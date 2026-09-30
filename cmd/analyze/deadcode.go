package analyze

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// DeadCodeResult contains the enhanced dead code analysis
type DeadCodeResult struct {
	DeadExports     []DeadExport     `json:"dead_exports"`
	DeadFiles       []DeadFile       `json:"dead_files"`
	Duplicates      []DuplicateGroup `json:"duplicates"`
	TotalExports    int              `json:"total_exports"`
	TotalDead       int              `json:"total_dead"`
	RemovalScore    float64          `json:"removal_score"` // 0-100, higher = more can be safely removed
}

// DeadExport is an unused export with confidence scoring
type DeadExport struct {
	Name       string  `json:"name"`
	File       string  `json:"file"`
	Kind       string  `json:"kind"`
	Line       int     `json:"line"`
	Confidence float64 `json:"confidence"` // 0-1, higher = more confident it's dead
	Reasons    []string `json:"reasons"`
	DaysSinceLastChange int `json:"days_since_last_change"`
	IsPublic   bool    `json:"is_public"`
	References int     `json:"references"` // how many other files reference this
}

// DeadFile is a file with no incoming edges (orphan)
type DeadFile struct {
	Path          string  `json:"path"`
	Exports       int     `json:"exports"`
	Confidence    float64 `json:"confidence"`
	DaysSinceLastChange int `json:"days_since_last_change"`
	Reason        string  `json:"reason"`
}

// DuplicateGroup is a set of near-identical code blocks
type DuplicateGroup struct {
	ID          int              `json:"id"`
	Files       []DuplicateEntry `json:"files"`
	Hash        string           `json:"hash"`
	Similarity  float64          `json:"similarity"`
	Removeable  bool             `json:"removable"`
}

// DuplicateEntry is a single duplicate occurrence
type DuplicateEntry struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	EndLine  int    `json:"end_line"`
	Name     string `json:"name"`
	Kind     string `json:"kind"`
}

// AnalyzeDeadCodeEnhanced performs git-aware dead code analysis
func AnalyzeDeadCodeEnhanced(graph *ModuleGraph, rootPath string) *DeadCodeResult {
	result := &DeadCodeResult{
		DeadExports: []DeadExport{},
		DeadFiles:   []DeadFile{},
		Duplicates:  []DuplicateGroup{},
	}

	// Count total exports
	for _, file := range graph.Files {
		result.TotalExports += len(file.Exports)
	}

	// Find dead exports
	result.DeadExports = findDeadExports(graph, rootPath)

	// Find orphan files
	result.DeadFiles = findOrphanFiles(graph, rootPath)

	// Find duplicates
	result.Duplicates = findDuplicates(graph, rootPath)

	result.TotalDead = len(result.DeadExports) + len(result.DeadFiles)
	result.RemovalScore = calculateRemovalScore(result)

	return result
}

// findDeadExports finds exports that are not referenced by any other file
func findDeadExports(graph *ModuleGraph, rootPath string) []DeadExport {
	var dead []DeadExport

	for filePath, file := range graph.Files {
		for _, exp := range file.Exports {
			refs := countReferences(graph, filePath, exp.Name)
			if refs == 0 {
				daysSince := getDaysSinceLastChange(filePath, rootPath)
				confidence := calculateDeadConfidence(exp.Kind, daysSince, false, 0)
				reasons := buildReasons(exp.Kind, daysSince, false)

				dead = append(dead, DeadExport{
					Name:                exp.Name,
					File:                filePath,
					Kind:                exp.Kind,
					Line:                exp.Line,
					Confidence:          confidence,
					Reasons:             reasons,
					DaysSinceLastChange: daysSince,
					IsPublic:            isExportedName(exp.Name),
					References:          0,
				})
			}
		}
	}

	// Sort by confidence descending
	sort.Slice(dead, func(i, j int) bool {
		return dead[i].Confidence > dead[j].Confidence
	})

	return dead
}

