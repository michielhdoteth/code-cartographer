package codegraph

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
)

type MapEntry struct {
	Node  Node    `json:"node"`
	Score float64 `json:"score"`
}

type ImportSummary struct {
	File    string   `json:"file"`
	Imports []string `json:"imports"`
	Score   float64  `json:"score"`
}

type RepoMap struct {
	RootPath       string          `json:"root_path"`
	TokenBudget    int             `json:"token_budget_estimate"`
	TokensUsed     int             `json:"tokens_used_estimate"`
	SymbolsOmitted int             `json:"symbols_omitted"`
	ImportsOmitted int             `json:"imports_omitted"`
	Entries        []MapEntry      `json:"entries"`
	Imports        []ImportSummary `json:"imports,omitempty"`
	Savings        *Savings        `json:"savings,omitempty"`
}

type PackOptions struct {
	TokenBudget int
	Include     []string
	Exclude     []string
	Compress    bool
}

type PackResult struct {
	RootPath      string `json:"root_path"`
	TokenBudget   int    `json:"token_budget_estimate"`
	TokensUsed    int    `json:"tokens_used_estimate"`
	FilesIncluded int    `json:"files_included"`
	FilesOmitted  int    `json:"files_omitted"`
	Compressed    bool   `json:"compressed"`
	Truncated     bool    `json:"truncated"`
	Content       string  `json:"content"`
	Savings       *Savings `json:"savings,omitempty"`
	Summary       string  `json:"summary,omitempty"`
	Tree          []string `json:"tree,omitempty"`
	TokenLedger   map[string]int `json:"token_ledger,omitempty"`
}

// GenerateRepoMap ranks declarations using PageRank over resolved calls and
// imports, then selects the highest-ranked signatures that fit the estimate.
func (index *Index) GenerateRepoMap(tokenBudget int, focus []string, useChurn ...bool) RepoMap {
	if tokenBudget <= 0 {
		tokenBudget = 2048
	}
	churn := map[string]int{}
	if len(useChurn) > 0 && useChurn[0] {
		churn = gitChurn(index.RootPath)
	}
	maxChurn := 0
	for _, count := range churn {
		if count > maxChurn {
			maxChurn = count
		}
	}
	churnBoost := func(file string) float64 {
		if maxChurn == 0 {
			return 0
		}
		return 0.1 * float64(churn[file]) / float64(maxChurn)
	}
	ranks := index.centrality()
	maxRank := 0.0
	for _, rank := range ranks {
		if rank > maxRank {
			maxRank = rank
		}
	}
	focus = normalizeFocus(focus)
	type candidate struct {
		score         float64
		file          string
		line          int
		kind          string
		entry         *MapEntry
		importSummary *ImportSummary
	}
	candidates := make([]candidate, 0)
	totalSymbols, totalImports := 0, 0
	for _, node := range index.Nodes {
		if node.Kind == "file" {
			continue
		}
		totalSymbols++
		score := normalizedRank(ranks[node.ID], maxRank) + churnBoost(node.File)
		if node.Exported {
			score += 0.05
		}
		for _, path := range focus {
			if node.File == path || strings.HasPrefix(node.File, path+"/") {
				score += 1
				break
			}
		}
		entry := MapEntry{Node: node, Score: score}
		candidates = append(candidates, candidate{score: score, file: node.File, line: node.StartLine, kind: "symbol", entry: &entry})
	}
	for file, parsed := range index.Files {
		if len(parsed.Imports) == 0 {
			continue
		}
		imports := make([]string, 0, len(parsed.Imports))
		for _, imp := range parsed.Imports {
			imports = append(imports, imp.Source)
		}
		sort.Strings(imports)
		imports = uniqueStrings(imports)
		summary := ImportSummary{File: file, Imports: imports, Score: normalizedRank(ranks[fileNodeID(file)], maxRank) + churnBoost(file)}
		for _, path := range focus {
			if file == path || strings.HasPrefix(file, path+"/") {
				summary.Score += 1
				break
			}
		}
		totalImports++
		candidates = append(candidates, candidate{score: summary.Score, file: file, kind: "imports", importSummary: &summary})
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].score != candidates[j].score {
			return candidates[i].score > candidates[j].score
		}
		if candidates[i].file != candidates[j].file {
			return candidates[i].file < candidates[j].file
		}
		if candidates[i].line != candidates[j].line {
			return candidates[i].line < candidates[j].line
		}
		if candidates[i].kind != candidates[j].kind {
			return candidates[i].kind < candidates[j].kind
		}
		return candidateText(candidates[i]) < candidateText(candidates[j])
	})

	result := RepoMap{RootPath: index.RootPath, TokenBudget: tokenBudget, Entries: []MapEntry{}, Imports: []ImportSummary{}}
	selected := make([]candidate, 0)
	used := estimatedTokens("# Repository map\n\n")
	for _, candidate := range candidates {
		cost := estimatedTokens(candidateText(candidate))
		if used+cost > tokenBudget {
			continue
		}
		selected = append(selected, candidate)
		used += cost
	}
	for {
		result.Entries = []MapEntry{}
		result.Imports = []ImportSummary{}
		for _, item := range selected {
			if item.entry != nil {
				result.Entries = append(result.Entries, *item.entry)
			} else if item.importSummary != nil {
				result.Imports = append(result.Imports, *item.importSummary)
			}
		}
		result.SymbolsOmitted = totalSymbols - len(result.Entries)
		result.ImportsOmitted = totalImports - len(result.Imports)
		result.TokensUsed = estimatedTokens(result.markdown())
		if result.TokensUsed <= tokenBudget || len(selected) == 0 {
			break
		}
		selected = selected[:len(selected)-1]
	}
	result.Savings = index.EstimateSavings(result.TokensUsed)
	return result
}

