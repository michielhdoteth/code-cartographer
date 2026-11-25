package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissingReturnsDefaults(t *testing.T) {
	cfg, err := Load(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Version != 1 || cfg.Defaults.ContextTokens != 8000 || cfg.Structure.MaxFindings <= 0 || cfg.Retrieval.K1 != 1.2 {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
}

func TestSaveLoadAndValidate(t *testing.T) {
	root := t.TempDir()
	cfg := Default()
	cfg.Defaults.ContextTokens = 12345
	if err := Save(root, cfg); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Defaults.ContextTokens != 12345 {
		t.Fatalf("token budget not persisted: %+v", loaded.Defaults)
	}
	if problems := Validate(root); len(problems) != 0 {
		t.Fatalf("unexpected validation problems: %v", problems)
	}

	if err := os.WriteFile(filepath.Join(root, ".carto", "config.json"), []byte("{bad"), 0644); err != nil {
		t.Fatal(err)
	}
	if problems := Validate(root); len(problems) == 0 {
		t.Fatal("expected validation problem for malformed config")
	}
}
