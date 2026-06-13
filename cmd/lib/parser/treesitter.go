//go:build cgo
// +build cgo

package parser

import (
	"context"
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/bash"
	"github.com/smacker/go-tree-sitter/c"
	"github.com/smacker/go-tree-sitter/cpp"
	"github.com/smacker/go-tree-sitter/csharp"
	"github.com/smacker/go-tree-sitter/css"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/html"
	"github.com/smacker/go-tree-sitter/java"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/kotlin"
	"github.com/smacker/go-tree-sitter/markdown"
	"github.com/smacker/go-tree-sitter/php"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/ruby"
	"github.com/smacker/go-tree-sitter/rust"
	"github.com/smacker/go-tree-sitter/scala"
	"github.com/smacker/go-tree-sitter/sql"
	"github.com/smacker/go-tree-sitter/swift"
	"github.com/smacker/go-tree-sitter/toml"
	treesitter_ts "github.com/smacker/go-tree-sitter/typescript/typescript"
	"github.com/smacker/go-tree-sitter/yaml"
)

// TreeSitterParser is a parser that uses tree-sitter for accurate AST parsing
type TreeSitterParser struct {
	language Language
	langPtr  *sitter.Language
}

// NewTreeSitterParser creates a new tree-sitter parser for the given language
func NewTreeSitterParser(lang Language) *TreeSitterParser {
	var langPtr *sitter.Language

	switch lang {
	case LanguageGo:
		langPtr = golang.GetLanguage()
	case LanguageJavaScript:
		langPtr = javascript.GetLanguage()
	case LanguageTypeScript:
		langPtr = treesitter_ts.GetLanguage()
	case LanguagePython:
		langPtr = python.GetLanguage()
	case LanguageJava:
		langPtr = java.GetLanguage()
	case LanguageC:
		langPtr = c.GetLanguage()
	case LanguageCpp:
		langPtr = cpp.GetLanguage()
	case LanguageCSharp:
		langPtr = csharp.GetLanguage()
	case LanguageRuby:
		langPtr = ruby.GetLanguage()
	case LanguagePHP:
		langPtr = php.GetLanguage()
	case LanguageRust:
		langPtr = rust.GetLanguage()
	case LanguageSwift:
		langPtr = swift.GetLanguage()
	case LanguageKotlin:
		langPtr = kotlin.GetLanguage()
	case LanguageScala:
		langPtr = scala.GetLanguage()
	case LanguageHTML:
		langPtr = html.GetLanguage()
	case LanguageCSS:
		langPtr = css.GetLanguage()
	case LanguageSQL:
		langPtr = sql.GetLanguage()
	case LanguageShell:
		langPtr = bash.GetLanguage()
	case LanguageYAML:
		langPtr = yaml.GetLanguage()
	case LanguageTOML:
		langPtr = toml.GetLanguage()
	case LanguageMarkdown:
		langPtr = markdown.GetLanguage()
	}

	if langPtr == nil {
		return nil
	}

	return &TreeSitterParser{
		language: lang,
		langPtr:  langPtr,
	}
}

// Language returns the language this parser handles
func (p *TreeSitterParser) Language() Language {
	return p.language
}

// SupportedExtensions returns file extensions this parser handles
func (p *TreeSitterParser) SupportedExtensions() []string {
	return GetLanguageConfig(p.language).Extensions
}

// Parse parses source code using tree-sitter and returns the AST
func (p *TreeSitterParser) Parse(code string, filePath string) (*ParseResult, error) {
	if p.langPtr == nil {
		return nil, fmt.Errorf("language not supported for tree-sitter: %s", p.language)
	}

	result := &ParseResult{
		FilePath: filePath,
		Language: p.language,
		Nodes:    []Node{},
		Edges:    []Edge{},
		Exports:  []Export{},
		Imports:  []Import{},
		Metrics:  Metrics{},
	}

	parser := sitter.NewParser()
	defer parser.Close()
	parser.SetLanguage(p.langPtr)

	tree, err := parser.ParseCtx(context.Background(), nil, []byte(code))
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", filePath, err)
	}
	defer tree.Close()

	root := tree.RootNode()

	switch p.language {
	case LanguageGo:
		p.parseGo(root, []byte(code), filePath, result)
	case LanguageJavaScript:
		p.parseJavaScript(root, []byte(code), filePath, result)
	case LanguageTypeScript:
		p.parseTypeScript(root, []byte(code), filePath, result)
	default:
		p.walkGenericTree(root, []byte(code), filePath, result)
	}

	// Calculate metrics
	result.CalculateMetrics()

	return result, nil
}

