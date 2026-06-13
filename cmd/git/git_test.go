package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func setupTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Init repo
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}

	run("init")
	run("config", "user.email", "test@test.com")
	run("config", "user.name", "Test")

	// Create initial commit
	os.WriteFile(filepath.Join(dir, "test.txt"), []byte("hello"), 0644)
	run("add", ".")
	run("commit", "-m", "initial")

	return dir
}

func TestGetHeadSHA(t *testing.T) {
	dir := setupTestRepo(t)
	sha, err := GetHeadSHA(dir)
	if err != nil {
		t.Fatalf("GetHeadSHA failed: %v", err)
	}
	if len(sha) != 40 {
		t.Errorf("expected 40-char SHA, got %d chars: %s", len(sha), sha)
	}
}

func TestGetBranch(t *testing.T) {
	dir := setupTestRepo(t)
	branch, err := GetBranch(dir)
	if err != nil {
		t.Fatalf("GetBranch failed: %v", err)
	}
	if branch != "main" && branch != "master" {
		t.Errorf("expected main or master, got %s", branch)
	}
}

func TestIsRepo(t *testing.T) {
	dir := setupTestRepo(t)
	if !IsRepo(dir) {
		t.Error("expected IsRepo to return true for git repo")
	}

	tmpDir := t.TempDir()
	if IsRepo(tmpDir) {
		t.Error("expected IsRepo to return false for non-git dir")
	}
}

func TestGetStatus(t *testing.T) {
	dir := setupTestRepo(t)

	// Clean repo
	statuses, err := GetStatus(dir)
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}
	if len(statuses) != 0 {
		t.Errorf("expected 0 changes, got %d", len(statuses))
	}

	// Add a file
	os.WriteFile(filepath.Join(dir, "new.txt"), []byte("world"), 0644)
	statuses, err = GetStatus(dir)
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}
	if len(statuses) != 1 {
		t.Errorf("expected 1 change, got %d", len(statuses))
	}
}

func TestGetLastCommitSHA(t *testing.T) {
	dir := setupTestRepo(t)

	// Make a second commit
	os.WriteFile(filepath.Join(dir, "test2.txt"), []byte("world"), 0644)
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = dir
	cmd.Run()
	cmd = exec.Command("git", "commit", "-m", "second")
	cmd.Dir = dir
	cmd.Run()

	sha, err := GetLastCommitSHA(dir)
	if err != nil {
		t.Fatalf("GetLastCommitSHA failed: %v", err)
	}
	if len(sha) != 40 {
		t.Errorf("expected 40-char SHA, got %d chars", len(sha))
	}
}

func TestGetChangedFiles(t *testing.T) {
	dir := setupTestRepo(t)

	// Get initial SHA
	sha1, _ := GetHeadSHA(dir)

	// Make a change
	os.WriteFile(filepath.Join(dir, "new.txt"), []byte("world"), 0644)
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = dir
	cmd.Run()
	cmd = exec.Command("git", "commit", "-m", "add new")
	cmd.Dir = dir
	cmd.Run()

	sha2, _ := GetHeadSHA(dir)

	files, err := GetChangedFiles(dir, sha1, sha2)
	if err != nil {
		t.Fatalf("GetChangedFiles failed: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 changed file, got %d", len(files))
	}
}
