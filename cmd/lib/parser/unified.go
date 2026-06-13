package parser

import (
	"fmt"
	"regexp"
	"strings"
)

// UnifiedParser is a regex-based parser that works for multiple languages
type UnifiedParser struct {
	language Language
}

// NewUnifiedParser creates a new unified parser
func NewUnifiedParser(lang Language) *UnifiedParser {
	return &UnifiedParser{language: lang}
}

// Language returns the language
func (p *UnifiedParser) Language() Language {
	return p.language
}

// SupportedExtensions returns supported file extensions
func (p *UnifiedParser) SupportedExtensions() []string {
	return GetLanguageConfig(p.language).Extensions
}

// Parse parses source code using regex patterns
func (p *UnifiedParser) Parse(code string, filePath string) (*ParseResult, error) {
	result := &ParseResult{
		FilePath: filePath,
		Language: p.language,
		Nodes:    []Node{},
		Edges:    []Edge{},
		Exports:  []Export{},
		Imports:  []Import{},
		Metrics:  Metrics{},
	}

	switch p.language {
	case LanguageGo:
		p.parseGo(code, filePath, result)
	case LanguageJavaScript, LanguageTypeScript:
		p.parseJavaScript(code, filePath, result)
	case LanguagePython:
		p.parsePython(code, filePath, result)
	default:
		p.parseGeneric(code, filePath, result)
	}

	// Extract imports for all languages
	p.extractImports(code, filePath, result)

	// Calculate metrics
	result.CalculateMetrics()

	return result, nil
}

// parseGo parses Go source code
func (p *UnifiedParser) parseGo(code string, filePath string, result *ParseResult) {
	lines := strings.Split(code, "\n")

	// Function declarations: func name(...) or func (receiver) name(...)
	funcRegex := regexp.MustCompile(`^\s*func\s+(?:\([^)]+\)\s+)?(\w+)\s*\(`)
	// Type declarations: type Name struct/interface
	typeRegex := regexp.MustCompile(`^\s*type\s+(\w+)\s+(struct|interface)\s*`)
	// Const/Var declarations at package level
	constRegex := regexp.MustCompile(`^\s*(?:const|var)\s+(\w+)`)

	for i, line := range lines {
		// Functions
		if matches := funcRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			nodeType := NodeTypeFunction
			if strings.HasPrefix(line, "func (") {
				nodeType = NodeTypeMethod
			}
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:func:%s", filePath, name),
				Name:      name,
				Type:      nodeType,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
				Exported:  isExported(name),
				Metrics:   p.calculateLineMetrics(code, i),
			})
			if isExported(name) {
				result.Exports = append(result.Exports, Export{
					Name:     name,
					File:     filePath,
					Type:     nodeType,
					Line:     i + 1,
					Exported: true,
				})
			}
		}

		// Types
		if matches := typeRegex.FindStringSubmatch(line); len(matches) > 2 {
			name := matches[1]
			kind := NodeTypeStruct
			if matches[2] == "interface" {
				kind = NodeTypeInterface
			}
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:type:%s", filePath, name),
				Name:      name,
				Type:      kind,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
				Exported:  isExported(name),
			})
			if isExported(name) {
				result.Exports = append(result.Exports, Export{
					Name:     name,
					File:     filePath,
					Type:     kind,
					Line:     i + 1,
					Exported: true,
				})
			}
		}

		// Constants
		if matches := constRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			if isExported(name) {
				result.Exports = append(result.Exports, Export{
					Name:     name,
					File:     filePath,
					Type:     NodeTypeConstant,
					Line:     i + 1,
					Exported: true,
				})
			}
		}
	}
}

