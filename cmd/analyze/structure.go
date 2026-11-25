package analyze

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const StructureReportVersion = 1

// StructureConfig holds tunable thresholds for structural analysis.
// Every threshold is heuristic; findings are signals, not proofs.
type StructureConfig struct {
	ShallowUseRatio    float64 `json:"shallow_use_ratio"`
	ShallowMinExports  int     `json:"shallow_min_exports"`
	MaxLinesPerExport  int     `json:"max_lines_per_export"`
	BroadMinExports    int     `json:"broad_min_exports"`
	BroadIdleRatio     float64 `json:"broad_idle_ratio"`
	TangleMinEdges     int     `json:"tangle_min_edges"`
	ScatterMinModules  int     `json:"scatter_min_modules"`
	ScatterMinTopDirs  int     `json:"scatter_min_top_dirs"`
	MaxFindings        int     `json:"max_findings"`
	AllowPath          string  `json:"allow_path,omitempty"`
	ExtraAllow         []string `json:"extra_allow,omitempty"`
}

// DefaultStructureConfig returns conservative defaults, calibrated to avoid
// flooding a typical repository with low-confidence findings.
func DefaultStructureConfig() StructureConfig {
	return StructureConfig{
		ShallowUseRatio:   0.50,
		ShallowMinExports: 4,
		MaxLinesPerExport: 25,
		BroadMinExports:   8,
		BroadIdleRatio:    0.60,
		TangleMinEdges:    3,
		ScatterMinModules: 4,
		ScatterMinTopDirs: 3,
		MaxFindings:       50,
	}
}

// ModuleInfo is one structural module (a package directory, collapsed upward
// when a directory holds fewer than two source files).
type ModuleInfo struct {
	Path          string   `json:"path"`
	Files         []string `json:"files"`
	Exports       []string `json:"exports"`
	Lines         int      `json:"lines"`
	InternalEdges int      `json:"internal_edges"`
	FanIn         int      `json:"fan_in"`
	FanOut        int      `json:"fan_out"`
	Depth         float64  `json:"depth"`
}

type StructureEvidence struct {
	UsedExports   int      `json:"used_exports,omitempty"`
	TotalExports  int      `json:"total_exports,omitempty"`
	LinesPerExport float64 `json:"lines_per_export,omitempty"`
	MutualEdges   int      `json:"mutual_edges,omitempty"`
	ReachModules  int      `json:"reach_modules,omitempty"`
	TopDirs       []string `json:"top_dirs,omitempty"`
	Symbols       []string `json:"symbols,omitempty"`
	Modules       []string `json:"modules,omitempty"`
}

type StructureFinding struct {
	ID       string           `json:"id"`
	Rule     string           `json:"rule"`
	Severity string           `json:"severity"` // "critical", "warning", "info"
	Module   string           `json:"module"`
	Message  string           `json:"message"`
	Evidence StructureEvidence `json:"evidence"`
	Hint     string           `json:"hint"`
	Files    []string         `json:"files,omitempty"`
}

type StructureSummary struct {
	Modules  int `json:"modules"`
	Critical int `json:"critical"`
	Warning  int `json:"warning"`
	Info     int `json:"info"`
	Healthy  int `json:"healthy"`
	Total    int `json:"total"`
}

type StructureReport struct {
	Version    int                `json:"version"`
	RootPath   string             `json:"root_path"`
	Summary    StructureSummary   `json:"summary"`
	Modules    []ModuleInfo       `json:"modules"`
	Findings   []StructureFinding `json:"findings"`
	Suppressed []string           `json:"suppressed,omitempty"`
}

type structureFacts struct {
	fileModule    map[string]string
	modules       []ModuleInfo
	byName        map[string]*ModuleInfo
	edgeCount     map[string]int
	fanIn         map[string]int
	fanOut        map[string]int
	internalEdges map[string]int
	symbolUsers   map[string]map[string]int // module -> symbol -> distinct external modules (approx by count)
	symbolModules map[string]map[string]bool // symbol -> modules using it
}

