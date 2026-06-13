package find

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// SearchOptions defines options for search
type SearchOptions struct {
	CaseSensitive bool
	UseRegex      bool
	ContextLines  int
	MaxResults    int
}

// SearchResult represents a single search match
type SearchResult struct {
	File     string
	Line     int
	Column   int
	Content  string
	Match    string
}

// Search searches for a pattern in files within a directory
func Search(rootPath string, pattern string, opts SearchOptions) ([]SearchResult, error) {
	var results []SearchResult
	var regex *regexp.Regexp

	// Compile regex if needed
	if opts.UseRegex {
		flags := ""
		if !opts.CaseSensitive {
			flags = "(?i)"
		}
		var err error
		regex, err = regexp.Compile(flags + pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid regex: %w", err)
		}
	}

	// Walk directory
	walkErr := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			name := info.Name()
			if name == "node_modules" || name == ".git" || name == "vendor" ||
				name == "dist" || name == "build" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if it's a source file
		ext := strings.ToLower(filepath.Ext(path))
		if !isSourceFile(ext) {
			return nil
		}

		// Read file
		content, err := ioutil.ReadFile(path)
		if err != nil {
			return nil
		}

		// Search in file
		fileResults := searchInContent(string(content), path, pattern, regex, opts)
		results = append(results, fileResults...)

		// Check max results
		if opts.MaxResults > 0 && len(results) >= opts.MaxResults {
			return filepath.SkipAll
		}

		return nil
	})

	if walkErr != nil {
		return nil, walkErr
	}

	return results, nil
}

// searchInContent searches for a pattern in file content
func searchInContent(content, filePath, pattern string, regex *regexp.Regexp, opts SearchOptions) []SearchResult {
	var results []SearchResult
	lines := strings.Split(content, "\n")

	for lineNum, line := range lines {
		var matches []int

		if regex != nil {
			// Regex search
			loc := regex.FindStringIndex(line)
			if loc != nil {
				matches = []int{loc[0], loc[1]}
			}
		} else {
			// Literal search
			idx := strings.Index(line, pattern)
			if idx == -1 {
				continue
			}
			if !opts.CaseSensitive {
				lowerLine := strings.ToLower(line)
				lowerPattern := strings.ToLower(pattern)
				idx = strings.Index(lowerLine, lowerPattern)
			}
			if idx >= 0 {
				matches = []int{idx, idx + len(pattern)}
			}
		}

		for _, match := range matches {
			results = append(results, SearchResult{
				File:    filePath,
				Line:    lineNum + 1,
				Column:  match + 1,
				Content: strings.TrimSpace(line),
				Match:   line[match:min(match+len(pattern), len(line))],
			})
		}
	}

	return results
}

// isSourceFile checks if a file extension is a source file
func isSourceFile(ext string) bool {
	sourceExts := []string{".go", ".ts", ".tsx", ".js", ".jsx", ".py", ".rs", ".java", ".cpp", ".c", ".h", ".hpp", ".rb", ".php", ".cs", ".swift", ".kt", ".scala"}

	for _, e := range sourceExts {
		if ext == e {
			return true
		}
	}
	return false
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// FormatResults formats search results for display
func FormatResults(results []SearchResult, maxResults int) string {
	if len(results) == 0 {
		return "No matches found.\n"
	}

	if maxResults > 0 && len(results) > maxResults {
		results = results[:maxResults]
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d matches:\n\n", len(results)))

	for _, r := range results {
		sb.WriteString(fmt.Sprintf("%s:%d:%d\n", r.File, r.Line, r.Column))
		sb.WriteString(fmt.Sprintf("  %s\n\n", r.Content))
	}

	if maxResults > 0 && len(results) >= maxResults {
		sb.WriteString(fmt.Sprintf("... and %d more matches", len(results)-maxResults))
	}

	return sb.String()
}