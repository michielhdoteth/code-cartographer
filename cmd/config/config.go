package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"carto/analyze"
	"carto/retrieval"
)

// Config is the optional per-repository carto configuration. It lives at
// .carto/config.json so it is ignored by git by default but local to the repo.
type Config struct {
	Version   int                       `json:"version"`
	Defaults  Defaults                  `json:"defaults"`
	Structure analyze.StructureConfig   `json:"structure"`
	Retrieval retrieval.Settings        `json:"retrieval,omitempty"`
}

// Defaults are token budgets used when a CLI flag is not provided.
type Defaults struct {
	ContextTokens  int      `json:"context_tokens"`
	RepoMapTokens  int      `json:"repo_map_tokens"`
	AskTokens      int      `json:"ask_tokens"`
	PackTokens     int `json:"pack_tokens"`
}

func Path(root string) string {
	return filepath.Join(root, ".carto", "config.json")
}

func Default() *Config {
	return &Config{
		Version:  1,
		Defaults: Defaults{ContextTokens: 8000, RepoMapTokens: 2048, AskTokens: 4000, PackTokens: 8000},
		Structure: analyze.DefaultStructureConfig(),
		Retrieval: retrieval.DefaultSettings(),
	}
}

func Load(root string) (*Config, error) {
	cfg := Default()
	data, err := os.ReadFile(Path(root))
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if cfg.Version != 1 {
		return nil, fmt.Errorf("unsupported config version %d", cfg.Version)
	}
	fillDefaults(cfg)
	return cfg, nil
}

func fillDefaults(cfg *Config) {
	d := Default()
	if cfg.Defaults.ContextTokens <= 0 {
		cfg.Defaults.ContextTokens = d.Defaults.ContextTokens
	}
	if cfg.Defaults.RepoMapTokens <= 0 {
		cfg.Defaults.RepoMapTokens = d.Defaults.RepoMapTokens
	}
	if cfg.Defaults.AskTokens <= 0 {
		cfg.Defaults.AskTokens = d.Defaults.AskTokens
	}
	if cfg.Defaults.PackTokens <= 0 {
		cfg.Defaults.PackTokens = d.Defaults.PackTokens
	}
	if cfg.Structure.MaxFindings <= 0 {
		cfg.Structure = d.Structure
	}
	cfg.Retrieval.FillDefaults()
}

func Save(root string, cfg *Config) error {
	if err := os.MkdirAll(filepath.Join(root, ".carto"), 0755); err != nil {
		return fmt.Errorf("create .carto: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(Path(root), append(data, '\n'), 0644)
}

func Validate(root string) []string {
	if _, err := os.Stat(Path(root)); os.IsNotExist(err) {
		return nil
	}
	_, err := Load(root)
	if err != nil {
		return []string{err.Error()}
	}
	return nil
}
