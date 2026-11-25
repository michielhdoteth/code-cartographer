package codegraph

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"carto/config"
	"carto/retrieval"
)

func (index *Index) retrievalState() (*retrieval.Cache, retrieval.Settings) {
	settings := retrieval.DefaultSettings()
	if cfg, err := config.Load(index.RootPath); err == nil {
		settings = cfg.Retrieval
	}
	settings.FillDefaults()
	fileHashes := make(map[string]string, len(index.Files))
	for path, file := range index.Files {
		fileHashes[path] = file.Hash
	}
	graphHash := retrieval.ComputeGraphHash(fileHashes)
	settingsHash := settings.Hash()
	if cache, err := retrieval.LoadCache(index.RootPath, graphHash, settingsHash); err == nil {
		return cache, settings
	}
	docs := index.retrievalDocuments(settings)
	cache := retrieval.BuildCache(graphHash, settingsHash, docs)
	// Cache writes are best-effort: read-only checkouts must still support search.
	_ = retrieval.SaveCache(index.RootPath, cache)
	return cache, settings
}

func (index *Index) retrievalDocuments(settings retrieval.Settings) []retrieval.Document {
	commentCache := make(map[string][]string)
	docs := make([]retrieval.Document, 0, len(index.Nodes))
	for _, node := range index.Nodes {
		if node.Kind == "file" {
			continue
		}
		body := nodeText(index.RootPath, node)
		bodyTokens := retrieval.Tokenize(body)
		if settings.BodyCap > 0 && len(bodyTokens) > settings.BodyCap {
			bodyTokens = bodyTokens[:settings.BodyCap]
		}
		qualified := strings.TrimSuffix(node.File, filepath.Ext(node.File)) + "::" + node.Name
		docTokens := commentCache[node.File]
		if docTokens == nil {
			docTokens = precedingComments(index.RootPath, node.File, node.StartLine)
			commentCache[node.File] = docTokens
		}
		fields := map[string][]string{
			"name":      append(retrieval.Tokenize(node.Name), retrieval.RawTokens(node.Name)...),
			"qualified": append(retrieval.Tokenize(qualified), retrieval.RawTokens(qualified)...),
			"signature": append(retrieval.Tokenize(node.Signature), retrieval.RawTokens(node.Signature)...),
			"doc":       docTokens,
			"path":      append(retrieval.Tokenize(node.File), retrieval.RawTokens(node.File)...),
			"kind":      retrieval.Tokenize(node.Kind),
			"body":      bodyTokens,
		}
		docs = append(docs, retrieval.Document{ID: node.ID, Fields: fields})
	}
	return docs
}

func precedingComments(root, file string, line int) []string {
	if line <= 1 {
		return nil
	}
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(file)))
	if err != nil {
		return nil
	}
	lines := strings.Split(string(data), "\n")
	end := line - 1
	if end > len(lines) {
		end = len(lines)
	}
	start := end - 8
	if start < 0 {
		start = 0
	}
	var tokens []string
	for i := start; i < end; i++ {
		trimmed := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") ||
			strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*") {
			tokens = append(tokens, retrieval.Tokenize(trimmed)...)
		}
	}
	return tokens
}

func queryTokens(query string, settings retrieval.Settings, stats retrieval.Stats) []string {
	tokens := retrieval.Tokenize(query)
	if settings.QueryCap > 0 && len(tokens) > settings.QueryCap {
		tokens = tokens[:settings.QueryCap]
	}
	if stats.DocCount == 0 {
		return tokens
	}
	filtered := make([]string, 0, len(tokens))
	for _, token := range tokens {
		ratio := float64(stats.DocFreq[token]) / float64(stats.DocCount)
		if ratio > 0.9 {
			continue
		}
		filtered = append(filtered, token)
	}
	if len(filtered) == 0 {
		return tokens
	}
	return filtered
}

func rankIDs(ranked []retrieval.Ranked, limit int) []string {
	if limit <= 0 || limit > len(ranked) {
		limit = len(ranked)
	}
	out := make([]string, 0, limit)
	for _, r := range ranked[:limit] {
		out = append(out, r.ID)
	}
	return out
}

func symbolRoute(query string, nodes []Node, limit int) []string {
	q := strings.ToLower(strings.TrimSpace(query))
	type candidate struct {
		id    string
		score float64
	}
	var candidates []candidate
	for _, node := range nodes {
		if node.Kind == "file" {
			continue
		}
		name := strings.ToLower(node.Name)
		score := 0.0
		if name == q {
			score = 100
		} else if strings.HasPrefix(name, q) && len(q) >= 3 {
			score = 50
		} else if strings.Contains(name, q) && len(q) >= 3 {
			score = 20
		}
		if score > 0 {
			candidates = append(candidates, candidate{node.ID, score})
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].score != candidates[j].score {
			return candidates[i].score > candidates[j].score
		}
		return candidates[i].id < candidates[j].id
	})
	out := make([]string, 0, minInt(limit, len(candidates)))
	for _, c := range candidates {
		out = append(out, c.id)
		if len(out) >= limit {
			break
		}
	}
	return out
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// nameOverlapRoute ranks documents by how many query tokens occur in the symbol
// name. It gives natural-language queries a symbol-anchored signal.
func nameOverlapRoute(queryTokens []string, docs map[string]retrieval.Document, limit int) []string {
	type candidate struct {
		id    string
		score int
	}
	var candidates []candidate
	for id, doc := range docs {
		nameSet := map[string]bool{}
		for _, token := range doc.Fields["name"] {
			nameSet[token] = true
		}
		score := 0
		for _, token := range queryTokens {
			if nameSet[token] {
				score++
			}
		}
		if score > 0 {
			candidates = append(candidates, candidate{id, score})
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].score != candidates[j].score {
			return candidates[i].score > candidates[j].score
		}
		return candidates[i].id < candidates[j].id
	})
	out := make([]string, 0, minInt(limit, len(candidates)))
	for _, c := range candidates {
		out = append(out, c.id)
		if len(out) >= limit {
			break
		}
	}
	return out
}

