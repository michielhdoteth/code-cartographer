package mcp

import (
	"testing"
)

func TestDiffSinceLastMapInput(t *testing.T) {
	input := DiffSinceLastMapInput{
		Path: "/test/path",
	}
	if input.Path != "/test/path" {
		t.Errorf("Expected path '/test/path', got '%s'", input.Path)
	}
}

func TestDiffSinceLastMapOutput(t *testing.T) {
	output := DiffSinceLastMapOutput{
		ChangedFiles: []string{"file1.go", "file2.go"},
		Count:        2,
		FromCommit:   "abc123",
		ToCommit:     "def456",
		Message:      "Changed 2 files since commit abc123",
	}
	if output.Count != 2 {
		t.Errorf("Expected count 2, got %d", output.Count)
	}
	if len(output.ChangedFiles) != 2 {
		t.Errorf("Expected 2 changed files, got %d", len(output.ChangedFiles))
	}
}

func TestGetChangedFilesInput(t *testing.T) {
	input := GetChangedFilesInput{
		Path: "/test/path",
	}
	if input.Path != "/test/path" {
		t.Errorf("Expected path '/test/path', got '%s'", input.Path)
	}
}

func TestGetChangedFilesOutput(t *testing.T) {
	output := GetChangedFilesOutput{
		Added:    []string{"new.go"},
		Modified: []string{"changed.go"},
		Deleted:  []string{"old.go"},
		Total:    3,
		Message:  "3 uncommitted changes",
	}
	if output.Total != 3 {
		t.Errorf("Expected total 3, got %d", output.Total)
	}
	if len(output.Added) != 1 {
		t.Errorf("Expected 1 added file, got %d", len(output.Added))
	}
	if len(output.Modified) != 1 {
		t.Errorf("Expected 1 modified file, got %d", len(output.Modified))
	}
	if len(output.Deleted) != 1 {
		t.Errorf("Expected 1 deleted file, got %d", len(output.Deleted))
	}
}

func TestImpactAnalysisInput(t *testing.T) {
	input := ImpactAnalysisInput{
		Path: "/test/path",
		File: "main.go",
	}
	if input.Path != "/test/path" {
		t.Errorf("Expected path '/test/path', got '%s'", input.Path)
	}
	if input.File != "main.go" {
		t.Errorf("Expected file 'main.go', got '%s'", input.File)
	}
}

func TestImpactAnalysisOutput(t *testing.T) {
	output := ImpactAnalysisOutput{
		Dependents:   []string{"file1.go", "file2.go"},
		Dependencies: []string{"utils.go"},
		Risk:         "medium",
		ReadSet:      []string{"main.go", "file1.go", "file2.go"},
		Message:      "File main.go has 2 dependents, risk: medium",
	}
	if output.Risk != "medium" {
		t.Errorf("Expected risk 'medium', got '%s'", output.Risk)
	}
	if len(output.Dependents) != 2 {
		t.Errorf("Expected 2 dependents, got %d", len(output.Dependents))
	}
	if len(output.ReadSet) != 3 {
		t.Errorf("Expected 3 read set items, got %d", len(output.ReadSet))
	}
}