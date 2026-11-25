package codegraph

import (
	"fmt"
	"sort"
	"strings"
)

const (
	ChunkerVersion   = 1
	ChunkTargetTokens = 300
	ChunkMinTokens    = 120
	ChunkMaxTokens    = 400
)

type ChunkManifest struct {
	SchemaVersion  int          `json:"schema_version"`
	FilePath       string       `json:"file_path"`
	Level          int          `json:"level"`
	FileHash       string       `json:"file_hash"`
	ChunkerVersion int          `json:"chunker_version"`
	TokenBudget    int          `json:"budget_tokens"`
	TotalLines     int          `json:"total_lines"`
	TotalChunks    int          `json:"total_chunks"`
	Chunks         []ChunkMeta  `json:"chunks"`
	SuggestedNext  string       `json:"suggested_next"`
	Savings        *Savings     `json:"savings,omitempty"`
}

type ChunkMeta struct {
	ChunkIndex   int      `json:"chunk_index"`
	ChunkID      string   `json:"chunk_id"`
	LineStart    int      `json:"line_start"`
	LineEnd      int      `json:"line_end"`
	Lines        int      `json:"lines"`
	Tokens       int      `json:"tokens"`
	Symbols      []string `json:"symbols"`
	SplitSymbol  bool     `json:"split_symbol,omitempty"`
	ApproxBounds bool     `json:"approx_bounds,omitempty"`
}

type Chunk struct {
	ChunkMeta
	SchemaVersion string   `json:"schema_version"`
	Content       string   `json:"content"`
	NextChunk     int      `json:"next_chunk"`
	Truncated     bool     `json:"truncated"`
	SuggestedNext string   `json:"suggested_next"`
	Savings       *Savings `json:"savings,omitempty"`
}

type chunkSegment struct {
	start, end int
	symbols    []string
	split      bool
	approx     bool
}

func (index *Index) BuildChunkManifest(path string, budget int) (ChunkManifest, error) {
	rel := normalizeRelativePath(index.RootPath, path)
	manifest := ChunkManifest{
		SchemaVersion: 1, FilePath: rel, Level: 2, ChunkerVersion: ChunkerVersion,
		TokenBudget: budget, Chunks: []ChunkMeta{},
	}
	file, ok := index.Files[rel]
	if !ok {
		return manifest, fmt.Errorf("file not in graph: %s", rel)
	}
	lines := readFileLines(index.RootPath, rel)
	if len(lines) == 0 {
		return manifest, fmt.Errorf("file is empty or unreadable: %s", rel)
	}
	manifest.TotalLines = len(lines)
	manifest.FileHash = shortHash(file.Hash)
	if budget <= 0 {
		budget = ChunkTargetTokens
	}
	if budget > ChunkMaxTokens {
		budget = ChunkMaxTokens
	}

	segments := buildChunkSegments(index, rel, lines)
	metas := packChunkSegments(segments, lines, budget)
	prefix := chunkIDPrefix(file.Hash)
	for i := range metas {
		metas[i].ChunkIndex = i
		metas[i].ChunkID = fmt.Sprintf("%s:v%d:%d", prefix, ChunkerVersion, i)
	}
	manifest.Chunks = metas
	manifest.TotalChunks = len(metas)
	if manifest.TotalChunks > 0 {
		manifest.SuggestedNext = fmt.Sprintf("carto chunk %s --chunk 0 --json", rel)
	}
	manifest.Savings = BuildSavings(0, 0, int(file.Size/4), file.Lines, "file")
	return manifest, nil
}

func (index *Index) GenerateChunk(path string, chunkIndex, budget, maxChunks int) (Chunk, error) {
	manifest, err := index.BuildChunkManifest(path, budget)
	if err != nil {
		return Chunk{}, err
	}
	if chunkIndex < 0 || chunkIndex >= manifest.TotalChunks {
		return Chunk{}, fmt.Errorf("chunk index %d out of range (0..%d)", chunkIndex, manifest.TotalChunks-1)
	}
	if maxChunks > 0 && manifest.TotalChunks > maxChunks {
		return Chunk{}, fmt.Errorf("file has %d chunks, exceeding --max-chunks %d", manifest.TotalChunks, maxChunks)
	}
	meta := manifest.Chunks[chunkIndex]
	lines := readFileLines(index.RootPath, manifest.FilePath)
	content := strings.Join(symbolLines(lines, meta.LineStart, meta.LineEnd), "\n")
	next := chunkIndex + 1
	if next >= manifest.TotalChunks {
		next = -1
	}
	suggested := ""
	if next >= 0 {
		suggested = fmt.Sprintf("carto chunk %s --chunk %d --json", manifest.FilePath, next)
	}
	return Chunk{
		ChunkMeta: meta, SchemaVersion: "1", Content: content, NextChunk: next,
		SuggestedNext: suggested,
		Savings: BuildSavings(meta.Tokens, meta.Lines, int(linesByteEstimate(lines)/4), len(lines), "file"),
	}, nil
}