// parseGo extracts nodes from Go source code
func (p *TreeSitterParser) parseGo(root *sitter.Node, source []byte, filePath string, result *ParseResult) {
	cursor := sitter.NewTreeCursor(root)
	defer cursor.Close()
	p.walkGoTree(cursor, source, filePath, result)
}

// walkGoTree recursively walks Go AST to find declarations
func (p *TreeSitterParser) walkGoTree(cursor *sitter.TreeCursor, source []byte, filePath string, result *ParseResult) {
	node := cursor.CurrentNode()

	switch node.Type() {
	case "function_declaration":
		name := p.getNodeText(node.ChildByFieldName("name"), source)
		if name != "" {
			start := node.StartPoint()
			end := node.EndPoint()
			nodeType := NodeTypeFunction
			if p.isMethod(node) {
				nodeType = NodeTypeMethod
			}
			result.Nodes = append(result.Nodes, Node{
				ID:         fmt.Sprintf("%s:func:%s", filePath, name),
				Name:       name,
				Type:       nodeType,
				File:       filePath,
				StartLine:  int(start.Row) + 1,
				EndLine:    int(end.Row) + 1,
				Exported:   isExported(name),
				Metrics:    p.calculateNodeMetrics(node, source),
			})
			if isExported(name) {
				result.Exports = append(result.Exports, Export{
					Name:     name,
					File:     filePath,
					Type:     nodeType,
					Line:     int(start.Row) + 1,
					Exported: true,
				})
			}
		}

	case "type_declaration":
		for i := uint32(0); i < node.ChildCount(); i++ {
			child := node.Child(i)
			if child.Type() == "type_spec" {
				name := p.getNodeText(child.ChildByFieldName("name"), source)
				if name != "" {
					start := node.StartPoint()
					kind := NodeTypeStruct
					if child.ChildByFieldName("type") != nil {
						typeNode := child.ChildByFieldName("type")
						if typeNode != nil && typeNode.Child(0) != nil && typeNode.Child(0).Type() == "interface_type" {
							kind = NodeTypeInterface
						}
					}
					result.Nodes = append(result.Nodes, Node{
						ID:         fmt.Sprintf("%s:type:%s", filePath, name),
						Name:       name,
						Type:       kind,
						File:       filePath,
						StartLine:  int(start.Row) + 1,
						EndLine:    int(node.EndPoint().Row) + 1,
						Exported:   isExported(name),
					})
					if isExported(name) {
						result.Exports = append(result.Exports, Export{
							Name:     name,
							File:     filePath,
							Type:     kind,
							Line:     int(start.Row) + 1,
							Exported: true,
						})
					}
				}
			}
		}

	case "import_declaration":
		importSpec := node.ChildByFieldName("import")
		if importSpec != nil {
			path := p.extractImportPath(importSpec, source)
			if path != "" {
				start := node.StartPoint()
				result.Imports = append(result.Imports, Import{
					Name:   extractNameFromPath(path),
					Source: path,
					Line:   int(start.Row) + 1,
				})
			}
		}

	case "const_declaration", "var_declaration":
		for i := uint32(0); i < node.ChildCount(); i++ {
			child := node.Child(i)
			if child.Type() == "const_spec" || child.Type() == "var_spec" {
				nameNode := child.ChildByFieldName("name")
				if nameNode != nil {
					name := p.getNodeText(nameNode, source)
					if name != "" && isExported(name) {
						start := node.StartPoint()
						result.Exports = append(result.Exports, Export{
							Name:     name,
							File:     filePath,
							Type:     NodeTypeConstant,
							Line:     int(start.Row) + 1,
							Exported: true,
						})
					}
				}
			}
		}
	}

	if cursor.GoToFirstChild() {
		p.walkGoTree(cursor, source, filePath, result)
		for cursor.GoToNextSibling() {
			p.walkGoTree(cursor, source, filePath, result)
		}
		cursor.GoToParent()
	}
}

// isMethod checks if a function is a method (has a receiver)
func (p *TreeSitterParser) isMethod(node *sitter.Node) bool {
	if node.ChildCount() > 0 && node.Child(0).Type() == "parameter_list" {
		return false
	}
	if node.ChildCount() > 1 && node.Child(0).Type() == "parenthesized_expression" {
		return true
	}
	if node.ChildCount() > 0 {
		firstChild := node.Child(0)
		if firstChild.Type() == "parameter_list" {
			return false
		}
	}
	return false
}

