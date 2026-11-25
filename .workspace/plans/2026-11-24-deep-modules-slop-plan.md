# Structural Analysis for Code Cartographer — deep modules + agent context efficiency

**Goal:** Evolve Code Cartographer from a passive mapper into an architectural-quality linter by adapting the *ideas* of deslop-please (deep modules, narrow interfaces) and `untangle-codebase` (agent-side refactor loop), into carto's existing Go/CLI + graph architecture. Add measurable token-efficiency features and an optional learned re-ranker.

**Architecture (target):** Implement the structural analyzer natively in Go first (`cmd/analyze/structure.go`) so it ships with the existing CLI. Keep the TypeScript web UI unchanged for v1 (CLI/JSON only). A Rust `crates/carto-core` rewrite of the analysis core is a *later, gated* strangler-fig migration, not a prerequisite. The versioned `.carto/graph.json` contract is the boundary (additive `schema_version: 2` with member-access facts).

**Tech:** Go 1.25 (existing CLI + first implementation), Rust (future core rewrite, tree-sitter/petgraph), TypeScript/React (UI, unchanged in v1), Node/Vitest and `go test` for gates.

**Research base (2026-11-24):** `.workspace/research/2026-11-24-codebase-context-tools.md` (competitive landscape) and `.workspace/research/2026-11-24-small-code-models-for-context-selection.md` (neural options). Key conclusions: carto's defensible niche is deterministic single-binary retrieval + built-in eval; the biggest competitive gap is an MCP server; neural parsing is a non-starter, ~137M embedding re-ranking is an optional benchmarked spike only.

---

## 0. Corrections to the brief (read first)

The brief contained assumptions that do not match this repository. The plan is built on what was verified, not on the brief:

1. **This is a Go project, not Rust.** `cmd/` is Go 1.25 + Cobra + smacker/go-tree-sitter; the UI is React/TS. **Decision:** ship the structural analyzer in Go first; a Rust `crates/carto-core` rewrite is a later, gated strangler-fig migration (`--engine=rust`), never a flag-day rewrite. The Go CLI remains the reference implementation and shipping artifact until parity is proven.
2. **"Architecture → Contracts → Defects → Spec" pipeline does not exist here.** The real pipeline is: `map` (walk + detect languages) → `.carto/graph.json` (versioned cache, content-hash incremental) → `analyze` (dead-code, complexity, circular, health, blast-radius) → `context` / `repo-map` / `pack` (token-budgeted agent context) → `serve` (Web UI). The slop rules plug into `analyze` and feed `context`; there is no spec layer to wire into.
3. **"Rust is faster than Go" is not a safe assumption.** For I/O-bound scanning, parity is likely. Rust's real wins here: tree-sitter's first-class Rust bindings, no GC pauses on large graphs, fearless parallelism (rayon), and ecosystem alignment with deslop-please. The migration must be justified by **parity + new capabilities**, not by speed promises.
4. **deslop-please is someone else's MIT-licensed project.** This plan treats it as design inspiration. We re-implement ideas (with credit in NOTICE/README), we do not port code, command names, or copy. That also matches your instruction: grab what's liked, make Code Cartographer better.
5. **"Tokens/task dropped after deslopping" is an anecdote, not a law.** Treat it as a hypothesis to *measure* via `carto benchmark`, not a promised outcome.
6. **Neural networks:** research verdict — no neural parsers, no default-on model, no >500M model, no OpenRAIL-licensed weights. Optional spikes only: a tiny scorer over hand-crafted features, then (if justified) a ~137M MIT CodeRankEmbed int8 ONNX re-ranker on a shortlist. Both killed if they don't beat the deterministic baseline on carto's own benchmark.
7. **Working tree is dirty** (`README.md`, `cmd/context/context.go`, `cmd/lib/parser/*`, `cmd/main.go`, `cmd/serve/http.go`, `ui/*` are modified). Work on top of it; don't discard user changes.

---

## 1. Current state (verified)

