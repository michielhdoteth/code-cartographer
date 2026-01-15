---
name: code-visualizer
description: Canvas and graph generation for code visualization
triggers:
  - "show canvas"
  - "generate graph"
  - "visualize code"
  - "open visualization"
  - "state diagram"
  - "show graph"
  - "render canvas"
model: sonnet
---

# Code Visualizer Skill

Specialized skill for visualization output generation.

## Triggers

Auto-activates when:
- User mentions "canvas", "visualize", "graph"
- User wants to see code structure visually
- User asks for diagrams or renderings

## Capabilities

1. **Infinite Canvas Export**
   - Generate code-map-board.json
   - Position blocks for classes/functions
   - Create connectors for relationships
   - Support chunked loading for large graphs

2. **ASCII Graph Generation**
   - Generate code-graph.md
   - Tree view of code structure
   - Relationship diagrams
   - Dependency graphs

3. **State Machine Visualization**
   - Detect state machine patterns
   - Generate state diagrams
   - Show transitions between states
   - Visualize state variables

4. **Multi-Format Output**
   - TOON format for persistence
   - JSON for canvas rendering
   - Markdown for LLM context
   - SVG for static rendering

## Usage Examples

- "Show me the canvas" -> Opens visualization
- "Generate ASCII graph" -> Creates code-graph.md
- "Visualize state machine" -> Shows state diagram
- "Render code map" -> Exports to canvas format

## Entry Points

- `canvas()` - Start canvas server
- `export_canvas()` - Export to canvas format
- `generate_graph()` - Generate ASCII graph
- `get_canvas_data()` - Get visualization data

## Output

Returns:
- URL for canvas server
- JSON data for canvas rendering
- Markdown with ASCII visualization
- Statistics about visualization
