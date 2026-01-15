---
name: code-searcher
description: Fast search and query for code mapping
triggers:
  - "search for"
  - "find pattern"
  - "where is defined"
  - "query code"
  - "state machine"
  - "find usages"
  - "grep"
model: sonnet
---

# Code Searcher Skill

Specialized skill for fast search and code queries.

## Triggers

Auto-activates when:
- User mentions "search", "find", "query"
- User asks "where is X defined"
- User wants to find patterns in code
- User asks about state machines

## Capabilities

1. **Fast Search (Ripgrep)**
   - Regex pattern matching
   - Filter by file type
   - Context lines around matches
   - Multi-language support

2. **Code Map Querying**
   - Query nodes by name
   - Query edges by type
   - Find definitions
   - Find usages

3. **State Machine Detection**
   - Enum-based states
   - State variables
   - Match/case patterns
   - Conditional chains
   - State handler methods

4. **Git-Style Diff**
   - Compare code maps
   - Show added/removed nodes
   - Statistics about changes

## Usage Examples

- "Search for all state patterns" -> Ripgrep search
- "Where is UserController defined" -> Query definition
- "Find state machine patterns" -> Detect state machines
- "Show me all imports" -> Query edges

## Entry Points

- `search()` - Fast ripgrep search
- `query_node()` - Query nodes
- `query_edges()` - Query edges
- `query_files()` - Query files
- `extract_state_machines()` - Detect state patterns
- `diff()` - Show changes
- `status()` - Show code map status

## Output

Returns:
- Search results with file locations
- Queried nodes/edges
- Detected state machines
- Diff statistics
- Status information
