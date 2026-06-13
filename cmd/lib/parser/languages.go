package parser

import (
	"path/filepath"
	"strings"
)

// Language represents a supported programming language
type Language string

// Supported languages
const (
	LanguageGo          Language = "go"
	LanguageJavaScript  Language = "javascript"
	LanguageTypeScript  Language = "typescript"
	LanguagePython      Language = "python"
	LanguageJava        Language = "java"
	LanguageC           Language = "c"
	LanguageCpp         Language = "cpp"
	LanguageCSharp      Language = "csharp"
	LanguageRuby        Language = "ruby"
	LanguagePHP         Language = "php"
	LanguageRust        Language = "rust"
	LanguageSwift       Language = "swift"
	LanguageKotlin      Language = "kotlin"
	LanguageScala       Language = "scala"
	LanguageHTML        Language = "html"
	LanguageCSS         Language = "css"
	LanguageSQL         Language = "sql"
	LanguageShell       Language = "shell"
	LanguageYAML        Language = "yaml"
	LanguageJSON        Language = "json"
	LanguageTOML        Language = "toml"
	LanguageMarkdown    Language = "markdown"
	LanguageUnknown     Language = "unknown"
)

// LanguageConfig holds configuration for a language
type LanguageConfig struct {
	Language    Language
	Name        string
	Extensions  []string
	TreeSitter  string // tree-sitter language name
	Comments    []string
	LineComment string
}

// LanguageConfigs maps language to its configuration
var LanguageConfigs = map[Language]LanguageConfig{
	LanguageGo: {
		Language:    LanguageGo,
		Name:        "Go",
		Extensions:  []string{".go"},
		TreeSitter:  "go",
		Comments:    []string{"/*", "*/"},
		LineComment: "//",
	},
	LanguageJavaScript: {
		Language:    LanguageJavaScript,
		Name:        "JavaScript",
		Extensions:  []string{".js", ".jsx", ".mjs", ".cjs"},
		TreeSitter:  "javascript",
		Comments:    []string{"/*", "*/"},
		LineComment: "//",
	},
	LanguageTypeScript: {
		Language:    LanguageTypeScript,
		Name:        "TypeScript",
		Extensions:  []string{".ts", ".tsx", ".mts", ".cts"},
		TreeSitter:  "typescript",
		Comments:    []string{"/*", "*/"},
		LineComment: "//",
	},
	LanguagePython: {
		Language:    LanguagePython,
		Name:        "Python",
		Extensions:  []string{".py", ".pyi"},
		TreeSitter:  "python",
		Comments:    []string{`"""`, `'''`},
		LineComment: "#",
	},
	LanguageJava: {
		Language:    LanguageJava,
		Name:        "Java",
		Extensions:  []string{".java"},
		TreeSitter:  "java",
		Comments:    []string{"/*", "*/"},
		LineComment: "//",
	},
	LanguageC: {
		Language:    LanguageC,
		Name:        "C",
		Extensions:  []string{".c", ".h"},
		TreeSitter:  "c",
		Comments:    []string{"/*", "*/"},
		LineComment: "//",
	},
	LanguageCpp: {
		Language:    LanguageCpp,
		Name:        "C++",
		Extensions:  []string{".cpp", ".cc", ".cxx", ".hpp", ".hh", ".hxx"},
		TreeSitter:  "cpp",
		Comments:    []string{"/*", "*/"},
		LineComment: "//",
	},
	LanguageCSharp: {
		Language:    LanguageCSharp,
		Name:        "C#",
		Extensions:  []string{".cs"},
		TreeSitter:  "csharp",
		Comments:    []string{"/*", "*/"},
		LineComment: "//",
	},
	LanguageRuby: {
		Language:    LanguageRuby,
		Name:        "Ruby",
		Extensions:  []string{".rb"},
		TreeSitter:  "ruby",
		Comments:    []string{"=begin", "=end"},
		LineComment: "#",
	},
	LanguagePHP: {
		Language:    LanguagePHP,
		Name:        "PHP",
		Extensions:  []string{".php"},
		TreeSitter:  "php",
		Comments:    []string{"/*", "*/"},
		LineComment: "//",
	},
	LanguageRust: {
		Language:    LanguageRust,
		Name:        "Rust",
		Extensions:  []string{".rs"},
		TreeSitter:  "rust",
		Comments:    []string{"/*", "*/"},
		LineComment: "//",
	},
	LanguageSwift: {
		Language:    LanguageSwift,
		Name:        "Swift",
		Extensions:  []string{".swift"},
		TreeSitter:  "swift",
		Comments:    []string{"/*", "*/"},
		LineComment: "//",
	},
	LanguageKotlin: {
		Language:    LanguageKotlin,
		Name:        "Kotlin",
		Extensions:  []string{".kt", ".kts"},
		TreeSitter:  "kotlin",
		Comments:    []string{"/*", "*/"},
		LineComment: "//",
	},
	LanguageScala: {
		Language:    LanguageScala,
		Name:        "Scala",
		Extensions:  []string{".scala"},
		TreeSitter:  "scala",
		Comments:    []string{"/*", "*/"},
		LineComment: "//",
	},
	LanguageHTML: {
		Language:    LanguageHTML,
		Name:        "HTML",
		Extensions:  []string{".html", ".htm"},
		TreeSitter:  "html",
		Comments:    []string{"<!--", "-->"},
		LineComment: "",
	},
	LanguageCSS: {
		Language:    LanguageCSS,
		Name:        "CSS",
		Extensions:  []string{".css", ".scss", ".sass", ".less"},
		TreeSitter:  "css",
		Comments:    []string{"/*", "*/"},
		LineComment: "",
	},
	LanguageSQL: {
		Language:    LanguageSQL,
		Name:        "SQL",
		Extensions:  []string{".sql"},
		TreeSitter:  "sql",
		Comments:    []string{"/*", "*/"},
		LineComment: "--",
	},
	LanguageShell: {
		Language:    LanguageShell,
		Name:        "Shell",
		Extensions:  []string{".sh", ".bash", ".zsh"},
		TreeSitter:  "bash",
		Comments:    []string{},
		LineComment: "#",
	},
	LanguageYAML: {
		Language:    LanguageYAML,
		Name:        "YAML",
		Extensions:  []string{".yaml", ".yml"},
		TreeSitter:  "yaml",
		Comments:    []string{},
		LineComment: "#",
	},
	LanguageJSON: {
		Language:    LanguageJSON,
		Name:        "JSON",
		Extensions:  []string{".json"},
		TreeSitter:  "json",
		Comments:    []string{},
		LineComment: "",
	},
	LanguageTOML: {
		Language:    LanguageTOML,
		Name:        "TOML",
		Extensions:  []string{".toml"},
		TreeSitter:  "toml",
		Comments:    []string{},
		LineComment: "#",
	},
	LanguageMarkdown: {
		Language:    LanguageMarkdown,
		Name:        "Markdown",
		Extensions:  []string{".md", ".markdown"},
		TreeSitter:  "markdown",
		Comments:    []string{},
		LineComment: "",
	},
}