func AnalyzeStructure(graph *ModuleGraph, cfg StructureConfig) *StructureReport {
	if cfg.MaxFindings <= 0 {
		cfg = DefaultStructureConfig()
	}
	report := &StructureReport{
		Version:  StructureReportVersion,
		RootPath: graph.RootPath,
		Modules:  []ModuleInfo{},
		Findings: []StructureFinding{},
	}

	facts := buildStructureFacts(graph)
	report.Modules = facts.modules
	report.Summary.Modules = len(facts.modules)

	findings := make([]StructureFinding, 0)
	findings = append(findings, detectModuleCycles(facts)...)
	findings = append(findings, detectTangles(facts, cfg)...)
	shallow := detectShallowModules(facts, cfg)
	findings = append(findings, shallow...)
	shallowModules := make(map[string]bool)
	for _, f := range shallow {
		shallowModules[f.Module] = true
	}
	findings = append(findings, detectBroadInterfaces(facts, cfg, shallowModules)...)
	findings = append(findings, detectScatteredFeatures(facts, cfg)...)

	sortFindings(findings)

	allowed := loadStructureAllowlist(graph.RootPath, cfg)
	flagged := make(map[string]bool)
	for _, f := range findings {
		if allowed[f.ID] || allowed[f.Rule+"|"+f.Module] {
			report.Suppressed = append(report.Suppressed, f.ID)
			continue
		}
		if len(report.Findings) >= cfg.MaxFindings {
			break
		}
		report.Findings = append(report.Findings, f)
		flagged[f.Module] = true
		switch f.Severity {
		case "critical":
			report.Summary.Critical++
		case "warning":
			report.Summary.Warning++
		default:
			report.Summary.Info++
		}
	}
	for _, f := range report.Findings {
		flagged[f.Module] = true
	}
	for _, m := range facts.modules {
		if !flagged[m.Path] {
			report.Summary.Healthy++
		}
	}
	report.Summary.Total = len(report.Findings)
	sort.Strings(report.Suppressed)

	return report
}

func sortFindings(findings []StructureFinding) {
	sort.Slice(findings, func(i, j int) bool {
		ri, rj := severityRank(findings[i].Severity), severityRank(findings[j].Severity)
		if ri != rj {
			return ri < rj
		}
		if findings[i].Rule != findings[j].Rule {
			return findings[i].Rule < findings[j].Rule
		}
		if findings[i].Module != findings[j].Module {
			return findings[i].Module < findings[j].Module
		}
		return findings[i].ID < findings[j].ID
	})
}

func severityRank(s string) int {
	switch s {
	case "critical":
		return 0
	case "warning":
		return 1
	default:
		return 2
	}
}

