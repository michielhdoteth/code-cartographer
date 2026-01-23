# Code Cartographer v2.0 - TSX Rewrite Build Progress

## Completed Components

### 1. Project Setup
- `package.json` - Dependencies configured (React, D3.js, TypeScript, Vite)
- `tsconfig.json` - TypeScript strict mode enabled with path aliases
- `tsconfig.node.json` - Build tools configuration
- `vite.config.ts` - Vite build config for single HTML output

### 2. Core Data Models (`src/models/`)
- `types.ts` - Complete type definitions (Language, NodeType, EdgeType, Severity, HealthGrade, etc.)
- `codeMap.ts` - Core data structures:
  - `CodeNode` - Code elements (classes, functions, methods)
  - `CodeEdge` - Relationships (inheritance, calls, imports, etc.)
  - `CodeFile` - File metadata
  - `CodeMap` - Main container with graph algorithms
  - Methods for: dependency analysis, blast radius, circular dependency detection, coupling metrics
- `index.ts` - Export barrel

### 3. Language Parsers (`src/parsers/`)
- `base.ts` - Abstract BaseParser class with common utilities
- `javascript.ts` - JS/JSX parser using Acorn AST library
- `typescript.ts` - TS/TSX parser (strips annotations, uses JS parser)
- `python.ts` - Python parser (regex-based, handles classes, functions, decorators)
- `java.ts` - Java parser (classes, interfaces, methods, annotations)
- `go.ts` - Go parser (functions, structs, interfaces)
- `rust.ts` - Rust parser (functions, structs, traits, enums)
- `cpp.ts` - C++ parser (classes, structs, functions)
- `ruby.ts` - Ruby parser (classes, methods, modules)
- `php.ts` - PHP parser (classes, interfaces, traits, functions)
- `index.ts` - Export barrel

### 4. Code Analyzers (`src/analyzers/`)

#### Security Scanner (`securityScanner.ts`)
- Detects: hardcoded secrets, AWS keys, private keys, SQL injection, eval() usage, XSS, command execution
- Severity levels: critical, high, medium, low
- Regex patterns with detailed suggestions
- Returns: SecurityIssue[] with location and code snippets

#### Pattern Detector (`patternDetector.ts`)
- **Design Patterns**: Singleton, Factory, Observer/Event
- **Anti-Patterns**: God Objects, Long Files, Complex Functions, Dead Code
- **Architecture Violations**: Layer violations (UI → Service → Utils)
- Returns: PatternMatch[] with severity and descriptions

#### Health Scorer (`healthScorer.ts`)
- Calculates health metrics:
  - Overall score (0-100) converted to A-F grade
  - Dead code percentage
  - Circular dependencies count
  - Average coupling (fan-in/fan-out)
  - Security issues (critical, high, medium)
  - Complex functions and God Objects count
- Provides recommendations for improvement
- Returns: HealthMetrics with grade and actionable insights

## Architecture Overview

### Data Flow
```
Source Files
  → Language Parsers (Acorn/Regex)
  → CodeNode/CodeEdge objects
  → CodeMap (graph container)
  → Analyzers (Security, Patterns, Health)
  → AnalysisResult
  → React Components + D3.js Visualizers
  → Interactive Browser UI
```

### Type System
All components use strongly-typed interfaces:
- Language support: 9 languages with specific parsers
- Node types: 15 different code elements
- Edge types: 10 relationship types
- Severity levels for issues
- Health grades A-F

## Remaining Work (Priority Order)

### Phase 1: React Components (`src/components/`)
1. **App.tsx** - Main component with state management
2. **FileTree.tsx** - Directory/file navigator
3. **DetailPanel.tsx** - Shows selected node/edge details
4. **SearchBar.tsx** - Search and filter nodes
5. **SettingsPanel.tsx** - Visualization settings (layout, colors, filters)
6. **HealthScore.tsx** - Displays health metrics and recommendations

**Dependencies**: React hooks for state, CodeMap instance

### Phase 2: D3.js Visualizers (`src/visualizers/`)
1. **forceDirectedGraph.tsx** - Physics-based layout (default)
2. **radialLayout.tsx** - Circular/hub arrangement
3. **hierarchicalLayout.tsx** - Top-down layered
4. **gridLayout.tsx** - Structured grid arrangement
5. **metroLayout.tsx** - Subway-map style
6. **treemapLayout.tsx** - Space-filling rectangles (size by LOC)
7. **dependencyMatrix.tsx** - Heat map + Sankey diagram

**Dependencies**: D3.js, SVG rendering, interaction handlers

### Phase 3: Utils & Helpers (`src/utils/`)
1. **acornParser.ts** - Acorn integration wrapper
2. **fileSystem.ts** - Browser File API helpers
3. **git.ts** - Git info parsing (if available)

### Phase 4: Entry Point & Build
1. **public/index.html** - HTML container
2. **src/index.tsx** - React root, App mount
3. **Build process** - Vite bundle to single HTML file

### Phase 5: Claude Code Plugin Integration
1. Update `commands/*.md` to reference new visualization
2. Update `agents/*.md` with new capabilities
3. Modify `skills/*.md` to use TSX components
4. Update `.claude-plugin/plugin.json` if needed

## Key Technical Decisions

1. **Browser-Based**: No backend needed, works locally
2. **Single HTML Output**: Like cloned repo, easy distribution
3. **Regex + AST Parsing**: Acorn for JS/TS, regex for others (more portable)
4. **React 18**: Modern hooks-based components
5. **D3.js v7**: Professional visualization library
6. **TypeScript Strict Mode**: Full type safety
7. **Vite Build Tool**: Fast builds, modern tooling

## Statistics

- **9 Language Parsers**: 50+ lines each
- **3 Major Analyzers**: 300+ lines total
- **Core Models**: 400+ lines with algorithms
- **Type Definitions**: 150+ lines of strict types
- **Total Code So Far**: ~2,500+ lines of TypeScript

## Next Steps

1. Create React entry point (index.tsx + public/index.html)
2. Implement main App component with state management
3. Add basic UI components (FileTree, DetailPanel)
4. Integrate first visualizer (force-directed graph)
5. Wire up analyzers to UI
6. Test with sample code files
7. Add remaining visualizers
8. Bundle to single HTML file
9. Integrate with Claude Code plugin

## Build & Run Commands

```bash
npm install           # Install dependencies
npm run dev          # Start dev server
npm run build        # Build single HTML file
npm run type-check   # Check types
```

## Notes

- The security scanner uses proven patterns for vulnerability detection
- Pattern detector uses both AST and naming conventions for detection
- Health scorer uses weighted scoring system (customizable)
- All analyzers can run on partial/incomplete code
- Visualizers are designed to handle large graphs (1000+ nodes)
