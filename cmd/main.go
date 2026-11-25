package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"carto/analyze"
	"carto/context"
	"carto/exitcode"
	"carto/find"
	"carto/git"
	"carto/lib/parser"
	"carto/manifest"
	"carto/serve"

	"github.com/spf13/cobra"
)

var jsonOut bool
var jsonLines bool
var langFilter string
var analyzeType string
var caseSensitive bool
var useRegex bool
var servePort string
var includeStructure bool
var structureFailOn string
var structureAllowPath string
var packFailOnOverBudget bool

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
		Version: "1.0.0",
	}

	rootCmd.PersistentFlags().BoolVar(&jsonOut, "json", false, "JSON output")
	rootCmd.PersistentFlags().BoolVar(&jsonLines, "jsonl", false, "Stream JSON Lines events")

	rootCmd.AddCommand(mapCmd)
	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(findCmd)
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(diffCmd)
	rootCmd.AddCommand(changedCmd)
	rootCmd.AddCommand(impactCmd)
	rootCmd.AddCommand(contextCmd)
	rootCmd.AddCommand(symbolsCmd)
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(indexCmd)
	rootCmd.AddCommand(askCmd)
	rootCmd.AddCommand(traceCmd)
	rootCmd.AddCommand(outlineCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(repoMapCmd)
	rootCmd.AddCommand(packCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(fixCmd)
	rootCmd.AddCommand(briefCmd)
	rootCmd.AddCommand(chunkCmd)
	rootCmd.AddCommand(skillCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(schemaCmd)
	rootCmd.AddCommand(initCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exitcode.Code(err))
	}
}

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
		start := time.Now()

		var wantedLangs []string
		if langFilter != "" {
			wantedLangs = strings.Split(langFilter, ",")
		}

		files := []string{}

		if err := filepath.Walk(absPath, func(p string, info os.FileInfo, err error) error {
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

			lang := string(parser.DetectLanguage(name))
			if lang == "" || lang == "unknown" { return nil }

			if len(wantedLangs) > 0 {
				found := false
				for _, l := range wantedLangs {
					if strings.ToLower(l) == strings.ToLower(lang) { found = true }
				}
				if !found { return nil }
			}

			relPath, _ := filepath.Rel(absPath, p)
			files = append(files, relPath)
			return nil
		}); err != nil {
			return fmt.Errorf("walk project: %w", err)
		}

		byLang := map[string]int{}
		for _, f := range files {
			lang := string(parser.DetectLanguage(filepath.Ext(f)))
			byLang[lang]++
		}

		fmt.Printf("\nScanned %d files in %dms\n", len(files), time.Since(start).Milliseconds())
		fmt.Println("\nLanguages:")
		for lang, c := range byLang {
			fmt.Printf("  %s: %d files\n", lang, c)
		}

		if git.IsRepo(absPath) {
			sha, _ := git.GetHeadSHA(absPath)
			branch, _ := git.GetBranch(absPath)
			m := &manifest.Manifest{
				Repo:      filepath.Base(absPath),
				Commit:    sha,
				Branch:    branch,
				MappedAt:  time.Now(),
				FileCount: len(files),
				Languages: byLang,
			}
			if err := manifest.Save(absPath, m); err == nil {
				fmt.Printf("\nMap saved for commit %s (%s)\n", sha[:8], branch)
			}
		}

		return nil
	},
}

func init() {
	mapCmd.Flags().StringVar(&langFilter, "lang", "", "Languages to scan (comma-separated)")
}

