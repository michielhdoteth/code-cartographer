package fix

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"carto/analyze"
	"carto/codegraph"
)

// GeneratePrompt creates an LLM-ready refactor prompt for one structural
// finding. It never calls a model; it only assembles deterministic context.
func GeneratePrompt(root string, finding analyze.StructureFinding, maxTokens int) (string, error) {
	if maxTokens <= 0 {
		maxTokens = 4000
	}
	index, err := codegraph.Build(root)
	if err != nil {
		return "", fmt.Errorf("build code graph: %w", err)
	}

	symbols := append([]string{}, finding.Evidence.Symbols...)
	if len(symbols) == 0 {
		for _, file := range finding.Files {
			rel, relErr := filepath.Rel(root, file)
			if relErr != nil {
				continue
			}
			for _, node := range index.Outline(filepath.ToSlash(rel)) {
				symbols = append(symbols, node.Name)
			}
		}
	}
	sort.Strings(symbols)

	callSites := make(map[string]bool)
	for _, symbol := range symbols {
		for _, hit := range index.TraceDepth(symbol, 1, "callers") {
			line := ""
			if hit.Line > 0 {
				line = fmt.Sprintf(":%d", hit.Line)
			}
			callSites[fmt.Sprintf("%s%s calls %s", hit.Node.File, line, symbol)] = true
		}
	}
	calls := sortedKeys(callSites)

	var sb strings.Builder
	sb.WriteString("You are refactoring a codebase to follow deep modules and narrow interfaces.\n\n")
	sb.WriteString("## Problem\n")
	sb.WriteString(fmt.Sprintf("%s [%s]\n%s\n\n", finding.Rule, finding.Severity, finding.Message))
	sb.WriteString("## Evidence\n")
	sb.WriteString(fmt.Sprintf("- Module: %s\n", finding.Module))
	if finding.Evidence.TotalExports > 0 {
		sb.WriteString(fmt.Sprintf("- Exports used externally: %d of %d (%.1f lines per export)\n",
			finding.Evidence.UsedExports, finding.Evidence.TotalExports, finding.Evidence.LinesPerExport))
	}
	if finding.Evidence.MutualEdges > 0 {
		sb.WriteString(fmt.Sprintf("- Mutual references: %d\n", finding.Evidence.MutualEdges))
	}
	if len(finding.Evidence.TopDirs) > 0 {
		sb.WriteString(fmt.Sprintf("- Top-level areas: %s\n", strings.Join(finding.Evidence.TopDirs, ", ")))
	}
	sb.WriteString("\n")

	if len(finding.Files) > 0 {
		sb.WriteString("## Files in scope\n")
		for _, file := range finding.Files {
			sb.WriteString(fmt.Sprintf("- %s\n", file))
		}
		sb.WriteString("\n")
	}
	if len(symbols) > 0 {
		sb.WriteString("## Symbols involved\n")
		sb.WriteString("- " + strings.Join(symbols, ", ") + "\n\n")
	}
	if len(calls) > 0 {
		sb.WriteString("## Call sites (from the local code graph)\n")
		for _, call := range calls {
			sb.WriteString("- " + call + "\n")
		}
		sb.WriteString("\n")
	}
	sb.WriteString("## Task\n")
	sb.WriteString("1. Inventory every place that must change together; do not edit yet.\n")
	sb.WriteString("2. Work out the intended layer or responsibility from first principles. Look for one change that removes several problems; do not fix symptoms one by one.\n")
	sb.WriteString("3. Make the module deep: a small stable interface that hides implementation and vocabulary.\n")
	sb.WriteString("4. Update callers to use only that interface. Remove the now-unnecessary special cases and dead wrappers.\n")
	sb.WriteString("5. Run `carto analyze structure` again and confirm this finding is gone and no new ones appeared.\n\n")
	sb.WriteString("## Constraints\n")
	sb.WriteString("- Preserve intended behaviour and public API semantics unless the finding explicitly targets them.\n")
	sb.WriteString("- Do not add speculative abstractions, configuration, or wrappers.\n")
	sb.WriteString("- Keep changes scoped to the files above unless a caller must be updated.\n")
	sb.WriteString(fmt.Sprintf("- Keep your response within roughly %d tokens.\n", maxTokens))

	prompt := sb.String()
	if codegraph.EstimateTokens(prompt) > maxTokens {
		prompt = truncateToTokenBudget(prompt, maxTokens)
	}
	return prompt, nil
}

func truncateToTokenBudget(text string, maxTokens int) string {
	marker := "\n\n[truncated to fit token budget]"
	limit := maxTokens*4 - len(marker)
	if limit <= 0 {
		return marker
	}
	if len(text) > limit {
		text = text[:limit]
	}
	return strings.TrimRight(text, "\n") + marker
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
