package parser

import (
	"testing"
)

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		filePath string
		expected Language
	}{
		{"main.go", LanguageGo},
		{"app.js", LanguageJavaScript},
		{"app.jsx", LanguageJavaScript},
		{"app.ts", LanguageTypeScript},
		{"app.tsx", LanguageTypeScript},
		{"main.py", LanguagePython},
		{"Main.java", LanguageJava},
		{"main.c", LanguageC},
		{"main.cpp", LanguageCpp},
		{"main.cc", LanguageCpp},
		{"app.cs", LanguageCSharp},
		{"app.rb", LanguageRuby},
		{"app.php", LanguagePHP},
		{"main.rs", LanguageRust},
		{"main.swift", LanguageSwift},
		{"Main.kt", LanguageKotlin},
		{"Main.scala", LanguageScala},
		{"index.html", LanguageHTML},
		{"style.css", LanguageCSS},
		{"query.sql", LanguageSQL},
		{"script.sh", LanguageShell},
		{"config.yaml", LanguageYAML},
		{"config.yml", LanguageYAML},
		{"data.json", LanguageJSON},
		{"config.toml", LanguageTOML},
		{"README.md", LanguageMarkdown},
		{"Makefile", LanguageShell},
		{"Dockerfile", LanguageShell},
		{"unknown.xyz", LanguageUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			result := DetectLanguage(tt.filePath)
			if result != tt.expected {
				t.Errorf("DetectLanguage(%q) = %q, want %q", tt.filePath, result, tt.expected)
			}
		})
	}
}

func TestIsInterpreted(t *testing.T) {
	interpretedLanguages := []Language{
		LanguageJavaScript,
		LanguageTypeScript,
		LanguagePython,
		LanguageRuby,
		LanguagePHP,
		LanguageShell,
		LanguageHTML,
		LanguageCSS,
		LanguageSQL,
		LanguageYAML,
		LanguageJSON,
		LanguageTOML,
		LanguageMarkdown,
	}

	compiledLanguages := []Language{
		LanguageGo,
		LanguageJava,
		LanguageC,
		LanguageCpp,
		LanguageCSharp,
		LanguageRust,
		LanguageSwift,
		LanguageKotlin,
		LanguageScala,
	}

	for _, lang := range interpretedLanguages {
		if !IsInterpreted(lang) {
			t.Errorf("IsInterpreted(%q) = false, want true", lang)
		}
	}

	for _, lang := range compiledLanguages {
		if IsInterpreted(lang) {
			t.Errorf("IsInterpreted(%q) = true, want false", lang)
		}
	}
}

func TestGetAllLanguages(t *testing.T) {
	langs := GetAllLanguages()
	if len(langs) < 21 {
		t.Errorf("GetAllLanguages() returned %d languages, want at least 21", len(langs))
	}
}

func TestGetLanguageConfig(t *testing.T) {
	config := GetLanguageConfig(LanguageGo)
	if config.Name != "Go" {
		t.Errorf("GetLanguageConfig(LanguageGo).Name = %q, want %q", config.Name, "Go")
	}
	if len(config.Extensions) == 0 {
		t.Error("GetLanguageConfig(LanguageGo).Extensions is empty")
	}
}
