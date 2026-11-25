---
title: "Small neural models for Code Cartographer: worth it for context selection?"
status: draft
tags: [research]
---

# Small neural models vs. PageRank/tree-sitter for budget-aware code context

**Question:** For a deterministic local/offline CLI that ranks symbols for a token-budgeted context window, can ~100–500M-param neural models improve token efficiency, and is any of it worth building?
**Date:** 2026-11-24. All URLs fetched this session.

## Answer

Keep the core deterministic (tree-sitter + graph ranking). The published evidence says learned retrieval is real but *incremental* on top of structural retrieval, and the best small model (CodeRankEmbed, 137M, MIT) buys a few points of localization quality at the cost of a ~550MB dependency and new failure modes. Ship nothing neural by default; prototype one optional flag (`--rerank=neural`) as a spike, benchmark it locally, and decide with data.

## 1. Candidate models (all verified on HF this session)

| Model | Params | License | Use | Local OK? | Notes |
|---|---|---|---|---|---|
| CodeBERT (microsoft/codebert-base) | ~125M | MIT (repo) | Encoder, embeddings | Yes, ONNX quants on HF (8 variants) | 2020, CodeSearchNet, 6 langs, dated |
| GraphCodeBERT | ~125M | MIT (repo) | Encoder + dataflow | Yes | 12L/768d/512ctx; adds data-flow edges at pretrain time |
| UniXcoder | ~125M | Apache-2.0 (card) | Enc/dec embeddings | Yes (7 quants listed) | 768-d; RepoCoder used it as dense retriever |
| CodeT5+ 220M (Salesforce/codet5p-220m) | 220M | BSD-3-Clause | Enc-dec, gen+embed | Yes (1 quant) | CoIR NDCG@10 45.9 (per CodeRankEmbed card) |
| StarEncoder (bigcode) | ~125M | **OpenRAIL-M (gated)** | Encoder | Restricted | 80+ langs, The Stack; license is the problem |
| Qwen2.5-Coder-0.5B / 1.5B | 0.49B / 1.54B | Apache-2.0 | **Generative** (not embedders) | Yes (GGUF, 65 quants) | Too big/generative for ranking duty |
| **CodeRankEmbed (nomic-ai)** | **137M** | **MIT** | **Bi-encoder retrieval** | **Yes, CPU-feasible; 17 quants listed** | CoIR 60.1 NDCG@10, CSN 77.9 MRR; 8192 ctx; 54.9k downloads/mo |
| Voyage code (voyage-code-3/4; API) | n/a (cloud) | proprietary | API benchmark only | No (offline CLI) | docs.voyageai.com; can't ship or call offline |
| Voyage-4-nano | small, open-weight | (open weights, see HF) | Embedding | Maybe | Announced 2026; not yet evaluated here — unverified |

