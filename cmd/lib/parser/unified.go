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
	case LanguageJava:
		p.parseJava(code, filePath, result)
	case LanguageC:
		p.parseC(code, filePath, result)
	case LanguageCpp:
		p.parseCpp(code, filePath, result)
	case LanguageCSharp:
		p.parseCSharp(code, filePath, result)
	case LanguageRuby:
		p.parseRuby(code, filePath, result)
	case LanguagePHP:
		p.parsePHP(code, filePath, result)
	case LanguageRust:
		p.parseRust(code, filePath, result)
	case LanguageSwift:
		p.parseSwift(code, filePath, result)
	case LanguageKotlin:
		p.parseKotlin(code, filePath, result)
	case LanguageScala:
		p.parseScala(code, filePath, result)
	case LanguageShell:
		p.parseShell(code, filePath, result)
	case LanguageHTML:
		p.parseHTML(code, filePath, result)
	case LanguageCSS:
		p.parseCSS(code, filePath, result)
	case LanguageSQL:
		p.parseSQL(code, filePath, result)
	case LanguageYAML:
		p.parseYAML(code, filePath, result)
	case LanguageJSON:
		p.parseJSON(code, filePath, result)
	case LanguageTOML:
		p.parseTOML(code, filePath, result)
	case LanguageMarkdown:
		p.parseMarkdown(code, filePath, result)
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

	funcRegex := regexp.MustCompile(`^\s*func\s+(?:\([^)]+\)\s+)?(\w+)\s*\(`)
	typeRegex := regexp.MustCompile(`^\s*type\s+(\w+)\s+(struct|interface)\s*`)
	constRegex := regexp.MustCompile(`^\s*(?:const|var)\s+(\w+)`)

	for i, line := range lines {
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

	funcRegex := regexp.MustCompile(`^\s*(?:async\s+)?function\s+(\w+)\s*\(`)
	arrowRegex := regexp.MustCompile(`^\s*(?:const|let|var)\s+(\w+)\s*=\s*(?:async\s+)?\([^)]*\)\s*=>`)
	classRegex := regexp.MustCompile(`^\s*class\s+(\w+)`)
	methodRegex := regexp.MustCompile(`^\s{2,}(\w+)\s*\([^)]*\)\s*\{`)
	exportRegex := regexp.MustCompile(`^\s*export\s+(?:default\s+)?(?:const|let|var|function|class)\s+(\w+)`)

	for i, line := range lines {
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

	funcRegex := regexp.MustCompile(`^\s*(?:async\s+)?def\s+(\w+)\s*\(`)
	classRegex := regexp.MustCompile(`^\s*class\s+(\w+)`)
	decoratorRegex := regexp.MustCompile(`^\s*@(\w+)`)
	constRegex := regexp.MustCompile(`^\s*([A-Z][A-Z0-9_]*)\s*=`)

	for i, line := range lines {
		indent := len(line) - len(strings.TrimLeft(line, " \t"))

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

// parseJava parses Java source code
func (p *UnifiedParser) parseJava(code string, filePath string, result *ParseResult) {
	lines := strings.Split(code, "\n")

	classRegex := regexp.MustCompile(`^\s*(?:public\s+)?(?:abstract\s+)?class\s+(\w+)`)
	interfaceRegex := regexp.MustCompile(`^\s*(?:public\s+)?interface\s+(\w+)`)
	enumRegex := regexp.MustCompile(`^\s*(?:public\s+)?enum\s+(\w+)`)
	methodRegex := regexp.MustCompile(`^\s*(?:public|private|protected)\s+(?:static\s+)?(?:\w+(?:<[^>]+>)?)\s+(\w+)\s*\(`)
	constructorRegex := regexp.MustCompile(`^\s*(?:public|private|protected)\s+(\w+)\s*\(`)

	for i, line := range lines {
		if matches := classRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:class:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeClass,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
				Exported:  strings.Contains(line, "public"),
			})
			if strings.Contains(line, "public") {
				result.Exports = append(result.Exports, Export{
					Name:     name,
					File:     filePath,
					Type:     NodeTypeClass,
					Line:     i + 1,
					Exported: true,
				})
			}
		}

		if matches := interfaceRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:interface:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeInterface,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
				Exported:  strings.Contains(line, "public"),
			})
			if strings.Contains(line, "public") {
				result.Exports = append(result.Exports, Export{
					Name:     name,
					File:     filePath,
					Type:     NodeTypeInterface,
					Line:     i + 1,
					Exported: true,
				})
			}
		}

		if matches := enumRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:enum:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeEnum,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
				Exported:  strings.Contains(line, "public"),
			})
			if strings.Contains(line, "public") {
				result.Exports = append(result.Exports, Export{
					Name:     name,
					File:     filePath,
					Type:     NodeTypeEnum,
					Line:     i + 1,
					Exported: true,
				})
			}
		}

		if matches := methodRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			// Skip class name (constructors)
			if strings.Contains(line, "(") {
				result.Nodes = append(result.Nodes, Node{
					ID:        fmt.Sprintf("%s:method:%s", filePath, name),
					Name:      name,
					Type:      NodeTypeMethod,
					File:      filePath,
					StartLine: i + 1,
					EndLine:   i + 1,
					Exported:  strings.Contains(line, "public"),
					Metrics:   p.calculateLineMetrics(code, i),
				})
			}
		}

		if matches := constructorRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			// Only if followed by ( and the name looks like a constructor
			if strings.Contains(line, "(") && !strings.Contains(line, "=") {
				result.Nodes = append(result.Nodes, Node{
					ID:        fmt.Sprintf("%s:constructor:%s", filePath, name),
					Name:      name,
					Type:      NodeTypeMethod,
					File:      filePath,
					StartLine: i + 1,
					EndLine:   i + 1,
					Exported:  strings.Contains(line, "public"),
				})
			}
		}
	}
}

