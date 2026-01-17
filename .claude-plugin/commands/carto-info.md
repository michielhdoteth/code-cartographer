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

Show code map information, status, and changes.

## Purpose

Check the current state of your code map:
- **Default**: Show overall statistics (file count, node count, languages)
- **With --diff**: Show git-style diff of changes made to the code map

Combines the old `carto-status` and `carto-diff` commands into one.

## Parameters

- **diff** (optional): Show changes instead of status (default: false)

## Usage Examples

### Show current status
```
/carto:carto-info
```

### Show changes to code map
```
/carto:carto-info --diff
```

## Output

**Status Format**:
```
Code Map Status
===============

Project: my-project
Location: ./my-project/.code-map/
Last Updated: 2024-01-16 14:32:15

File Statistics:
  - Total Files: 47
  - Languages: 3 (TypeScript, Python, JavaScript)

Code Statistics:
  - Nodes: 312 (functions, classes, variables)
  - Edges: 528 (imports, dependencies, references)
  - Modules: 18

Coverage:
  - Source Code: 95% (47/47 files)
  - AST Extraction: 98%
  - Analysis: 92%

Next Steps:
  1. Run /carto:carto-parse to update code structure
  2. Run /carto:carto-analyze to assess code health
  3. Run /carto:carto-visualize to explore the graph
```

**Diff Format**:
```
Code Map Changes
================

Modified Files: 5
  M src/controllers/user.ts (+12 lines, -4 lines)
  M src/services/auth.ts (+8 lines)
  A src/models/role.ts (new file, 24 lines)
  D src/utils/deprecated.ts (removed)
  M test/integration/auth.test.ts (+6 lines)

Changes in Nodes:
  + UserController.getProfile() [NEW]
  + UserController.updateProfile() [NEW]
  - UserController.logout() [REMOVED]
  ~ UserService.authenticate() [MODIFIED]

Changes in Edges:
  + UserController --imports--> RoleService [NEW]
  - UserService --imports--> OldAuthLib [REMOVED]

Complexity Impact:
  - UserService: 32 → 35 (cyclomatic complexity +3)
  - RoleService: 12 → 18 (new dependencies +3)
```

## Workflow

Use `carto-info` to monitor your code map:
- Check current state → default
- Review recent changes → `--diff`
- Track before/after analysis → use with `carto-analyze`