CodeRankEmbed card (https://huggingface.co/nomic-ai/CodeRankEmbed, fetched 2026-11-24): 137M, MIT, 8k ctx, outperforms CodeSage-Large (1.3B, 59.4 CoIR) and voyage-code-002 (56.3). CodeBERT/GraphCodeBERT cards list no explicit license field; the microsoft/CodeBERT repo is MIT.

## 2. What the literature says about repo-level retrieval

- **RepoCoder** (arXiv:2303.12570, EMNLP 2023): RAG over repo snippets beats in-file context by >10% EM/ES; **a Jaccard bag-of-words retriever matched a UniXcoder dense retriever** ("the simple sparse retriever achieves equivalent performance", §4.2/App. B, fetched from arXiv HTML). The useful wins came from retrieval + iteration, not from the embedder.
- **RepoGraph** (arXiv:2410.14684, ICLR 2025): line-level tree-sitter-style graph + k-hop ego-graphs plugged into RAG / Agentless / SWE-agent / AutoCodeRover: **+2.3 to +2.7 resolve points on SWE-bench Lite** (e.g. Agentless 27.33→29.67; RAG 2.67→5.33), localization accuracy +2–5.6 points, at only **+11–32% tokens** (Agentless 42,376→47,323; RAG 11,736→15,439; SWE-agent 498k→519k). CrossCodeEval EM roughly doubled (DeepSeek-Coder-V2-Lite 10.2→19.7). This is graph-over-tokens, not neural-over-graph.
- **Agentless** (arXiv:2407.01489v2): 32% SWE-bench Lite at **$0.70/issue**; localization = LLM-prompted file ranking **plus** embedding retrieval (OpenAI text-embedding-3-small, 512-token chunks) unioned with LLM picks — i.e., embeddings as one of two weak-but-complementary signals, not the sole ranker.
- **LocAgent** (arXiv:2503.09089, ACL 2025): graph-guided localization (multi-hop over a hetero graph, with BM25) reported to improve localization vs prior methods; abstract/intro verified, detailed deltas not extracted — treat magnitude as unverified.
- **CoIR benchmark** (arXiv:2407.02883, ACL 2025; 10 tasks, 2M docs): even top dense retrievers land ~60 NDCG@10; retrieval is far from solved, so a 125M model will not "solve" context selection either. CodeRankEmbed tops small models at 60.1.
- **Token economics** (arXiv:2604.22750, Apr 2026, 8 frontier models on SWE-bench Verified): agentic tasks burn ~1000× more tokens than chat; input tokens dominate; same-task reruns vary up to 30×; accuracy plateaus at high token spend. Cutting input tokens has real money behind it, but gains saturate.
- RepoFusion (OpenReview) trains models to *use* repo context — training-time, not a shippable retriever.

**What moves the needle:** (1) graph structure from the AST (call/import/containment edges), (2) retrieval of *some* relevant context at all vs none, (3) cheap lexical signals fused with graph priors. Learned dense re-ranking is a refinement layer worth single-digit points, not the foundation.

## 3. Token-reduction evidence vs graph heuristics

No published head-to-head shows a ≤500M neural reranker beating a tuned PageRank/AST ranking on bytes-fetched-vs-solve-rate. Adjacent evidence: RepoCoder (Jaccard ≈ UniXcoder), RepoGraph (+tokens for +2-3 points), Agentless (embedding leg is one of two signals; graph/LLM legs carry the rest), CoIR (dense models ~60% NDCG@10 — far from oracle). Honest read: **expect ≈0–5% task-relevant context recall gain from a neural re-ranker over a good graph+BM25 baseline, measurable only on a large eval set.** That is below what a single better PageRank damping/personalization pass or a cheap feature scorer often buys. Numbers like "50% token reduction" floating around marketing pages are not from peer-reviewed head-to-heads I could verify.

## 4. Integration paths (Go/Rust CLI)

- **ONNX Runtime**: `onnxruntime-go` (Go) or `tract`/`ort` (Rust) — both run BERT-class encoders; CodeBERT/UniXcoder/CodeRankEmbed export cleanly via `optimum-cli export onnx`; quantized community repos exist for all three. Weights cached under a user cache dir; pin SHA + model revision in the flag's help text.
- **candle (Rust)**: native BERT impls exist; good if the CLI is Rust and you want zero C deps.
- **Python sidecar**: simplest for prototyping (sentence-transformers), worst for distribution — avoid as the shipped path.
- **Latency (estimate, not measured)**: ~125M encoder over 50 snippets ≤512 tokens ≈ 0.5–2 s single-thread-ish on a modern x86 core, <0.5 s int8-quantized multi-thread; first-run model download ~100–550 MB. Treat as estimate; measure before committing.
- Determinism: pin model revision, fix threads, record hash; fp32 vs int8 changes rankings slightly.

## 5. Risks

- **License**: avoid StarEncoder (OpenRAIL-M, gated). CodeBERT/GraphCodeBERT MIT (repo). UniXcoder/CodeT5+/Qwen2.5-Coder/CodeRankEmbed permissive. Voyage = API-only, can't meet "offline".
- **Freshness**: encoders are 2020–2024-era, 6–80 langs, frozen knowledge; tree-sitter tracks language grammars continuously. Neural ranker goes stale silently; graph signals don't.
- **Leakage/benchmarks**: CodeSearchNet/The Stack overlap GitHub; RepoEval used post-2022 repos deliberately. Don't quote CoIR numbers as "your gain". SWE-bench contamination is documented (Lite-SS paper).
- **Determinism/support**: two-runtime problem (Go/Rust + ONNX) in a tool that's otherwise pure logic + tree-sitter; error UX for missing model/download failure eats solo-maintainer time.

## 6. Verdict

**Ship now**
- Nothing neural. Keep/extend: tree-sitter graph + PageRank, plus a lexical (BM25/Jaccard) leg and simple feature scoring — this is the RepoCoder/Aider-validated 80% solution and costs zero deps. Add the token/solve-rate harness first so any later neural claim is measurable.

**Spike later (one weekend, behind `--rerank=neural`, default off)**
- CodeRankEmbed-137M (MIT) re-scoring the top ~20–50 graph candidates; ONNX int8 via onnxruntime-go or tract; compare vs graph-only on RepoEval/CrossCodeEval-style fixture; keep only if it beats baseline on *both* recall@budget and end-task metric, with pinned version + offline cache.
- A tiny learned scorer (logistic regression/GBDT over hand-crafted features: graph centrality, path depth, recency, lexical overlap) — cheaper, fully deterministic, and likely to capture most of what the neural net adds; preferred over the raw embedding.

**Avoid**
- Any default-on model, >500M models, StarEncoder/OpenRAIL-M, Qwen-Coder-as-ranker (generative, wrong tool), Voyage APIs in an "offline" CLI, and benchmark-driven claims without a local A/B.

## Confidence
Medium. Model/license facts verified against live cards/repos; benchmark deltas are from papers (fetched), but no head-to-head vs a tuned PageRank baseline exists in print — the central "learned ≈ small gain" claim is an inference from RepoCoder + RepoGraph + CoIR, flagged as such. Unverified: exact CPU latencies, LocAgent's quantitative deltas, voyage-4-nano quality, quantized-build numerical drift.

## Sources
- hf.co/microsoft/codebert-base, /graphcodebert-base, /unixcoder-base, /microsoft/CodeBERT (MIT, fetched 2026-11-24)
- hf.co/Salesforce/codet5p-220m (BSD-3), hf.co/bigcode/starencoder (OpenRAIL-M), hf.co/Qwen/Qwen2.5-Coder-{0.5B,1.5B} (Apache-2.0), hf.co/nomic-ai/CodeRankEmbed (MIT; CoIR 60.1)
- docs.voyageai.com/docs/embeddings (voyage-code-3/4; API-only)
- arXiv:2303.12570 (RepoCoder), arXiv:2407.01489 (Agentless), arXiv:2410.14684 (RepoGraph, ICLR 2025), arXiv:2503.09089 (LocAgent), arXiv:2407.02883 (CoIR, ACL 2025), arXiv:2604.22750 (token consumption, 2026)
