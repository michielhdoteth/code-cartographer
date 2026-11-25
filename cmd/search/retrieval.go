package search

import (
	"sort"
	"strings"

	"carto/config"
	"carto/retrieval"
)

func (idx *Index) rank(query string, maxResults int) []SearchResult {
	tokens := retrieval.Tokenize(query)
	if len(tokens) == 0 {
		return nil
	}
	settings := retrieval.DefaultSettings()
	if cfg, err := config.Load(idx.RootPath); err == nil {
		settings = cfg.Retrieval
	}
	settings.FillDefaults()

	termsByFile := map[string][]string{}
	for term, entries := range idx.Terms {
		for _, entry := range entries {
			termsByFile[entry.File] = append(termsByFile[entry.File], term)
		}
	}
	docs := make([]retrieval.Document, 0, len(termsByFile))
	for file, terms := range termsByFile {
		sort.Strings(terms)
		docs = append(docs, retrieval.Document{
			ID: file,
			Fields: map[string][]string{
				"name": retrieval.Tokenize(file),
				"path": retrieval.Tokenize(file),
				"body": terms,
				"kind": {"file"},
			},
		})
	}
	sort.Slice(docs, func(i, j int) bool { return docs[i].ID < docs[j].ID })
	scorer := settings.Scorer(retrieval.NewStats(docs))

	bestLines := map[string]int{}
	bestContexts := map[string]string{}
	for _, word := range tokens {
		for term, entries := range idx.Terms {
			if strings.Contains(strings.ToLower(term), word) {
				for _, entry := range entries {
					if bestLines[entry.File] == 0 || entry.Line < bestLines[entry.File] {
						bestLines[entry.File] = entry.Line
						bestContexts[entry.File] = entry.Context
					}
				}
			}
		}
	}

	type scored struct {
		file  string
		score float64
	}
	var sorted []scored
	for _, doc := range docs {
		score := scorer.Score(tokens, doc)
		if score > 0 {
			sorted = append(sorted, scored{doc.ID, score})
		}
	}
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].score != sorted[j].score {
			return sorted[i].score > sorted[j].score
		}
		return sorted[i].file < sorted[j].file
	})
	results := make([]SearchResult, 0, maxResults)
	for i, s := range sorted {
		if i >= maxResults {
			break
		}
		results = append(results, SearchResult{File: s.file, Line: bestLines[s.file], Context: bestContexts[s.file], Score: s.score})
	}
	return results
}
