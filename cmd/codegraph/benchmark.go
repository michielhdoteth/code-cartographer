package codegraph

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

// BenchmarkSuite is a provider-free, offline set of code retrieval questions.
type BenchmarkSuite struct {
	Version int             `json:"version"`
	Cases   []BenchmarkCase `json:"cases"`
}

type BenchmarkCase struct {
	ID       string   `json:"id"`
	Query    string   `json:"query"`
	Expected []string `json:"expected"`
}

type BenchmarkCaseResult struct {
	ID                 string  `json:"id"`
	Query              string  `json:"query"`
	Precision          float64 `json:"precision_at_limit"`
	Recall             float64 `json:"recall_at_limit"`
	TokensUsedEstimate int     `json:"tokens_used_estimate"`
	LatencyMillis      int64   `json:"latency_ms"`
	Deterministic      bool    `json:"deterministic"`
	Relevant           int     `json:"relevant"`
	Returned           int     `json:"returned"`
}

type BenchmarkReport struct {
	SuiteVersion  int                   `json:"suite_version"`
	Limit         int                   `json:"limit"`
	Cases         []BenchmarkCaseResult `json:"cases"`
	MeanPrecision float64               `json:"mean_precision_at_limit"`
	MeanRecall    float64               `json:"mean_recall_at_limit"`
	MeanTokens    int                   `json:"mean_tokens_used_estimate"`
	MeanLatency   int64                 `json:"mean_latency_ms"`
	Deterministic bool                  `json:"deterministic"`
}

// RunBenchmark evaluates Ask results against expected symbol names, file paths,
// or stable symbol IDs. It does not invoke an LLM or access the network.
func (index *Index) RunBenchmark(suite BenchmarkSuite, limit int) (BenchmarkReport, error) {
	if suite.Version != 1 {
		return BenchmarkReport{}, fmt.Errorf("unsupported benchmark suite version %d", suite.Version)
	}
	if len(suite.Cases) == 0 {
		return BenchmarkReport{}, fmt.Errorf("benchmark suite has no cases")
	}
	if limit <= 0 {
		limit = 8
	}
	report := BenchmarkReport{SuiteVersion: suite.Version, Limit: limit, Cases: make([]BenchmarkCaseResult, 0, len(suite.Cases)), Deterministic: true}
	var precisionTotal, recallTotal float64
	var tokenTotal int
	var latencyTotal int64
	for _, testCase := range suite.Cases {
		if strings.TrimSpace(testCase.ID) == "" || strings.TrimSpace(testCase.Query) == "" || len(testCase.Expected) == 0 {
			return BenchmarkReport{}, fmt.Errorf("each benchmark case requires id, query, and expected values")
		}
		start := time.Now()
		hits := index.Ask(testCase.Query, limit, 1<<30)
		elapsed := time.Since(start).Milliseconds()
		again := index.Ask(testCase.Query, limit, 1<<30)
		deterministic := sameHitOrder(hits, again)
		expected := make(map[string]bool, len(testCase.Expected))
		for _, value := range testCase.Expected {
			expected[strings.ToLower(strings.TrimSpace(filepathSlash(value)))] = true
		}
		relevant := 0
		matchedExpected := make(map[string]bool)
		tokens := 0
		for _, hit := range hits {
			keys := []string{strings.ToLower(hit.Node.Name), strings.ToLower(filepathSlash(hit.Node.File)), strings.ToLower(hit.Node.ID)}
			matched := false
			for _, key := range keys {
				if expected[key] {
					matched = true
					matchedExpected[key] = true
					break
				}
			}
			if matched {
				relevant++
			}
			tokens += estimatedTokens(fmt.Sprintf("%s %s:%d\n%s\n", hit.Node.Name, hit.Node.File, hit.Node.StartLine, hit.Snippet))
		}
		precision, recall := 0.0, 0.0
		if len(hits) > 0 {
			precision = float64(relevant) / float64(len(hits))
		}
		if len(expected) > 0 {
			recall = float64(len(matchedExpected)) / float64(len(expected))
		}
		caseResult := BenchmarkCaseResult{ID: testCase.ID, Query: testCase.Query, Precision: precision, Recall: recall, TokensUsedEstimate: tokens, LatencyMillis: elapsed, Deterministic: deterministic, Relevant: relevant, Returned: len(hits)}
		report.Cases = append(report.Cases, caseResult)
		precisionTotal += precision
		recallTotal += recall
		tokenTotal += tokens
		latencyTotal += elapsed
		if !deterministic {
			report.Deterministic = false
		}
	}
	count := float64(len(report.Cases))
	report.MeanPrecision = precisionTotal / count
	report.MeanRecall = recallTotal / count
	report.MeanTokens = tokenTotal / len(report.Cases)
	report.MeanLatency = latencyTotal / int64(len(report.Cases))
	return report, nil
}

func sameHitOrder(left, right []QueryHit) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if !reflect.DeepEqual(left[i], right[i]) {
			return false
		}
	}
	return true
}

func filepathSlash(value string) string { return strings.ReplaceAll(value, "\\", "/") }

func (report BenchmarkReport) RenderMarkdown() string {
	var b strings.Builder
	b.WriteString("# Code Cartographer offline benchmark\n\n")
	b.WriteString(fmt.Sprintf("Cases: %d · limit: %d · deterministic: %t\n\n", len(report.Cases), report.Limit, report.Deterministic))
	b.WriteString("| Case | Precision | Recall | Est. tokens | Latency ms | Deterministic |\n| --- | ---: | ---: | ---: | ---: | :---: |\n")
	for _, result := range report.Cases {
		b.WriteString(fmt.Sprintf("| %s | %.2f | %.2f | %d | %d | %t |\n", result.ID, result.Precision, result.Recall, result.TokensUsedEstimate, result.LatencyMillis, result.Deterministic))
	}
	b.WriteString(fmt.Sprintf("\nMean precision: %.2f · mean recall: %.2f · mean estimated tokens: %d · mean latency: %d ms\n", report.MeanPrecision, report.MeanRecall, report.MeanTokens, report.MeanLatency))
	return b.String()
}