// DetectLanguage detects the language from a file path
func DetectLanguage(filePath string) Language {
	ext := strings.ToLower(filepath.Ext(filePath))

	// Special cases
	baseName := strings.ToLower(filepath.Base(filePath))
	if baseName == "makefile" || baseName == "makefile.inc" {
		return LanguageShell
	}
	if baseName == "dockerfile" {
		return LanguageShell
	}

	for lang, config := range LanguageConfigs {
		for _, e := range config.Extensions {
			if strings.ToLower(e) == ext {
				return lang
			}
		}
	}

	return LanguageUnknown
}

// GetLanguageConfig returns the configuration for a language
func GetLanguageConfig(lang Language) LanguageConfig {
	if config, ok := LanguageConfigs[lang]; ok {
		return config
	}
	return LanguageConfig{Language: lang, Name: string(lang)}
}

// GetAllLanguages returns all supported languages
func GetAllLanguages() []Language {
	languages := make([]Language, 0, len(LanguageConfigs))
	for lang := range LanguageConfigs {
		languages = append(languages, lang)
	}
	return languages
}

// IsInterpreted returns true if the language is typically interpreted
func IsInterpreted(lang Language) bool {
	interpreted := map[Language]bool{
		LanguageJavaScript: true,
		LanguageTypeScript: true,
		LanguagePython:     true,
		LanguageRuby:       true,
		LanguagePHP:        true,
		LanguageShell:      true,
		LanguageHTML:       true,
		LanguageCSS:        true,
		LanguageSQL:        true,
		LanguageYAML:       true,
		LanguageJSON:       true,
		LanguageTOML:       true,
		LanguageMarkdown:   true,
		LanguageGo:         false, // compiled
		LanguageJava:       false, // compiled
		LanguageC:          false, // compiled
		LanguageCpp:        false, // compiled
		LanguageCSharp:     false, // compiled
		LanguageRust:       false, // compiled
		LanguageSwift:      false, // compiled
		LanguageKotlin:     false, // compiled
		LanguageScala:      false, // compiled
	}
	return interpreted[lang]
}