package manifest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type Manifest struct {
	Repo      string         `json:"repo"`
	Commit    string         `json:"commit"`
	Branch    string         `json:"branch"`
	MappedAt  time.Time      `json:"mapped_at"`
	FileCount int            `json:"file_count"`
	Languages map[string]int `json:"languages"`
}

const manifestDir = ".carto"
const manifestFile = "manifest.json"

func savePath(dir string) string {
	return filepath.Join(dir, manifestDir, manifestFile)
}

// Save writes the manifest to .carto/manifest.json
func Save(dir string, manifest *Manifest) error {
	path := savePath(dir)
	os.MkdirAll(filepath.Dir(path), 0755)
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// Load reads the manifest from .carto/manifest.json
func Load(dir string) (*Manifest, error) {
	path := savePath(dir)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// Exists checks if a manifest file exists
func Exists(dir string) bool {
	_, err := os.Stat(savePath(dir))
	return err == nil
}
