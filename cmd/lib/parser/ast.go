package parser

import (
	"fmt"
	"sort"
	"strings"
)

// NodeType represents the type of a code node
type NodeType string

// Node types
const (
	NodeTypeFunction  NodeType = "function"
	NodeTypeMethod    NodeType = "method"
	NodeTypeClass     NodeType = "class"
	NodeTypeStruct    NodeType = "struct"
	NodeTypeInterface NodeType = "interface"
	NodeTypeEnum      NodeType = "enum"
	NodeTypeVariable  NodeType = "variable"
	NodeTypeConstant  NodeType = "constant"
	NodeTypeProperty  NodeType = "property"
	NodeTypeModule    NodeType = "module"
	NodeTypePackage   NodeType = "package"
	NodeTypeFile      NodeType = "file"
	NodeTypeDecorator NodeType = "decorator"
	NodeTypeType      NodeType = "type"
	NodeTypeImport    NodeType = "import"
	NodeTypeExport    NodeType = "export"
	NodeTypeUnknown   NodeType = "unknown"
)

// EdgeType represents the type of relationship between nodes
type EdgeType string

// Edge types
const (
	EdgeTypeImports    EdgeType = "imports"
	EdgeTypeCalls      EdgeType = "calls"
	EdgeTypeInherits   EdgeType = "inherits"
	EdgeTypeContains   EdgeType = "contains"
	EdgeTypeUses       EdgeType = "uses"
	EdgeTypeReferences EdgeType = "references"
	EdgeTypeDecorates  EdgeType = "decorates"
)

// Node represents a code element (function, class, etc.)
type Node struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        NodeType  `json:"type"`
	File        string    `json:"file"`
	StartLine   int       `json:"start_line"`
	StartCol    int       `json:"start_col"`
	EndLine     int       `json:"end_line"`
	EndCol      int       `json:"end_col"`
	ParentID    string    `json:"parent_id,omitempty"`
	Exported    bool      `json:"exported"`
	Metrics     Metrics   `json:"metrics"`
	Decorators  []string  `json:"decorators,omitempty"`
	Annotations []string  `json:"annotations,omitempty"`
	Signature   string    `json:"signature,omitempty"`
	DocComment  string    `json:"doc_comment,omitempty"`
}

// Edge represents a relationship between nodes
type Edge struct {
	ID     string   `json:"id"`
	Source string   `json:"source"` // Node ID or file path
	Target string   `json:"target"` // Node ID or file path
	Type   EdgeType `json:"type"`
	Weight float64  `json:"weight,omitempty"`
}

// Metrics contains code metrics for a node
type Metrics struct {
	LinesOfCode    int `json:"lines_of_code"`
	Cyclomatic     int `json:"cyclomatic"`
	Cognitive      int `json:"cognitive"`
	MaxNesting     int `json:"max_nesting"`
	ParameterCount int `json:"parameter_count"`
	ReturnCount    int `json:"return_count"`
}

// Export represents an exported symbol
type Export struct {
	Name    string   `json:"name"`
	File    string   `json:"file"`
	Type    NodeType `json:"type"`
	Line    int      `json:"line"`
	Exported bool    `json:"exported"`
}

// Import represents an imported symbol
type Import struct {
	Name   string `json:"name"`
	Source string `json:"source"` // module path or file path
	Alias  string `json:"alias,omitempty"`
	Line   int    `json:"line"`
	Used   bool   `json:"used"`
}

// ParseError represents a parsing error
type ParseError struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Col     int    `json:"col"`
	Message string `json:"message"`
}

// ParseResult contains the result of parsing a file
type ParseResult struct {
	FilePath string   `json:"file_path"`
	Language Language `json:"language"`
	AST      interface{} `json:"ast,omitempty"` // Raw AST, language-specific
	Nodes    []Node   `json:"nodes"`
	Edges    []Edge   `json:"edges"`
	Exports  []Export `json:"exports"`
	Imports  []Import `json:"imports"`
	Metrics  Metrics  `json:"metrics"`
	Errors   []ParseError `json:"errors,omitempty"`
}

