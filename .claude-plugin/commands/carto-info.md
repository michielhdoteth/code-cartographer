---
description: Show code map information and changes
parameters:
  diff:
    type: boolean
    description: Show changes instead of status
    required: false
    default: false
---

# carto-info

Show code map status and track changes.

## Purpose

Monitor your code map state and understand what changed:
- **Default**: Show current statistics (file count, node count, languages)
- **With --diff**: Show git-style diff of changes to code map

## Parameters

- **diff** (optional): Show changes instead of status (default: false)

## Usage Examples

### Check current status
```
/carto:carto-info
```

### View changes to code map
```
/carto:carto-info --diff
```

## Output

**Status** (default):
```
Code Map Status
===============

Project: my-project
Location: ./my-project/.code-map/
Last Updated: 2024-01-16 14:32:15

Statistics:
  - Files: 47
  - Languages: 3 (TypeScript, Python, JavaScript)
  - Nodes: 312
  - Edges: 528
```

**Diff** (`--diff`):
```
Code Map Changes
================

Modified Files: 5
  M src/controllers/user.ts (+12 lines, -4 lines)
  A src/models/role.ts (new file)
  D src/utils/deprecated.ts (removed)

Node Changes:
  + UserController.getProfile() [NEW]
  - UserController.logout() [REMOVED]

Edge Changes:
  + UserController --imports--> RoleService [NEW]
```

## Workflow

Use `carto-info` to monitor your analysis:
- Check status before starting analysis
- Track changes after code modifications
- Compare metrics before/after refactoring
- Identify what code changed in recent updates
