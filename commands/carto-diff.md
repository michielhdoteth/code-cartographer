---
description: Show git-style diff of code map changes
---

# carto-diff

Show git-style diff of the code map.

## Usage

```
/carto:carto-diff
```

## Output

```
diff --git a/.code-map/code-map.toon b/.code-map/code-map.toon
index 28..418 100644
--- a/.code-map/code-map.toon
+++ b/.code-map/code-map.toon
@@ -1 +1,28 @@
+Files: 45
+Classes: 23
+Functions: 156
+Edges: 342

+By Type:
+  class: 23
+  function: 156
+  module: 45
```

## Use Case

Review changes before committing to git:

```bash
git add .code-map/
git commit -m "Update code map"
git diff .code-map/
```

## Example

```
/carto:carto-diff

Output:
diff --git a/.code-map/code-map.toon b/.code-map/code-map.toon
index 28..418 100644
--- a/.code-map/code-map.toon
+++ b/.code-map/code-map.toon
@@ -1 +1,45 @@
+Files: 45
+Classes: 23
+Functions: 156
+Edges: 342
```
