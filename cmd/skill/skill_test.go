package skill

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTextHasFrontmatterAndCommands(t *testing.T) {
	if !strings.HasPrefix(Text, "---\nname: carto") {
		t.Fatalf("skill is missing frontmatter")
	}
	for _, want := range []string{"carto doctor", "carto repo-map", "carto analyze structure", "carto fix", "Exit codes"} {
		if !strings.Contains(Text, want) {
			t.Fatalf("skill missing %q", want)
		}
	}
}

func TestInstallAndCheck(t *testing.T) {
	root := t.TempDir()
	files, err := Install(root, "both")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %v", files)
	}
	if problems := Check(root, "both"); len(problems) != 0 {
		t.Fatalf("fresh install reported drift: %v", problems)
	}
	path := filepath.Join(root, ".claude", "skills", "carto", "SKILL.md")
	if err := os.WriteFile(path, []byte("stale"), 0644); err != nil {
		t.Fatal(err)
	}
	if problems := Check(root, "claude"); len(problems) == 0 {
		t.Fatal("expected drift after modifying the skill copy")
	}

	// Re-install must not duplicate the AGENTS.md section.
	if _, err := Install(root, "agents"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(data), AgentsMarkerBegin) != 1 {
		t.Fatalf("AGENTS.md section duplicated: %s", data)
	}
}
