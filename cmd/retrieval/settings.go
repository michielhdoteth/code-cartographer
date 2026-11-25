package retrieval

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

type Settings struct {
	K1       float64            `json:"k1"`
	B        map[string]float64 `json:"b"`
	Weights  map[string]float64 `json:"weights"`
	Floor    float64            `json:"idf_floor"`
	BodyCap  int                `json:"body_token_cap"`
	QueryCap int                `json:"query_token_cap"`
}

func DefaultSettings() Settings {
	return Settings{
		K1: 1.2, B: DefaultB(), Weights: DefaultWeights(),
		Floor: -2.0, BodyCap: 256, QueryCap: 12,
	}
}

func (s *Settings) FillDefaults() {
	d := DefaultSettings()
	if s.K1 <= 0 {
		s.K1 = d.K1
	}
	if len(s.B) == 0 {
		s.B = d.B
	}
	if len(s.Weights) == 0 {
		s.Weights = d.Weights
	}
	if s.Floor == 0 {
		s.Floor = d.Floor
	}
	if s.BodyCap <= 0 {
		s.BodyCap = d.BodyCap
	}
	if s.QueryCap <= 0 {
		s.QueryCap = d.QueryCap
	}
}

func (s Settings) Hash() string {
	payload, _ := json.Marshal(struct {
		K1, Floor float64
		B, Weights map[string]float64
	}{s.K1, s.Floor, s.B, s.Weights})
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])[:16]
}

func (s Settings) Scorer(stats Stats) Scorer {
	return Scorer{K1: s.K1, B: s.B, Weights: s.Weights, Floor: s.Floor, Stats: stats}
}
