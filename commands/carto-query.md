---
description: Query the code map
parameters:
  query:
    type: string
    description: Query string
    required: true
  type:
    type: string
    description: Query type
    required: false
    enum: [node, edge, file, all]
    default: all
---

# carto-query

Query the code map by name, type, or relationship.

## Usage

```
/carto:carto-query query="UserController"
/carto:carto-query query="AuthService" type=node
/carto:carto-query query="calls" type=edge
/carto:carto-query query="*.py" type=file
```

## Query Types

| Type | Searches |
|------|----------|
| `node` | Classes, functions, methods by name |
| `edge` | Relationships (calls, imports, inherits) |
| `file` | Source files by path |
| `all` | Everything |

## Examples

### Find a class
```
/carto:carto-query query="UserController" type=node

Output:
[CLS] UserController
  Methods: 12
  File: src/users/controller.py
```

### Find all functions
```
/carto:carto-query query="def " type=node

Output:
- process_data (engine/processor.py:45)
- validate_input (engine/validator.py:23)
- ...
```

### Find import dependencies
```
/carto:carto-query query="imports" type=edge

Output:
UserController imports AuthService
ProcessEngine imports Database
```

### Find file
```
/carto:carto-query query="src/users" type=file

Output:
- src/users/controller.py
- src/users/service.py
- src/users/model.py
```