// parseC parses C source code
func (p *UnifiedParser) parseC(code string, filePath string, result *ParseResult) {
	lines := strings.Split(code, "\n")

	// C function: return_type name(args) { or ;
	funcRegex := regexp.MustCompile(`^\s*(?:static\s+)?(?:inline\s+)?(?:const\s+)?(?:\w+\s*[*]+)\s*(\w+)\s*\(`)
	// Typedef struct: typedef struct { ... } Name;
	typedefRegex := regexp.MustCompile(`^\s*typedef\s+struct\s*(?:\w+)?\s*\{`)
	// Enum: enum Name { or enum Name {
	enumRegex := regexp.MustCompile(`^\s*enum\s+(\w+)`)
	// Define: #define NAME
	defineRegex := regexp.MustCompile(`^\s*#define\s+(\w+)`)

	for i, line := range lines {
		// Skip preprocessor directives for function matching
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}

		if matches := funcRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			// Skip keywords
			if name == "if" || name == "for" || name == "while" || name == "switch" || name == "return" {
				continue
			}
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

		if matches := typedefRegex.FindStringSubmatch(line); len(matches) > 0 {
			// Look for the name after the closing brace
			for j := i; j < len(lines) && j < i+20; j++ {
				closingRegex := regexp.MustCompile(`\}\s*(\w+)\s*;`)
				if m := closingRegex.FindStringSubmatch(lines[j]); len(m) > 1 {
					result.Nodes = append(result.Nodes, Node{
						ID:        fmt.Sprintf("%s:struct:%s", filePath, m[1]),
						Name:      m[1],
						Type:      NodeTypeStruct,
						File:      filePath,
						StartLine: i + 1,
						EndLine:   j + 1,
					})
					break
				}
			}
		}

		if matches := enumRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:enum:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeEnum,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
			})
		}

		if matches := defineRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			if name[0] >= 'A' && name[0] <= 'Z' {
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

// parseCpp parses C++ source code
func (p *UnifiedParser) parseCpp(code string, filePath string, result *ParseResult) {
	// C++ has everything C has plus classes, namespaces, templates
	lines := strings.Split(code, "\n")

	funcRegex := regexp.MustCompile(`^\s*(?:static\s+)?(?:virtual\s+)?(?:const\s+)?(?:\w+\s*[*&]+)\s*(\w+)\s*\(`)
	classRegex := regexp.MustCompile(`^\s*(?:class|struct)\s+(\w+)\s*(?:\{|:)`)
	namespaceRegex := regexp.MustCompile(`^\s*namespace\s+(\w+)`)
	enumRegex := regexp.MustCompile(`^\s*enum\s+(?:class\s+)?(\w+)`)
	methodRegex := regexp.MustCompile(`^\s*(?:virtual\s+)?(?:\w+\s*[*&]+\s+)?(\w+::\w+)\s*\(`)
	defineRegex := regexp.MustCompile(`^\s*#define\s+(\w+)`)

	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}

		if matches := funcRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			if name == "if" || name == "for" || name == "while" || name == "switch" || name == "return" {
				continue
			}
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

		if matches := classRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:class:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeClass,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
			})
		}

		if matches := namespaceRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:module:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeModule,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
			})
		}

		if matches := enumRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:enum:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeEnum,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
			})
		}

		if matches := methodRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			parts := strings.Split(name, "::")
			if len(parts) == 2 {
				result.Nodes = append(result.Nodes, Node{
					ID:        fmt.Sprintf("%s:method:%s", filePath, parts[1]),
					Name:      parts[1],
					Type:      NodeTypeMethod,
					File:      filePath,
					StartLine: i + 1,
					EndLine:   i + 1,
					Metrics:   p.calculateLineMetrics(code, i),
				})
			}
		}

		if matches := defineRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			if name[0] >= 'A' && name[0] <= 'Z' {
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

// parseCSharp parses C# source code
func (p *UnifiedParser) parseCSharp(code string, filePath string, result *ParseResult) {
	lines := strings.Split(code, "\n")

	classRegex := regexp.MustCompile(`^\s*(?:public\s+)?(?:abstract\s+)?(?:sealed\s+)?class\s+(\w+)`)
	interfaceRegex := regexp.MustCompile(`^\s*(?:public\s+)?interface\s+(\w+)`)
	enumRegex := regexp.MustCompile(`^\s*(?:public\s+)?enum\s+(\w+)`)
	structRegex := regexp.MustCompile(`^\s*(?:public\s+)?struct\s+(\w+)`)
	methodRegex := regexp.MustCompile(`^\s*(?:public|private|protected)\s+(?:static\s+)?(?:async\s+)?(?:\w+(?:<[^>]+>)?)\s+(\w+)\s*\(`)
	namespaceRegex := regexp.MustCompile(`^\s*namespace\s+([\w.]+)`)

	for i, line := range lines {
		if matches := classRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:class:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeClass,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
				Exported:  strings.Contains(line, "public"),
			})
			if strings.Contains(line, "public") {
				result.Exports = append(result.Exports, Export{
					Name:     name,
					File:     filePath,
					Type:     NodeTypeClass,
					Line:     i + 1,
					Exported: true,
				})
			}
		}

		if matches := interfaceRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:interface:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeInterface,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
				Exported:  strings.Contains(line, "public"),
			})
			if strings.Contains(line, "public") {
				result.Exports = append(result.Exports, Export{
					Name:     name,
					File:     filePath,
					Type:     NodeTypeInterface,
					Line:     i + 1,
					Exported: true,
				})
			}
		}

		if matches := enumRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:enum:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeEnum,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
				Exported:  strings.Contains(line, "public"),
			})
			if strings.Contains(line, "public") {
				result.Exports = append(result.Exports, Export{
					Name:     name,
					File:     filePath,
					Type:     NodeTypeEnum,
					Line:     i + 1,
					Exported: true,
				})
			}
		}

		if matches := structRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:struct:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeStruct,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
				Exported:  strings.Contains(line, "public"),
			})
			if strings.Contains(line, "public") {
				result.Exports = append(result.Exports, Export{
					Name:     name,
					File:     filePath,
					Type:     NodeTypeStruct,
					Line:     i + 1,
					Exported: true,
				})
			}
		}

		if matches := namespaceRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:module:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeModule,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
			})
		}

		if matches := methodRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:method:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeMethod,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
				Exported:  strings.Contains(line, "public"),
				Metrics:   p.calculateLineMetrics(code, i),
			})
		}
	}
}

