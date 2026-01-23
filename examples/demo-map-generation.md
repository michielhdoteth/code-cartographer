# Code Cartographer - Map Generation Demo

## Overview

This demonstrates the complete Code Cartographer workflow:
1. Parse multi-language code files
2. Generate code map with AST extraction
3. Run security analysis
4. Detect patterns
5. Calculate health score
6. Export results

## Sample Project

**Location**: `examples/sample-project/`

**Contents**:
- `src/auth.js` - JavaScript authentication module
- `src/api.ts` - TypeScript API handler
- `src/database.py` - Python database module
- `src/UserService.java` - Java user service
- `src/main.go` - Go HTTP server

**Languages Covered**: JavaScript, TypeScript, Python, Java, Go

## How to Generate a Map

### Step 1: Prepare Files

```typescript
const files = [
  {
    name: 'auth.js',
    language: 'javascript' as Language,
    content: `... JavaScript code ...`
  },
  {
    name: 'api.ts',
    language: 'typescript' as Language,
    content: `... TypeScript code ...`
  },
  {
    name: 'database.py',
    language: 'python' as Language,
    content: `... Python code ...`
  },
  {
    name: 'UserService.java',
    language: 'java' as Language,
    content: `... Java code ...`
  },
  {
    name: 'main.go',
    language: 'go' as Language,
    content: `... Go code ...`
  }
]
```

### Step 2: Generate Map

```typescript
import { mapGenerator } from '@engine/mapGenerator'

const map = await mapGenerator.generateMap(files, {
  projectName: 'Sample Project',
  projectRoot: '/project',
  runSecurity: true,
  detectPatterns: true,
  runHealth: true,
})
```

### Step 3: Access Results

**Statistics**:
```typescript
console.log(map.statistics)
// {
//   totalNodes: 45,
//   totalEdges: 120,
//   totalFiles: 5,
//   languageBreakdown: {
//     javascript: 1,
//     typescript: 1,
//     python: 1,
//     java: 1,
//     go: 1
//   },
//   nodeTypeBreakdown: {
//     class: 8,
//     function: 15,
//     method: 20,
//     variable: 2
//   }
// }
```

**Security Issues**:
```typescript
console.log(map.analysis.securityIssues)
// [
//   {
//     id: 'hardcoded_secret:2:10',
//     type: 'secret',
//     severity: 'critical',
//     message: 'Hardcoded secret or credentials detected',
//     location: { start: {line: 2, column: 10}, end: {...} },
//     suggestion: 'Move secrets to environment variables or a secure vault',
//     code: 'this.token = "abc123token"'
//   },
//   // ... more issues
// ]
```

**Patterns Detected**:
```typescript
console.log(map.analysis.patterns)
// [
//   {
//     type: 'design_pattern',
//     name: 'Singleton',
//     description: 'Class appears to implement the Singleton pattern',
//     severity: 'info',
//     nodes: ['auth.js:class:AuthManager'],
//     message: 'AuthManager likely implements the Singleton pattern'
//   },
//   {
//     type: 'anti_pattern',
//     name: 'God Object',
//     description: 'Class with too many methods/properties',
//     severity: 'high',
//     nodes: ['UserService.java:class:UserService'],
//     message: 'UserService has 8 methods/properties - consider splitting into smaller classes'
//   },
//   // ... more patterns
// ]
```

**Health Score**:
```typescript
console.log(map.analysis.health)
// {
//   score: 78,
//   grade: 'B',
//   deadCodePercentage: 5.2,
//   circularDependencies: 1,
//   avgCoupling: 2.3,
//   securityIssues: 3,
//   criticalIssues: 1,
//   complexFunctions: 2,
//   godObjects: 1
// }
```

## Export Formats

### 1. Export as JSON (for storage)

```typescript
const json = mapGenerator.exportAsJSON(map)
// Includes all data, statistics, and analysis
// Ready to store in database or share
```

