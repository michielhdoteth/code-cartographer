# Code Cartographer for Claude Code - Complete Plugin Guide

## What is Code Cartographer?

**Code Cartographer** is a specialized Claude Code plugin that maps and visualizes your codebase like a real cartographer maps geography.

Instead of Claude guessing at code structure, Code Cartographer:
- **Maps** your entire codebase architecture
- **Identifies** all relationships and dependencies
- **Visualizes** code as interactive diagrams
- **Detects** security issues automatically
- **Scores** code health with actionable insights
- **Integrates** seamlessly with Claude Code

**Why "Cartographer"?** Just as a cartographer creates accurate maps of physical terrain, Code Cartographer creates accurate maps of code structure, letting Claude understand your codebase precisely without guessing.

---

## Core Concept: Machine Diagrams & Maps

### Traditional Approach (Without Maps)
```
Claude: "What does this file do?"
You: "It's the user service..."
Claude: [guesses at structure]
Result: Incomplete understanding, wrong assumptions
```

### With Code Cartographer (Exact Maps)
```
Claude: /carto:carto-analyze
Code Cartographer: [generates complete map]
Result: Claude has exact structure, relationships, issues
```

**The Difference**: Claude now has **machine-readable diagrams** instead of guessing.

---

## How It Works

### The Pipeline

```
Your Code
    ↓
[Parser Layer] - 9 language AST extractors
    ↓
[Map Generation] - Creates CodeMap with nodes & edges
    ↓
[Analysis Layer] - Security, patterns, health scoring
    ↓
[Visualization] - Interactive diagrams via D3.js
    ↓
Claude Understanding - Precise, accurate, complete
```

### What Gets Mapped

**Nodes** (Code Elements):
- Classes, functions, methods
- Interfaces, traits, modules
- Variables, properties, constants

**Edges** (Relationships):
- Inheritance (extends, implements)
- Dependencies (imports, uses)
- Calls (function calls, method calls)
- Nesting (package structure)

**Metrics** (Quality Data):
- Complexity (cyclomatic, cognitive)
- Lines of code
- Coupling (fan-in, fan-out)
- Health scores

---

## Using Code Cartographer in Claude Code

### Quick Start

```bash
# 1. Scan your project
/carto:carto-scan /path/to/project

# 2. Parse the code
/carto:carto-parse

# 3. Analyze everything
/carto:carto-analyze

# 4. View results
/carto:carto-health
/carto:carto-canvas
```

### Available Commands

#### 1. Initialize
```bash
/carto:carto-init /path/to/project
```
Sets up the cartographer for a new project.
- Creates `.code-map/` directory
- Initializes metadata
- Prepares for scanning

#### 2. Scan
```bash
/carto:carto-scan /path/to/project
```
Discovers all code files.
- Finds all supported languages
- Indexes file structure
- Prepares for parsing
- Returns: List of files found

#### 3. Parse
```bash
/carto:carto-parse
```
Extracts code structure from all files.
- Runs language-specific parsers
- Builds AST for each file
- Creates CodeNodes and CodeEdges
- Returns: `code-map.toon` file with complete structure

#### 4. Analyze
```bash
/carto:carto-analyze
```
Runs all analysis engines.
- Security scanning
- Pattern detection
- Health scoring
- Returns: Analysis results with recommendations

#### 5. Health
```bash
/carto:carto-health
```
Shows code health metrics.
- Grade: A-F
- Score: 0-100
- Metrics breakdown
- Recommendations

```
Grade: B
Score: 78/100

Issues Found:
- 5% dead code
- 1 circular dependency
- 3 security issues (1 critical)
- 2 complex functions
- 1 God Object
```

#### 6. Canvas
```bash
/carto:carto-canvas
```
Opens interactive visualization.
- Force-directed graph
- Multiple layout modes
- Zoom and pan
- Click nodes for details
- 7 different views

#### 7. Query
```bash
/carto:carto-query function:getUserById
/carto:carto-query type:class language:python
/carto:carto-query calls --from=api.ts
```
Search and query the code map.
- Find specific nodes
- Track dependencies
- Show blast radius

#### 8. Detect
```bash
/carto:carto-detect
```
Auto-detect frameworks and libraries.
- Django, Flask, FastAPI (Python)
- React, Vue, Angular (JavaScript)
- Spring, Hibernate (Java)
- And 100+ more frameworks

#### 9. Diff
```bash
/carto:carto-diff HEAD~1
```
Show what changed in code structure.
- New nodes added
- Dependencies changed
- Breaking changes

