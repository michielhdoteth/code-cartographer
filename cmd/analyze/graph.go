package analyze

import (
	"carto/lib/parser"
	"fmt"
	"path/filepath"
	"strings"
)

// ModuleGraph represents the dependency graph of a codebase
type ModuleGraph struct {
	Files    map[string]*FileNode
	Exports  map[string]ExportRef // "file:exportName" -> references
	Imports  map[string][]Import  // file -> imports
	Language string
	RootPath string
}

// FileNode represents a single source file in the graph
type FileNode struct {
	Path    string
	Name    string
	Language string
	Exports  []Export
	Imports  []Import
}

// Export represents an exported symbol
type Export struct {
	Name string
	Kind string // "function", "class", "const", "type", "var", "interface"
	Line int
	Doc  string
}

// Import represents an imported symbol
type Import struct {
	Name   string // imported name (may be aliased)
	Source string // module path or relative path
	Local  string // local name if aliased
	Line   int
}

// ExportRef tracks where an export is referenced
type ExportRef struct {
	File   string
	Export string
	Line   int
}

// NewModuleGraph creates a new empty module graph
func NewModuleGraph(rootPath string, language string) *ModuleGraph {
	return &ModuleGraph{
		Files:    make(map[string]*FileNode),
		Exports:  make(map[string]ExportRef),
		Imports:  make(map[string][]Import),
		Language: language,
		RootPath: rootPath,
	}
}

// BuildGraph scans a directory and builds the dependency graph
func BuildGraph(rootPath string, language string) (*ModuleGraph, error) {
	graph := NewModuleGraph(rootPath, language)

	// Use the unified parser
	codebase, err := parser.ParseDirectory(rootPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to parse directory: %w", err)
	}

	// Convert parser.ParseResult to FileNode
	for filePath, result := range codebase.Files {
		fileNode := &FileNode{
			Path:    filePath,
			Name:    filepath.Base(filePath),
			Language: string(result.Language),
			Exports:  []Export{},
			Imports:  []Import{},
		}

		// Convert exports
		for _, exp := range result.Exports {
			fileNode.Exports = append(fileNode.Exports, Export{
				Name: exp.Name,
				Kind: string(exp.Type),
				Line: exp.Line,
			})
		}

		// Convert imports
		for _, imp := range result.Imports {
			fileNode.Imports = append(fileNode.Imports, Import{
				Name:   imp.Name,
				Source: imp.Source,
				Local:  imp.Alias,
				Line:   imp.Line,
			})
		}

		graph.Files[filePath] = fileNode

		// Register exports
		for _, exp := range fileNode.Exports {
			key := filePath + ":" + exp.Name
			graph.Exports[key] = ExportRef{
				File:   filePath,
				Export: exp.Name,
				Line:   exp.Line,
			}
		}

		// Register imports
		graph.Imports[filePath] = fileNode.Imports
	}

	return graph, nil
}

// BuildGraphFromCodebase builds a graph from an already-parsed codebase
func BuildGraphFromCodebase(codebase *parser.Codebase) *ModuleGraph {
	graph := NewModuleGraph(codebase.RootPath, string(codebase.Language))

	for filePath, result := range codebase.Files {
		fileNode := &FileNode{
			Path:    filePath,
			Name:    filepath.Base(filePath),
			Language: string(result.Language),
			Exports:  []Export{},
			Imports:  []Import{},
		}

		for _, exp := range result.Exports {
			fileNode.Exports = append(fileNode.Exports, Export{
				Name: exp.Name,
				Kind: string(exp.Type),
				Line: exp.Line,
			})
		}

		for _, imp := range result.Imports {
			fileNode.Imports = append(fileNode.Imports, Import{
				Name:   imp.Name,
				Source: imp.Source,
				Local:  imp.Alias,
				Line:   imp.Line,
			})
		}

		graph.Files[filePath] = fileNode

		for _, exp := range fileNode.Exports {
			key := filePath + ":" + exp.Name
			graph.Exports[key] = ExportRef{
				File:   filePath,
				Export: exp.Name,
				Line:   exp.Line,
			}
		}

		graph.Imports[filePath] = fileNode.Imports
	}

	return graph
}

// GetFilesByLanguage returns all files of a specific language
func (g *ModuleGraph) GetFilesByLanguage(lang string) []*FileNode {
	var result []*FileNode
	for _, file := range g.Files {
		if strings.ToLower(file.Language) == strings.ToLower(lang) {
			result = append(result, file)
		}
	}
	return result
}

// GetExportedFunctions returns all exported functions
func (g *ModuleGraph) GetExportedFunctions() []Export {
	var result []Export
	for _, file := range g.Files {
		for _, exp := range file.Exports {
			if exp.Kind == "function" {
				result = append(result, exp)
			}
		}
	}
	return result
}

// GetTotalLinesOfCode returns total lines of code
func (g *ModuleGraph) GetTotalLinesOfCode() int {
	// This would need to come from the parser metrics
	return 0
}