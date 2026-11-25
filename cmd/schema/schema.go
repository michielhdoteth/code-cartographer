package schema

type Field struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type Definition struct {
	Command       string  `json:"command"`
	SchemaVersion int     `json:"schema_version"`
	Fields        []Field `json:"fields"`
}

var definitions = map[string]Definition{
	"context": {
		Command: "context", SchemaVersion: 1,
		Fields: []Field{
			{"file_path", "string", "Absolute path of the main file"},
			{"tokens_used", "integer", "Estimated tokens in returned blocks"},
			{"token_budget", "integer", "Requested budget"},
			{"truncated", "boolean", "True when the budget cut content"},
			{"total_symbols", "integer", "Symbols parsed from the main file"},
			{"savings", "object?", "Estimated whole-repo comparison; basis is stated"},
			{"cohesion", "object?", "Structural cohesion score and top-level areas (with --cohesion)"},
			{"view", "string?", "skeleton, chunk, or selection when a view flag is used"},
			{"skeleton", "object?", "AST symbol skeleton with no bodies (--outline)"},
			{"chunk", "object?", "One AST-aligned chunk (--chunk)"},
			{"selection", "object?", "Requested symbol/line ranges"},
			{"blocks", "array", "Context blocks: file_path, tokens, content, type"},
		},
	},
	"structure": {
		Command: "structure", SchemaVersion: 1,
		Fields: []Field{
			{"version", "integer", "Report schema version"},
			{"root_path", "string", "Analyzed root"},
			{"summary", "object", "modules, critical, warning, info, healthy, total"},
			{"modules", "array", "path, files, exports, lines, internal_edges, fan_in, fan_out, depth"},
			{"findings", "array", "id, rule, severity, module, message, evidence, hint, files"},
			{"suppressed", "array?", "IDs suppressed by the allowlist"},
		},
	},
	"ask": {
		Command: "ask", SchemaVersion: 1,
		Fields: []Field{
			{"query", "string", "Original query"},
			{"hits", "array", "node, BM25F score, optional snippet"},
			{"facts", "array?", "graph facts without snippets with --answer-only"},
			{"savings", "object", "Estimated tokens saved versus whole-repo dump"},
			{"warnings", "array", "Parser warnings from the graph build"},
		},
	},
	"brief": {
		Command: "brief", SchemaVersion: 1,
		Fields: []Field{
			{"root_path", "string", "Repository root"},
			{"token_budget_estimate", "integer", "Requested budget"},
			{"tokens_used_estimate", "integer", "Estimated tokens emitted"},
			{"files", "array", "path, lines, symbols, imports, top_symbols"},
			{"communities", "array?", "connected subsystem groups"},
			{"suggested_next", "string", "Recommended next command"},
		},
	},
	"chunk": {
		Command: "chunk", SchemaVersion: 1,
		Fields: []Field{
			{"file_path", "string", "Target file"},
			{"chunker_version", "integer", "Chunking algorithm version"},
			{"total_chunks", "integer", "Number of chunks"},
			{"chunks", "array", "chunk_index, chunk_id, line_start, line_end, tokens, symbols"},
			{"content", "string", "Chunk body (only when --chunk is used)"},
			{"next_chunk", "integer", "Next chunk index or -1"},
		},
	},
	"search": {
		Command: "search", SchemaVersion: 1,
		Fields: []Field{
			{"query", "string", "Original query"},
			{"path", "string", "Searched repository root"},
			{"results", "array", "file, line, context, BM25F score"},
		},
	},
	"impact": {
		Command: "impact", SchemaVersion: 1,
		Fields: []Field{
			{"seed_file", "string", "File whose blast radius was measured"},
			{"risk_tier", "string", "low, medium, high, or critical"},
			{"risk_score", "number", "Depth-weighted risk score"},
			{"affected_files", "array", "file, depth, score, via_symbols"},
			{"hubs", "array", "file, fan_in, pct_of_repo, symbol"},
			{"concentrated_risk", "object?", "critical_symbols and safe_symbols in the seed file"},
			{"affected_tests", "array?", "test files near the affected set"},
			{"test_gap_count", "integer", "affected files without a matching test"},
		},
	},
	"repo-map": {
		Command: "repo-map", SchemaVersion: 1,
		Fields: []Field{
			{"root_path", "string", "Repository root"},
			{"tokens_used_estimate", "integer", "Estimated tokens emitted"},
			{"token_budget_estimate", "integer", "Requested budget"},
			{"savings", "object", "Estimated savings with stated basis"},
			{"entries", "array", "Ranked declarations with signatures"},
			{"imports", "array?", "Ranked import summaries"},
		},
	},
	"pack": {
		Command: "pack", SchemaVersion: 1,
		Fields: []Field{
			{"root_path", "string", "Repository root"},
			{"tokens_used_estimate", "integer", "Estimated tokens emitted"},
			{"savings", "object", "Estimated savings with stated basis"},
			{"files_included", "integer", "Files selected"},
			{"files_omitted", "integer", "Files excluded by globs or budget"},
			{"content", "string", "Packed context"},
			{"summary", "string?", "compact inventory summary"},
			{"tree", "array?", "included file paths"},
			{"token_ledger", "object?", "source, emitted, and budget tokens"},
		},
	},
	"doctor": {
		Command: "doctor", SchemaVersion: 1,
		Fields: []Field{
			{"schema_version", "integer", "Report schema version"},
			{"root_path", "string", "Repository root"},
			{"status", "string", "ok, warn, or error"},
			{"files_in_graph", "integer", "Indexed source files"},
			{"symbols_in_graph", "integer", "Indexed non-file symbols"},
			{"checks", "array", "name, status, detail"},
		},
	},
	"fix": {
		Command: "fix", SchemaVersion: 1,
		Fields: []Field{
			{"finding_id", "string", "Stable ID from carto analyze structure"},
			{"rule", "string", "Rule that produced the finding"},
			{"prompt", "string", "LLM-ready refactor prompt; no model is called"},
		},
	},
	"analyze": {
		Command: "analyze", SchemaVersion: 1,
		Fields: []Field{
			{"type", "string", "Requested analysis type"},
			{"path", "string", "Analyzed root"},
			{"duration", "integer", "Analysis time in milliseconds"},
			{"results", "object", "Analysis-specific payloads (dead-code, complexity, circular, structure, health)"},
		},
	},
	"agent-bench": {
		Command: "agent-bench", SchemaVersion: 1,
		Fields: []Field{
			{"schema_version", "integer", "Report schema version"},
			{"cases", "array", "Task results: id, query, tokens_used_estimate, recall, relevant, returned"},
			{"mean_tokens", "number", "Mean tokens used per task"},
			{"whole_repo_tokens_estimate", "integer", "Baseline token estimate for all indexed source"},
			{"mean_recall", "number", "Mean fraction of expected symbols/files retrieved"},
		},
	},
}

func Get(name string) (Definition, bool) {
	d, ok := definitions[name]
	return d, ok
}

func List() []string {
	out := make([]string, 0, len(definitions))
	for name := range definitions {
		out = append(out, name)
	}
	return out
}