#### 10. Graph
```bash
/carto:carto-graph --style=call-graph
/carto:carto-graph --style=dependency-tree
/carto:carto-graph --style=class-hierarchy
```
Generate specific diagram types.
- Call graphs
- Dependency trees
- Class hierarchies
- State machines

---

## Real-World Usage Examples

### Example 1: Understand a New Codebase

**Scenario**: You're joining a project and need to understand it fast.

```bash
# Step 1: Scan the project
/carto:carto-scan /path/to/new-project

# Step 2: Parse to get structure
/carto:carto-parse

# Result: Complete map of all code

# Step 3: Claude, analyze the architecture
Claude: "Based on the code map, here's the architecture:
- 24 classes organized in 3 layers
- 156 functions with average complexity 3.2
- 8 critical dependencies from UI to database
- 12% dead code mainly in legacy module
- Health score: C (needs refactoring)"
```

### Example 2: Code Review with Confidence

**Scenario**: Reviewing a PR and need to know impact.

```bash
# Generate map for the PR changes
/carto:carto-parse --branch=feature/new-api

# Query the blast radius
/carto:carto-query blast-radius --file=user-service.ts

# Result:
# Files that will be affected by changes to user-service.ts:
# - api-handler.ts (direct call)
# - auth-middleware.ts (depends on exports)
# - main.ts (loads module)
# - 5 test files (import for testing)

Claude: "This PR modifies core user service - will affect 8 files.
Here are the key impacts:
- API responses change (update client code)
- Auth flow affected (test middleware)
- 3 tests will need updates"
```

### Example 3: Security Audit

**Scenario**: Need to find security issues fast.

```bash
# Run analysis
/carto:carto-analyze

# Show security issues
Claude: "Security scan complete. Found 8 issues:

CRITICAL (1):
- Hardcoded API key in config.js line 42

HIGH (3):
- SQL injection risk in user-query.py
- XSS vulnerability in render.ts
- Command execution in deployment.sh

MEDIUM (4):
- Weak cryptography (MD5)
- Debug statements in production
- Missing input validation
- Unescaped user input

Recommendations:
1. Immediately move API key to env vars
2. Use prepared statements for SQL
3. Sanitize HTML output
4. Replace MD5 with SHA-256"
```

### Example 4: Refactoring Planning

**Scenario**: Need to refactor but want to know impact.

```bash
# Analyze the problematic module
/carto:carto-query type:class --complexity:>10

# Result: Shows classes with high complexity

Claude: "Found 3 God Objects to refactor:

UserService (15 methods):
- Handles authentication (5 methods) → extract to AuthService
- Manages profiles (4 methods) → extract to ProfileService
- Handles permissions (6 methods) → extract to PermissionService

This refactoring will:
- Reduce average complexity from 4.2 to 2.1
- Improve health score from C to B
- Make testing easier (isolated concerns)"
```

### Example 5: Performance Analysis

**Scenario**: App is slow, need to find bottlenecks.

```bash
# Get coupling metrics
/carto:carto-analyze

# Show high-coupling modules
Claude: "High coupling detected:

Main bottle necks:
- UserController has 12 dependencies (high fan-in)
- DatabaseLayer depends on 8 modules (high fan-out)
- CacheManager is used everywhere (tight coupling)

Recommendations:
1. Inject dependencies into UserController
2. Move DatabaseLayer abstraction higher
3. Create facade for CacheManager
4. Add service locator pattern"
```

---

## Understanding the Map Output

### The Code Map Structure

```typescript
CodeMap {
  id: "map_1705417200000"
  name: "my-project"

  nodes: [
    {
      id: "auth.js:class:AuthManager",
      name: "AuthManager",
      type: "class",
      language: "javascript",
      metrics: {
        linesOfCode: 156,
        cyclomatic: 4,
        nesting: 2
      }
    },
    // ... 44 more nodes
  ]

  edges: [
    {
      source: "api.ts:class:APIHandler",
      target: "auth.js:class:AuthManager",
      type: "imports"
    },
    // ... 119 more edges
  ]
}
```

### Statistics Summary

```
Total Nodes: 45
├─ Classes: 8
├─ Functions: 15
├─ Methods: 20
└─ Variables: 2

Total Edges: 120
├─ Imports: 30
├─ Calls: 50
├─ Inherits: 10
└─ Depends: 30

Languages: 5
├─ JavaScript: 15 nodes
├─ TypeScript: 12 nodes
├─ Python: 8 nodes
├─ Java: 6 nodes
└─ Go: 4 nodes

Health: B (78/100)
```

---

## Map Exports & Sharing

### Export Formats

#### 1. JSON (for storage/sharing)
```bash
/carto:carto-export --format=json --file=map.json
```
Complete machine-readable map for:
- Storing in database
- Sharing with team
- Processing by other tools
- Feeding to Claude

