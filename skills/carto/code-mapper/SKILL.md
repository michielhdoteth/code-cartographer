---
name: code-mapper
description: AST parsing and structure analysis for code mapping
triggers:
  - "map this codebase"
  - "parse code"
  - "analyze structure"
  - "find classes"
  - "find functions"
  - "extract nodes"
  - "build graph"
  - "show inheritance"
model: sonnet
---

# Code Mapper Skill

Specialized skill for AST parsing and structure analysis.

## Triggers

Auto-activates when:
- User mentions "map", "parse", "analyze" of code structure
- User asks about classes, functions, or relationships
- User wants to extract nodes from code

## Capabilities

1. **AST Parsing**
   - Parse Python files with Python AST
   - Parse JavaScript files
   - Parse TypeScript files
   - Extract full syntax tree

2. **Structure Extraction**
   - Extract classes, functions, methods
   - Identify inheritance relationships
   - Track imports and dependencies
   - Extract docstrings and comments

3. **Relationship Building**
   - Build contains relationships
   - Build inheritance edges
   - Build call graphs
   - Track parent/child relationships

4. **Multi-Language Support**
   - Python: Full AST support with complexity analysis
   - JavaScript: Classes, functions, arrow functions
   - TypeScript: Interfaces, types, generics

## Usage Examples

- "Map this codebase" -> Creates complete code map
- "Parse all Python files" -> Extract AST nodes
- "Find all controller classes" -> Query structure
- "Show inheritance hierarchy" -> Build inheritance graph

## Entry Points

- `init()` - Initialize a new code map
- `scan()` - Scan files in directory
- `parse()` - Parse files and extract nodes
- `full()` - Run complete workflow

## Output

Returns:
- CodeMap with nodes and edges
- Statistics about codebase
- File listings
- Relationship data
