package codegraph

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Skeleton struct {
	SchemaVersion  int              `json:"schema_version"`
	FilePath       string           `json:"file_path"`
	View           string           `json:"view"`
	Level          int              `json:"level"`
	Language       string           `json:"language"`
	GraphFresh     bool             `json:"graph_fresh"`
	FileHash       string           `json:"file_hash"`
	TokenBudget    int              `json:"token_budget_estimate"`
	TokensUsed     int              `json:"tokens_used_estimate"`
	TotalLines     int              `json:"total_lines"`
	TotalSymbols   int              `json:"total_symbols"`
	Imports        []SkeletonImport `json:"imports"`
	Symbols        []SkeletonSymbol `json:"symbols"`
	OmittedSymbols int              `json:"omitted_symbols"`
	SuggestedNext  string           `json:"suggested_next"`
	Savings        *Savings         `json:"savings,omitempty"`
}

type SkeletonImport struct {
	Source string `json:"source"`
	Line   int    `json:"line"`
}

type SkeletonSymbol struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Kind               string `json:"kind"`
	Signature          string `json:"signature,omitempty"`
	StartLine          int    `json:"start_line"`
	EndLine            int    `json:"end_line"`
	LineSpan           int    `json:"line_span"`
	BodyTokensEstimate int    `json:"body_tokens_estimate"`
	Exported           bool   `json:"exported,omitempty"`
	RangeConfidence    string `json:"range_confidence,omitempty"`
	CallerCount        int    `json:"caller_count"`
	CalleeCount        int    `json:"callee_count"`
	ParentID           string `json:"parent_id,omitempty"`
}

func (index *Index) GenerateSkeleton(path string, budget int) Skeleton {
	rel := normalizeRelativePath(index.RootPath, path)
	result := Skeleton{
		SchemaVersion: 1, FilePath: rel, View: "skeleton", Level: 1,
		TokenBudget: budget, Imports: []SkeletonImport{}, Symbols: []SkeletonSymbol{},
		SuggestedNext: fmt.Sprintf("carto chunk %s --list --json", rel),
	}
	if freshness, err := CheckFreshness(index.RootPath); err == nil {
		result.GraphFresh = freshness.Fresh
	}
	file, ok := index.Files[rel]
	if !ok {
		return result
	}
	result.Language = file.Language
	result.FileHash = shortHash(file.Hash)
	result.TotalLines = file.Lines
	for _, imp := range file.Imports {
		result.Imports = append(result.Imports, SkeletonImport{Source: imp.Source, Line: imp.Line})
	}
	lines := readFileLines(index.RootPath, rel)
	nodes := index.Outline(rel)
	result.TotalSymbols = len(nodes)
	for _, node := range nodes {
		symbol := SkeletonSymbol{
			ID: node.ID, Name: node.Name, Kind: node.Kind, Signature: node.Signature,
			StartLine: node.StartLine, EndLine: node.EndLine, Exported: node.Exported,
			RangeConfidence: node.RangeConfidence,
		}
		if symbol.Signature == "" {
			symbol.Signature = node.Name
		}
		if symbol.EndLine >= symbol.StartLine {
			symbol.LineSpan = symbol.EndLine - symbol.StartLine + 1
			symbol.BodyTokensEstimate = EstimateTokens(strings.Join(symbolLines(lines, symbol.StartLine, symbol.EndLine), "\n"))
		}
		symbol.CallerCount, symbol.CalleeCount = index.symbolEdgeCounts(node.ID)
		symbol.ParentID = containingSymbolID(nodes, node)
		result.Symbols = append(result.Symbols, symbol)
	}

	for len(result.Symbols) > 0 {
		result.TokensUsed = EstimateTokens(result.RenderMarkdown())
		if result.TokenBudget <= 0 || result.TokensUsed <= result.TokenBudget {
			break
		}
		result.Symbols = result.Symbols[:len(result.Symbols)-1]
		result.OmittedSymbols++
	}
	result.TokensUsed = EstimateTokens(result.RenderMarkdown())
	result.Savings = BuildSavings(result.TokensUsed, renderedLines(result.RenderMarkdown()), int(file.Size/4), file.Lines, "file")
	return result
}

