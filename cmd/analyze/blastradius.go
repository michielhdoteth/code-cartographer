package analyze

import (
	"fmt"
	"math"
	"path/filepath"
	"sort"
	"strings"
)

// BlastRadiusResult contains the full blast radius analysis
type BlastRadiusResult struct {
	SeedFile        string              `json:"seed_file"`
	TotalFiles      int                 `json:"total_files"`
	TotalEdges      int                 `json:"total_edges"`
	AffectedFiles   []AffectedFile      `json:"affected_files"`
	Hubs            []Hub               `json:"hubs"`
	MaxDepthReached int                 `json:"max_depth_reached"`
	RiskScore       float64             `json:"risk_score"`
	RiskTier        string              `json:"risk_tier"`
	Concentrated    *ConcentratedRisk   `json:"concentrated_risk,omitempty"`
}

// AffectedFile is a file impacted by the seed change
type AffectedFile struct {
	File          string   `json:"file"`
	Depth         int      `json:"depth"`
	Score         float64  `json:"score"`
	Depths        []int    `json:"depths"`      // multiple paths can reach same file at different depths
	ViaSymbols    []string `json:"via_symbols"` // which symbols carry the blast
	DirectDependents int  `json:"direct_dependents"`
}

// Hub is a file with unusually high fan-in
type Hub struct {
	File       string  `json:"file"`
	FanIn      int     `json:"fan_in"`
	PctOfRepo  float64 `json:"pct_of_repo"`
	Symbol     string  `json:"symbol,omitempty"` // most-imported symbol if known
}

// ConcentratedRisk shows which parts of a file are safe vs critical
type ConcentratedRisk struct {
	File           string             `json:"file"`
	TotalSymbols   int                `json:"total_symbols"`
	CriticalSymbols []SymbolRisk      `json:"critical_symbols"`
	SafeSymbols    []string           `json:"safe_symbols"`
}

// SymbolRisk is a symbol with its risk level
type SymbolRisk struct {
	Name     string  `json:"name"`
	Kind     string  `json:"kind"`
	Line     int     `json:"line"`
	Score    float64 `json:"score"`
	Reach    int     `json:"reach"` // number of files this symbol affects
}

// AnalyzeBlastRadius performs a full blast radius analysis for a seed file
func AnalyzeBlastRadius(graph *ModuleGraph, seedFile string) *BlastRadiusResult {
	result := &BlastRadiusResult{
		SeedFile:      seedFile,
		TotalFiles:    len(graph.Files),
		AffectedFiles: []AffectedFile{},
		Hubs:          []Hub{},
	}

	// Build reverse adjacency (who imports what)
	reverseGraph := buildReverseGraph(graph)
	result.TotalEdges = countEdges(graph)

	// BFS from seed file through reverse graph
	affected := make(map[string]*AffectedFile)
	queue := []BFSNode{{File: seedFile, Depth: 0, ViaSymbol: ""}}
	visited := make(map[string]bool)
	maxDepth := 0

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current.File] {
			// Update existing entry with shorter path if found
			if existing, ok := affected[current.File]; ok {
				if current.Depth < existing.Depth {
					existing.Depth = current.Depth
				}
				existing.Depths = append(existing.Depths, current.Depth)
				if current.ViaSymbol != "" {
					existing.ViaSymbols = appendIfMissing(existing.ViaSymbols, current.ViaSymbol)
				}
			}
			continue
		}
		visited[current.File] = true

		if current.Depth > maxDepth {
			maxDepth = current.Depth
		}

		if current.Depth > 0 {
			affected[current.File] = &AffectedFile{
				File:       current.File,
				Depth:      current.Depth,
				Score:      calculateDepthScore(current.Depth),
				Depths:     []int{current.Depth},
				ViaSymbols: ifEmpty(nil, current.ViaSymbol),
			}
		}

		// Find files that import this one
		if importers, ok := reverseGraph[current.File]; ok {
			for _, imp := range importers {
				if !visited[imp.File] {
					queue = append(queue, BFSNode{
						File:      imp.File,
						Depth:     current.Depth + 1,
						ViaSymbol: imp.Symbol,
					})
				}
			}
		}
	}

	// Calculate risk score: Σ 0.6^(depth-1) for each affected file
	totalScore := 0.0
	for _, af := range affected {
		totalScore += af.Score
	}
	result.RiskScore = totalScore
	result.RiskTier = scoreToTier(totalScore, result.TotalFiles)
	result.MaxDepthReached = maxDepth

	// Sort by score descending
	for _, af := range affected {
		result.AffectedFiles = append(result.AffectedFiles, *af)
	}
	sort.Slice(result.AffectedFiles, func(i, j int) bool {
		return result.AffectedFiles[i].Score > result.AffectedFiles[j].Score
	})

	// Detect hubs
	result.Hubs = detectHubs(graph)

	// Analyze concentrated risk for seed file
	result.Concentrated = analyzeConcentratedRisk(graph, seedFile, affected)

	return result
}

