package codegraph

import (
	"fmt"
	"path/filepath"
	"reflect"
	"sort"
)

// TaskSuite is an offline suite of agent-style questions with gold
// symbols/files. No LLM is involved; results measure retrieval recall and the
// token cost of the context an agent would read.
type TaskSuite struct {
	Version int         `json:"version"`
	Cases   []Task `json:"cases"`
}

type Task struct {
	ID              string   `json:"id"`
	Query           string   `json:"query"`
	Expected        []string `json:"expected"`
	QueryType       string   `json:"query_type,omitempty"`
	ExpectedSymbols []string `json:"expected_symbols,omitempty"`
	ExpectedFiles   []string `json:"expected_files,omitempty"`
	Difficulty      string   `json:"difficulty,omitempty"`
	Split           string   `json:"split,omitempty"`
	Tokens          int      `json:"tokens,omitempty"`
}

// Gold returns the union of v1 expected values and v2 symbol/file labels.
func (t Task) Gold() []string {
	seen := map[string]bool{}
	var out []string
	for _, list := range [][]string{t.Expected, t.ExpectedSymbols, t.ExpectedFiles} {
		for _, value := range list {
			if value != "" && !seen[value] {
				seen[value] = true
				out = append(out, value)
			}
		}
	}
	return out
}

func (index *Index) fullFileTokensForTask(task Task) int {
	files := make(map[string]bool)
	for _, path := range task.ExpectedFiles {
		files[filepath.ToSlash(path)] = true
	}
	for _, value := range task.Gold() {
		path := filepath.ToSlash(value)
		if _, ok := index.Files[path]; ok {
			files[path] = true
		}
	}
	total := 0
	for path := range files {
		if file, ok := index.Files[path]; ok {
			total += int(file.Size / 4)
		}
	}
	return total
}

type TaskResult struct {
	ID                     string  `json:"id"`
	Query                  string  `json:"query"`
	QueryType              string  `json:"query_type,omitempty"`
	RecallAt1              float64 `json:"recall_at_1"`
	RecallAt5              float64 `json:"recall_at_5"`
	RecallAt10             float64 `json:"recall_at_10"`
	TokensUsedEstimate     int     `json:"tokens_used_estimate"`
	FullFileTokensEstimate int     `json:"full_file_tokens_estimate"`
	WholeRepoTokensEstimate int    `json:"whole_repo_tokens_estimate"`
	SavedTokensEstimate    int     `json:"saved_tokens_estimate"`
	SavedPercentEstimate   float64 `json:"saved_percent_estimate"`
	Recall                 float64 `json:"recall"`
	MRR                    float64 `json:"mrr"`
	Relevant               int     `json:"relevant"`
	Returned               int     `json:"returned"`
	Strategy               string  `json:"strategy,omitempty"`
	TokensPerCorrectSymbol float64 `json:"tokens_per_correct_symbol"`
}

type TaskBenchReport struct {
	SchemaVersion          int               `json:"schema_version"`
	SuiteVersion           int               `json:"suite_version"`
	Cases                  []TaskResult `json:"cases"`
	MeanTokens             float64           `json:"mean_tokens"`
	MeanFullFileTokens     float64           `json:"mean_full_file_tokens"`
	WholeRepoTokensEstimate int               `json:"whole_repo_tokens_estimate"`
	MeanSavedPercent       float64           `json:"mean_saved_percent"`
	MeanRecall             float64           `json:"mean_recall"`
	MeanMRR                float64           `json:"mean_mrr"`
	Strategy               string            `json:"strategy"`
	MeanTokensPerCorrect   float64           `json:"mean_tokens_per_correct_symbol"`
	Deterministic          bool              `json:"deterministic"`
}

func (index *Index) RunTaskBenchmark(suite TaskSuite, defaultBudget int) (*TaskBenchReport, error) {
	return index.RunTaskBenchmarkWithStrategy(suite, defaultBudget, "ask")
}

// RunTaskBenchmarkWithStrategy supports the consult workflow in addition to
// plain retrieval. Both strategies remain offline and deterministic.
func (index *Index) RunTaskBenchmarkWithStrategy(suite TaskSuite, defaultBudget int, strategy string) (*TaskBenchReport, error) {
	if strategy == "" {
		strategy = "ask"
	}
	report, err := index.runTaskBenchmark(suite, defaultBudget, strategy)
	if err != nil {
		return nil, err
	}
	again, err := index.runTaskBenchmark(suite, defaultBudget, strategy)
	if err != nil || !reflect.DeepEqual(report, again) {
		report.Deterministic = false
	}
	return report, nil
}

