---
description: Analyze codebase for patterns, frameworks, and health metrics
parameters:
  type:
    type: string
    description: Analysis type
    required: false
    default: all
    enum: [frameworks, patterns, health, metrics, all]
  focus:
    type: string
    description: Focus on specific file or node
    required: false
---

# carto-analyze

Analyze codebase for frameworks, design patterns, code health, and metrics.

## Purpose

Run comprehensive code analysis to understand:
- **Frameworks**: Detect Django, React, Spring, Express, etc.
- **Patterns**: Design patterns and anti-patterns
- **Health**: A-F grading and code quality scores
- **Metrics**: Code churn, complexity, blast radius


## Parameters

- **type** (optional): Analysis scope - `frameworks`, `patterns`, `health`, `metrics`, or `all` (default: `all`)
- **focus** (optional): Focus analysis on specific file or node

## Usage Examples

### Full analysis (all types)
```
/carto:carto-analyze
```

### Framework detection only
```
/carto:carto-analyze type=frameworks
```

### Health report with grades
```
/carto:carto-analyze type=health
```

### Analyze specific file
```
/carto:carto-analyze focus="src/auth/service.ts"
```

## Output

**Frameworks**:
```
Backend:
  - Express 4.18.2
  - TypeScript 5.1.0

Frontend:
  - React 18.2.0
  - TailwindCSS 3.3.0
```

**Health**:
```
Overall Grade: B

Metrics:
  - Complexity: 6.2/10 (Moderate)
  - Test Coverage: 78% (Good)

Top Issues:
  1. High complexity in src/services/auth.ts (CC=42)
  2. Circular dependency: UserService <-> AuthService
```

**Patterns**:
```
Creational:
  - Factory Pattern: src/factories/serviceFactory.ts
  - Singleton Pattern: src/services/database.ts

Anti-Patterns:
  - God Object: UserService (1,200 lines)
```

## Workflow

Use `carto-analyze` to understand code quality:
- `type=frameworks` - Know dependencies and tech stack
- `type=health` - Get actionable quality metrics
- `type=patterns` - Identify design issues early
- `focus=file` - Deep dive on specific areas
