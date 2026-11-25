package parser

import (
	"fmt"
	"regexp"
	"strings"
)

func (p *UnifiedParser) parseHTML(code string, filePath string, result *ParseResult) {
	lines := strings.Split(code, "\n")

	tagRegex := regexp.MustCompile(`<(\w+)(?:\s[^>]*)?>`)
	titleRegex := regexp.MustCompile(`<title>([^<]+)</title>`)
	headingRegex := regexp.MustCompile(`<h([1-6])[^>]*>([^<]+)</h[1-6]>`)

	for i, line := range lines {
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

		if matches := tagRegex.FindStringSubmatch(line); len(matches) > 1 {
			tag := matches[1]
			switch tag {
			case "script", "style", "link", "meta", "title",
				"h1", "h2", "h3", "h4", "h5", "h6",
				"div", "section", "nav", "header", "footer", "main",
				"article", "aside", "form", "table":
			}
		}
	}
}


func (p *UnifiedParser) parseCSS(code string, filePath string, result *ParseResult) {
	lines := strings.Split(code, "\n")

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


func (p *UnifiedParser) parseYAML(code string, filePath string, result *ParseResult) {
	lines := strings.Split(code, "\n")

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


