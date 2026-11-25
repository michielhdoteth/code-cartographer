package parser

import (
	"fmt"
	"regexp"
	"strings"
)

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


func (p *UnifiedParser) parseC(code string, filePath string, result *ParseResult) {
	lines := strings.Split(code, "\n")

	funcRegex := regexp.MustCompile(`^\s*(?:static\s+)?(?:inline\s+)?(?:const\s+)?(?:\w+\s*[*]+)\s*(\w+)\s*\(`)
	typedefRegex := regexp.MustCompile(`^\s*typedef\s+struct\s*(?:\w+)?\s*\{`)
	enumRegex := regexp.MustCompile(`^\s*enum\s+(\w+)`)
	defineRegex := regexp.MustCompile(`^\s*#define\s+(\w+)`)

	for i, line := range lines {
		// Skip preprocessor directives for function matching
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

		if matches := typedefRegex.FindStringSubmatch(line); len(matches) > 0 {
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


