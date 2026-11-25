package context

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"carto/codegraph"
	"carto/lib/parser"
)

type Options struct {
	FilePath        string
	TokenBudget     int
	IncludeImports  bool
	IncludeCallers  bool
	IncludeCohesion bool
	RootPath        string
	// Progressive-disclosure view selectors. When any is set, the default
	// whole-file path is skipped.
	Outline     bool
	Symbols     []string
	Lines       []string
	ChunkIndex  int
	ChunkBudget int
}

// ContextSavings compares produced context with a naive baseline. It is an
// alias of the shared codegraph Savings type so line-level efficiency fields
// stay in sync across commands.
type ContextSavings = codegraph.Savings

type Result struct {
	FilePath     string          `json:"file_path"`
	TokensUsed   int             `json:"tokens_used"`
	TokenBudget  int             `json:"token_budget"`
	Blocks       []ContextBlock  `json:"blocks"`
	Truncated    bool            `json:"truncated"`
	TotalSymbols int             `json:"total_symbols"`
	Savings      *ContextSavings   `json:"savings,omitempty"`
	Cohesion     *CohesionInfo     `json:"cohesion,omitempty"`
	View         string            `json:"view,omitempty"`
	Skeleton     *codegraph.Skeleton `json:"skeleton,omitempty"`
	Chunk        *codegraph.Chunk    `json:"chunk,omitempty"`
	Selection    *Selection         `json:"selection,omitempty"`
	SuggestedNext string            `json:"suggested_next,omitempty"`
}

type ContextBlock struct {
	FilePath string `json:"file_path"`
	Line     int    `json:"line,omitempty"`
	Tokens   int    `json:"tokens"`
	Content  string `json:"content"`
	Type     string `json:"type"` // "main", "import", "caller"
}

func Generate(opts Options) (*Result, error) {
	absPath, err := filepath.Abs(opts.FilePath)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}
	if opts.Outline || len(opts.Symbols) > 0 || len(opts.Lines) > 0 || (opts.ChunkIndex >= 0 && opts.ChunkBudget > 0) {
		return generateView(opts, absPath)
	}

	code, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	codeStr := string(code)
	tokens := estimateTokens(codeStr)

	result := &Result{
		FilePath:    absPath,
		TokenBudget: opts.TokenBudget,
		Blocks:      []ContextBlock{},
	}

	index, _ := codegraph.Build(opts.RootPath)
	if index != nil {
		result.Savings = calculateSavings(index, tokens)
		if opts.IncludeCohesion {
			result.Cohesion = calculateCohesion(index, absPath)
		}
	}

	mainBlock := ContextBlock{
		FilePath: absPath,
		Tokens:   tokens,
		Content:  codeStr,
		Type:     "main",
	}
	result.Blocks = append(result.Blocks, mainBlock)
	result.TokensUsed += tokens

	lang := parser.DetectLanguage(absPath)
	p := parser.GetParser(lang)
	if p != nil {
		parseResult, err := p.Parse(codeStr, absPath)
		if err == nil {
			result.TotalSymbols = len(parseResult.Exports)
		}
	}

	if opts.IncludeImports && result.TokensUsed < opts.TokenBudget {
		result.Blocks = append(result.Blocks, addImportContext(absPath, opts.RootPath, &result.TokensUsed, opts.TokenBudget)...)
	}

	if opts.IncludeCallers && result.TokensUsed < opts.TokenBudget && index != nil {
		result.Blocks = append(result.Blocks, addCallerContext(absPath, index, &result.TokensUsed, opts.TokenBudget)...)
	}

	if result.TokensUsed >= opts.TokenBudget {
		result.Truncated = true
	}

	return result, nil
}

// estimateTokens estimates token count (~4 chars per token)
func estimateTokens(code string) int {
	return codegraph.EstimateTokens(code)
}

// calculateSavings compares context size with a whole-repository dump. The
// baseline uses source-file byte counts from the graph index divided by four,
// the same provider-neutral estimate used everywhere else in carto.
func calculateSavings(index *codegraph.Index, mainTokens int) *ContextSavings {
	return index.EstimateSavings(mainTokens)
}

func addImportContext(filePath string, rootPath string, used *int, budget int) []ContextBlock {
	var blocks []ContextBlock

	lang := parser.DetectLanguage(filePath)
	p := parser.GetParser(lang)
	if p == nil {
		return blocks
	}

	code, err := os.ReadFile(filePath)
	if err != nil {
		return blocks
	}

	result, err := p.Parse(string(code), filePath)
	if err != nil {
		return blocks
	}

	for _, imp := range result.Imports {
		if *used >= budget {
			break
		}

		impPath := resolveImport(imp.Source, rootPath)
		if impPath == "" {
			continue
		}

		impCode, err := os.ReadFile(impPath)
		if err != nil {
			continue
		}

		impTokens := estimateTokens(string(impCode))
		if *used+impTokens > budget {
			remaining := budget - *used
			truncated := truncateToTokens(string(impCode), remaining)
			blocks = append(blocks, ContextBlock{
				FilePath: impPath,
				Tokens:   remaining,
				Content:  truncated,
				Type:     "import",
			})
			*used = budget
			break
		}

		blocks = append(blocks, ContextBlock{
			FilePath: impPath,
			Tokens:   impTokens,
			Content:  string(impCode),
			Type:     "import",
		})
		*used += impTokens
	}

	return blocks
}