var analyzeCmd = &cobra.Command{
	Use:   "analyze [type] [path]",
	Short: "Analyze mapped codebase",
	Long: `Analyze the current codebase for patterns and issues.

Analysis types:
  dead-code   - Unused files, exports, and dependencies
  complexity - Functions with high cyclomatic complexity
  circular   - Circular dependencies
  structure  - Deep modules, narrow interfaces, and context scatter
  health     - Overall code health metrics
  all        - Run all analyses (default)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 1 {
			path = args[1]
		} else if len(args) > 0 && !isAnalysisType(args[0]) {
			path = args[0]
		}

		analysisType := analyzeType
		if len(args) > 0 && isAnalysisType(args[0]) {
			analysisType = args[0]
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

		if !jsonOut {
			fmt.Printf("Analyzing %s...\n", absPath)
		}
		start := time.Now()

		graph, err := analyze.BuildGraph(absPath, "go")
		if err != nil {
			return fmt.Errorf("failed to build graph: %w", err)
		}

		elapsed := time.Since(start).Milliseconds()

		var structureReport *analyze.StructureReport
		if analysisType == "structure" || analysisType == "all" || includeStructure {
			structureReport = analyze.AnalyzeStructure(graph, currentStructureConfig(absPath))
		}

		if jsonOut {
			result := runAnalysis(graph, analysisType, structureReport)
			output := map[string]interface{}{
				"type":     analysisType,
				"path":     absPath,
				"duration": elapsed,
				"results":  result,
			}
			jsonBytes, _ := json.MarshalIndent(output, "", "  ")
			fmt.Println(string(jsonBytes))
		} else {
			printAnalysisResults(graph, analysisType, elapsed, structureReport)
		}
		if structureReport != nil && structureFailOn != "never" {
			count := structureReport.Summary.Critical
			if structureFailOn == "warning" {
				count += structureReport.Summary.Warning
			}
			if count > 0 {
				return fmt.Errorf("%w: structure analysis found %d issue(s) at or above '%s'", exitcode.ErrFindings, count, structureFailOn)
			}
		}
		return nil
	},
}

func init() {
	analyzeCmd.Flags().StringVar(&analyzeType, "type", "all", "Analysis type")
	analyzeCmd.Flags().StringVar(&structureFailOn, "fail-on", "never", "Exit non-zero when structure findings reach this tier")
	analyzeCmd.Flags().StringVar(&structureAllowPath, "allow", "", "Path to a structure allowlist")
	analyzeCmd.Flags().BoolVar(&includeStructure, "include-structure", false, "Include structural findings in the health score")
}

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
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show current map info",
	Long: `Show information about the current mapped codebase.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Code Cartographer v1.0.0")
		fmt.Println("")
		fmt.Println("Commands:")
		fmt.Println("  carto map <path>       - Scan and map a codebase")
		fmt.Println("  carto analyze [type]   - Analyze for dead code, complexity, etc.")
		fmt.Println("  carto find <pattern>   - Search code in mapped files")
		fmt.Println("  carto serve            - Start interactive web UI")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  carto map ./my-project")
		fmt.Println("  carto analyze dead-code ./my-project")
		fmt.Println("  carto find 'functionName' ./my-project")
		return nil
	},
}

var diffCmd = &cobra.Command{
	Use:   "diff [ref]",
	Short: "Show changes since last map",
	Long: `Compare current state against the last mapped commit.
Shows which files have changed and groups them by language.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}
		absPath, _ := filepath.Abs(path)

		if !git.IsRepo(absPath) {
			return fmt.Errorf("not a git repository: %s", absPath)
		}

		m, err := manifest.Load(absPath)
		if err != nil {
			return fmt.Errorf("no manifest found. Run 'carto map' first")
		}

		currentSHA, _ := git.GetHeadSHA(absPath)
		if m.Commit == currentSHA {
			fmt.Println("No changes since last map.")
			return nil
		}

		files, err := git.GetChangedFiles(absPath, m.Commit, currentSHA)
		if err != nil {
			return fmt.Errorf("failed to get changes: %w", err)
		}

		if len(files) == 0 {
			fmt.Println("No file changes detected since last map.")
			return nil
		}

		byLang := map[string][]string{}
		for _, f := range files {
			lang := string(parser.DetectLanguage(filepath.Ext(f)))
			if lang == "" {
				lang = "unknown"
			}
			byLang[lang] = append(byLang[lang], f)
		}

		fmt.Printf("Changed %d files since %s\n\n", len(files), m.Commit[:8])
		for lang, fns := range byLang {
			fmt.Printf("[%s] (%d files)\n", lang, len(fns))
			for _, f := range fns {
				fmt.Printf("  %s\n", f)
			}
		}
		return nil
	},
}

var changedCmd = &cobra.Command{
	Use:   "changed",
	Short: "Show current uncommitted changes",
	Long: `Show files that have been modified, added, or deleted
but not yet committed.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}
		absPath, _ := filepath.Abs(path)

		if !git.IsRepo(absPath) {
			return fmt.Errorf("not a git repository: %s", absPath)
		}

		statuses, err := git.GetStatus(absPath)
		if err != nil {
			return fmt.Errorf("failed to get status: %w", err)
		}

		if len(statuses) == 0 {
			fmt.Println("Working tree clean.")
			return nil
		}

		added := []string{}
		modified := []string{}
		deleted := []string{}

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

		fmt.Printf("Changed files: %d\n\n", len(statuses))
		if len(added) > 0 {
			fmt.Printf("Added (%d):\n", len(added))
			for _, f := range added {
				fmt.Printf("  + %s\n", f)
			}
		}
		if len(modified) > 0 {
			fmt.Printf("Modified (%d):\n", len(modified))
			for _, f := range modified {
				fmt.Printf("  ~ %s\n", f)
			}
		}
		if len(deleted) > 0 {
			fmt.Printf("Deleted (%d):\n", len(deleted))
			for _, f := range deleted {
				fmt.Printf("  - %s\n", f)
			}
		}
		return nil
	},
}

