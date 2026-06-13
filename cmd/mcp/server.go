package mcp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"carto/analyze"
	"carto/find"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server wraps the MCP server instance
type Server struct {
	server *mcp.Server
}

// NewServer creates a new MCP server for Code Cartographer
func NewServer() *Server {
	s := mcp.NewServer(&mcp.Implementation{
		Name:    "carto",
		Version: "2.0.0",
	}, nil)

	// Register tools
	registerTools(s)

	return &Server{server: s}
}

// registerTools registers all MCP tools
func registerTools(s *mcp.Server) {
	// Map codebase tool
	mcp.AddTool(s, &mcp.Tool{
		Name:        "map_codebase",
		Description: "Scan and map a codebase directory, returning file structure and language distribution",
	}, mapCodebaseTool)

	// Analyze dead code tool
	mcp.AddTool(s, &mcp.Tool{
		Name:        "analyze_dead_code",
		Description: "Find unused exports, files, and dependencies in the codebase",
	}, analyzeDeadCodeTool)

	// Analyze complexity tool
	mcp.AddTool(s, &mcp.Tool{
		Name:        "analyze_complexity",
		Description: "Find functions with high cyclomatic or cognitive complexity",
	}, analyzeComplexityTool)

	// Find circular dependencies tool
	mcp.AddTool(s, &mcp.Tool{
		Name:        "find_circular_deps",
		Description: "Detect circular dependencies in the codebase",
	}, findCircularDepsTool)

	// Health check tool
	mcp.AddTool(s, &mcp.Tool{
		Name:        "code_health",
		Description: "Get overall code health metrics and issues",
	}, codeHealthTool)

	// Search tool
	mcp.AddTool(s, &mcp.Tool{
		Name:        "search_code",
		Description: "Search for text or regex patterns in the codebase",
	}, searchCodeTool)
}

// Tool input/output types
type MapCodebaseInput struct {
	Path     string `json:"path" jsonschema:"required,description=The directory path to map"`
	Language string `json:"language" jsonschema:"description=Language to filter (go, typescript, etc.)"`
}

type MapCodebaseOutput struct {
	Files     int            `json:"files" jsonschema:"description=Total files scanned"`
	Languages map[string]int `json:"languages" jsonschema:"description=Files per language"`
	Message   string         `json:"message"`
}

type AnalyzeDeadCodeInput struct {
	Path     string `json:"path" jsonschema:"required,description=The directory path to analyze"`
	Language string `json:"language" jsonschema:"description=Language to analyze (go, typescript, python, etc.)"`
}

type AnalyzeDeadCodeOutput struct {
	UnusedExports int    `json:"unused_exports"`
	UnusedFiles   int    `json:"unused_files"`
	TotalIssues   int    `json:"total_issues"`
	Message       string `json:"message"`
}

type AnalyzeComplexityInput struct {
	Path     string `json:"path" jsonschema:"required,description=The directory path to analyze"`
	Language string `json:"language" jsonschema:"description=Language to analyze (go, typescript, python, etc.)"`
	Top      int    `json:"top" jsonschema:"description=Number of top complex functions to return"`
}

type AnalyzeComplexityOutput struct {
	Functions     int     `json:"functions"`
	AvgComplexity float64 `json:"avg_complexity"`
	Issues        int     `json:"issues"`
	Message       string  `json:"message"`
}

type FindCircularDepsInput struct {
	Path     string `json:"path" jsonschema:"required,description=The directory path to analyze"`
	Language string `json:"language" jsonschema:"description=Language to analyze (go, typescript, python, etc.)"`
}

type FindCircularDepsOutput struct {
	Cycles  int    `json:"cycles"`
	Message string `json:"message"`
}

type CodeHealthInput struct {
	Path     string `json:"path" jsonschema:"required,description=The directory path to analyze"`
	Language string `json:"language" jsonschema:"description=Language to analyze (go, typescript, python, etc.)"`
}

type CodeHealthOutput struct {
	Score   float64 `json:"score"`
	Grade   string  `json:"grade"`
	Issues  int     `json:"issues"`
	Message string  `json:"message"`
}

type SearchCodeInput struct {
	Path         string `json:"path" jsonschema:"required,description=The directory path to search"`
	Query        string `json:"query" jsonschema:"required,description=Search pattern"`
	CaseSensitive bool   `json:"case_sensitive" jsonschema:"description=Case sensitive search"`
	UseRegex     bool   `json:"use_regex" jsonschema:"description=Use regex pattern"`
}

type SearchCodeOutput struct {
	Results int    `json:"results"`
	Message string `json:"message"`
}

// Tool handlers

func mapCodebaseTool(ctx context.Context, req *mcp.CallToolRequest, input MapCodebaseInput) (*mcp.CallToolResult, MapCodebaseOutput, error) {
	path := input.Path
	if path == "" {
		path = "."
	}

	absPath, _ := filepath.Abs(path)
	graph, err := analyze.BuildGraph(absPath, input.Language)
	if err != nil {
		return nil, MapCodebaseOutput{Message: fmt.Sprintf("Error: %v", err)}, nil
	}

	languages := make(map[string]int)
	for _, f := range graph.Files {
		languages[f.Language]++
	}

	output := MapCodebaseOutput{
		Files:     len(graph.Files),
		Languages: languages,
		Message:   fmt.Sprintf("Scanned %d files", len(graph.Files)),
	}
	return nil, output, nil
}