func normalizedRank(rank, maximum float64) float64 {
	if maximum == 0 {
		return 0
	}
	return rank / maximum
}

func candidateText(item struct {
	score         float64
	file          string
	line          int
	kind          string
	entry         *MapEntry
	importSummary *ImportSummary
}) string {
	if item.entry != nil {
		return renderMapEntry(*item.entry)
	}
	return fmt.Sprintf("- %s imports: %s\n", item.file, strings.Join(item.importSummary.Imports, ", "))
}

func uniqueStrings(values []string) []string {
	if len(values) == 0 {
		return values
	}
	result := values[:1]
	for _, value := range values[1:] {
		if value != result[len(result)-1] {
			result = append(result, value)
		}
	}
	return result
}

func renderMapEntry(entry MapEntry) string {
	label := strings.TrimSpace(entry.Node.Signature)
	if label == "" {
		label = entry.Node.Name
	}
	label = strings.ReplaceAll(label, "\n", " ")
	return fmt.Sprintf("- %s — %s:%d\n", label, entry.Node.File, entry.Node.StartLine)
}

// RenderMarkdown formats a map without exposing implementation bodies.
func (m RepoMap) RenderMarkdown() string {
	return truncateEstimated(m.markdown(), m.TokenBudget)
}

func (m RepoMap) markdown() string {
	var b strings.Builder
	b.WriteString("# Repository map\n\n")
	b.WriteString(fmt.Sprintf("Estimated tokens: %d / %d\n\n", m.TokensUsed, m.TokenBudget))
	for _, entry := range m.Entries {
		b.WriteString(renderMapEntry(entry))
	}
	if len(m.Imports) > 0 {
		b.WriteString("\n## Imports\n\n")
		for _, summary := range m.Imports {
			b.WriteString(fmt.Sprintf("- %s imports: %s\n", summary.File, strings.Join(summary.Imports, ", ")))
		}
	}
	if m.SymbolsOmitted > 0 {
		b.WriteString(fmt.Sprintf("\n… %d lower-ranked symbols omitted\n", m.SymbolsOmitted))
	}
	if m.ImportsOmitted > 0 {
		b.WriteString(fmt.Sprintf("… %d lower-ranked import summaries omitted\n", m.ImportsOmitted))
	}
	return b.String()
}

