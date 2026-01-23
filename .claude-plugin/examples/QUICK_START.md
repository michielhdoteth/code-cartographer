# Code Cartographer - Quick Start

## 3-Step Workflow

### 1. Scan
```bash
/carto:carto-scan path="./examples/sample-project"
```
**Output:** 3 files found (177 lines total)

### 2. Parse
```bash
/carto:carto-parse
```
**Output:** 23 nodes, 31 relationships extracted

### 3. Visualize
```bash
/carto:carto-graph style=tree
```
**Output:** Tree visualization of all code

```
Code Cartographer Map
├── 📄 api.ts (TypeScript, 59L)
│   └── ◆ APIHandler (C:5)
│       ├── → handle()
│       ├── → getUsers()
│       └── → createUser()
├── 📄 auth.js (JavaScript, 43L)
│   └── ◆ AuthManager (C:4)
│       ├── → login()
│       └── → isAuthenticated()
└── 📄 database.py (Python, 75L)
    └── ◆ Database (C:6)
        ├── → connect()
        └── → get_user()
```

## See More Examples

- **carto-scan-output.md** - Scan results
- **carto-parse-output.md** - AST extraction
- **carto-graph-output.md** - Graph visualizations
- **WORKFLOW.md** - Full workflow guide

## Web UI

Open interactive visualization at: **http://localhost:5173**

1. Click "Scan Current Project"
2. Explore the D3.js graph in the center
3. Navigate code structure on the right
4. Use controls to change colors/sizes

## Commands Reference

| Command | Purpose |
|---------|---------|
| `/carto:carto-scan` | Find source files |
| `/carto:carto-parse` | Extract code structure |
| `/carto:carto-graph` | Generate visualizations |
| `/carto:carto-canvas` | Open interactive web UI |
| `/carto:carto-detect` | Find patterns & issues |
| `/carto:carto-query` | Search the code map |
| `/carto:carto-search` | Full-text search |
| `/carto:carto-status` | Show map statistics |
| `/carto:carto-diff` | Compare maps |

## Language Support

- Python 🐍
- JavaScript/TypeScript 📘
- Java ☕
- Go 🐹
- Rust 🦀
- C++ 🔨
- Ruby 💎
- PHP 🐘
