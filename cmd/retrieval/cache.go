package retrieval

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Cache struct {
	Version         int                 `json:"version"`
	TokenizerVersion int                `json:"tokenizer_version"`
	SettingsHash    string              `json:"settings_hash"`
	GraphHash       string              `json:"graph_hash"`
	Documents       map[string]Document `json:"documents"`
	Stats           Stats               `json:"stats"`
}

var ErrCacheInvalid = errors.New("retrieval cache is missing or invalid")

func CachePath(root string) string {
	return filepath.Join(root, ".carto", "retrieval.json")
}

func ComputeGraphHash(fileHashes map[string]string) string {
	keys := make([]string, 0, len(fileHashes))
	for key := range fileHashes {
		keys = append(keys, filepath.ToSlash(key))
	}
	sort.Strings(keys)
	var sb strings.Builder
	for _, key := range keys {
		sb.WriteString(key)
		sb.WriteByte(0)
		sb.WriteString(fileHashes[key])
		sb.WriteByte('\n')
	}
	sum := sha256.Sum256([]byte(sb.String()))
	return hex.EncodeToString(sum[:])
}

func LoadCache(root, graphHash, settingsHash string) (*Cache, error) {
	data, err := os.ReadFile(CachePath(root))
	if err != nil {
		return nil, ErrCacheInvalid
	}
	var cache Cache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, ErrCacheInvalid
	}
	if cache.Version != 1 || cache.TokenizerVersion != TokenizerVersion ||
		cache.GraphHash != graphHash || cache.SettingsHash != settingsHash {
		return nil, ErrCacheInvalid
	}
	return &cache, nil
}

func BuildCache(graphHash, settingsHash string, docs []Document) *Cache {
	cache := &Cache{
		Version: 1, TokenizerVersion: TokenizerVersion,
		SettingsHash: settingsHash, GraphHash: graphHash,
		Documents: map[string]Document{},
	}
	for _, doc := range docs {
		cache.Documents[doc.ID] = doc
	}
	cache.Stats = NewStats(docs)
	return cache
}

func SaveCache(root string, cache *Cache) error {
	path := CachePath(root)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
