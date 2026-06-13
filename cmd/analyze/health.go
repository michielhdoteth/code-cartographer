package analyze

import (
	"fmt"
	"math"
	"strings"
)

// HealthResult contains overall code health metrics
type HealthResult struct {
	Score           float64         // Overall health score (0-100)
	Grade           string          // Letter grade (A-F)
	Files           int             // Total files analyzed
	Lines           int             // Total lines of code
	Functions       int             // Total functions
	Exports         int             // Total exports
	Imports         int             // Total imports
	DeadCode        *DeadCodeResult // Dead code analysis
	Complexity      *ComplexityResult // Complexity analysis
	Circular        *CircularDepsResult // Circular deps
	Issues          []HealthIssue   // All identified issues
}

// HealthIssue represents a single health issue
type HealthIssue struct {
	Severity string // "error", "warning", "info"
	Category string // "dead-code", "complexity", "circular"
	Message  string
	File     string
	Line     int
}

// CalculateHealth calculates overall health metrics from all analyses
func CalculateHealth(graph *ModuleGraph) *HealthResult {
	result := &HealthResult{
		Files:     len(graph.Files),
		Functions: 0,
		Exports:   0,
		Imports:   0,
		Lines:     0,
	}

	// Run all analyses
	result.DeadCode = AnalyzeDeadCode(graph)
	result.Complexity = AnalyzeComplexity(graph)
	result.Circular = AnalyzeCircularDeps(graph)

	// Count functions, exports, imports
	for _, fileNode := range graph.Files {
		result.Functions += len(fileNode.Exports) // Approximation
		result.Exports += len(fileNode.Exports)
		result.Imports += len(fileNode.Imports)
	}

	// Collect all issues
	result.Issues = result.collectIssues()

	// Calculate health score
	result.Score = result.calculateScore()
	result.Grade = result.calculateGrade()

	return result
}

// collectIssues gathers all issues from analyses
func (r *HealthResult) collectIssues() []HealthIssue {
	var issues []HealthIssue

	// Dead code issues
	for _, ue := range r.DeadCode.UnusedExports {
		issues = append(issues, HealthIssue{
			Severity: "warning",
			Category: "dead-code",
			Message:  fmt.Sprintf("Unused export: %s", ue.Export),
			File:     ue.File,
			Line:     ue.Line,
		})
	}

	for _, uf := range r.DeadCode.UnusedFiles {
		issues = append(issues, HealthIssue{
			Severity: "warning",
			Category: "dead-code",
			Message:  "Unused file: no imports reference this file",
			File:     uf.Path,
			Line:     0,
		})
	}

	// Complexity issues
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

	// Circular dependency issues
	for _, cycle := range r.Circular.Cycles {
		issues = append(issues, HealthIssue{
			Severity: "error",
			Category: "circular",
			Message:  fmt.Sprintf("Circular dependency: %s", strings.Join(cycle.Files, " -> ")),
			File:     cycle.Files[0],
			Line:     0,
		})
	}

	return issues
}

// calculateScore calculates the overall health score (0-100)
func (r *HealthResult) calculateScore() float64 {
	score := 100.0

	// Deduct for dead code (max 30 points)
	deadCodePenalty := float64(r.DeadCode.TotalIssues) * 2
	score -= math.Min(deadCodePenalty, 30)

	// Deduct for complexity issues (max 40 points)
	complexityPenalty := float64(r.Complexity.Issues) * 5
	score -= math.Min(complexityPenalty, 40)

	// Deduct for circular dependencies (max 30 points)
	circularPenalty := float64(r.Circular.TotalCount) * 10
	score -= math.Min(circularPenalty, 30)

	// Ensure score is within bounds
	return math.Max(0, math.Min(100, score))
}

// calculateGrade calculates letter grade from score
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

// String returns a human-readable summary of health analysis
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
	sb.WriteString(fmt.Sprintf("  Dead Code: %d issues\n", r.DeadCode.TotalIssues))
	sb.WriteString(fmt.Sprintf("  Complexity: %d issues\n", r.Complexity.Issues))
	sb.WriteString(fmt.Sprintf("  Circular Dependencies: %d\n", r.Circular.TotalCount))
	sb.WriteString(fmt.Sprintf("  Total Issues: %d\n\n", len(r.Issues)))

	// Show top issues
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

// GetSummary returns a brief summary for CLI output
func (r *HealthResult) GetSummary() string {
	return fmt.Sprintf("Score: %.0f/100 (Grade: %s) - %d issues found",
		r.Score, r.Grade, len(r.Issues))
}