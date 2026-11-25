---
title: "Code Cartographer competitive landscape: codebase-context tools for AI agents"
status: draft
tags: [research]
---

# Code Cartographer vs. the codebase-context tool landscape

Date: 2026-11-24. Star counts are as shown on GitHub today and drift; treat as approximate. Sources: repo pages fetched this session. Decision this unblocks: which ideas to adopt in carto and what Carto's defensible niche is.

## Comparison table

| Tool | Lang / License | Stars (2026-11) | Approach | Token strategy | MCP? | Steal-able idea |
|---|---|---|---|---|---|---|
| repomix (yamadashy) | TS, MIT | ~29.4k | Dump-to-file (XML/MD/JSON/plain), optional tree-sitter `--compress` signatures-only mode, git-aware sort, secretlint scan | tiktoken (`o200k_base` default, `--token-count-encoding`); `--token-budget` CI guard | Yes (`--mcp`) | `--compress` signature-extraction mode; `--token-budget` exit-code guard; `--split-output`; git-change-frequency file ordering |
| code2prompt (mufeedvh) | Rust, MIT | ~7.8k | Same dump category, Handlebars templates, TUI, Python SDK | tiktoken-style per-file counts, tree-sitter `entity-map` experimental | Yes | Handlebars prompt templating; entity-map (compact symbol map) as an optional pack mode |
| gpt-repository-loader (mpoon) | Python, (no license file seen) | ~3k, stale (2024) | Naive dump | none notable | No | Nothing — superseded |
| gitingest (coderamp-labs) | Python, MIT | ~16.2k | Dump, web UX, "hub→ingest" URL trick | tiktoken | No | Web UX / instant preview is the adoption driver |
| aider (Aider-AI) | Python, Apache-2.0 | ~50.2k | tree-sitter repo-map, PageRank-ish graph ranking of definitions, `--map-tokens` (default 1k, expands when no files in chat) | tiktoken (via litellm) | No (CLI) | Repo map = signatures only, ranked by graph centrality, budget-aware, expand-on-uncertainty |
| Continue.dev | TS, Apache-2.0 | ~36.5k (repo now read-only/2.0 final) | Agent + optional embeddings (local or API) for codebase indexing | n/a | No (IDE) | Local embeddings option; `.continueignore` pattern |
| claude-context (zilliztech) | TS, MIT | ~12.7k | Embeddings (AST chunks) + Milvus/Zilliz hybrid BM25+vector; Merkle-tree incremental indexing | n/a (retrieval, not packing) | Yes | Merkle-tree incremental reindex; hybrid BM25+vector is the bar embeddings tools set |
| code-review-graph (tirth8205) | Python, MIT | ~32.9k | tree-sitter → SQLite graph; blast-radius/risk scoring; incremental (SHA-256, ~2.5s/2-file edit); MCP + hooks | ~4 bytes/token heuristic + tiktoken cross-check (`--verify`) | Yes | **Best-in-class reference**: blast radius + risk score + `context_savings` field on every MCP response + `--verify` against cl100k |
| codebase-memory-mcp (DeusData) | C, MIT | ~49.1k | tree-sitter (vendored grammars) → persistent SQLite graph, MCP, hybrid LSP, 46 agent integrations | graph-query answers instead of file reads | Yes | Local-first persistent graph + "answer from graph, not files"; architecture-decision records in graph |
| graphify (Graphify-Labs) | Python, Apache-2.0/MIT | ~153k (suspicious, unverified) | tree-sitter AST → Leiden-clustered graph, docs/PDF/media too; no embeddings; local | n/a | Yes | Leiden community detection → architecture summary; EXTRACTED/INFERRED edge-confidence tags |
| code-graph-rag (vitali87) | Python, MIT | ~5.4k | tree-sitter + Neo4j/memgraph RAG for monorepos | n/a | Yes | Graph-DB option; multi-hop retrieval |
| axon (harshkedia177) | Python | ~0.8k | tree-sitter graph → MCP + CLI | n/a | Yes | Similar architecture, smaller |
| SocratiCode (giancarloerra) | TS | ~3.4k | hybrid semantic search + graph, enterprise/local | n/a | Yes | "40m+ LOC" positioning |
| SCIP (scip-code) | Go/protobuf, Apache-2.0 | ~895 | Language-agnostic precise index format (protobuf) from language servers/compilers | n/a | No (index format) | Symbol-precise cross-repo references; richer than tree-sitter name matching |
| sourcegraph / Cody | — | repos 404 today | Was the reference for SCIP/precise code intel; Cody open-sourced historically | — | — | Sourcegraph repos returned 404 on 2026-11-24; treat as unavailable/changed, verify before relying |

## Measured token-reduction claims (with sources)

- **code-review-graph**: ~63× median per-question reduction vs whole-corpus baseline across 6 repos (range 34.6×–358×; fastapi 948,793→2,653 tokens). Crucially, vs a *grep-and-read-top-3-files* agent baseline: **~6.65× median** (4.2×–72.6×), 2026-09-22 capture. Impact "recall 1.0" is circular (graph-derived ground truth); they say so openly. Source: README + docs/REPRODUCING.md, fetched 2026-11-24.
- **codebase-memory-mcp** (arXiv:2603.27277, submitted 2026-03-28): 83% answer quality vs 92% for file-exploration agent, at **10× fewer tokens and 2.1× fewer tool calls**, across 31 repos; matches/beats explorer on graph-native queries in 19/31 languages.
- **claude-context** (Zilliz): claims "~40% token reduction at equivalent retrieval quality" (vendor eval, methodology in repo `/evaluation`; not independently replicated).
- **graphify**: claims LOCOMO recall@10 0.497, "0 LLM credits" for graph build; not comparable to packing metrics.
- No major tool publishes a **head-to-head vs aider-style repo-map** number; code-review-graph's 6.65× vs grep baseline is the most honest public number.

