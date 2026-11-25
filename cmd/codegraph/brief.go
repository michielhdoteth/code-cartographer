package codegraph

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

type Brief struct {
	SchemaVersion int          `json:"schema_version"`
	RootPath      string       `json:"root_path"`
	TokenBudget   int          `json:"token_budget_estimate"`
	TokensUsed    int          `json:"tokens_used_estimate"`
	Focus         []string     `json:"focus,omitempty"`
	Files         []BriefFile  `json:"files"`
	MapEntries    int          `json:"map_entries"`
	SuggestedNext string       `json:"suggested_next"`
	Savings       *Savings     `json:"savings,omitempty"`
	Communities   [][]string   `json:"communities,omitempty"`
}

type BriefFile struct {
	Path         string   `json:"path"`
	Lines        int      `json:"lines"`
	Symbols      int      `json:"symbols"`
	Imports      int      `json:"imports"`
	TopSymbols   []string `json:"top_symbols"`
}

func (index *Index) GenerateBrief(budget, maxFiles int, focus []string) Brief {
	if budget <= 0 {
		budget = 1200
	}
	if maxFiles <= 0 {
		maxFiles = 30
	}
	result := Brief{
		SchemaVersion: 1, RootPath: index.RootPath, TokenBudget: budget,
		Focus: focus, Files: []BriefFile{},
		SuggestedNext: "carto context <file> --outline --json",
	}

	symbolsByFile := map[string][]string{}
	for _, node := range index.Nodes {
		if node.Kind == "file" {
			continue
		}
		symbolsByFile[node.File] = append(symbolsByFile[node.File], node.Name)
	}
	centrality := index.centrality()
	type rankedFile struct {
		path  string
		score float64
	}
	var ranked []rankedFile
	for path := range index.Files {
		ranked = append(ranked, rankedFile{path, centrality[briefFileNodeID(path)]})
	}
	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].score != ranked[j].score {
			return ranked[i].score > ranked[j].score
		}
		return ranked[i].path < ranked[j].path
	})

	for _, item := range ranked {
		file := index.Files[item.path]
		info := BriefFile{
			Path: item.path, Lines: file.Lines, Imports: len(file.Imports),
			TopSymbols: []string{},
		}
		names := append([]string{}, symbolsByFile[item.path]...)
		sort.Strings(names)
		info.Symbols = len(names)
		if len(names) > 3 {
			names = names[:3]
		}
		info.TopSymbols = names
		result.Files = append(result.Files, info)
		if len(result.Files) >= maxFiles {
			break
		}
	}
	repoMap := index.GenerateRepoMap(budget/2, focus)
	result.MapEntries = len(repoMap.Entries)
	result.Communities = fileCommunities(index)

	for len(result.Files) > 0 {
		result.TokensUsed = EstimateTokens(result.RenderMarkdown())
		if result.TokensUsed <= budget {
			break
		}
		result.Files = result.Files[:len(result.Files)-1]
	}
	result.TokensUsed = EstimateTokens(result.RenderMarkdown())
	result.Savings = index.EstimateSavings(result.TokensUsed)
	return result
}

func fileCommunities(index *Index) [][]string {
	parent := make(map[string]string)
	find := func(x string) string {
		root := x
		for parent[root] != "" && parent[root] != root {
			root = parent[root]
		}
		for parent[x] != "" && parent[x] != root {
			next := parent[x]
			parent[x] = root
			x = next
		}
		return root
	}
	for path := range index.Files {
		parent[path] = path
	}
	for _, edge := range index.Edges {
		if edge.Type != "imports" {
			continue
		}
		a := strings.TrimPrefix(edge.Source, "file:")
		b := strings.TrimPrefix(edge.Target, "file:")
		if _, ok := parent[a]; !ok {
			continue
		}
		if _, ok := parent[b]; !ok {
			continue
		}
		ra, rb := find(a), find(b)
		if ra != rb {
			parent[ra] = rb
		}
	}
	groups := make(map[string][]string)
	for path := range parent {
		root := find(path)
		groups[root] = append(groups[root], path)
	}
	out := make([][]string, 0, len(groups))
	for _, group := range groups {
		sort.Strings(group)
		out = append(out, group)
	}
	sort.Slice(out, func(i, j int) bool {
		if len(out[i]) != len(out[j]) {
			return len(out[i]) > len(out[j])
		}
		return out[i][0] < out[j][0]
	})
	return out
}

func briefFileNodeID(path string) string {
	return "file:" + filepath.ToSlash(path)
}

func (b Brief) RenderMarkdown() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "=== carto brief ===\nTokens: %d / %d\n\n", b.TokensUsed, b.TokenBudget)
	fmt.Fprintf(&sb, "Repo map entries: %d · Files: %d · Subsystems: %d\n\n", b.MapEntries, len(b.Files), len(b.Communities))
	for i, group := range b.Communities {
		if i >= 5 {
			break
		}
		fmt.Fprintf(&sb, "  subsystem %d: %s\n", i+1, strings.Join(group, ", "))
	}
	if len(b.Communities) > 0 {
		fmt.Fprintln(&sb)
	}
	for _, file := range b.Files {
		fmt.Fprintf(&sb, "  %-50s %5d ln  %3d sym  %3d imp  %s\n",
			file.Path, file.Lines, file.Symbols, file.Imports, strings.Join(file.TopSymbols, ", "))
	}
	fmt.Fprintf(&sb, "\nnext: %s\n", b.SuggestedNext)
	return sb.String()
}
