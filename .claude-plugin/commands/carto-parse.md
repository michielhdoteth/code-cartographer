---
description: Parse source files with AST and extract code structure
parameters:
  languages:
    type: array
    description: Programming languages to parse
    required: false
---

# carto-parse

Parse source files using AST and extract code structure (nodes and edges).

## Usage

```
/carto:carto-parse
/carto:carto-parse languages=["python", "javascript"]
```

## What Gets Extracted

- **Classes**: Name, inheritance, methods, decorators
- **Functions**: Name, parameters, return type, decorators
- **Methods**: Same as functions, plus class association
- **Calls**: Function call relationships
- **Imports**: Module dependencies

## Output

Updates `code-map.toon` with:
- `nodes`: All extracted code elements
- `edges`: Relationships (contains, calls, inherits, imports)
- `statistics`: Counts by type

## Statistics

```
Classes: 23
Functions: 156
Methods: 29
Edges: 342
```

## Example

### Parse All Languages
```
/carto:carto-parse
```

Output:
```
Parsing 3 files...
✓ Extracted 23 nodes
  - 3 classes
  - 2 interfaces
  - 18 methods/functions
✓ Found 31 relationships
  - 5 imports
  - 12 calls
  - 14 dependencies
✓ Calculated metrics for all entities

Code map ready for visualization!
```

### Parse Specific Languages
```
/carto:carto-parse languages=["typescript", "python"]
```

Output:
```
Parsing TypeScript and Python files...
✓ Found 2 files (api.ts, database.py)
✓ Extracted 19 nodes
✓ Found 28 relationships
```

## Workflow

This is **Step 2** of the optimized workflow:

1. **Map** (`/carto:carto-map`) - Initialize and find files
2. **Parse** (`/carto:carto-parse`) - Extract code structure
3. **Analyze** (`/carto:carto-analyze`) - Assess code quality
4. **Visualize** (`/carto:carto-visualize`) - Generate views and reports
5. **Find** (`/carto:carto-find`) - Search code map or source
6. **Info** (`/carto:carto-info`) - Check status and changes

See `WORKFLOW.md` in examples for complete walkthrough.
