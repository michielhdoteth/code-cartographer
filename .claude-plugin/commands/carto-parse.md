---
description: Parse source files with AST and extract code structure
parameters:
  languages:
    type: array
    description: Languages to parse (all if omitted)
    required: false
---

# carto-parse

Parse source files using AST to extract code structure.

## Purpose

Extract code elements and relationships from all source files:
- **Classes**: Name, inheritance, methods, decorators
- **Functions**: Name, parameters, return type
- **Calls**: Function call relationships
- **Imports**: Module dependencies

## Parameters

- **languages** (optional): Filter to specific languages (e.g., `["python", "javascript"]`). If omitted, parses all detected languages.

## Usage Examples

### Parse all languages
```
/carto:carto-parse
```

### Parse specific languages only
```
/carto:carto-parse languages=["python", "javascript"]
```

### Parse TypeScript and Python
```
/carto:carto-parse languages=["typescript", "python"]
```

## Output

```
Parsing 3 files...
Extracted 23 nodes
  - 3 classes
  - 2 interfaces
  - 18 methods/functions
Found 31 relationships
  - 5 imports
  - 12 calls
  - 14 dependencies

Code map ready for visualization!
```

## Workflow

**Step 2** of the workflow:
1. `carto-map` - Initialize and find files
2. `carto-parse` - Extract code structure (this command)
3. `carto-analyze` - Assess code quality
4. `carto-visualize` - Generate views
5. `carto-find` - Search code
6. `carto-info` - Check status
