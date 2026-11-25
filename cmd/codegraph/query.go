package codegraph

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type QueryHit struct {
	Node    Node    `json:"node"`
	Score   float64 `json:"score"`
	Snippet string  `json:"snippet,omitempty"`
}

type TraceHit struct {
	Direction  string `json:"direction"`
	Node       Node   `json:"node"`
	Confidence string `json:"confidence"`
	Line       int    `json:"line,omitempty"`
	Depth      int    `json:"depth"`
}

func (index *Index) Ask(query string, limit, tokenBudget int) []QueryHit {
	return index.rank(query, limit, tokenBudget)
}

func (index *Index) Trace(name string) []TraceHit {
	return index.TraceDepth(name, 1, "both")
}

// TraceDepth walks resolved call edges up to maxDepth hops. Direction accepts
// callers, callees, or both. Ambiguous direct callers are retained as evidence.
func (index *Index) TraceDepth(name string, maxDepth int, direction string) []TraceHit {
	if maxDepth <= 0 {
		maxDepth = 1
	}
	if direction == "" {
		direction = "both"
	}
	var targets []Node
	for _, node := range index.Nodes {
		if node.Kind != "file" && strings.EqualFold(node.Name, name) {
			targets = append(targets, node)
		}
	}
	if len(targets) == 0 {
		for _, node := range index.Nodes {
			if node.Kind == "file" && strings.EqualFold(node.File, name) {
				targets = append(targets, node)
			}
		}
	}
	result := []TraceHit{}
	type state struct {
		node  Node
		depth int
	}
	seen := make(map[string]bool)
	for _, target := range targets {
		queue := []state{{node: target}}
		visited := map[string]bool{target.ID: true}
		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]
			if current.depth >= maxDepth {
				continue
			}
			for _, edge := range index.Edges {
				if edge.Type != "calls" {
					continue
				}
				var nextID, hitDirection, confidence string
				line := edge.Line
				if direction == "callers" || direction == "both" {
					if edge.Target == current.node.ID {
						nextID, hitDirection, confidence = edge.Source, "caller", edge.Confidence
					} else if current.depth == 0 && edge.Confidence == "ambiguous" && edge.Target == "symbol-name:"+normalizeName(current.node.Name) {
						nextID, hitDirection, confidence = edge.Source, "caller", "ambiguous"
					}
				}
				if nextID == "" && (direction == "callees" || direction == "both") && edge.Source == current.node.ID {
					nextID, hitDirection, confidence = edge.Target, "callee", edge.Confidence
				}
				next, ok := index.nodeByID(nextID)
				if !ok {
					continue
				}
				key := hitDirection + ":" + next.ID
				if !seen[key] {
					result = append(result, TraceHit{Direction: hitDirection, Node: next, Confidence: confidence, Line: line, Depth: current.depth + 1})
					seen[key] = true
				}
				if !visited[next.ID] {
					visited[next.ID] = true
					queue = append(queue, state{node: next, depth: current.depth + 1})
				}
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Direction != result[j].Direction {
			return result[i].Direction < result[j].Direction
		}
		if result[i].Depth != result[j].Depth {
			return result[i].Depth < result[j].Depth
		}
		if result[i].Node.File != result[j].Node.File {
			return result[i].Node.File < result[j].Node.File
		}
		return result[i].Node.StartLine < result[j].Node.StartLine
	})
	return result
}

func (index *Index) Outline(path string) []Node {
	clean := filepath.ToSlash(filepath.Clean(path))
	var result []Node
	for _, node := range index.Nodes {
		if node.Kind != "file" && (node.File == clean || strings.HasSuffix(node.File, "/"+clean)) {
			result = append(result, node)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].StartLine != result[j].StartLine {
			return result[i].StartLine < result[j].StartLine
		}
		return result[i].Name < result[j].Name
	})
	return result
}

func (index *Index) nodeByID(id string) (Node, bool) {
	for _, node := range index.Nodes {
		if node.ID == id {
			return node, true
		}
	}
	return Node{}, false
}

func nodeText(root string, node Node) string {
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(node.File)))
	if err != nil {
		return ""
	}
	lines := strings.Split(string(data), "\n")
	start := node.StartLine
	end := node.EndLine
	if start < 1 {
		start = 1
	}
	if end < start {
		end = start
	}
	if end > start+24 {
		end = start + 24
	}
	if end > len(lines) {
		end = len(lines)
	}
	if start > len(lines) {
		return ""
	}
	return strings.Join(lines[start-1:end], "\n")
}