## Steal this (ranked)

1. **Adopt code-review-graph's `context_savings` pattern**: every carto JSON/MCP response carries `{estimated, saved_tokens, saved_percent}` and an optional `--verify` flag cross-checking against tiktoken cl100k_base. Turns carto's 4-bytes/token guess into a defensible number and gives users a built-in demo of value.
2. **Ship an MCP server mode for carto** (`carto mcp`) exposing `query`, `blast_radius`, `repo_map`, `context`. Every serious entrant (code-review-graph, codebase-memory-mcp, graphify, claude-context) is MCP-first; carto currently isn't.
3. **Incremental re-index by content hash** (they use SHA-256 per file; ~2.5 s for a 2-file edit on a 3k-file repo). carto has content-hash caching for graph.json — make sure `context`/`repo-map`/`blast-radius` all ride the same incremental path and document the latency.
4. **Blast radius with risk tiers + "affected tests"**: code-review-graph maps diffs → affected symbols → risk score + test gaps. carto has blast-radius BFS but no risk tiers or test-gap output; add both — it's the single most "agent-useful" output.
5. **Edge confidence tags** (Graphify's EXTRACTED/INFERRED): carto's import-resolution is suffix/heuristic; tag edges so agents know what's certain vs guessed. Cheap, high trust value.
6. **`--compress` signature-only pack mode tuning**: repomix's tree-sitter signature extraction with `⋮` elision is the proven pattern; carto's `pack --compress` likely matches — verify it emits doc-comments + signatures and elides bodies, and add a `--token-budget` non-zero-exit guard for CI.
7. **Repo-map "expand when uncertain" behavior** (aider): when no files are pinned in chat, let the map grow past budget; when files are pinned, shrink it. Mirrors aider's dynamic behavior and is trivial in carto's budget loop.
8. **Leiden/Louvain community detection** over the call graph (graphify, code-review-graph, codebase-memory all use it) to emit an architecture summary ("subsystems + hubs") — carto has hubs (fan-in≥10%) but no community detection; this is the missing piece for "architecture at a glance".
9. **Git-change-frequency ordering**: repomix sorts files by git churn by default; carto's `repo-map` ranks by graph centrality. Blend both (centrality × recency) — churn is a strong prior for what the agent is about to touch.
10. **Token counting honesty**: keep 4 bytes/token as the default (provider-neutral, matches code-review-graph), but add optional `--tokenizer tiktoken|o200k_base` (repomix/gitingest both ship tiktoken) so power users get exact counts. Document both.
11. **"Answer from graph, not files" MCP tool**: codebase-memory's arXiv result (83% quality at 10× fewer tokens) validates a `carto ask --answer`-style tool that returns synthesized graph facts (callers, paths, hubs) without file bodies; carto's `ask` already returns bounded snippets — add an `--answer-only` mode.
12. **Persistent SQLite graph instead of JSON cache** if graph.json is getting large: code-review-graph and codebase-memory both use SQLite (+FTS5) for incremental queries; carto's JSON cache won't scale past ~50k files or support FTS.

## Measured-claims honesty box

- code-review-graph: ~63× vs whole-corpus (inflated baseline), **6.65× vs realistic grep baseline** — cite the latter.
- codebase-memory-mcp: 10× fewer tokens, 2.1× fewer calls, 83% vs 92% quality — peer-reviewable preprint (arXiv 2603.27277, Mar 2026).
- claude-context: ~40% reduction, vendor-run eval, unreplicated.
- Nobody publishes head-to-head vs aider repo-map or vs a plain `grep` on the *same* harness; treat all single-vendor numbers as directional.

## Gaps none of them fill (carto's opening)

1. **Deterministic, model-free benchmark harness** — carto's `benchmark` (precision/recall, no LLM) is unique; nobody else ships retrieval eval in the box. Extend it to the code-review-graph scenarios (6-repo suite) and publish numbers.
2. **Provider-neutral 4-bytes/token default** — everyone else hard-wires tiktoken or hides the estimate; carto's explicit estimate is honest for non-OpenAI models.
3. **Single static Go binary, no Python/Node/Neo4j/vector DB** — every credible competitor requires Python+deps or a vector DB or Docker. carto's zero-dependency local-first story (one binary + React UI) is unmatched for CI/air-gapped use.
4. **Health/dead-code/circular-dependency analysis + blast radius in one tool** — competitors do graphs *or* analysis, not both; carto can be the only one answering "what breaks + is it healthy + what's dead" in one pass.
5. **No one does SCIP-level precision cheaply** — an optional scip-typescript/scip-go import path would be a differentiator none of the packers offer.

## Unverified / caveats

- graphify's 153k stars looks anomalous (repo created recently, ROSS Index claim); treat popularity as unverified. Several star counts may be inflated across this ecosystem.
- sourcegraph/sourcegraph and sourcegraph/cody returned 404 on 2026-11-24; SCIP itself is alive (scip-code/scip, Apache-2.0, ~895 stars). Sourcegraph's current open-source status is unverified.
- code2prompt's exact tokenizer (README only says "parallel per-file token counts") unverified.
- mpoon/gpt-repository-loader ~3k stars, unmaintained (last update 2024-06); effectively dead.
- "4 bytes/token" is a heuristic; code content averages ~3–4 bytes/token depending on language/indentation — carto should state its measured error band, not just the constant.