// parseRuby parses Ruby source code
func (p *UnifiedParser) parseRuby(code string, filePath string, result *ParseResult) {
	lines := strings.Split(code, "\n")

	classRegex := regexp.MustCompile(`^\s*class\s+(\w+)`)
	moduleRegex := regexp.MustCompile(`^\s*module\s+(\w+)`)
	defRegex := regexp.MustCompile(`^\s*def\s+(?:self\.)?(\w+)`)
	attrRegex := regexp.MustCompile(`^\s*attr_(?:reader|writer|accessor)\s+(:?\w+)`)

	for i, line := range lines {
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

		if matches := moduleRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:module:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeModule,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
				Exported:  name[0] >= 'A' && name[0] <= 'Z',
			})
			if name[0] >= 'A' && name[0] <= 'Z' {
				result.Exports = append(result.Exports, Export{
					Name:     name,
					File:     filePath,
					Type:     NodeTypeModule,
					Line:     i + 1,
					Exported: true,
				})
			}
		}

		if matches := defRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			nodeType := NodeTypeMethod
			if strings.HasPrefix(line, "  def") || strings.HasPrefix(line, "\tdef") {
				nodeType = NodeTypeMethod
			}
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:method:%s", filePath, name),
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

		if matches := attrRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			name = strings.TrimPrefix(name, ":")
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:property:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeProperty,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
			})
		}
	}
}