// findOrphanFiles finds files that nothing imports
func findOrphanFiles(graph *ModuleGraph, rootPath string) []DeadFile {
	var orphans []DeadFile

	// Build reverse graph
	importedBy := make(map[string]int)
	for _, file := range graph.Files {
		for _, imp := range file.Imports {
			resolved := resolveImportFile(imp.Source, graph)
			if resolved != "" {
				importedBy[resolved]++
			}
		}
	}

	for filePath, file := range graph.Files {
		if importedBy[filePath] == 0 {
			// This file is not imported by anything
			daysSince := getDaysSinceLastChange(filePath, rootPath)
			confidence := 0.5 // base confidence for orphan files

			// Boost confidence if old and has no exports
			if daysSince > 90 && len(file.Exports) == 0 {
				confidence = 0.8
			} else if daysSince > 30 {
				confidence = 0.6
			}

			// Check if it's an entry point (main, index, etc.)
			base := filepath.Base(filePath)
			if isEntryPoint(base) {
				confidence *= 0.3 // Much less likely to be dead
			}

			orphans = append(orphans, DeadFile{
				Path:                filePath,
				Exports:             len(file.Exports),
				Confidence:          confidence,
				DaysSinceLastChange: daysSince,
				Reason:              "not imported by any file",
			})
		}
	}

	sort.Slice(orphans, func(i, j int) bool {
		return orphans[i].Confidence > orphans[j].Confidence
	})

	return orphans
}

// findDuplicates finds near-identical code blocks via content hashing
func findDuplicates(graph *ModuleGraph, rootPath string) []DuplicateGroup {
	type fileContent struct {
		path    string
		content string
		lines   []string
	}

	var files []fileContent
	for filePath := range graph.Files {
		absPath := filepath.Join(rootPath, filePath)
		data, err := os.ReadFile(absPath)
		if err != nil {
			continue
		}
		content := string(data)
		lines := strings.Split(content, "\n")
		files = append(files, fileContent{path: filePath, content: content, lines: lines})
	}

	// Normalize and hash each function/block
	type blockHash struct {
		hash    string
		entry   DuplicateEntry
		content string
	}

	blocksByHash := make(map[string][]blockHash)

	for _, fc := range files {
		functions := extractFunctionBlocks(fc.path, fc.lines)
		for _, fn := range functions {
			// Extract content from source lines
			content := ""
			if fn.EndLine <= len(fc.lines) && fn.Line > 0 {
				content = strings.Join(fc.lines[fn.Line-1:fn.EndLine], "\n")
			}
			normalized := normalizeCode(content)
			hash := fmt.Sprintf("%x", sha256.Sum256([]byte(normalized)))[:16]
			blocksByHash[hash] = append(blocksByHash[hash], blockHash{
				hash:    hash,
				entry:   fn,
				content: content,
			})
		}
	}

	var groups []DuplicateGroup
	groupID := 0
	for hash, entries := range blocksByHash {
		if len(entries) < 2 {
			continue
		}

		var dupEntries []DuplicateEntry
		for _, e := range entries {
			dupEntries = append(dupEntries, e.entry)
		}

		groupID++
		groups = append(groups, DuplicateGroup{
			ID:         groupID,
			Files:      dupEntries,
			Hash:       hash,
			Similarity: 1.0, // exact match
			Removeable: len(dupEntries) > 1,
		})
	}

	sort.Slice(groups, func(i, j int) bool {
		return len(groups[i].Files) > len(groups[j].Files)
	})

	return groups
}

