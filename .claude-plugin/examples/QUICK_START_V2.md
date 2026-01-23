# Code Cartographer v2.0 - Quick Start Guide

## 3-Minute Setup

Get started mapping your codebase in 3 simple steps:

### Step 1: Map Your Project
```bash
/carto:carto-map path="./my-project"
```

This initializes the code map and scans for all source files. You'll see:
- Number of files found
- Supported programming languages detected
- Ready-to-parse message

### Step 2: Parse Code Structure
```bash
/carto:carto-parse
```

This extracts the code structure (functions, classes, imports). You'll get:
- Node count (code elements)
- Edge count (relationships)
- Extraction statistics

### Step 3: Explore
Choose your next step:

**Visualize the structure:**
```bash
/carto:carto-visualize --interactive
```

**Analyze code quality:**
```bash
/carto:carto-analyze
```

**Search for code:**
```bash
/carto:carto-find "FunctionName"
```

---

## The 6 Essential Commands

### 1. carto-map
**Initialize and scan a project**
```bash
/carto:carto-map path="./project"
/carto:carto-map path="./project" rescan=true
```

### 2. carto-parse
**Extract code structure**
```bash
/carto:carto-parse
/carto:carto-parse languages=["typescript", "python"]
```

### 3. carto-find
**Search code map or source**
```bash
/carto:carto-find "UserController"                    # Search everywhere
/carto:carto-find "UserController" mode=map           # Search parsed code
/carto:carto-find "def.*main" mode=source             # Search source files
```

### 4. carto-analyze
**Analyze code quality and patterns**
```bash
/carto:carto-analyze                                  # Full analysis
/carto:carto-analyze type=health                      # Health report
/carto:carto-analyze type=frameworks                  # Framework detection
```

### 5. carto-visualize
**Generate visualizations and reports**
```bash
/carto:carto-visualize --format=tree                  # File hierarchy
/carto:carto-visualize --format=deps                  # Dependencies
/carto:carto-visualize --interactive                  # Interactive canvas
/carto:carto-visualize --format=health                # Health report
```

### 6. carto-info
**Show status and changes**
```bash
/carto:carto-info                                     # Current status
/carto:carto-info --diff                              # Show changes
```

---

## Common Workflows

### Understand a New Codebase

```bash
# 1. Map and parse
/carto:carto-map path="./new-project"
/carto:carto-parse

# 2. Get overview
/carto:carto-visualize --format=tree

# 3. Assess quality
/carto:carto-analyze type=health

# 4. Find key files
/carto:carto-find "main" mode=map
```

### Interactive Exploration

```bash
# 1. Setup
/carto:carto-map path="./project"
/carto:carto-parse

# 2. Launch canvas
/carto:carto-visualize --interactive

# 3. Search while exploring
/carto:carto-find "FunctionName"
```

### Code Review

```bash
# 1. Map project
/carto:carto-map path="./project"
/carto:carto-parse

# 2. Detect changes
/carto:carto-info --diff

# 3. Analyze impact
/carto:carto-analyze type=metrics

# 4. Visualize affected areas
/carto:carto-visualize --format=deps --metric=blast
```

---

## What You Get

✓ **Automatic File Discovery** - Scans all supported languages
✓ **Code Structure Extraction** - Classes, functions, imports, relationships
✓ **Code Quality Metrics** - Complexity, churn, health scores
✓ **Pattern Detection** - Design patterns, anti-patterns
✓ **Interactive Visualization** - Explore your codebase interactively
✓ **Search Capabilities** - Find code across map and source

---

## Supported Languages

- TypeScript / JavaScript
- Python
- Java
- Go
- Rust
- C++
- Ruby
- PHP

---

## Need Help?

- Read the **MIGRATION_GUIDE.md** if upgrading from v1.x
- Check **WORKFLOW.md** for detailed workflow examples
- See individual command docs for parameters and options
- Review **LIVE_ANALYSIS.md** for real-world analysis scenarios

---

## Next Steps

Once familiar with basic commands, explore:
- Advanced filtering in `/carto:carto-find`
- Different visualization formats
- Metrics overlays on visualizations
- Batch analysis workflows

Happy mapping!