// parsePHP parses PHP source code
func (p *UnifiedParser) parsePHP(code string, filePath string, result *ParseResult) {
	lines := strings.Split(code, "\n")

	classRegex := regexp.MustCompile(`^\s*(?:abstract\s+)?class\s+(\w+)`)
	interfaceRegex := regexp.MustCompile(`^\s*interface\s+(\w+)`)
	traitRegex := regexp.MustCompile(`^\s*trait\s+(\w+)`)
	methodRegex := regexp.MustCompile(`^\s*(?:public|private|protected)\s+(?:static\s+)?function\s+(\w+)\s*\(`)
	namespaceRegex := regexp.MustCompile(`^\s*namespace\s+([\w\\]+)`)

	for i, line := range lines {
		if matches := classRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:class:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeClass,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
				Exported:  strings.Contains(line, "public"),
			})
			if strings.Contains(line, "public") {
				result.Exports = append(result.Exports, Export{
					Name:     name,
					File:     filePath,
					Type:     NodeTypeClass,
					Line:     i + 1,
					Exported: true,
				})
			}
		}

		if matches := interfaceRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:interface:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeInterface,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
			})
			result.Exports = append(result.Exports, Export{
				Name:     name,
				File:     filePath,
				Type:     NodeTypeInterface,
				Line:     i + 1,
				Exported: true,
			})
		}

		if matches := traitRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:trait:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeType,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
			})
			result.Exports = append(result.Exports, Export{
				Name:     name,
				File:     filePath,
				Type:     NodeTypeType,
				Line:     i + 1,
				Exported: true,
			})
		}

		if matches := namespaceRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:module:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeModule,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
			})
		}

		if matches := methodRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:method:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeMethod,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
				Exported:  strings.Contains(line, "public"),
				Metrics:   p.calculateLineMetrics(code, i),
			})
		}
	}
}

// parseRust parses Rust source code
func (p *UnifiedParser) parseRust(code string, filePath string, result *ParseResult) {
	lines := strings.Split(code, "\n")

	fnRegex := regexp.MustCompile(`^\s*(?:pub\s+)?(?:unsafe\s+)?fn\s+(\w+)`)
	structRegex := regexp.MustCompile(`^\s*(?:pub\s+)?struct\s+(\w+)`)
	traitRegex := regexp.MustCompile(`^\s*(?:pub\s+)?trait\s+(\w+)`)
	enumRegex := regexp.MustCompile(`^\s*(?:pub\s+)?enum\s+(\w+)`)
	implRegex := regexp.MustCompile(`^\s*(?:pub\s+)?impl(?:<[^>]+>)?\s+(\w+)`)
	typeRegex := regexp.MustCompile(`^\s*(?:pub\s+)?type\s+(\w+)`)
	constRegex := regexp.MustCompile(`^\s*(?:pub\s+)?const\s+(\w+)`)
	modRegex := regexp.MustCompile(`^\s*(?:pub\s+)?mod\s+(\w+)`)
	automodRegex := regexp.MustCompile(`^\s*mod\s+(\w+)\s*;`)

	for i, line := range lines {
		if matches := fnRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			nodeType := NodeTypeFunction
			if strings.Contains(line, "impl") || (i > 0 && strings.Contains(lines[max(0, i-1)], "impl")) {
				nodeType = NodeTypeMethod
			}
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:func:%s", filePath, name),
				Name:      name,
				Type:      nodeType,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
				Exported:  strings.Contains(line, "pub"),
				Metrics:   p.calculateLineMetrics(code, i),
			})
			if strings.Contains(line, "pub") {
				result.Exports = append(result.Exports, Export{
					Name:     name,
					File:     filePath,
					Type:     nodeType,
					Line:     i + 1,
					Exported: true,
				})
			}
		}

		if matches := structRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:struct:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeStruct,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
				Exported:  strings.Contains(line, "pub"),
			})
			if strings.Contains(line, "pub") {
				result.Exports = append(result.Exports, Export{
					Name:     name,
					File:     filePath,
					Type:     NodeTypeStruct,
					Line:     i + 1,
					Exported: true,
				})
			}
		}

		if matches := traitRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:trait:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeType,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
				Exported:  strings.Contains(line, "pub"),
			})
			if strings.Contains(line, "pub") {
				result.Exports = append(result.Exports, Export{
					Name:     name,
					File:     filePath,
					Type:     NodeTypeType,
					Line:     i + 1,
					Exported: true,
				})
			}
		}

		if matches := enumRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:enum:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeEnum,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
				Exported:  strings.Contains(line, "pub"),
			})
			if strings.Contains(line, "pub") {
				result.Exports = append(result.Exports, Export{
					Name:     name,
					File:     filePath,
					Type:     NodeTypeEnum,
					Line:     i + 1,
					Exported: true,
				})
			}
		}

		if matches := implRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:impl:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeType,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
			})
		}

		if matches := typeRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:type:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeType,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
				Exported:  strings.Contains(line, "pub"),
			})
			if strings.Contains(line, "pub") {
				result.Exports = append(result.Exports, Export{
					Name:     name,
					File:     filePath,
					Type:     NodeTypeType,
					Line:     i + 1,
					Exported: true,
				})
			}
		}

		if matches := constRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:const:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeConstant,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
				Exported:  strings.Contains(line, "pub"),
			})
			if strings.Contains(line, "pub") {
				result.Exports = append(result.Exports, Export{
					Name:     name,
					File:     filePath,
					Type:     NodeTypeConstant,
					Line:     i + 1,
					Exported: true,
				})
			}
		}

		if matches := modRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:module:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeModule,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
				Exported:  strings.Contains(line, "pub"),
			})
		}

		if matches := automodRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:module:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeModule,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
			})
		}
	}
}