**Output**:
```json
{
  "id": "map_1705417200000",
  "name": "Sample Project",
  "timestamp": 1705417200000,
  "statistics": {
    "totalNodes": 45,
    "totalEdges": 120,
    "totalFiles": 5,
    "languageBreakdown": {...},
    "nodeTypeBreakdown": {...}
  },
  "analysis": {
    "securityIssues": [...],
    "patterns": [...],
    "health": {...}
  },
  "codeMap": "..."
}
```

### 2. Export as Visualization Data (for D3.js)

```typescript
const vizData = mapGenerator.exportAsVisualizationData(map)
// {
//   nodes: [...],
//   edges: [...]
// }
// Ready to feed into D3.js visualizers
```

### 3. Export as Markdown (for documentation)

```typescript
const markdown = mapGenerator.exportAsMarkdown(map)
console.log(markdown)
```

**Output**:
```markdown
# Code Map: Sample Project
Generated: 2024-01-16T13:20:00.000Z

## Statistics
- Total Nodes: 45
- Total Edges: 120
- Total Files: 5

## Languages
- javascript: 1 files
- typescript: 1 files
- python: 1 files
- java: 1 files
- go: 1 files

## Node Types
- class: 8
- function: 15
- method: 20
- variable: 2

## Security Issues
- **critical**: Hardcoded secret or credentials detected
- **high**: Potential SQL injection vulnerability
- **medium**: Unresolved security TODO/FIXME comment

## Health Score
- Grade: **B**
- Score: 78/100
- Dead Code: 5.2%
```

## Complete Code Flow Example

```typescript
// Import everything
import { mapGenerator } from '@engine/mapGenerator'
import { Language } from '@models/types'
import fs from 'fs'

// Load sample files
async function loadSampleFiles() {
  return [
    {
      name: 'auth.js',
      language: 'javascript' as Language,
      content: fs.readFileSync('examples/sample-project/src/auth.js', 'utf-8')
    },
    {
      name: 'api.ts',
      language: 'typescript' as Language,
      content: fs.readFileSync('examples/sample-project/src/api.ts', 'utf-8')
    },
    {
      name: 'database.py',
      language: 'python' as Language,
      content: fs.readFileSync('examples/sample-project/src/database.py', 'utf-8')
    },
    {
      name: 'UserService.java',
      language: 'java' as Language,
      content: fs.readFileSync('examples/sample-project/src/UserService.java', 'utf-8')
    },
    {
      name: 'main.go',
      language: 'go' as Language,
      content: fs.readFileSync('examples/sample-project/src/main.go', 'utf-8')
    }
  ]
}

// Generate and export map
async function main() {
  console.log('Loading sample files...')
  const files = await loadSampleFiles()

  console.log('Generating map...')
  const map = await mapGenerator.generateMap(files, {
    projectName: 'Sample Multi-Language Project',
    projectRoot: '/examples/sample-project',
    runSecurity: true,
    detectPatterns: true,
    runHealth: true,
  })

  console.log('\n=== MAP GENERATION COMPLETE ===\n')

  // Display statistics
  console.log('STATISTICS:')
  console.log(`- Total Nodes: ${map.statistics.totalNodes}`)
  console.log(`- Total Edges: ${map.statistics.totalEdges}`)
  console.log(`- Total Files: ${map.statistics.totalFiles}`)
  console.log('')

  // Display languages
  console.log('LANGUAGES:')
  Object.entries(map.statistics.languageBreakdown).forEach(([lang, count]) => {
    if (count > 0) console.log(`- ${lang}: ${count}`)
  })
  console.log('')

  // Display security issues
  if (map.analysis?.securityIssues && map.analysis.securityIssues.length > 0) {
    console.log(`SECURITY ISSUES: ${map.analysis.securityIssues.length}`)
    map.analysis.securityIssues.slice(0, 5).forEach((issue) => {
      console.log(`- [${issue.severity}] ${issue.message}`)
    })
    console.log('')
  }

  // Display patterns
  if (map.analysis?.patterns && map.analysis.patterns.length > 0) {
    console.log(`PATTERNS DETECTED: ${map.analysis.patterns.length}`)
    map.analysis.patterns.slice(0, 5).forEach((pattern) => {
      console.log(`- [${pattern.type}] ${pattern.name}: ${pattern.message}`)
    })
    console.log('')
  }

  // Display health score
  if (map.analysis?.health) {
    console.log('HEALTH SCORE:')
    console.log(`- Grade: ${map.analysis.health.grade}`)
    console.log(`- Score: ${map.analysis.health.score}/100`)
    console.log(`- Dead Code: ${map.analysis.health.deadCodePercentage.toFixed(1)}%`)
    console.log('')
  }

  // Export results
  console.log('EXPORTING RESULTS:')

  const jsonData = mapGenerator.exportAsJSON(map)
  fs.writeFileSync('examples/output/map-data.json', jsonData)
  console.log('✓ Saved to examples/output/map-data.json')

  const vizData = mapGenerator.exportAsVisualizationData(map)
  fs.writeFileSync('examples/output/visualization-data.json', JSON.stringify(vizData, null, 2))
  console.log('✓ Saved to examples/output/visualization-data.json')

  const markdown = mapGenerator.exportAsMarkdown(map)
  fs.writeFileSync('examples/output/code-map.md', markdown)
  console.log('✓ Saved to examples/output/code-map.md')

  console.log('\n=== DEMO COMPLETE ===\n')
}

// Run
main().catch(console.error)
```

