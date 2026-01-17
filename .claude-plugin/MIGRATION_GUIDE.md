# Code Cartographer v2.0 - Command Migration Guide

## Overview

Code Cartographer v2.0 streamlines the plugin from **11 commands to 6 optimized commands** for a simpler, more intuitive user experience.

### What Changed

| Old Commands | New Command | Rationale |
|---|---|---|
| `carto-init` + `carto-scan` | `carto-map` | Combined sequential operations into single workflow |
| `carto-query` + `carto-search` | `carto-find` | Unified search interface for map and source code |
| `carto-graph` + `carto-canvas` + `carto-report` | `carto-visualize` | Single interface for all visualizations |
| `carto-detect` | `carto-analyze` | Expanded to include health and metrics analysis |
| `carto-status` + `carto-diff` | `carto-info` | Combined state reporting commands |
| `carto-parse` | `carto-parse` | **No change** - already optimal |

---

## Migration by Use Case

### Setting Up a New Project

**Before (3 steps):**
```bash
/carto:carto-init path="./project"
/carto:carto-scan path="./project"
/carto:carto-parse
```

**After (2 steps):**
```bash
/carto:carto-map path="./project"
/carto:carto-parse
```

**Benefit:** Reduced from 3 commands to 2. The map operation combines initialization and scanning.

---

### Searching Your Code

**Before (two separate commands):**
```bash
# Search in parsed code map
/carto:carto-query "UserController"

# Search in source code
/carto:carto-search "def.*main"
```

**After (unified command):**
```bash
# Search everywhere (default)
/carto:carto-find "UserController"

# Search only source
/carto:carto-find "def.*main" mode=source

# Search only map
/carto:carto-find "UserController" mode=map type=node
```

**Benefit:** One command searches both, with optional mode specification for advanced use.

---

### Generating Visualizations

**Before (multiple commands for different outputs):**
```bash
# Generate ASCII graph
/carto:carto-graph --style=tree

# Open interactive canvas
/carto:carto-canvas

# Generate LLM report
/carto:carto-report --format=deps

# Generate health report
/carto:carto-report --format=health
```

**After (one command, different formats):**
```bash
# Generate ASCII tree
/carto:carto-visualize --format=tree

# Open interactive canvas
/carto:carto-visualize --interactive

# Generate dependency graph
/carto:carto-visualize --format=deps

# Generate health report
/carto:carto-visualize --format=health
```

**Benefit:** All visualization types accessible from one command with consistent parameters.

---

### Analyzing Code Quality

**Before:**
```bash
/carto:carto-detect
```

**After (expanded capabilities):**
```bash
# All analyses (equivalent to old detect)
/carto:carto-analyze

# Just frameworks
/carto:carto-analyze type=frameworks

# Health metrics
/carto:carto-analyze type=health

# All analysis types
/carto:carto-analyze type=all
```

**Benefit:** More control and expanded analysis types beyond just framework detection.

---

### Checking Project Status

**Before (two commands):**
```bash
# Show status
/carto:carto-status

# Show changes
/carto:carto-diff
```

**After (one command):**
```bash
# Show status (default)
/carto:carto-info

# Show changes
/carto:carto-info --diff
```

**Benefit:** Related operations consolidated into single command.

---

## Complete Command Reference

### New 6-Command Structure

1. **carto-map** - Initialize and scan in one step
   - Replaces: `carto-init` + `carto-scan`
   - Usage: `/carto:carto-map path="./project"`

2. **carto-parse** - Extract code structure (unchanged)
   - Usage: `/carto:carto-parse`

3. **carto-find** - Unified search
   - Replaces: `carto-query` + `carto-search`
   - Usage: `/carto:carto-find "pattern"`

4. **carto-analyze** - Code analysis
   - Replaces: `carto-detect`
   - Usage: `/carto:carto-analyze` or `/carto:carto-analyze type=health`

5. **carto-visualize** - All visualizations
   - Replaces: `carto-graph` + `carto-canvas` + `carto-report`
   - Usage: `/carto:carto-visualize --format=tree` or `--interactive`

6. **carto-info** - Status and changes
   - Replaces: `carto-status` + `carto-diff`
   - Usage: `/carto:carto-info` or `/carto:carto-info --diff`

---

## Recommended Workflow

### Standard Analysis Workflow

```bash
# 1. Initialize and scan
/carto:carto-map path="./my-project"

# 2. Parse code structure
/carto:carto-parse

# 3. Analyze code quality and patterns
/carto:carto-analyze

# 4. Generate visualizations
/carto:carto-visualize --format=tree

# 5. Search for specific code
/carto:carto-find "UserService"

# 6. Check changes
/carto:carto-info --diff
```

### Interactive Exploration

```bash
# Map and parse
/carto:carto-map path="./project"
/carto:carto-parse

# Open interactive canvas
/carto:carto-visualize --interactive

# Search for code
/carto:carto-find "pattern"

# Check health
/carto:carto-analyze type=health
```

---

## FAQ

### Q: Will my old commands still work?

**A:** No. v2.0 is a breaking change with immediate replacement. The old 11 commands have been removed and replaced with 6 optimized commands. Update your scripts and bookmarks.

### Q: Did I lose any functionality?

**A:** No. Every feature from the 11 commands is still available in the 6 new commands, often with better organization and smarter defaults.

### Q: Why consolidate commands?

**A:** The consolidation reduces cognitive load from remembering 11 commands to 6, improves discoverability, and creates more intuitive command grouping around key workflows.

### Q: How do I search both map and source?

**A:** Use `/carto:carto-find pattern` with no mode specified (defaults to `all`), or explicitly use `mode=all`.

### Q: Can I still visualize the interactive canvas?

**A:** Yes! Use `/carto:carto-visualize --interactive`

### Q: Where are the different visualization styles?

**A:** Use the `--format` parameter: tree, deps, calls, hierarchy, health, coverage, summary

---

## Breaking Changes Summary

- 11 commands → 6 commands
- Consolidated overlapping functionality
- Simplified parameter structure
- Smarter defaults reduce parameter requirements
- Zero feature loss

---

## Questions?

Refer to individual command documentation for detailed parameters and examples:
- `/carto:carto-map --help`
- `/carto:carto-find --help`
- `/carto:carto-visualize --help`
- And more...

Or check the `WORKFLOW.md` file in examples for complete workflow walkthroughs.
