---
description: Generate visualizations and reports from code map
parameters:
  format:
    type: string
    description: Visualization format
    required: false
    default: tree
    enum: [tree, deps, calls, hierarchy, health, coverage, summary]
  focus:
    type: string
    description: Focus on specific node or file
    required: false
  interactive:
    type: boolean
    description: Launch interactive web canvas
    required: false
    default: false
  metric:
    type: string
    description: Metric to overlay on visualization
    required: false
    enum: [churn, blast, complexity]
---

# carto-visualize

Generate all types of visualizations and reports from your code map.

## Purpose

One command for all visualization needs:
- ASCII graphs for LLM analysis
- Interactive web canvas for exploration
- Hierarchical trees for structure
- Dependency graphs for relationships
- Health reports for quality metrics
- Call graphs for tracing execution

Replaces `carto-graph`, `carto-canvas`, and `carto-report` with unified interface.

## Parameters

- **format** (optional): Type of visualization - `tree`, `deps`, `calls`, `hierarchy`, `health`, `coverage`, `summary` (default: tree)
- **focus** (optional): Focus on specific node/file for detailed view
- **interactive** (optional): Launch interactive web canvas instead of ASCII (default: false)
- **metric** (optional): Overlay metric on visualization - `churn`, `blast`, `complexity`

## Usage Examples

### File hierarchy tree (LLM-optimized)
```
/carto:carto-visualize --format=tree
```

### Dependency graph
```
/carto:carto-visualize --format=deps
```

### Call graph focused on main function
```
/carto:carto-visualize --format=calls --focus=main
```

### Interactive visualization with churn overlay
```
/carto:carto-visualize --interactive --metric=churn
```

### Health report with complexity metrics
```
/carto:carto-visualize --format=health --metric=complexity
```

### Class hierarchy
```
/carto:carto-visualize --format=hierarchy
```

## Output

**Tree Format** (LLM-optimized):
```
src/
├── controllers/
│   ├── user.ts (UserController)
│   └── auth.ts (AuthController)
├── services/
│   ├── user.ts (UserService)
│   └── auth.ts (AuthService)
└── models/
    └── user.ts (User, UserRole)
```

**Dependency Format**:
```
UserController
  ├─imports→ UserService
  ├─imports→ express
  └─imports→ @types/node

UserService
  ├─imports→ User
  └─imports→ database
```

**Call Graph**:
```
main()
├─ initialize()
│  ├─ loadConfig()
│  └─ setupDatabase()
├─ startServer()
│  └─ createExpressApp()
│     └─ registerRoutes()
└─ registerErrorHandlers()
```

**Health Format**:
```
Code Health Report
==================

Overall Grade: B

Metrics:
  - Complexity: 6.2/10 (Moderate)
  - Test Coverage: 78% (Good)
  - Code Churn: 12 commits/file (Low)
  - Dependency Health: 8.5/10 (Excellent)

Top Issues:
  1. High complexity in src/services/auth.ts (CC=42)
  2. No tests for src/utils/parser.ts
  3. Circular dependency: UserService ↔ AuthService
```

**Interactive Mode**: Opens localhost:5173 with full interactive canvas

## Workflow

Use `carto-visualize` to understand your codebase:
- Quick overview → `--format=summary`
- Understand structure → `--format=tree`
- Find dependencies → `--format=deps`
- Trace calls → `--format=calls --focus=functionName`
- Assess quality → `--format=health`
- Explore interactively → `--interactive`
