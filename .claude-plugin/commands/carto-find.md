---
description: Search code map and source code
parameters:
  pattern:
    type: string
    description: Search pattern (literal or regex)
    required: true
  mode:
    type: string
    description: Where to search
    required: false
    default: all
    enum: [map, source, all]
  type:
    type: string
    description: Entity type for map search
    required: false
    enum: [node, edge, file]
  context:
    type: number
    description: Context lines for source search
    required: false
    default: 2
---

# carto-find

Unified search across both the parsed code map AND raw source code.

## Purpose

Find code anywhere in your codebase:
- **Map mode**: Search parsed structure (functions, classes, imports)
- **Source mode**: Search raw source code with regex (like ripgrep)
- **All mode**: Search both simultaneously (default)


## Parameters

- **pattern** (required): What to search for (literal string or regex)
- **mode** (optional): Search scope - `map`, `source`, or `all` (default: `all`)
- **type** (optional): For map mode - `node`, `edge`, or `file`
- **context** (optional): Context lines for source mode (default: 2)

## Usage Examples

### Search everywhere (default)
```
/carto:carto-find "UserController"
```

### Search only parsed code map
```
/carto:carto-find "UserController" mode=map type=node
```

### Search only source code with regex
```
/carto:carto-find "def.*main" mode=source context=5
```

### Find all import edges
```
/carto:carto-find "express" mode=map type=edge
```

## Output

**Map Results**:
```
Found 3 matches in code map:

Nodes:
  - UserController (class) at src/controllers/user.ts:12
  - UserService (class) at src/services/user.ts:8

Edges:
  - UserController --imports--> UserService
```

**Source Results**:
```
Found 5 matches in source code:

src/controllers/user.ts:12
  10 |
  11 | export class UserController {
> 12 |   constructor(private userService: UserService) {}
  13 |
  14 |   async getUser(id: string) {
```

## Workflow

Use `carto-find` in your analysis workflow to locate specific code:
- Finding definitions → `mode=map type=node`
- Finding dependencies → `mode=map type=edge`
- Finding patterns → `mode=source` (regex)
- Quick search → default `mode=all`
