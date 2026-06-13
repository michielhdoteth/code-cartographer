package analyze

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// DeadCodeResult contains the results of dead code analysis
type DeadCodeResult struct {
	UnusedExports []UnusedExport
	UnusedFiles   []UnusedFile
	UnusedImports []UnusedImport
	TotalIssues   int
}

// UnusedExport represents an exported symbol that is not used
type UnusedExport struct {
	File   string
	Export string
	Kind   string
	Line   int
	Reason string
}

// UnusedFile represents a file that is not imported by any other file
type UnusedFile struct {
	Path   string
	Reason string
}

// UnusedImport represents an imported symbol that is not used
type UnusedImport struct {
	File   string
	Import string
	Source string
	Line   int
	Reason string
}

// AnalyzeDeadCode performs dead code analysis on the module graph
func AnalyzeDeadCode(graph *ModuleGraph) *DeadCodeResult {
	result := &DeadCodeResult{
		UnusedExports: []UnusedExport{},
		UnusedFiles:   []UnusedFile{},
		UnusedImports: []UnusedImport{},
	}

	if graph == nil || len(graph.Files) == 0 {
		result.TotalIssues = 0
		return result
	}

	// Track used exports
	usedExports := make(map[string]bool)

	// Collect all usage patterns
	// 1. Direct imports: import path "symbol" -> marks that symbol as used

	// Step 1: Find direct imports that reference exports
	for _, imports := range graph.Imports {
		for _, imp := range imports {
			// Check if any export in the codebase matches this import
			for exportKey := range graph.Exports {
				parts := strings.SplitN(exportKey, ":", 2)
				if len(parts) != 2 {
					continue
				}
				expFile := parts[0]
				expName := parts[1]

				// Only mark as used if:
				// - Import name matches export name OR
				// - Import source matches the file containing the export
				if matchesImport(imp, expName, expFile) {
					usedExports[exportKey] = true
				}
			}
		}
	}

	// Step 2: Find unused exports
	for exportKey, exportRef := range graph.Exports {
		if !usedExports[exportKey] {
			// Find the export details
			fileNode := graph.Files[exportRef.File]
			if fileNode != nil {
				for _, exp := range fileNode.Exports {
					if exp.Name == exportRef.Export {
						result.UnusedExports = append(result.UnusedExports, UnusedExport{
							File:   exportRef.File,
							Export: exportRef.Export,
							Kind:   exp.Kind,
							Line:   exp.Line,
							Reason: "No imports reference this export",
						})
						break
					}
				}
			}
		}
	}

	// Step 3: Find unused files (files with exports but never imported)
	for filePath, fileNode := range graph.Files {
		if len(fileNode.Exports) == 0 {
			continue // Skip files with no exports
		}

		// Check if any of this file's exports are used
		isUsed := false
		for _, exp := range fileNode.Exports {
			key := filePath + ":" + exp.Name
			if usedExports[key] {
				isUsed = true
				break
			}
		}

		if !isUsed && len(fileNode.Exports) > 0 {
			result.UnusedFiles = append(result.UnusedFiles, UnusedFile{
				Path:   filePath,
				Reason: "None of the exports are imported by any file",
			})
		}
	}

	// Step 4: Find unused imports
	// An import is unused if no export in the file matches the import name
	for filePath, imports := range graph.Imports {
		fileNode := graph.Files[filePath]
		if fileNode == nil {
			continue
		}

		for _, imp := range imports {
			isUsed := false

			// Check if this import name matches any export in the same file
			for _, exp := range fileNode.Exports {
				if exp.Name == imp.Local || exp.Name == imp.Name {
					isUsed = true
					break
				}
			}

			// Also check if the import source matches any file path
			if !isUsed {
				for otherPath := range graph.Files {
					if strings.Contains(imp.Source, otherPath) ||
						strings.HasSuffix(imp.Source, filepath.Base(otherPath)) {
						isUsed = true
						break
					}
				}
			}

			if !isUsed {
				result.UnusedImports = append(result.UnusedImports, UnusedImport{
					File:   filePath,
					Import: imp.Name,
					Source: imp.Source,
					Line:   imp.Line,
					Reason: "Import not used in file",
				})
			}
		}
	}

	result.TotalIssues = len(result.UnusedExports) + len(result.UnusedFiles) + len(result.UnusedImports)
	return result
}

// matchesImport checks if an import matches an export
func matchesImport(imp Import, expName string, expFile string) bool {
	// Direct name match
	if imp.Local == expName || imp.Name == expName {
		return true
	}

	// Check if the import source matches the export file
	expFileName := strings.TrimSuffix(expFile, ".go")
	impSourceName := strings.Trim(imp.Source, `"`)

	// Handle relative imports
	if strings.HasPrefix(impSourceName, ".") {
		return false // Skip relative imports for now
	}

	// Handle package imports
	if impSourceName == expFileName {
		return true
	}

	// Check if import source contains the export file name
	if strings.Contains(impSourceName, expFileName) {
		return true
	}

	return false
}

// String returns a human-readable summary of the dead code analysis
func (r *DeadCodeResult) String() string {
	var sb strings.Builder

	sb.WriteString("\n=== Dead Code Analysis ===\n\n")

	if len(r.UnusedExports) > 0 {
		sb.WriteString(fmt.Sprintf("Unused Exports (%d):\n", len(r.UnusedExports)))
		sort.Slice(r.UnusedExports, func(i, j int) bool {
			if r.UnusedExports[i].File != r.UnusedExports[j].File {
				return r.UnusedExports[i].File < r.UnusedExports[j].File
			}
			return r.UnusedExports[i].Line < r.UnusedExports[j].Line
		})
		for _, ue := range r.UnusedExports {
			sb.WriteString(fmt.Sprintf(" %s:%d - %s (%s)\n", ue.File, ue.Line, ue.Export, ue.Kind))
		}
		sb.WriteString("\n")
	}

	if len(r.UnusedFiles) > 0 {
		sb.WriteString(fmt.Sprintf("Unused Files (%d):\n", len(r.UnusedFiles)))
		for _, uf := range r.UnusedFiles {
			sb.WriteString(fmt.Sprintf(" %s\n", uf.Path))
		}
		sb.WriteString("\n")
	}

	if len(r.UnusedImports) > 0 {
		sb.WriteString(fmt.Sprintf("Unused Imports (%d):\n", len(r.UnusedImports)))
		for _, ui := range r.UnusedImports {
			sb.WriteString(fmt.Sprintf(" %s:%d - %s from %s\n", ui.File, ui.Line, ui.Import, ui.Source))
		}
		sb.WriteString("\n")
	}

	if r.TotalIssues == 0 {
		sb.WriteString("No dead code issues found.\n")
	} else {
		sb.WriteString(fmt.Sprintf("Total issues: %d\n", r.TotalIssues))
	}

	return sb.String()
}

// GetSummary returns a brief summary for CLI output
func (r *DeadCodeResult) GetSummary() string {
	if r.TotalIssues == 0 {
		return "No dead code issues found"
	}

	var parts []string
	if len(r.UnusedExports) > 0 {
		parts = append(parts, fmt.Sprintf("%d unused exports", len(r.UnusedExports)))
	}
	if len(r.UnusedFiles) > 0 {
		parts = append(parts, fmt.Sprintf("%d unused files", len(r.UnusedFiles)))
	}
	if len(r.UnusedImports) > 0 {
		parts = append(parts, fmt.Sprintf("%d unused imports", len(r.UnusedImports)))
	}

	return strings.Join(parts, ", ")
}