func buildStructureFacts(graph *ModuleGraph) *structureFacts {
	facts := &structureFacts{
		fileModule:    make(map[string]string),
		byName:        make(map[string]*ModuleInfo),
		edgeCount:     make(map[string]int),
		fanIn:         make(map[string]int),
		fanOut:        make(map[string]int),
		internalEdges: make(map[string]int),
		symbolUsers:   make(map[string]map[string]int),
		symbolModules: make(map[string]map[string]bool),
	}

	dirCounts := make(map[string]int)
	for file := range graph.Files {
		dirCounts[moduleDirOf(file)]++
	}
	for file := range graph.Files {
		dir := moduleDirOf(file)
		for strings.Contains(dir, "/") && dirCounts[dir] < 2 {
			parent := filepath.ToSlash(filepath.Dir(dir))
			if parent == dir || parent == "" {
				break
			}
			dir = parent
		}
		facts.fileModule[file] = dir
		if _, ok := facts.byName[dir]; !ok {
			info := &ModuleInfo{Path: dir, Files: []string{}, Exports: []string{}}
			facts.byName[dir] = info
			facts.modules = append(facts.modules, *info)
		}
	}

	for i := range facts.modules {
		mod := &facts.modules[i]
		for file, m := range facts.fileModule {
			if m != mod.Path {
				continue
			}
			mod.Files = append(mod.Files, file)
			node := graph.Files[file]
			if node == nil {
				continue
			}
			mod.Lines += countSourceLines(graph.RootPath, file, len(node.Exports))
			for _, exp := range node.Exports {
				mod.Exports = append(mod.Exports, exp.Name)
			}
		}
		sort.Strings(mod.Files)
		sort.Strings(mod.Exports)
		facts.byName[mod.Path] = mod
	}

	for file, node := range graph.Files {
		from := facts.fileModule[file]
		for _, imp := range node.Imports {
			resolved := resolveImportToFilePath(imp.Source, graph)
			if resolved == "" || resolved == file {
				continue
			}
			to, ok := facts.fileModule[resolved]
			if !ok {
				continue
			}
			if from == to {
				facts.internalEdges[from]++
				continue
			}
			facts.fanOut[from]++
			facts.fanIn[to]++
			facts.edgeCount[from+"|"+to]++
			if facts.symbolUsers[to] == nil {
				facts.symbolUsers[to] = make(map[string]int)
			}
			facts.symbolUsers[to][imp.Name]++
			if facts.symbolModules[imp.Name] == nil {
				facts.symbolModules[imp.Name] = make(map[string]bool)
			}
			facts.symbolModules[imp.Name][from] = true
		}
	}

	for i := range facts.modules {
		mod := &facts.modules[i]
		mod.FanIn = facts.fanIn[mod.Path]
		mod.FanOut = facts.fanOut[mod.Path]
		mod.InternalEdges = facts.internalEdges[mod.Path]
		if len(mod.Exports) > 0 {
			mod.Depth = math.Log(1+float64(mod.InternalEdges)) / float64(len(mod.Exports))
		}
		facts.byName[mod.Path] = mod
	}
	sort.Slice(facts.modules, func(i, j int) bool { return facts.modules[i].Path < facts.modules[j].Path })
	return facts
}

func moduleDirOf(file string) string {
	dir := filepath.ToSlash(filepath.Dir(file))
	if dir == "" {
		return "."
	}
	return dir
}

// estimatedLinesPerExport approximates file length when the source is unreadable.
const estimatedLinesPerExport = 10

func countSourceLines(root, file string, fallbackExports int) int {
	if root != "" && file != "" {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(file)))
		if err == nil {
			lines := 0
			for _, line := range strings.Split(string(data), "\n") {
				if strings.TrimSpace(line) != "" {
					lines++
				}
			}
			return lines
		}
	}
	return fallbackExports * estimatedLinesPerExport
}

func detectModuleCycles(facts *structureFacts) []StructureFinding {
	adjacency := make(map[string][]string)
	for _, m := range facts.modules {
		for key, count := range facts.edgeCount {
			if count == 0 {
				continue
			}
			parts := strings.SplitN(key, "|", 2)
			if len(parts) == 2 && parts[0] == m.Path {
				adjacency[m.Path] = append(adjacency[m.Path], parts[1])
			}
		}
		sort.Strings(adjacency[m.Path])
	}

	var result []StructureFinding
	for _, scc := range tarjanSCC(adjacency, facts.modules) {
		if len(scc) < 2 {
			continue
		}
		// A parent/child mutual dependency is reported as a tangle (wrong
		// layering), not as a generic module cycle.
		if len(scc) == 2 && (isAncestorModule(scc[0], scc[1]) || isAncestorModule(scc[1], scc[0])) {
			continue
		}
		sort.Strings(scc)
		files := make(map[string]bool)
		for _, m := range scc {
			for _, f := range facts.byName[m].Files {
				files[f] = true
			}
		}
		fileList := sortedKeys(files)
		result = append(result, StructureFinding{
			ID:       findingID("arch/module-cycle", strings.Join(scc, ",")),
			Rule:     "arch/module-cycle",
			Severity: "critical",
			Module:   strings.Join(scc, " + "),
			Message:  fmt.Sprintf("module cycle across %d modules: %s", len(scc), strings.Join(scc, " -> ")),
			Evidence: StructureEvidence{Modules: scc},
			Hint:     "Put the shared contract in a lower layer both modules can depend on; do not break the cycle by duplicating types.",
			Files:    fileList,
		})
	}
	return result
}

