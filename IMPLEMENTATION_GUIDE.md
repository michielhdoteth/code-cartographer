# Code Cartographer v2.0 - Complete Implementation Guide

## Current Status

**Completed**: 70% - All core infrastructure and backend logic
- ✅ TypeScript project setup
- ✅ Data models with graph algorithms
- ✅ 9 language parsers
- ✅ Security scanner, pattern detector, health scorer
- ✅ React entry point with file upload
- ✅ Base styles

**Remaining**: 30% - Frontend components and visualizers
- ⏳ Advanced React components
- ⏳ D3.js visualizers (7 modes)
- ⏳ Full build & bundle
- ⏳ Claude Code plugin integration

---

## Phase 1: Complete React Components

### 1. FileTree Component (`src/components/FileTree.tsx`)

```typescript
// Features:
// - Hierarchical file browser
// - Filter by language
// - Click to select nodes
// - Expand/collapse folders

// Key Props:
// - codeMap: CodeMap
// - onSelectNode: (nodeId: string) => void
// - filter?: Language | null

// Implementation pattern:
// - Use recursive rendering for tree structure
// - Store expanded state with Set<string>
// - Color code by language type
```

### 2. DetailPanel Component (`src/components/DetailPanel.tsx`)

```typescript
// Features:
// - Show selected node information
// - Display complexity metrics
// - Show related edges (calls, dependencies)
// - Display security issues if any
// - Show health metrics for file

// Key Props:
// - codeMap: CodeMap | null
// - selectedNodeId: string | null
// - analysis?: AnalysisResult

// Display:
// - Node type, language, location
// - Metrics: LOC, cyclomatic complexity, parameters
// - Edges: incoming calls, outgoing calls, dependencies
// - Security issues in this node/file
```

### 3. SearchBar Component (`src/components/SearchBar.tsx`)

```typescript
// Features:
// - Search nodes by name
// - Filter by node type (function, class, etc.)
// - Filter by language
// - Filter by complexity range
// - Real-time highlighting

// Key Props:
// - codeMap: CodeMap
// - onSearch: (results: CodeNode[]) => void
// - onFilter: (config: VisualizationConfig) => void

// Implementation:
// - Debounce search input
// - Use fuzzy matching for names
// - Highlight matches in visualization
```

### 4. SettingsPanel Component (`src/components/SettingsPanel.tsx`)

```typescript
// Features:
// - Change visualization layout (7 modes)
// - Change coloring scheme (folder, layer, churn, blast, security, health)
// - Adjust visualization parameters
// - Toggle edges, labels
// - Node size options (uniform, size, lines, complexity)

// Key Props:
// - config: VisualizationConfig
// - onConfigChange: (config: VisualizationConfig) => void

// Controls:
// - Radio buttons for layout selection
// - Checkboxes for toggles
// - Sliders for numeric values (link distance, charge strength)
// - Dropdowns for filtering
```

### 5. HealthScore Component (`src/components/HealthScore.tsx`)

```typescript
// Features:
// - Display A-F grade prominently
// - Show metrics breakdown
// - List recommendations
// - Color code by severity

// Key Props:
// - health: HealthMetrics
// - patterns: PatternMatch[]

// Display:
// - Large grade badge with color
// - Metrics grid (dead code %, coupling, security score)
// - Actionable recommendations
// - Critical issues list
```

### 6. AnalysisResults Component (`src/components/AnalysisResults.tsx`)

```typescript
// Features:
// - Tabbed interface (Security, Patterns, Health)
// - Security issues list with severity coloring
// - Pattern detection results
// - Health recommendations
// - Click to navigate to affected nodes

// Implementation:
// - Tab component (Security | Patterns | Health)
// - Issue cards with details and suggestions
// - Severity color coding
// - Link to source location
```

---

## Phase 2: D3.js Visualizers

### Base Visualizer Pattern (`src/visualizers/BaseVisualizer.tsx`)

```typescript
// Common interface all visualizers implement:
interface Visualizer {
  render(container: SVGElement, data: VisualizerData): void
  onNodeClick?(nodeId: string): void
  onNodeHover?(nodeId: string): void
  destroy(): void
}

// VisualizerData structure:
interface VisualizerData {
  nodes: VisNode[]
  edges: VisEdge[]
  config: VisualizationConfig
}

interface VisNode {
  id: string
  name: string
  type: NodeType
  color: string
  size: number
  x?: number
  y?: number
}

interface VisEdge {
  source: string
  target: string
  type: EdgeType
  weight: number
  color: string
}
```

### 1. Force-Directed Graph (`src/visualizers/forceDirectedGraph.tsx`)