func addCallerContext(filePath string, graph *codegraph.Index, used *int, budget int) []ContextBlock {
	var blocks []ContextBlock
	relPath, _ := filepath.Rel(graph.RootPath, filePath)
	seen := make(map[string]bool)
	for _, symbol := range graph.Outline(filepath.ToSlash(relPath)) {
		for _, hit := range graph.Trace(symbol.Name) {
			if hit.Direction != "caller" || hit.Node.File == filepath.ToSlash(relPath) || seen[hit.Node.File] {
				continue
			}
			seen[hit.Node.File] = true
			if *used >= budget {
				return blocks
			}
			callerPath := filepath.Join(graph.RootPath, filepath.FromSlash(hit.Node.File))
			callerCode, err := os.ReadFile(callerPath)
			if err != nil {
				continue
			}
			callerTokens := estimateTokens(string(callerCode))
			if *used+callerTokens > budget {
				remaining := budget - *used
				if remaining <= 0 {
					return blocks
				}
				callerCode = []byte(truncateToTokens(string(callerCode), remaining))
				callerTokens = remaining
			}
			blocks = append(blocks, ContextBlock{FilePath: callerPath, Tokens: callerTokens, Content: string(callerCode), Type: "caller"})
			*used += callerTokens
		}
	}
	// Import-only callers remain useful when a language parser cannot resolve calls.
	for path, file := range graph.Files {
		if *used >= budget {
			break
		}
		for _, imp := range file.Imports {
			if strings.Contains(imp.Source, relPath) || strings.Contains(imp.Source, filepath.Base(filePath)) {
				if seen[path] {
					break
				}
				callerPath := filepath.Join(graph.RootPath, filepath.FromSlash(path))
				callerCode, err := os.ReadFile(callerPath)
				if err != nil {
					continue
				}

				callerTokens := estimateTokens(string(callerCode))
				if *used+callerTokens > budget {
					remaining := budget - *used
					truncated := truncateToTokens(string(callerCode), remaining)
					blocks = append(blocks, ContextBlock{
						FilePath: callerPath,
						Tokens:   remaining,
						Content:  truncated,
						Type:     "caller",
					})
					*used = budget
					break
				}

				blocks = append(blocks, ContextBlock{
					FilePath: callerPath,
					Tokens:   callerTokens,
					Content:  string(callerCode),
					Type:     "caller",
				})
				*used += callerTokens
				break
			}
		}
	}

	return blocks
}

func resolveImport(importPath string, rootPath string) string {
	candidates := []string{
		filepath.Join(rootPath, importPath),
		filepath.Join(rootPath, importPath+".go"),
		filepath.Join(rootPath, importPath+".ts"),
		filepath.Join(rootPath, importPath+".js"),
		filepath.Join(rootPath, importPath+".py"),
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}

	return ""
}

func truncateToTokens(code string, maxTokens int) string {
	if maxTokens <= 0 {
		return ""
	}
	if codegraph.EstimateTokens(code) <= maxTokens {
		return code
	}
	marker := "\n// ... truncated"
	contentBudget := maxTokens - codegraph.EstimateTokens(marker)
	if contentBudget <= 0 {
		return marker[:min(len(marker), maxTokens*4)]
	}
	cut := min(len(code), contentBudget*4)
	for cut > 0 && cut < len(code) && (code[cut]&0xC0) == 0x80 {
		cut--
	}
	return strings.TrimRight(code[:cut], "\r\n") + marker
}

func FormatResult(result *Result) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("=== carto context %s ===\n", filepath.Base(result.FilePath)))
	sb.WriteString(fmt.Sprintf("Tokens: %d / %d", result.TokensUsed, result.TokenBudget))
	if result.Truncated {
		sb.WriteString(" (truncated)")
	}
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("Symbols: %d\n", result.TotalSymbols))
	if result.Savings != nil {
		sb.WriteString(fmt.Sprintf("Context savings: ~%d tokens saved (%.0f%% vs whole-repo estimate of %d)\n",
			result.Savings.SavedTokens, result.Savings.SavedPercent, result.Savings.WholeRepoTokens))
	}
	if result.Cohesion != nil {
		sb.WriteString(fmt.Sprintf("Structural cohesion: %d/100 — %s\n", result.Cohesion.Score, result.Cohesion.Verdict))
	}
	sb.WriteString("\n")

	for _, block := range result.Blocks {
		relPath, _ := filepath.Rel(result.FilePath, block.FilePath)
		if relPath == "" {
			relPath = filepath.Base(block.FilePath)
		}

		switch block.Type {
		case "main":
			sb.WriteString(fmt.Sprintf("--- %s (%d tokens) ---\n", filepath.Base(block.FilePath), block.Tokens))
		case "import":
			sb.WriteString(fmt.Sprintf("--- [import] %s (%d tokens) ---\n", relPath, block.Tokens))
		case "caller":
			sb.WriteString(fmt.Sprintf("--- [caller] %s (%d tokens) ---\n", relPath, block.Tokens))
		}

		sb.WriteString(block.Content)
		sb.WriteString("\n\n")
	}

	return sb.String()
}