func tarjanSCC(adjacency map[string][]string, modules []ModuleInfo) [][]string {
	index := make(map[string]int)
	low := make(map[string]int)
	onStack := make(map[string]bool)
	var stack []string
	var result [][]string
	next := 0

	var strongconnect func(string)
	strongconnect = func(v string) {
		index[v] = next
		low[v] = next
		next++
		stack = append(stack, v)
		onStack[v] = true
		for _, w := range adjacency[v] {
			if _, ok := index[w]; !ok {
				strongconnect(w)
				if low[w] < low[v] {
					low[v] = low[w]
				}
			} else if onStack[w] && index[w] < low[v] {
				low[v] = index[w]
			}
		}
		if low[v] != index[v] {
			return
		}
		var component []string
		for {
			w := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			onStack[w] = false
			component = append(component, w)
			if w == v {
				break
			}
		}
		result = append(result, component)
	}
	for _, m := range modules {
		if _, ok := index[m.Path]; !ok {
			strongconnect(m.Path)
		}
	}
	return result
}

// detectTangles finds parent/child modules that know each other. Unlike a
// peer-to-peer module cycle, this signals a layer that leaks upward or downward.
func detectTangles(facts *structureFacts, cfg StructureConfig) []StructureFinding {
	var result []StructureFinding
	seen := make(map[string]bool)
	for key, count := range facts.edgeCount {
		parts := strings.SplitN(key, "|", 2)
		if len(parts) != 2 {
			continue
		}
		a, b := parts[0], parts[1]
		if !isAncestorModule(a, b) && !isAncestorModule(b, a) {
			continue // peer cycles are handled by arch/module-cycle
		}
		pair := a + "|" + b
		reverse := b + "|" + a
		if seen[pair] || seen[reverse] {
			continue
		}
		reverseCount := facts.edgeCount[reverse]
		if reverseCount == 0 || count+reverseCount < cfg.TangleMinEdges {
			continue
		}
		seen[pair] = true
		files := append([]string{}, facts.byName[a].Files...)
		files = append(files, facts.byName[b].Files...)
		result = append(result, StructureFinding{
			ID:       findingID("arch/tangle", pair),
			Rule:     "arch/tangle",
			Severity: "warning",
			Module:   a + " <-> " + b,
			Message:  fmt.Sprintf("mutual references between %s and %s (%d edges) without a module cycle", a, b, count+reverseCount),
			Evidence: StructureEvidence{MutualEdges: count + reverseCount, Modules: []string{a, b}},
			Hint:     "Decide which module owns the shared idea and let the other call down; remove the reverse dependency.",
			Files:    files,
		})
	}
	return result
}

func moduleAdjacency(facts *structureFacts) map[string][]string {
	adj := make(map[string][]string)
	for key, count := range facts.edgeCount {
		if count == 0 {
			continue
		}
		parts := strings.SplitN(key, "|", 2)
		if len(parts) == 2 {
			adj[parts[0]] = append(adj[parts[0]], parts[1])
		}
	}
	return adj
}

func isAncestorModule(a, b string) bool {
	return strings.HasPrefix(b, a+"/")
}

