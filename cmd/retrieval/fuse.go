package retrieval

// RRF fuses multiple ranked lists using reciprocal rank fusion (k=60).
func RRF(lists [][]string, k int) map[string]float64 {
	if k <= 0 {
		k = 60
	}
	out := map[string]float64{}
	for _, list := range lists {
		for rank, id := range list {
			out[id] += 1 / float64(k+rank+1)
		}
	}
	return out
}

func WeightedRRF(lists [][]string, weights []float64, k int) map[string]float64 {
	if k <= 0 {
		k = 60
	}
	out := map[string]float64{}
	for i, list := range lists {
		weight := 1.0
		if i < len(weights) {
			weight = weights[i]
		}
		for rank, id := range list {
			out[id] += weight / float64(k+rank+1)
		}
	}
	return out
}

// AdaptiveDepth returns candidate depth, doubling from 200 up to 2000 until
// enough candidates could be collected. limit is the number of final hits.
func AdaptiveDepth(limit int) int {
	depth := 200
	target := 5 * limit
	if target < 50 {
		target = 50
	}
	for depth < 2000 && depth < target {
		depth *= 2
	}
	if depth > 2000 {
		return 2000
	}
	return depth
}
