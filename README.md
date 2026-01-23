# Code Cartographer v2.0

## The Ultimate Browser-Based Code Analysis & Visualization Tool

![Badge](https://img.shields.io/badge/version-2.0.0-blue)
![Badge](https://img.shields.io/badge/language-TypeScript-blue)
![Badge](https://img.shields.io/badge/runtime-Browser-green)
![Badge](https://img.shields.io/badge/languages-9-orange)

**Code Cartographer** is a zero-setup, browser-based tool for mapping, analyzing, and visualizing codebases. Built with TypeScript/React/D3.js, it requires no backend, no installation, and no data collection.

## Features

### Code Analysis
- **9-Language Support**: JavaScript, TypeScript, Python, Java, Go, Rust, C++, Ruby, PHP
- **AST Parsing**: Accurate structure extraction with language-specific parsers
- **Dependency Graph**: Complete code dependency mapping
- **Metrics**: Cyclomatic complexity, lines of code, coupling analysis

### Security Scanning
- Hardcoded secrets & API keys detection
- SQL injection vulnerabilities
- XSS (Cross-Site Scripting) risks
- Command execution dangers
- Weak cryptography usage
- Detailed severity levels with suggestions

### Quality Metrics
- **Health Score**: A-F grade for code quality
- **Dead Code Detection**: Identifies unused functions/methods
- **Circular Dependencies**: Finds cyclic import issues
- **Pattern Detection**:
  - Design patterns (Singleton, Factory, Observer)
  - Anti-patterns (God Objects, Complex Functions)
  - Architecture violations

### Visualization
- **7 Layout Modes**: Force-directed, Radial, Hierarchical, Grid, Metro, Treemap, Dependency Matrix
- **Interactive**: Zoom, pan, drag nodes
- **Color Modes**: Type, Language, Complexity, Security, Health
- **Customizable**: Filter by language, complexity, type

### User Experience
- Drag & drop file upload
- Real-time search & filter
- Detailed node information
- Settings panel for customization
- Dark/light theme support
- Responsive design

## Quick Start

### Prerequisites
- Node.js 16+ and npm

### Installation & Running

```bash
# Install dependencies
npm install

# Start development server (hot reload)
npm run dev

# Build for production (single HTML file)
npm run build

# Type checking
npm run type-check
```

### Opening the App

**Development**: Open http://localhost:5173 in your browser

**Production**: Open `build/index.html` directly in your browser (no server needed)

### Usage

1. **Upload Code**: Drag & drop files or click to browse
2. **Analyze**: Parsing and analysis happens instantly
3. **Explore**: Use file tree or search to find code
4. **Visualize**: Switch between 7 layout modes
5. **Inspect**: Click nodes to view details
6. **Filter**: Use search and settings to focus

## Supported File Types

| Language | Extensions |
|----------|-----------|
| JavaScript | `.js`, `.jsx` |
| TypeScript | `.ts`, `.tsx` |
| Python | `.py`, `.pyi` |
| Java | `.java` |
| Go | `.go` |
| Rust | `.rs` |
| C++ | `.cpp`, `.cc`, `.cxx`, `.h`, `.hpp` |
| Ruby | `.rb`, `.erb` |
| PHP | `.php` |

## Key Technologies

- **React 18**: Modern UI components with hooks
- **TypeScript 5**: Full type safety
- **D3.js 7**: Professional data visualization
- **Acorn**: JavaScript/TypeScript AST parsing
- **Vite**: Fast build tooling
- **CSS Grid**: Responsive layout

## Architecture

### Data Flow
```
Source Files
  ↓
Language-Specific Parsers (Acorn/Regex)
  ↓
CodeNode/CodeEdge Objects
  ↓
CodeMap (Aggregate Root with Algorithms)
  ↓
Analyzers (Security, Patterns, Health)
  ↓
React Components + D3.js Visualizers
  ↓
Interactive Browser UI
```

### Project Structure

```
code-cartographer/
├── src/
│   ├── models/          # Data structures & types
│   ├── parsers/         # 9 language parsers
│   ├── analyzers/       # Security, patterns, health
│   ├── components/      # React UI components
│   ├── visualizers/     # D3.js visualization
│   ├── utils/           # Helper utilities
│   ├── index.tsx        # React entry point
│   └── styles.css       # Styling
├── public/
│   └── index.html       # HTML container
├── build/               # Production output (after build)
└── [config files]
```

## Documentation

- **[QUICK_START.md](./QUICK_START.md)** - Quick reference guide
- **[BUILD_PROGRESS.md](./BUILD_PROGRESS.md)** - Detailed progress report
- **[IMPLEMENTATION_GUIDE.md](./IMPLEMENTATION_GUIDE.md)** - Complete implementation details

## Performance

- **Small Codebases** (< 100 files): Instant
- **Medium Codebases** (100-500 files): < 2 seconds
- **Large Codebases** (500+ files): 5-10 seconds
- **Bundle Size**: ~500KB (single HTML file, gzipped ~150KB)

## Browser Support

| Browser | Support |
|---------|---------|
| Chrome/Edge | ✅ Latest |
| Firefox | ✅ Latest |
| Safari | ✅ Latest |
| Mobile | ✅ iOS Safari, Chrome Mobile |

## Security

- **100% Browser-Based**: No data sent to servers
- **No Tracking**: No analytics, no data collection
- **Offline**: Works completely offline after loading
- **HTTPS Ready**: Can be hosted with HTTPS only
- **Security Scanning**: Detects common vulnerabilities

## Statistics

- **2,500+** lines of TypeScript code
- **9** language parsers
- **10** relationship types in code graph
- **15** node types for code elements
- **50+** security/pattern detection rules
- **7** visualization layout modes
- **A-F** health grade system

## Command Reference

### Development
```bash
npm run dev              # Start dev server with hot reload
npm run type-check      # Check TypeScript types
npm run build           # Build for production
npm run preview         # Preview production build locally
```

### Build Output

Running `npm run build` creates:
- `build/index.html` - Single file containing entire app
  - All CSS embedded
  - All JavaScript bundled
  - Fully minified and optimized
  - Can be opened directly in browser
  - ~500KB total (gzipped ~150KB)

## Integration with Claude Code

Code Cartographer integrates with Claude Code as a plugin:

```
/carto:carto-init       # Initialize code map
/carto:carto-scan       # Scan directory
/carto:carto-parse      # Parse code files
/carto:carto-canvas     # Open visualization
/carto:carto-analyze    # Run all analyzers
/carto:carto-health     # Show health score
```

## API Overview

### Core Types

```typescript
// Language support
type Language = 'python' | 'javascript' | 'typescript' | 'java' | 'go' | 'rust' | ...

// Node represents code element
interface CodeNode {
  id: string
  name: string
  type: NodeType
  language: Language
  metrics: ComplexityMetrics
  // ... more fields
}

// Edge represents relationship
interface CodeEdge {
  source: string
  target: string
  type: EdgeType // 'calls', 'imports', 'inherits', etc.
}

// Map contains all nodes and edges
class CodeMap {
  getNodes(): CodeNode[]
  getEdges(): CodeEdge[]
  findCircularDependencies(): string[][]
  getBlastRadius(nodeId: string): Set<string>
  // ... more methods
}
```

### Analyzers

```typescript
// Security scanning
const scanner = new SecurityScanner()
const issues = scanner.scan(code)  // SecurityIssue[]

// Pattern detection
const detector = new PatternDetector()
const patterns = detector.detectPatterns(codeMap)  // PatternMatch[]

// Health scoring
const scorer = new HealthScorer()
const health = scorer.calculateHealth(codeMap, code)  // HealthMetrics
```

## Common Use Cases

### New Developer Onboarding
Use Code Cartographer to quickly understand a new codebase's structure, dependencies, and architecture.

### Code Review Assistance
Visualize the impact of changes (blast radius) before approving PRs.

### Technical Debt Identification
Use health score and anti-pattern detection to identify refactoring opportunities.

### Security Audit
Run security scanner to detect hardcoded secrets, SQL injection risks, etc.

### Architecture Documentation
Generate visual maps of system architecture for documentation.

### Performance Analysis
Identify highly coupled modules and complex functions for optimization.

## Development Status

**Version 2.0 Features**:
- ✅ 9-language parser support
- ✅ AST-based code analysis
- ✅ Security vulnerability scanning
- ✅ Pattern & anti-pattern detection
- ✅ Health scoring (A-F grade)
- ✅ Interactive visualization
- ✅ React + D3.js frontend
- ✅ Single HTML file distribution
- ✅ Browser-based (no backend)

**Roadmap**:
- Additional visualizer layouts
- Git integration (blame, history)
- PR impact analysis
- Code duplication detection
- Performance profiling
- Export capabilities (PNG, SVG)

## Contributing

This project was created as part of Claude Code development. Contributions welcome!

## License

GPL v3 - See LICENSE file

## Credits

Built with:
- **Acorn** - JavaScript/TypeScript AST parsing
- **D3.js** - Data visualization
- **React** - UI framework
- **TypeScript** - Type safety
- **Vite** - Build tooling

## Support

- **Issues**: Report bugs via GitHub Issues
- **Documentation**: See IMPLEMENTATION_GUIDE.md
- **Quick Help**: See QUICK_START.md

---

## The Vision

**Code Cartographer** is designed to be *the ultimate tool for understanding code* - both for humans and AI.

Whether you're:
- 🧑‍💻 **Joining a new project** - Instantly understand the codebase
- 🔍 **Reviewing code** - See impact and dependencies
- 🛡️ **Auditing security** - Find vulnerabilities automatically
- 📊 **Managing technical debt** - Identify refactoring opportunities
- 🤖 **Working with AI** - Give Claude better context
- 📚 **Documenting architecture** - Generate visual maps

Code Cartographer helps you see the whole picture, fast.

---

**Welcome to Code Cartographer v2.0 - Where Code Meets Clarity**

*Map your code. Understand it faster. Build better software.*