// parseJavaScript parses JavaScript/TypeScript
func (p *UnifiedParser) parseJavaScript(code string, filePath string, result *ParseResult) {
	lines := strings.Split(code, "\n")

	// Function declarations: function name(...) or async function name(...)
	funcRegex := regexp.MustCompile(`^\s*(?:async\s+)?function\s+(\w+)\s*\(`)
	// Arrow functions: const/let name = (...) => or const/let name = async (...) =>
	arrowRegex := regexp.MustCompile(`^\s*(?:const|let|var)\s+(\w+)\s*=\s*(?:async\s+)?\([^)]*\)\s*=>`)
	// Class declarations: class Name
	classRegex := regexp.MustCompile(`^\s*class\s+(\w+)`)
	// Method definitions in classes: name(...) {
	methodRegex := regexp.MustCompile(`^\s{2,}(\w+)\s*\([^)]*\)\s*\{`)
	// Export statements
	exportRegex := regexp.MustCompile(`^\s*export\s+(?:default\s+)?(?:const|let|var|function|class)\s+(\w+)`)

	for i, line := range lines {
		// Functions
		if matches := funcRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:func:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeFunction,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
				Exported:  true,
				Metrics:   p.calculateLineMetrics(code, i),
			})
			result.Exports = append(result.Exports, Export{
				Name:     name,
				File:     filePath,
				Type:     NodeTypeFunction,
				Line:     i + 1,
				Exported: true,
			})
		}

		// Arrow functions
		if matches := arrowRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:func:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeFunction,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
				Exported:  false,
				Metrics:   p.calculateLineMetrics(code, i),
			})
		}

		// Classes
		if matches := classRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:class:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeClass,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
				Exported:  true,
			})
			result.Exports = append(result.Exports, Export{
				Name:     name,
				File:     filePath,
				Type:     NodeTypeClass,
				Line:     i + 1,
				Exported: true,
			})
		}

		// Methods
		if matches := methodRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:method:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeMethod,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
				Metrics:   p.calculateLineMetrics(code, i),
			})
		}

		// Exports
		if matches := exportRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Exports = append(result.Exports, Export{
				Name:     name,
				File:     filePath,
				Type:     NodeTypeFunction,
				Line:     i + 1,
				Exported: true,
			})
		}
	}
}

// parsePython parses Python source code
func (p *UnifiedParser) parsePython(code string, filePath string, result *ParseResult) {
	lines := strings.Split(code, "\n")

	// Function definitions: def name(...)
	funcRegex := regexp.MustCompile(`^\s*(?:async\s+)?def\s+(\w+)\s*\(`)
	// Class definitions: class Name(...)
	classRegex := regexp.MustCompile(`^\s*class\s+(\w+)`)
	// Decorators: @name
	decoratorRegex := regexp.MustCompile(`^\s*@(\w+)`)
	// Module-level constants: CONST_NAME =
	constRegex := regexp.MustCompile(`^\s*([A-Z][A-Z0-9_]*)\s*=`)

	for i, line := range lines {
		indent := len(line) - len(strings.TrimLeft(line, " \t"))

		// Functions
		if matches := funcRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			nodeType := NodeTypeFunction
			if indent > 0 {
				nodeType = NodeTypeMethod
			}
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:func:%s", filePath, name),
				Name:      name,
				Type:      nodeType,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
				Exported:  name[0] >= 'A' && name[0] <= 'Z',
				Metrics:   p.calculateLineMetrics(code, i),
			})
			if name[0] >= 'A' && name[0] <= 'Z' {
				result.Exports = append(result.Exports, Export{
					Name:     name,
					File:     filePath,
					Type:     nodeType,
					Line:     i + 1,
					Exported: true,
				})
			}
		}

		// Classes
		if matches := classRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:class:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeClass,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
				Exported:  name[0] >= 'A' && name[0] <= 'Z',
			})
			if name[0] >= 'A' && name[0] <= 'Z' {
				result.Exports = append(result.Exports, Export{
					Name:     name,
					File:     filePath,
					Type:     NodeTypeClass,
					Line:     i + 1,
					Exported: true,
				})
			}
		}

		// Decorators
		if matches := decoratorRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:decorator:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeDecorator,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
			})
		}

		// Constants
		if matches := constRegex.FindStringSubmatch(line); len(matches) > 1 && indent == 0 {
			name := matches[1]
			result.Exports = append(result.Exports, Export{
				Name:     name,
				File:     filePath,
				Type:     NodeTypeConstant,
				Line:     i + 1,
				Exported: true,
			})
		}
	}
}

// parseGeneric parses using generic patterns for unknown languages
func (p *UnifiedParser) parseGeneric(code string, filePath string, result *ParseResult) {
	lines := strings.Split(code, "\n")

	// Generic function-like patterns
	funcPatterns := []string{
		`^\s*(?:public|private|protected)?\s*(?:static)?\s*\w+\s+(\w+)\s*\(`,
		`^\s*(?:def|fn|func|function)\s+(\w+)\s*\(`,
	}

	for i, line := range lines {
		for _, pattern := range funcPatterns {
			if matches := regexp.MustCompile(pattern).FindStringSubmatch(line); len(matches) > 1 {
				name := matches[1]
				result.Nodes = append(result.Nodes, Node{
					ID:        fmt.Sprintf("%s:func:%s", filePath, name),
					Name:      name,
					Type:      NodeTypeFunction,
					File:      filePath,
					StartLine: i + 1,
					EndLine:   i + 1,
					Metrics:   p.calculateLineMetrics(code, i),
				})
			}
		}
	}
}

