package skill

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//go:embed SKILL.md
var Text string

const (
	AgentsMarkerBegin = "<!-- carto-guide:begin -->"
	AgentsMarkerEnd   = "<!-- carto-guide:end -->"
)

func AgentsSection() string {
	return AgentsMarkerBegin + "\n" +
		"## carto — codebase map\n\n" +
		"Use `carto` for codebase navigation instead of reading whole trees. Run `carto skill` for the full guide.\n\n" +
		"1. `carto doctor . --json` — graph freshness and health.\n" +
		"2. `carto repo-map . --tokens 2048 --json` — orient cheaply.\n" +
		"3. `carto ask \"<question>\" . --json` — locate symbols.\n" +
		"4. `carto context <file> --imports --callers --cohesion --json` — assemble only needed context.\n" +
		"5. `carto analyze structure . --json` and `carto fix <finding-id>` — before and after structural edits.\n\n" +
		"Exit codes: 0 ok, 1 findings threshold, 2 usage, 3 stale graph (run `carto map`), 4 I/O/config.\n" +
		AgentsMarkerEnd + "\n"
}

func Install(root, target string) ([]string, error) {
	var written []string
	switch target {
	case "claude", "both":
		dir := filepath.Join(root, ".claude", "skills", "carto")
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create skill directory: %w", err)
		}
		path := filepath.Join(dir, "SKILL.md")
		if err := os.WriteFile(path, []byte(Text), 0644); err != nil {
			return nil, fmt.Errorf("write skill: %w", err)
		}
		written = append(written, path)
	}
	switch target {
	case "agents", "both":
		path := filepath.Join(root, "AGENTS.md")
		existing, _ := os.ReadFile(path)
		updated := upsertAgentsSection(string(existing))
		if err := os.WriteFile(path, []byte(updated), 0644); err != nil {
			return nil, fmt.Errorf("write AGENTS.md: %w", err)
		}
		written = append(written, path)
	}
	return written, nil
}

func upsertAgentsSection(existing string) string {
	section := AgentsSection()
	start := strings.Index(existing, AgentsMarkerBegin)
	end := strings.Index(existing, AgentsMarkerEnd)
	if start >= 0 && end > start {
		return existing[:start] + strings.TrimSpace(section) + existing[end+len(AgentsMarkerEnd):]
	}
	if strings.TrimSpace(existing) == "" {
		return section
	}
	return strings.TrimRight(existing, "\n") + "\n\n" + section
}

func Check(root, target string) []string {
	var problems []string
	switch target {
	case "claude", "both":
		path := filepath.Join(root, ".claude", "skills", "carto", "SKILL.md")
		data, err := os.ReadFile(path)
		if err != nil {
			problems = append(problems, "missing or unreadable: "+path)
		} else if string(data) != Text {
			problems = append(problems, "out of date: "+path)
		}
	}
	switch target {
	case "agents", "both":
		path := filepath.Join(root, "AGENTS.md")
		data, err := os.ReadFile(path)
		if err != nil {
			problems = append(problems, "missing or unreadable: "+path)
		} else if !strings.Contains(string(data), AgentsMarkerBegin) || !strings.Contains(string(data), AgentsMarkerEnd) {
			problems = append(problems, "missing carto guide section: "+path)
		}
	}
	return problems
}
