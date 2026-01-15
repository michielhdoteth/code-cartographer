---
description: Fast code search using ripgrep
parameters:
  pattern:
    type: string
    description: Search pattern (regex supported)
    required: true
  file_types:
    type: array
    description: File types to search (python, javascript, typescript, java, go, rust)
    required: false
  context:
    type: number
    description: Number of context lines around matches
    required: false
    default: 2
---

# carto-search

Fast code search using ripgrep.

## Usage

```
/carto:carto-search pattern="def.*main"
/carto:carto-search pattern="class.*Controller" file_types=["python"]
/carto:carto-search pattern="import.*react" file_types=["javascript", "typescript"]
/carto:carto-search pattern="console.log" file_types=["javascript"] context=3
```

## Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| **pattern** | string | Search pattern (regex supported) |
| **file_types** | array | File types to search |
| **context** | number | Context lines around matches (default: 2) |

## File Types

| Type | Extensions |
|------|------------|
| python | .py, .pyi |
| javascript | .js, .jsx |
|typescript | .ts, .tsx |
| java | .java |
| go | .go |
| rust | .rs |

## Examples

```
Search for all functions:
/carto:carto-search pattern="^def\s+\w+"

Search for React components:
/carto:carto-search pattern="class.*extends.*Component" file_types=["javascript"]

Search for imports:
/carto:carto-search pattern="^import.*from" file_types=["javascript", "typescript"]

Search for state machines:
/carto:carto-search pattern="self\.state\s*=" file_types=["python"]
```

## Output

Returns matches with:
- File path
- Line number
- Column
- Matched content with context

## Notes

- Uses ripgrep for fast searching
- Automatically searches all relevant file types if not specified
- Results are limited to 100 matches
- Patterns are regex-based