#### 2. Markdown (for documentation)
```bash
/carto:carto-export --format=markdown --file=architecture.md
```
Human-readable documentation:
- Project overview
- Architecture diagram (as ASCII)
- Module relationships
- Health recommendations

#### 3. D3.js Visualization
```bash
/carto:carto-canvas --export=html
```
Interactive HTML visualization:
- Pan and zoom
- Click nodes for details
- Multiple layout modes
- Exportable diagram

#### 4. Diff Report
```bash
/carto:carto-export --format=diff --from=HEAD~1
```
What changed:
- New modules added
- Deleted modules
- Relationship changes
- Complexity changes

---

## Claude Using the Map

### How Claude Understands Your Code

**Without Map** (Traditional):
```
You: "What's the architecture?"
Claude: "Based on file names, I see services,
controllers, and... probably a database layer?
But I'm guessing about the exact structure."
```

**With Code Cartographer Map** (Exact):
```
You: /carto:carto-parse
Claude: "I can see the exact structure:
- 8 classes in a 3-layer architecture
- 156 functions with their exact complexity
- 120 specific relationships between components
- 5% dead code in exactly which functions
- Health score: B with specific issues to fix"
```

### Claude's Accurate Insights

With the map, Claude can:

✅ **Architecture Understanding**
- Know exact layer boundaries
- Understand module dependencies
- See coupling points

✅ **Code Quality**
- Know complexity of each function
- Identify dead code precisely
- Find refactoring opportunities

✅ **Impact Analysis**
- Calculate blast radius exactly
- Predict test coverage needs
- Know backward compatibility issues

✅ **Performance**
- Identify coupling bottlenecks
- Find circular dependencies
- Suggest architectural improvements

✅ **Security**
- Know exact vulnerability locations
- Understand data flow paths
- Identify sensitive function exposure

---

## Integration Workflow

### Complete Workflow: From Code to Insight

```
1. YOUR CODE
   ↓
2. /carto:carto-init
   Initialize cartographer
   ↓
3. /carto:carto-scan
   Find all code files (45 files found)
   ↓
4. /carto:carto-parse
   Extract structure (45 nodes, 120 edges)
   ↓
5. /carto:carto-analyze
   Run security, patterns, health
   ↓
6. CLAUDE HAS THE MAP
   ├─ Knows exact structure
   ├─ Sees all relationships
   ├─ Understands security issues
   ├─ Calculates health metrics
   └─ Can make precise recommendations
   ↓
7. /carto:carto-health
   Show recommendations
   ↓
8. /carto:carto-canvas
   Visualize the map
   ↓
9. ACTIONABLE INSIGHTS
   ├─ Specific security fixes
   ├─ Refactoring recommendations
   ├─ Architecture improvements
   ├─ Dead code removal
   └─ Performance optimizations
```

---

## Command Reference

| Command | Purpose | Output |
|---------|---------|--------|
| `carto-init` | Initialize for project | `.code-map/` directory |
| `carto-scan` | Find code files | File list and count |
| `carto-parse` | Extract code structure | `code-map.toon` with all nodes/edges |
| `carto-analyze` | Run all analysis | Security, patterns, health |
| `carto-health` | Show health score | Grade, metrics, recommendations |
| `carto-canvas` | Visualize code | Interactive D3.js diagram |
| `carto-query` | Search the map | Query results |
| `carto-detect` | Find frameworks | Framework list |
| `carto-diff` | Show changes | Diff report |
| `carto-graph` | Generate diagrams | Specific visualization |

---

## Benefits for Claude

### 1. Accuracy
- No guessing at code structure
- Exact node and edge data
- Precise complexity metrics
- Verified relationships

### 2. Speed
- Analysis takes seconds
- Results ready immediately
- No manual exploration needed
- Quick recommendations

### 3. Completeness
- Every file analyzed
- Every relationship mapped
- All issues detected
- Nothing missed

### 4. Actionability
- Specific file locations
- Line numbers for issues
- Blast radius calculated
- Impact understood

### 5. Confidence
- Data-driven insights
- Proven patterns
- Verifiable metrics
- No assumptions

---

## Examples: Before vs After

### Before (Without Maps)

**Question**: "Is this code healthy?"

**Claude's Answer**: "Based on the file names and what I can see, it looks... okay? There might be some circular dependencies, but I can't tell for sure."

**Problem**: Guessing, incomplete, uncertain

### After (With Maps)

**Question**: "Is this code healthy?"

