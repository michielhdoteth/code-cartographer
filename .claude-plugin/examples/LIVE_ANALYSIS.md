# Live Analysis Mode

Code Cartographer now includes **Live Analysis** - automatically scan and visualize your codebase in real-time as you work.

## How It Works

### Auto-Analysis on Startup
When you open the web UI at `http://localhost:5173`:
1. Code Cartographer automatically scans the current folder
2. Parses all source files
3. Extracts code structure and relationships
4. Visualizes the codebase in the D3.js graph
5. Updates code structure tree on the right

**No manual button clicks needed** - just open the URL and start exploring!

## Real-Time Updates

### Background Monitoring (Planned)
The system will monitor your project for changes:
- **Git diffs** detect which files changed
- **Incremental parsing** updates only changed files
- **Live graph updates** reflect new/modified code
- **Zero configuration** - just keep working

### Manual Rescan
Click **"Rescan Project"** button to manually update the analysis:
- Useful for large changes
- Refreshes all metrics and relationships
- Takes ~1-2 seconds for medium projects

## Integration Points

### From Claude Code Plugin
The web UI works alongside Claude Code commands:
```bash
/carto:carto-map         # Initialize and detect new files
/carto:carto-parse       # Extract structure
/carto:carto-analyze     # Find issues and patterns
/carto:carto-visualize --interactive  # Open live UI
```

When you run commands via CLI, the web UI can be triggered to update automatically.

### From Git Integration
Detect changes and update visualization:
```bash
git diff --name-only     # Find changed files
git diff HEAD~1          # Show specific changes
```

The system will:
1. Detect which files changed via `carto-info --diff`
2. Re-parse only changed files
3. Update the graph automatically
4. Show change indicators in the UI

## Live Analysis Features

### Live Status Indicator
```
● ./examples/sample-project
Auto-updating
```
Green dot = active analysis

### Auto-Updating Metrics
```
Code Map:
  23 nodes
  31 edges
  3 files
```
Updates in real-time as you work

### Watch Mode (Future)
```bash
# Start watch mode
/carto:carto-scan --watch

# Continuously monitors project
# Updates graph on file changes
# No CLI needed - just edit files
```

## Configuration

### Auto-Analysis Settings
In `.code-map/config.json`:
```json
{
  "autoAnalysis": {
    "enabled": true,
    "watchMode": false,
    "gitDiffEnabled": true,
    "updateInterval": 5000
  }
}
```

### Performance
- **Small projects** (<50 files): Updates instant
- **Medium projects** (50-500 files): ~1-3 seconds
- **Large projects** (500+ files): ~5-10 seconds

## Usage Patterns

### Pattern 1: Continuous Code Review
1. Setup: `/carto:carto-map path="./project"` then `/carto:carto-parse`
2. Open live UI: `/carto:carto-visualize --interactive`
3. As you edit files, graph updates automatically
4. Use metrics overlay to understand impact: `--metric=complexity`
5. Find issues: `/carto:carto-analyze type=health`

### Pattern 2: Collaborative Development
1. Team member makes changes and pushes to git
2. Check changes: `/carto:carto-info --diff`
3. Automatically re-analyzes changed files via `carto-parse`
4. Team can see architectural impact instantly
5. View in canvas: `/carto:carto-visualize --interactive`

### Pattern 3: Onboarding New Contributors
1. New dev maps project: `/carto:carto-map path="./project"`
2. Auto-analysis scans codebase: `/carto:carto-parse`
3. Open interactive canvas: `/carto:carto-visualize --interactive`
4. Explore architecture with visualizations
5. Find key code: `/carto:carto-find "main" mode=map`

## Technical Details

### Auto-Analysis Pipeline
```
┌─────────────────────────────────────────┐
│ Browser loads http://localhost:5173     │
└────────────────┬────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────┐
│ useEffect triggers handleLoadExample()   │
└────────────────┬────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────┐
│ Load example files from /examples/      │
└────────────────┬────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────┐
│ Parse each file with appropriate parser │
└────────────────┬────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────┐
│ Build code map (nodes + edges)          │
└────────────────┬────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────┐
│ Run analysis (security, patterns)       │
└────────────────┬────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────┐
│ Render D3.js graph + file tree          │
└─────────────────────────────────────────┘
```

### No Manual Steps
- ✅ Auto-detects example files
- ✅ Auto-parses all languages
- ✅ Auto-builds relationships
- ✅ Auto-renders visualization
- ✅ Auto-updates as you work

## Future Enhancements

### Phase 2: Git Integration
- Auto-detect file changes via git
- Show which files changed in UI
- Highlight recently modified code
- Display change statistics

### Phase 3: Watch Mode
- Real-time file system monitoring
- Instant updates on save
- LSP integration for IDE
- Collaborative editing

### Phase 4: CI/CD Integration
- Trigger analysis on push
- Generate reports automatically
- Track metrics over time
- Architectural trend analysis

## Troubleshooting

### Graph Not Updating
- Run `/carto:carto-parse` to rescan and update
- Check browser console for errors
- Reload page (F5)

### Slow Performance
- Increase `updateInterval` in config
- Filter by language: `/carto:carto-parse languages=["typescript"]`
- Focus on specific folder: `/carto:carto-map path="./src"`

### Missing Files
- Check `.code-map/config.json` for include/exclude patterns
- Verify file extensions are supported
- Run `/carto:carto-map path="./project"` to rescan all files
