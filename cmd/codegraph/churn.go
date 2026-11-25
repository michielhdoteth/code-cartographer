package codegraph

import (
	"os/exec"
	"path/filepath"
	"strings"
)

// gitChurn counts recent file mentions in git history. It is a local-only
// signal; non-repositories return an empty map.
func gitChurn(root string) map[string]int {
	cmd := exec.Command("git", "log", "--format=", "--name-only", "-n", "200")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return map[string]int{}
	}
	result := make(map[string]int)
	for _, line := range strings.Split(string(out), "\n") {
		path := filepath.ToSlash(strings.TrimSpace(line))
		if path == "" {
			continue
		}
		result[path]++
	}
	return result
}