**Claude's Answer**:
```
Grade: B (78/100)

Issues:
✗ 5.2% dead code (3 unused functions)
✗ 1 circular dependency (UserService ↔ AuthService)
✗ 5 security issues (1 critical: hardcoded token)
✗ 3 complex functions (> 10 cyclomatic complexity)
✗ 1 God Object (UserService with 15 methods)

Recommendations:
1. Remove 3 unused functions
2. Break circular dependency using dependency injection
3. Move hardcoded token to environment variables
4. Refactor complex functions and split God Object
5. Expected improvement: Grade A (95/100)
```

**Benefits**: Specific, complete, actionable, verifiable

---

## Setup Instructions

### 1. Install Code Cartographer

The plugin is already configured in your project:
```bash
cat .claude-plugin/plugin.json
```

### 2. Enable the Plugin

In Claude Code settings, enable "carto" plugin:
```
Settings → Plugins → Code Cartographer → Enable
```

### 3. Start Using

```bash
/carto:carto-init /path/to/your/project
```

### 4. Explore Commands

```bash
# Show all available commands
/carto:help

# Get command details
/carto:carto-scan --help
```

---

## Tips & Tricks

### Tip 1: Quick Health Check
```bash
/carto:carto-parse && /carto:carto-health
```

### Tip 2: Focus on Issues
```bash
/carto:carto-analyze --severity=critical
```
Shows only critical issues.

### Tip 3: Compare Branches
```bash
/carto:carto-diff main --branch=feature/refactor
```
See what changed.

### Tip 4: Targeted Analysis
```bash
/carto:carto-query security
/carto:carto-query patterns
/carto:carto-query dead-code
```

### Tip 5: Export for Sharing
```bash
/carto:carto-export --format=markdown
```
Creates shareable report.

---

## Troubleshooting

### "Parser not found for language X"
**Solution**: Language might not be supported yet. Supported: JS, TS, Python, Java, Go, Rust, C++, Ruby, PHP

### "Too many nodes in visualization"
**Solution**: Use `--filter` to focus on specific modules or layer

### "Security scan found too many issues"
**Solution**: Use `--severity=critical` to prioritize

### "Map generation is slow"
**Solution**: Run on a subset first with `--include-path`

---

## Integration with Claude's Analysis

### Claude Can Now:

1. **Understand Architecture Instantly**
   ```
   "I can see this is a 3-layer MVC architecture with
   24 classes across 5 modules. The database layer is
   tightly coupled to 8 other modules, which might
   cause performance issues."
   ```

2. **Identify Exact Problems**
   ```
   "UserService.java line 145 has a circular dependency
   with AuthService.java. This will cause initialization
   issues. Here's how to fix it..."
   ```

3. **Calculate Precise Impact**
   ```
   "Changing the User model will affect 12 files:
   - 3 service files (direct)
   - 4 controller files (indirect)
   - 5 test files (tests)
   These are the exact changes needed..."
   ```

4. **Make Data-Driven Decisions**
   ```
   "Your health score is B. To reach A, you need to:
   1. Remove 3 unused functions (8 min)
   2. Refactor GetUserById (12 min, complexity 14→4)
   3. Break circular dependency (20 min)
   Total effort: 40 minutes for +17 points."
   ```

5. **Provide Security Insights**
   ```
   "Found hardcoded API key in config.js:42.
   This is exposed in 3 places. Here's the fix...
   Also: SQL injection risk in database.py:156
   detected by pattern matching SQL string concatenation."
   ```

---

## Real Value

### Before Code Cartographer
- Claude guesses at structure ❌
- Misses issues ❌
- Provides uncertain recommendations ❌
- Can't calculate impact ❌
- Incomplete analysis ❌

### After Code Cartographer
- Claude knows exact structure ✅
- Finds all issues ✅
- Provides confident recommendations ✅
- Calculates precise impact ✅
- Complete, actionable analysis ✅

---

## Summary

**Code Cartographer** is your codebase's map and Claude is the cartographer's assistant.

Together, they:
- Map your entire codebase
- Identify all issues precisely
- Visualize relationships clearly
- Understand impact exactly
- Provide actionable recommendations

**No more guessing. Just facts, maps, and insights.**

---

## Next Steps

1. **Try it**: `/carto:carto-init /your/project`
2. **Scan**: `/carto:carto-scan`
3. **Parse**: `/carto:carto-parse`
4. **Analyze**: `/carto:carto-analyze`
5. **View**: `/carto:carto-health`
6. **Visualize**: `/carto:carto-canvas`

**Welcome to precise code understanding with Code Cartographer!** 🗺️

---

## Questions?

```bash
/carto:help
/carto:carto-query --help
/carto:carto-parse --help
```

Or ask Claude: "How would you analyze this with the code map?"

**Code Cartographer: Because precision matters.**
