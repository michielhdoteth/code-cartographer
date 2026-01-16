---
description: Generate LLM-friendly code analysis reports (tree, dependencies, health, coverage)
parameters:
  format:
    type: string
    description: Report format
    required: false
    default: tree
    enum: [tree, deps, health, coverage]
  focus:
    type: string
    description: Focus file or node for blast radius analysis
    required: false
  metric:
    type: string
    description: Sort by metric (churn, blast, complexity)
    required: false
    enum: [churn, blast, complexity]
  force:
    type: boolean
    description: Overwrite existing report
    required: false
    default: false
---

# carto-report

Generate code analysis reports optimized for LLM consumption. All output is plain text designed to be easily parsed and reasoned about by Claude.

## Usage

```
/carto:carto-report
/carto:carto-report format=tree
/carto:carto-report format=deps
/carto:carto-report format=health
/carto:carto-report format=coverage metric=churn
/carto:carto-report format=coverage focus=src/api.ts
```

## Report Formats

### Tree View (Default)
Shows code structure as hierarchical tree with metrics for each file and function.

```
# Code Structure

src/
  api.ts (245 LOC, 8 items) [churn: 12]
  components/
    Header.tsx (120 LOC, 3 items)
  services/
    auth.ts (180 LOC, 6 items) [HIGH IMPACT]
```

**Use cases:**
- Overview of project structure
- Identifying large files
- Planning refactoring targets

### Dependencies (deps)
Shows import and call relationships organized by edge type.

```
## IMPORTS
  api.ts -> auth.ts (imports)
  components/Header.tsx -> api.ts (imports)

## CALLS
  Header.tsx -> api.getUsers (calls)
  api.getUsers -> auth.validateToken (calls)

## DEPENDS_ON
  services/api.ts -> utils/validation.ts (depends_on)
```

**Use cases:**
- Understanding module coupling
- Identifying circular dependencies
- Planning decoupling strategies

### Health Report
Comprehensive analysis of code quality with issues and recommendations.

```
# Code Health Report

## Statistics
  Total Nodes: 87
  Total Edges: 142
  Circular Dependencies: 2
  Layer Violations: 3
  Security Issues: 1

## Issues Found
  [CRITICAL] Circular: api.ts -> auth.ts -> api.ts
  [HIGH] Layer Violation: data/models.ts imports from services/api.ts
  [MEDIUM] Unused Function: old_handler() in utils.ts

## Recommendations
  - Refactor circular dependency between api.ts and auth.ts
  - Move data models to appropriate layer
  - Remove or refactor old_handler()
```

**Use cases:**
- Code quality assessment
- Finding technical debt
- Planning improvement roadmap

### Coverage (Hotspots)
Identifies high-risk files by combining churn and blast radius metrics.

```
# Coverage & Hotspots

File | LOC | Churn | Impact | Risk Level
----------------------------------------
src/api.ts | 245 | 18 | 8 | CRITICAL
src/auth.ts | 180 | 12 | 5 | HIGH
src/components/Header.tsx | 120 | 5 | 3 | MEDIUM

## Summary
  Average Churn: 8.2
  Total Impact Score: 2145.3
  Critical Files: 2
  High Priority Files: 5
```

**Use cases:**
- Finding files needing attention
- Prioritizing code reviews
- Identifying hotspots for refactoring

## Metrics

### Churn
Number of commits affecting a file. Indicates:
- Frequently modified files (potential instability)
- Unfinished features (many recent changes)
- Files needing careful testing

### Blast Radius
How many other files depend on this file. Indicates:
- Risk of breaking changes
- Importance for other modules
- Coupling strength

### Complexity
Cyclomatic complexity score. Indicates:
- Code difficulty
- Testing challenges
- Refactoring priority

## Blast Radius Analysis

Use `focus` parameter to analyze impact of changing a specific file:

```
/carto:carto-report format=coverage focus=src/api.ts

# Blast Radius Analysis: src/api.ts

## Risk Level: CRITICAL
## Direct Dependents: 8
## Transitive Dependents: 23
## Impact Score: 12.5

## Directly Affected Files (8)
  - src/components/Header.tsx
  - src/components/Dashboard.tsx
  - src/services/notifications.ts
  - ... and 5 more

## Dependencies (4)
  - src/auth.ts
  - src/utils/validation.ts
  - src/config/api.ts
  - src/models/user.ts
```

## Integration with Workflow

This is **Step 3** of the complete analysis workflow:

1. **Scan** (`/carto:carto-scan`) - Find all code files
2. **Parse** (`/carto:carto-parse`) - Extract code structure
3. **Report** (`/carto:carto-report`) - Generate analysis reports
4. **Canvas** (`/carto:carto-canvas`) - Interactive visualization
5. **Detect** (`/carto:carto-detect`) - Find issues and patterns

## Examples

### Quick Overview for LLM
```
/carto:carto-report format=tree
```
Shows file structure with metrics. Perfect for giving Claude context about your codebase.

### Find Problem Areas
```
/carto:carto-report format=coverage metric=churn
```
Shows files that change most frequently - good candidates for testing and documentation.

### Understand Architecture
```
/carto:carto-report format=deps
```
Shows how modules depend on each other. Use to understand architecture and plan refactoring.

### Assess Code Quality
```
/carto:carto-report format=health
```
Full health assessment with issues and recommendations.

### Analyze Change Impact
```
/carto:carto-report format=coverage focus=src/core/api.ts
```
Shows what breaks if you change this file. Use before making changes.

## Output

All reports are saved as markdown files in `.code-map/reports/`:
- `code-structure.md` - Tree format
- `dependencies.md` - Dependencies format
- `code-health.md` - Health report
- `hotspots.md` - Coverage report

## Tips

- Run `/carto:carto-report format=tree` first to get familiar with codebase structure
- Use `/carto:carto-report format=health` to identify technical debt
- Use `/carto:carto-report format=coverage` to find files needing tests
- Use `/carto:carto-report format=deps` to understand coupling and plan refactoring
- Combine with `/carto:carto-graph` for visual representations

## Advanced Usage

### Identify Critical Paths
```
/carto:carto-report format=deps | grep IMPORTS
```
Find import chains that create tight coupling.

### Find Dead Code
```
/carto:carto-report format=health
```
Health report includes list of unused functions.

### Plan Refactoring
```
/carto:carto-report format=coverage metric=blast
```
Sort by impact to identify high-value refactoring targets.
