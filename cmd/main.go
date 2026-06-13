package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"carto/analyze"
	"carto/find"
	"carto/mcp"
	"carto/serve"

	"github.com/spf13/cobra"
)

var verbose bool
var jsonOut bool
var langFilter string
var analyzeType string
var caseSensitive bool
var useRegex bool
var servePort string
var serveOpen bool

func main() {
	rootCmd := &cobra.Command{
		Use:   "carto",
		Short: "carto - Map, analyze, and visualize codebases",
		Long: `carto - Fast codebase analyzer and visualizer

Zero-config CLI for mapping code structure, detecting patterns,
and visualizing your codebase.

Examples:
  carto map ./my-project       # Scan and map a codebase
  carto analyze               # Analyze mapped codebase
  carto find "function_name"  # Search code
  carto serve                 # Start web UI`,
		Version: "0.1.0",
	}

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().BoolVar(&jsonOut, "json", false, "JSON output")

	rootCmd.AddCommand(mapCmd)
	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(findCmd)
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(mcpCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// MAP COMMAND
var mapCmd = &cobra.Command{
	Use:   "map <path>",
	Short: "Map a codebase - scan files and structure",
	Long: `Scan a directory and map its code structure.

Finds all source files, extracts language, and builds
a module graph for analysis and visualization.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("path does not exist: %s", path)
			}
			return err
		}
		if !info.IsDir() {
			return fmt.Errorf("path is not a directory: %s", path)
		}

		absPath, _ := filepath.Abs(path)
		fmt.Printf("Scanning %s...\n", absPath)

		var wantedLangs []string
		if langFilter != "" {
			wantedLangs = strings.Split(langFilter, ",")
		}

		files := []string{}
		count := 0

		filepath.Walk(absPath, func(p string, info os.FileInfo, err error) error {
			if err != nil { return err }
			
			name := info.Name()
			if info.IsDir() {
				if name == "node_modules" || name == ".git" || name == "vendor" ||
				   name == "dist" || name == "build" || name == "__pycache__" ||
				   name == "target" || strings.HasPrefix(name, ".") {
					return filepath.SkipDir
				}
				return nil
			}

			lang := detectLanguage(name)
			if lang == "" { return nil }

			if len(wantedLangs) > 0 {
				found := false
				for _, l := range wantedLangs {
					if strings.ToLower(l) == strings.ToLower(lang) { found = true }
				}
				if !found { return nil }
			}

			relPath, _ := filepath.Rel(absPath, p)
			files = append(files, relPath)
			count++
			return nil
		})

		byLang := map[string]int{}
		for _, f := range files {
			lang := detectLanguage(filepath.Ext(f))
			byLang[lang]++
		}

		fmt.Printf("\nScanned %d files in %.0fms\n", count, 0.0)
		fmt.Println("\nLanguages:")
		for lang, c := range byLang {
			fmt.Printf("  %s: %d files\n", lang, c)
		}

		return nil
	},
}

func init() {
	mapCmd.Flags().StringVar(&langFilter, "lang", "", "Languages to scan (comma-separated)")
}

// ANALYZE COMMAND
var analyzeCmd = &cobra.Command{
	Use:   "analyze [type] [path]",
	Short: "Analyze mapped codebase",
	Long: `Analyze the current codebase for patterns and issues.

Analysis types:
  dead-code   - Unused files, exports, and dependencies
  complexity - Functions with high cyclomatic complexity
  circular   - Circular dependencies
  health     - Overall code health metrics
  all        - Run all analyses (default)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 1 {
			path = args[1]
		} else if len(args) > 0 && args[0] != "all" && args[0] != "dead-code" && args[0] != "complexity" && args[0] != "circular" && args[0] != "health" {
			path = args[0]
		}

		analysisType := analyzeType
		if len(args) > 0 {
			if args[0] == "dead-code" || args[0] == "complexity" || args[0] == "circular" || args[0] == "health" || args[0] == "all" {
				analysisType = args[0]
			}
		}

		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("invalid path: %w", err)
		}

		info, err := os.Stat(absPath)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("path does not exist: %s", absPath)
			}
			return err
		}
		if !info.IsDir() {
			return fmt.Errorf("path is not a directory: %s", absPath)
		}

		fmt.Printf("Analyzing %s...\n", absPath)
		start := time.Now()

		// Build the module graph
		graph, err := analyze.BuildGraph(absPath, "go")
		if err != nil {
			return fmt.Errorf("failed to build graph: %w", err)
		}

		elapsed := time.Since(start).Milliseconds()

		// Run the requested analysis
		if jsonOut {
			result := runAnalysis(graph, analysisType)
			output := map[string]interface{}{
				"type":     analysisType,
				"path":     absPath,
				"duration": elapsed,
				"results":  result,
			}
			jsonBytes, _ := json.MarshalIndent(output, "", "  ")
			fmt.Println(string(jsonBytes))
		} else {
			printAnalysisResults(graph, analysisType, elapsed)
		}
		return nil
	},
}

func init() {
	analyzeCmd.Flags().StringVar(&analyzeType, "type", "all", "Analysis type")
}

