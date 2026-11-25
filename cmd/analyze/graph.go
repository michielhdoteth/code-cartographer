package analyze

import (
	"carto/lib/parser"
	"fmt"
	"path/filepath"
	"strings"
)

type ModuleGraph struct {
	Files    map[string]*FileNode
	Exports  map[string]ExportRef // "file:exportName" -> references
	Imports  map[string][]Import
	Language string
	RootPath string
}

type FileNode struct {
	Path    string
	Name    string
	Language string
	Exports  []Export
	Imports  []Import
}

type Export struct {
	Name string
	Kind string // "function", "class", "const", "type", "var", "interface"
	Line int
	Doc  string
}

type Import struct {
	Name   string // imported name (may be aliased)
	Source string // module path or relative path
	Local  string // local name if aliased
	Line   int
}

type ExportRef struct {
	File   string
	Export string
	Line   int
}

func NewModuleGraph(rootPath string, language string) *ModuleGraph {
	return &ModuleGraph{
		Files:    make(map[string]*FileNode),
		Exports:  make(map[string]ExportRef),
		Imports:  make(map[string][]Import),
		Language: language,
		RootPath: rootPath,
	}
}

func BuildGraph(rootPath string, language string) (*ModuleGraph, error) {
	graph := NewModuleGraph(rootPath, language)

	codebase, err := parser.ParseDirectory(rootPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to parse directory: %w", err)
	}

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

	resolveGoImportSources(rootPath, graph)
	return graph, nil
}

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

func (g *ModuleGraph) GetFilesByLanguage(lang string) []*FileNode {
	var result []*FileNode
	for _, file := range g.Files {
		if strings.ToLower(file.Language) == strings.ToLower(lang) {
			result = append(result, file)
		}
	}
	return result
}

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

func (g *ModuleGraph) GetTotalLinesOfCode() int {
	total := 0
	for _, file := range g.Files {
		// Count exports as a proxy for lines of code
		total += len(file.Exports) * 10
	}
	return total
}

func (g *ModuleGraph) GetSymbolInfo(name string) []ExportRef {
	var refs []ExportRef
	for key, ref := range g.Exports {
		if strings.Contains(key, ":"+name) || ref.Export == name {
			refs = append(refs, ref)
		}
	}
	return refs
}

func (g *ModuleGraph) GetCallGraph(filePath string) (callers []string, callees []string) {
	for f, imports := range g.Imports {
		for _, imp := range imports {
			if strings.Contains(imp.Source, filePath) || strings.Contains(f, filePath) {
				callers = append(callers, f)
			}
		}
	}
	return callers, callees
}

func (g *ModuleGraph) GetAllSymbols() []Export {
	var result []Export
	for _, file := range g.Files {
		result = append(result, file.Exports...)
	}
	return result
}

func (g *ModuleGraph) GetSymbolsByType(symbolType string) []Export {
	var result []Export
	for _, file := range g.Files {
		for _, exp := range file.Exports {
			if exp.Kind == symbolType {
				result = append(result, exp)
			}
		}
	}
	return result
}

func (g *ModuleGraph) GetStats() map[string]interface{} {
	byLang := map[string]int{}
	byType := map[string]int{}
	totalExports := 0
	totalImports := 0

	for _, file := range g.Files {
		byLang[file.Language]++
		totalExports += len(file.Exports)
		totalImports += len(file.Imports)
		for _, exp := range file.Exports {
			byType[exp.Kind]++
		}
	}

	return map[string]interface{}{
		"total_files":    len(g.Files),
		"total_exports":  totalExports,
		"total_imports":  totalImports,
		"by_language":    byLang,
		"by_type":        byType,
	}
}