package analyze

import (
	"fmt"
	"strings"
)

type CircularDepsResult struct {
	Cycles     []DependencyCycle
	TotalCount int
}

type DependencyCycle struct {
	Files []string
	Length int
	Type   string // "import", "export", "type"
}

func AnalyzeCircularDeps(graph *ModuleGraph) *CircularDepsResult {
	result := &CircularDepsResult{
		Cycles: []DependencyCycle{},
	}

	if graph == nil || len(graph.Files) == 0 {
		return result
	}

	adjacency := make(map[string][]string)
	for filePath, imports := range graph.Imports {
		var deps []string
		for _, imp := range imports {
			for otherFile := range graph.Files {
				if matchesImportToFile(imp, otherFile) {
					deps = append(deps, otherFile)
				}
			}
		}
		adjacency[filePath] = deps
	}

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

	for filePath := range graph.Files {
		if !visited[filePath] {
			dfs(filePath)
		}
	}

	result.Cycles = deduplicateCycles(result.Cycles)
	result.TotalCount = len(result.Cycles)

	return result
}

func matchesImportToFile(imp Import, filePath string) bool {
	impSource := strings.Trim(imp.Source, `"`)
	fileName := strings.TrimSuffix(filePath, ".go")

	if impSource == filePath || impSource == fileName {
		return true
	}

	if strings.HasSuffix(impSource, "/"+fileName) || strings.HasSuffix(impSource, fileName) {
		return true
	}

	return false
}

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

	cycleFiles = append(cycleFiles, start)

	return DependencyCycle{
		Files:  cycleFiles,
		Length: len(cycleFiles) - 1,
		Type:   "import",
	}
}

func deduplicateCycles(cycles []DependencyCycle) []DependencyCycle {
	seen := make(map[string]bool)
	var result []DependencyCycle

	for _, cycle := range cycles {
		canonical := canonicalizeCycle(cycle.Files)
		if !seen[canonical] {
			seen[canonical] = true
			result = append(result, cycle)
		}
	}

	return result
}

func canonicalizeCycle(files []string) string {
	if len(files) == 0 {
		return ""
	}

	minIdx := 0
	for i := 1; i < len(files); i++ {
		if files[i] < files[minIdx] {
			minIdx = i
		}
	}

	var canonical []string
	canonical = append(canonical, files[minIdx:]...)
	canonical = append(canonical, files[:minIdx]...)

	return strings.Join(canonical, " -> ")
}

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

func (r *CircularDepsResult) GetSummary() string {
	if r.TotalCount == 0 {
		return "No circular dependencies"
	}

	return fmt.Sprintf("%d circular dependencies found", r.TotalCount)
}