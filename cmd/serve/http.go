package serve

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"carto/analyze"
	"carto/codegraph"
	"carto/find"
)

type Server struct {
	Port     string
	BuildDir string
}

func NewServer(port string, buildDir string) *Server {
	return &Server{
		Port:     port,
		BuildDir: buildDir,
	}
}

func (s *Server) Run() error {
	addr := fmt.Sprintf(":%s", s.Port)
	fmt.Printf("Starting Code Cartographer UI...\n")
	fmt.Printf(" Web UI: http://localhost%s\n", addr)

	mux := http.NewServeMux()

	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/map", s.handleMap)
	mux.HandleFunc("/api/analyze", s.handleAnalyze)
	mux.HandleFunc("/api/search", s.handleSearch)
	mux.HandleFunc("/api/files/", s.handleFile)
	mux.HandleFunc("/api/graph", s.handleGraph)
	mux.HandleFunc("/api/symbol-graph", s.handleSymbolGraph)

	if s.BuildDir != "" && dirExists(s.BuildDir) {
		mux.Handle("/", http.FileServer(http.Dir(s.BuildDir)))
	} else {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(htmlIndex))
		})
	}

	fmt.Printf("\nServer running. Press Ctrl+C to stop.\n")
	return http.ListenAndServe(addr, mux)
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "version": "2.0.0"})
}

func (s *Server) handleMap(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := r.URL.Query().Get("path")
	if path == "" {
		path = "."
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "invalid path: %v"}`, err), 400)
		return
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, fmt.Sprintf(`{"error": "path does not exist: %s"}`, absPath), 404)
			return
		}
		http.Error(w, fmt.Sprintf(`{"error": %v}`, err), 500)
		return
	}

	if !info.IsDir() {
		http.Error(w, `{"error": "path is not a directory"}`, 400)
		return
	}

	lang := r.URL.Query().Get("language")
	if lang == "" {
		lang = "auto"
	}
	graph, err := analyze.BuildGraph(absPath, lang)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "failed to build graph: %v"}`, err), 500)
		return
	}

	result := map[string]interface{}{
		"path":      absPath,
		"fileCount": len(graph.Files),
		"files":     getFileList(graph),
	}

	json.NewEncoder(w).Encode(result)
}

func (s *Server) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := r.URL.Query().Get("path")
	if path == "" {
		path = "."
	}

	analysisType := r.URL.Query().Get("type")
	if analysisType == "" {
		analysisType = "all"
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "invalid path: %v"}`, err), 400)
		return
	}

	lang := r.URL.Query().Get("language")
	if lang == "" {
		lang = "auto"
	}
	graph, err := analyze.BuildGraph(absPath, lang)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "failed to build graph: %v"}`, err), 500)
		return
	}

	var result interface{}
	switch analysisType {
	case "dead-code":
		result = analyze.AnalyzeDeadCode(graph)
	case "complexity":
		result = analyze.AnalyzeComplexity(graph)
	case "circular":
		result = analyze.AnalyzeCircularDeps(graph)
	case "structure":
		result = analyze.AnalyzeStructure(graph, analyze.DefaultStructureConfig())
	case "health":
		result = analyze.CalculateHealth(graph)
	default:
		result = map[string]interface{}{
			"dead-code":  analyze.AnalyzeDeadCode(graph),
			"complexity": analyze.AnalyzeComplexity(graph),
			"circular":   analyze.AnalyzeCircularDeps(graph),
			"structure":  analyze.AnalyzeStructure(graph, analyze.DefaultStructureConfig()),
			"health":     analyze.CalculateHealth(graph),
		}
	}

	response := map[string]interface{}{
		"type":      analysisType,
		"path":      absPath,
		"fileCount": len(graph.Files),
		"results":   result,
	}

	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "Missing query parameter 'q'"})
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		path = "."
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "invalid path: %v"}`, err), 400)
		return
	}

	opts := find.SearchOptions{
		CaseSensitive: r.URL.Query().Get("case-sensitive") == "true",
		UseRegex:      r.URL.Query().Get("regex") == "true",
		MaxResults:    100,
	}

	results, err := find.Search(absPath, query, opts)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "search failed: %v"}`, err), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"query":   query,
		"path":    absPath,
		"results": results,
		"count":   len(results),
	})
}

func (s *Server) handleFile(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/files/")
	if path == "" {
		http.NotFound(w, r)
		return
	}

	content, err := os.ReadFile(path)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write(content)
}

func (s *Server) handleGraph(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := r.URL.Query().Get("path")
	if path == "" {
		path = "."
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "invalid path: %v"}`, err), 400)
		return
	}

	graph, err := analyze.BuildGraph(absPath, "go")
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "failed to build graph: %v"}`, err), 500)
		return
	}

	result := map[string]interface{}{
		"path":         absPath,
		"fileCount":    len(graph.Files),
		"exportCount":  len(graph.Exports),
		"files":        getFileList(graph),
		"dependencies": getDependencies(graph),
	}

	json.NewEncoder(w).Encode(result)
}

func (s *Server) handleSymbolGraph(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		path = "."
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		http.Error(w, `{"error":"invalid repository path"}`, http.StatusBadRequest)
		return
	}
	graph, err := codegraph.Build(absPath)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(graph); err != nil {
		fmt.Fprintf(os.Stderr, "write graph response: %v\n", err)
	}
}

func getFileList(graph *analyze.ModuleGraph) []map[string]interface{} {
	var files []map[string]interface{}
	for path, node := range graph.Files {
		exports := make([]map[string]string, len(node.Exports))
		for i, exp := range node.Exports {
			exports[i] = map[string]string{
				"name": exp.Name,
				"kind": exp.Kind,
			}
		}
		files = append(files, map[string]interface{}{
			"path":    path,
			"exports": exports,
		})
	}
	return files
}

func getDependencies(graph *analyze.ModuleGraph) []map[string]interface{} {
	var deps []map[string]interface{}
	for file, imports := range graph.Imports {
		for _, imp := range imports {
			deps = append(deps, map[string]interface{}{
				"file":   file,
				"import": imp.Name,
				"source": imp.Source,
			})
		}
	}
	return deps
}

var htmlIndex = `
<!DOCTYPE html>
<html>
<head>
<title>Code Cartographer</title>
<style>
body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; background: #f5f5f5; }
.card { background: white; border-radius: 8px; padding: 30px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
h1 { color: #333; }
.commands { background: #f0f0f0; padding: 15px; border-radius: 4px; font-family: monospace; margin: 20px 0; }
.link { display: inline-block; margin: 5px; padding: 8px 16px; background: #007bff; color: white; text-decoration: none; border-radius: 4px; }
.link:hover { background: #0056b3; }
</style>
</head>
<body>
<div class="card">
<h1>Code Cartographer v2.0</h1>
<p>Fast, CLI-first codebase analyzer and visualizer.</p>
<div class="commands">
# Map a codebase<br>
carto map ./my-project<br><br>
# Analyze for issues<br>
carto analyze<br><br>
# Search code<br>
carto find "functionName"<br><br>
# Start web UI<br>
carto serve
</div>
<p>For more information, see the CLI commands above.</p>
</div>
</body>
</html>
`
