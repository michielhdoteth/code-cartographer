package codegraph

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"carto/lib/parser"
)

func testIndex(t *testing.T) *Index {
	t.Helper()
	root := t.TempDir()
	files := map[string]File{
		"src/a.go": {Language: "go", Lines: 4, Nodes: []parser.Node{{Name: "Start", Type: parser.NodeTypeFunction, StartLine: 1, EndLine: 2, Signature: "func Start()", Exported: true}}, Imports: []parser.Import{{Source: "./b", Line: 1}}},
		"src/b.go": {Language: "go", Lines: 4, Nodes: []parser.Node{{Name: "Middle", Type: parser.NodeTypeFunction, StartLine: 1, EndLine: 2, Signature: "func Middle()"}}},
		"src/c.go": {Language: "go", Lines: 4, Nodes: []parser.Node{{Name: "Finish", Type: parser.NodeTypeFunction, StartLine: 1, EndLine: 2, Signature: "func Finish()"}}},
	}
	contents := map[string]string{
		"src/a.go": "package a\nfunc Start() { Middle() }\nfunc noisy() {}\n",
		"src/b.go": "package b\nfunc Middle() { Finish() }\nfunc unused() {}\n",
		"src/c.go": "package c\nfunc Finish() {}\n",
	}
	for path, body := range contents {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(root, path)), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, path), []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	for path, file := range files {
		file.Hash = "fixture"
		files[path] = file
	}
	index := &Index{RootPath: root, Files: files, Nodes: []Node{
		{ID: "file:src/a.go", Kind: "file", Name: "a.go", File: "src/a.go"},
		{ID: "file:src/b.go", Kind: "file", Name: "b.go", File: "src/b.go"},
		{ID: "file:src/c.go", Kind: "file", Name: "c.go", File: "src/c.go"},
		{ID: "symbol:start", Name: "Start", Kind: "function", File: "src/a.go", StartLine: 2, EndLine: 2, Signature: "func Start()", Exported: true},
		{ID: "symbol:middle", Name: "Middle", Kind: "function", File: "src/b.go", StartLine: 2, EndLine: 2, Signature: "func Middle()"},
		{ID: "symbol:finish", Name: "Finish", Kind: "function", File: "src/c.go", StartLine: 2, EndLine: 2, Signature: "func Finish()"},
	}, Edges: []Edge{
		{ID: "a-b", Source: "symbol:start", Target: "symbol:middle", Type: "calls", Confidence: "name", Line: 2},
		{ID: "b-c", Source: "symbol:middle", Target: "symbol:finish", Type: "calls", Confidence: "exact", Line: 2},
		{ID: "a-import-b", Source: "file:src/a.go", Target: "file:src/b.go", Type: "imports", Confidence: "exact", Line: 1},
	}}
	return index
}

func TestGenerateRepoMapIsStableFocusedAndBudgeted(t *testing.T) {
	index := testIndex(t)
	first := index.GenerateRepoMap(1000, nil)
	second := index.GenerateRepoMap(1000, nil)
	if first.RenderMarkdown() != second.RenderMarkdown() {
		t.Fatal("repository map changed between identical runs")
	}
	focused := index.GenerateRepoMap(1000, []string{"src/c.go"})
	if len(focused.Entries) == 0 || focused.Entries[0].Node.File != "src/c.go" {
		t.Fatalf("focused path did not rank first: %+v", focused.Entries)
	}
	small := index.GenerateRepoMap(12, nil)
	if got := EstimateTokens(small.RenderMarkdown()); got > small.TokenBudget {
		t.Fatalf("map used %d estimated tokens with budget %d", got, small.TokenBudget)
	}
	if strings.Contains(first.RenderMarkdown(), "func noisy") {
		t.Fatal("signature map unexpectedly included implementation bodies")
	}
	if len(first.Imports) == 0 || !strings.Contains(first.RenderMarkdown(), "imports: ./b") {
		t.Fatal("ranked map omitted repository import summaries")
	}
}

func TestGeneratePackFiltersCompressesAndRespectsBudget(t *testing.T) {
	index := testIndex(t)
	filtered := index.GeneratePack(PackOptions{TokenBudget: 1000, Include: []string{"src/**"}, Exclude: []string{"**/b.go"}, Compress: true})
	if filtered.FilesIncluded != 2 || strings.Contains(filtered.Content, "src/b.go") {
		t.Fatalf("unexpected filtered pack: included=%d content=%s", filtered.FilesIncluded, filtered.Content)
	}
	if strings.Contains(filtered.Content, "noisy") || !strings.Contains(filtered.Content, "func Start()") {
		t.Fatalf("compressed output should contain signatures but omit bodies: %s", filtered.Content)
	}
	small := index.GeneratePack(PackOptions{TokenBudget: 40})
	if small.TokensUsed > small.TokenBudget {
		t.Fatalf("pack used %d estimated tokens with budget %d", small.TokensUsed, small.TokenBudget)
	}
}

func TestTraceDepthDirectionAndAmbiguity(t *testing.T) {
	index := testIndex(t)
	callers := index.TraceDepth("Finish", 2, "callers")
	if len(callers) != 2 || callers[0].Depth != 1 || callers[1].Depth != 2 {
		t.Fatalf("expected two caller hops, got %+v", callers)
	}
	callees := index.TraceDepth("Start", 2, "callees")
	if len(callees) != 2 || callees[0].Direction != "callee" {
		t.Fatalf("expected two callee hops, got %+v", callees)
	}
	if hits := index.TraceDepth("Start", 2, "callers"); len(hits) != 0 {
		t.Fatalf("caller-only traversal returned unrelated nodes: %+v", hits)
	}
	index.Edges = append(index.Edges, Edge{ID: "ambiguous", Source: "file:src/a.go", Target: "symbol-name:finish", Type: "calls", Confidence: "ambiguous", Line: 3})
	hits := index.TraceDepth("Finish", 1, "callers")
	found := false
	for _, hit := range hits {
		found = found || (hit.Node.File == "src/a.go" && hit.Confidence == "ambiguous")
	}
	if !found {
		t.Fatal("direct ambiguous caller evidence was not retained")
	}
}

func TestOfflineBenchmarkMeasuresRetrievalAndDeterminism(t *testing.T) {
	index := testIndex(t)
	report, err := index.RunBenchmark(BenchmarkSuite{Version: 1, Cases: []BenchmarkCase{{ID: "finish", Query: "Finish", Expected: []string{"Finish"}}}}, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Cases) != 1 || report.Cases[0].Recall != 1 || !report.Deterministic {
		t.Fatalf("unexpected benchmark result: %+v", report)
	}
	if _, err := index.RunBenchmark(BenchmarkSuite{Version: 2, Cases: []BenchmarkCase{{ID: "x", Query: "x", Expected: []string{"x"}}}}, 1); err == nil {
		t.Fatal("unsupported suite version should fail")
	}
}

func TestGlobMatchRecursiveAndSingleLevel(t *testing.T) {
	if !globMatch("src/**/*.go", "src/nested/file.go") || !globMatch("src/*.go", "src/file.go") {
		t.Fatal("expected supported glob to match")
	}
	if globMatch("src/*.go", "src/nested/file.go") {
		t.Fatal("single-level wildcard should not cross directory boundaries")
	}
}
