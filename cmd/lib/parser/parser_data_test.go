package parser

import (
	"testing"
)

func assertExtracted(t *testing.T, result *ParseResult, language string) {
	t.Helper()
	if result == nil {
		t.Fatalf("expected non-nil %s parse result", language)
	}
	if len(result.Nodes) == 0 && len(result.Imports) == 0 && len(result.Exports) == 0 {
		t.Fatalf("%s parser extracted no symbols, imports, or exports", language)
	}
}


func TestParseHTML(t *testing.T) {
	code := `<!DOCTYPE html>
<html>
<head>
    <title>My Page</title>
    <link rel="stylesheet" href="style.css">
    <script src="app.js"></script>
</head>
<body>
    <h1>Hello World</h1>
    <div class="content">
        <p>Content here</p>
    </div>
</body>
</html>
`
	parser := NewUnifiedParser(LanguageHTML)
	result, err := parser.Parse(code, "index.html")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result.Nodes) == 0 {
		t.Error("Expected nodes from HTML parse")
	}

	if len(result.Imports) < 1 {
		t.Error("Expected at least 1 import from HTML")
	}
}


func TestParseCSS(t *testing.T) {
	code := `@media (max-width: 768px) {
    .container {
        width: 100%;
    }
}

.header {
    background-color: #333;
    color: white;
}

#main-content {
    padding: 20px;
}

@keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
}
`
	parser := NewUnifiedParser(LanguageCSS)
	result, err := parser.Parse(code, "style.css")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	assertExtracted(t, result, "CSS")
}


func TestParseSQL(t *testing.T) {
	code := `CREATE TABLE users (
    id INT PRIMARY KEY,
    name VARCHAR(100),
    email VARCHAR(255)
);

CREATE INDEX idx_users_email ON users(email);

CREATE VIEW active_users AS
SELECT * FROM users WHERE active = true;

CREATE FUNCTION get_user(id INT)
RETURNS VARCHAR(255)
BEGIN
    RETURN SELECT name FROM users WHERE id = id;
END;
`
	parser := NewUnifiedParser(LanguageSQL)
	result, err := parser.Parse(code, "schema.sql")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	assertExtracted(t, result, "SQL")
}


func TestParseYAML(t *testing.T) {
	code := `name: my-project
version: 1.0.0
dependencies:
  - react: ^18.0.0
  - next: ^14.0.0
scripts:
  dev: next dev
  build: next build
`
	parser := NewUnifiedParser(LanguageYAML)
	result, err := parser.Parse(code, "package.yaml")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	assertExtracted(t, result, "YAML")
}


func TestParseJSON(t *testing.T) {
	code := `{
  "name": "my-project",
  "version": "1.0.0",
  "dependencies": {
    "react": "^18.0.0"
  }
}`
	parser := NewUnifiedParser(LanguageJSON)
	result, err := parser.Parse(code, "package.json")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	assertExtracted(t, result, "JSON")
}


func TestParseTOML(t *testing.T) {
	code := `[package]
name = "my-project"
version = "0.1.0"

[dependencies]
serde = "1.0"
tokio = { version = "1.0", features = ["full"] }

[dev-dependencies]
criterion = "0.5"
`
	parser := NewUnifiedParser(LanguageTOML)
	result, err := parser.Parse(code, "Cargo.toml")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	assertExtracted(t, result, "TOML")
}


func TestParseMarkdown(t *testing.T) {
	code := `# My Project

## Installation

Run the following:

` + "```bash" + `
npm install
` + "```" + `

## Usage

Use it like this.
`
	parser := NewUnifiedParser(LanguageMarkdown)
	result, err := parser.Parse(code, "README.md")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	assertExtracted(t, result, "Markdown")
}

