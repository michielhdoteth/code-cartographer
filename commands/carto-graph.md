---
description: Generate code visualizations (ASCII, call graph, dependency graph, class hierarchy, state machines)
parameters:
  style:
    type: string
    description: Graph style
    required: false
    default: overview
    enum: [overview, tree, relationships, dependencies, calls, hierarchy, summary, full, call_graph, dependency_graph, class_hierarchy, state_machines]
  focus:
    type: string
    description: Focus node for call graph (function name or ID)
    required: false
  force:
    type: boolean
    description: Overwrite existing file
    required: false
    default: false
---

# carto-graph

Generate code visualizations in multiple formats for LLM prompts and interactive viewing.

## Usage

```
/carto:carto-graph
/carto:carto-graph style=tree
/carto:carto-graph style=call_graph focus=process_data
/carto:carto-graph style=dependency_graph
/carto:carto-graph style=class_hierarchy
/carto:carto-graph style=state_machines
/carto:carto-graph style=hierarchy force=true
```

## ASCII Styles (LLM Context)

| Style | Description | Output |
|-------|-------------|--------|
| `overview` | Statistics and file list | code-graph.md |
| `tree` | Class/function hierarchy | code-graph.md |
| `relationships` | Inheritance and calls | code-graph.md |
| `dependencies` | Import dependencies | code-graph.md |
| `calls` | Call graph | code-graph.md |
| `hierarchy` | Class inheritance tree | code-graph.md |
| `summary` | Quick overview for LLMs | code-graph.md |
| `full` | Complete report | code-graph.md |

## Canvas Visualizations (Interactive)

| Style | Description | Output File |
|-------|-------------|-------------|
| `call_graph` | Interactive call graph visualization | call-graph.json |
| `dependency_graph` | Interactive dependency visualization | dependency-graph.json |
| `class_hierarchy` | Interactive inheritance tree | class-hierarchy.json |
| `state_machines` | State machine diagrams | state-machines.json |

## Graph Types Explained

### Call Graph
Shows which functions call which other functions. Useful for:
- Understanding control flow
- Finding critical paths
- Identifying hot functions

### Dependency Graph
Shows module/file dependencies via imports. Useful for:
- Understanding architecture
- Finding coupling
- Planning refactoring

### Class Hierarchy
Shows inheritance relationships. Useful for:
- Understanding class design
- Finding base classes
- Identifying design patterns

### State Machines
Shows state patterns in Python code. Detects:
- Enum-based states
- Match/case patterns
- State variables
- State handler methods

## Output

ASCII graphs are saved to `.code-map/code-graph.md`.
Canvas visualizations are saved to `.code-map/*.json`.

## Example: Call Graph

```
Function Call Graph:
main() calls init() calls setup() calls configure()
setup() calls load_config() calls validate()
configure() calls apply_settings() calls start_services()
```

## Example: State Machines Detected

```json
{
  "name": "StateMachine:OrderProcessor",
  "states": ["pending", "processing", "shipped", "delivered", "cancelled"],
  "transitions": [
    ["pending", "submit", "processing"],
    ["processing", "ship", "shipped"],
    ["shipped", "deliver", "delivered"]
  ]
}
```
