package analyze

import (
	"fmt"
	"strings"
)

// CircularDepsResult contains the results of circular dependency analysis
type CircularDepsResult struct {
	Cycles     []DependencyCycle
	TotalCount int
}

// DependencyCycle represents a circular dependency path
type DependencyCycle struct {
	Files []string // Files in the cycle
	Length int
	Type   string // "import", "export", "type"
}

// AnalyzeCircularDeps detects circular dependencies in the module graph
func AnalyzeCircularDeps(graph *ModuleGraph) *CircularDepsResult {
	result := &CircularDepsResult{
		Cycles: []DependencyCycle{},
	}

	if graph == nil || len(graph.Files) == 0 {
		return result
	}

	// Build adjacency list for imports
	adjacency := make(map[string][]string)
	for filePath, imports := range graph.Imports {
		var deps []string
		for _, imp := range imports {
			// Try to find which files this import refers to
			for otherFile := range graph.Files {
				// Check if the import matches this file
				if matchesImportToFile(imp, otherFile) {
					deps = append(deps, otherFile)
				}
			}
		}
		adjacency[filePath] = deps
	}

	// Find cycles using DFS
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)
	path := []string{}

	var dfs func(node string) bool
	dfs = func(node string) bool {
		visited[node] = true
		recursionStack[node] = true
		path = append(path, node)

		for _, neighbor := range adjacency[node] {
			if !visited[neighbor] {
				if dfs(neighbor) {
					return true
				}
			} else if recursionStack[neighbor] {
				// Found a cycle
				cycle := extractCycle(path, neighbor)
				if cycle.Length > 1 { // Ignore self-loops
					result.Cycles = append(result.Cycles, cycle)
				}
				return true
			}
		}

		path = path[:len(path)-1]
		recursionStack[node] = false
		return false
	}

	// Run DFS from each node
	for filePath := range graph.Files {
		if !visited[filePath] {
			dfs(filePath)
		}
	}

	// Remove duplicate cycles
	result.Cycles = deduplicateCycles(result.Cycles)
	result.TotalCount = len(result.Cycles)

	return result
}

// matchesImportToFile checks if an import could refer to a specific file
func matchesImportToFile(imp Import, filePath string) bool {
	impSource := strings.Trim(imp.Source, `"`)
	fileName := strings.TrimSuffix(filePath, ".go")

	// Direct match
	if impSource == filePath || impSource == fileName {
		return true
	}

	// Check if import source ends with the file name
	if strings.HasSuffix(impSource, "/"+fileName) || strings.HasSuffix(impSource, fileName) {
		return true
	}

	return false
}

// extractCycle extracts the cycle from the current path
func extractCycle(path []string, start string) DependencyCycle {
	var cycleFiles []string
	found := false

	for _, file := range path {
		if file == start {
			found = true
		}
		if found {
			cycleFiles = append(cycleFiles, file)
		}
	}

	// Add the starting node to complete the cycle
	cycleFiles = append(cycleFiles, start)

	return DependencyCycle{
		Files:  cycleFiles,
		Length: len(cycleFiles) - 1,
		Type:   "import",
	}
}

// deduplicateCycles removes duplicate cycles
func deduplicateCycles(cycles []DependencyCycle) []DependencyCycle {
	seen := make(map[string]bool)
	var result []DependencyCycle

	for _, cycle := range cycles {
		// Create a canonical representation of the cycle
		canonical := canonicalizeCycle(cycle.Files)
		if !seen[canonical] {
			seen[canonical] = true
			result = append(result, cycle)
		}
	}

	return result
}

// canonicalizeCycle creates a canonical string representation of a cycle
func canonicalizeCycle(files []string) string {
	if len(files) == 0 {
		return ""
	}

	// Find the minimum element to normalize the cycle
	minIdx := 0
	for i := 1; i < len(files); i++ {
		if files[i] < files[minIdx] {
			minIdx = i
		}
	}

	// Rotate the cycle to start from the minimum
	var canonical []string
	canonical = append(canonical, files[minIdx:]...)
	canonical = append(canonical, files[:minIdx]...)

	return strings.Join(canonical, " -> ")
}

// String returns a human-readable summary of circular dependency analysis
func (r *CircularDepsResult) String() string {
	var sb strings.Builder

	sb.WriteString("\n=== Circular Dependencies ===\n\n")

	if len(r.Cycles) == 0 {
		sb.WriteString("No circular dependencies found.\n")
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("Found %d circular dependencies:\n\n", r.TotalCount))

	for i, cycle := range r.Cycles {
		sb.WriteString(fmt.Sprintf("Cycle %d:\n", i+1))
		for j, file := range cycle.Files {
			if j > 0 {
				sb.WriteString(" -> ")
			}
			sb.WriteString(file)
		}
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("  Length: %d\n\n", cycle.Length))
	}

	return sb.String()
}

// GetSummary returns a brief summary for CLI output
func (r *CircularDepsResult) GetSummary() string {
	if r.TotalCount == 0 {
		return "No circular dependencies"
	}

	return fmt.Sprintf("%d circular dependencies found", r.TotalCount)
}