// parseJavaScript extracts nodes from JavaScript source code
func (p *TreeSitterParser) parseJavaScript(root *sitter.Node, source []byte, filePath string, result *ParseResult) {
	cursor := sitter.NewTreeCursor(root)
	defer cursor.Close()
	p.walkJSTree(cursor, source, filePath, result, false)
}

// parseTypeScript extracts nodes from TypeScript source code
func (p *TreeSitterParser) parseTypeScript(root *sitter.Node, source []byte, filePath string, result *ParseResult) {
	cursor := sitter.NewTreeCursor(root)
	defer cursor.Close()
	p.walkJSTree(cursor, source, filePath, result, true)
}

// walkJSTree recursively walks JS/TS AST
func (p *TreeSitterParser) walkJSTree(cursor *sitter.TreeCursor, source []byte, filePath string, result *ParseResult, isTypeScript bool) {
	node := cursor.CurrentNode()

	switch node.Type() {
	case "function_declaration", "function":
		name := p.getNodeText(node.ChildByFieldName("name"), source)
		if name != "" {
			start := node.StartPoint()
			end := node.EndPoint()
			result.Nodes = append(result.Nodes, Node{
				ID:         fmt.Sprintf("%s:func:%s", filePath, name),
				Name:       name,
				Type:       NodeTypeFunction,
				File:       filePath,
				StartLine:  int(start.Row) + 1,
				EndLine:    int(end.Row) + 1,
				Exported:   p.isExportedInJS(node, source),
				Metrics:    p.calculateNodeMetrics(node, source),
			})
			if p.isExportedInJS(node, source) {
				result.Exports = append(result.Exports, Export{
					Name:     name,
					File:     filePath,
					Type:     NodeTypeFunction,
					Line:     int(start.Row) + 1,
					Exported: true,
				})
			}
		}

	case "class_declaration", "class":
		name := p.getNodeText(node.ChildByFieldName("name"), source)
		if name != "" {
			start := node.StartPoint()
			end := node.EndPoint()
			result.Nodes = append(result.Nodes, Node{
				ID:         fmt.Sprintf("%s:class:%s", filePath, name),
				Name:       name,
				Type:       NodeTypeClass,
				File:       filePath,
				StartLine:  int(start.Row) + 1,
				EndLine:    int(end.Row) + 1,
				Exported:   p.isExportedInJS(node, source),
			})
			if p.isExportedInJS(node, source) {
				result.Exports = append(result.Exports, Export{
					Name:     name,
					File:     filePath,
					Type:     NodeTypeClass,
					Line:     int(start.Row) + 1,
					Exported: true,
				})
			}
		}

	case "lexical_declaration", "variable_declaration":
		for i := uint32(0); i < node.ChildCount(); i++ {
			child := node.Child(i)
			if child.Type() == "variable_declarator" {
				nameNode := child.ChildByFieldName("name")
				if nameNode != nil {
					name := p.getNodeText(nameNode, source)
					if name != "" {
						start := node.StartPoint()
						isExport := p.isExportedInJS(node, source)
						if isExport {
							result.Exports = append(result.Exports, Export{
								Name:     name,
								File:     filePath,
								Type:     NodeTypeVariable,
								Line:     int(start.Row) + 1,
								Exported: true,
							})
						}
					}
				}
			}
		}

	case "import_statement", "import_clause":
		sourceNode := node.ChildByFieldName("source")
		if sourceNode != nil {
			path := p.getNodeText(sourceNode, source)
			path = strings.Trim(path, "'\"")
			if path != "" {
				start := node.StartPoint()
				for i := uint32(0); i < node.ChildCount(); i++ {
					child := node.Child(i)
					if child.Type() == "import_specifier" ||
						child.Type() == "named_imports" ||
						child.Type() == "identifier" {
						name := p.getNodeText(child.ChildByFieldName("name"), source)
						if name == "" {
							name = p.getNodeText(child, source)
						}
						if name != "" && name != "{" && name != "}" && name != "," {
							result.Imports = append(result.Imports, Import{
								Name:   name,
								Source: path,
								Line:   int(start.Row) + 1,
							})
						}
					}
				}
				if node.ChildByFieldName("default") != nil {
					name := p.getNodeText(node.ChildByFieldName("default"), source)
					if name != "" {
						result.Imports = append(result.Imports, Import{
							Name:   name,
							Source: path,
							Line:   int(start.Row) + 1,
						})
					}
				}
			}
		}

	case "export_statement", "export":
		if node.ChildByFieldName("declaration") != nil {
			decl := node.ChildByFieldName("declaration")
			var name string
			var declType NodeType

			switch decl.Type() {
			case "function_declaration", "function":
				name = p.getNodeText(decl.ChildByFieldName("name"), source)
				declType = NodeTypeFunction
			case "class_declaration", "class":
				name = p.getNodeText(decl.ChildByFieldName("name"), source)
				declType = NodeTypeClass
			default:
				name = ""
			}

			if name != "" {
				start := node.StartPoint()
				result.Exports = append(result.Exports, Export{
					Name:     name,
					File:     filePath,
					Type:     declType,
					Line:     int(start.Row) + 1,
					Exported: true,
				})
			}
		}
	}

	if cursor.GoToFirstChild() {
		p.walkJSTree(cursor, source, filePath, result, isTypeScript)
		for cursor.GoToNextSibling() {
			p.walkJSTree(cursor, source, filePath, result, isTypeScript)
		}
		cursor.GoToParent()
	}
}

