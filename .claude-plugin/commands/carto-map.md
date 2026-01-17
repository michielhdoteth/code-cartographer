---
description: Initialize and scan codebase in one step
parameters:
  path:
    type: string
    description: Root path to map
    required: true
  name:
    type: string
    description: Project name (defaults to directory name)
    required: false
  rescan:
    type: boolean
    description: Update file list only (skip initialization)
    required: false
    default: false
---

# carto-map

Initialize a new code map OR rescan an existing one in a single step.

## Purpose

Creates `.code-map/` folder and scans for all source files by language in one streamlined step.

## Parameters

- **path** (required): Root directory to map
- **name** (optional): Custom name for code map (defaults to directory name)
- **rescan** (optional): Set to `true` to skip initialization and just update files

## Usage Examples

### First-time setup
```
/carto:carto-map path="./my-project"
```

### Custom project name
```
/carto:carto-map path="./backend" name="api-service"
```

### Update file list (after adding/removing files)
```
/carto:carto-map path="./my-project" rescan=true
```

## Output

```
Initialized code map: my-project
Location: ./my-project/.code-map/

Scanning for source files...
Found 47 files across 3 languages

Files by language:
  - TypeScript: 28 files (3,421 lines)
  - Python: 15 files (1,982 lines)
  - JavaScript: 4 files (312 lines)

Next step: Run /carto:carto-parse to extract code structure
```

## Workflow

**Step 1 of 6** in the standard workflow:
1. **Map** (`/carto:carto-map`) - Initialize and find files
2. **Parse** (`/carto:carto-parse`) - Extract code structure
3. **Analyze** (`/carto:carto-analyze`) - Assess quality
4. **Visualize** (`/carto:carto-visualize`) - Generate views
5. **Find** (`/carto:carto-find`) - Search code
6. **Info** (`/carto:carto-info`) - Check status
