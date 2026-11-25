package analyze

import (
	"fmt"
	"strings"
	"testing"
)

func testNode(path string, exports []string, imports []Import) *FileNode {
	nodes := make([]Export, 0, len(exports))
	for _, e := range exports {
		nodes = append(nodes, Export{Name: e, Kind: "function", Line: 1})
	}
	return &FileNode{Path: path, Name: path, Exports: nodes, Imports: imports}
}

func rulesIn(report *StructureReport) map[string]bool {
	out := make(map[string]bool)
	for _, f := range report.Findings {
		out[f.Rule] = true
	}
	return out
}

func TestAnalyzeStructureModuleCycle(t *testing.T) {
	g := NewModuleGraph("", "go")
	g.Files["a/a.go"] = testNode("a/a.go", []string{"A"}, []Import{{Name: "B", Source: "b/b.go"}})
	g.Files["b/b.go"] = testNode("b/b.go", []string{"B"}, []Import{{Name: "A", Source: "a/a.go"}})

	report := AnalyzeStructure(g, DefaultStructureConfig())
	if report.Summary.Critical != 1 {
		t.Fatalf("expected 1 critical module cycle, got %d: %+v", report.Summary.Critical, report.Findings)
	}
	if !rulesIn(report)["arch/module-cycle"] {
		t.Fatalf("expected module-cycle rule, got %v", rulesIn(report))
	}
}

func TestAnalyzeStructureParentChildTangle(t *testing.T) {
	g := NewModuleGraph("", "go")
	g.Files["core/core.go"] = testNode("core/core.go", []string{"Run"}, []Import{
		{Name: "Tick", Source: "core/timers/tick.go"},
		{Name: "Stop", Source: "core/timers/tick.go"},
	})
	g.Files["core/timers/tick.go"] = testNode("core/timers/tick.go", []string{"Tick", "Stop"}, []Import{
		{Name: "Run", Source: "core/core.go"},
		{Name: "Job", Source: "core/core.go"},
	})
	g.Files["core/timers/timer.go"] = testNode("core/timers/timer.go", []string{"Timer"}, nil)

	report := AnalyzeStructure(g, DefaultStructureConfig())
	rules := rulesIn(report)
	if !rules["arch/tangle"] {
		t.Fatalf("expected tangle finding, got %v", report.Findings)
	}
	if rules["arch/module-cycle"] {
		t.Fatalf("parent/child tangle must not be double-reported as a cycle: %v", report.Findings)
	}
}

func TestAnalyzeStructureShallowModule(t *testing.T) {
	g := NewModuleGraph("", "go")
	utilsExports := []string{"Upper", "Lower", "Trim", "Split", "Join", "Pad"}
	g.Files["utils/helpers.go"] = testNode("utils/helpers.go", utilsExports, nil)
	for _, pair := range []struct{ file string; names []string }{
		{"app/handler.go", []string{"Upper", "Lower", "Trim"}},
		{"app/parser.go", []string{"Upper", "Split", "Join"}},
		{"cli/main.go", []string{"Upper", "Lower", "Pad"}},
	} {
		var imports []Import
		for _, name := range pair.names {
			imports = append(imports, Import{Name: name, Source: "utils/helpers.go"})
		}
		g.Files[pair.file] = testNode(pair.file, []string{"Do"}, imports)
	}

	report := AnalyzeStructure(g, DefaultStructureConfig())
	var shallow *StructureFinding
	for i := range report.Findings {
		if report.Findings[i].Rule == "arch/shallow-module" {
			shallow = &report.Findings[i]
		}
	}
	if shallow == nil {
		t.Fatalf("expected shallow-module finding, got %v", report.Findings)
	}
	if shallow.Module != "utils" {
		t.Fatalf("expected module 'utils', got %s", shallow.Module)
	}
	if shallow.Evidence.UsedExports != 6 || shallow.Evidence.TotalExports != 6 {
		t.Fatalf("unexpected evidence: %+v", shallow.Evidence)
	}
	if !strings.Contains(shallow.Hint, "Upper") {
		t.Fatalf("expected front-door symbol in hint, got %s", shallow.Hint)
	}
}

func TestAnalyzeStructureNoBroadInterfaceWhenUnresolved(t *testing.T) {
	g := NewModuleGraph("", "go")
	var exports []string
	for i := 0; i < 12; i++ {
		exports = append(exports, fmt.Sprintf("Export%d", i))
	}
	g.Files["utils/all.go"] = testNode("utils/all.go", exports, nil)
	g.Files["app/main.go"] = testNode("app/main.go", []string{"Main"}, []Import{{Name: "Unrelated", Source: "utils/all.go"}})

	report := AnalyzeStructure(g, DefaultStructureConfig())
	if rulesIn(report)["arch/broad-interface"] {
		t.Fatalf("unresolved symbols must not produce broad-interface findings: %v", report.Findings)
	}
}

func TestAnalyzeStructureAllowlist(t *testing.T) {
	g := NewModuleGraph("", "go")
	g.Files["a/a.go"] = testNode("a/a.go", []string{"A"}, []Import{{Name: "B", Source: "b/b.go"}})
	g.Files["b/b.go"] = testNode("b/b.go", []string{"B"}, []Import{{Name: "A", Source: "a/a.go"}})

	cfg := DefaultStructureConfig()
	cfg.AllowPath = "nonexistent-allow-file"
	cfg.ExtraAllow = []string{"arch/module-cycle|a + b"}
	report := AnalyzeStructure(g, cfg)
	if len(report.Findings) != 0 {
		t.Fatalf("expected suppressed findings, got %v", report.Findings)
	}
	if len(report.Suppressed) != 1 {
		t.Fatalf("expected 1 suppressed finding, got %v", report.Suppressed)
	}
}

func TestAnalyzeStructureDeterministicOutput(t *testing.T) {
	g := NewModuleGraph("", "go")
	g.Files["a/a.go"] = testNode("a/a.go", []string{"A"}, []Import{{Name: "B", Source: "b/b.go"}})
	g.Files["b/b.go"] = testNode("b/b.go", []string{"B"}, []Import{{Name: "A", Source: "a/a.go"}})
	g.Files["c/c.go"] = testNode("c/c.go", []string{"C"}, nil)

	first := AnalyzeStructure(g, DefaultStructureConfig()).String()
	for i := 0; i < 5; i++ {
		if got := AnalyzeStructure(g, DefaultStructureConfig()).String(); got != first {
			t.Fatalf("non-deterministic output on run %d", i)
		}
	}
	if !strings.Contains(first, "Structural Health Report") {
		t.Fatalf("missing report header: %s", first)
	}
}
