## Simplification report

### Changed
Strict comment policy applied to Go files under `cmd/` only. Removed section banners, step-by-step restatements, and exported/type doc comments that only restated names or obvious behavior. Kept WHY comments: algorithm rationale (`Σ 0.6^(depth-1)`, IDF floor, PageRank, RRF k=60, ~4 bytes/token), cycle-vs-tangle policy, parser-vocabulary uncertainty, best-effort cache writes for read-only repos, allowlist format, opt-in/backward-compatibility constraints, enum/format constraints, build tags, and test comments describing invariants or fixture intent.

| File | Comments removed |
|---|---|
| `cmd/main.go` | 21 |
| `cmd/commands_map.go` | 2 |
| `cmd/analyze/structure.go` | 7 |
| `cmd/analyze/blastradius.go` | 20 |
| `cmd/analyze/deadcode.go` | 22 |
| `cmd/analyze/complexity.go` | 18 |
| `cmd/analyze/circular.go` | 23 |
| `cmd/analyze/graph.go` | 23 |
| `cmd/analyze/health.go` | 27 |
| `cmd/codegraph/agent_bench.go` | 3 |
| `cmd/codegraph/agent_context.go` | 6 |
| `cmd/codegraph/benchmark.go` | 1 |
| `cmd/codegraph/brief.go` | 4 |
| `cmd/codegraph/chunk.go` | 5 |
| `cmd/codegraph/query.go` | 1 |
| `cmd/codegraph/retrieval.go` | 4 |
| `cmd/codegraph/savings.go` | 4 |
| `cmd/codegraph/skeleton.go` | 3 |
| `cmd/config/config.go` | 5 |
| `cmd/context/context.go` | 20 |
| `cmd/context/cohesion.go` | 1 |
| `cmd/context/consult.go` | 2 |
| `cmd/doctor/doctor.go` | 4 |
| `cmd/find/search.go` | 15 |
| `cmd/git/git.go` | 6 |
| `cmd/git/git_test.go` | 6 |
| `cmd/manifest/manifest.go` | 3 |
| `cmd/manifest/manifest_test.go` | 3 |
| `cmd/retrieval/bm25f.go` | 7 |
| `cmd/retrieval/cache.go` | 6 |
| `cmd/retrieval/fuse.go` | 1 |
| `cmd/retrieval/settings.go` | 4 |
| `cmd/retrieval/tokenize.go` | 2 |
| `cmd/schema/schema.go` | 4 |
| `cmd/search/index.go` | 15 |
| `cmd/search/retrieval.go` | 1 |
| `cmd/serve/http.go` | 21 |
| `cmd/skill/skill.go` | 4 |
| `cmd/lib/parser/ast.go` | 23 |
| `cmd/lib/parser/languages.go` | 9 |
| `cmd/lib/parser/parser.go` | 24 |
| `cmd/lib/parser/parser_test.go` | 4 |
| `cmd/lib/parser/parser_compiled_test.go` | 15 |
| `cmd/lib/parser/parser_data_test.go` | 2 |
| `cmd/lib/parser/parser_scripting_test.go` | 8 |
| `cmd/lib/parser/treesitter.go` | 20 |
| `cmd/lib/parser/treesitter_stub.go` | 2 |
| `cmd/lib/parser/unified.go` | 10 |
| `cmd/lib/parser/unified_clike.go` | 13 |
| `cmd/lib/parser/unified_extract.go` | 11 |
| `cmd/lib/parser/unified_markup_data.go` | 15 |
| `cmd/lib/parser/unified_mobile_jvm.go` | 3 |
| `cmd/lib/parser/unified_scripting.go` | 7 |

### Measured
- Comments removed: 490 across 53 files
- Verification: compared every modified file against a byte-for-byte pre-edit snapshot; all 475 changed hunks were comment deletions or trailing-comment strips. No code, identifier, blank line, or file-ending changes.
- `go build ./...` (workdir `cmd`) → exit 0
- `go test ./...` (workdir `cmd`) → exit 0, all packages ok

### Left alone
- WHY/constraint comments listed above, including misplaced-but-load-bearing multi-line blocks where a WHY line shared the block with a restatement.
- Build tags (`//go:build`, `// +build`, `//go:embed`) and test comments stating invariants (e.g. `// Re-install must not duplicate the AGENTS.md section.`).
- Files with no removable comments: `cmd/commands_agent.go`, `cmd/commands_context.go`, `cmd/commands_diagnostics.go`, `cmd/helpers.go`, `cmd/codegraph/index.go`, `cmd/exitcode/exitcode.go`, `cmd/fix/fix.go`, `cmd/skill/skill_test.go`, `cmd/tuner/tuner.go`, and all test files without comments.
- Pre-existing working-tree changes outside the comment deletions were not touched.