// extractFunctionBlocks extracts function/method blocks from source
func extractFunctionBlocks(filePath string, lines []string) []DuplicateEntry {
	var entries []DuplicateEntry
	inFunction := false
	startLine := 0
	braceCount := 0
	funcName := ""

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if !inFunction {
			// Look for function declarations
			if strings.HasPrefix(trimmed, "func ") || strings.HasPrefix(trimmed, "function ") ||
				strings.HasPrefix(trimmed, "def ") || strings.HasPrefix(trimmed, "fn ") {
				inFunction = true
				startLine = i + 1
				braceCount = 0
				// Extract name
				parts := strings.Fields(trimmed)
				if len(parts) >= 2 {
					funcName = parts[1]
					funcName = strings.TrimSuffix(funcName, "(")
					funcName = strings.TrimSuffix(funcName, ":")
				}
			}
		} else {
			braceCount += strings.Count(line, "{") - strings.Count(line, "}")
			if braceCount <= 0 && (strings.Contains(line, "}") || trimmed == "") {
				content := strings.Join(lines[startLine-1:i+1], "\n")
				if len(strings.Split(content, "\n")) >= 5 { // only consider blocks >= 5 lines
					entries = append(entries, DuplicateEntry{
						File:    filePath,
						Line:    startLine,
						EndLine: i + 1,
						Name:    funcName,
						Kind:    "function",
					})
				}
				inFunction = false
			}
		}
	}

	return entries
}

// normalizeCode normalizes code for comparison (strip names, comments, whitespace)
func normalizeCode(code string) string {
	lines := strings.Split(code, "\n")
	var normalized []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") {
			continue
		}
		// Skip doc strings
		if strings.HasPrefix(trimmed, "///") || strings.HasPrefix(trimmed, "//!") {
			continue
		}
		// Normalize whitespace
		trimmed = strings.Join(strings.Fields(trimmed), " ")
		normalized = append(normalized, trimmed)
	}

	return strings.Join(normalized, "\n")
}

// countReferences counts how many files reference a given symbol
func countReferences(graph *ModuleGraph, filePath string, symbolName string) int {
	count := 0
	for otherFile, otherNode := range graph.Files {
		if otherFile == filePath {
			continue
		}
		for _, imp := range otherNode.Imports {
			if imp.Name == symbolName || strings.Contains(imp.Source, symbolName) {
				count++
				break
			}
		}
	}
	return count
}

// getDaysSinceLastChange returns days since file was last modified via git
func getDaysSinceLastChange(filePath string, rootPath string) int {
	absPath := filepath.Join(rootPath, filePath)
	cmd := exec.Command("git", "log", "-1", "--format=%at", "--", absPath)
	cmd.Dir = rootPath
	out, err := cmd.Output()
	if err != nil {
		return 999 // default to old if git fails
	}

	timestamp, err := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64)
	if err != nil {
		return 999
	}

	lastChange := time.Unix(timestamp, 0)
	return int(time.Since(lastChange).Hours() / 24)
}

// calculateDeadConfidence calculates confidence that a symbol is truly dead
func calculateDeadConfidence(kind string, daysSince int, isPublic bool, refs int) float64 {
	confidence := 0.5 // base

	// Time factor: older = more confident
	switch {
	case daysSince > 180:
		confidence += 0.25
	case daysSince > 90:
		confidence += 0.15
	case daysSince > 30:
	confidence += 0.05
	case daysSince < 7:
		confidence -= 0.2 // recently changed, might be WIP
	}

	// Scope factor: private = more confident
	if !isPublic {
		confidence += 0.1
	} else {
		confidence -= 0.1 // public APIs might be used externally
	}

	// Kind factor
	switch kind {
	case "function", "method":
		confidence += 0.05
	case "struct", "class", "interface":
		confidence -= 0.05 // types might be used in type assertions
	}

	// Clamp to [0, 1]
	if confidence < 0 {
		confidence = 0
	}
	if confidence > 1 {
		confidence = 1
	}

	return confidence
}

// buildReasons builds human-readable reasons for dead code classification
func buildReasons(kind string, daysSince int, isPublic bool) []string {
	var reasons []string

	if daysSince > 90 {
		reasons = append(reasons, fmt.Sprintf("no changes in %d days", daysSince))
	}
	if !isPublic {
		reasons = append(reasons, "not exported (private)")
	}
	reasons = append(reasons, "no references found in codebase")

	return reasons
}

// isExportedName checks if a name is exported (starts with uppercase)
func isExportedName(name string) bool {
	if len(name) == 0 {
		return false
	}
	return name[0] >= 'A' && name[0] <= 'Z'
}

