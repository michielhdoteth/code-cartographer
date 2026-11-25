package parser

import (
	"fmt"
	"regexp"
	"strings"
)

func (p *UnifiedParser) extractImports(code string, filePath string, result *ParseResult) {
	lines := strings.Split(code, "\n")

	switch p.language {
	case LanguageGo:
		singleImportRegex := regexp.MustCompile(`^\s*import\s+"([^"]+)"`)
		// Block-style individual lines: "path" inside import ( ... )
		blockImportRegex := regexp.MustCompile(`^\s*"([^"]+)"`)
		inImportBlock := false
		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			if singleImportRegex.MatchString(line) {
				if matches := singleImportRegex.FindStringSubmatch(line); len(matches) > 1 {
					result.Imports = append(result.Imports, Import{
						Name:   extractNameFromPath(matches[1]),
						Source: matches[1],
						Line:   i + 1,
					})
				}
				continue
			}
			if trimmed == "import (" {
				inImportBlock = true
				continue
			}
			if inImportBlock && trimmed == ")" {
				inImportBlock = false
				continue
			}
			if inImportBlock {
				if matches := blockImportRegex.FindStringSubmatch(line); len(matches) > 1 {
					result.Imports = append(result.Imports, Import{
						Name:   extractNameFromPath(matches[1]),
						Source: matches[1],
						Line:   i + 1,
					})
				}
			}
		}

	case LanguageJavaScript, LanguageTypeScript:
		importRegex := regexp.MustCompile(`import\s+(?:{[^}]+}|\*\s+as\s+\w+|\w+)\s+from\s+['"]([^'"]+)['"]`)
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

	case LanguageJava:
		importRegex := regexp.MustCompile(`^\s*import\s+([\w.*]+)\s*;`)
		for i, line := range lines {
			if matches := importRegex.FindStringSubmatch(line); len(matches) > 1 {
				result.Imports = append(result.Imports, Import{
					Name:   extractNameFromPath(matches[1]),
					Source: matches[1],
					Line:   i + 1,
				})
			}
		}

	case LanguageC, LanguageCpp:
		includeRegex := regexp.MustCompile(`^\s*#include\s+[<"]([^>"]+)[>"]`)
		for i, line := range lines {
			if matches := includeRegex.FindStringSubmatch(line); len(matches) > 1 {
				result.Imports = append(result.Imports, Import{
					Name:   extractNameFromPath(matches[1]),
					Source: matches[1],
					Line:   i + 1,
				})
			}
		}
		// C++20 imports
		if p.language == LanguageCpp {
			importRegex := regexp.MustCompile(`^\s*import\s+<([^>]+)>`)
			for i, line := range lines {
				if matches := importRegex.FindStringSubmatch(line); len(matches) > 1 {
					result.Imports = append(result.Imports, Import{
						Name:   extractNameFromPath(matches[1]),
						Source: matches[1],
						Line:   i + 1,
					})
				}
			}
		}

	case LanguageCSharp:
		usingRegex := regexp.MustCompile(`^\s*using\s+([\w.]+)\s*;`)
		for i, line := range lines {
			if matches := usingRegex.FindStringSubmatch(line); len(matches) > 1 {
				result.Imports = append(result.Imports, Import{
					Name:   extractNameFromPath(matches[1]),
					Source: matches[1],
					Line:   i + 1,
				})
			}
		}

	case LanguageRuby:
		requireRegex := regexp.MustCompile(`^\s*(?:require_relative|require|gem)\s+['"]([^'"]+)['"]`)
		for i, line := range lines {
			if matches := requireRegex.FindStringSubmatch(line); len(matches) > 1 {
				result.Imports = append(result.Imports, Import{
					Name:   extractNameFromPath(matches[1]),
					Source: matches[1],
					Line:   i + 1,
				})
			}
		}

	case LanguagePHP:
		useRegex := regexp.MustCompile(`^\s*use\s+([\w\\]+)\s*;`)
		requireRegex := regexp.MustCompile(`^\s*(?:require|include)(?:_once)?\s+['"]([^'"]+)['"]`)
		for i, line := range lines {
			if matches := useRegex.FindStringSubmatch(line); len(matches) > 1 {
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

	case LanguageRust:
		useRegex := regexp.MustCompile(`^\s*use\s+([\w:]+)\s*;`)
		externCrateRegex := regexp.MustCompile(`^\s*extern\s+crate\s+(\w+)`)
		for i, line := range lines {
			if matches := useRegex.FindStringSubmatch(line); len(matches) > 1 {
				result.Imports = append(result.Imports, Import{
					Name:   extractNameFromPath(matches[1]),
					Source: matches[1],
					Line:   i + 1,
				})
			}
			if matches := externCrateRegex.FindStringSubmatch(line); len(matches) > 1 {
				result.Imports = append(result.Imports, Import{
					Name:   extractNameFromPath(matches[1]),
					Source: matches[1],
					Line:   i + 1,
				})
			}
		}

	case LanguageSwift:
		importRegex := regexp.MustCompile(`^\s*import\s+(\w+)`)
		for i, line := range lines {
			if matches := importRegex.FindStringSubmatch(line); len(matches) > 1 {
				result.Imports = append(result.Imports, Import{
					Name:   extractNameFromPath(matches[1]),
					Source: matches[1],
					Line:   i + 1,
				})
			}
		}

	case LanguageKotlin:
		importRegex := regexp.MustCompile(`^\s*import\s+([\w.]+)`)
		for i, line := range lines {
			if matches := importRegex.FindStringSubmatch(line); len(matches) > 1 {
				result.Imports = append(result.Imports, Import{
					Name:   extractNameFromPath(matches[1]),
					Source: matches[1],
					Line:   i + 1,
				})
			}
		}

	case LanguageScala:
		importRegex := regexp.MustCompile(`^\s*import\s+([\w.]+)`)
		for i, line := range lines {
			if matches := importRegex.FindStringSubmatch(line); len(matches) > 1 {
				result.Imports = append(result.Imports, Import{
					Name:   extractNameFromPath(matches[1]),
					Source: matches[1],
					Line:   i + 1,
				})
			}
		}

	case LanguageShell:
		sourceRegex := regexp.MustCompile(`^\s*(?:source|\.)\s+(\S+)`)
		for i, line := range lines {
			if matches := sourceRegex.FindStringSubmatch(line); len(matches) > 1 {
				result.Imports = append(result.Imports, Import{
					Name:   extractNameFromPath(matches[1]),
					Source: matches[1],
					Line:   i + 1,
				})
			}
		}

	case LanguageHTML:
		linkRegex := regexp.MustCompile(`<link\s+[^>]*href="([^"]+)"`)
		scriptRegex := regexp.MustCompile(`<script\s+[^>]*src="([^"]+)"`)
		imgRegex := regexp.MustCompile(`<img\s+[^>]*src="([^"]+)"`)
		for i, line := range lines {
			if matches := linkRegex.FindStringSubmatch(line); len(matches) > 1 {
				result.Imports = append(result.Imports, Import{
					Name:   extractNameFromPath(matches[1]),
					Source: matches[1],
					Line:   i + 1,
				})
			}
			if matches := scriptRegex.FindStringSubmatch(line); len(matches) > 1 {
				result.Imports = append(result.Imports, Import{
					Name:   extractNameFromPath(matches[1]),
					Source: matches[1],
					Line:   i + 1,
				})
			}
			if matches := imgRegex.FindStringSubmatch(line); len(matches) > 1 {
				result.Imports = append(result.Imports, Import{
					Name:   extractNameFromPath(matches[1]),
					Source: matches[1],
					Line:   i + 1,
				})
			}
		}
	}
}


func (p *UnifiedParser) extractCalls(code string, filePath string, result *ParseResult) {
	lines := strings.Split(code, "\n")

	callRegex := regexp.MustCompile(`\b([a-zA-Z_]\w*)\s*\(`)
	methodRegex := regexp.MustCompile(`\b(\w+)\.(\w+)\s*\(`)

	skipCalls := map[string]bool{
		"if": true, "for": true, "while": true, "switch": true,
		"return": true, "func": true, "function": true, "class": true,
		"import": true, "require": true, "use": true, "def": true,
		"var": true, "let": true, "const": true, "type": true,
		"struct": true, "interface": true, "enum": true, "trait": true,
		"new": true, "make": true, "len": true, "cap": true,
	}

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") {
			continue
		}

		if matches := methodRegex.FindAllStringSubmatch(line, -1); matches != nil {
			for _, match := range matches {
				if len(match) > 2 {
					obj := match[1]
					method := match[2]
					if skipCalls[obj] || skipCalls[method] {
						continue
					}
					result.Edges = append(result.Edges, Edge{
						ID:     fmt.Sprintf("%s:calls:%s.%s:%d", filePath, obj, method, i+1),
						Source: filePath,
						Target: obj + "." + method,
						Type:   EdgeTypeCalls,
						Line:   i + 1,
					})
				}
			}
		}

		if matches := callRegex.FindAllStringSubmatch(line, -1); matches != nil {
			for _, match := range matches {
				if len(match) > 1 {
					name := match[1]
					if skipCalls[name] {
						continue
					}
					// Skip if it's part of a method call (already handled)
					if i > 0 {
						prevChar := ' '
						idx := strings.Index(line, name)
						if idx > 0 {
							prevChar = rune(line[idx-1])
						}
						if prevChar == '.' {
							continue
						}
					}
					result.Edges = append(result.Edges, Edge{
						ID:     fmt.Sprintf("%s:calls:%s:%d", filePath, name, i+1),
						Source: filePath,
						Target: name,
						Type:   EdgeTypeCalls,
						Line:   i + 1,
					})
				}
			}
		}
	}
}


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

	controlFlow := regexp.MustCompile(`\b(if|else|for|while|switch|case|catch|&&|\|\|)\b`)
	matches := controlFlow.FindAllString(line, -1)
	metrics.Cyclomatic += len(matches)
	metrics.Cognitive = metrics.Cyclomatic

	return metrics
}



func isExported(name string) bool {
	if len(name) == 0 {
		return false
	}
	return name[0] >= 'A' && name[0] <= 'Z'
}


func extractNameFromPath(path string) string {
	// Handle Rust paths (use std::collections::HashMap)
	if strings.Contains(path, "::") {
		parts := strings.Split(path, "::")
		return parts[len(parts)-1]
	}
	// Handle PHP paths (use App\Models\User)
	if strings.Contains(path, "\\") {
		parts := strings.Split(path, "\\")
		return parts[len(parts)-1]
	}
	parts := strings.Split(path, "/")
	name := parts[len(parts)-1]
	name = strings.TrimSuffix(name, ".go")
	name = strings.TrimSuffix(name, ".js")
	name = strings.TrimSuffix(name, ".ts")
	name = strings.TrimSuffix(name, ".py")
	name = strings.TrimSuffix(name, ".jsx")
	name = strings.TrimSuffix(name, ".tsx")
	return name
}

// RegisterUnifiedParsers registers regex-based parsers for all supported languages
// Tree-sitter parsers take precedence when available (CGO enabled)