- `cmd/main.go` — Cobra root; `map`, `analyze [dead-code|complexity|circular|health|all]`, `context`, `repo-map`, `pack`, `ask`, `trace`, `impact`, `check`, `benchmark`, `serve`, etc. Global `--json`.
- `cmd/lib/parser/` — tree-sitter via `smacker/go-tree-sitter`. Extracts nodes (functions, classes, …), exports, imports, calls. **Does not emit member-access events** (`x.field`), which `field-leak` needs.
- `cmd/analyze/graph.go` — `ModuleGraph`: `Files`, `Imports`, `Exports`; file-granularity edges; import resolution is suffix/heuristic matching. `health.go` bundles dead-code + complexity + circular into a 0–100 score with severity issues. `blastradius.go` has reverse-graph BFS, `detectHubs` (fan-in ≥ 10% of repo) — the natural neighbor of a "shallow module" detector.
- `cmd/codegraph/` — `Index` (`.carto/graph.json`, `schema_version: 1`), `Node`/`Edge`, PageRank-ish `centrality()`, `GenerateRepoMap`, `GeneratePack`, `RunBenchmark`.
- `cmd/context/context.go` — token budgeting (~4 bytes/token), `--imports` / `--callers` expansion.
- Tests: `go test ./...` covers parser, agent_context, git, manifest. No CI (`.github/` contains only `banner.png`). Test fixtures: `test-data/squish`.
- UI: `ui/components/HealthScoreDisplay.tsx` renders analyze health; expects a `{ score, issues[] }`-shaped payload (confirm exact prop shape before wiring new fields).

### deslop-please concepts to adapt (MIT, credit required)
- **module-cycle** — strongly-connected components over module graph; reframe as "wrong layering" signal.
- **tangle** — two modules that know each other (mutual symbol references) without a formal import cycle; suggests a missing shared abstraction or misplaced responsibility.
- **field-leak** — struct fields known/used outside the defining module; the canonical "narrow your interface" smell.
- **shallow-module** — callers use a large fraction of a module's surface instead of few entry points; low depth-to-surface ratio.
- **Allowlist** — `.deslop-please.allow`-style suppression with reasons; carto-native equivalent: `.carto-slop-allow` with mandatory `reason =` comments.

---

## 2. Target design

```
 source ──> Go walker + parser (unchanged, extended)
                 │  writes .carto/graph.json v2
                 │  (adds: module ids, member_accesses[], module_edges[])
                 ▼
        carto-core (Rust, crates/carto-core)
          ├─ graph ingest + validation (schema_version 2)
          ├─ module resolver (Go: dir/package; TS/JS: nearest package.json|tsconfig; Py: __init__.py pkg; others: directory)
          ├─ rules: module_cycle, tangle, field_leak, shallow_module, broad_interface
          ├─ Structural Cohesion Score (per module & per "feature cluster")
          └─ findings JSON (stable schema, severity, evidence, fix hint)
                 │
                 ▼
        Go CLI ──> `carto analyze slop` (text + --json) ──> exit code policy
                ├─> `carto health` folds slop into score (opt-in flag at first)
                ├─> `carto context --cohesion` annotates token waste
                └─> `carto fix --prompt <finding-id>` emits an LLM-ready refactor prompt
```

Hard rules:
- `graph.json` v2 is **additive**; v1 readers keep working (unknown keys ignored).
- Rust core is **stateless per invocation** (file in → JSON out). No daemon, no network.
- Every rule must be **deterministic** (sorted iteration, stable tie-breaks) so JSON output is byte-diffable for tests.
- All thresholds configurable in `.carto/slop.toml`; defaults ship calibrated on ≥3 real repos (this repo + 2 others).

---

## 3. Rule specifications (adapted, not copied)

Let `M` = module, `E(M)` = exported symbols of M, `I(M)` = external modules importing from M, `uses(c, M)` = exported symbols of M referenced by module c, `impl_size(M)` = non-export LOC + internal edges.

