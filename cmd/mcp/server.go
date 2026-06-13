package mcp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"carto/analyze"
	"carto/find"
	"carto/git"
	"carto/manifest"

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

	// Diff since last map tool
	mcp.AddTool(s, &mcp.Tool{
		Name:        "diff_since_last_map",
		Description: "Show files changed since the last carto map was created",
	}, diffSinceLastMapTool)

	// Get changed files tool
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_changed_files",
		Description: "Show uncommitted changes (added, modified, deleted files)",
	}, getChangedFilesTool)

	// Impact analysis tool
	mcp.AddTool(s, &mcp.Tool{
		Name:        "impact_analysis",
		Description: "Analyze the impact of changing a specific file - shows dependents and risk level",
	}, impactAnalysisTool)
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

type DiffSinceLastMapInput struct {
	Path string `json:"path" jsonschema:"required,description=The directory path to check"`
}

type DiffSinceLastMapOutput struct {
	ChangedFiles []string `json:"changed_files"`
	Count        int      `json:"count"`
	FromCommit   string   `json:"from_commit"`
	ToCommit     string   `json:"to_commit"`
	Message      string   `json:"message"`
}

type GetChangedFilesInput struct {
	Path string `json:"path" jsonschema:"required,description=The directory path to check"`
}

type GetChangedFilesOutput struct {
	Added    []string `json:"added"`
	Modified []string `json:"modified"`
	Deleted  []string `json:"deleted"`
	Total    int      `json:"total"`
	Message  string   `json:"message"`
}

type ImpactAnalysisInput struct {
	Path string `json:"path" jsonschema:"required,description=The directory path to analyze"`
	File string `json:"file" jsonschema:"required,description=The file path to analyze impact for"`
}

type ImpactAnalysisOutput struct {
	Dependents   []string `json:"dependents"`
	Dependencies []string `json:"dependencies"`
	Risk         string   `json:"risk"`
	ReadSet      []string `json:"read_set"`
	Message      string   `json:"message"`
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

func diffSinceLastMapTool(ctx context.Context, req *mcp.CallToolRequest, input DiffSinceLastMapInput) (*mcp.CallToolResult, DiffSinceLastMapOutput, error) {
	path := input.Path
	if path == "" {
		path = "."
	}
	absPath, _ := filepath.Abs(path)

	if !git.IsRepo(absPath) {
		return nil, DiffSinceLastMapOutput{Message: "Not a git repository"}, nil
	}

	m, err := manifest.Load(absPath)
	if err != nil {
		return nil, DiffSinceLastMapOutput{Message: "No manifest found. Run 'carto map' first"}, nil
	}

	currentSHA, _ := git.GetHeadSHA(absPath)
	if m.Commit == currentSHA {
		return nil, DiffSinceLastMapOutput{Message: "No changes since last map"}, nil
	}

	files, err := git.GetChangedFiles(absPath, m.Commit, currentSHA)
	if err != nil {
		return nil, DiffSinceLastMapOutput{Message: fmt.Sprintf("Error: %v", err)}, nil
	}

	return nil, DiffSinceLastMapOutput{
		ChangedFiles: files,
		Count:        len(files),
		FromCommit:   m.Commit,
		ToCommit:     currentSHA,
		Message:      fmt.Sprintf("Changed %d files since commit %s", len(files), m.Commit[:8]),
	}, nil
}

func getChangedFilesTool(ctx context.Context, req *mcp.CallToolRequest, input GetChangedFilesInput) (*mcp.CallToolResult, GetChangedFilesOutput, error) {
	path := input.Path
	if path == "" {
		path = "."
	}
	absPath, _ := filepath.Abs(path)

	if !git.IsRepo(absPath) {
		return nil, GetChangedFilesOutput{Message: "Not a git repository"}, nil
	}

	statuses, err := git.GetStatus(absPath)
	if err != nil {
		return nil, GetChangedFilesOutput{Message: fmt.Sprintf("Error: %v", err)}, nil
	}

	var added, modified, deleted []string
	for _, s := range statuses {
		if len(s) < 3 {
			continue
		}
		code := s[:2]
		file := s[3:]
		switch {
		case strings.Contains(code, "A"):
			added = append(added, file)
		case strings.Contains(code, "D"):
			deleted = append(deleted, file)
		default:
			modified = append(modified, file)
		}
	}

	return nil, GetChangedFilesOutput{
		Added:    added,
		Modified: modified,
		Deleted:  deleted,
		Total:    len(statuses),
		Message:  fmt.Sprintf("%d uncommitted changes", len(statuses)),
	}, nil
}

func impactAnalysisTool(ctx context.Context, req *mcp.CallToolRequest, input ImpactAnalysisInput) (*mcp.CallToolResult, ImpactAnalysisOutput, error) {
	absPath, _ := filepath.Abs(input.Path)

	graph, err := analyze.BuildGraph(absPath, "go")
	if err != nil {
		return nil, ImpactAnalysisOutput{Message: fmt.Sprintf("Error: %v", err)}, nil
	}

	var dependents []string
	for filePath, imports := range graph.Imports {
		for _, imp := range imports {
			if strings.Contains(imp.Source, input.File) || strings.Contains(filePath, input.File) {
				dependents = append(dependents, filePath)
				break
			}
		}
	}

	var dependencies []string
	absFile := filepath.Join(absPath, input.File)
	if fileNode, ok := graph.Files[absFile]; ok {
		for _, imp := range fileNode.Imports {
			dependencies = append(dependencies, imp.Source)
		}
	}

	risk := "low"
	if len(dependents) > 5 {
		risk = "high"
	} else if len(dependents) > 2 {
		risk = "medium"
	}

	readSet := []string{input.File}
	for i, d := range dependents {
		if i >= 4 {
			break
		}
		readSet = append(readSet, d)
	}

	return nil, ImpactAnalysisOutput{
		Dependents:   dependents,
		Dependencies: dependencies,
		Risk:         risk,
		ReadSet:      readSet,
		Message:      fmt.Sprintf("File %s has %d dependents, risk: %s", input.File, len(dependents), risk),
	}, nil
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