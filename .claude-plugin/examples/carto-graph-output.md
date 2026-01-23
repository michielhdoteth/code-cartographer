# Code Cartographer - Graph Output Example

## Command
```
/carto:carto-graph style=tree
```

## ASCII Tree Visualization

```
Code Cartographer Map
├── 📄 api.ts (TypeScript, 59L)
│   ├── 🔹 Request (interface)
│   ├── 🔹 Response (interface)
│   └── ◆ APIHandler (class, C:5)
│       ├── → constructor(auth: AuthManager)
│       ├── → initializeRoutes(): void
│       ├── → handle(request: Request): Response
│       ├── → getUsers(request: Request): Response
│       ├── → createUser(request: Request): Response
│       └── → getProfile(request: Request): Response
│
├── 📄 auth.js (JavaScript, 43L)
│   └── ◆ AuthManager (class, C:4)
│       ├── → constructor(config)
│       ├── → login(email, password): Promise
│       ├── → logout(): void
│       ├── → isAuthenticated(): boolean
│       └── → getToken(): string
│
└── 📄 database.py (Python, 75L)
    └── ◆ Database (class, C:6)
        ├── → __init__(db_path: str)
        ├── → connect(): void
        ├── → create_tables(): void
        ├── → get_user(user_id: int): Dict
        ├── → create_user(email: str, password: str): int
        ├── → get_user_posts(user_id: int): List
        └── → close(): void
```

## Dependency Graph

```
DEPENDENCIES:
  api.ts
    ├──> imports: AuthManager (from ./auth)
    └──> uses: Request, Response (interfaces)

  auth.js
    ├──> depends_on: fetch API
    └──> exports: AuthManager

  database.py
    ├──> imports: sqlite3
    ├──> defines: Database, users table, posts table
    └──> calls: sqlite3.connect(), cursor.execute()

RELATIONSHIPS:
  APIHandler ──calls──> AuthManager.isAuthenticated()
  APIHandler ──calls──> AuthManager.getToken()
  Database ──implements──> CRUD operations for users & posts
```

## Statistics

```
Overall Metrics:
  Total Nodes:      23 (3 classes + 2 interfaces + 18 methods)
  Total Edges:      31 (imports, calls, dependencies)
  Cyclomatic Complexity: Avg 4.2
  Code Health:      B+ (Good)

Language Distribution:
  TypeScript: 33.3% (2 files)
  JavaScript: 33.3% (1 file)
  Python:     33.3% (1 file)

Complexity Breakdown:
  Low (1-3):    8 methods   (47%)
  Medium (4-6): 8 methods   (44%)
  High (7+):    2 methods   (9%)
```

## Canvas Visualization

Run `/carto:carto-canvas` to open an interactive force-directed graph where you can:
- **Zoom** - Scroll to zoom in/out
- **Pan** - Click and drag to move around
- **Drag Nodes** - Click and drag individual nodes to rearrange
- **Color By** - Type | Language | Complexity
- **Size By** - Fixed | Lines of Code | Complexity
- **Toggle Labels** - Show/hide node names

## Available Styles

```
ASCII Styles:
  /carto:carto-graph style=tree           (hierarchical tree)
  /carto:carto-graph style=relationships  (dependency graph)
  /carto:carto-graph style=calls          (call graph)
  /carto:carto-graph style=overview       (class overview)
  /carto:carto-graph style=dependencies   (import graph)
  /carto:carto-graph style=hierarchy      (class hierarchy)

Canvas Styles:
  /carto:carto-graph style=call_graph     (D3.js call graph)
  /carto:carto-graph style=dependency_graph (D3.js dependency graph)
  /carto:carto-graph style=class_hierarchy  (D3.js class hierarchy)
```

## Next Steps

1. **Analyze Quality**
   ```
   /carto:carto-detect
   ```
   Detects design patterns, anti-patterns, and code smells

2. **Query the Map**
   ```
   /carto:carto-query "find all classes with complexity > 5"
   /carto:carto-query "show dependencies for APIHandler"
   ```

3. **Search Code**
   ```
   /carto:carto-search "login"
   /carto:carto-search type:class language:typescript
   ```