var impactCmd = &cobra.Command{
	Use:   "impact <file>",
	Short: "Analyze blast radius of changing a file",
	Long: `Analyze the transitive impact of changing a file.

Uses BFS traversal to find all files that depend on the given file,
directly or transitively. Shows depth-weighted risk scoring, hub detection,
and concentrated risk analysis for the seed file.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		relFile := args[0]

		absPath, _ := filepath.Abs(path)

		graph, err := analyze.BuildGraph(absPath, "")
		if err != nil {
			return fmt.Errorf("failed to build graph: %w", err)
		}

		var seedFile string
		wanted := filepath.ToSlash(relFile)
		for filePath := range graph.Files {
			normalized := filepath.ToSlash(filePath)
			if normalized == wanted || strings.HasSuffix(normalized, "/"+wanted) {
				seedFile = filePath
				break
			}
		}

		if seedFile == "" {
			if _, err := os.Stat(relFile); err == nil {
				seedFile = relFile
			} else {
				return fmt.Errorf("file not found in codebase: %s", relFile)
			}
		}

		result := analyze.AnalyzeBlastRadius(graph, seedFile)

		if jsonOut {
			output, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(output))
		} else {
			fmt.Print(analyze.FormatBlastRadiusResult(result))
		}

		return nil
	},
}


var contextCmd = &cobra.Command{
	Use:   "context <file>",
	Short: "Get token-budgeted code context",
	Long: `Generate code context within a token budget.

Reads the specified file and optionally includes imports and callers,
all within the specified token limit. Useful for feeding context to
AI agents without exceeding token limits.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		tokens, _ := cmd.Flags().GetInt("tokens")
		rootPath, _ := cmd.Flags().GetString("root")
		tokens = tokenBudgetFor(cmd, rootPath, tokens)
		includeImports, _ := cmd.Flags().GetBool("imports")
		includeCallers, _ := cmd.Flags().GetBool("callers")
		includeCohesion, _ := cmd.Flags().GetBool("cohesion")
		outline, _ := cmd.Flags().GetBool("outline")
		skeletonAlias, _ := cmd.Flags().GetBool("skeleton")
		symbols, _ := cmd.Flags().GetStringArray("symbols")
		lines, _ := cmd.Flags().GetStringArray("lines")
		chunkIndex, _ := cmd.Flags().GetInt("chunk")
		chunkBudget, _ := cmd.Flags().GetInt("chunk-budget")

		absPath, _ := filepath.Abs(path)
		if _, err := os.Stat(absPath); err != nil {
			return fmt.Errorf("file not found: %s", path)
		}

		absRoot, _ := filepath.Abs(rootPath)
		opts := context.Options{
			FilePath:        absPath,
			TokenBudget:     tokens,
			IncludeImports:  includeImports,
			IncludeCallers:  includeCallers,
			IncludeCohesion: includeCohesion,
			RootPath:        absRoot,
			Outline:         outline || skeletonAlias,
			Symbols:         symbols,
			Lines:           lines,
			ChunkIndex:      chunkIndex,
			ChunkBudget:     chunkBudget,
		}

		result, err := context.Generate(opts)
		if err != nil {
			return fmt.Errorf("failed to generate context: %w", err)
		}

		if jsonLines {
			emitJSONL("begin", map[string]interface{}{"file": result.FilePath, "tokens": result.TokensUsed})
			for _, block := range result.Blocks {
				emitJSONL("block", map[string]interface{}{"file": block.FilePath, "type": block.Type, "tokens": block.Tokens, "content": block.Content})
			}
			emitJSONL("summary", map[string]interface{}{"tokens": result.TokensUsed, "savings": result.Savings})
		} else if jsonOut {
			output, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(output))
		} else if result.View != "" {
			fmt.Print(context.FormatView(result))
		} else {
			fmt.Print(context.FormatResult(result))
		}
		return nil
	},
}

