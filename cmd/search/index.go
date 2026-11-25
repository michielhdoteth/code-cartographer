package search

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"carto/retrieval"
)

type Index struct {
	RootPath    string              `json:"root_path"`
	BuiltAt     time.Time           `json:"built_at"`
	Terms       map[string][]Entry  `json:"terms"`       // word -> entries
	Files       map[string]FileInfo `json:"files"`       // path -> file info
	TotalFiles  int                 `json:"total_files"`
	TotalTerms  int                 `json:"total_terms"`
}

type Entry struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Context string `json:"context"` // surrounding line
}

type FileInfo struct {
	Path     string `json:"path"`
	Lines    int    `json:"lines"`
	Size     int64  `json:"size"`
	ModTime  time.Time `json:"mod_time"`
}

type SearchResult struct {
	File    string  `json:"file"`
	Line    int     `json:"line"`
	Context string  `json:"context"`
	Score   float64 `json:"score"`
}

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

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") {
			continue
		}

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

func extractWords(line string) []string {
	return retrieval.Tokenize(line)
}

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

func (idx *Index) Search(query string, maxResults int) []SearchResult {
	return idx.rank(query, maxResults)
}

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
