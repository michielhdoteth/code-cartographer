package context

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"carto/codegraph"
)

type CohesionInfo struct {
	Score   int      `json:"score"`
	TopDirs []string `json:"top_dirs"`
	Modules []string `json:"modules"`
	Files   int      `json:"files"`
	Verdict string   `json:"verdict"`
}

// calculateCohesion scores how close the related context is to the target file.
// A feature spread across many top-level directories costs agents more tokens
// because each additional directory needs another fetch. The score is a
// heuristic signal for refactoring, not an exact token measurement.
func calculateCohesion(index *codegraph.Index, filePath string) *CohesionInfo {
	if index == nil {
		return nil
	}
	rel, err := filepath.Rel(index.RootPath, filePath)
	if err != nil {
		return nil
	}
	rel = filepath.ToSlash(rel)

	files := make(map[string]bool)
	for _, node := range index.Outline(rel) {
		for _, hit := range index.TraceDepth(node.Name, 1, "both") {
			if hit.Node.File != "" && hit.Node.File != rel {
				files[hit.Node.File] = true
			}
		}
	}
	// Import-only neighbours cover languages with unresolved call edges.
	for path, file := range index.Files {
		if path == rel {
			continue
		}
		for _, imp := range file.Imports {
			if importResolvesTo(imp.Source, rel) {
				files[path] = true
				break
			}
		}
	}

	topDirs := make(map[string]bool)
	if parts := strings.Split(rel, "/"); len(parts) > 0 {
		topDirs[parts[0]] = true
	}
	mods := map[string]bool{filepath.ToSlash(filepath.Dir(rel)): true}
	for file := range files {
		parts := strings.Split(file, "/")
		if len(parts) > 0 {
			topDirs[parts[0]] = true
		}
	}
	for file := range files {
		mods[filepath.ToSlash(filepath.Dir(file))] = true
	}

	dirList := sortedStrings(topDirs)
	modList := sortedStrings(mods)
	score := 100
	if len(dirList) > 1 {
		score -= (len(dirList) - 1) * 15
		if score < 0 {
			score = 0
		}
	}
	verdict := "cohesive: related context lives close to the target"
	if len(dirList) > 1 {
		verdict = fmt.Sprintf("scattered: related context spans %d top-level areas (%s)", len(dirList), strings.Join(dirList, ", "))
	}

	return &CohesionInfo{
		Score:   score,
		TopDirs: dirList,
		Modules: modList,
		Files:   len(files),
		Verdict: verdict,
	}
}

// importResolvesTo mirrors the suffix-based import resolution used elsewhere
// in carto, bounded by a path-segment boundary to avoid false positives.
func importResolvesTo(source, target string) bool {
	clean := filepath.ToSlash(source)
	if clean == target {
		return true
	}
	candidates := []string{target, strings.TrimSuffix(target, filepath.Ext(target))}
	for _, candidate := range candidates {
		if len(clean) == 0 || len(candidate) < len(clean) {
			continue
		}
		if strings.HasSuffix(candidate, clean) {
			prefix := candidate[:len(candidate)-len(clean)]
			if prefix == "" || strings.HasSuffix(prefix, "/") {
				return true
			}
		}
	}
	return false
}

func sortedStrings(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