// Codebase represents a parsed codebase
type Codebase struct {
	RootPath string                  `json:"root_path"`
	Language Language                `json:"language"`
	Files    map[string]*ParseResult `json:"files"`
	Graph    *DependencyGraph        `json:"graph"`
	Stats    CodebaseStats           `json:"stats"`
}

// CodebaseStats contains statistics about the codebase
type CodebaseStats struct {
	TotalFiles     int            `json:"total_files"`
	TotalNodes     int            `json:"total_nodes"`
	TotalEdges     int            `json:"total_edges"`
	TotalExports   int            `json:"total_exports"`
	TotalImports   int            `json:"total_imports"`
	LinesOfCode    int            `json:"lines_of_code"`
	LanguageCounts map[string]int `json:"language_counts"`
}

// DependencyGraph represents the dependency graph
type DependencyGraph struct {
	Nodes map[string]*GraphNode `json:"nodes"`
	Edges []GraphEdge           `json:"edges"`
}

// GraphNode represents a node in the dependency graph
type GraphNode struct {
	ID       string   `json:"id"`
	File     string   `json:"file"`
	Exports  []string `json:"exports"`
	Imports  []string `json:"imports"`
	Metrics  Metrics  `json:"metrics"`
}

// GraphEdge represents an edge in the dependency graph
type GraphEdge struct {
	Source string   `json:"source"`
	Target string   `json:"target"`
	Type   EdgeType `json:"type"`
}

// String returns a human-readable summary
func (r *ParseResult) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("File: %s (%s)\n", r.FilePath, r.Language))
	sb.WriteString(fmt.Sprintf("  Nodes: %d\n", len(r.Nodes)))
	sb.WriteString(fmt.Sprintf("  Edges: %d\n", len(r.Edges)))
	sb.WriteString(fmt.Sprintf("  Exports: %d\n", len(r.Exports)))
	sb.WriteString(fmt.Sprintf("  Imports: %d\n", len(r.Imports)))
	sb.WriteString(fmt.Sprintf("  Lines: %d\n", r.Metrics.LinesOfCode))
	return sb.String()
}

// GetNodeByID returns a node by its ID
func (r *ParseResult) GetNodeByID(id string) *Node {
	for i := range r.Nodes {
		if r.Nodes[i].ID == id {
			return &r.Nodes[i]
		}
	}
	return nil
}

// GetNodesByType returns all nodes of a specific type
func (r *ParseResult) GetNodesByType(nodeType NodeType) []Node {
	var result []Node
	for _, n := range r.Nodes {
		if n.Type == nodeType {
			result = append(result, n)
		}
	}
	return result
}

// GetExportedNodes returns all exported nodes
func (r *ParseResult) GetExportedNodes() []Node {
	var result []Node
	for _, n := range r.Nodes {
		if n.Exported {
			result = append(result, n)
		}
	}
	return result
}

// CalculateMetrics calculates overall metrics for the parse result
func (r *ParseResult) CalculateMetrics() {
	r.Metrics.LinesOfCode = 0
	r.Metrics.Cyclomatic = 0
	r.Metrics.Cognitive = 0

	for _, node := range r.Nodes {
		r.Metrics.LinesOfCode += node.Metrics.LinesOfCode
		r.Metrics.Cyclomatic += node.Metrics.Cyclomatic
		r.Metrics.Cognitive += node.Metrics.Cognitive
		if node.Metrics.MaxNesting > r.Metrics.MaxNesting {
			r.Metrics.MaxNesting = node.Metrics.MaxNesting
		}
	}
}

// SortNodes sorts nodes by line number
func (r *ParseResult) SortNodes() {
	sort.Slice(r.Nodes, func(i, j int) bool {
		if r.Nodes[i].File != r.Nodes[j].File {
			return r.Nodes[i].File < r.Nodes[j].File
		}
		return r.Nodes[i].StartLine < r.Nodes[j].StartLine
	})
}

// SortEdges sorts edges by source
func (r *ParseResult) SortEdges() {
	sort.Slice(r.Edges, func(i, j int) bool {
		if r.Edges[i].Source != r.Edges[j].Source {
			return r.Edges[i].Source < r.Edges[j].Source
		}
		return r.Edges[i].Target < r.Edges[j].Target
	})
}