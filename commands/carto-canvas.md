---
description: Open the infinite canvas visualization
parameters:
  port:
    type: number
    description: Port for canvas server
    required: false
    default: 8080
---

# carto-canvas

Open the infinite canvas visualization in a browser.

## Usage

```
/carto:carto-canvas
/carto:carto-canvas port=9000
```

## What Happens

1. Exports code map to canvas format (code-map-board.json)
2. Starts canvas server on specified port
3. Opens browser to the canvas URL
4. Displays code map as interactive visualization

## Prerequisites

Run these commands first to build the code map:
```
/carto:carto-init path="."
/carto:carto-scan
/carto:carto-parse
```

## Canvas Features

- **Pan and zoom**: Scroll to zoom, drag to pan
- **Block-based visualization**: Classes and functions as blocks
- **SVG connectors**: Shows relationships between elements
- **Minimap navigation**: Overview of entire map
- **Connection filtering**: Filter by relationship type

## Controls

| Control | Action |
|---------|--------|
| Scroll wheel | Zoom in/out |
| Click + drag | Pan around |
| Click block | View details |
| Connection mode dropdown | Filter relationships |

## Manual Start

```bash
python serve_canvas.py --port 8080 --dir .code-map
# Open http://localhost:8080/infinite-canvas.html
```

## Example

```
/carto:carto-canvas

Output:
Canvas server starting at http://localhost:8080/infinite-canvas.html
```
