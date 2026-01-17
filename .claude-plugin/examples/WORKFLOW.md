# Code Cartographer Workflow Example

## Quick Start: Map Your Project in 3 Steps

### Step 1: Map (Initialize and Scan)
```bash
/carto:carto-map path="./examples/sample-project"
```

**Output:**
```
Initialized code map: sample-project
Location: ./examples/sample-project/.code-map/

Scanning for source files...
✓ Found 3 files across 2 languages

Files by language:
  - TypeScript: 2 files (59 lines, 43 lines)
  - Python: 1 file (75 lines)

Next step: Run /carto:carto-parse to extract code structure
```

---

### Step 2: Parse
```bash
/carto:carto-parse
```

**Output:**
```
Parsing 3 files...
✓ Extracted 23 nodes
  - 3 classes
  - 2 interfaces
  - 18 methods/functions
✓ Found 31 relationships
  - 5 imports
  - 12 calls
  - 14 dependencies
✓ Calculated metrics for all entities

Code map ready for visualization!
```

---

### Step 3: Visualize
```bash
/carto:carto-visualize --format=tree
```

**Output:**
```
src/
├── api.ts (APIHandler)
│   ├── handle()
│   ├── getUsers()
│   └── createUser()
├── auth.ts (AuthManager)
│   ├── login()
│   └── isAuthenticated()
└── database.py (Database)
    ├── connect()
    └── get_user()
```

---

## Full Workflow with Analysis

### 1. Initialize and Map Project
```bash
/carto:carto-map path="./my-project"
```
Creates .code-map/ folder and scans for all source files by language

### 2. Parse & Extract Structure
```bash
/carto:carto-parse languages=["typescript", "python"]
```
Extracts classes, functions, dependencies, and metrics

### 3. Analyze Code Quality
```bash
/carto:carto-analyze
```
Detects:
- Design patterns (Singleton, Factory, Observer)
- Anti-patterns (God Object, Long Methods)
- Architectural violations (Layer violations)
- Code health metrics (A-F grading)
- Complexity and churn analysis

### 4. Visualize Dependencies
```bash
/carto:carto-visualize --format=deps
```
Shows import/dependency relationships in ASCII format (LLM-optimized)

### 5. Open Interactive Canvas
```bash
/carto:carto-visualize --interactive
```
Explore the code map interactively at http://localhost:5173:
- Drag nodes to rearrange
- Zoom to explore details
- Color by type/language/complexity
- Size by lines/complexity

### 6. Search and Find Code
```bash
/carto:carto-find "AuthManager" mode=map
/carto:carto-find "login" mode=source
/carto:carto-find "database" mode=all
```
Unified search across parsed code map and source files

### 7. Check Status and Changes
```bash
/carto:carto-info
/carto:carto-info --diff
```
Monitor code map status and track changes

---

## Working with Multiple Projects

### Create New Map
```bash
/carto:carto-map path="/path/to/project"
```

### Check Current Status
```bash
/carto:carto-info
```

### View Changes
```bash
/carto:carto-info --diff
```

Shows what changed in your code map

---

## Example: Real-World Workflow

### Scenario: Code Review
1. **Map** the repository
2. **Parse** to understand structure
3. **Analyze** to find patterns and issues
4. **Find** suspicious areas
5. **Visualize** in canvas for team discussion

### Scenario: Refactoring Planning
1. **Map** to understand scope
2. **Visualize** with complexity coloring (`--metric=complexity`)
3. **Analyze** to identify high-complexity areas
4. **Find** all dependencies
5. **Plan** refactoring with full context

### Scenario: New Contributor Onboarding
1. **Map** the project
2. **Open canvas** for exploration (`--interactive`)
3. **Find** entry points and key classes
4. **Visualize** architecture
5. **Understand** relationships before coding

---

## Tips & Tricks

### Visualization with Metrics
```bash
# Overlay complexity metrics
/carto:carto-visualize --format=deps --metric=complexity

# Show code churn (frequently modified)
/carto:carto-visualize --interactive --metric=churn

# Show blast radius impact
/carto:carto-visualize --format=calls --metric=blast
```

### Focused Analysis
```bash
# Only TypeScript files
/carto:carto-parse languages=["typescript"]

# Analyze specific file
/carto:carto-analyze focus="src/auth.ts"

# Find in specific language
/carto:carto-find "login" mode=source
```

### Search Patterns
```bash
# Search by type (parsed code map)
/carto:carto-find "AuthManager" mode=map type=node

# Search with regex (source code)
/carto:carto-find "def.*auth" mode=source

# Search everywhere (default)
/carto:carto-find "AuthService"
```

---

## Integration with Web UI

Access the **interactive web UI** at `http://localhost:5173`:

1. Click **"Scan Current Project"**
2. See **Code Structure** tree on the right
3. Interact with **D3.js Graph** in the center
4. Use **visualization controls** to change color/size
5. Click **nodes** to see details in the detail panel

The web UI works alongside the CLI commands for a complete code analysis experience.
