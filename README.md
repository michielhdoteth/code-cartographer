# Code Cartographer

A fast Go context engine that maps large repos so AI agents know where to look before spending tokens.

## Installation

### CLI (Go)

```bash
cd cmd && go build -o ../bin/carto.exe .
```

### Web UI

```bash
bun install
bun run build
```

## Quick Start

```bash
# Map a codebase
carto map ./my-project

# Analyze for issues
carto analyze

# Search code
carto find "functionName"

# Start web UI
carto serve
```

## CLI Commands

| Command | Description |
|---------|-------------|
| `carto map <path>` | Scan and map a codebase (saves commit manifest) |
| `carto analyze [type]` | Analyze for dead code, complexity, etc. |
| `carto find <pattern>` | Search code in mapped files |
| `carto diff [ref]` | Show changes since last map |
| `carto changed` | Show current uncommitted changes |
| `carto impact <file>` | Analyze impact of changing a file |
| `carto serve` | Start interactive web UI |
| `carto mcp` | Start MCP server for Claude integration |

### Analysis Types

| Type | Description |
|------|-------------|
| `dead-code` | Unused files, exports, and dependencies |
| `complexity` | Functions with high cyclomatic complexity |
| `circular` | Circular dependency detection |
| `health` | Overall code health metrics |
| `all` | Run all analyses (default) |

### Flags

```bash
carto map --lang typescript,python ./src   # Filter by language
carto analyze --type dead-code              # Specific analysis
carto find --regex "func\w+"               # Regex search
carto find --case-sensitive "MyFunc"        # Case-sensitive search
carto analyze --json                        # JSON output
carto serve --port 8080                     # Custom port
```

## Features

- **Fast scanning** - Sub-second for typical projects
- **Multi-language** - TypeScript, JavaScript, Python, Go, Rust, Java, C++, Ruby, PHP
- **Tree-sitter parsing** - Accurate AST analysis for all supported languages
- **Zero config** - Works out of the box
- **CLI-first** - Fast integration in any workflow
- **Web UI** - Interactive visualization with multiple views (force-directed, tree, treemap, matrix, radial, flow)
- **Analysis** - Dead code detection, complexity scoring, circular dependency detection, health metrics
- **MCP server** - Claude Desktop integration via Model Context Protocol
- **JSON output** - Machine-readable results for scripting and CI/CD

## Commit-Aware Mapping

Carto tracks which commit your map was built from, enabling targeted context for AI agents:

- **`carto map`** saves a manifest with commit SHA, branch, and timestamp
- **`carto diff`** shows what changed since your last map
- **`carto changed`** shows uncommitted changes grouped by status
- **`carto impact`** analyzes dependencies and risk for a specific file

This means agents can ask "what changed?" instead of rescanning everything.

## Web UI Components

- **File Explorer** - Browse codebase structure with folder hierarchy
- **Language Distribution** - Pie chart breakdown of languages used
- **Health Score Display** - Overall codebase health visualization
- **Architecture Issues Panel** - Surface detected patterns and anti-patterns
- **Git Diff Visualization** - View changes across commits
- **Hub Visualization** - Identify hub/dependency-heavy modules
- **Multiple visualization modes** - Force-directed graph, tree, treemap, matrix, radial layout, flow diagram

## Tech Stack

- **CLI**: Go with Cobra, tree-sitter for AST parsing
- **Web UI**: React + TypeScript, Vite, Tailwind CSS, D3.js, WebGL
- **MCP**: Go SDK for Model Context Protocol
- **Package manager**: Bun

### MCP Tools

- `diff_since_last_map` - Show files changed since last map
- `get_changed_files` - Show uncommitted changes
- `impact_analysis` - Analyze impact of changing a file

## Project Structure

```
code-cartographer/
  cmd/              # Go CLI source
    analyze/        # Analysis commands (dead-code, complexity, circular, health)
    find/           # Code search
    lib/parser/     # Tree-sitter based multi-language parser
    mcp/            # MCP server for Claude integration
    plugins/        # Plugin system
    serve/          # HTTP server for web UI
  ui/               # React web UI
    components/     # UI components
    visualizers/    # Visualization engines (WebGL, D3, canvas)
    parsers/        # Client-side parsers
    analyzers/      # Client-side analyzers
    models/         # Data models
    engine/         # Core engine logic
  scanner/          # Scanner library
  bin/              # Built CLI binaries (gitignored)
  build/            # Web UI output (gitignored)
```

## Development

```bash
# Install dependencies
bun install

# Start dev server
bun run dev

# Run tests
bun run test

# Type check
bun run type-check

# Build for production
bun run build
```

## License

MIT
