package search

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSearchRanksByBM25F(t *testing.T) {
	root := t.TempDir()
	for name, body := range map[string]string{
		"main.go": "package main\nfunc Main() { Helper() }\n",
		"util.go": "package main\nfunc Helper() string { return \"help\" }\n",
	} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	idx, err := BuildIndex(root)
	if err != nil {
		t.Fatal(err)
	}
	results := idx.Search("Helper", 5)
	if len(results) == 0 {
		t.Fatal("expected ranked results")
	}
	if results[0].File == "" || results[0].Score <= 0 {
		t.Fatalf("unexpected result: %+v", results[0])
	}
	if err := idx.Save(root); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadIndex(root)
	if err != nil || len(loaded.Terms) == 0 {
		t.Fatalf("index round trip failed: %v", err)
	}
}
