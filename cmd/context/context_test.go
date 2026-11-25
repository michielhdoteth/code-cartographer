package context

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"carto/codegraph"
)

func makeFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		"util.go": "package main\nfunc Helper() string { return \"x\" }\n",
		"main.go": "package main\nfunc Main() string { return Helper() }\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := codegraph.Build(root); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestGenerateReportsSavings(t *testing.T) {
	root := makeFixture(t)
	result, err := Generate(Options{FilePath: filepath.Join(root, "main.go"), TokenBudget: 8000, RootPath: root})
	if err != nil {
		t.Fatal(err)
	}
	if result.Savings == nil {
		t.Fatal("expected context savings")
	}
	if result.Savings.WholeRepoTokens <= 0 || result.Savings.UsedTokens <= 0 {
		t.Fatalf("unexpected savings: %+v", result.Savings)
	}
	if result.Savings.SavedTokens < 0 || result.Savings.SavedPercent > 100 {
		t.Fatalf("invalid savings: %+v", result.Savings)
	}
	if result.Savings.Basis == "" {
		t.Fatal("savings must state its estimation basis")
	}
}

func TestCohesionDetectsScatteredContext(t *testing.T) {
	root := t.TempDir()
	for dir, content := range map[string]string{
		"a": "package a\nimport \"b\"\nfunc A() string { return b.B() }\n",
		"b": "package b\nfunc B() string { return \"b\" }\n",
	} {
		path := filepath.Join(root, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(path, dir+".go"), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := codegraph.Build(root); err != nil {
		t.Fatal(err)
	}
	result, err := Generate(Options{
		FilePath:        filepath.Join(root, "a", "a.go"),
		TokenBudget:     8000,
		RootPath:        root,
		IncludeCohesion: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Cohesion == nil {
		t.Fatal("expected cohesion info")
	}
	if result.Cohesion.Score >= 100 {
		t.Fatalf("expected a scattered score below 100, got %d: %+v", result.Cohesion.Score, result.Cohesion)
	}
	if len(result.Cohesion.TopDirs) < 2 {
		t.Fatalf("expected two top-level areas, got %v", result.Cohesion.TopDirs)
	}
}

func TestGenerateCohesion(t *testing.T) {
	root := makeFixture(t)
	result, err := Generate(Options{
		FilePath:        filepath.Join(root, "main.go"),
		TokenBudget:     8000,
		RootPath:        root,
		IncludeCohesion: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Cohesion == nil {
		t.Fatal("expected cohesion info when requested")
	}
	if result.Cohesion.Score < 0 || result.Cohesion.Score > 100 {
		t.Fatalf("invalid cohesion score: %d", result.Cohesion.Score)
	}
	if result.Cohesion.Verdict == "" {
		t.Fatal("expected cohesion verdict")
	}
	out := FormatResult(result)
	if !strings.Contains(out, "Structural cohesion") || !strings.Contains(out, "Context savings") {
		t.Fatalf("formatted result missing new fields: %s", out)
	}
}