// parseSwift parses Swift source code
func (p *UnifiedParser) parseSwift(code string, filePath string, result *ParseResult) {
	lines := strings.Split(code, "\n")

	funcRegex := regexp.MustCompile(`^\s*(?:public\s+)?(?:private\s+)?(?:fileprivate\s+)?(?:internal\s+)?(?:open\s+)?func\s+(\w+)`)
	classRegex := regexp.MustCompile(`^\s*(?:public\s+)?class\s+(\w+)`)
	structRegex := regexp.MustCompile(`^\s*(?:public\s+)?struct\s+(\w+)`)
	enumRegex := regexp.MustCompile(`^\s*(?:public\s+)?enum\s+(\w+)`)
	protocolRegex := regexp.MustCompile(`^\s*(?:public\s+)?protocol\s+(\w+)`)
	varRegex := regexp.MustCompile(`^\s*(?:public\s+)?(?:let|var)\s+(\w+)`)

	for i, line := range lines {
		if matches := funcRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:func:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeFunction,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
				Exported:  strings.Contains(line, "public"),
				Metrics:   p.calculateLineMetrics(code, i),
			})
			if strings.Contains(line, "public") {
				result.Exports = append(result.Exports, Export{
					Name:     name,
					File:     filePath,
					Type:     NodeTypeFunction,
					Line:     i + 1,
					Exported: true,
				})
			}
		}

		if matches := classRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:class:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeClass,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
				Exported:  strings.Contains(line, "public"),
			})
			if strings.Contains(line, "public") {
				result.Exports = append(result.Exports, Export{
					Name:     name,
					File:     filePath,
					Type:     NodeTypeClass,
					Line:     i + 1,
					Exported: true,
				})
			}
		}

		if matches := structRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:struct:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeStruct,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
				Exported:  strings.Contains(line, "public"),
			})
			if strings.Contains(line, "public") {
				result.Exports = append(result.Exports, Export{
					Name:     name,
					File:     filePath,
					Type:     NodeTypeStruct,
					Line:     i + 1,
					Exported: true,
				})
			}
		}

		if matches := enumRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:enum:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeEnum,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
				Exported:  strings.Contains(line, "public"),
			})
			if strings.Contains(line, "public") {
				result.Exports = append(result.Exports, Export{
					Name:     name,
					File:     filePath,
					Type:     NodeTypeEnum,
					Line:     i + 1,
					Exported: true,
				})
			}
		}

		if matches := protocolRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:interface:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeInterface,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
				Exported:  strings.Contains(line, "public"),
			})
			if strings.Contains(line, "public") {
				result.Exports = append(result.Exports, Export{
					Name:     name,
					File:     filePath,
					Type:     NodeTypeInterface,
					Line:     i + 1,
					Exported: true,
				})
			}
		}

		if matches := varRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			if name[0] >= 'A' && name[0] <= 'Z' && strings.Contains(line, "public") {
				result.Exports = append(result.Exports, Export{
					Name:     name,
					File:     filePath,
					Type:     NodeTypeVariable,
					Line:     i + 1,
					Exported: true,
				})
			}
		}
	}
}

