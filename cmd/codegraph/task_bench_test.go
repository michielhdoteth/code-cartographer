package codegraph

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunTaskBenchmark(t *testing.T) {
	root := t.TempDir()
	for name, body := range map[string]string{
		"main.go": "package main\nfunc Main() string { return \"main\" }\n",
		"util.go": "package main\nfunc Helper() string { return \"help\" }\n",
	} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	index, err := Build(root)
	if err != nil {
		t.Fatal(err)
	}
	suite := TaskSuite{
		Version: 1,
		Cases: []Task{
			{ID: "helper", Query: "Helper", Expected: []string{"Helper"}, Tokens: 1000},
		},
	}
	report, err := index.RunTaskBenchmark(suite, 1000)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Cases) != 1 || report.SchemaVersion != 1 {
		t.Fatalf("unexpected report: %+v", report)
	}
	if !report.Deterministic {
		t.Fatal("benchmark should be deterministic")
	}
	if report.Cases[0].Recall <= 0 {
		t.Fatalf("expected to retrieve Helper, got %+v", report.Cases[0])
	}
	if report.Cases[0].SavedTokensEstimate < 0 {
		t.Fatal("negative savings")
	}
}