// extractImports extracts import statements
func (p *UnifiedParser) extractImports(code string, filePath string, result *ParseResult) {
	lines := strings.Split(code, "\n")

	switch p.language {
	case LanguageGo:
		// import "path" or import (
		//   "path1"
		//   "path2"
		// )
		importRegex := regexp.MustCompile(`^\s*import\s+(?:\(\s*)?["']([^"']+)["']`)
		for i, line := range lines {
			if matches := importRegex.FindStringSubmatch(line); len(matches) > 1 {
				result.Imports = append(result.Imports, Import{
					Name:   extractNameFromPath(matches[1]),
					Source: matches[1],
					Line:   i + 1,
				})
			}
		}

	case LanguageJavaScript, LanguageTypeScript:
		// import x from 'path' or import { x } from 'path'
		importRegex := regexp.MustCompile(`import\s+(?:{[^}]+}|\*\s+as\s+\w+|\w+)\s+from\s+['"]([^'"]+)['"]`)
		// require('path')
		requireRegex := regexp.MustCompile(`require\s*\(\s*['"]([^'"]+)['"]\s*\)`)
		for i, line := range lines {
			if matches := importRegex.FindStringSubmatch(line); len(matches) > 1 {
				result.Imports = append(result.Imports, Import{
					Name:   extractNameFromPath(matches[1]),
					Source: matches[1],
					Line:   i + 1,
				})
			}
			if matches := requireRegex.FindStringSubmatch(line); len(matches) > 1 {
				result.Imports = append(result.Imports, Import{
					Name:   extractNameFromPath(matches[1]),
					Source: matches[1],
					Line:   i + 1,
				})
			}
		}

	case LanguagePython:
		// from x import y or import x
		fromImportRegex := regexp.MustCompile(`^\s*from\s+(\S+)\s+import\s+`)
		importRegex := regexp.MustCompile(`^\s*import\s+(\S+)`)
		for i, line := range lines {
			if matches := fromImportRegex.FindStringSubmatch(line); len(matches) > 1 {
				result.Imports = append(result.Imports, Import{
					Name:   extractNameFromPath(matches[1]),
					Source: matches[1],
					Line:   i + 1,
				})
			}
			if matches := importRegex.FindStringSubmatch(line); len(matches) > 1 {
				result.Imports = append(result.Imports, Import{
					Name:   extractNameFromPath(matches[1]),
					Source: matches[1],
					Line:   i + 1,
				})
			}
		}
	}
}

// calculateLineMetrics calculates metrics for a single line
func (p *UnifiedParser) calculateLineMetrics(code string, lineNum int) Metrics {
	lines := strings.Split(code, "\n")
	if lineNum >= len(lines) {
		return Metrics{}
	}

	line := lines[lineNum]
	metrics := Metrics{
		LinesOfCode: 1,
		Cyclomatic:  1,
		Cognitive:   0,
	}

	// Count control flow keywords
	controlFlow := regexp.MustCompile(`\b(if|else|for|while|switch|case|catch|&&|\|\|)\b`)
	matches := controlFlow.FindAllString(line, -1)
	metrics.Cyclomatic += len(matches)
	metrics.Cognitive = metrics.Cyclomatic

	return metrics
}

// Helper functions

func isExported(name string) bool {
	if len(name) == 0 {
		return false
	}
	return name[0] >= 'A' && name[0] <= 'Z'
}

func extractNameFromPath(path string) string {
	parts := strings.Split(path, "/")
	name := parts[len(parts)-1]
	// Remove common extensions
	name = strings.TrimSuffix(name, ".go")
	name = strings.TrimSuffix(name, ".js")
	name = strings.TrimSuffix(name, ".ts")
	name = strings.TrimSuffix(name, ".py")
	name = strings.TrimSuffix(name, ".jsx")
	name = strings.TrimSuffix(name, ".tsx")
	return name
}

// RegisterUnifiedParsers registers parsers for all supported languages
func RegisterUnifiedParsers() {
	languages := []Language{
		LanguagePython,
		LanguageJava,
		LanguageC,
		LanguageCpp,
		LanguageRuby,
		LanguagePHP,
		LanguageRust,
		LanguageSwift,
		LanguageKotlin,
	}

	for _, lang := range languages {
		RegisterParser(NewUnifiedParser(lang))
	}
}

// init registers the unified parsers
func init() {
	// Register tree-sitter parsers first (they take precedence)
	RegisterTreeSitterParsers()
	// Then register regex-based parsers for languages without tree-sitter support
	RegisterUnifiedParsers()
}