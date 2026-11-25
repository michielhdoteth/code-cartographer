package parser


type UnifiedParser struct {
	language Language
}

func NewUnifiedParser(lang Language) *UnifiedParser {
	return &UnifiedParser{language: lang}
}

func (p *UnifiedParser) Language() Language {
	return p.language
}

func (p *UnifiedParser) SupportedExtensions() []string {
	return GetLanguageConfig(p.language).Extensions
}

func (p *UnifiedParser) Parse(code string, filePath string) (*ParseResult, error) {
	result := &ParseResult{
		FilePath: filePath,
		Language: p.language,
		Nodes:    []Node{},
		Edges:    []Edge{},
		Exports:  []Export{},
		Imports:  []Import{},
		Metrics:  Metrics{},
	}

	switch p.language {
	case LanguageGo:
		p.parseGo(code, filePath, result)
	case LanguageJavaScript, LanguageTypeScript:
		p.parseJavaScript(code, filePath, result)
	case LanguagePython:
		p.parsePython(code, filePath, result)
	case LanguageJava:
		p.parseJava(code, filePath, result)
	case LanguageC:
		p.parseC(code, filePath, result)
	case LanguageCpp:
		p.parseCpp(code, filePath, result)
	case LanguageCSharp:
		p.parseCSharp(code, filePath, result)
	case LanguageRuby:
		p.parseRuby(code, filePath, result)
	case LanguagePHP:
		p.parsePHP(code, filePath, result)
	case LanguageRust:
		p.parseRust(code, filePath, result)
	case LanguageSwift:
		p.parseSwift(code, filePath, result)
	case LanguageKotlin:
		p.parseKotlin(code, filePath, result)
	case LanguageScala:
		p.parseScala(code, filePath, result)
	case LanguageShell:
		p.parseShell(code, filePath, result)
	case LanguageHTML:
		p.parseHTML(code, filePath, result)
	case LanguageCSS:
		p.parseCSS(code, filePath, result)
	case LanguageSQL:
		p.parseSQL(code, filePath, result)
	case LanguageYAML:
		p.parseYAML(code, filePath, result)
	case LanguageJSON:
		p.parseJSON(code, filePath, result)
	case LanguageTOML:
		p.parseTOML(code, filePath, result)
	case LanguageMarkdown:
		p.parseMarkdown(code, filePath, result)
	default:
		p.parseGeneric(code, filePath, result)
	}

	p.extractImports(code, filePath, result)

	p.extractCalls(code, filePath, result)

	result.CalculateMetrics()

	return result, nil
}

func RegisterUnifiedParsers() {
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
		// Only register if no tree-sitter parser is already registered
		if GetParser(lang) == nil {
			RegisterParser(NewUnifiedParser(lang))
		}
	}
}

func init() {
	// Register tree-sitter parsers first (they take precedence)
	RegisterTreeSitterParsers()
	// Then register regex-based parsers for languages without tree-sitter support
	RegisterUnifiedParsers()
}