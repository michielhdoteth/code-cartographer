package main

import (
	"encoding/json"
	"fmt"
	"os"

	"carto/codegraph"
	"carto/exitcode"

	"github.com/spf13/cobra"
)

var askCmd = &cobra.Command{
	Use:   "ask <question> [path]",
	Short: "Find relevant code for a question (local and deterministic)",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		root := "."
		if len(args) > 1 {
			root = args[1]
		}
		limit, _ := cmd.Flags().GetInt("limit")
		tokens, _ := cmd.Flags().GetInt("tokens")
		tokens = tokenBudgetFor(cmd, root, tokens)
		index, err := codegraph.Build(root)
		if err != nil {
			return err
		}
		hits := index.Ask(args[0], limit, tokens)
		answerOnly, _ := cmd.Flags().GetBool("answer-only")
		if answerOnly {
			facts := make([]map[string]interface{}, 0, len(hits))
			for _, hit := range hits {
				callers, callees := 0, 0
				for _, edge := range index.Edges {
					if edge.Type != "calls" {
						continue
					}
					if edge.Target == hit.Node.ID {
						callers++
					}
					if edge.Source == hit.Node.ID {
						callees++
					}
				}
				facts = append(facts, map[string]interface{}{
					"name": hit.Node.Name, "kind": hit.Node.Kind, "file": hit.Node.File,
					"line": hit.Node.StartLine, "callers": callers, "callees": callees,
					"flags": hit.Node.Flags,
				})
			}
			if jsonOut {
				out, _ := json.MarshalIndent(map[string]interface{}{"query": args[0], "facts": facts}, "", "  ")
				fmt.Println(string(out))
			} else {
				for _, fact := range facts {
					fmt.Printf("%s %s:%v (callers %v, callees %v)\n", fact["name"], fact["file"], fact["line"], fact["callers"], fact["callees"])
				}
			}
			return nil
		}
		if jsonOut {
			used := 0
			for _, hit := range hits {
				used += codegraph.EstimateTokens(hit.Snippet)
			}
			data, e := json.MarshalIndent(map[string]interface{}{
				"query": args[0], "hits": hits, "warnings": index.Warnings,
				"savings": index.EstimateSavings(used),
			}, "", "  ")
			if e != nil {
				return e
			}
			fmt.Println(string(data))
			return nil
		}
		if len(hits) == 0 {
			fmt.Println("No relevant code found.")
			return nil
		}
		for i, hit := range hits {
			fmt.Printf("%d. %s %s:%d (score %.1f)\n", i+1, hit.Node.Name, hit.Node.File, hit.Node.StartLine, hit.Score)
			if hit.Snippet != "" {
				fmt.Println(hit.Snippet)
			}
			fmt.Println()
		}
		return nil
	},
}

func init() {
	askCmd.Flags().Int("limit", 8, "Maximum number of code results")
	askCmd.Flags().IntP("tokens", "t", 4000, "Approximate token budget for snippets")
	askCmd.Flags().Bool("answer-only", false, "Return graph facts without source snippets")
}

var traceCmd = &cobra.Command{
	Use:   "trace <symbol> [path]",
	Short: "Show callers and callees for a symbol",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		root := "."
		if len(args) > 1 {
			root = args[1]
		}
		depth, _ := cmd.Flags().GetInt("depth")
		direction, _ := cmd.Flags().GetString("direction")
		if direction != "callers" && direction != "callees" && direction != "both" {
			return fmt.Errorf("direction must be callers, callees, or both")
		}
		index, err := codegraph.Build(root)
		if err != nil {
			return err
		}
		hits := index.TraceDepth(args[0], depth, direction)
		if jsonOut {
			data, e := json.MarshalIndent(hits, "", "  ")
			if e != nil {
				return e
			}
			fmt.Println(string(data))
			return nil
		}
		for _, hit := range hits {
			fmt.Printf("%-7s depth %-2d %-10s %s (%s", hit.Direction, hit.Depth, hit.Node.Kind, hit.Node.File, hit.Confidence)
			if hit.Line > 0 {
				fmt.Printf(", line %d", hit.Line)
			}
			fmt.Println(")")
		}
		return nil
	},
}

func init() {
	traceCmd.Flags().Int("depth", 1, "Maximum call-graph hops")
	traceCmd.Flags().String("direction", "both", "Traversal direction: callers, callees, or both")
}

var outlineCmd = &cobra.Command{
	Use:   "outline <file> [path]",
	Short: "Show a file's symbol API and source ranges",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		root := "."
		if len(args) > 1 {
			root = args[1]
		}
		index, err := codegraph.Build(root)
		if err != nil {
			return err
		}
		nodes := index.Outline(args[0])
		if jsonOut {
			data, e := json.MarshalIndent(nodes, "", "  ")
			if e != nil {
				return e
			}
			fmt.Println(string(data))
			return nil
		}
		for _, node := range nodes {
			label := node.Signature
			if label == "" {
				label = node.Name
			}
			fmt.Printf("%5d-%-5d %-12s %s\n", node.StartLine, node.EndLine, node.Kind, label)
		}
		return nil
	},
}

