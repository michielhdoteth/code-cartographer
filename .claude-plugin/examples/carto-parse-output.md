# Code Cartographer - Parse Output Example

## Command
```
/carto:carto-parse
```

## Parse Results

**Parse Time:** 1,247ms
**Files Parsed:** 3
**Nodes Extracted:** 12

### Code Structure

#### api.ts (TypeScript)
```
📄 api.ts (59 lines)
├── 🔹 Request (interface)
├── 🔹 Response (interface)
└── ◆ APIHandler (class)
    ├── → constructor() [9-23]
    ├── → initializeRoutes() [25-29]
    ├── → handle() [31-44]
    ├── → getUsers() [46-48]
    ├── → createUser() [50-52]
    └── → getProfile() [54-56]
```

#### auth.js (JavaScript)
```
📄 auth.js (43 lines)
└── ◆ AuthManager (class)
    ├── → constructor() [2-6]
    ├── → login() [8-26]
    ├── → logout() [28-31]
    ├── → isAuthenticated() [33-35]
    └── → getToken() [37-39]
```

#### database.py (Python)
```
📄 database.py (75 lines)
└── ◆ Database (class)
    ├── → __init__() [9-12]
    ├── → connect() [14-17]
    ├── → create_tables() [19-44]
    ├── → get_user() [46-50]
    ├── → create_user() [52-60]
    ├── → get_user_posts() [62-69]
    └── → close() [71-74]
```

### Extracted Metrics

| Entity | Type | Lines | Complexity | Language |
|--------|------|-------|-----------|----------|
| APIHandler | class | 42 | 5 | TypeScript |
| AuthManager | class | 40 | 4 | JavaScript |
| Database | class | 65 | 6 | Python |
| handle | method | 13 | 3 | TypeScript |
| login | method | 18 | 4 | JavaScript |
| create_tables | method | 25 | 2 | Python |

### Dependencies Found

- `api.ts` imports `AuthManager` from `./auth`
- `AuthManager` depends on `fetch` API
- `Database` depends on `sqlite3` module

### Summary
- **Classes:** 3
- **Methods/Functions:** 18
- **Interfaces:** 2
- **Total Metrics:** Average Complexity: 3.8

### Next Step
Run `/carto:carto-graph style=tree` to visualize the code structure.
