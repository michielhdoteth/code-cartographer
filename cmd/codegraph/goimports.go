package codegraph

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type goListPackage struct {
	ImportPath string   `json:"ImportPath"`
	Dir        string   `json:"Dir"`
	GoFiles    []string `json:"GoFiles"`
	CgoFiles   []string `json:"CgoFiles"`
}

// loadGoImportMap maps Go import paths to repository-relative file paths using
// the local toolchain. GOPROXY=off keeps resolution offline; any failure leaves
// the heuristic resolver in place.
func loadGoImportMap(root string, files []string) map[string]string {
	result := make(map[string]string)
	known := make(map[string]bool, len(files))
	for _, file := range files {
		known[file] = true
	}
	if len(known) == 0 {
		return result
	}

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
		return result
	}

	decoder := json.NewDecoder(strings.NewReader(string(out)))
	for decoder.More() {
		var pkg goListPackage
		if err := decoder.Decode(&pkg); err != nil {
			continue
		}
		if pkg.ImportPath == "" || pkg.Dir == "" {
			continue
		}
		names := append(append([]string{}, pkg.GoFiles...), pkg.CgoFiles...)
		sort.Strings(names)
		for _, name := range names {
			abs := filepath.Join(pkg.Dir, name)
			rel, err := filepath.Rel(root, abs)
			if err != nil {
				continue
			}
			rel = filepath.ToSlash(rel)
			if !known[rel] {
				continue
			}
			if _, exists := result[pkg.ImportPath]; !exists {
				result[pkg.ImportPath] = rel
			}
		}
	}
	return result
}