// parseKotlin parses Kotlin source code
func (p *UnifiedParser) parseKotlin(code string, filePath string, result *ParseResult) {
	lines := strings.Split(code, "\n")

	funRegex := regexp.MustCompile(`^\s*(?:fun\s+|suspend\s+fun\s+)(\w+)`)
	classRegex := regexp.MustCompile(`^\s*(?:data\s+)?(?:open\s+)?(?:abstract\s+)?class\s+(\w+)`)
	interfaceRegex := regexp.MustCompile(`^\s*interface\s+(\w+)`)
	objectRegex := regexp.MustCompile(`^\s*(?:data\s+)?object\s+(\w+)`)
	valRegex := regexp.MustCompile(`^\s*(?:val|var)\s+(\w+)`)
	packageRegex := regexp.MustCompile(`^\s*package\s+([\w.]+)`)

	for i, line := range lines {
		if matches := funRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:func:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeFunction,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
				Exported:  !strings.HasPrefix(strings.TrimSpace(line), "private"),
				Metrics:   p.calculateLineMetrics(code, i),
			})
			if !strings.HasPrefix(strings.TrimSpace(line), "private") {
				result.Exports = append(result.Exports, Export{
					Name:     name,
					File:     filePath,
					Type:     NodeTypeFunction,
					Line:     i + 1,
					Exported: true,
				})
			}
		}

		if matches := classRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:class:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeClass,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
				Exported:  !strings.HasPrefix(strings.TrimSpace(line), "private"),
			})
			if !strings.HasPrefix(strings.TrimSpace(line), "private") {
				result.Exports = append(result.Exports, Export{
					Name:     name,
					File:     filePath,
					Type:     NodeTypeClass,
					Line:     i + 1,
					Exported: true,
				})
			}
		}

		if matches := interfaceRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:interface:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeInterface,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
				Exported:  !strings.HasPrefix(strings.TrimSpace(line), "private"),
			})
			if !strings.HasPrefix(strings.TrimSpace(line), "private") {
				result.Exports = append(result.Exports, Export{
					Name:     name,
					File:     filePath,
					Type:     NodeTypeInterface,
					Line:     i + 1,
					Exported: true,
				})
			}
		}

		if matches := objectRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:class:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeClass,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
				Exported:  !strings.HasPrefix(strings.TrimSpace(line), "private"),
			})
			if !strings.HasPrefix(strings.TrimSpace(line), "private") {
				result.Exports = append(result.Exports, Export{
					Name:     name,
					File:     filePath,
					Type:     NodeTypeClass,
					Line:     i + 1,
					Exported: true,
				})
			}
		}

		if matches := packageRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:module:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeModule,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
			})
		}

		_ = valRegex // Used in future for exported vals
	}
}

// parseScala parses Scala source code
func (p *UnifiedParser) parseScala(code string, filePath string, result *ParseResult) {
	lines := strings.Split(code, "\n")

	defRegex := regexp.MustCompile(`^\s*(?:override\s+)?def\s+(\w+)`)
	classRegex := regexp.MustCompile(`^\s*(?:case\s+)?class\s+(\w+)`)
	traitRegex := regexp.MustCompile(`^\s*trait\s+(\w+)`)
	objectRegex := regexp.MustCompile(`^\s*object\s+(\w+)`)
	valRegex := regexp.MustCompile(`^\s*(?:val|var)\s+(\w+)`)
	packageRegex := regexp.MustCompile(`^\s*package\s+([\w.]+)`)

	for i, line := range lines {
		if matches := defRegex.FindStringSubmatch(line); len(matches) > 1 {
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

		if matches := classRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:class:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeClass,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
			})
			result.Exports = append(result.Exports, Export{
				Name:     name,
				File:     filePath,
				Type:     NodeTypeClass,
				Line:     i + 1,
				Exported: true,
			})
		}

		if matches := traitRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:trait:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeType,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
			})
			result.Exports = append(result.Exports, Export{
				Name:     name,
				File:     filePath,
				Type:     NodeTypeType,
				Line:     i + 1,
				Exported: true,
			})
		}

		if matches := objectRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:object:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeClass,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
			})
			result.Exports = append(result.Exports, Export{
				Name:     name,
				File:     filePath,
				Type:     NodeTypeClass,
				Line:     i + 1,
				Exported: true,
			})
		}

		if matches := valRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			if name[0] >= 'A' && name[0] <= 'Z' {
				result.Exports = append(result.Exports, Export{
					Name:     name,
					File:     filePath,
					Type:     NodeTypeVariable,
					Line:     i + 1,
					Exported: true,
				})
			}
		}

		if matches := packageRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:module:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeModule,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
			})
		}
	}
}