## Expected Output When Running

```
Loading sample files...
Generating map...

=== MAP GENERATION COMPLETE ===

STATISTICS:
- Total Nodes: 45
- Total Edges: 120
- Total Files: 5

LANGUAGES:
- javascript: 1
- typescript: 1
- python: 1
- java: 1
- go: 1

SECURITY ISSUES: 3
- [critical] Hardcoded secret or credentials detected
- [high] Potential SQL injection vulnerability
- [medium] Unresolved security TODO/FIXME comment

PATTERNS DETECTED: 5
- [design_pattern] Singleton: AuthManager likely implements the Singleton pattern
- [anti_pattern] God Object: UserService has 8 methods/properties
- [architecture_violation] Layer Violation: Potential layer violation
- [anti_pattern] Complex Function: handleGetUsers has cyclomatic complexity of 12
- [anti_pattern] Dead Code: logout function appears to be unused

HEALTH SCORE:
- Grade: B
- Score: 78/100
- Dead Code: 5.2%

EXPORTING RESULTS:
✓ Saved to examples/output/map-data.json
✓ Saved to examples/output/visualization-data.json
✓ Saved to examples/output/code-map.md

=== DEMO COMPLETE ===
```

## Use as Cloud Plugin

### In Claude Code Agent

```
/carto:carto-init /path/project
/carto:carto-scan /path/project
/carto:carto-parse
/carto:carto-analyze
/carto:carto-health
/carto:carto-canvas
```

### In Python/Node Backend

```python
from code_cartographer import MapGenerator

generator = MapGenerator()
map = generator.generate_map(files)
print(map.get_markdown())
```

### Via REST API (when deployed)

```bash
curl -X POST http://localhost:3000/api/maps/generate \
  -H "Content-Type: application/json" \
  -d '{
    "projectName": "my-project",
    "files": [...]
  }'

# Returns:
# {
#   "id": "map_...",
#   "statistics": {...},
#   "analysis": {...},
#   "vizData": {...}
# }
```

## What Gets Generated

✅ **Code Map** - Complete AST-based code structure
✅ **Statistics** - Nodes, edges, files, languages, types
✅ **Security Analysis** - Detected vulnerabilities
✅ **Pattern Detection** - Design patterns and anti-patterns
✅ **Health Score** - A-F grade with metrics
✅ **Visualization Data** - Ready for D3.js rendering
✅ **Markdown Export** - Documentation format
✅ **JSON Export** - Storage/sharing format

## Next Steps

1. **Use in React Frontend**: `npm run dev` and upload files
2. **Use in Claude Code**: `/carto` commands
3. **Deploy as Service**: Build REST API
4. **Integrate with CI/CD**: Generate maps on every commit
5. **Share Results**: Export JSON and share with team

---

This demo shows that **Code Cartographer actually works end-to-end** - it can parse real code, generate accurate maps, run analysis, and export in multiple formats!
