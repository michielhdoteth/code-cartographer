// Package retrieval provides deterministic, dependency-free lexical ranking for
// code search. It is a leaf package: it never imports other carto packages.
package retrieval

import (
	"strings"
	"unicode"
)

const TokenizerVersion = 2

var codeStopwords = map[string]bool{
	"a": true, "an": true, "the": true, "and": true, "or": true, "to": true,
	"of": true, "in": true, "for": true, "on": true, "is": true, "are": true,
	"be": true, "as": true, "at": true, "by": true, "with": true, "from": true,
	"func": true, "function": true, "return": true, "type": true, "struct": true,
	"interface": true, "package": true, "import": true, "var": true, "const": true,
	"let": true, "def": true, "class": true, "this": true, "self": true,
	"nil": true, "true": true, "false": true, "ctx": true, "args": true,
	"val": true, "new": true, "make": true,
	"where": true, "what": true, "which": true, "when": true, "why": true, "how": true,
	"does": true, "do": true, "did": true, "was": true, "were": true, "can": true,
	"could": true, "should": true, "would": true, "will": true, "has": true,
	"have": true, "had": true, "that": true, "these": true,
	"those": true, "there": true, "here": true, "into": true, "about": true,
	"than": true, "then": true, "each": true, "some": true, "any": true,
	"all": true, "every": true, "both": true, "such": true, "same": true,
	"other": true, "another": true, "using": true, "used": true, "use": true,
	"via": true, "within": true, "across": true, "between": true, "find": true,
	"show": true, "tell": true, "explain": true, "many": true,
	"much": true, "most": true, "only": true, "just": true, "also": true,
}

func IsStopword(token string) bool { return codeStopwords[token] }

// SplitIdentifier splits one identifier span into meaningful subtokens.
// HTTPServer -> [http server], AnalyzeBlastRadius -> [analyze blast radius],
// sha256_hash -> [sha 256 hash].
func SplitIdentifier(span string) []string {
	if span == "" {
		return nil
	}
	runes := []rune(span)
	var words []string
	start := 0
	for i := 1; i < len(runes); i++ {
		prev, curr := runes[i-1], runes[i]
		boundary := false
		switch {
		case prev == '_' || curr == '_':
			boundary = true
		case unicode.IsLower(prev) && unicode.IsUpper(curr):
			boundary = true
		case unicode.IsUpper(prev) && unicode.IsUpper(curr) && i+1 < len(runes) && unicode.IsLower(runes[i+1]):
			boundary = true
		case unicode.IsLetter(prev) && unicode.IsDigit(curr), unicode.IsDigit(prev) && unicode.IsLetter(curr):
			boundary = true
		}
		if boundary {
			word := strings.Trim(string(runes[start:i]), "_")
			if word != "" {
				words = append(words, strings.ToLower(word))
			}
			start = i
			if curr == '_' {
				start = i + 1
			}
		}
	}
	last := strings.Trim(string(runes[start:]), "_")
	if last != "" {
		words = append(words, strings.ToLower(last))
	}
	return words
}

// Tokenize returns split tokens for a field, dropping stopwords and 1-char tokens.
func Tokenize(text string) []string {
	var result []string
	for _, span := range strings.FieldsFunc(text, func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_')
	}) {
		for _, token := range SplitIdentifier(span) {
			if len([]rune(token)) > 1 && !IsStopword(token) {
				result = append(result, token)
			}
		}
	}
	return result
}

func RawTokens(text string) []string {
	var result []string
	for _, span := range strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_')
	}) {
		if span != "" {
			result = append(result, span)
		}
	}
	return result
}
