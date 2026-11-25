package doctor

import (
	"fmt"
	"os"
	"path/filepath"

	"carto/codegraph"
	"carto/config"
	"carto/git"
	"carto/retrieval"
	"carto/skill"
)

type Check struct {
	Name   string `json:"name"`
	Status string `json:"status"` // "ok", "warn", "error"
	Detail string `json:"detail"`
}

type Report struct {
	SchemaVersion int     `json:"schema_version"`
	RootPath      string  `json:"root_path"`
	Status        string  `json:"status"`
	Checks        []Check `json:"checks"`
	Files         int     `json:"files_in_graph"`
	Symbols       int     `json:"symbols_in_graph"`
}

func add(report *Report, name, status, detail string) {
	report.Checks = append(report.Checks, Check{Name: name, Status: status, Detail: detail})
}

func Run(root string) *Report {
	abs, err := filepath.Abs(root)
	if err != nil {
		abs = root
	}
	report := &Report{SchemaVersion: 1, RootPath: abs, Checks: []Check{}, Status: "ok"}

	for _, problem := range config.Validate(abs) {
		add(report, "config", "error", problem)
	}
	if len(report.Checks) == 0 {
		add(report, "config", "ok", "no config problems")
	}

	cacheDir := filepath.Join(abs, ".carto")
	if info, err := os.Stat(cacheDir); err != nil || !info.IsDir() {
		add(report, "cache", "error", "no .carto directory; run 'carto map <path>'")
	} else {
		graphPath := filepath.Join(cacheDir, "graph.json")
		if info, err := os.Stat(graphPath); err != nil {
			add(report, "cache", "error", "no graph cache; run 'carto map <path>'")
		} else {
			add(report, "cache", "ok", fmt.Sprintf("graph cache present (%d bytes)", info.Size()))
		}
	}

	freshness, err := codegraph.CheckFreshness(abs)
	switch {
	case err != nil:
		add(report, "freshness", "error", err.Error())
	case !freshness.Fresh:
		add(report, "freshness", "warn", fmt.Sprintf("stale graph: %d added, %d changed, %d deleted; run 'carto map'",
			len(freshness.Added), len(freshness.Changed), len(freshness.Deleted)))
	default:
		add(report, "freshness", "ok", "graph matches working tree")
	}

	index, err := codegraph.Load(abs)
	if err != nil {
		add(report, "parser_coverage", "error", err.Error())
	} else {
		report.Files = len(index.Files)
		parsed := 0
		for _, file := range index.Files {
			if len(file.Nodes) > 0 {
				parsed++
			}
		}
		for _, node := range index.Nodes {
			if node.Kind != "file" {
				report.Symbols++
			}
		}
		if len(index.Files) == 0 {
			add(report, "parser_coverage", "warn", "graph is empty")
		} else if parsed == 0 {
			add(report, "parser_coverage", "warn", "no files produced symbols; check language support")
		} else {
			add(report, "parser_coverage", "ok", fmt.Sprintf("%d/%d files produced symbols", parsed, len(index.Files)))
		}
		for _, warning := range index.Warnings {
			add(report, "parser_coverage", "warn", warning.File+": "+warning.Message)
		}
	}

	if index != nil {
		fileHashes := make(map[string]string, len(index.Files))
		for path, file := range index.Files {
			fileHashes[path] = file.Hash
		}
		graphHash := retrieval.ComputeGraphHash(fileHashes)
		settings := retrieval.DefaultSettings()
		if cfg, err := config.Load(abs); err == nil {
			settings = cfg.Retrieval
		}
		settings.FillDefaults()
		cache, err := retrieval.LoadCache(abs, graphHash, settings.Hash())
		switch {
		case err != nil:
			add(report, "retrieval_cache", "warn", "not built for current graph/settings; it will rebuild on first ask")
		case cache != nil:
			add(report, "retrieval_cache", "ok", fmt.Sprintf("%d tokenized documents", len(cache.Documents)))
		}
	}

	if git.IsRepo(abs) {
		statuses, err := git.GetStatus(abs)
		if err == nil && len(statuses) > 0 {
			add(report, "git", "warn", fmt.Sprintf("working tree has %d uncommitted file(s)", len(statuses)))
		} else {
			add(report, "git", "ok", "working tree clean or status unavailable")
		}
	} else {
		add(report, "git", "ok", "not a git repository")
	}

	if problems := skill.Check(abs, "both"); len(problems) > 0 {
		for _, p := range problems {
			add(report, "skill", "warn", p)
		}
	} else {
		add(report, "skill", "ok", "installed skill and AGENTS.md are current")
	}

	for _, c := range report.Checks {
		if c.Status == "error" {
			report.Status = "error"
			break
		}
		if c.Status == "warn" {
			report.Status = "warn"
		}
	}
	return report
}

func (r *Report) String() string {
	out := fmt.Sprintf("carto doctor: %s (%d files, %d symbols)\n", r.Status, r.Files, r.Symbols)
	for _, c := range r.Checks {
		out += fmt.Sprintf("  [%s] %s: %s\n", c.Status, c.Name, c.Detail)
	}
	return out
}
