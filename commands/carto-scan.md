---
description: Scan directory for source files
parameters:
  path:
    type: string
    description: Path to scan
    required: false
---

# carto-scan

Scan a directory for source files and add them to the code map.

## Usage

```
/carto:carto-scan
/carto:carto-scan path="/path/to/project"
```

## What Happens

1. Scans for `.py`, `.js`, `.ts`, `.java`, `.go`, `.rs` files
2. Creates `CodeFile` entries for each file
3. Stores file metadata (size, line count, checksum)
4. Updates `code-map.toon`

## Supported Languages

| Extension | Language |
|-----------|----------|
| .py | Python |
| .js | JavaScript |
| .ts | TypeScript |
| .tsx | TSX |
| .java | Java |
| .go | Go |
| .rs | Rust |

## Example

```
/carto:carto-scan

Output:
Scanned 45 files
Total files: 45
```