func linesByteEstimate(lines []string) int64 {
	total := int64(0)
	for _, line := range lines {
		total += int64(len(line)) + 1
	}
	return total
}

func chunkIDPrefix(hash string) string {
	if len(hash) < 6 {
		return hash
	}
	return hash[:6]
}

func buildChunkSegments(index *Index, rel string, lines []string) []chunkSegment {
	var nodes []Node
	for _, node := range index.Nodes {
		if node.File == rel && node.Kind != "file" {
			nodes = append(nodes, node)
		}
	}
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].StartLine != nodes[j].StartLine {
			return nodes[i].StartLine < nodes[j].StartLine
		}
		return nodes[i].EndLine < nodes[j].EndLine
	})
	var segments []chunkSegment
	previous := 1
	for _, node := range nodes {
		start := node.StartLine
		if start < 1 {
			start = 1
		}
		if start > len(lines) {
			continue
		}
		if start > previous {
			segments = append(segments, chunkSegment{start: previous, end: start - 1})
		}
		end := node.EndLine
		if end < start {
			end = start
		}
		if end > len(lines) {
			end = len(lines)
		}
		segment := chunkSegment{
			start: start, end: end, symbols: []string{node.Name},
			approx: node.RangeConfidence == "estimated",
		}
		segments = append(segments, splitOversizedSegment(segment, lines)...)
		previous = end + 1
	}
	if previous <= len(lines) {
		segments = append(segments, chunkSegment{start: previous, end: len(lines)})
	}
	return segments
}

func splitOversizedSegment(segment chunkSegment, lines []string) []chunkSegment {
	tokens := EstimateTokens(strings.Join(symbolLines(lines, segment.start, segment.end), "\n"))
	if tokens <= ChunkMaxTokens {
		return []chunkSegment{segment}
	}
	span := segment.end - segment.start + 1
	parts := (tokens + ChunkMaxTokens - 1) / ChunkMaxTokens
	if parts < 2 {
		parts = 2
	}
	per := (span + parts - 1) / parts
	var out []chunkSegment
	for start := segment.start; start <= segment.end; start += per {
		end := start + per - 1
		if end > segment.end {
			end = segment.end
		}
		out = append(out, chunkSegment{
			start: start, end: end, symbols: segment.symbols,
			split: true, approx: segment.approx,
		})
	}
	return out
}

func packChunkSegments(segments []chunkSegment, lines []string, budget int) []ChunkMeta {
	var metas []ChunkMeta
	current := chunkSegment{}
	hasCurrent := false
	flush := func() {
		if !hasCurrent || current.end < current.start {
			return
		}
		metas = append(metas, ChunkMeta{
			LineStart: current.start, LineEnd: current.end,
			Lines: current.end - current.start + 1,
			Tokens: EstimateTokens(strings.Join(symbolLines(lines, current.start, current.end), "\n")),
			Symbols: append([]string{}, current.symbols...),
			SplitSymbol: current.split, ApproxBounds: current.approx,
		})
		current = chunkSegment{}
		hasCurrent = false
	}
	for _, segment := range segments {
		if !hasCurrent {
			current = segment
			hasCurrent = true
			continue
		}
		if segment.start != current.end+1 {
			flush()
			current = segment
			hasCurrent = true
			continue
		}
		candidate := chunkSegment{start: current.start, end: segment.end}
		candidate.symbols = append(append([]string{}, current.symbols...), segment.symbols...)
		candidate.split = current.split || segment.split
		candidate.approx = current.approx || segment.approx
		candidateTokens := EstimateTokens(strings.Join(symbolLines(lines, candidate.start, candidate.end), "\n"))
		if candidateTokens > budget {
			flush()
			current = segment
			hasCurrent = true
			continue
		}
		current = candidate
	}
	flush()

	for i := 0; i < len(metas)-1; {
		if metas[i].Tokens >= ChunkMinTokens {
			i++
			continue
		}
		mergedEnd := metas[i+1].EndLine()
		mergedTokens := EstimateTokens(strings.Join(symbolLines(lines, metas[i].LineStart, mergedEnd), "\n"))
		if mergedTokens > ChunkMaxTokens {
			i++
			continue
		}
		metas[i] = ChunkMeta{
			LineStart: metas[i].LineStart, LineEnd: mergedEnd,
			Lines: mergedEnd - metas[i].LineStart + 1, Tokens: mergedTokens,
			Symbols: append(append([]string{}, metas[i].Symbols...), metas[i+1].Symbols...),
			SplitSymbol: metas[i].SplitSymbol || metas[i+1].SplitSymbol,
			ApproxBounds: metas[i].ApproxBounds || metas[i+1].ApproxBounds,
		}
		metas = append(metas[:i+1], metas[i+2:]...)
	}
	return metas
}

func (m ChunkMeta) EndLine() int { return m.LineEnd }