| Rule | ID | Detection (v1) | Severity |
|---|---|---|---|
| module-cycle | `slop/module-cycle` | SCC of size > 1 over module import graph (Tarjan). Distinct from existing file-level `circular.go` (which stays; it's file-granular). | critical |
| tangle | `slop/tangle` | Bidirectional symbol references between two modules with no import cycle and no shared parent module; report the minimal mutual-reference set. | warning |
| field-leak | `slop/field-leak` | Member-access events `x.field` where `field` is declared on a type in M, access site ∉ M, count ≥ N (default 3) across ≥2 external modules. Go: exported field accessed in another package; TS: public property accessed cross-module. | warning→critical by count |
| shallow-module | `slop/shallow-module` | Depth-surface ratio: `depth = log(1 + internal_edges) / (1 + |E(M)|)`; flag when callers collectively touch `> 50%` of E(M) AND `impl_size < k * |E(M)|` (k≈25 LOC/export, calibrated). Report the 2–4 "front door" symbols actually used most. | warning |
| broad-interface | `slop/broad-interface` | `|E(M)|` exceeds P95 for repo **and** > 60% of exports used by ≤1 external module (speculative generality). | info/warning |

**Structural Cohesion Score (for `carto context`):** for the target file/feature, `cohesion = 1 − (spread / max_spread)` where spread = module count spanned by the symbol's forward+reverse call cone, normalized by repo module count. Surfaced as `structural_cohesion: 0–100` plus a one-line verdict ("scattered across 6 modules") in context output and JSON. This is the token-optimization lever: it tells the agent *why* the context is expensive, and flags refactors that would collapse it.

**Why not copy deslop-please verbatim:** its visibility model is Rust `pub` semantics; mapping onto Go/TS/Python needs per-language "export visibility" rules, and its "4 call sites in 2 modules" heuristic is a starting point, not a spec. Thresholds get calibrated here, documented, and versioned in `.carto/slop.toml`.

---

## 4. Rust core sketch (illustrative, not final API)

```rust
// crates/carto-core/src/slop.rs (sketch)
pub struct Module { pub id: ModuleId, pub exports: Vec<SymbolId>, pub loc: u32, pub internal_edges: u32 }

pub fn shallow_module(m: &Module, usage: &UsageMatrix, cfg: &SlopConfig) -> Option<Finding> {
    let used: usize = usage.distinct_used_exports(m).count();          // how many exports outsiders touch
    let ratio = used as f64 / m.exports.len().max(1) as f64;           // surface utilization
    let density = m.internal_edges as f64 / m.loc.max(1) as f64;       // implementation mass
    if ratio >= cfg.shallow_use_ratio && density < cfg.shallow_density_max {
        Some(Finding {
            rule: "slop/shallow-module",
            severity: Severity::Warning,
            evidence: Evidence::shallow(m.id, used, m.exports.len(), density),
            hint: format!("{} of {} exports are used; collapse the rest behind {} front doors",
                          used, m.exports.len(), cfg.front_door_target),
        })
    } else { None }
}
// module-cycle: petgraph::algo::tarjan_scc over module_edges from graph.json
// tangle: cross-module reference counts minus SCC members minus ancestor modules
// field-leak: group member_access events by (declaring_module, field); count distinct accessing modules
```

Rust value here: `petgraph` + `rayon` over a content-hash-cached graph, deterministic output, single-binary distribution later. No ML frameworks in the core crate.

---

## 5. CLI surface & output mockup

New commands (all support `--json`):

- `carto analyze slop [path] [--fail-on=critical|warning|never] [--top N]`
- `carto analyze slop --explain <finding-id>` — why it fired + the exact call sites
- `carto fix --prompt <finding-id> [--tokens 4000]` — prints an LLM-ready refactor prompt built from the finding + `context` of the involved symbols (this is deliverable #4)
- `carto context <file> --cohesion` — adds `structural_cohesion` block
- `carto benchmark` gains a `slop-suite.json` mode (before/after token cost of a refactor)

`carto analyze slop` text output (IDs are stable hashes of file+rule+anchor, so prompts survive re-runs):

```
Slop Report — D:\src\shop (847 files, 3 modules mapped)        1.9s

CRITICAL
  ✗ slop/module-cycle  2 cycles
      src/billing ↔ src/invoicing (via ApplyCredit, InvoiceTotal)
      src/auth    ↔ src/session   (via ParseToken, RefreshSession)
      why: mutual knowledge, no shared layer. Move the shared contract to src/billingreceipt or src/contracts.

WARNING
  ▲ slop/shallow-module  src/utils (score 0.18)
      callers touch 11 of 14 exports; only 2 are real entry points (NewPool, Close)
      38 caller sites across 21 files — context cost ≈ 3,900 tokens per task touching utils
      hint: split into `pool` and `strings` (2 deep modules) or collapse behind NewPool/Close
  ▲ slop/field-leak      src/store Queue.items read outside store by 4 modules (17 sites)
  ▲ slop/tangle          src/scheduler ⇄ src/cron (7 mutual references, no cycle)

HEALTHY
  ● deep module  src/raster  depth 4.2  (2 exports, 61 internal edges) — model citizen
  ● deep module  src/cache              depth 3.1

Summary: 2 critical · 3 warnings · 41 healthy modules
Run `carto fix --prompt slop/shallow-module:9f3a…` for a ready-to-paste refactor prompt.
Exit code: 1 (critical present)  [--fail-on=critical]
```

JSON shape (stable, versioned): `{ "version": 1, "root": …, "summary": {…}, "findings": [ { id, rule, severity, module, evidence{…}, span, fix_hint } ] }`.

---

## 6. Phased implementation

Conventions: each phase lands green (`go test ./...`, `cargo test`, `bun run test` if UI touched) with no behavior change to existing commands unless the phase says so. Work happens on a `feat/slop-lint` branch; the current dirty tree must be committed/stashed first.

**Phase 0 — Baseline & contracts (0.5–1 day)**
1. Commit/stash current work; create branch. Record golden outputs: `carto analyze all`, `carto context` on 3 fixtures, `carto repo-map`.
2. Extend parser to emit member-access events: `member_accesses: [{ receiver_type?, member, file, line, module }]` — Go via `selector_expression`, TS/TSX via `member_expression`, Python via `attribute`. Additive to `graph.json` → `schema_version: 2` (readers tolerate both).
3. Define `.carto/slop.toml` schema + JSON Schema for findings v1. Publish in `docs/slop.md`.
   - Acceptance: `graph.json` v2 round-trips; v1 consumers unaffected; tests green.

**Phase 1 — Rust workspace + module resolution + module-cycle (2–4 days)**
4. `crates/carto-core` (+`Cargo.toml` workspace at repo root), `carto-core analyze --graph .carto/graph.json --rules cycle` CLI; Go shell-outs only when `.carto/slop.toml` or explicit `--engine` flag present (pure-Go fallback stays default for one release).
5. Module resolver (Go package / TS package.json-or-tsconfig root / Python package / fallback dir), `module_edges` builder, Tarjan SCC; findings JSON; unit tests on synthetic fixtures.
   - Acceptance: finds a known cycle in a fixture; zero findings on `cmd/` itself (or document true positives); deterministic bytes on repeat runs.

**Phase 2 — shallow-module + broad-interface (2–3 days)**
6. Usage matrix from edges + member accesses; implement scoring in §4; calibrate thresholds against this repo + 2 real OSS repos (record calibration data in `docs/slop-calibration.md`).
   - Acceptance: each rule has ≥3 synthetic fixtures (true positive, true negative, boundary) + integration snapshot on this repo.

**Phase 3 — field-leak + tangle (2–3 days)**
7. field-leak from member_accesses (needs Phase 0); tangle from cross-module reference counts minus SCC/parent exclusions.
   - Acceptance: fixtures for Go (`internal/` respect), TS (barrel files), Python (leading-underscore treated as private); no cross-module false positives on a TS barrel re-export pattern (test with `ui/`).

**Phase 4 — Slop Report + fix prompts + context integration (2–4 days)**
8. `carto analyze slop` text/JSON renderers (extend `cmd/analyze/` with `slop.go`; reuse `FormatBlastRadiusResult` style). `--fail-on` for CI.
9. `carto fix --prompt <id>`: loads finding + pulls symbol context via existing `codegraph` + `context.Generate`, emits a constrained prompt (file list, invariants, "do not move exports, deepen the module", diff budget). No LLM calls in-process.
10. `carto context --cohesion` adds `structural_cohesion` to human + JSON output; `carto health` gains an *opt-in* `--include-slop` panel; UI: add a Slop/Health tab reading the new JSON (behind a flag; follow `HealthScoreDisplay.tsx` patterns; verify props before editing).
    - Acceptance: `carto analyze slop --json` validates against schema; prompt generated for every finding kind; UI builds untouched or with flag off.

**Phase 5 — Optional learned layer (only if data exists) (3–5 days, gated)**
11. Build labeled set: hand-label ≥150 findings from fixtures + 2 OSS repos (useful/not). Features: export ratio, density, fan-in/out, member-access counts, SCC size, LOC, language one-hot. Model: tiny MLP or gradient-boosted stumps, trained offline in Python (not shipped), exported as plain JSON weights (≤ 100 KB). Inference in Rust: 50-line forward pass, `--ml` flag, off by default.
    - Gate: if labeling yields <150 usable examples or AUC < 0.75 on held-out, **do not ship the model**; ship heuristics only and record why.
    - Never network. Model file versioned; fallback to heuristics on any load error.

**Phase 6 — Full Rust migration of the analysis core (optional, phased, per-command)**
12. One command at a time behind `--engine=rust`: `circular` → `blast-radius` → `dead-code` → `health`. Each port requires byte-identical or documented-delta golden outputs vs Go on `test-data/squish` + this repo. Go implementations deleted only after two green releases.
    - Note: parsers stay in Go until tree-sitter Rust grammar coverage is verified per language; this phase is explicitly *not* "rewrite everything", it's strangler-fig.

**Phase 7 — CI, docs, attribution (1 day)**
13. `.github/workflows/ci.yml`: go test, cargo test, `carto analyze slop --fail-on critical --json` on this repo (start `continue-on-error: true`, flip after one calibration cycle). `.carto-slop-allow` with `# reason:` per entry, enforced by the reporter (unknown keys → error). README section + `NOTICE` crediting deslop-please (MIT) as inspiration. `AGENTS.md` snippet: run slop analysis before structural edits; never `--allow` without a reason.

**Ordering guarantee:** Phases 0–4 deliver user value without any Rust-graph rewrite risk if 1 slips (Go shell-out boundary isolates it). Phase 5 is droppable. Phase 6 is independent.

---

## 7. Evaluation plan (proves the "fewer tokens" claim)

- Extend `carto benchmark` with a `slop-refactor` suite: 3 fixture repos, each with a known-bad module; measure tokens for a fixed edit task before/after an applied deep-module refactor (paired). Report median Δ tokens in `docs/slop-calibration.md`. This replaces the anecdote with a number before it goes in the README.
- Track slop-score trend in this repo itself each release.

## 8. Risks & mitigations

| Risk | Mitigation |
|---|---|
| Rust rewrite stalls the Go CLI | Phases 1–4 keep Go default; Rust is opt-in via `--engine=rust` until parity proven |
| False positives on generated/barrel files | Exclude `*.pb.go`, `*_generated.*`, `index.ts` re-export barrels by default; `.carto-slop-allow` with required reasons |
| Visibility inference wrong per language | Conservative: unknown → treat as exported; document per-language matrix; tests per language |
| Schema churn breaks `.carto/graph.json` consumers | Additive v2 only; `schema_version` gate in loader; UI untouched until Phase 4 |
| Scope creep into "rewrite everything in Rust" | Phase 6 strictly strangler-fig with golden tests; web UI stays TS |
| Perf regression on huge repos | Benchmark on a ≥5k-file repo before enabling by default; rayon parallelism; cache preserved |
| Working tree dirty during Phase 0 | Build on top of it; never discard user changes |
| Licensing | Ideas only, clean reimplementation, NOTICE credit (MIT); no copied source |

## 9. Decisions made (2026-11-24)

1. **Rust end-state:** Go CLI stays the shipping/reference artifact; Rust is an opt-in gated core rewrite (`--engine=rust`) after parity, not a prerequisite.
2. **CI gate:** warn-only / `continue-on-error` first; flip to blocking only after calibration + ~2–4 clean weeks. Never block on warnings in v1.
3. **UI scope:** CLI/JSON only for v1; no React work.
4. **Calibration repos:** allowed for local analysis; publish only aggregate numbers.
5. **Naming:** `carto analyze structure`, report "Structural Health Report", rule IDs `arch/*`, config `.carto/structure.toml`. "Slop" is informal shorthand only — not in the stable CLI surface.
6. **untangle-codebase integration:** inform `carto fix --prompt` output (inventory → first-principles redesign → fix → verify → rerun `carto analyze`). Our own prompt text; inspiration only, no copied text (no LICENSE in that repo).
7. **Neural tier:** tiny scorer first; CodeRankEmbed-137M (MIT) int8 ONNX re-ranker as a benchmarked spike; kill both if they don't beat the graph heuristic on carto's benchmark suite. No neural parsers, no default-on models.

## 9a. Research-driven feature adds (in priority order)

From `.workspace/research/2026-11-24-codebase-context-tools.md`:

1. `context_savings` field + optional `--verify` tiktoken check on context/pack/ask responses (credibility + built-in value demo).
2. `carto mcp` server exposing `ask`, `blast_radius`, `repo_map`, `structure` (every serious competitor is MCP-first; carto's biggest gap).
3. Edge confidence tags (`extracted` / `inferred`) for heuristic import resolution.
4. Incremental re-index by content hash on all commands (extend existing graph.json caching).
5. Blast radius with risk tiers + affected-test output.
6. `--compress` audit + `--token-budget` non-zero-exit CI guard.
7. Leiden/Louvain communities → architecture summary (also feeds deep-module scoring).
8. Git-churn × centrality ranking for `repo-map`.
9. Keep 4 B/token default; optional tiktoken mode; document error band.
10. `carto ask --answer-only` ("answer from graph, not files").
11. Consider SQLite graph storage past ~50k files (JSON cache limit).

Honest token claims to cite: code-review-graph ~6.65× vs grep baseline (not 63× vs whole corpus), codebase-memory-mcp 10× fewer tokens at 83% vs 92% answer quality (arXiv:2603.27277). Nobody benchmarks head-to-head vs aider's repo-map — carto can publish the first honest harness numbers.

## 10. Out of scope

- Fixing findings automatically (only prompt *generation*; never auto-apply).
- LLM calls inside `carto` itself.
- Publishing to crates.io/npm (separate decision).
- Changes to the token estimator (~4 bytes/token stays).