func (index *Index) symbolEdgeCounts(id string) (callers, callees int) {
	for _, edge := range index.Edges {
		if edge.Type != "calls" {
			continue
		}
		if edge.Target == id {
			callers++
		}
		if edge.Source == id {
			callees++
		}
	}
	return
}

func containingSymbolID(nodes []Node, node Node) string {
	best := ""
	bestSpan := int(^uint(0) >> 1)
	for _, candidate := range nodes {
		if candidate.ID == node.ID || candidate.StartLine >= node.StartLine || candidate.EndLine <= node.EndLine {
			continue
		}
		span := candidate.EndLine - candidate.StartLine
		if span < bestSpan {
			bestSpan = span
			best = candidate.ID
		}
	}
	return best
}

func normalizeRelativePath(root, path string) string {
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	abs, err := filepath.Abs(path)
	if err == nil {
		if rel, err := filepath.Rel(root, abs); err == nil {
			return filepath.ToSlash(rel)
		}
	}
	return filepath.ToSlash(filepath.Clean(path))
}

func shortHash(hash string) string {
	if len(hash) > 12 {
		return hash[:12]
	}
	return hash
}

func readFileLines(root, rel string) []string {
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		return nil
	}
	return strings.Split(string(data), "\n")
}

func symbolLines(lines []string, start, end int) []string {
	if start < 1 || start > len(lines) {
		return nil
	}
	if end > len(lines) {
		end = len(lines)
	}
	if end < start {
		return nil
	}
	return lines[start-1 : end]
}

func renderedLines(text string) int {
	if text == "" {
		return 0
	}
	return strings.Count(text, "\n") + 1
}

func (s Skeleton) RenderMarkdown() string {
	var b strings.Builder
	fmt.Fprintf(&b, "=== carto skeleton %s ===\n", s.FilePath)
	fmt.Fprintf(&b, "Tokens: %d / %d · Level 1/3", s.TokensUsed, s.TokenBudget)
	if s.Savings != nil {
		fmt.Fprintf(&b, " · Efficiency %.0f/100", s.Savings.ContextEfficiencyScore)
	}
	b.WriteString("\n")
	fmt.Fprintf(&b, "File: %s · %d lines · %d symbols · hash %s · fresh %t\n\n",
		s.FilePath, s.TotalLines, s.TotalSymbols, s.FileHash, s.GraphFresh)
	if len(s.Imports) > 0 {
		fmt.Fprintf(&b, "imports (%d): ", len(s.Imports))
		for i, imp := range s.Imports {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(imp.Source)
		}
		b.WriteString("\n\n")
	}
	for _, symbol := range s.Symbols {
		flags := ""
		if symbol.Exported {
			flags = " exported"
		}
		fmt.Fprintf(&b, "  %5d-%-5d %-40s %4d ln %5d tok  callers %d callees %d%s\n",
			symbol.StartLine, symbol.EndLine, symbol.Signature, symbol.LineSpan,
			symbol.BodyTokensEstimate, symbol.CallerCount, symbol.CalleeCount, flags)
	}
	if s.OmittedSymbols > 0 {
		fmt.Fprintf(&b, "… %d symbols omitted\n", s.OmittedSymbols)
	}
	if s.Savings != nil {
		fmt.Fprintf(&b, "… %d lines avoided\n", s.Savings.AvoidedLines)
	}
	fmt.Fprintf(&b, "next: %s\n", s.SuggestedNext)
	return b.String()
}

// SortedImports returns imports sorted by source for stable rendering.
func (s Skeleton) SortedImports() []SkeletonImport {
	out := append([]SkeletonImport{}, s.Imports...)
	sort.Slice(out, func(i, j int) bool { return out[i].Source < out[j].Source })
	return out
}
