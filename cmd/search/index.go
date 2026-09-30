package search

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Index represents a search index for a codebase
type Index struct {
	RootPath    string              `json:"root_path"`
	BuiltAt     time.Time           `json:"built_at"`
	Terms       map[string][]Entry  `json:"terms"`       // word -> entries
	Files       map[string]FileInfo `json:"files"`       // path -> file info
	TotalFiles  int                 `json:"total_files"`
	TotalTerms  int                 `json:"total_terms"`
}

// Entry is a single occurrence of a term in a file
type Entry struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Context string `json:"context"` // surrounding line
}

// FileInfo stores metadata about a indexed file
type FileInfo struct {
	Path     string `json:"path"`
	Lines    int    `json:"lines"`
	Size     int64  `json:"size"`
	ModTime  time.Time `json:"mod_time"`
}

// SearchResult is a single search result
type SearchResult struct {
	File    string  `json:"file"`
	Line    int     `json:"line"`
	Context string  `json:"context"`
	Score   float64 `json:"score"`
}

// BuildIndex creates a search index for the given directory
func BuildIndex(rootPath string) (*Index, error) {
	absPath, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, err
	}

	index := &Index{
		RootPath: absPath,
		BuiltAt:  time.Now(),
		Terms:    make(map[string][]Entry),
		Files:    make(map[string]FileInfo),
	}

	err = filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
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

		// Only index source files
		ext := filepath.Ext(path)
		if !isSourceFile(ext) {
			return nil
		}

		relPath, _ := filepath.Rel(absPath, path)
		indexFile(path, relPath, index)
		index.TotalFiles++

		return nil
	})

	if err != nil {
		return nil, err
	}

	index.TotalTerms = len(index.Terms)
	return index, nil
}

// indexFile adds a file's content to the index
func indexFile(absPath, relPath string, index *Index) {
	code, err := os.ReadFile(absPath)
	if err != nil {
		return
	}

	info, _ := os.Stat(absPath)
	lines := strings.Split(string(code), "\n")

	fileInfo := FileInfo{
		Path:    relPath,
		Lines:   len(lines),
		Size:    info.Size(),
		ModTime: info.ModTime(),
	}
	index.Files[relPath] = fileInfo

	// Index each line
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Extract words from the line
		words := extractWords(line)
		for _, word := range words {
			if len(word) < 2 {
				continue
			}
			entry := Entry{
				File:    relPath,
				Line:    i + 1,
				Context: trimmed,
			}
			index.Terms[word] = append(index.Terms[word], entry)
		}
	}
}

// extractWords splits a line into indexable words
func extractWords(line string) []string {
	// Split on non-alphanumeric characters
	words := strings.FieldsFunc(line, func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_')
	})
	return words
}

// isSourceFile checks if a file extension is a source file
func isSourceFile(ext string) bool {
	sourceExts := map[string]bool{
		".go": true, ".js": true, ".ts": true, ".tsx": true, ".jsx": true,
		".py": true, ".java": true, ".c": true, ".cpp": true, ".h": true,
		".cs": true, ".rb": true, ".php": true, ".rs": true, ".swift": true,
		".kt": true, ".scala": true, ".sh": true, ".bash": true,
		".html": true, ".css": true, ".sql": true, ".yaml": true, ".yml": true,
		".json": true, ".toml": true, ".md": true, ".txt": true,
	}
	return sourceExts[ext]
}

// Search searches the index for a query
func (idx *Index) Search(query string, maxResults int) []SearchResult {
	query = strings.ToLower(query)
	queryWords := strings.Fields(query)

	if len(queryWords) == 0 {
		return nil
	}

	// Score each file based on matching terms
	scores := make(map[string]float64)
	bestLines := make(map[string]int)
	bestContexts := make(map[string]string)

	for _, word := range queryWords {
		for term, entries := range idx.Terms {
			if strings.Contains(strings.ToLower(term), word) {
				for _, entry := range entries {
					scores[entry.File] += 1.0
					if bestLines[entry.File] == 0 || entry.Line < bestLines[entry.File] {
						bestLines[entry.File] = entry.Line
						bestContexts[entry.File] = entry.Context
					}
				}
			}
		}
	}

	// Sort by score
	type scored struct {
		file  string
		score float64
	}
	var sorted []scored
	for file, score := range scores {
		sorted = append(sorted, scored{file, score})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].score > sorted[j].score
	})

	// Build results
	var results []SearchResult
	for i, s := range sorted {
		if i >= maxResults {
			break
		}
		results = append(results, SearchResult{
			File:    s.file,
			Line:    bestLines[s.file],
			Context: bestContexts[s.file],
			Score:   s.score,
		})
	}

	return results
}

// Save saves the index to a file
func (idx *Index) Save(rootPath string) error {
	indexPathDir := filepath.Join(rootPath, ".carto")
	os.MkdirAll(indexPathDir, 0755)
	indexPathFile := filepath.Join(indexPathDir, "search-index.json")

	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(indexPathFile, data, 0644)
}

// LoadIndex loads an index from disk
func LoadIndex(rootPath string) (*Index, error) {
	indexPathFile := filepath.Join(rootPath, ".carto", "search-index.json")
	data, err := os.ReadFile(indexPathFile)
	if err != nil {
		return nil, err
	}
	var index Index
	err = json.Unmarshal(data, &index)
	if err != nil {
		return nil, err
	}
	return &index, nil
}

// FormatResults formats search results for display
func FormatResults(results []SearchResult, query string) string {
	if len(results) == 0 {
		return fmt.Sprintf("No results for \"%s\"", query)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d results for \"%s\":\n\n", len(results), query))

	for _, r := range results {
		sb.WriteString(fmt.Sprintf("  %s:%d  (score: %.1f)\n", r.File, r.Line, r.Score))
		sb.WriteString(fmt.Sprintf("    %s\n\n", r.Context))
	}

	return sb.String()
}