// isExportedInJS checks if a JS/TS declaration is exported
func (p *TreeSitterParser) isExportedInJS(node *sitter.Node, source []byte) bool {
	parent := node.Parent()
	if parent != nil && (parent.Type() == "export_statement" || parent.Type() == "export") {
		return true
	}
	if node.NextSibling() != nil && node.NextSibling().Type() == "export" {
		return true
	}
	return false
}

// walkGenericTree is a generic tree-walker that looks for common node types across all tree-sitter grammars
func (p *TreeSitterParser) walkGenericTree(root *sitter.Node, source []byte, filePath string, result *ParseResult) {
	cursor := sitter.NewTreeCursor(root)
	defer cursor.Close()
	p.walkGenericTreeCursor(cursor, source, filePath, result)
}

// walkGenericTreeCursor recursively walks a generic AST
func (p *TreeSitterParser) walkGenericTreeCursor(cursor *sitter.TreeCursor, source []byte, filePath string, result *ParseResult) {
	node := cursor.CurrentNode()
	nodeType := node.Type()

	switch nodeType {
	case "function_definition", "function_declaration", "method_declaration", "function_item":
		name := p.getNodeText(node.ChildByFieldName("name"), source)
		if name == "" {
			name = p.getNodeText(node.ChildByFieldName("declarator"), source)
		}
		if name != "" {
			start := node.StartPoint()
			end := node.EndPoint()
			nt := NodeTypeFunction
			if nodeType == "method_declaration" {
				nt = NodeTypeMethod
			}
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:func:%s", filePath, name),
				Name:      name,
				Type:      nt,
				File:      filePath,
				StartLine: int(start.Row) + 1,
				EndLine:   int(end.Row) + 1,
			})
		}

	case "class_definition", "class_declaration", "class_specifier":
		name := p.getNodeText(node.ChildByFieldName("name"), source)
		if name != "" {
			start := node.StartPoint()
			end := node.EndPoint()
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:class:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeClass,
				File:      filePath,
				StartLine: int(start.Row) + 1,
				EndLine:   int(end.Row) + 1,
			})
		}

	case "struct_declaration", "struct_specifier":
		name := p.getNodeText(node.ChildByFieldName("name"), source)
		if name != "" {
			start := node.StartPoint()
			end := node.EndPoint()
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:struct:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeStruct,
				File:      filePath,
				StartLine: int(start.Row) + 1,
				EndLine:   int(end.Row) + 1,
			})
		}

	case "interface_declaration", "interface_type":
		name := p.getNodeText(node.ChildByFieldName("name"), source)
		if name != "" {
			start := node.StartPoint()
			end := node.EndPoint()
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:interface:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeInterface,
				File:      filePath,
				StartLine: int(start.Row) + 1,
				EndLine:   int(end.Row) + 1,
			})
		}

	case "enum_declaration", "enum_specifier":
		name := p.getNodeText(node.ChildByFieldName("name"), source)
		if name != "" {
			start := node.StartPoint()
			end := node.EndPoint()
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:enum:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeEnum,
				File:      filePath,
				StartLine: int(start.Row) + 1,
				EndLine:   int(end.Row) + 1,
			})
		}

	case "import_statement", "import_declaration", "use_declaration", "use_statement":
		start := node.StartPoint()
		text := p.getNodeText(node, source)
		// Try to extract the source path
		sourcePath := ""
		if node.ChildByFieldName("path") != nil {
			sourcePath = p.getNodeText(node.ChildByFieldName("path"), source)
		}
		if sourcePath == "" {
			sourcePath = strings.TrimSpace(text)
			sourcePath = strings.TrimPrefix(sourcePath, "import ")
			sourcePath = strings.TrimPrefix(sourcePath, "use ")
			sourcePath = strings.Trim(sourcePath, ";\n\r\t \"'")
		}
		if sourcePath != "" {
			result.Imports = append(result.Imports, Import{
				Name:   extractNameFromPath(sourcePath),
				Source: sourcePath,
				Line:   int(start.Row) + 1,
			})
		}

	case "package_declaration":
		start := node.StartPoint()
		name := p.getNodeText(node, source)
		name = strings.TrimPrefix(name, "package ")
		name = strings.TrimSpace(name)
		if name != "" {
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:module:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeModule,
				File:      filePath,
				StartLine: int(start.Row) + 1,
				EndLine:   int(start.Row) + 1,
			})
		}

	case "trait_declaration":
		name := p.getNodeText(node.ChildByFieldName("name"), source)
		if name != "" {
			start := node.StartPoint()
			end := node.EndPoint()
			result.Nodes = append(result.Nodes, Node{
				ID:        fmt.Sprintf("%s:trait:%s", filePath, name),
				Name:      name,
				Type:      NodeTypeType,
				File:      filePath,
				StartLine: int(start.Row) + 1,
				EndLine:   int(end.Row) + 1,
			})
		}
	}

	if cursor.GoToFirstChild() {
		p.walkGenericTreeCursor(cursor, source, filePath, result)
		for cursor.GoToNextSibling() {
			p.walkGenericTreeCursor(cursor, source, filePath, result)
		}
		cursor.GoToParent()
	}
}

