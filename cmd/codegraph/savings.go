package codegraph

// Savings compares an agent-facing payload with a naive whole-repository dump.
// Values are provider-neutral estimates, not tokenizer-exact counts.
type Savings struct {
	WholeRepoTokens       int     `json:"whole_repo_tokens_estimate"`
	UsedTokens            int     `json:"tokens_used_estimate"`
	SavedTokens           int     `json:"saved_tokens_estimate"`
	SavedPercent          float64 `json:"saved_percent_estimate"`
	Basis                 string  `json:"basis"`

	BaselineKind           string  `json:"baseline_kind,omitempty"`
	BaselineLines          int     `json:"baseline_lines,omitempty"`
	EmittedLines           int     `json:"emitted_lines,omitempty"`
	AvoidedLines           int     `json:"avoided_lines,omitempty"`
	AvoidedTokens          int     `json:"avoided_tokens,omitempty"`
	ContextEfficiencyScore float64 `json:"context_efficiency_score,omitempty"`
}

const savingsBasis = "estimated ~4 UTF-8 bytes per token; baseline = indexed source bytes / 4"

func BuildSavings(usedTokens, emittedLines, baselineTokens, baselineLines int, baselineKind string) *Savings {
	if baselineTokens <= 0 {
		return nil
	}
	saved := baselineTokens - usedTokens
	if saved < 0 {
		saved = 0
	}
	avoidedLines := baselineLines - emittedLines
	if avoidedLines < 0 {
		avoidedLines = 0
	}
	return &Savings{
		WholeRepoTokens:        baselineTokens,
		UsedTokens:             usedTokens,
		SavedTokens:            saved,
		SavedPercent:           float64(saved) / float64(baselineTokens) * 100,
		BaselineKind:           baselineKind,
		BaselineLines:          baselineLines,
		EmittedLines:           emittedLines,
		AvoidedLines:           avoidedLines,
		AvoidedTokens:          saved,
		ContextEfficiencyScore: float64(saved) / float64(baselineTokens) * 100,
		Basis:                  savingsBasis,
	}
}

func (index *Index) EstimateSavings(used int) *Savings {
	if index == nil {
		return nil
	}
	bytesTotal := int64(0)
	linesTotal := 0
	for _, file := range index.Files {
		bytesTotal += file.Size
		linesTotal += file.Lines
	}
	baseline := int(bytesTotal / 4)
	saved := baseline - used
	if saved < 0 {
		saved = 0
	}
	return &Savings{
		WholeRepoTokens: baseline,
		UsedTokens:      used,
		SavedTokens:     saved,
		SavedPercent:    float64(saved) / float64(baseline) * 100,
		Basis:           savingsBasis,
	}
}

func (index *Index) EstimateSavingsWithLines(usedTokens, usedLines int, rel string) *Savings {
	if index == nil {
		return nil
	}
	file, ok := index.Files[rel]
	if !ok {
		return index.EstimateSavings(usedTokens)
	}
	return BuildSavings(usedTokens, usedLines, int(file.Size/4), file.Lines, "file")
}