func init() {
	contextCmd.Flags().IntP("tokens", "t", 8000, "Token budget")
	contextCmd.Flags().Bool("imports", false, "Include imported files")
	contextCmd.Flags().Bool("callers", false, "Include files that call this file's symbols")
	contextCmd.Flags().Bool("cohesion", false, "Report a structural cohesion score")
	contextCmd.Flags().String("root", ".", "Project root used to resolve imports and callers")
	contextCmd.Flags().Bool("outline", false, "Emit a budgeted AST skeleton")
	contextCmd.Flags().Bool("skeleton", false, "Alias for --outline")
	contextCmd.Flags().StringArray("symbols", nil, "Include only these symbol ranges")
	contextCmd.Flags().StringArray("lines", nil, "Include line range a-b; repeatable")
	contextCmd.Flags().Int("chunk", -1, "Fetch one AST-aligned chunk by index")
	contextCmd.Flags().Int("chunk-budget", 400, "Token budget per chunk")
	contextCmd.MarkFlagsMutuallyExclusive("outline", "skeleton", "chunk")
}

var symbolsCmd = &cobra.Command{
	Use:   "symbols <path>",
	Short: "List all symbols in a file or directory",
	Long: `List all symbols (functions, classes, types, etc.) in the specified path.

Use --type to filter by symbol type (function, class, struct, etc.)`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		symbolType, _ := cmd.Flags().GetString("type")

		absPath, _ := filepath.Abs(path)
		if _, err := os.Stat(absPath); err != nil {
			return fmt.Errorf("path not found: %s", path)
		}

		info, _ := os.Stat(absPath)
		var symbols []analyze.Export

		if !info.IsDir() {
			code, err := os.ReadFile(absPath)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}
			lang := parser.DetectLanguage(absPath)
			p := parser.GetParser(lang)
			if p == nil {
				return fmt.Errorf("unsupported file type: %s", absPath)
			}
			result, err := p.Parse(string(code), absPath)
			if err != nil {
				return fmt.Errorf("failed to parse file: %w", err)
			}
			for _, node := range result.Nodes {
				kind := string(node.Type)
				if symbolType == "" || kind == symbolType {
					symbols = append(symbols, analyze.Export{
						Name: node.Name,
						Kind: kind,
						Line: node.StartLine,
					})
				}
			}
		} else {
			graph, err := analyze.BuildGraph(absPath, "")
			if err != nil {
				return fmt.Errorf("failed to build graph: %w", err)
			}
			if symbolType != "" {
				symbols = graph.GetSymbolsByType(symbolType)
			} else {
				symbols = graph.GetAllSymbols()
			}
		}

		if jsonOut {
			output, _ := json.MarshalIndent(symbols, "", "  ")
			fmt.Println(string(output))
		} else {
			if len(symbols) == 0 {
				fmt.Println("No symbols found.")
				return nil
			}
			fmt.Printf("Found %d symbols:\n\n", len(symbols))
			for _, sym := range symbols {
				fmt.Printf("  %-10s %-30s line %d\n", sym.Kind, sym.Name, sym.Line)
			}
		}
		return nil
	},
}

func init() {
	symbolsCmd.Flags().StringP("type", "t", "", "Filter by type (function, class, struct, method, etc.)")
}

// No duplicate language detection - using parser.DetectLanguage from lib/parser