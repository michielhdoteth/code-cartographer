//go:build !cgo
// +build !cgo

package parser

// TreeSitterParser is a stub when CGO is not available
type TreeSitterParser struct {
	language Language
}

// NewTreeSitterParser returns nil when CGO is not available
func NewTreeSitterParser(lang Language) *TreeSitterParser {
	return nil
}

// Language returns the language this parser handles
func (p *TreeSitterParser) Language() Language {
	if p == nil {
		return LanguageUnknown
	}
	return p.language
}

// SupportedExtensions returns file extensions this parser handles
func (p *TreeSitterParser) SupportedExtensions() []string {
	if p == nil {
		return nil
	}
	return GetLanguageConfig(p.language).Extensions
}

// Parse is a stub that returns an error when CGO is not available
func (p *TreeSitterParser) Parse(code string, filePath string) (*ParseResult, error) {
	return nil, nil
}

// RegisterTreeSitterParsers is a no-op when CGO is not available
func RegisterTreeSitterParsers() {
	// Tree-sitter requires CGO; fall back to regex-based parsers
}
