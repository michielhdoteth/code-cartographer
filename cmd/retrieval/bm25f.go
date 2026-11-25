package retrieval

import (
	"math"
	"sort"
)

// Document is a tokenized retrieval document. Fields are keyed by name,
// qualified, signature, doc, path, kind, and body.
type Document struct {
	ID     string              `json:"id"`
	Fields map[string][]string `json:"fields"`
}

type Stats struct {
	DocCount        int                       `json:"doc_count"`
	DocFreq         map[string]int            `json:"doc_freq"`
	FieldLengths    map[string]map[string]int `json:"field_lengths"`
	AvgFieldLengths map[string]float64        `json:"avg_field_lengths"`
}

func NewStats(docs []Document) Stats {
	s := Stats{
		DocCount:        len(docs),
		DocFreq:         map[string]int{},
		FieldLengths:    map[string]map[string]int{},
		AvgFieldLengths: map[string]float64{},
	}
	totals := map[string]int{}
	for _, doc := range docs {
		seen := map[string]bool{}
		for field, tokens := range doc.Fields {
			if s.FieldLengths[field] == nil {
				s.FieldLengths[field] = map[string]int{}
			}
			s.FieldLengths[field][doc.ID] = len(tokens)
			totals[field] += len(tokens)
			for _, token := range tokens {
				seen[token] = true
			}
		}
		for token := range seen {
			s.DocFreq[token]++
		}
	}
	for field, total := range totals {
		if s.DocCount > 0 {
			s.AvgFieldLengths[field] = float64(total) / float64(s.DocCount)
		}
	}
	return s
}

// IDF is smoothed Robertson/Sparck-Jones IDF.
func IDF(docCount, docFreq float64) float64 {
	if docCount <= 0 {
		return 0
	}
	return math.Log(1 + (docCount-docFreq+0.5)/(docFreq+0.5))
}

type Scorer struct {
	K1      float64
	B       map[string]float64
	Weights map[string]float64
	Floor   float64
	Stats   Stats
}

// DefaultWeights mirrors the field importance used by carto's CLI.
func DefaultWeights() map[string]float64 {
	return map[string]float64{
		"name": 4.0, "qualified": 3.0, "signature": 2.0,
		"doc": 1.5, "path": 2.0, "kind": 0.5, "body": 1.0,
	}
}

func DefaultB() map[string]float64 {
	return map[string]float64{
		"name": 0.3, "qualified": 0.4, "signature": 0.5,
		"doc": 0.5, "path": 0.5, "kind": 0.3, "body": 0.75,
	}
}

func (s Scorer) Score(query []string, doc Document) float64 {
	if len(query) == 0 || s.Stats.DocCount == 0 {
		return 0
	}
	termScores := map[string]float64{}
	for field, tokens := range doc.Fields {
		weight := s.Weights[field]
		if weight == 0 {
			continue
		}
		b := s.B[field]
		dl := float64(len(tokens))
		avdl := s.Stats.AvgFieldLengths[field]
		if avdl <= 0 {
			avdl = 1
		}
		norm := (1-b) + b*dl/avdl
		if norm <= 0 {
			norm = 1
		}
		tf := map[string]int{}
		for _, token := range tokens {
			tf[token]++
		}
		for _, term := range query {
			if tf[term] == 0 {
				continue
			}
			termScores[term] += weight * (float64(tf[term]) / norm)
		}
	}
	total := 0.0
	for _, term := range query {
		tf, ok := termScores[term]
		if !ok {
			continue
		}
		idf := IDF(float64(s.Stats.DocCount), float64(s.Stats.DocFreq[term]))
		if idf < s.Floor {
			idf = s.Floor
		}
		total += idf * ((s.K1 + 1) * tf / (s.K1 + tf))
	}
	return total
}

type Ranked struct {
	ID    string  `json:"id"`
	Score float64 `json:"score"`
}

func (s Scorer) Rank(query []string, docs []Document) []Ranked {
	out := make([]Ranked, 0, len(docs))
	for _, doc := range docs {
		out = append(out, Ranked{ID: doc.ID, Score: s.Score(query, doc)})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		return out[i].ID < out[j].ID
	})
	return out
}
