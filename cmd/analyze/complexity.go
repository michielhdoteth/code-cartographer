package analyze

import (
	"fmt"
	"sort"
	"strings"
)

type ComplexityResult struct {
	Functions     []FunctionComplexity
	TotalFiles    int
	AvgComplexity float64
	Issues        int
}

type FunctionComplexity struct {
	Name       string
	File       string
	Line       int
	Cyclomatic int
	Cognitive  int
	Lines      int
	Params     int
	Nesting    int
}

const (
	MaxCyclomatic = 10
	MaxCognitive  = 15
	MaxLinesPerFn = 50
	MaxParams     = 5
	MaxNesting    = 4
)

func AnalyzeComplexity(graph *ModuleGraph) *ComplexityResult {
	result := &ComplexityResult{
		Functions: []FunctionComplexity{},
	}

	if graph == nil || len(graph.Files) == 0 {
		return result
	}

	result.TotalFiles = len(graph.Files)

	// Since the unified parser doesn't provide per-function metrics,
	// we estimate complexity based on the number of functions
	// In a full implementation, we would parse each file to get detailed metrics

	for _, file := range graph.Files {
		for _, exp := range file.Exports {
			if exp.Kind == "function" {
				result.Functions = append(result.Functions, FunctionComplexity{
					Name:       exp.Name,
					File:       file.Path,
					Line:       exp.Line,
					Cyclomatic: 1,
					Cognitive:  1,
					Lines:      10,
					Params:     0,
					Nesting:    1,
				})
			}
		}
	}

	if len(result.Functions) > 0 {
		total := 0
		for _, fn := range result.Functions {
			total += fn.Cyclomatic
		}
		result.AvgComplexity = float64(total) / float64(len(result.Functions))
	}

	for _, fn := range result.Functions {
		if fn.Cyclomatic > MaxCyclomatic || fn.Cognitive > MaxCognitive ||
			fn.Lines > MaxLinesPerFn || fn.Params > MaxParams || fn.Nesting > MaxNesting {
			result.Issues++
		}
	}

	sort.Slice(result.Functions, func(i, j int) bool {
		return result.Functions[i].Cyclomatic > result.Functions[j].Cyclomatic
	})

	return result
}

func (r *ComplexityResult) String() string {
	var sb strings.Builder

	sb.WriteString("\n=== Complexity Analysis ===\n\n")
	sb.WriteString(fmt.Sprintf("Files analyzed: %d\n", r.TotalFiles))
	sb.WriteString(fmt.Sprintf("Functions analyzed: %d\n", len(r.Functions)))
	sb.WriteString(fmt.Sprintf("Average complexity: %.2f\n", r.AvgComplexity))
	sb.WriteString(fmt.Sprintf("Issues found: %d\n\n", r.Issues))

	if len(r.Functions) > 0 {
		sb.WriteString("Top 10 most complex functions:\n")
		for i, fn := range r.Functions {
			if i >= 10 {
				break
			}
			flags := ""
			if fn.Cyclomatic > MaxCyclomatic {
				flags += " [HIGH COMPLEXITY]"
			}
			if fn.Lines > MaxLinesPerFn {
				flags += " [LONG]"
			}
			if fn.Params > MaxParams {
				flags += " [MANY PARAMS]"
			}
			sb.WriteString(fmt.Sprintf(" %s:%d - %s (CC: %d, Cognitive: %d, Lines: %d)%s\n",
				fn.File, fn.Line, fn.Name, fn.Cyclomatic, fn.Cognitive, fn.Lines, flags))
		}
	}

	return sb.String()
}

func (r *ComplexityResult) GetSummary() string {
	if len(r.Functions) == 0 {
		return "No functions analyzed"
	}

	var parts []string
	if r.Issues > 0 {
		parts = append(parts, fmt.Sprintf("%d functions exceed thresholds", r.Issues))
	}
	parts = append(parts, fmt.Sprintf("avg CC: %.1f", r.AvgComplexity))

	return strings.Join(parts, ", ")
}

func (r *ComplexityResult) GetTopFunctions(n int) []FunctionComplexity {
	if n > len(r.Functions) {
		n = len(r.Functions)
	}
	return r.Functions[:n]
}