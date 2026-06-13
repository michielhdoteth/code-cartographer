package manifest

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()

	m := &Manifest{
		Repo:      "test-repo",
		Commit:    "abc123def456",
		Branch:    "main",
		MappedAt:  time.Now(),
		FileCount: 42,
		Languages: map[string]int{"go": 10, "typescript": 32},
	}

	if err := Save(dir, m); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists
	if !Exists(dir) {
		t.Fatal("manifest should exist after save")
	}

	// Load and compare
	loaded, err := Load(dir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Repo != m.Repo {
		t.Errorf("repo: got %s, want %s", loaded.Repo, m.Repo)
	}
	if loaded.Commit != m.Commit {
		t.Errorf("commit: got %s, want %s", loaded.Commit, m.Commit)
	}
	if loaded.Branch != m.Branch {
		t.Errorf("branch: got %s, want %s", loaded.Branch, m.Branch)
	}
	if loaded.FileCount != m.FileCount {
		t.Errorf("file_count: got %d, want %d", loaded.FileCount, m.FileCount)
	}
	if loaded.Languages["go"] != 10 {
		t.Errorf("languages[go]: got %d, want 10", loaded.Languages["go"])
	}
}

func TestExists(t *testing.T) {
	dir := t.TempDir()

	if Exists(dir) {
		t.Error("manifest should not exist in empty dir")
	}

	// Create .carto/manifest.json
	os.MkdirAll(filepath.Join(dir, ".carto"), 0755)
	os.WriteFile(filepath.Join(dir, ".carto", "manifest.json"), []byte(`{}`), 0644)

	if !Exists(dir) {
		t.Error("manifest should exist after file creation")
	}
}

func TestLoadNotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := Load(dir)
	if err == nil {
		t.Error("expected error when loading non-existent manifest")
	}
}
