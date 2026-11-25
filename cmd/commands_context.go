package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"carto/analyze"
	"carto/codegraph"
	"carto/exitcode"
	"carto/fix"
	"carto/search"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <query> [path]",
	Short: "Search the local full-text index with BM25F ranking",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		root := "."
		if len(args) > 1 {
			root = args[1]
		}
		limit, _ := cmd.Flags().GetInt("limit")
		freshness, err := codegraph.CheckFreshness(root)
		if err == nil && !freshness.Fresh {
			return fmt.Errorf("%w: run 'carto map %s' then 'carto index %s'", exitcode.ErrStaleGraph, root, root)
		}
		index, err := search.LoadIndex(root)
		if err != nil {
			return fmt.Errorf("no search index found; run 'carto index %s': %w", root, err)
		}
		results := index.Search(args[0], limit)
		if jsonOut {
			out, _ := json.MarshalIndent(map[string]interface{}{"query": args[0], "path": root, "results": results}, "", "  ")
			fmt.Println(string(out))
		} else {
			fmt.Print(search.FormatResults(results, args[0]))
		}
		return nil
	},
}

func init() {
	searchCmd.Flags().Int("limit", 8, "Maximum number of files to return")
}

var briefCmd = &cobra.Command{
	Use:   "brief [path]",
	Short: "Compact token-budgeted repository orientation",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root := "."
		if len(args) > 0 {
			root = args[0]
		}
		budget, _ := cmd.Flags().GetInt("tokens")
		maxFiles, _ := cmd.Flags().GetInt("max-files")
		focus, _ := cmd.Flags().GetStringArray("focus")
		index, err := codegraph.Build(root)
		if err != nil {
			return err
		}
		brief := index.GenerateBrief(budget, maxFiles, focus)
		if jsonOut {
			out, _ := json.MarshalIndent(brief, "", "  ")
			fmt.Println(string(out))
		} else {
			fmt.Print(brief.RenderMarkdown())
		}
		return nil
	},
}

func init() {
	briefCmd.Flags().IntP("tokens", "t", 1200, "Estimated token budget")
	briefCmd.Flags().Int("max-files", 30, "Maximum files in the inventory")
	briefCmd.Flags().StringArray("focus", nil, "Boost a path in the repo map; repeatable")
}

var chunkCmd = &cobra.Command{
	Use:   "chunk <file> [path]",
	Short: "List or fetch AST-aligned chunks of a large file",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		root := "."
		if len(args) > 1 {
			root = args[1]
		}
		budget, _ := cmd.Flags().GetInt("budget")
		chunkIndex, _ := cmd.Flags().GetInt("chunk")
		list, _ := cmd.Flags().GetBool("list")
		maxChunks, _ := cmd.Flags().GetInt("max-chunks")
		index, err := codegraph.Build(root)
		if err != nil {
			return err
		}
		if list {
			chunkIndex = -1
		}
		manifest, err := index.BuildChunkManifest(args[0], budget)
		if err != nil {
			return err
		}
		if chunkIndex < 0 {
			if jsonOut {
				out, _ := json.MarshalIndent(manifest, "", "  ")
				fmt.Println(string(out))
			} else {
				fmt.Printf("=== chunks %s ===\n%d chunks · %d lines\n\n", manifest.FilePath, manifest.TotalChunks, manifest.TotalLines)
				for _, c := range manifest.Chunks {
					fmt.Printf("  %2d  lines %5d-%-5d %4d tokens  %s\n", c.ChunkIndex, c.LineStart, c.LineEnd, c.Tokens, strings.Join(c.Symbols, ", "))
				}
				fmt.Printf("\nnext: %s\n", manifest.SuggestedNext)
			}
			return nil
		}
		chunk, err := index.GenerateChunk(args[0], chunkIndex, budget, maxChunks)
		if err != nil {
			return err
		}
		if jsonOut {
			out, _ := json.MarshalIndent(chunk, "", "  ")
			fmt.Println(string(out))
		} else {
			fmt.Printf("=== chunk %d/%d %s ===\nlines %d-%d · %d tokens\n\n%s\n",
				chunk.ChunkIndex+1, manifest.TotalChunks, chunk.ChunkID,
				chunk.LineStart, chunk.LineEnd, chunk.Tokens, chunk.Content)
			if chunk.NextChunk >= 0 {
				fmt.Printf("\nnext: %s\n", chunk.SuggestedNext)
			}
		}
		return nil
	},
}

func init() {
	chunkCmd.Flags().IntP("budget", "b", 400, "Token budget per chunk")
	chunkCmd.Flags().IntP("chunk", "n", -1, "Chunk index to fetch (-1 lists chunks)")
	chunkCmd.Flags().Bool("list", false, "List chunk metadata only")
	chunkCmd.Flags().Int("max-chunks", 64, "Reject files with more chunks than this")
}

var fixCmd = &cobra.Command{
	Use:   "fix <finding-id> [path]",
	Short: "Generate an LLM-ready refactor prompt for a structural finding",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		root := "."
		if len(args) > 1 {
			root = args[1]
		}
		absRoot, err := filepath.Abs(root)
		if err != nil {
			return err
		}
		budget, _ := cmd.Flags().GetInt("tokens")
		graph, err := analyze.BuildGraph(absRoot, "")
		if err != nil {
			return fmt.Errorf("failed to build graph: %w", err)
		}
		report := analyze.AnalyzeStructure(graph, currentStructureConfig(absRoot))
		var found *analyze.StructureFinding
		for i := range report.Findings {
			if report.Findings[i].ID == args[0] || strings.HasPrefix(report.Findings[i].ID, args[0]) {
				found = &report.Findings[i]
				break
			}
		}
		if found == nil {
			return fmt.Errorf("finding %q not found; run 'carto analyze structure' to list finding IDs", args[0])
		}
		prompt, err := fix.GeneratePrompt(absRoot, *found, budget)
		if err != nil {
			return err
		}
		if jsonOut {
			out, _ := json.MarshalIndent(map[string]interface{}{"finding_id": found.ID, "rule": found.Rule, "prompt": prompt}, "", "  ")
			fmt.Println(string(out))
		} else {
			fmt.Print(prompt)
		}
		return nil
	},
}

func init() {
	fixCmd.Flags().IntP("tokens", "t", 4000, "Approximate token budget for the prompt")
}
