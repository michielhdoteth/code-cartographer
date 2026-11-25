// Package tuner provides deterministic offline grid search for retrieval weights.
package tuner

import (
	"encoding/json"
	"os"

	"carto/codegraph"
	"carto/config"
)

type Result struct {
	BaselineRecall float64 `json:"baseline_recall"`
	BaselineMRR    float64 `json:"baseline_mrr"`
	BestRecall     float64 `json:"best_recall"`
	BestMRR        float64 `json:"best_mrr"`
	MeanTokens     float64 `json:"mean_tokens"`
	Improved       bool    `json:"improved"`
	Written        string  `json:"written,omitempty"`
	Tried          int     `json:"tried"`
}

func cloneConfig(cfg *config.Config) (*config.Config, error) {
	data, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	var out config.Config
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Tune searches a small weight grid and writes the best config only if it
// improves recall on the supplied suite. The original config is restored.
func Tune(root string, suite codegraph.TaskSuite, outPath string) (*Result, error) {
	original, err := config.Load(root)
	if err != nil {
		return nil, err
	}
	_, statErr := os.Stat(config.Path(root))
	hadConfig := statErr == nil
	original, err = cloneConfig(original)
	if err != nil {
		return nil, err
	}
	original.Retrieval.FillDefaults()

	index, err := codegraph.Build(root)
	if err != nil {
		return nil, err
	}

	measure := func(cfg *config.Config) (*codegraph.TaskBenchReport, error) {
		if err := config.Save(root, cfg); err != nil {
			return nil, err
		}
		return index.RunTaskBenchmark(suite, 4000)
	}

	baseReport, err := measure(original)
	if err != nil {
		return nil, err
	}
	result := &Result{
		BaselineRecall: baseReport.MeanRecall,
		BaselineMRR:    baseReport.MeanMRR,
		BestRecall:     baseReport.MeanRecall,
		BestMRR:        baseReport.MeanMRR,
		MeanTokens:     baseReport.MeanTokens,
	}
	best, err := cloneConfig(original)
	if err != nil {
		return nil, err
	}

	bodyWeights := []float64{1.0, 1.5, 2.0}
	k1Values := []float64{1.0, 1.2, 1.4}
	floors := []float64{-2.0, -1.0}
	for _, bodyWeight := range bodyWeights {
		for _, k1 := range k1Values {
			for _, floor := range floors {
			candidate, err := cloneConfig(original)
			if err != nil {
				continue
			}
			candidate.Retrieval.Weights["body"] = bodyWeight
				candidate.Retrieval.Weights["doc"] = 1.5 + (bodyWeight-1)*0.5
				candidate.Retrieval.K1 = k1
				candidate.Retrieval.Floor = floor
				report, err := measure(candidate)
				if err != nil {
					continue
				}
				result.Tried++
				if report.MeanRecall > result.BestRecall ||
					(report.MeanRecall == result.BestRecall && report.MeanMRR > result.BestMRR) {
					result.BestRecall = report.MeanRecall
					result.BestMRR = report.MeanMRR
					result.MeanTokens = report.MeanTokens
					best = candidate
				}
			}
		}
	}

	if hadConfig {
		if err := config.Save(root, original); err != nil {
			return nil, err
		}
	} else {
		if err := os.Remove(config.Path(root)); err != nil && !os.IsNotExist(err) {
			return nil, err
		}
	}

	if result.BestRecall > result.BaselineRecall {
		if outPath == "" {
			outPath = config.Path(root)
		}
		if err := config.Save(root, best); err != nil {
			return nil, err
		}
		if err := os.WriteFile(outPath, mustJSON(best), 0644); err != nil && outPath != config.Path(root) {
			return nil, err
		}
		result.Improved = true
		result.Written = outPath
	}
	return result, nil
}

func mustJSON(cfg *config.Config) []byte {
	data, _ := json.MarshalIndent(cfg, "", "  ")
	return append(data, '\n')
}