func (index *Index) GeneratePack(opts PackOptions) PackResult {
	if opts.TokenBudget <= 0 {
		opts.TokenBudget = 8000
	}
	result := PackResult{RootPath: index.RootPath, TokenBudget: opts.TokenBudget, Compressed: opts.Compress}
	paths := make([]string, 0, len(index.Files))
	for path := range index.Files {
		if matchesAny(path, opts.Exclude) || (len(opts.Include) > 0 && !matchesAny(path, opts.Include)) {
			result.FilesOmitted++
			continue
		}
		paths = append(paths, path)
	}
	sort.Strings(paths)
	includedPaths := []string{}
	var b strings.Builder
	b.WriteString("# Repository context\n\n")
	used := estimatedTokens(b.String())
	for i, path := range paths {
		file, err := os.ReadFile(filepath.Join(index.RootPath, filepath.FromSlash(path)))
		if err != nil {
			result.FilesOmitted++
			continue
		}
		body := string(file)
		if opts.Compress {
			body = index.compressedFile(path)
		}
		block := fmt.Sprintf("## %s\n\n%s\n\n", path, strings.TrimRight(body, "\r\n"))
		cost := estimatedTokens(block)
		if used+cost > opts.TokenBudget {
			remaining := opts.TokenBudget - used
			if remaining > 0 {
				block = truncateEstimated(block, remaining)
				b.WriteString(block)
				used = opts.TokenBudget
				result.FilesIncluded++
				includedPaths = append(includedPaths, path)
				result.FilesOmitted += len(paths) - i - 1
			} else {
				result.FilesOmitted += len(paths) - i
			}
			result.Truncated = true
			break
		}
		b.WriteString(block)
		used += cost
		result.FilesIncluded++
		includedPaths = append(includedPaths, path)
	}
	result.Content = b.String()
	if result.FilesOmitted > 0 {
		footer := fmt.Sprintf("\n… %d files omitted or truncated to fit the estimate.\n", result.FilesOmitted)
		if estimatedTokens(result.Content+footer) <= result.TokenBudget {
			result.Content += footer
		} else {
			result.Truncated = true
			footerTokens := estimatedTokens(footer)
			if footerTokens >= result.TokenBudget {
				result.Content = truncateEstimated(footer, result.TokenBudget)
			} else {
				result.Content = truncateEstimated(result.Content, result.TokenBudget-footerTokens) + footer
			}
		}
	}
	result.TokensUsed = estimatedTokens(result.Content)
	if result.TokensUsed > result.TokenBudget {
		result.Content = truncateEstimated(result.Content, result.TokenBudget)
		result.TokensUsed = estimatedTokens(result.Content)
	}
	result.Savings = index.EstimateSavings(result.TokensUsed)
	sourceTokens := 0
	for _, path := range includedPaths {
		if file, ok := index.Files[path]; ok {
			sourceTokens += int(file.Size / 4)
		}
	}
	result.Summary = fmt.Sprintf("%d files included, %d omitted, ~%d tokens", result.FilesIncluded, result.FilesOmitted, result.TokensUsed)
	result.Tree = includedPaths
	result.TokenLedger = map[string]int{
		"source_tokens":  sourceTokens,
		"emitted_tokens": result.TokensUsed,
		"budget":         result.TokenBudget,
	}
	return result
}

func (index *Index) compressedFile(path string) string {
	file := index.Files[path]
	var b strings.Builder
	for _, imp := range file.Imports {
		b.WriteString(fmt.Sprintf("import %q\n", imp.Source))
	}
	for _, node := range index.Outline(path) {
		if node.File != path {
			continue
		}
		label := strings.TrimSpace(node.Signature)
		if label == "" {
			label = node.Name
		}
		b.WriteString(fmt.Sprintf("%s %s\n", node.Kind, label))
	}
	if b.Len() == 0 {
		return "(no declarations or imports extracted)"
	}
	return strings.TrimRight(b.String(), "\n")
}

