package fix

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"carto/analyze"
	"carto/codegraph"
)

func TestGeneratePromptIncludesEvidenceAndLoop(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\nfunc Main() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := codegraph.Build(root); err != nil {
		t.Fatal(err)
	}
	finding := analyze.StructureFinding{
		ID:       "arch/shallow-module:abc",
		Rule:     "arch/shallow-module",
		Severity: "warning",
		Module:   "utils",
		Message:  "shallow module: callers use 4 of 5 exports",
		Evidence: analyze.StructureEvidence{
			UsedExports: 4, TotalExports: 5, Symbols: []string{"Main"},
		},
		Files: []string{filepath.Join(root, "main.go")},
	}
	prompt, err := GeneratePrompt(root, finding, 4000)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"## Problem", "## Evidence", "arch/shallow-module", "## Task", "carto analyze structure", "## Constraints"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q:\n%s", want, prompt)
		}
	}
}

func TestGeneratePromptHonorsBudget(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\nfunc Main() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := codegraph.Build(root); err != nil {
		t.Fatal(err)
	}
	finding := analyze.StructureFinding{Rule: "arch/broad-interface", Severity: "info", Message: "broad interface"}
	prompt, err := GeneratePrompt(root, finding, 120)
	if err != nil {
		t.Fatal(err)
	}
	if codegraph.EstimateTokens(prompt) > 120 {
		t.Fatalf("prompt exceeds budget: %d tokens", codegraph.EstimateTokens(prompt))
	}
}
