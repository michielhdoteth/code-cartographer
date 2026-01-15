---
description: Initialize a new code map in .code-map/ folder
parameters:
  path:
    type: string
    description: Root path to scan
    required: true
  name:
    type: string
    description: Name for the code map
    required: false
    default: code-map
---

# carto-init

Initialize a new code map in the `.code-map/` folder.

## Usage

```
/carto:carto-init path="/path/to/project"
/carto:carto-init path="/path/to/project" name="my-project"
```

## What Gets Created

Creates `.code-map/` folder with:
- `code-map.toon` - Empty code map structure
- `code-graph.md` - Placeholder for ASCII output

## Next Steps

1. Run `/code-cartographer:carto-scan` to find files
2. Run `/code-cartographer:carto-parse` to extract structure
3. Run `/code-cartographer:carto-graph` to generate ASCII view
4. Run `/code-cartographer:carto-canvas` to open visualization

## Example

```
/carto:carto-init path="/workspace/myapp"
```

Output:
```
Initialized code map: myapp
Folder: /workspace/myapp/.code-map/
Files: code-map.toon, code-graph.md
```
