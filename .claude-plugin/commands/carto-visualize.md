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
    description: Metric overlay
    required: false
    enum: [churn, blast, complexity]
---

# carto-visualize

Generate visualizations and reports from your code map.

## Purpose

Create visual representations of your codebase:
- **tree**: File hierarchy (LLM-optimized for AI analysis)
- **deps**: Dependency graph showing imports and relationships
- **calls**: Function call hierarchy
- **hierarchy**: Class inheritance structure
- **health**: Quality metrics visualization
- **summary**: Quick overview report

## Parameters

- **format** (optional): Visualization type - `tree`, `deps`, `calls`, `hierarchy`, `health`, `coverage`, `summary` (default: `tree`)
- **focus** (optional): Focus on specific node or file for detailed view
- **interactive** (optional): Launch interactive web canvas at localhost:5173
- **metric** (optional): Overlay metric - `churn`, `blast`, `complexity`

## Usage Examples

### File hierarchy tree
```
/carto:carto-visualize --format=tree
```

### Dependency graph
```
/carto:carto-visualize --format=deps
```

### Call graph focused on function
```
/carto:carto-visualize --format=calls --focus=main
```

### Interactive canvas with churn metric
```
/carto:carto-visualize --interactive --metric=churn
```

### Health report with complexity overlay
```
/carto:carto-visualize --format=health --metric=complexity
```

## Output

**Tree**:
```
src/
+-- controllers/
|   +-- user.ts (UserController)
|   +-- auth.ts (AuthController)
+-- services/
|   +-- user.ts (UserService)
+-- models/
    +-- user.ts (User, UserRole)
```

**Deps**:
```
UserController
  +--imports--> UserService
  +--imports--> express

UserService
  +--imports--> User
  +--imports--> database
```

**Calls**:
```
main()
+-- initialize()
|   +-- loadConfig()
|   +-- setupDatabase()
+-- startServer()
    +-- createExpressApp()
```

**Interactive**: Opens localhost:5173 with full interactive canvas

## Workflow

Use `carto-visualize` to explore and understand:
- `--format=tree` - Get quick overview for LLM context
- `--format=deps` - Understand module relationships
- `--format=calls` - Trace execution paths
- `--interactive` - Explore interactively with team
- `--metric=*` - Overlay metrics for decision making