func (index *Index) centrality() map[string]float64 {
	ranks := make(map[string]float64)
	adjacency := make(map[string][]string)
	for _, node := range index.Nodes {
		ranks[node.ID] = 1
	}
	for _, edge := range index.Edges {
		if edge.Confidence != "exact" && edge.Confidence != "name" {
			continue
		}
		if _, ok := ranks[edge.Source]; !ok {
			continue
		}
		if _, ok := ranks[edge.Target]; !ok {
			continue
		}
		if edge.Type == "calls" || edge.Type == "imports" || edge.Type == "contains" {
			adjacency[edge.Source] = append(adjacency[edge.Source], edge.Target)
			adjacency[edge.Target] = append(adjacency[edge.Target], edge.Source)
		}
	}
	n := len(ranks)
	if n == 0 {
		return ranks
	}
	base := 1 / float64(n)
	ids := make([]string, 0, len(ranks))
	for id := range ranks {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		ranks[id] = base
	}
	for iteration := 0; iteration < 20; iteration++ {
		next := make(map[string]float64, n)
		for _, id := range ids {
			next[id] = 0.15 / float64(n)
		}
		for _, source := range ids {
			targets := adjacency[source]
			if len(targets) == 0 {
				continue
			}
			share := 0.85 * ranks[source] / float64(len(targets))
			for _, target := range targets {
				next[target] += share
			}
		}
		for _, id := range ids {
			ranks[id] = next[id]
		}
	}
	return ranks
}

func normalizeFocus(paths []string) []string {
	result := make([]string, 0, len(paths))
	for _, path := range paths {
		clean := filepath.ToSlash(filepath.Clean(path))
		clean = strings.TrimPrefix(clean, "./")
		if clean != "." && clean != "" {
			result = append(result, clean)
		}
	}
	sort.Strings(result)
	return result
}

// EstimateTokens returns a provider-neutral size estimate (~4 UTF-8 bytes per token).
func EstimateTokens(text string) int {
	if text == "" {
		return 0
	}
	return (len(text) + 3) / 4
}

func estimatedTokens(text string) int { return EstimateTokens(text) }

func truncateEstimated(text string, tokens int) string {
	if tokens <= 0 {
		return ""
	}
	maxBytes := tokens * 4
	if len(text) <= maxBytes {
		return text
	}
	cut := maxBytes
	for cut > 0 && cut < len(text) && (text[cut]&0xC0) == 0x80 {
		cut--
	}
	return strings.TrimRight(text[:cut], "\r\n")
}

func matchesAny(path string, patterns []string) bool {
	for _, pattern := range patterns {
		if globMatch(pattern, path) {
			return true
		}
	}
	return false
}

var globCache sync.Map

func globMatch(pattern, name string) bool {
	pattern = filepath.ToSlash(strings.TrimSpace(pattern))
	name = filepath.ToSlash(name)
	if pattern == "" {
		return false
	}
	if _, ok := globCache.Load(pattern); !ok {
		var b strings.Builder
		b.WriteString("^")
		for i := 0; i < len(pattern); i++ {
			switch pattern[i] {
			case '*':
				if i+1 < len(pattern) && pattern[i+1] == '*' {
					i++
					if i+1 < len(pattern) && pattern[i+1] == '/' {
						i++
						b.WriteString("(?:.*/)?")
					} else {
						b.WriteString(".*")
					}
				} else {
					b.WriteString("[^/]*")
				}
			case '?':
				b.WriteString("[^/]")
			default:
				b.WriteString(regexp.QuoteMeta(string(pattern[i])))
			}
		}
		b.WriteString("$")
		compiled, err := regexp.Compile(b.String())
		if err != nil {
			return false
		}
		globCache.Store(pattern, compiled)
	}
	compiled, ok := globCache.Load(pattern)
	return ok && compiled.(*regexp.Regexp).MatchString(name)
}
