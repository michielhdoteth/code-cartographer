---
description: Initialize and scan codebase in one step
parameters:
  path:
    type: string
    description: Root path to map
    required: true
  name:
    type: string
    description: Project name for the code map
    required: false
  rescan:
    type: boolean
    description: Skip initialization, just rescan files
    required: false
    default: false
---

# carto-map

Initialize a new code map OR rescan an existing one to update the file list.

## Purpose

Creates `.code-map/` folder structure and scans for source files in a single command. Combines the functionality of the old `carto-init` and `carto-scan` commands for a streamlined workflow.

## Parameters

- **path** (required): Root directory to map
- **name** (optional): Custom name for the code map. Defaults to directory name.
- **rescan** (optional): Set to `true` to skip initialization and just update the file list

## Usage Examples

### First-time setup (creates .code-map/ and scans)
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
✓ Found 47 files across 3 languages

Files by language:
  - TypeScript: 28 files (3,421 lines)
  - Python: 15 files (1,982 lines)
  - JavaScript: 4 files (312 lines)

Next step: Run /carto:carto-parse to extract code structure
```

## Workflow

**Step 1 of 3** in the standard workflow:
1. **Map** (`/carto:carto-map`) - Initialize and find files
2. **Parse** (`/carto:carto-parse`) - Extract code structure
3. **Analyze/Visualize** - Understand codebase