func detectShallowModules(facts *structureFacts, cfg StructureConfig) []StructureFinding {
	var result []StructureFinding
	for _, mod := range facts.modules {
		exportSet := make(map[string]bool, len(mod.Exports))
		for _, e := range mod.Exports {
			exportSet[e] = true
		}
		used := 0
		type symbolCount struct {
			name  string
			count int
		}
		var top []symbolCount
		for symbol, count := range facts.symbolUsers[mod.Path] {
			if !exportSet[symbol] {
				continue
			}
			used++
			top = append(top, symbolCount{symbol, count})
		}
		if len(mod.Exports) < cfg.ShallowMinExports || used == 0 {
			continue
		}
		ratio := float64(used) / float64(len(mod.Exports))
		linesPerExport := float64(mod.Lines) / float64(len(mod.Exports))
		if ratio < cfg.ShallowUseRatio || linesPerExport > float64(cfg.MaxLinesPerExport) {
			continue
		}
		sort.Slice(top, func(i, j int) bool {
			if top[i].count != top[j].count {
				return top[i].count > top[j].count
			}
			return top[i].name < top[j].name
		})
		var symbols []string
		for i, s := range top {
			if i >= 3 {
				break
			}
			symbols = append(symbols, s.name)
		}
		result = append(result, StructureFinding{
			ID:       findingID("arch/shallow-module", mod.Path),
			Rule:     "arch/shallow-module",
			Severity: "warning",
			Module:   mod.Path,
			Message:  fmt.Sprintf("shallow module: callers use %d of %d exports (%.0f%%), only %.0f lines per export", used, len(mod.Exports), ratio*100, linesPerExport),
			Evidence: StructureEvidence{
				UsedExports:    used,
				TotalExports:   len(mod.Exports),
				LinesPerExport: math.Round(linesPerExport*10) / 10,
				Symbols:        symbols,
			},
			Hint: fmt.Sprintf("Collapse the surface behind %d entry points (%s); hide helpers and group by what changes together.",
				minInt(3, len(symbols)), strings.Join(symbols, ", ")),
			Files: mod.Files,
		})
	}
	return result
}

func detectBroadInterfaces(facts *structureFacts, cfg StructureConfig, shallow map[string]bool) []StructureFinding {
	var exportCounts []int
	for _, m := range facts.modules {
		exportCounts = append(exportCounts, len(m.Exports))
	}
	sort.Ints(exportCounts)
	threshold := cfg.BroadMinExports
	if len(exportCounts) > 0 {
		p85 := exportCounts[(len(exportCounts)*85)/100]
		if p85 > threshold {
			threshold = p85
		}
	}

	var result []StructureFinding
	for _, mod := range facts.modules {
		if shallow[mod.Path] || len(mod.Exports) < threshold {
			continue
		}
		// If no external symbol reference resolved at all, the parser may simply
		// not understand this language's import vocabulary. Do not guess.
		resolved := 0
		idle := 0
		for _, exp := range mod.Exports {
			users := 0
			for symbol, modules := range facts.symbolModules {
				if symbol == exp {
					users = len(modules)
				}
			}
			if users > 0 {
				resolved++
			}
			if users <= 1 {
				idle++
			}
		}
		if resolved == 0 || mod.FanIn == 0 {
			continue
		}
		if float64(idle)/float64(len(mod.Exports)) < cfg.BroadIdleRatio {
			continue
		}
		result = append(result, StructureFinding{
			ID:       findingID("arch/broad-interface", mod.Path),
			Rule:     "arch/broad-interface",
			Severity: "info",
			Module:   mod.Path,
			Message:  fmt.Sprintf("broad interface: %d of %d exports are used by at most one module", idle, len(mod.Exports)),
			Evidence: StructureEvidence{UsedExports: idle, TotalExports: len(mod.Exports)},
			Hint:     "Remove speculative exports or make them package-private; keep only the vocabulary callers actually need.",
			Files:    mod.Files,
		})
	}
	return result
}

func detectScatteredFeatures(facts *structureFacts, cfg StructureConfig) []StructureFinding {
	var result []StructureFinding
	for symbol, modules := range facts.symbolModules {
		if len(modules) < cfg.ScatterMinModules {
			continue
		}
		topDirs := make(map[string]bool)
		for m := range modules {
			parts := strings.Split(m, "/")
			if len(parts) > 0 {
				topDirs[parts[0]] = true
			}
		}
		if len(topDirs) < cfg.ScatterMinTopDirs {
			continue
		}
		dirList := sortedKeys(topDirs)
		modList := sortedKeys(modules)
		result = append(result, StructureFinding{
			ID:       findingID("arch/scattered-feature", symbol),
			Rule:     "arch/scattered-feature",
			Severity: "warning",
			Module:   symbol,
			Message:  fmt.Sprintf("feature '%s' is used across %d modules in %d top-level areas: %s", symbol, len(modules), len(topDirs), strings.Join(dirList, ", ")),
			Evidence: StructureEvidence{ReachModules: len(modules), TopDirs: dirList, Symbols: []string{symbol}, Modules: modList},
			Hint:     "Colocate the code that changes with this feature; each extra directory makes agents fetch more context per task.",
		})
	}
	return result
}

