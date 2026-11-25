package parser

import (
	"fmt"
	"sort"
	"strings"
)

type NodeType string

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

type EdgeType string

const (
	EdgeTypeImports    EdgeType = "imports"
	EdgeTypeCalls      EdgeType = "calls"
	EdgeTypeInherits   EdgeType = "inherits"
	EdgeTypeContains   EdgeType = "contains"
	EdgeTypeUses       EdgeType = "uses"
	EdgeTypeReferences EdgeType = "references"
	EdgeTypeDecorates  EdgeType = "decorates"
)

type Node struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Type        NodeType `json:"type"`
	File        string   `json:"file"`
	StartLine   int      `json:"start_line"`
	StartCol    int      `json:"start_col"`
	EndLine     int      `json:"end_line"`
	EndCol      int      `json:"end_col"`
	ParentID    string   `json:"parent_id,omitempty"`
	Exported    bool     `json:"exported"`
	Metrics     Metrics  `json:"metrics"`
	Decorators  []string `json:"decorators,omitempty"`
	Annotations []string `json:"annotations,omitempty"`
	Signature   string   `json:"signature,omitempty"`
	DocComment  string   `json:"doc_comment,omitempty"`
}

type Edge struct {
	ID     string   `json:"id"`
	Source string   `json:"source"` // Node ID or file path
	Target string   `json:"target"` // Node ID or file path
	Type   EdgeType `json:"type"`
	Line   int      `json:"line,omitempty"`
	Weight float64  `json:"weight,omitempty"`
}

type Metrics struct {
	LinesOfCode    int `json:"lines_of_code"`
	Cyclomatic     int `json:"cyclomatic"`
	Cognitive      int `json:"cognitive"`
	MaxNesting     int `json:"max_nesting"`
	ParameterCount int `json:"parameter_count"`
	ReturnCount    int `json:"return_count"`
}

type Export struct {
	Name     string   `json:"name"`
	File     string   `json:"file"`
	Type     NodeType `json:"type"`
	Line     int      `json:"line"`
	Exported bool     `json:"exported"`
}

type Import struct {
	Name   string `json:"name"`
	Source string `json:"source"` // module path or file path
	Alias  string `json:"alias,omitempty"`
	Line   int    `json:"line"`
	Used   bool   `json:"used"`
}

type ParseError struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Col     int    `json:"col"`
	Message string `json:"message"`
}

type ParseResult struct {
	FilePath string       `json:"file_path"`
	Language Language     `json:"language"`
	AST      interface{}  `json:"ast,omitempty"` // Raw AST, language-specific
	Nodes    []Node       `json:"nodes"`
	Edges    []Edge       `json:"edges"`
	Exports  []Export     `json:"exports"`
	Imports  []Import     `json:"imports"`
	Metrics  Metrics      `json:"metrics"`
	Errors   []ParseError `json:"errors,omitempty"`
}

type Codebase struct {
	RootPath string                  `json:"root_path"`
	Language Language                `json:"language"`
	Files    map[string]*ParseResult `json:"files"`
	Graph    *DependencyGraph        `json:"graph"`
	Stats    CodebaseStats           `json:"stats"`
}

type CodebaseStats struct {
	TotalFiles     int            `json:"total_files"`
	TotalNodes     int            `json:"total_nodes"`
	TotalEdges     int            `json:"total_edges"`
	TotalExports   int            `json:"total_exports"`
	TotalImports   int            `json:"total_imports"`
	LinesOfCode    int            `json:"lines_of_code"`
	LanguageCounts map[string]int `json:"language_counts"`
}

type DependencyGraph struct {
	Nodes map[string]*GraphNode `json:"nodes"`
	Edges []GraphEdge           `json:"edges"`
}

type GraphNode struct {
	ID      string   `json:"id"`
	File    string   `json:"file"`
	Exports []string `json:"exports"`
	Imports []string `json:"imports"`
	Metrics Metrics  `json:"metrics"`
}

type GraphEdge struct {
	Source string   `json:"source"`
	Target string   `json:"target"`
	Type   EdgeType `json:"type"`
}

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

func (r *ParseResult) GetNodeByID(id string) *Node {
	for i := range r.Nodes {
		if r.Nodes[i].ID == id {
			return &r.Nodes[i]
		}
	}
	return nil
}

func (r *ParseResult) GetNodesByType(nodeType NodeType) []Node {
	var result []Node
	for _, n := range r.Nodes {
		if n.Type == nodeType {
			result = append(result, n)
		}
	}
	return result
}

func (r *ParseResult) GetExportedNodes() []Node {
	var result []Node
	for _, n := range r.Nodes {
		if n.Exported {
			result = append(result, n)
		}
	}
	return result
}

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

func (r *ParseResult) SortNodes() {
	sort.Slice(r.Nodes, func(i, j int) bool {
		if r.Nodes[i].File != r.Nodes[j].File {
			return r.Nodes[i].File < r.Nodes[j].File
		}
		return r.Nodes[i].StartLine < r.Nodes[j].StartLine
	})
}

func (r *ParseResult) SortEdges() {
	sort.Slice(r.Edges, func(i, j int) bool {
		if r.Edges[i].Source != r.Edges[j].Source {
			return r.Edges[i].Source < r.Edges[j].Source
		}
		return r.Edges[i].Target < r.Edges[j].Target
	})
}