// parseShell parses Shell/Bash source code
func (p *UnifiedParser) parseShell(code string, filePath string, result *ParseResult) {
	lines := strings.Split(code, "\n")

	// function name() { or name() {
	funcRegex := regexp.MustCompile(`^\s*(?:function\s+)?(\w+)\s*\(\s*\)`)
	variableRegex := regexp.MustCompile(`^([A-Z_][A-Z0-9_]*)=`)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		if matches := funcRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			// Skip if/while/for/case keywords
			if name == "if" || name == "while" || name == "for" || name == "case" || name == "until" {
				continue
			}
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:func:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeFunction,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
				Metrics:   p.calculateLineMetrics(code, i),
			})
			if name[0] >= 'A' && name[0] <= 'Z' {
				result.Exports = append(result.Exports, Export{
					Name:     name,
					File:     filePath,
					Type:     NodeTypeFunction,
					Line:     i + 1,
					Exported: true,
				})
			}
		}

		if matches := variableRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:variable:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeVariable,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
			})
		}
	}
}

// parseHTML parses HTML source code
func (p *UnifiedParser) parseHTML(code string, filePath string, result *ParseResult) {
	lines := strings.Split(code, "\n")

	// Match HTML tags for structural elements
	tagRegex := regexp.MustCompile(`<(\w+)(?:\s[^>]*)?>`)
	titleRegex := regexp.MustCompile(`<title>([^<]+)</title>`)
	headingRegex := regexp.MustCompile(`<h([1-6])[^>]*>([^<]+)</h[1-6]>`)

	for i, line := range lines {
		// Extract title
		if matches := titleRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := strings.TrimSpace(matches[1])
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:title:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeModule,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
			})
		}

		// Extract headings
		if matches := headingRegex.FindStringSubmatch(line); len(matches) > 2 {
			name := strings.TrimSpace(matches[2])
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:heading:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeConstant,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
			})
		}

		// Extract structural tags
		if matches := tagRegex.FindStringSubmatch(line); len(matches) > 1 {
			tag := matches[1]
			switch tag {
			case "script", "style", "link", "meta", "title",
				"h1", "h2", "h3", "h4", "h5", "h6",
				"div", "section", "nav", "header", "footer", "main",
				"article", "aside", "form", "table":
				// These are structural elements worth tracking
			}
		}
	}
}

// parseCSS parses CSS source code
func (p *UnifiedParser) parseCSS(code string, filePath string, result *ParseResult) {
	lines := strings.Split(code, "\n")

	// Match CSS selectors at the start of rules
	selectorRegex := regexp.MustCompile(`^([.#:\w][\w.:#>-]*)\s*\{`)
	mediaRegex := regexp.MustCompile(`^@media\s+`)
	keyframeRegex := regexp.MustCompile(`^@keyframes\s+(\w+)`)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if matches := keyframeRegex.FindStringSubmatch(trimmed); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:keyframe:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeType,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
			})
		}

		if mediaRegex.MatchString(trimmed) {
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:media:%d", filePath, i+1),
				Name:      "media-query",
				Type:      NodeTypeType,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
			})
		}

		if matches := selectorRegex.FindStringSubmatch(trimmed); len(matches) > 1 {
			selector := strings.TrimSpace(matches[1])
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:selector:%s", filePath, selector),
				Name:      selector,
				Type:      NodeTypeType,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
			})
		}
	}
}

// parseSQL parses SQL source code
func (p *UnifiedParser) parseSQL(code string, filePath string, result *ParseResult) {
	lines := strings.Split(code, "\n")

	createTableRegex := regexp.MustCompile(`(?i)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?(\w+)`)
	createIndexRegex := regexp.MustCompile(`(?i)CREATE\s+(?:UNIQUE\s+)?INDEX\s+(?:IF\s+NOT\s+EXISTS\s+)?(\w+)`)
	createViewRegex := regexp.MustCompile(`(?i)CREATE\s+VIEW\s+(?:IF\s+NOT\s+EXISTS\s+)?(\w+)`)
	createFuncRegex := regexp.MustCompile(`(?i)CREATE\s+(?:OR\s+REPLACE\s+)?(?:FUNCTION|PROCEDURE)\s+(\w+)`)
	dropRegex := regexp.MustCompile(`(?i)DROP\s+(?:TABLE|INDEX|VIEW|FUNCTION|PROCEDURE)\s+(?:IF\s+EXISTS\s+)?(\w+)`)

	for i, line := range lines {
		if matches := createTableRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:table:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeStruct,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
			})
			result.Exports = append(result.Exports, Export{
				Name:     name,
				File:     filePath,
				Type:     NodeTypeStruct,
				Line:     i + 1,
				Exported: true,
			})
		}

		if matches := createIndexRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:index:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeType,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
			})
		}

		if matches := createViewRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:view:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeType,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
			})
			result.Exports = append(result.Exports, Export{
				Name:     name,
				File:     filePath,
				Type:     NodeTypeType,
				Line:     i + 1,
				Exported: true,
			})
		}

		if matches := createFuncRegex.FindStringSubmatch(line); len(matches) > 1 {
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
			result.Exports = append(result.Exports, Export{
				Name:     name,
				File:     filePath,
				Type:     NodeTypeFunction,
				Line:     i + 1,
				Exported: true,
			})
		}

		_ = dropRegex // tracked but not creating nodes for drops
	}
}

