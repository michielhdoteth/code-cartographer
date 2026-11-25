package parser

import (
	"testing"
)

func TestTreeSitterParserAllLanguages(t *testing.T) {
	// Test that tree-sitter parsers can be created for all languages
	// that have grammar packages available
	languages := []Language{
		LanguageGo,
		LanguageJavaScript,
		LanguageTypeScript,
		LanguagePython,
		LanguageJava,
		LanguageC,
		LanguageCpp,
		LanguageCSharp,
		LanguageRuby,
		LanguagePHP,
		LanguageRust,
		LanguageSwift,
		LanguageKotlin,
		LanguageScala,
		LanguageHTML,
		LanguageCSS,
		LanguageSQL,
		LanguageShell,
		LanguageYAML,
		LanguageJSON,
		LanguageTOML,
		LanguageMarkdown,
	}

	for _, lang := range languages {
		t.Run(string(lang), func(t *testing.T) {
			parser := NewTreeSitterParser(lang)
			if parser == nil {
				if GetParser(lang) == nil {
					t.Fatalf("no parser available for %q (tree-sitter and fallback both missing)", lang)
				}
				return
			}
			if parser.Language() != lang {
				t.Errorf("parser.Language() = %q, want %q", parser.Language(), lang)
			}
		})
	}
}

func TestUnifiedParserAllLanguages(t *testing.T) {
	languages := []Language{
		LanguagePython,
		LanguageJava,
		LanguageC,
		LanguageCpp,
		LanguageCSharp,
		LanguageRuby,
		LanguagePHP,
		LanguageRust,
		LanguageSwift,
		LanguageKotlin,
		LanguageScala,
		LanguageShell,
		LanguageHTML,
		LanguageCSS,
		LanguageSQL,
		LanguageYAML,
		LanguageJSON,
		LanguageTOML,
		LanguageMarkdown,
	}

	for _, lang := range languages {
		t.Run(string(lang), func(t *testing.T) {
			parser := NewUnifiedParser(lang)
			if parser == nil {
				t.Fatalf("NewUnifiedParser(%q) returned nil", lang)
			}
			if parser.Language() != lang {
				t.Errorf("parser.Language() = %q, want %q", parser.Language(), lang)
			}
		})
	}
}

func TestRegisterAllParsers(t *testing.T) {
	parserRegistry = make(map[Language]Parser)

	RegisterTreeSitterParsers()
	RegisterUnifiedParsers()

	supported := GetSupportedLanguages()
	if len(supported) < 21 {
		t.Errorf("Expected at least 21 supported languages, got %d: %v", len(supported), supported)
	}

	mustHave := []Language{
		LanguageGo,
		LanguageJavaScript,
		LanguageTypeScript,
		LanguagePython,
		LanguageJava,
		LanguageC,
		LanguageCpp,
		LanguageCSharp,
		LanguageRuby,
		LanguagePHP,
		LanguageRust,
		LanguageSwift,
		LanguageKotlin,
		LanguageScala,
		LanguageHTML,
		LanguageCSS,
		LanguageSQL,
		LanguageShell,
		LanguageYAML,
		LanguageJSON,
		LanguageTOML,
		LanguageMarkdown,
	}

	for _, lang := range mustHave {
		if GetParser(lang) == nil {
			t.Errorf("No parser registered for %q", lang)
		}
	}
}