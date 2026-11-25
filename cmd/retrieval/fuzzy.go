package retrieval

// FuzzyScore implements a small Smith-Waterman-style subsequence score for
// identifier queries. It rewards starts of words, consecutive matches, and the
// first character; gaps are penalised. Zero means no subsequence match.
func FuzzyScore(query, candidate string) float64 {
	q := []rune(query)
	c := []rune(candidate)
	if len(q) == 0 || len(q) > len(c) {
		return 0
	}
	score := 0.0
	prev := -1
	for i, qr := range q {
		found := -1
		for j := prev + 1; j < len(c); j++ {
			if lowerRune(c[j]) == lowerRune(qr) {
				found = j
				break
			}
		}
		if found < 0 {
			return 0
		}
		score += 16
		if isWordStart(c, found) {
			score += 8
		}
		if found == prev+1 {
			score += 4
		} else if prev >= 0 {
			gap := found - prev - 1
			score -= 3 + float64(gap-1)
		}
		if i == 0 {
			score *= 2
		}
		prev = found
	}
	return score
}

func lowerRune(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r + ('a' - 'A')
	}
	return r
}

func isWordStart(s []rune, i int) bool {
	if i == 0 {
		return true
	}
	prev := s[i-1]
	if prev == '_' || prev == '-' || prev == '.' || prev == '/' {
		return true
	}
	return prev >= 'a' && prev <= 'z' && s[i] >= 'A' && s[i] <= 'Z'
}
