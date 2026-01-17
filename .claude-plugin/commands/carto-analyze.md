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

Analyze your codebase for frameworks, design patterns, code health, and metrics.

## Purpose

One command to understand code quality and architecture:
- **Frameworks**: Detect Django, React, Spring, Express, etc.
- **Patterns**: Design patterns and anti-patterns
- **Health**: A-F grading, code quality scores
- **Metrics**: Churn, complexity, blast radius
- **All** (default): Run comprehensive analysis

Replaces the old `carto-detect` command with expanded capabilities.

## Parameters

- **type** (optional): Analysis type - `frameworks`, `patterns`, `health`, `metrics`, or `all` (default)
- **focus** (optional): Focus on specific file or node for detailed analysis

## Usage Examples

### Run all analyses
```
/carto:carto-analyze
```

### Detect frameworks and libraries
```
/carto:carto-analyze type=frameworks
```

### Get code health report
```
/carto:carto-analyze type=health
```

### Analyze specific file
```
/carto:carto-analyze focus="src/auth/service.ts"
```

### Get complexity metrics
```
/carto:carto-analyze type=metrics
```

## Output

**Frameworks Analysis**:
```
Detected Frameworks & Libraries
================================

Backend:
  - Express 4.18.2
  - TypeScript 5.1.0
  - PostgreSQL (via pg)

Frontend:
  - React 18.2.0
  - React Router 6.4.0
  - TailwindCSS 3.3.0

Testing:
  - Jest 29.5.0
  - Vitest 1.0.0
```

**Health Report**:
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

**Patterns Analysis**:
```
Design Patterns Detected
========================

Creational:
  - Factory Pattern: src/factories/serviceFactory.ts
  - Singleton Pattern: src/services/database.ts

Structural:
  - Adapter Pattern: src/adapters/jsonAdapter.ts

Anti-Patterns:
  - God Object: UserService (1,200 lines)
  - Callback Hell: src/handlers/upload.ts:45-78
```

## Workflow

Use `carto-analyze` to assess code quality and architecture:
- Understand tech stack → `type=frameworks`
- Assess code quality → `type=health`
- Find design patterns → `type=patterns`
- Measure complexity → `type=metrics`
- Full analysis → default (all types)
