package doctor

import (
	"os"
	"path/filepath"
	"testing"

	"carto/codegraph"
	"carto/skill"
)

func TestRunReportsFreshGraph(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\nfunc Main() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := skill.Install(root, "both"); err != nil {
		t.Fatal(err)
	}
	if _, err := codegraph.Build(root); err != nil {
		t.Fatal(err)
	}
	report := Run(root)
	if report.Status != "ok" && report.Status != "warn" {
		t.Fatalf("unexpected status: %+v", report)
	}
	foundFreshness := false
	for _, c := range report.Checks {
		if c.Name == "freshness" {
			foundFreshness = true
			if c.Status != "ok" {
				t.Fatalf("graph should be fresh: %+v", c)
			}
		}
	}
	if !foundFreshness {
		t.Fatal("missing freshness check")
	}
	if report.Files < 1 {
		t.Fatalf("expected indexed files, got %d", report.Files)
	}
}