func (index *Index) runTaskBenchmark(suite TaskSuite, defaultBudget int, strategy string) (*TaskBenchReport, error) {
	if suite.Version != 1 && suite.Version != 2 {
		return nil, fmt.Errorf("unsupported task suite version %d", suite.Version)
	}
	if len(suite.Cases) == 0 {
		return nil, fmt.Errorf("task suite has no cases")
	}
	if defaultBudget <= 0 {
		defaultBudget = 4000
	}
	report := &TaskBenchReport{
		SchemaVersion:  1,
		SuiteVersion:   suite.Version,
		Cases:          []TaskResult{},
		Strategy:       strategy,
		Deterministic:  true,
	}
	baseline := 0
	if savings := index.EstimateSavings(0); savings != nil {
		baseline = savings.WholeRepoTokens
	}
	report.WholeRepoTokensEstimate = baseline

	for _, task := range suite.Cases {
		budget := task.Tokens
		if budget <= 0 {
			budget = defaultBudget
		}
		hits := index.Ask(task.Query, 12, budget)
		matched := 0
		expected := make(map[string]bool)
		for _, value := range task.Gold() {
			expected[value] = true
		}
		keys := make(map[string]bool)
		firstRank := 0
		hitSets := map[int]map[string]bool{1: {}, 5: {}, 10: {}}
		for i, hit := range hits {
			keys[hit.Node.Name] = true
			keys[hit.Node.File] = true
			if i < 10 {
				hitSets[10][hit.Node.Name] = true
				hitSets[10][hit.Node.File] = true
			}
			if i < 5 {
				hitSets[5][hit.Node.Name] = true
				hitSets[5][hit.Node.File] = true
			}
			if i < 1 {
				hitSets[1][hit.Node.Name] = true
				hitSets[1][hit.Node.File] = true
			}
			if firstRank == 0 && (expected[hit.Node.Name] || expected[hit.Node.File]) {
				firstRank = i + 1
			}
		}
		for key := range expected {
			if keys[key] {
				matched++
			}
		}
		recallAt := func(k int) float64 {
			if len(expected) == 0 {
				return 0
			}
			hitsAtK := 0
			for key := range expected {
				if hitSets[k][key] {
					hitsAtK++
				}
			}
			return float64(hitsAtK) / float64(len(expected))
		}
		mrr := 0.0
		if firstRank > 0 {
			mrr = 1 / float64(firstRank)
		}
		used := 0
		if strategy == "consult" {
			if len(hits) > 0 {
				topFile := hits[0].Node.File
				skeleton := index.GenerateSkeleton(topFile, 512)
				used += skeleton.TokensUsed
				if manifest, err := index.BuildChunkManifest(topFile, 400); err == nil && len(manifest.Chunks) > 0 {
					used += manifest.Chunks[0].Tokens
				}
			}
		} else {
			for _, hit := range hits {
				used += EstimateTokens(hit.Snippet)
			}
		}
		fullFileTokens := index.fullFileTokensForTask(task)
		recall := 0.0
		if len(expected) > 0 {
			recall = float64(matched) / float64(len(expected))
		}
		saved := baseline - used
		if saved < 0 {
			saved = 0
		}
		savedPct := 0.0
		if baseline > 0 {
			savedPct = float64(saved) / float64(baseline) * 100
		}
		report.Cases = append(report.Cases, TaskResult{
			ID: task.ID, Query: task.Query, QueryType: task.QueryType,
			RecallAt1: recallAt(1), RecallAt5: recallAt(5), RecallAt10: recallAt(10),
			TokensUsedEstimate: used,
			FullFileTokensEstimate: fullFileTokens,
			WholeRepoTokensEstimate: baseline,
			SavedTokensEstimate: saved,
			SavedPercentEstimate: savedPct,
			Recall: recall, MRR: mrr, Relevant: matched, Returned: len(hits),
			Strategy: strategy,
			TokensPerCorrectSymbol: float64(used) / float64(maxInt(matched, 1)),
		})
	}

	sort.Slice(report.Cases, func(i, j int) bool {
		if report.Cases[i].Recall != report.Cases[j].Recall {
			return report.Cases[i].Recall > report.Cases[j].Recall
		}
		return report.Cases[i].ID < report.Cases[j].ID
	})

	totalTokens, totalRecall, totalSaved, totalMRR, totalFullFile := 0.0, 0.0, 0.0, 0.0, 0.0
	totalCorrect := 0
	for _, c := range report.Cases {
		totalTokens += float64(c.TokensUsedEstimate)
		totalRecall += c.Recall
		totalSaved += c.SavedPercentEstimate
		totalMRR += c.MRR
		totalCorrect += c.Relevant
		totalFullFile += float64(c.FullFileTokensEstimate)
	}
	n := float64(len(report.Cases))
	report.MeanTokens = totalTokens / n
	report.MeanFullFileTokens = totalFullFile / n
	report.MeanRecall = totalRecall / n
	report.MeanMRR = totalMRR / n
	report.MeanSavedPercent = totalSaved / n
	if totalCorrect > 0 {
		report.MeanTokensPerCorrect = totalTokens / float64(totalCorrect)
	} else {
		report.MeanTokensPerCorrect = totalTokens
	}

	return report, nil
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (r *TaskBenchReport) RenderMarkdown() string {
	out := "# Agent task benchmark\n\n"
	out += fmt.Sprintf("Strategy: %s · Mean recall: %.2f · Mean MRR: %.2f · Mean tokens: %.0f · Full-file baseline: %.0f · Tokens/correct: %.0f · Baseline whole repo: %d · Mean savings: %.1f%%\n\n",
		r.Strategy, r.MeanRecall, r.MeanMRR, r.MeanTokens, r.MeanFullFileTokens, r.MeanTokensPerCorrect, r.WholeRepoTokensEstimate, r.MeanSavedPercent)
	out += "| Task | Recall | MRR | Tokens | Saved |\n|---|---|---|---|---|\n"
	for _, c := range r.Cases {
		out += fmt.Sprintf("| %s | %.2f | %.2f | %d | %.1f%% |\n", c.ID, c.Recall, c.MRR, c.TokensUsedEstimate, c.SavedPercentEstimate)
	}
	out += fmt.Sprintf("\nDeterministic: %t\n", r.Deterministic)
	return out
}







