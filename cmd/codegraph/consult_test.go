package codegraph

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func makeLargeFixture(t *testing.T) (string, string, []string) {
	t.Helper()
	root := t.TempDir()
	var lines []string
	lines = append(lines, "package main", "", "import \"fmt\"", "")
	for i := 0; i < 40; i++ {
		lines = append(lines, fmt.Sprintf("// Doc comment for Func%d", i))
		lines = append(lines, fmt.Sprintf("func Func%d() string {", i))
		lines = append(lines, fmt.Sprintf("\tfmt.Println(\"BODY_MARKER_%d\")", i))
		for j := 0; j < 40; j++ {
			lines = append(lines, "\tprintln(\"line\")")
		}
		lines = append(lines, "}", "")
	}
	path := filepath.Join(root, "huge.go")
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		t.Fatal(err)
	}
	index, err := Build(root)
	if err != nil || len(index.Files) == 0 {
		t.Fatalf("build failed: %v", err)
	}
	return root, "huge.go", lines
}

func TestGenerateSkeletonBudgetAndNoBodies(t *testing.T) {
	root, rel, _ := makeLargeFixture(t)
	index, err := Build(root)
	if err != nil {
		t.Fatal(err)
	}
	skeleton := index.GenerateSkeleton(rel, 512)
	if skeleton.TotalSymbols < 40 || len(skeleton.Symbols) == 0 {
		t.Fatalf("unexpected symbols: %+v", skeleton)
	}
	if EstimateTokens(skeleton.RenderMarkdown()) > 512 {
		t.Fatalf("skeleton exceeds budget: %d", EstimateTokens(skeleton.RenderMarkdown()))
	}
	if strings.Contains(skeleton.RenderMarkdown(), "BODY_MARKER") {
		t.Fatal("skeleton leaked function bodies")
	}
	if skeleton.Savings == nil || skeleton.Savings.ContextEfficiencyScore <= 0 {
		t.Fatalf("missing savings: %+v", skeleton.Savings)
	}
}

func TestChunkManifestContiguousAndBounded(t *testing.T) {
	root, rel, lines := makeLargeFixture(t)
	index, err := Build(root)
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := index.BuildChunkManifest(rel, 400)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.TotalChunks < 2 {
		t.Fatalf("expected multiple chunks, got %d: %+v (nodes=%d)", manifest.TotalChunks, manifest.Chunks, len(index.Files[rel].Nodes))
	}
	previousEnd := 0
	for i, chunk := range manifest.Chunks {
		if i > 0 && chunk.LineStart != previousEnd+1 {
			t.Fatalf("non-contiguous chunks at %d: %d != %d", i, chunk.LineStart, previousEnd+1)
		}
		if chunk.Tokens > 400 {
			t.Fatalf("chunk %d exceeds budget: %d", i, chunk.Tokens)
		}
		previousEnd = chunk.LineEnd
	}
	if previousEnd != len(lines) {
		t.Fatalf("chunks do not cover file: %d != %d", previousEnd, len(lines))
	}
	first, err := index.GenerateChunk(rel, 0, 400, 64)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(strings.Split(first.Content, "\n")[0]) != "package main" {
		t.Fatalf("unexpected chunk content: %s", first.Content[:80])
	}
	last, err := index.GenerateChunk(rel, manifest.TotalChunks-1, 400, 64)
	if err != nil || last.NextChunk != -1 {
		t.Fatalf("last chunk navigation wrong: %+v %v", last, err)
	}
	if _, err := index.GenerateChunk(rel, manifest.TotalChunks, 400, 64); err == nil {
		t.Fatal("expected out-of-range error")
	}
}
