---
name: carto
description: Auto-activate when working with codebases, mapping structure, or needing code overview for LLM context
triggers:
  - "map this codebase"
  - "show me the code structure"
  - "what functions are in"
  - "class hierarchy"
  - "analyze this code"
  - "code overview"
  - "generate code graph"
  - "where is.*defined"
  - "find.*class"
  - "find.*function"
  - "search for.*pattern"
  - "state machine"
sub_skills:
  - code-mapper
  - code-visualizer
  - code-searcher
---

# Code Cartographer Skill

This skill automatically activates when working with codebases to provide structure mapping, code analysis, and visualization.

## Sub-Skills

This skill delegates to specialized sub-skills:

1. **code-mapper** - AST parsing and structure analysis
   - Parse Python, JavaScript, TypeScript
   - Extract classes, functions, methods
   - Build inheritance and call graphs

2. **code-visualizer** - Canvas and graph generation
   - Infinite canvas export
   - ASCII graph generation
   - State machine visualization

3. **code-searcher** - Fast search and query
   - Ripgrep search
   - Query code map
   - State machine detection

## Auto-Activation Triggers

The skill activates when:
- User mentions "map", "structure", "overview" of code
- User asks about classes, functions, or relationships
- User needs code context for another task
- User asks "where is X defined"
- User wants to search for patterns in code
- User asks about state machines

## Capabilities

1. **Code Mapping**: Generate complete codebase maps using AST parsing
2. **Structure Analysis**: Identify classes, functions, inheritance, dependencies
3. **Visualization**: Render interactive infinite canvas
4. **LLM Context**: Generate token-efficient ASCII/MD for prompt context
5. **Querying**: Find definitions, relationships, dependencies
6. **Code Search**: Fast ripgrep search across codebase
7. **State Machine Detection**: Identify state machine patterns
8. **Multi-Language**: Python, JavaScript, TypeScript support

## Usage Examples

- "Map this codebase" -> Delegates to code-mapper
- "Show me the canvas" -> Delegates to code-visualizer
- "Search for all state patterns" -> Delegates to code-searcher
- "Find React components" -> Delegates to appropriate sub-skill

## Parameters

- **path**: Optional path to specific directory
- **style**: Output style (summary, tree, relationships, full)
- **include_dependencies**: Whether to analyze imports

## Output Format

The skill generates:
- `code-map.toon` - Full structured code map in TOON format
- `code-graph.md` - ASCII/MD overview for LLM context
- Search results with file locations
- Canvas visualization data

## State Machine Detection

Detects these patterns:
- Enum-based states: `class State(Enum): IDLE, RUNNING`
- State variables: `self.state = "idle"`
- Match/case: `match status: case "pending":`
- Conditional chains: `if state == "x": ... elif state == "y":`
- State handlers: `on_idle()`, `handle_running()`

## Git Integration

Track changes with git:
```bash
git add .code-map/
git commit -m "Update code map"
```