// isEntryPoint checks if a filename looks like an entry point
func isEntryPoint(name string) bool {
	entryPoints := []string{
		"main.go", "main.ts", "main.js", "main.py", "main.rs",
		"index.go", "index.ts", "index.js", "index.html",
		"app.go", "app.ts", "app.js", "app.py",
		"server.go", "server.ts", "server.js",
		"cli.go", "cli.ts", "cli.js",
		"cmd", "Cmd",
	}
	for _, ep := range entryPoints {
		if strings.EqualFold(name, ep) || strings.HasPrefix(name, ep) {
			return true
		}
	}
	return false
}

// resolveImportFile resolves an import to a file path (simplified)
func resolveImportFile(source string, graph *ModuleGraph) string {
	for filePath := range graph.Files {
		if strings.HasSuffix(filePath, "/"+source) || strings.HasSuffix(filePath, "\\"+source) {
			return filePath
		}
		if strings.Contains(filePath, source) {
			return filePath
		}
	}
	return ""
}

// calculateRemovalScore calculates a score for how much dead code can be safely removed
func calculateRemovalScore(result *DeadCodeResult) float64 {
	if result.TotalExports == 0 {
		return 0
	}

	score := 0.0
	for _, dead := range result.DeadExports {
		score += dead.Confidence
	}
	for _, df := range result.DeadFiles {
		score += df.Confidence
	}

	// Normalize to 0-100
	maxScore := float64(len(result.DeadExports) + len(result.DeadFiles))
	if maxScore == 0 {
		return 0
	}
	return (score / maxScore) * 100
}

// FormatDeadCodeResult formats the result for display
func FormatDeadCodeResult(result *DeadCodeResult) string {
	var sb strings.Builder

	sb.WriteString("=== Dead Code Analysis (Git-Aware) ===\n\n")
	sb.WriteString(fmt.Sprintf("Total exports: %d\n", result.TotalExports))
	sb.WriteString(fmt.Sprintf("Dead exports:  %d\n", len(result.DeadExports)))
	sb.WriteString(fmt.Sprintf("Orphan files:  %d\n", len(result.DeadFiles)))
	sb.WriteString(fmt.Sprintf("Duplicates:    %d groups\n", len(result.Duplicates)))
	sb.WriteString(fmt.Sprintf("Removal score: %.0f/100\n\n", result.RemovalScore))

	// Top dead exports
	if len(result.DeadExports) > 0 {
		sb.WriteString("Dead Exports (by confidence):\n")
	shown := 0
	for _, dead := range result.DeadExports {
			if shown >= 20 {
				sb.WriteString(fmt.Sprintf("  ... and %d more\n", len(result.DeadExports)-20))
				break
			}
			age := ""
			if dead.DaysSinceLastChange > 30 {
				age = fmt.Sprintf(" (%d days since change)", dead.DaysSinceLastChange)
			}
			pub := ""
			if dead.IsPublic {
				pub = " [exported]"
			}
			sb.WriteString(fmt.Sprintf("  %.0f%%  %s %s%s%s\n",
				dead.Confidence*100, dead.Kind, dead.Name, pub, age))
			shown++
		}
		sb.WriteString("\n")
	}

	// Duplicates
	if len(result.Duplicates) > 0 {
		sb.WriteString("Duplicate Code Groups:\n")
		for _, dup := range result.Duplicates {
			sb.WriteString(fmt.Sprintf("  Group %d (%.0f%% similar, %d occurrences):\n", dup.ID, dup.Similarity*100, len(dup.Files)))
			for _, entry := range dup.Files {
				sb.WriteString(fmt.Sprintf("    %s:%d-%d  %s\n", entry.File, entry.Line, entry.EndLine, entry.Name))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// AnalyzeDeadCode is the backward-compatible entry point that calls the enhanced version
func AnalyzeDeadCode(graph *ModuleGraph) *DeadCodeResult {
	return AnalyzeDeadCodeEnhanced(graph, graph.RootPath)
}