```typescript
// Default layout - physics-based simulation
// - Uses D3 force simulation
// - Drag nodes to move
// - Double-click to expand
// - Click to select

// Key Features:
// - Attractive forces between connected nodes
// - Repulsive forces between all nodes
// - Adjustable link distance and charge strength
// - Smooth animations
// - Zoom and pan support

// Implementation:
// - D3 force layout
// - SVG group for nodes and edges
// - Event handlers for interaction
// - Color by node type or file path
```

### 2. Radial Layout (`src/visualizers/radialLayout.tsx`)

```typescript
// Circular arrangement with center node
// - Root node at center
// - Related nodes in concentric circles
// - Edge distance by relationship strength

// Key Features:
// - Circle packing algorithm
// - Click node to re-center
// - Concentric layers by distance
// - Radial connection lines
```

### 3. Hierarchical Layout (`src/visualizers/hierarchicalLayout.tsx`)

```typescript
// Top-down layered layout
// - Classes at top
// - Methods/properties in middle
// - Dependencies at bottom

// Key Features:
// - Sugiyama algorithm for layering
// - Even vertical distribution
// - Minimized edge crossings
// - Clear parent-child relationships
```

### 4. Grid Layout (`src/visualizers/gridLayout.tsx`)

```typescript
// Structured grid arrangement
// - Group by node type in sections
// - Alphabetical ordering within groups

// Key Features:
// - Type-based grouping
// - Consistent spacing
// - Connected edges between grid cells
// - Suitable for structured code
```

### 5. Metro Layout (`src/visualizers/metroLayout.tsx`)

```typescript
// Subway-map style visualization
// - Nodes on rectilinear paths
// - Orthogonal edges only
// - Straight lines at cardinal directions

// Key Features:
// - Orthogonal path routing
// - Station-like node rendering
// - Color by layer/type
// - Minimal edge overlap
```

### 6. Treemap Layout (`src/visualizers/treemapLayout.tsx`)

```typescript
// Space-filling rectangles sized by LOC
// - Rectangle size = lines of code
// - Color by file path/language
// - Nested by hierarchy

// Key Features:
// - D3 treemap layout
// - Hierarchical grouping
// - Sized by metrics (LOC, complexity)
// - Click to zoom into group
```

### 7. Dependency Matrix (`src/visualizers/dependencyMatrix.tsx`)

```typescript
// Heat map matrix + Sankey diagram
// - Rows and columns for each node
// - Color intensity = dependency strength
// - Optional Sankey flow visualization

// Key Features:
// - Square matrix rendering
// - Color heat map (white to red)
// - Sankey diagram for flows
// - Click to highlight row/column
```

### Visualization Component Wrapper (`src/components/VisualizationPanel.tsx`)

```typescript
// Main component that:
// - Holds SVG container
// - Manages visualizer lifecycle
// - Switches between visualizer modes
// - Handles interactions

// Features:
// - Import all visualizers
// - Instantiate based on config.layout
// - Pass data and config
// - Handle events (click, hover, drag)
// - Cleanup on unmount
```

---

## Phase 3: Integration & Utils

### Utils Files

#### `src/utils/colors.ts`
```typescript
// Coloring functions:
// - colorByFolder(path) -> color
// - colorByLanguage(lang) -> color
// - colorBySeverity(severity) -> color
// - colorByComplexity(complexity) -> color
// - colorByHealth(grade) -> color

// Use consistent color palettes
// Generate contrasting colors for readability
```

#### `src/utils/layout.ts`
```typescript
// Layout calculation helpers:
// - positionNodesGrid(nodes)
// - positionNodesHierarchical(nodes, edges)
// - calculateNodeSize(metric)
// - generateEdgePaths(source, target, layout)

// Provide position data for each layout type
```

#### `src/utils/interactions.ts`
```typescript
// User interaction handlers:
// - zoom and pan
// - node dragging
// - tooltip positioning
// - selection highlighting
// - multi-select with Ctrl/Cmd

// Create reusable event handlers
```

---

## Phase 4: Build & Deployment

### Build Process

```bash
# 1. Install dependencies
npm install

# 2. Build for production
npm run build

# 3. Output: build/index.html (complete single-file app)
```

### Build Output

- Single HTML file with:
  - All CSS embedded
  - All JS (React + D3) bundled and minified
  - Fonts embedded
  - Can open directly in browser
  - No external dependencies needed

### Vite Configuration Notes

Current `vite.config.ts` already configured for:
- Inlining all assets
- CSS bundling
- Tree-shaking unused code
- Minification

---

## Phase 5: Claude Code Plugin Integration

### Update Commands

Modify `commands/*.md` to reference new visualization:

```markdown
/carto:carto-canvas
→ Now opens the new TSX-based single-file visualization
```

### Update Agents

