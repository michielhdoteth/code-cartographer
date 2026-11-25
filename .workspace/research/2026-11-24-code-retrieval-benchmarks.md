---
title: "Public datasets/benchmarks for code-context retrieval (carto)"
status: draft
tags: [research, retrieval, benchmarks]
---

# Public datasets and benchmarks for code-context retrieval

Decision this unblocks: what to evaluate `carto`'s BM25 retriever (and a possible small neural re-ranker) against locally, without training on tainted data.

Date checked: 2026-11-24. All URLs fetched this session unless marked unverified.

## Summary table

| Dataset | Task | Size | License | Access | Usable for BM25 tuning? |
|---|---|---|---|---|---|
| CoIR (10 tasks, `coir-eval` on HF `CoIR-Retrieval`) | Text→code, code→code, code→text, hybrid QA | 10 datasets, ~2M docs total; e.g. CSN 905k train q / 1M corpus; CosQA 20.6k pairs (500-q test) | Code Apache-2.0; datasets inherit source licenses: APPS/CSN/CosQA "MIT, non-commercial research", CodeTransOcean + SyntheticText2SQL + CodeFeedback Apache-2.0, StackOverflow QA CC-BY-SA 3.0 | `pip install coir-eval`; HF `coir`-style parquet, MTEB-compatible | **Eval only.** Not for training a commercial tool (several sub-tasks non-commercial). BM25 baseline published (paper Table 3, avg NDCG@10 ≈ 29.8) |
| CodeSearchNet (github/CodeSearchNet, archived 2023) | Code-to-docstring (train) + 99-query human-annotated challenge (test) | 2.33M doc↔code pairs (Go 348k, Java 543k, JS 158k, PHP 717k, Py 504k, Ruby 57k); 6.45M functions total; NDCG/MRR | Repo MIT; corpus per-language licenses in data; CoIR lists it "MIT non-commercial research"; HF mirrors (e.g. `code_search_net`, `Nan-Do/code-search-net-python` apache-2.0 tag) | S3 (~3.5–20 GB, README is inconsistent) or HF parquet mirrors | **Eval only**, legal grey zone for training commercial models; standard BM25/Elasticsearch baseline reported in paper |
| CodeXGLUE (microsoft/CodeXGLUE) | Code search (NL→code), summarization, etc. | CSN split same as above + Bing-query test set (WebQueryTest) | Repo MIT; **datasets = C-UDA (Computational Use of Data Agreement)** — restricts use | GitHub + HF `code_x_glue_*` | No for commercial; eval only, and even that is awkward |
| RepoBench (Leolty/repobench, ICLR 2024) | Repo-level completion; **RepoBench-R = retrieval of cross-file snippet** (gold = `gold_snippet_index`) | Python + Java; cff/cfr/if settings; (sizes not restated here — check `data/README.md`) | Repo CC-BY-4.0, but HF `tianyang/repobench-p` says **CC-BY-NC-ND-4.0** (conflict; NC+ND) | HF `tianyang/repobench_python/java_v1.1` | **Avoid for training** (ND); ok for a private sanity check, nothing more |
| RepoEval (RepoCoder, microsoft/CodeT) | Repo-level completion (line/API/function body); retrieval is a means, not a labeled task | "latest real-world repos"; exact counts not restated in sources fetched — check paper arXiv:2303.12570 / repo | Repo `microsoft/CodeT` — license not verified this session | github.com/microsoft/CodeT/tree/main/RepoCoder | No retrieval labels → not directly usable as retrieval eval |
| SWE-bench / Verified / Multimodal | Issue→patch; gold files = files touched by gold patch (retrieval-able); BM25-13K/27K/40K/50K variants ship pre-retrieved contexts | Full 2294 / Lite 300 / **Verified 500** / Multimodal-v2 480 (JS/TS) / Multi-SWE-bench 1,632 (7 langs) | Harness MIT; HF datasets: Verified card shows no explicit license; Multimodal paper: repos only allow **non-commercial** use; Multi-SWE-bench — check each repo | HF `SWE-bench/*`, `swebench` pip | **Yes for eval** (file-level recall@k vs gold patch files). Not for training a commercial tool without legal review |
| SWE-smith trajectories (`SWE-bench/SWE-smith-trajectories`) | Agent traces w/ tool calls; derive file-visit labels from successful runs | ~5,016 trajectories shown on card; dataset tag says 10k–100k rows, 76k rows/3 splits, 4.22 GB; 128 repos | **MIT** (HF license tag) | HF parquet | **Yes** — best available source for training/eval labels from real agent navigation |
| CrossCodeEval (amazon-science/cceval, NeurIPS 2023) | Cross-file completion; usable as retrieval benchmark (their framing) | ~10k examples, ~1k repos, 4 langs (Py/Java/TS/C#) | Not verified this session — check repo (Amazon Science, typically Apache-2.0) | github.com/amazon-science/cceval | Yes for eval once license confirmed; gold = cross-file dependency |
| "RepoRecall" | — | — | — | — | **Does not exist** as a benchmark; only vendor products/repos. You likely mean RepoEval/RepoBench. |

## 1. CoIR specifics
Paper: arXiv:2407.02883 (ACL 2025 main, v3 2025-06-06). 10 datasets / 8 subtasks / 7 domains / 14 PLs, EN only. Tasks: text→code (APPS 3.8k test, CosQA 500 test, SyntheticText2SQL 6K test), code→text (CSN), code→code (CSN-CCR context retrieval; CodeTransOcean DL/Contest), hybrid (StackOverflow-QA 2K test, CodeFeedback-ST 31K, CodeFeedback-MT 13K). Download: `pip install coir-eval`, `load_dataset("CoIR-Retrieval/<name>")`; BM25 reference NDCG@10 per task in paper Table 3 (avg 29.79; BM25 is decent on StackOverflow-QA 56.8, weak on APPS 0.95). No explicit hard negatives — single gold per query, nDCG@10; negatives are the rest of the corpus (in-batch if you train). Fits carto's "code-to-code + text-to-code" needs; skip Text2SQL and CodeFeedback if you want code-specific signal.

## 2. CodeSearchNet / CodeXGLUE
CSN: 80/10/10 split, queries = docstrings (train) vs 99 human NL queries + 4,026 annotations (test challenge). Query style mismatch is the known weakness: docstrings share vocabulary with code, real queries don't. CodeXGLUE adds a Bing-query test slice but under C-UDA. HF mirrors exist in parquet; sizes above verified from paper (Table 1) and HF.

## 3. Repo-family
RepoBench retrieval labels = `gold_snippet_index` per instance (cross-file snippet) — real retrieval labels, but licensed NC-ND on HF (repo says CC-BY-4.0 — conflicting; treat as NC-ND to be safe). RepoEval has no ranked relevance labels. No established "RepoRecall".

## 4. SWE-bench family
Gold patch file paths give file-level retrieval labels for free (parse `patch` diffs). Verified (500, 12 Python repos) is the cleanest anchor; Multimodal covers JS/TS (relevant to carto's TS) but is non-commercial per the paper; Multi-SWE-bench adds Go/Rust/etc. BM25 variants (`SWE-bench_bm25_13K/27K/40K`, `..._50k_llama`, `SWE-bench_oracle`) exist for RAG-style context eval. Retrieval is not the scored metric — you compute recall@k of gold files yourself.

## 5. Agent traces / navigation benchmarks
SWE-smith trajectories (HF, MIT) are the only release I verified with real tool-call traces from which file-visit labels can be derived. CoIR/CSN have none. `RepoQA`-style "find the file" agent benchmarks exist in papers but no maintained public dataset verified — treat as unverified.

## 6. Local/offline use
MTEB/CoIR subsets download as parquet (CosQA total 3.6 MB; CSN ~GBs). All of the above work offline after download. For a commercial tool: training on MIT/Apache data (SWE-smith trajectories, SyntheticText2SQL, CodeTransOcean, CodeFeedback) is fine; APPS/CSN/CosQA/CodeXGLUE/RepoBench/SWE-bench-Multimodal are eval-only or off-limits. Check each HF card's license tag at pin time.

## 7. Metrics & known pitfalls
Standard: **nDCG@10** (CoIR/BEIR/CoIR), **MRR** (CSN test), **Recall@k / file-level recall** (SWE-bench-style, k∈{1,5,10}, plus context-token budget). MAP used by CoIR as secondary. Vocabulary mismatch is documented since CodeSearchNet (Elasticsearch/BM25 competitively strong on docstring queries precisely because of term overlap; fails on short NL queries — that's the CosQA effect; CoIR shows BM25 avg 29.8 NDCG@10 vs dense models 36–56). See also: Cambronero et al. 2019 "When Deep Learning Met Code Search" (arXiv:1905.03813); Thakur et al. 2021 BEIR (arXiv:2104.08663) on BM25's strong zero-shot baseline; and the 2026 Findings-of-EMNLP paper "On the Challenges and Opportunities of Learned Sparse Retrieval for Code" (arXiv, aclanthology 2026.findings-emnlp.193; PDF binary, contents unverified) on subword/IDF issues in code. On "IDF leakage": boilerplate identifiers and generated code inflate document frequency; mitigate with camelCase/snake_case splitting, code stopword lists, and field weighting (BM25F over symbol vs body vs path) — but verify on your own eval, don't assume.

## 8. BM25 for code — references
- CoIR Table 3 (BM25 baselines per subtask) — the cleanest published code-specific BM25 numbers.
- CodeSearchNet paper §4.2 (Elasticsearch baseline, NDCG@within ≈ 0.34 avg).
- BM25S (arXiv:2407.03618) — fast Python/NumPy BM25; good for local tooling. Pyserini/Lucene is the reference implementation for BM25/BM25F. No published BM25F-for-code parameter study found — that part is genuinely open; tune (k1, b, field weights, tokenizer) on your own eval set.

## Recommendation (smallest defensible eval)
1. **Build 100–200 queries from this repo** (git log + issues; label gold files by hand; cover symbol-name, NL-intent, stack-trace, and "where is X handled" query types). Compute Recall@{1,5,10} + MRR. This is your regression gate for carto.
2. **Sanity-check anchor**: CoIR `CosQA` test (500 q, text→code) + one code→code task (CodeSearchNet-CCR sample) for BM25 vs dense sanity — eval-only, never trained on, note license. Plus SWE-bench Verified file-recall@k if you want a repo-level number (parse `patch` for gold files; Python-only, 500 instances).
3. **Do not** train or distill on CosQA/APPS/CSN/CodeXGLUE/RepoBench/SWE-bench for a commercial tool; if you need training pairs, use SWE-smith trajectories (MIT) + CodeFeedback/SyntheticText2SQL (Apache-2.0) + Program-Aided self-labeling on your own repos.
4. Ship BM25 with identifier-splitting + field weighting; measure before adding any neural re-ranker. CoIR suggests dense gains over BM25 are real but modest on hybrid tasks — a re-ranker only pays off if your lexical features are already maxed.

## Conflicts / gaps
- RepoBench license conflict (GitHub CC-BY-4.0 vs HF CC-BY-NC-ND-4.0) — treat as NC-ND.
- SWE-bench HF datasets carry no explicit license tag; GitHub harness is MIT; Multimodal paper declares repos non-commercial-only. Legal review recommended before shipping derived data.
- CSN size discrepancy (README says 3.5 GB and ~20 GB in two places; table above uses paper counts).
- "RepoRecall" as a named benchmark: not found; likely a conflation.
- Hard-negative construction for code retrieval: no public code benchmark ships explicit hard negatives; you must mine them (e.g., same-file random negatives) yourself.

Sources (all fetched 2026-11-24): arxiv.org/abs/2407.02883 + /html/2407.02883v3; github.com/CoIR-team/coir; arxiv.org/abs/1909.09436; github.com/github/CodeSearchNet; github.com/microsoft/CodeXGLUE; github.com/Leolty/repobench; huggingface.co/datasets/tianyang/repobench-p; arxiv.org/abs/2303.12570; github.com/microsoft/CodeT/tree/main/RepoCoder; github.com/SWE-bench/SWE-bench; huggingface.co/datasets/SWE-bench/{SWE-bench_Verified,SWE-smith-trajectories,SWE-bench_Multimodal}; huggingface.co/datasets/CoIR-Retrieval/cosqa; arxiv.org/abs/2306.03091; github.com/amazon-science/cceval (URL only, contents unverified); aclanthology.org/2026.findings-emnlp.193 (existence only).
