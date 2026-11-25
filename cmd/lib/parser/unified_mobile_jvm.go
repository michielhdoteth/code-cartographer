package parser

import (
	"fmt"
	"regexp"
	"strings"
)

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


