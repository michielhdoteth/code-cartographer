---
name: carto
description: >
  Use carto to understand a codebase without reading the whole repository.
  Triggers on "carto", "code map", "where is this symbol", "what depends on this file",
  "blast radius", "dead code", "circular dependencies", "code health", "structural health",
  "deep modules", "shallow module", "refactor prompt", "context budget", "repo map",
  "codebase question", "find usages", "trace callers", "which files to edit".
  Always use carto for codebase navigation instead of guessing paths or reading whole trees.
---

# carto â€” Codebase Map and Agent Context CLI

carto is a local CLI built to be used by AI agents. The agent is the LLM: it reasons and
decides what to inspect; carto only maps the codebase and returns small, structured answers
so the agent's context window stays useful. carto has no LLM inside and never makes an LLM
or network call. Humans can use the same commands directly. No API keys, no daemons;
everything is deterministic.

**Binary:** `carto` (Go, single binary; build with `cd cmd && go build -o ../bin/carto .`)
**Cache:** `.carto/graph.json` plus `.carto/retrieval.json` (tokenized documents) under the target project. Run `carto map <path>` first, then `carto check <path>`.

## Quick reference

| Command | Use it when |
|---|---|
| `carto brief [path] --tokens 1200 --json` | L0: compact repo map + file inventory |
| `carto check [path] --json` | First call in any session: is the graph fresh? |
| `carto map <path>` | Build/refresh the graph (writes `.carto/graph.json`) |
| `carto doctor [path] --json` | Diagnose freshness, parser coverage, config problems |
| `carto repo-map [path] --tokens 2048 --focus src/auth --json` | Ranked signature map; `--expand` doubles budget when no focus is given |
| `carto ask "question" [path] --limit 8 --tokens 4000 --json` | Find symbols with BM25F + graph fusion; `--answer-only` returns facts without bodies |
| `carto search "query" [path] --json` | Search the local text index with BM25F (run `carto index` first) |
| `carto context <file> --outline --tokens 512 --json` | L1: AST skeleton (signatures + line ranges, no bodies) |
| `carto context <file> --symbols Start,Load --json` | L1: only the requested symbol ranges |
| `carto context <file> --lines 100-160 --json` | L1: only a raw line window |
| `carto chunk <file> --list --json` | L2: list AST-aligned chunks (~300 tokens each) without content |
| `carto chunk <file> --chunk 0 --json` | L2: fetch one chunk, with `next_chunk` navigation |
| `carto context <file> --tokens 8000 --imports --callers --cohesion --json` | L3: full file context with savings + cohesion |
| `carto pack [path] --include 'src/**' --compress --tokens 8000 --output ctx.md` | Pack repo context into a file |
| `carto impact <file> [path] --json` | Blast radius before editing a file |
| `carto trace <symbol> [path] --direction both --depth 3 --json` | Callers/callees of a symbol |
| `carto analyze structure [path] --json --fail-on critical` | Deep-module / narrow-interface linter |
| `carto analyze all [path] --json` | Dead code, complexity, cycles, structure, health |
| `carto fix <finding-id> [path] --tokens 4000` | Get an LLM-ready refactor prompt for a finding |
| `carto schema [command]` | Print the JSON contract for a command |

All commands accept the global `--json` flag. Prefer JSON when parsing output programmatically.

## Progressive disclosure (avoid reading huge files)

Use the four-level ladder and stop as soon as you have enough:

1. L0 `carto brief . --tokens 1200 --json` â€” repo map + per-file inventory.
2. L1 `carto context <file> --outline --tokens 512 --json` â€” symbol signatures and line ranges, zero bodies. Use `--symbols`/`--lines` to select exact ranges.
3. L2 `carto chunk <file> --list --json`, then `carto chunk <file> --chunk N --json` â€” fetch one ~300-token AST-aligned chunk at a time; follow `next_chunk`.
4. L3 `carto context <file> --tokens 8000 --json` â€” only when you truly need the full file.

Every view reports `savings.avoided_lines`, `savings.avoided_tokens`, and `savings.context_efficiency_score`. `suggested_next` tells you the recommended next command.

## Recommended workflow

1. `carto doctor . --json` â€” if status is `error`, run `carto map .` and repeat.
2. `carto repo-map . --tokens 2048 --json` â€” orient cheaply before reading files.
3. `carto ask "<question>" . --json` â€” locate symbols; follow with `carto trace` or `carto impact`.
4. `carto context <target> --imports --callers --cohesion --tokens 8000 --json` â€” read only what is needed.
5. `carto analyze structure . --json` â€” before a structural edit, check for shallow modules/cycles.
6. After a refactor, re-run the same analysis to verify the finding disappeared.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | Success (or findings below the configured threshold) |
| 1 | Findings reached `--fail-on` threshold |
| 2 | Usage error / bad arguments / schema mismatch |
| 3 | Stale graph â€” run `carto map <path>` |
| 4 | I/O or configuration error |

`carto check` returns 3 on a stale graph. Scripts should handle 3 by remapping, not by assuming failure.

## Ranking and tuning

`ask` and `search` use deterministic BM25F lexical ranking with identifier splitting and graph-aware fusion. Per-repository weights live in `.carto/config.json` under `retrieval`; changing them invalidates only the scoring cache. `carto retrieval tune` grid-searches weights against the offline suite and writes the config only when recall improves.

## Interpreting results

- `context.savings` and `repo_map.savings` are **estimates** (~4 UTF-8 bytes/token), not model tokenizer counts.
- `context.cohesion.score` (0â€“100) falls when a feature's related context spans multiple top-level directories. Below 70 is a refactor signal.
- `structure.findings[].id` is stable. Feed it to `carto fix <id>`; do not hand-write a refactor prompt.
- `structure.findings[].severity`: `critical` (cycles), `warning` (tangles/shallow/scattered), `info` (broad interface).
- Suppress findings only in `.carto/structure-allow` with a `reason` comment; never suppress silently.
- Import/call edges are inferred where parsers cannot resolve them; confidence is reported in `trace`/`ask` JSON.

## Gotchas

- `carto impact <file>` resolves relative to the current directory; run it from the project root.
- A stale graph produces confident-looking wrong answers. Always `carto check` first in a new session.
- `--json` is machine-readable: no progress lines are written to stdout in JSON mode.
- `carto fix` only prints a prompt. It never edits code; you remain responsible for the change and tests.

## Install this skill

Run `carto skill --install` to write this file to `.claude/skills/carto/SKILL.md` and to add a carto
section to `AGENTS.md`. Use `carto skill --check` in CI to detect drift.