// BFSNode is a node in the BFS traversal
type BFSNode struct {
	File      string
	Depth     int
	ViaSymbol string
}

// reverseEdge is an edge in the reverse dependency graph
type reverseEdge struct {
	File   string
	Symbol string
}

// buildReverseGraph builds a map from file -> files that import it
func buildReverseGraph(graph *ModuleGraph) map[string][]reverseEdge {
	reverse := make(map[string][]reverseEdge)

	for filePath, file := range graph.Files {
		for _, imp := range file.Imports {
			// Try to resolve import to a file
			resolved := resolveImportToFilePath(imp.Source, graph)
			if resolved != "" && resolved != filePath {
				reverse[resolved] = append(reverse[resolved], reverseEdge{
					File:   filePath,
					Symbol: imp.Name,
				})
			}
		}

		// Also check CALLS edges (if we have them in imports metadata)
		for _, exp := range file.Exports {
			key := filePath + ":" + exp.Name
			if refs, ok := graph.Exports[key]; ok {
				_ = refs // For now, exports don't directly give us callers
			}
		}
	}

	return reverse
}

// resolveImportToFilePath tries to resolve an import path to an actual file path
func resolveImportToFilePath(importSource string, graph *ModuleGraph) string {
	// Direct match
	if _, ok := graph.Files[importSource]; ok {
		return importSource
	}

	// Try to find by suffix
	for filePath := range graph.Files {
		if strings.HasSuffix(filePath, "/"+importSource) || strings.HasSuffix(filePath, "\\"+importSource) {
			return filePath
		}
		if strings.HasSuffix(filePath, importSource) {
			return filePath
		}
	}

	// Try without extension
	for filePath := range graph.Files {
		base := strings.TrimSuffix(filePath, ".go")
		base = strings.TrimSuffix(base, ".ts")
		base = strings.TrimSuffix(base, ".js")
		if strings.HasSuffix(base, "/"+importSource) || strings.HasSuffix(base, "\\"+importSource) {
			return filePath
		}
	}

	return ""
}

// calculateDepthScore returns the depth-weighted score for a file at given depth
func calculateDepthScore(depth int) float64 {
	if depth == 0 {
		return 0
	}
	return math.Pow(0.6, float64(depth-1))
}

// scoreToTier converts a raw score to a risk tier
func scoreToTier(score float64, totalFiles int) string {
	if totalFiles == 0 {
		return "unknown"
	}
	pct := (score / float64(totalFiles)) * 100
	switch {
	case pct >= 25:
		return "critical"
	case pct >= 10:
		return "high"
	case pct >= 3:
		return "medium"
	default:
		return "low"
	}
}

// detectHubs finds files with high fan-in (imported by many others)
func detectHubs(graph *ModuleGraph) []Hub {
	fanIn := make(map[string]int)
	symbolFanIn := make(map[string]map[string]int) // file -> symbol -> count

	for _, file := range graph.Files {
		for _, imp := range file.Imports {
			resolved := resolveImportToFilePath(imp.Source, graph)
			if resolved != "" {
				fanIn[resolved]++
				if symbolFanIn[resolved] == nil {
					symbolFanIn[resolved] = make(map[string]int)
				}
				symbolFanIn[resolved][imp.Name]++
			}
		}
	}

	var hubs []Hub
	threshold := float64(len(graph.Files)) * 0.10 // 10% of repo

	for file, count := range fanIn {
		if float64(count) >= threshold {
			// Find most imported symbol
			bestSymbol := ""
			bestCount := 0
			for sym, symCount := range symbolFanIn[file] {
				if symCount > bestCount {
					bestSymbol = sym
					bestCount = symCount
				}
			}
			hubs = append(hubs, Hub{
				File:      file,
				FanIn:     count,
				PctOfRepo: (float64(count) / float64(len(graph.Files))) * 100,
				Symbol:    bestSymbol,
			})
		}
	}

	sort.Slice(hubs, func(i, j int) bool {
		return hubs[i].FanIn > hubs[j].FanIn
	})

	return hubs
}

