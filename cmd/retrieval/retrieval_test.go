package retrieval

import (
	"math"
	"testing"
)

func TestTokenizeSplitsIdentifiers(t *testing.T) {
	cases := map[string][]string{
		"AnalyzeBlastRadius": {"analyze", "blast", "radius"},
		"HTTPServer":         {"http", "server"},
		"sha256_hash":        {"sha", "256", "hash"},
	}
	for input, want := range cases {
		got := Tokenize(input)
		if len(got) != len(want) {
			t.Fatalf("Tokenize(%q) = %v, want %v", input, got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("Tokenize(%q) = %v, want %v", input, got, want)
			}
		}
	}
	if tokens := Tokenize("return the error"); len(tokens) != 1 || tokens[0] != "error" {
		t.Fatalf("stopwords not removed: %v", tokens)
	}
}

func TestIDFDecreasesWithDocFreq(t *testing.T) {
	a := IDF(100, 1)
	b := IDF(100, 50)
	c := IDF(100, 100)
	if !(a > b && b > c) {
		t.Fatalf("idf not monotonic: %f %f %f", a, b, c)
	}
	if math.IsNaN(a) || math.IsNaN(c) {
		t.Fatal("NaN idf")
	}
}

func TestBM25FNameWeightAndSaturation(t *testing.T) {
	docs := []Document{
		{ID: "a", Fields: map[string][]string{"name": {"parse", "parse", "token"}, "body": {"other"}}},
		{ID: "b", Fields: map[string][]string{"name": {"other"}, "body": {"parse", "parse", "parse"}}},
	}
	stats := NewStats(docs)
	s := DefaultSettings().Scorer(stats)
	if s.Score([]string{"parse"}, docs[0]) <= s.Score([]string{"parse"}, docs[1]) {
		t.Fatal("name field should outrank body field at equal tf")
	}
	one := s.Score([]string{"parse"}, Document{ID: "x", Fields: map[string][]string{"name": {"parse"}}})
	two := s.Score([]string{"parse"}, Document{ID: "x", Fields: map[string][]string{"name": {"parse", "parse"}}})
	if two <= one {
		t.Fatal("tf should increase score")
	}
}

func TestFuzzyScorePrefersWordStartsAndConsecutiveMatches(t *testing.T) {
	good := FuzzyScore("AnalyzeBlastRadius", "AnalyzeBlastRadius")
	bad := FuzzyScore("AnalyzeBlastRadius", "xAnalyzeBlastRadius")
	if good <= 0 || bad <= 0 {
		t.Fatalf("expected positive fuzzy scores, got %f and %f", good, bad)
	}
	if good <= bad {
		t.Fatalf("exact identifier should outrank prefixed match: %f <= %f", good, bad)
	}
	if FuzzyScore("zzz", "abc") != 0 {
		t.Fatal("non-subsequence should score zero")
	}
}

func TestRRFAndAdaptiveDepth(t *testing.T) {
	fused := RRF([][]string{{"a", "b"}, {"b", "c"}}, 60)
	if fused["b"] <= fused["a"] || fused["b"] <= fused["c"] {
		t.Fatalf("rrf fusion wrong: %+v", fused)
	}
	weighted := WeightedRRF([][]string{{"a"}, {"b"}}, []float64{5, 1}, 60)
	if weighted["a"] <= weighted["b"] {
		t.Fatalf("weighted RRF wrong: %+v", weighted)
	}
	if AdaptiveDepth(8) != 200 || AdaptiveDepth(100) > 2000 {
		t.Fatalf("adaptive depth wrong: %d %d", AdaptiveDepth(8), AdaptiveDepth(100))
	}
}

func TestCacheRoundTripAndInvalidation(t *testing.T) {
	root := t.TempDir()
	docs := []Document{{ID: "a", Fields: map[string][]string{"name": {"parse"}}}}
	cache := BuildCache("graph1", "settings1", docs)
	if err := SaveCache(root, cache); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadCache(root, "graph1", "settings1")
	if err != nil || loaded.Documents["a"].ID != "a" {
		t.Fatalf("cache round trip failed: %v %+v", err, loaded)
	}
	if _, err := LoadCache(root, "graph2", "settings1"); err != ErrCacheInvalid {
		t.Fatalf("expected invalid cache on graph change, got %v", err)
	}
}
