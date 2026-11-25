package analyze

import (
	"fmt"
	"math"
	"strings"
)

type HealthResult struct {
	Score           float64         // Overall health score (0-100)
	Grade           string          // Letter grade (A-F)
	Files           int
	Lines           int
	Functions       int
	Exports         int
	Imports         int
	DeadCode        *DeadCodeResult
	Complexity      *ComplexityResult
	Circular        *CircularDepsResult
	Structure       *StructureReport    // Optional structural analysis (opt-in)
	Issues          []HealthIssue
}

type HealthIssue struct {
	Severity string // "error", "warning", "info"
	Category string // "dead-code", "complexity", "circular"
	Message  string
	File     string
	Line     int
}

func CalculateHealth(graph *ModuleGraph) *HealthResult {
	return CalculateHealthWithStructure(graph, nil)
}

// CalculateHealthWithStructure includes structural findings when a report is
// supplied. Callers opt in explicitly so the existing health score stays stable.
func CalculateHealthWithStructure(graph *ModuleGraph, structure *StructureReport) *HealthResult {
	result := &HealthResult{
		Files:     len(graph.Files),
		Functions: 0,
		Exports:   0,
		Imports:   0,
		Lines:     0,
	}

	result.DeadCode = AnalyzeDeadCode(graph)
	result.Complexity = AnalyzeComplexity(graph)
	result.Circular = AnalyzeCircularDeps(graph)
	result.Structure = structure

	for _, fileNode := range graph.Files {
		result.Functions += len(fileNode.Exports) // Approximation
		result.Exports += len(fileNode.Exports)
		result.Imports += len(fileNode.Imports)
	}

	result.Issues = result.collectIssues()

	result.Score = result.calculateScore()
	result.Grade = result.calculateGrade()

	return result
}

func (r *HealthResult) collectIssues() []HealthIssue {
	var issues []HealthIssue

	for _, de := range r.DeadCode.DeadExports {
		issues = append(issues, HealthIssue{
			Severity: "warning",
			Category: "dead-code",
			Message:  fmt.Sprintf("Unused export: %s (%.0f%% confidence)", de.Name, de.Confidence*100),
			File:     de.File,
			Line:     de.Line,
		})
	}

	for _, df := range r.DeadCode.DeadFiles {
		issues = append(issues, HealthIssue{
			Severity: "warning",
			Category: "dead-code",
			Message:  "Dead file: no imports reference this file",
			File:     df.Path,
			Line:     0,
		})
	}

	for _, fn := range r.Complexity.Functions {
		if fn.Cyclomatic > MaxCyclomatic {
			issues = append(issues, HealthIssue{
				Severity: "warning",
				Category: "complexity",
				Message:  fmt.Sprintf("High cyclomatic complexity: %d (threshold: %d)", fn.Cyclomatic, MaxCyclomatic),
				File:     fn.File,
				Line:     fn.Line,
			})
		}
		if fn.Cognitive > MaxCognitive {
			issues = append(issues, HealthIssue{
				Severity: "warning",
				Category: "complexity",
				Message:  fmt.Sprintf("High cognitive complexity: %d (threshold: %d)", fn.Cognitive, MaxCognitive),
				File:     fn.File,
				Line:     fn.Line,
			})
		}
		if fn.Lines > MaxLinesPerFn {
			issues = append(issues, HealthIssue{
				Severity: "info",
				Category: "complexity",
				Message:  fmt.Sprintf("Long function: %d lines (threshold: %d)", fn.Lines, MaxLinesPerFn),
				File:     fn.File,
				Line:     fn.Line,
			})
		}
	}

	for _, cycle := range r.Circular.Cycles {
		issues = append(issues, HealthIssue{
			Severity: "error",
			Category: "circular",
			Message:  fmt.Sprintf("Circular dependency: %s", strings.Join(cycle.Files, " -> ")),
			File:     cycle.Files[0],
			Line:     0,
		})
	}

	if r.Structure != nil {
		for _, f := range r.Structure.Findings {
			severity := "warning"
			if f.Severity == "critical" {
				severity = "error"
			} else if f.Severity == "info" {
				severity = "info"
			}
			file := ""
			if len(f.Files) > 0 {
				file = f.Files[0]
			}
			issues = append(issues, HealthIssue{
				Severity: severity,
				Category: "structure",
				Message:  fmt.Sprintf("%s: %s", f.Rule, f.Message),
				File:     file,
			})
		}
	}

	return issues
}

func (r *HealthResult) calculateScore() float64 {
	score := 100.0

	// Deduct for dead code (max 30 points)
	deadCodePenalty := float64(r.DeadCode.TotalDead) * 2
	score -= math.Min(deadCodePenalty, 30)

	// Deduct for complexity issues (max 40 points)
	complexityPenalty := float64(r.Complexity.Issues) * 5
	score -= math.Min(complexityPenalty, 40)

	// Deduct for circular dependencies (max 30 points)
	circularPenalty := float64(r.Circular.TotalCount) * 10
	score -= math.Min(circularPenalty, 30)

	// Deduct for structural findings (max 15 points, opt-in only)
	if r.Structure != nil {
		structurePenalty := float64(r.Structure.Summary.Critical)*5 + float64(r.Structure.Summary.Warning)*2
		score -= math.Min(structurePenalty, 15)
	}

	return math.Max(0, math.Min(100, score))
}

func (r *HealthResult) calculateGrade() string {
	switch {
	case r.Score >= 90:
		return "A"
	case r.Score >= 80:
		return "B"
	case r.Score >= 70:
		return "C"
	case r.Score >= 60:
		return "D"
	default:
		return "F"
	}
}

func (r *HealthResult) String() string {
	var sb strings.Builder

	sb.WriteString("\n=== Code Health Report ===\n\n")
	sb.WriteString(fmt.Sprintf("Health Score: %.0f/100 (Grade: %s)\n\n", r.Score, r.Grade))

	sb.WriteString("Metrics:\n")
	sb.WriteString(fmt.Sprintf("  Files: %d\n", r.Files))
	sb.WriteString(fmt.Sprintf("  Functions analyzed: %d\n", len(r.Complexity.Functions)))
	sb.WriteString(fmt.Sprintf("  Exports: %d\n", r.Exports))
	sb.WriteString(fmt.Sprintf("  Imports: %d\n", r.Imports))

	sb.WriteString("\nIssues Summary:\n")
	sb.WriteString(fmt.Sprintf("  Dead Code: %d issues\n", r.DeadCode.TotalDead))
	sb.WriteString(fmt.Sprintf("  Complexity: %d issues\n", r.Complexity.Issues))
	sb.WriteString(fmt.Sprintf("  Circular Dependencies: %d\n", r.Circular.TotalCount))
	sb.WriteString(fmt.Sprintf("  Total Issues: %d\n\n", len(r.Issues)))

	if len(r.Issues) > 0 {
		sb.WriteString("Top Issues:\n")
		for i, issue := range r.Issues {
			if i >= 10 {
				break
			}
			sb.WriteString(fmt.Sprintf("  [%s] %s\n", strings.ToUpper(issue.Severity), issue.Message))
			if issue.File != "" {
				sb.WriteString(fmt.Sprintf("    -> %s", issue.File))
				if issue.Line > 0 {
					sb.WriteString(fmt.Sprintf(":%d", issue.Line))
				}
				sb.WriteString("\n")
			}
		}
	}

	return sb.String()
}

func (r *HealthResult) GetSummary() string {
	return fmt.Sprintf("Score: %.0f/100 (Grade: %s) - %d issues found",
		r.Score, r.Grade, len(r.Issues))
}