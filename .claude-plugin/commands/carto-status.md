---
description: Show current status of the code map
parameters: {}
---

# carto-status

Show the current status of the code map.

## Usage

```
/carto:carto-status
```

## Output

```json
{
  "initialized": true,
  "root_path": "/path/to/project",
  "folder_exists": true,
  "toon_exists": true,
  "md_exists": true,
  "files": 45,
  "nodes": 418,
  "edges": 342
}
```

## Status Values

| Field | Description |
|-------|-------------|
| `initialized` | Whether init() has been run |
| `folder_exists` | Whether `.code-map/` exists |
| `toon_exists` | Whether `code-map.toon` exists |
| `md_exists` | Whether `code-graph.md` exists |
| `files` | Number of source files |
| `nodes` | Number of code elements |
| `edges` | Number of relationships |

## Example

```
/carto:carto-status

Output:
Status: initialized
Files: 45 | Nodes: 418 | Edges: 342
TOON: /path/to/project/.code-map/code-map.toon
MD: /path/to/project/.code-map/code-graph.md
```
