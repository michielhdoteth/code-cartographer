package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"carto/analyze"
	"carto/search"

	"github.com/spf13/cobra"
)
var statsCmd = &cobra.Command{
	Use:   "stats [path]",
	Short: "Show codebase statistics",
	Long: `Display quick statistics about a codebase including file counts,
language distribution, and symbol counts.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		absPath, _ := filepath.Abs(path)
		if _, err := os.Stat(absPath); err != nil {
			return fmt.Errorf("path not found: %s", path)
		}

		start := time.Now()
		graph, err := analyze.BuildGraph(absPath, "")
		if err != nil {
			return fmt.Errorf("failed to build graph: %w", err)
		}
		elapsed := time.Since(start).Milliseconds()

		stats := graph.GetStats()

		if jsonOut {
			output, _ := json.MarshalIndent(stats, "", "  ")
			fmt.Println(string(output))
		} else {
			fmt.Printf("Codebase Stats (%s)\n", absPath)
			fmt.Printf("Scan time: %dms\n\n", elapsed)

			if totalFiles, ok := stats["total_files"].(int); ok {
				fmt.Printf("Files:     %d\n", totalFiles)
			}
			if totalExports, ok := stats["total_exports"].(int); ok {
				fmt.Printf("Symbols:   %d\n", totalExports)
			}
			if totalImports, ok := stats["total_imports"].(int); ok {
				fmt.Printf("Imports:   %d\n", totalImports)
			}

			if byLang, ok := stats["by_language"].(map[string]int); ok {
				fmt.Println("\nLanguages:")
				for lang, count := range byLang {
					fmt.Printf("  %-15s %d files\n", lang, count)
				}
			}

			if byType, ok := stats["by_type"].(map[string]int); ok {
				fmt.Println("\nSymbols by type:")
				for symType, count := range byType {
					fmt.Printf("  %-15s %d\n", symType, count)
				}
			}
		}
		return nil
	},
}


var indexCmd = &cobra.Command{
	Use:   "index [path]",
	Short: "Build search index for fast full-text search",
	Long: `Build a search index for the codebase.

The index enables instant search across all source files.
Run this after significant changes to refresh the index.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		absPath, _ := filepath.Abs(path)
		if _, err := os.Stat(absPath); err != nil {
			return fmt.Errorf("path not found: %s", path)
		}

		fmt.Printf("Building search index for %s...\n", absPath)
		start := time.Now()

		index, err := search.BuildIndex(absPath)
		if err != nil {
			return fmt.Errorf("failed to build index: %w", err)
		}

		if err := index.Save(absPath); err != nil {
			return fmt.Errorf("failed to save index: %w", err)
		}

		elapsed := time.Since(start).Milliseconds()
		fmt.Printf("Indexed %d files with %d unique terms in %dms\n",
			index.TotalFiles, index.TotalTerms, elapsed)
		fmt.Printf("Index saved to .carto/search-index.json\n")

		return nil
	},
}


