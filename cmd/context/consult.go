package context

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"carto/codegraph"
)

type Selection struct {
	FilePath      string   `json:"file_path"`
	Ranges        []Range  `json:"ranges"`
	TokensUsed    int      `json:"tokens_used_estimate"`
	SuggestedNext string   `json:"suggested_next"`
	Content       string   `json:"content"`
}

type Range struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

// generateView handles the progressive-disclosure views. The default path in
// Generate stays untouched when no view flag is set.
func generateView(opts Options, absPath string) (*Result, error) {
	root := opts.RootPath
	if root == "" {
		root = "."
	}
	index, err := codegraph.Build(root)
	if err != nil {
		return nil, err
	}
	rel, err := filepath.Rel(root, absPath)
	if err != nil {
		rel = absPath
	}
	rel = filepath.ToSlash(rel)

	result := &Result{
		FilePath:    absPath,
		TokenBudget: opts.TokenBudget,
		Blocks:      []ContextBlock{},
	}

	switch {
	case opts.Outline:
		skeleton := index.GenerateSkeleton(rel, opts.TokenBudget)
		result.View = "skeleton"
		result.Skeleton = &skeleton
		result.TokensUsed = skeleton.TokensUsed
		result.SuggestedNext = skeleton.SuggestedNext
		result.Savings = extendSavings(result.Savings, skeleton.Savings)
	case opts.ChunkIndex >= 0:
		chunk, err := index.GenerateChunk(rel, opts.ChunkIndex, opts.ChunkBudget, 64)
		if err != nil {
			return nil, err
		}
		result.View = "chunk"
		result.Chunk = &chunk
		result.TokensUsed = chunk.Tokens
		result.SuggestedNext = chunk.SuggestedNext
		result.Savings = extendSavings(result.Savings, chunk.Savings)
	default:
		ranges, err := selectionRanges(index, rel, opts)
		if err != nil {
			return nil, err
		}
		lines := readLines(absPath)
		var content strings.Builder
		var emitted []Range
		used := 0
		for _, r := range ranges {
			if r.Start < 1 || r.End < r.Start || r.End > len(lines) {
				continue
			}
			chunk := strings.Join(lines[r.Start-1:r.End], "\n")
			chunkTokens := codegraph.EstimateTokens(chunk)
			if used+chunkTokens > opts.TokenBudget {
				remaining := opts.TokenBudget - used
				if remaining <= 0 {
					break
				}
				chunk = truncateChunk(chunk, remaining)
				chunkTokens = codegraph.EstimateTokens(chunk)
			}
			content.WriteString(chunk)
			content.WriteString("\n")
			used += chunkTokens
			emitted = append(emitted, r)
		}
		result.View = "selection"
		result.Selection = &Selection{
			FilePath: rel, Ranges: emitted, TokensUsed: used,
			Content: content.String(), SuggestedNext: fmt.Sprintf("carto context %s --outline --json", rel),
		}
		result.TokensUsed = used
		result.SuggestedNext = result.Selection.SuggestedNext
		if file, ok := index.Files[rel]; ok {
			result.Savings = codegraph.BuildSavings(used, sumRanges(emitted), int(file.Size/4), file.Lines, "file")
		}
	}
	result.Truncated = result.TokensUsed >= opts.TokenBudget
	return result, nil
}

func selectionRanges(index *codegraph.Index, rel string, opts Options) ([]Range, error) {
	var ranges []Range
	for _, spec := range opts.Lines {
		r, err := parseRange(spec)
		if err != nil {
			return nil, err
		}
		ranges = append(ranges, r)
	}
	if len(opts.Symbols) > 0 {
		want := make(map[string]bool)
		for _, name := range opts.Symbols {
			for _, part := range strings.Split(name, ",") {
				if trimmed := strings.TrimSpace(part); trimmed != "" {
					want[trimmed] = true
				}
			}
		}
		for _, node := range index.Outline(rel) {
			if !want[node.Name] {
				continue
			}
			ranges = append(ranges, Range{Start: node.StartLine, End: node.EndLine})
		}
	}
	sort.Slice(ranges, func(i, j int) bool { return ranges[i].Start < ranges[j].Start })
	return ranges, nil
}

func parseRange(spec string) (Range, error) {
	parts := strings.SplitN(strings.TrimSpace(spec), "-", 2)
	if len(parts) != 2 {
		return Range{}, fmt.Errorf("invalid line range %q (expected a-b)", spec)
	}
	start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return Range{}, fmt.Errorf("invalid line range %q", spec)
	}
	end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil || end < start {
		return Range{}, fmt.Errorf("invalid line range %q", spec)
	}
	return Range{Start: start, End: end}, nil
}

func readLines(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	return strings.Split(string(data), "\n")
}

func truncateChunk(text string, maxTokens int) string {
	marker := "\n// ... truncated"
	budget := maxTokens*4 - len(marker)
	if budget <= 0 {
		return marker
	}
	if len(text) > budget {
		text = text[:budget]
	}
	return strings.TrimRight(text, "\n") + marker
}

func sumRanges(ranges []Range) int {
	total := 0
	for _, r := range ranges {
		total += r.End - r.Start + 1
	}
	return total
}

func extendSavings(current, extra *codegraph.Savings) *codegraph.Savings {
	if extra != nil {
		return extra
	}
	return current
}

func FormatView(result *Result) string {
	if result.Skeleton != nil {
		return result.Skeleton.RenderMarkdown()
	}
	if result.Chunk != nil {
		var b strings.Builder
		fmt.Fprintf(&b, "=== chunk %d/%d %s ===\n", result.Chunk.ChunkIndex+1, 0, result.Chunk.ChunkID)
		fmt.Fprintf(&b, "lines %d-%d · %d tokens\n\n", result.Chunk.LineStart, result.Chunk.LineEnd, result.Chunk.Tokens)
		b.WriteString(result.Chunk.Content)
		if result.Chunk.NextChunk >= 0 {
			fmt.Fprintf(&b, "\n\nnext: %s\n", result.Chunk.SuggestedNext)
		}
		return b.String()
	}
	if result.Selection != nil {
		return result.Selection.Content
	}
	return FormatResult(result)
}