// parseYAML parses YAML source code
func (p *UnifiedParser) parseYAML(code string, filePath string, result *ParseResult) {
	lines := strings.Split(code, "\n")

	// Top-level keys (no leading whitespace or minimal)
	keyRegex := regexp.MustCompile(`^(\w[\w-]*)\s*:`)

	for i, line := range lines {
		if matches := keyRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:key:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeProperty,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
			})
		}
	}
}

// parseJSON parses JSON source code
func (p *UnifiedParser) parseJSON(code string, filePath string, result *ParseResult) {
	// For JSON, just count top-level keys
	keyRegex := regexp.MustCompile(`"(\w+)"\s*:`)
	matches := keyRegex.FindAllStringSubmatch(code, -1)

	seen := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 {
			name := match[1]
			if !seen[name] {
				seen[name] = true
				result.Nodes = append(result.Nodes, Node{
					ID:   fmt.Sprintf("%s:key:%s", filePath, name),
					Name: name,
					Type: NodeTypeProperty,
					File: filePath,
				})
			}
		}
	}
}

// parseTOML parses TOML source code
func (p *UnifiedParser) parseTOML(code string, filePath string, result *ParseResult) {
	lines := strings.Split(code, "\n")

	sectionRegex := regexp.MustCompile(`^\[([\w.]+)\]`)
	keyRegex := regexp.MustCompile(`^(\w[\w-]*)\s*=`)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if matches := sectionRegex.FindStringSubmatch(trimmed); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:section:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeModule,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
			})
		}

		if matches := keyRegex.FindStringSubmatch(trimmed); len(matches) > 1 {
			name := matches[1]
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:key:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeProperty,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
			})
		}
	}
}

// parseMarkdown parses Markdown source code
func (p *UnifiedParser) parseMarkdown(code string, filePath string, result *ParseResult) {
	lines := strings.Split(code, "\n")

	headingRegex := regexp.MustCompile(`^(#{1,6})\s+(.+)`)
	codeBlockRegex := regexp.MustCompile("^```(\\w*)")

	for i, line := range lines {
		if matches := headingRegex.FindStringSubmatch(line); len(matches) > 2 {
			level := len(matches[1])
			name := strings.TrimSpace(matches[2])
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:heading:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeConstant,
				File:      filePath,
				StartLine: i + 1,
				EndLine:   i + 1,
			})
			_ = level
		}

		if matches := codeBlockRegex.FindStringSubmatch(line); len(matches) > 1 {
			lang := matches[1]
			if lang != "" {
				result.Nodes = append(result.Nodes, Node{
					ID:        fmt.Sprintf("%s:codeblock:%s:%d", filePath, lang, i+1),
					Name:      lang,
					Type:      NodeTypeType,
					File:      filePath,
					StartLine: i + 1,
					EndLine:   i + 1,
				})
			}
		}
	}
}

// parseGeneric parses using generic patterns for unknown languages
func (p *UnifiedParser) parseGeneric(code string, filePath string, result *ParseResult) {
	lines := strings.Split(code, "\n")

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
		// Single-line: import "path"
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
	// Remove common extensions
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
func RegisterUnifiedParsers() {
	languages := []Language{
		LanguageGo,
		LanguageJavaScript,
		LanguageTypeScript,
		LanguagePython,
		LanguageJava,
		LanguageC,
		LanguageCpp,
		LanguageCSharp,
		LanguageRuby,
		LanguagePHP,
		LanguageRust,
		LanguageSwift,
		LanguageKotlin,
		LanguageScala,
		LanguageShell,
		LanguageHTML,
		LanguageCSS,
		LanguageSQL,
		LanguageYAML,
		LanguageJSON,
		LanguageTOML,
		LanguageMarkdown,
	}

	for _, lang := range languages {
		// Only register if no tree-sitter parser is already registered
		if GetParser(lang) == nil {
			RegisterParser(NewUnifiedParser(lang))
		}
	}
}

// init registers the parsers
func init() {
	// Register tree-sitter parsers first (they take precedence)
	RegisterTreeSitterParsers()
	// Then register regex-based parsers for languages without tree-sitter support
	RegisterUnifiedParsers()
}