var checkCmd = &cobra.Command{
	Use:   "check [path]",
	Short: "Check whether the cached code graph matches the working tree",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root := "."
		if len(args) > 0 {
			root = args[0]
		}
		freshness, err := codegraph.CheckFreshness(root)
		if err != nil {
			return err
		}
		if jsonOut {
			data, _ := json.MarshalIndent(freshness, "", "  ")
			fmt.Println(string(data))
		} else if freshness.Fresh {
			fmt.Println("Code graph is fresh.")
		} else {
			fmt.Printf("Code graph is stale: %d added, %d changed, %d deleted. Run 'carto map %s' to refresh.\n",
				len(freshness.Added), len(freshness.Changed), len(freshness.Deleted), root)
		}
		if !freshness.Fresh {
			return fmt.Errorf("%w: run 'carto map %s'", exitcode.ErrStaleGraph, root)
		}
		return nil
	},
}

var repoMapCmd = &cobra.Command{
	Use:   "repo-map [path]",
	Short: "Generate a ranked, token-budgeted repository signature map",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root := "."
		if len(args) > 0 {
			root = args[0]
		}
		budget, _ := cmd.Flags().GetInt("tokens")
		budget = tokenBudgetFor(cmd, root, budget)
		focus, _ := cmd.Flags().GetStringArray("focus")
		expand, _ := cmd.Flags().GetBool("expand")
		if expand && len(focus) == 0 {
			budget *= 2
		}
		index, err := codegraph.Build(root)
		if err != nil {
			return err
		}
		result := index.GenerateRepoMap(budget, focus, true)
		if jsonOut {
			data, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
		} else {
			fmt.Print(result.RenderMarkdown())
		}
		return nil
	},
}

func init() {
	repoMapCmd.Flags().IntP("tokens", "t", 2048, "Estimated token budget for the map")
	repoMapCmd.Flags().StringArray("focus", nil, "Boost a path or directory in ranking; repeatable")
	repoMapCmd.Flags().Bool("expand", false, "Double the map budget when no --focus is given")
}

var packCmd = &cobra.Command{
	Use:   "pack [path]",
	Short: "Pack filtered repository context within an estimated token budget",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root := "."
		if len(args) > 0 {
			root = args[0]
		}
		budget, _ := cmd.Flags().GetInt("tokens")
		budget = tokenBudgetFor(cmd, root, budget)
		include, _ := cmd.Flags().GetStringArray("include")
		exclude, _ := cmd.Flags().GetStringArray("exclude")
		compress, _ := cmd.Flags().GetBool("compress")
		outputPath, _ := cmd.Flags().GetString("output")
		index, err := codegraph.Build(root)
		if err != nil {
			return err
		}
		result := index.GeneratePack(codegraph.PackOptions{TokenBudget: budget, Include: include, Exclude: exclude, Compress: compress})
		if outputPath != "" {
			if jsonOut {
				return fmt.Errorf("--json cannot be combined with --output")
			}
			if err := os.WriteFile(outputPath, []byte(result.Content), 0644); err != nil {
				return fmt.Errorf("write pack output: %w", err)
			}
			fmt.Fprintf(os.Stderr, "Packed %d files (%d estimated tokens) to %s\n", result.FilesIncluded, result.TokensUsed, outputPath)
			return nil
		}
		if jsonLines {
			emitJSONL("begin", map[string]interface{}{"files": result.FilesIncluded, "tokens": result.TokensUsed})
			emitJSONL("content", map[string]interface{}{"text": result.Content})
			emitJSONL("summary", map[string]interface{}{"summary": result.Summary, "tree": result.Tree, "token_ledger": result.TokenLedger})
		} else if jsonOut {
			data, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
		} else {
			fmt.Print(result.Content)
		}
		if packFailOnOverBudget && result.TokensUsed > budget {
			return fmt.Errorf("%w: pack used %d estimated tokens, budget is %d", exitcode.ErrFindings, result.TokensUsed, budget)
		}
		return nil
	},
}

func init() {
	packCmd.Flags().IntP("tokens", "t", 8000, "Estimated token budget")
	packCmd.Flags().BoolVar(&packFailOnOverBudget, "fail-on-over-budget", false, "Exit 1 when the pack exceeds its token budget")
	packCmd.Flags().StringArray("include", nil, "Include matching path glob; repeatable")
	packCmd.Flags().StringArray("exclude", nil, "Exclude matching path glob; repeatable")
	packCmd.Flags().Bool("compress", false, "Emit declarations and imports without implementation bodies")
	packCmd.Flags().StringP("output", "o", "", "Write Markdown context to a file instead of stdout")
}
