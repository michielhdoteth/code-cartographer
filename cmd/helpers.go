package main

import (
	"encoding/json"
	"fmt"

	"carto/analyze"
	"carto/config"

	"github.com/spf13/cobra"
)

func emitJSONL(event string, data interface{}) {
	line, _ := json.Marshal(map[string]interface{}{"type": event, "data": data})
	fmt.Println(string(line))
}

func isAnalysisType(name string) bool {
	switch name {
	case "dead-code", "complexity", "circular", "structure", "health", "all":
		return true
	default:
		return false
	}
}

func currentStructureConfig(root string) analyze.StructureConfig {
	cfg := analyze.DefaultStructureConfig()
	if loaded, err := config.Load(root); err == nil {
		cfg = loaded.Structure
	}
	if structureAllowPath != "" {
		cfg.AllowPath = structureAllowPath
	}
	return cfg
}

func tokenBudgetFor(cmd *cobra.Command, root string, fallback int) int {
	if cmd.Flags().Changed("tokens") {
		return fallback
	}
	cfg, err := config.Load(root)
	if err != nil {
		return fallback
	}
	switch cmd.Name() {
	case "context":
		return cfg.Defaults.ContextTokens
	case "repo-map":
		return cfg.Defaults.RepoMapTokens
	case "ask":
		return cfg.Defaults.AskTokens
	case "pack", "fix":
		return cfg.Defaults.PackTokens
	default:
		return fallback
	}
}

func runAnalysis(graph *analyze.ModuleGraph, analysisType string, structure *analyze.StructureReport) interface{} {
	switch analysisType {
	case "dead-code":
		return analyze.AnalyzeDeadCode(graph)
	case "complexity":
		return analyze.AnalyzeComplexity(graph)
	case "circular":
		return analyze.AnalyzeCircularDeps(graph)
	case "structure":
		if structure == nil {
			structure = analyze.AnalyzeStructure(graph, currentStructureConfig(graph.RootPath))
		}
		return structure
	case "health":
		return analyze.CalculateHealthWithStructure(graph, structure)
	case "all":
		var healthStructure *analyze.StructureReport
		if includeStructure {
			healthStructure = structure
		}
		return map[string]interface{}{
			"dead-code":  analyze.AnalyzeDeadCode(graph),
			"complexity": analyze.AnalyzeComplexity(graph),
			"circular":   analyze.AnalyzeCircularDeps(graph),
			"structure":  structure,
			"health":     analyze.CalculateHealthWithStructure(graph, healthStructure),
		}
	default:
		return analyze.AnalyzeDeadCode(graph)
	}
}

func printAnalysisResults(graph *analyze.ModuleGraph, analysisType string, elapsed int64, structure *analyze.StructureReport) {
	fmt.Printf("\nAnalysis completed in %dms\n\n", elapsed)
	switch analysisType {
	case "dead-code":
		result := analyze.AnalyzeDeadCode(graph)
		if jsonOut {
			output, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(output))
		} else {
			fmt.Print(analyze.FormatDeadCodeResult(result))
		}
	case "complexity":
		fmt.Print(analyze.AnalyzeComplexity(graph).String())
	case "circular":
		fmt.Print(analyze.AnalyzeCircularDeps(graph).String())
	case "structure":
		if structure == nil {
			structure = analyze.AnalyzeStructure(graph, currentStructureConfig(graph.RootPath))
		}
		fmt.Print(structure.String())
	case "health":
		fmt.Print(analyze.CalculateHealthWithStructure(graph, structure).String())
	case "all":
		fmt.Println("=== Dead Code Analysis (Git-Aware) ===")
		fmt.Print(analyze.FormatDeadCodeResult(analyze.AnalyzeDeadCode(graph)))
		fmt.Println("\n=== Complexity Analysis ===")
		fmt.Print(analyze.AnalyzeComplexity(graph).String())
		fmt.Println("\n=== Circular Dependencies ===")
		fmt.Print(analyze.AnalyzeCircularDeps(graph).String())
		if structure == nil {
			structure = analyze.AnalyzeStructure(graph, currentStructureConfig(graph.RootPath))
		}
		fmt.Println("\n=== Structure (Deep Modules) ===")
		fmt.Print(structure.String())
		fmt.Println("\n=== Overall Health ===")
		var healthStructure *analyze.StructureReport
		if includeStructure {
			healthStructure = structure
		}
		fmt.Print(analyze.CalculateHealthWithStructure(graph, healthStructure).String())
	}
}
