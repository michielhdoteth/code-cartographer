package context

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"carto/analyze"
	"carto/lib/parser"
)

// Options for generating context
type Options struct {
	FilePath       string
	TokenBudget    int
	IncludeImports bool
	IncludeCallers bool
	RootPath       string
}

// Result contains the generated context
type Result struct {
	FilePath     string         `json:"file_path"`
	TokensUsed   int            `json:"tokens_used"`
	TokenBudget  int            `json:"token_budget"`
	Blocks       []ContextBlock `json:"blocks"`
	Truncated    bool           `json:"truncated"`
	TotalSymbols int            `json:"total_symbols"`
}

// ContextBlock is a chunk of code with metadata
type ContextBlock struct {
	FilePath string `json:"file_path"`
	Line     int    `json:"line,omitempty"`
	Tokens   int    `json:"tokens"`
	Content  string `json:"content"`
	Type     string `json:"type"` // "main", "import", "caller"
}

// Generate produces token-budgeted context for a file
func Generate(opts Options) (*Result, error) {
	absPath, err := filepath.Abs(opts.FilePath)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	// Read the main file
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

	// Add main file
	mainBlock := ContextBlock{
		FilePath: absPath,
		Tokens:   tokens,
		Content:  codeStr,
		Type:     "main",
	}
	result.Blocks = append(result.Blocks, mainBlock)
	result.TokensUsed += tokens

	// Parse to get symbols
	lang := parser.DetectLanguage(absPath)
	p := parser.GetParser(lang)
	if p != nil {
		parseResult, err := p.Parse(codeStr, absPath)
		if err == nil {
			result.TotalSymbols = len(parseResult.Exports)
		}
	}

	// Add imports if requested and budget allows
	if opts.IncludeImports && result.TokensUsed < opts.TokenBudget {
		result.Blocks = append(result.Blocks, addImportContext(absPath, opts.RootPath, &result.TokensUsed, opts.TokenBudget)...)
	}

	// Add callers if requested and budget allows
	if opts.IncludeCallers && result.TokensUsed < opts.TokenBudget {
		graph, err := analyze.BuildGraph(opts.RootPath, "")
		if err == nil {
			result.Blocks = append(result.Blocks, addCallerContext(absPath, graph, &result.TokensUsed, opts.TokenBudget)...)
		}
	}

	// Check if truncated
	if result.TokensUsed >= opts.TokenBudget {
		result.Truncated = true
	}

	return result, nil
}

// estimateTokens estimates token count (~4 chars per token)
func estimateTokens(code string) int {
	return len(code) / 4
}

// addImportContext adds imported files within budget
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

		// Try to resolve the import to a file
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
			// Add truncated version
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

// addCallerContext adds files that call symbols from the given file
func addCallerContext(filePath string, graph *analyze.ModuleGraph, used *int, budget int) []ContextBlock {
	var blocks []ContextBlock

	relPath, _ := filepath.Rel(graph.RootPath, filePath)

	for _, fileNode := range graph.Files {
		if *used >= budget {
			break
		}

		for _, imp := range fileNode.Imports {
			if strings.Contains(imp.Source, relPath) || strings.Contains(imp.Source, filepath.Base(filePath)) {
				// This file imports from our target
				callerPath := filepath.Join(graph.RootPath, fileNode.Path)
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

// resolveImport attempts to resolve an import path to a file
func resolveImport(importPath string, rootPath string) string {
	// Try direct path
	candidates := []string{
		filepath.Join(rootPath, importPath),
		filepath.Join(rootPath, importPath+".go"),
		filepath.Join(rootPath, importPath+".ts"),
		filepath.Join(rootPath, importPath+".js"),
		filepath.Join(rootPath, importPath+".py"),
	}

	// Try relative to the root
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}

	return ""
}

// truncateToTokens truncates code to fit within token budget
func truncateToTokens(code string, maxTokens int) string {
	maxChars := maxTokens * 4
	if len(code) <= maxChars {
		return code
	}
	return code[:maxChars] + "\n// ... truncated"
}

// FormatResult formats the context result for display
func FormatResult(result *Result) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("=== carto context %s ===\n", filepath.Base(result.FilePath)))
	sb.WriteString(fmt.Sprintf("Tokens: %d / %d", result.TokensUsed, result.TokenBudget))
	if result.Truncated {
		sb.WriteString(" (truncated)")
	}
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("Symbols: %d\n\n", result.TotalSymbols))

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
