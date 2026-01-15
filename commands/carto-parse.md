---
description: Parse source files with AST and extract nodes/edges
parameters:
  languages:
    type: array
    description: Programming languages to parse
    required: false
    default: ["python"]
---

# carto-parse

Parse source files using AST and extract code structure.

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

```
/carto:carto-parse

Output:
Parsed 28 files
Nodes added: 418
Edges added: 342
Total nodes: 418
Total edges: 342
```