func findingID(rule, key string) string {
	sum := sha256.Sum256([]byte(rule + "|" + key))
	return rule + ":" + hex.EncodeToString(sum[:])[:10]
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// loadStructureAllowlist reads a carto structure allowlist. Blank lines and
// lines starting with '#' are comments. Each entry is a finding ID or rule|module.
func loadStructureAllowlist(root string, cfg StructureConfig) map[string]bool {
	allowed := make(map[string]bool)
	for _, id := range cfg.ExtraAllow {
		if id != "" {
			allowed[id] = true
		}
	}
	path := cfg.AllowPath
	if path == "" && root != "" {
		path = filepath.Join(root, ".carto", "structure-allow")
	}
	if path == "" {
		return allowed
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return allowed
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		allowed[line] = true
	}
	return allowed
}

func (r *StructureReport) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n=== Structural Health Report ===\n\n"))
	sb.WriteString(fmt.Sprintf("%d modules · %d findings (🔴 %d critical · 🟡 %d warning · 🟢 %d healthy)\n\n",
		r.Summary.Modules, r.Summary.Total, r.Summary.Critical, r.Summary.Warning, r.Summary.Healthy))

	for _, tier := range []struct {
		name     string
		severity string
		marker   string
	}{
		{"CRITICAL", "critical", "🔴"},
		{"WARNING", "warning", "🟡"},
		{"INFO", "info", "ℹ"},
	} {
		var group []StructureFinding
		for _, f := range r.Findings {
			if f.Severity == tier.severity {
				group = append(group, f)
			}
		}
		if len(group) == 0 {
			continue
		}
		sb.WriteString(fmt.Sprintf("%s %s (%d)\n", tier.marker, tier.name, len(group)))
		for _, f := range group {
			sb.WriteString(fmt.Sprintf("  %s  %s\n", f.ID, f.Message))
			sb.WriteString(fmt.Sprintf("      module: %s\n      hint:   %s\n", f.Module, f.Hint))
		}
		sb.WriteString("\n")
	}

	var healthy []ModuleInfo
	for _, m := range r.Modules {
		flagged := false
		for _, f := range r.Findings {
			if f.Module == m.Path || strings.Contains(f.Module, m.Path) {
				flagged = true
				break
			}
		}
		if !flagged && len(m.Exports) > 0 {
			healthy = append(healthy, m)
		}
	}
	if len(healthy) > 0 {
		sb.WriteString(fmt.Sprintf("🟢 GOOD (%d modules)\n", len(healthy)))
		for i, m := range healthy {
			if i >= 10 {
				sb.WriteString(fmt.Sprintf("  ... and %d more\n", len(healthy)-10))
				break
			}
			sb.WriteString(fmt.Sprintf("  %s  depth %.2f  (%d exports, %d internal edges)\n", m.Path, m.Depth, len(m.Exports), m.InternalEdges))
		}
		sb.WriteString("\n")
	}
	if len(r.Suppressed) > 0 {
		sb.WriteString(fmt.Sprintf("Suppressed by allowlist: %d\n\n", len(r.Suppressed)))
	}
	return sb.String()
}

func (r *StructureReport) GetSummary() string {
	return fmt.Sprintf("%d modules, %d findings (critical %d, warning %d, healthy %d)",
		r.Summary.Modules, r.Summary.Total, r.Summary.Critical, r.Summary.Warning, r.Summary.Healthy)
}
