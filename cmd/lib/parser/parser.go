package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Parser interface {
	Parse(code string, filePath string) (*ParseResult, error)

	Language() Language

	SupportedExtensions() []string
}

var parserRegistry = make(map[Language]Parser)

func RegisterParser(p Parser) {
	parserRegistry[p.Language()] = p
}

func GetParser(lang Language) Parser {
	if p, ok := parserRegistry[lang]; ok {
		return p
	}
	return nil
}

func GetSupportedLanguages() []Language {
	languages := make([]Language, 0, len(parserRegistry))
	for lang := range parserRegistry {
		languages = append(languages, lang)
	}
	return languages
}

func ParseFile(filePath string) (*ParseResult, error) {
	lang := DetectLanguage(filePath)
	p := GetParser(lang)
	if p == nil {
		return nil, fmt.Errorf("no parser for language: %s (file: %s)", lang, filePath)
	}

	code, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	result, err := p.Parse(string(code), filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", filePath, err)
	}

	return result, nil
}

func ParseDirectory(dirPath string, langFilter []Language) (*Codebase, error) {
	codebase := &Codebase{
		RootPath: dirPath,
		Files:    make(map[string]*ParseResult),
		Graph: &DependencyGraph{
			Nodes: make(map[string]*GraphNode),
			Edges: []GraphEdge{},
		},
	}

	langCounts := make(map[string]int)

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			name := info.Name()
			if name == "node_modules" || name == ".git" || name == "vendor" ||
				name == "dist" || name == "build" || name == "__pycache__" ||
				name == "target" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}

		lang := DetectLanguage(path)
		if lang == LanguageUnknown {
			return nil
		}

		if len(langFilter) > 0 {
			found := false
			for _, f := range langFilter {
				if f == lang {
					found = true
					break
				}
			}
			if !found {
				return nil
			}
		}

		relPath, _ := filepath.Rel(dirPath, path)
		result, err := ParseFile(path)
		if err != nil {
			// Log error but continue
			fmt.Printf("Warning: failed to parse %s: %v\n", path, err)
			return nil
		}

		codebase.Files[relPath] = result
		codebase.Language = lang
		langCounts[string(lang)]++

		node := &GraphNode{
			ID:      relPath,
			File:    relPath,
			Exports: make([]string, len(result.Exports)),
			Imports: make([]string, len(result.Imports)),
			Metrics: result.Metrics,
		}
		for i, exp := range result.Exports {
			node.Exports[i] = exp.Name
		}
		for i, imp := range result.Imports {
			node.Imports[i] = imp.Source
		}
		codebase.Graph.Nodes[relPath] = node

		for _, edge := range result.Edges {
			codebase.Graph.Edges = append(codebase.Graph.Edges, GraphEdge{
				Source: edge.Source,
				Target: edge.Target,
				Type:   edge.Type,
			})
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	codebase.Stats.TotalFiles = len(codebase.Files)
	codebase.Stats.LanguageCounts = langCounts

	for _, file := range codebase.Files {
		codebase.Stats.TotalNodes += len(file.Nodes)
		codebase.Stats.TotalEdges += len(file.Edges)
		codebase.Stats.TotalExports += len(file.Exports)
		codebase.Stats.TotalImports += len(file.Imports)
		codebase.Stats.LinesOfCode += file.Metrics.LinesOfCode
	}

	return codebase, nil
}

func BuildGraph(codebase *Codebase) *DependencyGraph {
	graph := &DependencyGraph{
		Nodes: make(map[string]*GraphNode),
		Edges: []GraphEdge{},
	}

	for path, file := range codebase.Files {
		node := &GraphNode{
			ID:      path,
			File:    path,
			Exports: make([]string, 0),
			Imports: make([]string, 0),
			Metrics: file.Metrics,
		}

		for _, exp := range file.Exports {
			node.Exports = append(node.Exports, exp.Name)
		}

		for _, imp := range file.Imports {
			node.Imports = append(node.Imports, imp.Source)
		}

		graph.Nodes[path] = node
	}

	for path, file := range codebase.Files {
		for _, imp := range file.Imports {
			for otherPath, otherFile := range codebase.Files {
				if path == otherPath {
					continue
				}
				for _, exp := range otherFile.Exports {
					if exp.Name == imp.Name || imp.Source == otherPath {
						graph.Edges = append(graph.Edges, GraphEdge{
							Source: path,
							Target: otherPath,
							Type:   EdgeTypeImports,
						})
						break
					}
				}
			}
		}
	}

	return graph
}