// analyzeConcentratedRisk shows which symbols in the seed file are critical
func analyzeConcentratedRisk(graph *ModuleGraph, seedFile string, affected map[string]*AffectedFile) *ConcentratedRisk {
	fileNode, ok := graph.Files[seedFile]
	if !ok {
		return nil
	}

	result := &ConcentratedRisk{
		File:            seedFile,
		TotalSymbols:    len(fileNode.Exports),
		CriticalSymbols: []SymbolRisk{},
		SafeSymbols:     []string{},
	}

	for _, exp := range fileNode.Exports {
		reach := countSymbolReach(graph, seedFile, exp.Name)
		score := float64(reach) / float64(max(1, len(graph.Files)))

		sr := SymbolRisk{
			Name:  exp.Name,
			Kind:  exp.Kind,
			Line:  exp.Line,
			Score: score,
			Reach: reach,
		}

		if score >= 0.03 {
			result.CriticalSymbols = append(result.CriticalSymbols, sr)
		} else {
			result.SafeSymbols = append(result.SafeSymbols, exp.Name)
		}
	}

	// Sort critical symbols by score
	sort.Slice(result.CriticalSymbols, func(i, j int) bool {
		return result.CriticalSymbols[i].Score > result.CriticalSymbols[j].Score
	})

	return result
}

// countSymbolReach counts how many files transitively depend on a symbol
func countSymbolReach(graph *ModuleGraph, filePath string, symbolName string) int {
	count := 0
	visited := make(map[string]bool)

	var dfs func(file string)
	dfs = func(file string) {
		if visited[file] {
			return
		}
		visited[file] = true

		for otherFile, otherNode := range graph.Files {
			if otherFile == file {
				continue
			}
			for _, imp := range otherNode.Imports {
				if imp.Name == symbolName || strings.Contains(imp.Source, symbolName) {
					if !visited[otherFile] {
						count++
						dfs(otherFile)
					}
				}
			}
		}
	}

	dfs(filePath)
	return count
}

func countEdges(graph *ModuleGraph) int {
	count := 0
	for _, file := range graph.Files {
		count += len(file.Imports)
	}
	return count
}

func appendIfMissing(slice []string, s string) []string {
	for _, v := range slice {
		if v == s {
			return slice
		}
	}
	return append(slice, s)
}

func ifEmpty(defaultVal []string, val string) []string {
	if val == "" {
		return defaultVal
	}
	return []string{val}
}

// FormatBlastRadiusResult formats the result for display
func FormatBlastRadiusResult(result *BlastRadiusResult) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("=== Blast Radius: %s ===\n", filepath.Base(result.SeedFile)))
	sb.WriteString(fmt.Sprintf("Risk: %s (%.1f%% of repo)\n\n", strings.ToUpper(result.RiskTier), result.RiskScore/float64(max(1, result.TotalFiles))*100))

	// Hubs
	if len(result.Hubs) > 0 {
		sb.WriteString("Architectural Hubs (imported by ≥10% of repo):\n")
		for _, h := range result.Hubs {
			sym := ""
			if h.Symbol != "" {
				sym = fmt.Sprintf(" [via %s]", h.Symbol)
			}
			sb.WriteString(fmt.Sprintf("  %s: %d imports (%.0f%%)%s\n", h.File, h.FanIn, h.PctOfRepo, sym))
		}
		sb.WriteString("\n")
	}

	// Affected files
	if len(result.AffectedFiles) > 0 {
		sb.WriteString(fmt.Sprintf("Affected Files (%d):\n", len(result.AffectedFiles)))
		for i, af := range result.AffectedFiles {
			if i >= 30 {
				sb.WriteString(fmt.Sprintf("  ... and %d more\n", len(result.AffectedFiles)-30))
				break
			}
			via := ""
			if len(af.ViaSymbols) > 0 {
				via = fmt.Sprintf(" via %s", strings.Join(af.ViaSymbols, ", "))
			}
			sb.WriteString(fmt.Sprintf("  depth %d  score %.3f  %s%s\n", af.Depth, af.Score, af.File, via))
		}
		sb.WriteString("\n")
	}

	// Concentrated risk
	if result.Concentrated != nil && len(result.Concentrated.CriticalSymbols) > 0 {
		sb.WriteString("Concentrated Risk — critical symbols in this file:\n")
		for _, cs := range result.Concentrated.CriticalSymbols {
			sb.WriteString(fmt.Sprintf("  %s %s (line %d) — affects %d files (%.1f%%)\n", cs.Kind, cs.Name, cs.Line, cs.Reach, cs.Score*100))
		}
		if len(result.Concentrated.SafeSymbols) > 0 {
			sb.WriteString(fmt.Sprintf("\nSafe to edit: %s\n", strings.Join(result.Concentrated.SafeSymbols, ", ")))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
