<!-- carto-guide:begin -->
## carto â€” codebase map

Use `carto` for codebase navigation instead of reading whole trees. Run `carto skill` for the full guide.

1. `carto doctor . --json` â€” graph freshness and health.
2. `carto repo-map . --tokens 2048 --json` â€” orient cheaply.
3. `carto ask "<question>" . --json` â€” locate symbols.
4. `carto context <file> --imports --callers --cohesion --json` â€” assemble only needed context.
5. `carto analyze structure . --json` and `carto fix <finding-id>` â€” before and after structural edits.

Exit codes: 0 ok, 1 findings threshold, 2 usage, 3 stale graph (run `carto map`), 4 I/O/config.
<!-- carto-guide:end -->