// FIND COMMAND
var findCmd = &cobra.Command{
	Use:   "find <pattern> [path]",
	Short: "Search code in mapped files",
	Long: `Search for text or patterns in the codebase.

Searches through all mapped files for the given pattern.
Uses regex if --regex is specified.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("pattern required")
		}

		pattern := args[0]
		path := "."
		if len(args) > 1 {
			path = args[1]
		}

		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("invalid path: %w", err)
		}

		info, err := os.Stat(absPath)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("path does not exist: %s", absPath)
			}
			return err
		}
		if !info.IsDir() {
			return fmt.Errorf("path is not a directory: %s", absPath)
		}

		fmt.Printf("Searching for: %s\n", pattern)
		start := time.Now()

		opts := find.SearchOptions{
			CaseSensitive: caseSensitive,
			UseRegex:      useRegex,
			MaxResults:    100,
		}

		results, err := find.Search(absPath, pattern, opts)
		if err != nil {
			return fmt.Errorf("search failed: %w", err)
		}

		elapsed := time.Since(start).Milliseconds()

		if jsonOut {
			output := map[string]interface{}{
				"pattern":  pattern,
				"path":     absPath,
				"duration": elapsed,
				"results":  results,
			}
			jsonBytes, _ := json.MarshalIndent(output, "", "  ")
			fmt.Println(string(jsonBytes))
		} else {
			fmt.Printf("\nFound %d matches in %dms\n", len(results), elapsed)
			fmt.Println(find.FormatResults(results, 50))
		}
		return nil
	},
}

func init() {
	findCmd.Flags().BoolVar(&caseSensitive, "case-sensitive", false, "Case sensitive search")
	findCmd.Flags().BoolVar(&useRegex, "regex", false, "Use regex pattern")
}

// SERVE COMMAND
var serveCmd = &cobra.Command{
	Use:   "serve [path]",
	Short: "Start interactive web UI",
	Long: `Start a local web server for interactive visualization.
 Opens a browser-based UI for exploring your codebase.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		buildDir := "./build"
		if len(args) > 0 {
			buildDir = args[0]
		}

		server := serve.NewServer(servePort, buildDir)
		return server.Run()
	},
}

func init() {
	serveCmd.Flags().StringVarP(&servePort, "port", "p", "3000", "Port to serve on")
	serveCmd.Flags().BoolVar(&serveOpen, "open", true, "Open browser automatically")
}

// INFO COMMAND
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show current map info",
	Long: `Show information about the current mapped codebase.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Code Cartographer v2.0.0")
		fmt.Println("")
		fmt.Println("Commands:")
		fmt.Println("  carto map <path>       - Scan and map a codebase")
		fmt.Println("  carto analyze [type]   - Analyze for dead code, complexity, etc.")
		fmt.Println("  carto find <pattern>   - Search code in mapped files")
		fmt.Println("  carto serve            - Start interactive web UI")
		fmt.Println("  carto mcp              - Start MCP server for Claude")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  carto map ./my-project")
		fmt.Println("  carto analyze dead-code ./my-project")
		fmt.Println("  carto find 'functionName' ./my-project")
		return nil
	},
}

// MCP COMMAND
var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP server for Claude integration",
	Long: `Start the Model Context Protocol server for Claude Desktop integration.

This enables Claude to use Code Cartographer tools directly.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		server := mcp.NewServer()
		return server.Run(context.Background())
	},
}

// HELPER FUNCTIONS

func runAnalysis(graph *analyze.ModuleGraph, analysisType string) interface{} {
	switch analysisType {
	case "dead-code":
		return analyze.AnalyzeDeadCode(graph)
	case "complexity":
		return analyze.AnalyzeComplexity(graph)
	case "circular":
		return analyze.AnalyzeCircularDeps(graph)
	case "health":
		return analyze.CalculateHealth(graph)
	case "all":
		return map[string]interface{}{
			"dead-code":  analyze.AnalyzeDeadCode(graph),
			"complexity": analyze.AnalyzeComplexity(graph),
			"circular":   analyze.AnalyzeCircularDeps(graph),
			"health":     analyze.CalculateHealth(graph),
		}
	default:
		return analyze.AnalyzeDeadCode(graph)
	}
}

func printAnalysisResults(graph *analyze.ModuleGraph, analysisType string, elapsed int64) {
	fmt.Printf("\nAnalysis completed in %dms\n\n", elapsed)

	switch analysisType {
	case "dead-code":
		result := analyze.AnalyzeDeadCode(graph)
		fmt.Print(result.String())
	case "complexity":
		result := analyze.AnalyzeComplexity(graph)
		fmt.Print(result.String())
	case "circular":
		result := analyze.AnalyzeCircularDeps(graph)
		fmt.Print(result.String())
	case "health":
		result := analyze.CalculateHealth(graph)
		fmt.Print(result.String())
	case "all":
		fmt.Println("=== Dead Code Analysis ===")
		result := analyze.AnalyzeDeadCode(graph)
		fmt.Print(result.String())

		fmt.Println("\n=== Complexity Analysis ===")
		compResult := analyze.AnalyzeComplexity(graph)
		fmt.Print(compResult.String())

		fmt.Println("\n=== Circular Dependencies ===")
		circResult := analyze.AnalyzeCircularDeps(graph)
		fmt.Print(circResult.String())

		fmt.Println("\n=== Overall Health ===")
		healthResult := analyze.CalculateHealth(graph)
		fmt.Print(healthResult.String())
	}
}

// HELPERS
var langExtensions = map[string][]string{
	"typescript": {".ts", ".tsx", ".mts", ".cts"},
	"javascript": {".js", ".jsx", ".mjs", ".cjs"},
	"python":     {".py", ".pyi"},
	"rust":      {".rs"},
	"go":        {".go"},
	"java":      {".java"},
	"cpp":       {".cpp", ".cc", ".cxx", ".h", ".hpp"},
	"ruby":      {".rb"},
	"php":       {".php"},
}

func detectLanguage(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	for lang, exts := range langExtensions {
		for _, e := range exts {
			if ext == e { return lang }
		}
	}
	return ""
}