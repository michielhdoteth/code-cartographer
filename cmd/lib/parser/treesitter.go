package parser

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/javascript"
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
		// TypeScript uses JavaScript grammar for now
		// Tree-sitter TypeScript grammar is available separately
		langPtr = javascript.GetLanguage()
	case LanguagePython:
		// Python grammar might not be in the main package
		// Fall back to nil and use regex parser
		return nil
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
	}

	// Calculate metrics
	result.CalculateMetrics()

	return result, nil
}

// parseGo extracts nodes from Go source code
func (p *TreeSitterParser) parseGo(root *sitter.Node, source []byte, filePath string, result *ParseResult) {
	cursor := sitter.NewTreeCursor(root)
	defer cursor.Close()

	// Walk the tree and extract declarations
	p.walkGoTree(cursor, source, filePath, result)
}

// walkGoTree recursively walks Go AST to find declarations
func (p *TreeSitterParser) walkGoTree(cursor *sitter.TreeCursor, source []byte, filePath string, result *ParseResult) {
	node := cursor.CurrentNode()

	// Process based on node type
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
		// Handle type X struct {} or type X interface {}
		for i := uint32(0); i < node.ChildCount(); i++ {
			child := node.Child(i)
			if child.Type() == "type_spec" {
				name := p.getNodeText(child.ChildByFieldName("name"), source)
				if name != "" {
					start := node.StartPoint()
					kind := NodeTypeStruct
					// Check if it's an interface
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
		// Handle import "path" or import (
		//   "path1"
		//   "path2"
		// )
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
		// Handle const/var declarations
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

	// Recurse into children
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
	// Check if first child is a parenthesized expression (receiver)
	if node.ChildCount() > 0 && node.Child(0).Type() == "parameter_list" {
		return false
	}
	if node.ChildCount() > 1 && node.Child(0).Type() == "parenthesized_expression" {
		return true
	}
	// Also check for method receiver pattern: func (x Type) Name()
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
		// Handle: const/let/var x = ... or const/let/var x = function() {}
		for i := uint32(0); i < node.ChildCount(); i++ {
			child := node.Child(i)
			if child.Type() == "variable_declarator" {
				nameNode := child.ChildByFieldName("name")
				if nameNode != nil {
					name := p.getNodeText(nameNode, source)
					if name != "" {
						start := node.StartPoint()
						// Check if it's exported (exported declaration)
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
		// Handle ES6 imports: import x from 'path' or import { x, y } from 'path'
		sourceNode := node.ChildByFieldName("source")
		if sourceNode != nil {
			path := p.getNodeText(sourceNode, source)
			path = strings.Trim(path, "'\"")
			if path != "" {
				start := node.StartPoint()
				// Extract imported names
				// For import { x, y } from 'path'
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
				// For default import: import x from 'path'
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
		// Handle exports
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

	// Recurse into children
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
	// Check if parent is export statement
	parent := node.Parent()
	if parent != nil && (parent.Type() == "export_statement" || parent.Type() == "export") {
		return true
	}
	// Check for export keyword
	if node.NextSibling() != nil && node.NextSibling().Type() == "export" {
		return true
	}
	return false
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
	// Handle import "path" format
	if node.Type() == "import_spec" {
		pathNode := node.ChildByFieldName("path")
		if pathNode != nil {
			return p.getNodeText(pathNode, source)
		}
	}
	// Handle string_literal directly
	if node.Type() == "string_literal" {
		return p.getNodeText(node, source)
	}
	// Recursively search for string_literal
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

	// Count lines
	start := node.StartPoint()
	end := node.EndPoint()
	metrics.LinesOfCode = int(end.Row - start.Row + 1)

	// Count control flow keywords in the node's subtree
	cursor := sitter.NewTreeCursor(node)
	defer cursor.Close()

	controlFlowKeywords := map[string]bool{
		"if_statement":        true,
		"else_clause":         true,
		"for_statement":       true,
		"while_statement":     true,
		"do_statement":        true,
		"switch_statement":    true,
		"case_clause":         true,
		"catch_clause":        true,
		"ternary_expression":  true, // ? :
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

// RegisterTreeSitterParsers registers tree-sitter based parsers
func RegisterTreeSitterParsers() {
	languages := []Language{
		LanguageGo,
		LanguageJavaScript,
		LanguageTypeScript,
	}

	for _, lang := range languages {
		parser := NewTreeSitterParser(lang)
		if parser != nil {
			RegisterParser(parser)
		}
	}
}