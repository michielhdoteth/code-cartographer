package parser

import (
	"fmt"
	"regexp"
	"strings"
)

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


func (p *UnifiedParser) parseShell(code string, filePath string, result *ParseResult) {
	lines := strings.Split(code, "\n")

	funcRegex := regexp.MustCompile(`^\s*(?:function\s+)?(\w+)\s*\(\s*\)`)
	variableRegex := regexp.MustCompile(`^([A-Z_][A-Z0-9_]*)=`)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		if matches := funcRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
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