// getNodeText extracts text content from a node
func (p *TreeSitterParser) getNodeText(node *sitter.Node, source []byte) string {
	if node == nil {
		return ""
	}
	return string(node.Content(source))
}

// extractImportPath extracts the import path from an import spec node
func (p *TreeSitterParser) extractImportPath(node *sitter.Node, source []byte) string {
	if node == nil {
		return ""
	}
	if node.Type() == "import_spec" {
		pathNode := node.ChildByFieldName("path")
		if pathNode != nil {
			return p.getNodeText(pathNode, source)
		}
	}
	if node.Type() == "string_literal" {
		return p.getNodeText(node, source)
	}
	for i := uint32(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child.Type() == "string_literal" {
			return p.getNodeText(child, source)
		}
		result := p.extractImportPath(child, source)
		if result != "" {
			return result
		}
	}
	return ""
}

// calculateNodeMetrics calculates metrics for a node
func (p *TreeSitterParser) calculateNodeMetrics(node *sitter.Node, source []byte) Metrics {
	metrics := Metrics{
		LinesOfCode: 1,
		Cyclomatic:  1,
		Cognitive:   0,
	}

	start := node.StartPoint()
	end := node.EndPoint()
	metrics.LinesOfCode = int(end.Row - start.Row + 1)

	cursor := sitter.NewTreeCursor(node)
	defer cursor.Close()

	controlFlowKeywords := map[string]bool{
		"if_statement":       true,
		"else_clause":        true,
		"for_statement":      true,
		"while_statement":    true,
		"do_statement":       true,
		"switch_statement":   true,
		"case_clause":        true,
		"catch_clause":       true,
		"ternary_expression": true,
	}

	if cursor.GoToFirstChild() {
		for {
			n := cursor.CurrentNode()
			if controlFlowKeywords[n.Type()] {
				metrics.Cyclomatic++
			}
			if !cursor.GoToNextSibling() {
				break
			}
		}
		cursor.GoToParent()
	}

	metrics.Cognitive = metrics.Cyclomatic
	return metrics
}

// RegisterTreeSitterParsers registers tree-sitter based parsers for all languages
func RegisterTreeSitterParsers() {
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
		LanguageHTML,
		LanguageCSS,
		LanguageSQL,
		LanguageShell,
		LanguageYAML,
		LanguageJSON,
		LanguageTOML,
		LanguageMarkdown,
	}

	for _, lang := range languages {
		parser := NewTreeSitterParser(lang)
		if parser != nil {
			RegisterParser(parser)
		}
	}
}
