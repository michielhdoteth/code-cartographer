---
name: Code Analyzer
description: Analyzes code patterns, detects anti-patterns, design patterns, and architectural violations in mapped codebases
capabilities:
  - Pattern detection (design patterns and anti-patterns)
  - Architectural analysis and layer violation detection
  - Code quality assessment
  - Dependency analysis and complexity scoring
  - Health metrics evaluation
skills:
  - /carto:carto-map
  - /carto:carto-visualize
  - /carto:carto-find
hooks:
  - event: post-scan
    description: Automatically analyze code patterns after scanning
  - event: query-analysis
    description: Provide pattern analysis when queried about code relationships
tools:
  - type: codebase-analysis
    name: Pattern Detection
    description: Detects design patterns, anti-patterns, and architectural violations
  - type: metrics
    name: Code Health Metrics
    description: Scores code quality across complexity, cohesion, and maintainability dimensions
  - type: dependency-graph
    name: Dependency Analysis
    description: Maps dependencies and identifies circular dependencies or layer violations
activation:
  triggers:
    - /carto:carto-analyze
    - /carto:carto-find pattern=*
    - codebase-analysis-request
  priority: normal
---

# Code Analyzer Subagent

Analyzes mapped codebases to detect patterns, anti-patterns, and architectural issues.

## Pattern Detection

### Design Patterns Detected
- **Singleton**: Classes with private constructors and static instances
- **Factory**: Classes with static factory methods (create*)
- **Observer/Event Emitter**: Classes with subscription/observation methods

### Anti-Patterns Detected
- **God Object**: Classes with excessive methods/properties (>15)
- **Long Files**: Files exceeding 500 lines of code
- **Complex Functions**: Functions with cyclomatic complexity >10
- **Dead Code**: Unused functions/methods

### Architectural Analysis
- **Layer Violations**: Detects when lower-level layers depend on higher-level layers
- **Dependency Cycles**: Identifies circular dependencies
- **Module Cohesion**: Analyzes coupling and cohesion metrics

## Code Health Metrics

Evaluates:
- **Complexity**: Cyclomatic complexity and cognitive load
- **Maintainability**: Code duplication and reusability
- **Test Coverage**: Estimated coverage based on structure
- **Security**: Pattern-based security concerns
- **Performance**: Potential bottleneck detection

## Integration with Skills

Works with:
- `code-mapper`: Uses parsed code maps for analysis
- `code-visualizer`: Visualizes detected patterns on infinite canvas
- `code-searcher`: Finds related code sections for patterns

## Hooks

- **post-scan**: Automatically runs pattern detection after code scanning
- **query-analysis**: Provides detailed analysis when queried