func fuzzyRoute(query string, nodes []Node, limit int) []string {
	trimmed := strings.TrimSpace(query)
	if len([]rune(trimmed)) < 2 || strings.ContainsAny(trimmed, " \t\n") {
		return nil
	}
	type candidate struct {
		id    string
		score float64
	}
	var candidates []candidate
	for _, node := range nodes {
		if node.Kind == "file" {
			continue
		}
		score := retrieval.FuzzyScore(trimmed, node.Name)
		if score > 0 {
			candidates = append(candidates, candidate{node.ID, score})
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].score != candidates[j].score {
			return candidates[i].score > candidates[j].score
		}
		return candidates[i].id < candidates[j].id
	})
	out := make([]string, 0, minInt(limit, len(candidates)))
	for _, c := range candidates {
		out = append(out, c.id)
		if len(out) >= limit {
			break
		}
	}
	return out
}

func (index *Index) rank(query string, limit, tokenBudget int) []QueryHit {
	cache, settings := index.retrievalState()
	docs := make([]retrieval.Document, 0, len(cache.Documents))
	docByID := make(map[string]retrieval.Document, len(cache.Documents))
	for _, doc := range cache.Documents {
		docs = append(docs, doc)
		docByID[doc.ID] = doc
	}
	sort.Slice(docs, func(i, j int) bool { return docs[i].ID < docs[j].ID })

	tokens := queryTokens(query, settings, cache.Stats)
	depth := retrieval.AdaptiveDepth(limit)
	ranked := settings.Scorer(cache.Stats).Rank(tokens, docs)
	lexical := rankIDs(ranked, depth)
	symbolIDs := symbolRoute(query, index.Nodes, depth)
	nameIDs := nameOverlapRoute(tokens, cache.Documents, depth)
	fuzzyIDs := fuzzyRoute(query, index.Nodes, depth)
	fused := retrieval.WeightedRRF([][]string{lexical, symbolIDs, nameIDs, fuzzyIDs}, []float64{5, 1, 1, 2}, 60)

	centrality := index.centrality()
	maxRank := 0.0
	for _, value := range centrality {
		if value > maxRank {
			maxRank = value
		}
	}
	q := strings.ToLower(strings.TrimSpace(query))
	qSplit := retrieval.Tokenize(query)
	qRaw := retrieval.RawTokens(query)

	type scored struct {
		id    string
		score float64
	}
	var candidates []scored
	for id, rrfScore := range fused {
		node, ok := index.nodeByID(id)
		if !ok || node.Kind == "file" {
			continue
		}
		score := rrfScore * 100
		if strings.EqualFold(q, node.Name) {
			score += 12
		}
		if q != "" && q == node.Name {
			score += 6
		}
		if generatedPath(node.File) {
			score -= 5
		}
		if len(qRaw) == 1 {
			for _, raw := range retrieval.RawTokens(node.Name) {
				if raw == qRaw[0] {
					score += 6
					break
				}
			}
		}
		doc := docByID[id]
		if phrasePresent(qSplit, doc.Fields["name"]) {
			score += 2
		}
		if phrasePresent(qSplit, doc.Fields["body"]) {
			score += 2
		}
		if q != "" && (q == strings.ToLower(node.File) || q == strings.ToLower(filepath.Base(node.File))) {
			score += 3
		}
		if maxRank > 0 {
			score += centrality[id] / maxRank * 0.3
		}
		if node.Exported {
			score += 0.05
		}
		for _, flag := range node.Flags {
			if flag == "test" {
				score -= 2
			}
		}
		candidates = append(candidates, scored{id, score})
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].score != candidates[j].score {
			return candidates[i].score > candidates[j].score
		}
		return candidates[i].id < candidates[j].id
	})

	used := 0
	selected := make([]QueryHit, 0, limit)
	for _, candidate := range candidates {
		if len(selected) >= limit || used >= tokenBudget {
			break
		}
		node, _ := index.nodeByID(candidate.id)
		snippet := nodeText(index.RootPath, node)
		remaining := tokenBudget - used
		if EstimateTokens(snippet) > remaining {
			if remaining <= 0 {
				continue
			}
			snippet = truncateEstimated(snippet, remaining)
		}
		used += EstimateTokens(snippet)
		selected = append(selected, QueryHit{Node: node, Score: candidate.score, Snippet: snippet})
	}
	return selected
}

func generatedPath(path string) bool {
	p := strings.ToLower(filepath.ToSlash(path))
	for _, prefix := range []string{"bin/", "build/", "archive/", "node_modules/", "dist/", "test-data/", ".carto/"} {
		if strings.HasPrefix(p, prefix) {
			return true
		}
	}
	return strings.HasSuffix(p, ".min.js") || strings.Contains(p, "/bundle.")
}

func phrasePresent(needle, haystack []string) bool {
	if len(needle) == 0 || len(needle) > len(haystack) {
		return false
	}
	for i := 0; i+len(needle) <= len(haystack); i++ {
		match := true
		for j := range needle {
			if needle[j] != haystack[i+j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}