func analyzeDeadCodeTool(ctx context.Context, req *mcp.CallToolRequest, input AnalyzeDeadCodeInput) (*mcp.CallToolResult, AnalyzeDeadCodeOutput, error) {
	absPath, _ := filepath.Abs(input.Path)
	lang := input.Language
	if lang == "" {
		lang = "auto"
	}
	graph, err := analyze.BuildGraph(absPath, lang)
	if err != nil {
		return nil, AnalyzeDeadCodeOutput{Message: fmt.Sprintf("Error: %v", err)}, nil
	}

	result := analyze.AnalyzeDeadCode(graph)
	output := AnalyzeDeadCodeOutput{
		UnusedExports: len(result.UnusedExports),
		UnusedFiles:   len(result.UnusedFiles),
		TotalIssues:   result.TotalIssues,
		Message:       result.GetSummary(),
	}
	return nil, output, nil
}

func analyzeComplexityTool(ctx context.Context, req *mcp.CallToolRequest, input AnalyzeComplexityInput) (*mcp.CallToolResult, AnalyzeComplexityOutput, error) {
	absPath, _ := filepath.Abs(input.Path)
	lang := input.Language
	if lang == "" {
		lang = "auto"
	}
	graph, err := analyze.BuildGraph(absPath, lang)
	if err != nil {
		return nil, AnalyzeComplexityOutput{Message: fmt.Sprintf("Error: %v", err)}, nil
	}

	result := analyze.AnalyzeComplexity(graph)
	output := AnalyzeComplexityOutput{
		Functions:     len(result.Functions),
		AvgComplexity: result.AvgComplexity,
		Issues:        result.Issues,
		Message:       result.GetSummary(),
	}
	return nil, output, nil
}

func findCircularDepsTool(ctx context.Context, req *mcp.CallToolRequest, input FindCircularDepsInput) (*mcp.CallToolResult, FindCircularDepsOutput, error) {
	absPath, _ := filepath.Abs(input.Path)
	lang := input.Language
	if lang == "" {
		lang = "auto"
	}
	graph, err := analyze.BuildGraph(absPath, lang)
	if err != nil {
		return nil, FindCircularDepsOutput{Message: fmt.Sprintf("Error: %v", err)}, nil
	}

	result := analyze.AnalyzeCircularDeps(graph)
	output := FindCircularDepsOutput{
		Cycles:  result.TotalCount,
		Message: result.GetSummary(),
	}
	return nil, output, nil
}

func codeHealthTool(ctx context.Context, req *mcp.CallToolRequest, input CodeHealthInput) (*mcp.CallToolResult, CodeHealthOutput, error) {
	absPath, _ := filepath.Abs(input.Path)
	lang := input.Language
	if lang == "" {
		lang = "auto"
	}
	graph, err := analyze.BuildGraph(absPath, lang)
	if err != nil {
		return nil, CodeHealthOutput{Message: fmt.Sprintf("Error: %v", err)}, nil
	}

	result := analyze.CalculateHealth(graph)
	output := CodeHealthOutput{
		Score:   result.Score,
		Grade:   result.Grade,
		Issues:  len(result.Issues),
		Message: result.GetSummary(),
	}
	return nil, output, nil
}

func searchCodeTool(ctx context.Context, req *mcp.CallToolRequest, input SearchCodeInput) (*mcp.CallToolResult, SearchCodeOutput, error) {
	absPath, _ := filepath.Abs(input.Path)

	opts := find.SearchOptions{
		CaseSensitive: input.CaseSensitive,
		UseRegex:      input.UseRegex,
		MaxResults:    100,
	}

	results, err := find.Search(absPath, input.Query, opts)
	if err != nil {
		return nil, SearchCodeOutput{Message: fmt.Sprintf("Error: %v", err)}, nil
	}

	output := SearchCodeOutput{
		Results: len(results),
		Message: fmt.Sprintf("Found %d matches", len(results)),
	}
	return nil, output, nil
}

// Run starts the MCP server over stdio
func (s *Server) Run(ctx context.Context) error {
	fmt.Println("Starting Code Cartographer MCP server...")
	fmt.Println("Use this server with Claude Desktop or other MCP clients")
	fmt.Println("")
	fmt.Println("To configure Claude Desktop, add this to your config:")
	fmt.Println(`{`)
	fmt.Println(` "mcpServers": {`)
	fmt.Println(` "carto": {`)
	fmt.Printf(` "command": "%s"\n`, os.Args[0])
	fmt.Println(` "args": ["mcp"]`)
	fmt.Println(` }`)
	fmt.Println(` }`)
	fmt.Println(`}`)
	fmt.Println("")

	return s.server.Run(ctx, &mcp.StdioTransport{})
}