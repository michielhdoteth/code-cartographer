package analyze

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type goPackage struct {
	ImportPath string   `json:"ImportPath"`
	Dir        string   `json:"Dir"`
	GoFiles    []string `json:"GoFiles"`
	CgoFiles   []string `json:"CgoFiles"`
}

// resolveGoImportSources replaces heuristic import paths with repository-relative
// file paths using the local Go toolchain. Offline only; failures are ignored.
func resolveGoImportSources(root string, graph *ModuleGraph) {
	toolDir := root
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		if _, err := os.Stat(filepath.Join(root, "cmd", "go.mod")); err == nil {
			toolDir = filepath.Join(root, "cmd")
		}
	}
	cmd := exec.Command("go", "list", "-json", "./...")
	cmd.Dir = toolDir
	cmd.Env = append(cmd.Environ(), "GOPROXY=off")
	out, err := cmd.Output()
	if err != nil {
		return
	}
	bySlash := make(map[string]string, len(graph.Files))
	for key := range graph.Files {
		bySlash[filepath.ToSlash(key)] = key
	}
	imports := make(map[string]string)
	decoder := json.NewDecoder(strings.NewReader(string(out)))
	for decoder.More() {
		var pkg goPackage
		if err := decoder.Decode(&pkg); err != nil || pkg.ImportPath == "" || pkg.Dir == "" {
			continue
		}
		names := append(append([]string{}, pkg.GoFiles...), pkg.CgoFiles...)
		sort.Strings(names)
		for _, name := range names {
			rel, err := filepath.Rel(root, filepath.Join(pkg.Dir, name))
			if err != nil {
				continue
			}
			rel = filepath.ToSlash(rel)
			if actual, ok := bySlash[rel]; ok {
				if _, exists := imports[pkg.ImportPath]; !exists {
					imports[pkg.ImportPath] = actual
				}
			}
		}
	}
	for path, node := range graph.Files {
		if !strings.HasSuffix(path, ".go") {
			continue
		}
		for i := range node.Imports {
			if target, ok := imports[node.Imports[i].Source]; ok {
				node.Imports[i].Source = target
			}
		}
	}
}