Modify `agents/carto.md` to mention:
- "Uses React + D3.js for visualization"
- "Provides 7 layout modes for code graphs"
- "Includes security, pattern, and health analysis"

### Update Skills

Modify `skills/carto/SKILL.md` to reference:
- code-mapper (parsing, AST analysis)
- code-analyzer (security, patterns, health)
- code-visualizer (D3.js rendering)

---

## Development Workflow

### 1. Local Development
```bash
npm run dev
# Starts Vite dev server at http://localhost:5173
# Hot reload enabled
```

### 2. Type Checking
```bash
npm run type-check
# Ensures all TypeScript is correct
```

### 3. Building
```bash
npm run build
# Creates single HTML file in build/
```

### 4. Testing the Build
```bash
# Open build/index.html directly in browser
# Or use a simple HTTP server:
python -m http.server --directory build 8000
# Then open http://localhost:8000
```

---

## Key Implementation Tips

### 1. State Management
Use React hooks (`useState`, `useContext`) instead of Redux for simplicity:
```typescript
const [selectedNode, setSelectedNode] = useState<string | null>(null)
const [config, setConfig] = useState<VisualizationConfig>(defaultConfig)
```

### 2. Performance Optimization
For large graphs (1000+ nodes):
- Use virtualization for FileTree
- Memoize expensive calculations with `useMemo`
- Debounce search and filter operations
- Use canvas rendering for D3 if needed

### 3. Color Consistency
Define color palettes in constants:
```typescript
const COLORS = {
  languages: { js: '#f1e05a', py: '#3572A5', ... },
  severity: { critical: '#ef4444', high: '#f59e0b', ... },
  nodeTypes: { class: '#8b5cf6', function: '#06b6d4', ... }
}
```

### 4. Error Handling
Wrap parser calls:
```typescript
try {
  parser.parseAndIntegrate(file, content, codeMap)
} catch (error) {
  console.error(`Failed to parse ${file}`)
  // Show user-friendly error
}
```

---

## Expected File Structure After Completion

```
code-cartographer/
├── src/
│   ├── components/
│   │   ├── App.tsx
│   │   ├── FileTree.tsx
│   │   ├── DetailPanel.tsx
│   │   ├── SearchBar.tsx
│   │   ├── SettingsPanel.tsx
│   │   ├── HealthScore.tsx
│   │   ├── AnalysisResults.tsx
│   │   └── VisualizationPanel.tsx
│   ├── visualizers/
│   │   ├── forceDirectedGraph.tsx
│   │   ├── radialLayout.tsx
│   │   ├── hierarchicalLayout.tsx
│   │   ├── gridLayout.tsx
│   │   ├── metroLayout.tsx
│   │   ├── treemapLayout.tsx
│   │   └── dependencyMatrix.tsx
│   ├── utils/
│   │   ├── colors.ts
│   │   ├── layout.ts
│   │   └── interactions.ts
│   ├── models/
│   │   ├── types.ts
│   │   ├── codeMap.ts
│   │   └── index.ts
│   ├── parsers/
│   │   └── [9 parser files]
│   ├── analyzers/
│   │   ├── securityScanner.ts
│   │   ├── patternDetector.ts
│   │   ├── healthScorer.ts
│   │   └── index.ts
│   ├── styles.css
│   └── index.tsx
├── public/
│   └── index.html
├── build/
│   └── index.html (after build)
├── package.json
├── tsconfig.json
├── vite.config.ts
└── [config files]
```

---

## Success Criteria

- [ ] App starts with npm run dev
- [ ] Can upload code files (JS, TS, Python, Java, Go, Rust, C++, Ruby, PHP)
- [ ] Displays file tree with syntax highlighting
- [ ] Shows node details when selected
- [ ] All 7 visualizers render correctly
- [ ] Search and filter work
- [ ] Security issues displayed
- [ ] Patterns detected and shown
- [ ] Health score calculated and displayed
- [ ] Settings panel controls visualization
- [ ] Single HTML file builds successfully
- [ ] Works offline in browser

---

## Next Steps to Get Started

1. **Install dependencies**: `npm install`
2. **Start dev server**: `npm run dev`
3. **Create React components** following the patterns above
4. **Implement D3 visualizers** starting with force-directed graph
5. **Add Utils** for colors, layouts, and interactions
6. **Test with sample code files**
7. **Build production bundle**: `npm run build`
8. **Update Claude Code plugin** references
9. **Test plugin integration**

---

## Resources

- **React Documentation**: https://react.dev
- **D3.js Documentation**: https://d3js.org
- **TypeScript Handbook**: https://www.typescriptlang.org/docs
- **Vite Documentation**: https://vitejs.dev

This is your ultimate cartographer - maps made easy for both humans and AI!
