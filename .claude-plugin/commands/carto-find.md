---
description: Search code map and source code
parameters:
  pattern:
    type: string
    description: Search pattern or query
    required: true
  mode:
    type: string
    description: Search mode (map, source, or both)
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

Unified search for both the parsed code map AND raw source code.

## Purpose

Search anywhere in your codebase:
- **Map mode**: Search parsed code structure (functions, classes, imports)
- **Source mode**: Search raw source code with regex (like ripgrep)
- **All mode** (default): Search both simultaneously

Replaces the old `carto-query` and `carto-search` commands with a single, more intuitive interface.

## Parameters

- **pattern** (required): What to search for (literal or regex)
- **mode** (optional): Where to search - `map`, `source`, or `all` (default)
- **type** (optional): For map mode - `node`, `edge`, or `file`
- **context** (optional): Number of context lines for source mode (default: 2)

## Usage Examples

### Search everywhere (default)
```
/carto:carto-find "UserController"
```
Output: Finds class definition in map + all source mentions

### Search only parsed code map
```
/carto:carto-find "UserController" mode=map type=node
```
Output: Returns the node representing UserController class

### Search only source code (regex)
```
/carto:carto-find "def.*main" mode=source context=5
```
Output: All lines matching regex with 5 lines of context

### Find all imports
```
/carto:carto-find "express" mode=map type=edge
```
Output: All import edges involving 'express'

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

Use `carto-find` whenever you need to locate code:
- Finding function definitions → `mode=map type=node`
- Finding dependencies → `mode=map type=edge`
- Finding regex patterns → `mode=source`
- General search → `mode=all` (default)
