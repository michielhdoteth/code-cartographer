package codegraph

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"carto/lib/parser"
)

const SchemaVersion = 1

type Index struct {
	SchemaVersion int             `json:"schema_version"`
	RootPath      string          `json:"root_path"`
	BuiltAt       time.Time       `json:"built_at"`
	Files         map[string]File `json:"files"`
	Nodes         []Node          `json:"nodes"`
	Edges         []Edge          `json:"edges"`
	Warnings      []Warning       `json:"warnings,omitempty"`
	goImports     map[string]string `json:"-"`
}

type Warning struct {
	File    string `json:"file"`
	Message string `json:"message"`
}

type File struct {
	Hash     string          `json:"hash"`
	Language string          `json:"language"`
	Size     int64           `json:"size"`
	Lines    int             `json:"lines"`
	Nodes    []parser.Node   `json:"nodes"`
	Calls    []parser.Edge   `json:"calls,omitempty"`
	Imports  []parser.Import `json:"imports,omitempty"`
	Warning  string          `json:"warning,omitempty"`
}

type Node struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Kind            string `json:"kind"`
	File            string `json:"file"`
	StartLine       int    `json:"start_line,omitempty"`
	EndLine         int    `json:"end_line,omitempty"`
	Signature       string `json:"signature,omitempty"`
	Exported        bool   `json:"exported,omitempty"`
	RangeConfidence string `json:"range_confidence,omitempty"`
	Flags           []string `json:"flags,omitempty"`
}

type Edge struct {
	ID         string `json:"id"`
	Source     string `json:"source"`
	Target     string `json:"target"`
	Type       string `json:"type"`
	Confidence string `json:"confidence"`
	Role       string `json:"role,omitempty"`
	Line       int    `json:"line,omitempty"`
}

type Freshness struct {
	Fresh   bool     `json:"fresh"`
	Added   []string `json:"added"`
	Changed []string `json:"changed"`
	Deleted []string `json:"deleted"`
}

func cachePath(root string) string { return filepath.Join(root, ".carto", "graph.json") }

// Build updates the local graph cache, parsing only files whose content hash changed.
func Build(root string) (*Index, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve repository path: %w", err)
	}
	info, err := os.Stat(absRoot)
	if err != nil {
		return nil, fmt.Errorf("repository path: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("repository path is not a directory: %s", absRoot)
	}

	previous, _ := Load(absRoot)
	files := make(map[string]File)
	err = filepath.Walk(absRoot, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			if path != absRoot && shouldSkipDir(info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		lang := parser.DetectLanguage(path)
		if lang == parser.LanguageUnknown {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		h := sha256.Sum256(data)
		hash := hex.EncodeToString(h[:])
		rel, relErr := filepath.Rel(absRoot, path)
		if relErr != nil {
			return relErr
		}
		rel = filepath.ToSlash(rel)
		if previous != nil {
			if cached, ok := previous.Files[rel]; ok && cached.Hash == hash {
				files[rel] = cached
				return nil
			}
		}
		p := parser.GetParser(lang)
		if p == nil {
			files[rel] = File{Hash: hash, Language: string(lang), Size: int64(len(data)), Lines: strings.Count(string(data), "\n") + 1, Warning: "no parser registered"}
			return nil
		}
		result, parseErr := p.Parse(string(data), path)
		if parseErr != nil {
			files[rel] = File{Hash: hash, Language: string(lang), Size: int64(len(data)), Lines: strings.Count(string(data), "\n") + 1, Warning: parseErr.Error()}
			return nil
		}
		nodes := append([]parser.Node(nil), result.Nodes...)
		for _, exp := range result.Exports {
			found := false
			for _, node := range nodes {
				if node.Name == exp.Name && node.StartLine == exp.Line {
					found = true
					break
				}
			}
			if !found {
				nodes = append(nodes, parser.Node{Name: exp.Name, Type: exp.Type, File: path, StartLine: exp.Line, EndLine: exp.Line, Exported: exp.Exported})
			}
		}
		file := File{Hash: hash, Language: string(lang), Size: int64(len(data)), Lines: strings.Count(string(data), "\n") + 1, Nodes: nodes, Imports: result.Imports}
		for _, edge := range result.Edges {
			if edge.Type == parser.EdgeTypeCalls {
				file.Calls = append(file.Calls, edge)
			}
		}
		files[rel] = file
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan repository: %w", err)
	}

	index := &Index{SchemaVersion: SchemaVersion, RootPath: absRoot, BuiltAt: time.Now().UTC(), Files: files, Warnings: []Warning{}}
	filePaths := make([]string, 0, len(files))
	for path := range files {
		filePaths = append(filePaths, path)
	}
	index.goImports = loadGoImportMap(absRoot, filePaths)
	index.rebuildGraph()
	if err := index.Save(); err != nil {
		return nil, err
	}
	return index, nil
}

func Load(root string) (*Index, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(cachePath(absRoot))
	if err != nil {
		return nil, err
	}
	var index Index
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, err
	}
	if index.SchemaVersion != SchemaVersion || !samePath(index.RootPath, absRoot) || index.Files == nil {
		return nil, fmt.Errorf("graph cache is incompatible or belongs to another repository")
	}
	return &index, nil
}

func (index *Index) Save() error {
	if err := os.MkdirAll(filepath.Dir(cachePath(index.RootPath)), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("encode graph cache: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(cachePath(index.RootPath)), "graph-*.tmp")
	if err != nil {
		return fmt.Errorf("create graph cache: %w", err)
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}
	if err := os.Rename(tmpPath, cachePath(index.RootPath)); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("replace graph cache: %w", err)
	}
	return nil
}

func CheckFreshness(root string) (*Freshness, error) {
	index, err := Load(root)
	if err != nil {
		return nil, fmt.Errorf("no usable graph cache; run 'carto map <path>' first: %w", err)
	}
	absRoot, _ := filepath.Abs(root)
	result := &Freshness{Added: []string{}, Changed: []string{}, Deleted: []string{}}
	seen := make(map[string]bool)
	err = filepath.Walk(absRoot, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			if path != absRoot && shouldSkipDir(info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if parser.DetectLanguage(path) == parser.LanguageUnknown {
			return nil
		}
		rel, e := filepath.Rel(absRoot, path)
		if e != nil {
			return e
		}
		rel = filepath.ToSlash(rel)
		seen[rel] = true
		data, e := os.ReadFile(path)
		if e != nil {
			return nil
		}
		h := sha256.Sum256(data)
		hash := hex.EncodeToString(h[:])
		old, ok := index.Files[rel]
		if !ok {
			result.Added = append(result.Added, rel)
		} else if old.Hash != hash {
			result.Changed = append(result.Changed, rel)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	for path := range index.Files {
		if !seen[path] {
			result.Deleted = append(result.Deleted, path)
		}
	}
	for _, paths := range [][]string{result.Added, result.Changed, result.Deleted} {
		sort.Strings(paths)
	}
	result.Fresh = len(result.Added)+len(result.Changed)+len(result.Deleted) == 0
	return result, nil
}

func (index *Index) rebuildGraph() {
	index.Nodes = []Node{}
	index.Edges = []Edge{}
	index.Warnings = []Warning{}
	files := make([]string, 0, len(index.Files))
	for path := range index.Files {
		files = append(files, path)
	}
	sort.Strings(files)
	byName := make(map[string][]Node)
	for _, path := range files {
		f := index.Files[path]
		if f.Warning != "" {
			index.Warnings = append(index.Warnings, Warning{File: path, Message: f.Warning})
		}
		fileID := fileNodeID(path)
		index.Nodes = append(index.Nodes, Node{ID: fileID, Name: filepath.Base(path), Kind: "file", File: path, Flags: nodeFlags(path)})
		parsedNodes := append([]parser.Node(nil), f.Nodes...)
		sort.SliceStable(parsedNodes, func(i, j int) bool { return parsedNodes[i].StartLine < parsedNodes[j].StartLine })
		for i, parsed := range parsedNodes {
			if strings.TrimSpace(parsed.Name) == "" {
				continue
			}
			start, end, spanConfidence := parsed.StartLine, parsed.EndLine, "exact"
			if start < 1 {
				start = 1
			}
			if end <= start {
				spanConfidence = "estimated"
				end = f.Lines
				for j := i + 1; j < len(parsedNodes); j++ {
					if parsedNodes[j].StartLine > start {
						end = parsedNodes[j].StartLine - 1
						break
					}
				}
				if end < start {
					end = start
				}
			}
			n := Node{ID: symbolID(path, parsed), Name: parsed.Name, Kind: string(parsed.Type), File: path, StartLine: start, EndLine: end, Signature: parsed.Signature, Exported: parsed.Exported, RangeConfidence: spanConfidence, Flags: nodeFlags(path)}
			index.Nodes = append(index.Nodes, n)
			byName[normalizeName(parsed.Name)] = append(byName[normalizeName(parsed.Name)], n)
			index.Edges = append(index.Edges, Edge{ID: "contains:" + fileID + ":" + n.ID, Source: fileID, Target: n.ID, Type: "contains", Confidence: "exact", Line: n.StartLine})
		}
	}
	for _, path := range files {
		f := index.Files[path]
		for _, call := range f.Calls {
			name := normalizeName(call.Target)
			candidates := byName[name]
			line := call.Line
			if line == 0 {
				line = edgeLine(call.ID)
			}
			source := fileNodeID(path)
			bestSpan := int(^uint(0) >> 1)
			for _, n := range index.Nodes {
				if n.File == path && n.Kind != "file" && line >= n.StartLine && (n.EndLine == 0 || line <= n.EndLine) {
					span := n.EndLine - n.StartLine
					if span < bestSpan {
						bestSpan = span
						source = n.ID
					}
				}
			}
			if len(candidates) == 1 {
				index.Edges = append(index.Edges, Edge{ID: "call:" + source + ":" + candidates[0].ID + ":" + strconv.Itoa(line), Source: source, Target: candidates[0].ID, Type: "calls", Confidence: "name", Line: line})
			} else if len(candidates) > 1 {
				index.Edges = append(index.Edges, Edge{ID: "call:" + source + ":ambiguous:" + name + ":" + strconv.Itoa(line), Source: source, Target: "symbol-name:" + name, Type: "calls", Confidence: "ambiguous", Line: line})
			} else {
				index.Edges = append(index.Edges, Edge{ID: "call:" + source + ":external:" + name + ":" + strconv.Itoa(line), Source: source, Target: "external:" + name, Type: "calls", Confidence: "unresolved", Line: line})
			}
		}
		for _, imp := range f.Imports {
			target, confidence := resolveImport(path, imp.Source, files, index.goImports)
			targetID := "external:" + imp.Source
			if confidence == "exact" || confidence == "name" {
				targetID = fileNodeID(target)
			} else if confidence == "ambiguous" {
				targetID = "import-name:" + normalizeName(imp.Source)
			}
			index.Edges = append(index.Edges, Edge{ID: "import:" + fileNodeID(path) + ":" + targetID + ":" + strconv.Itoa(imp.Line), Source: fileNodeID(path), Target: targetID, Type: "imports", Confidence: confidence, Line: imp.Line})
		}
	}
	for i := range index.Edges {
		index.Edges[i].Role = edgeRole(index.Edges[i].Confidence)
	}
	sort.Slice(index.Nodes, func(i, j int) bool { return index.Nodes[i].ID < index.Nodes[j].ID })
	sort.Slice(index.Edges, func(i, j int) bool {
		if index.Edges[i].Source != index.Edges[j].Source {
			return index.Edges[i].Source < index.Edges[j].Source
		}
		if index.Edges[i].Target != index.Edges[j].Target {
			return index.Edges[i].Target < index.Edges[j].Target
		}
		return index.Edges[i].Line < index.Edges[j].Line
	})
}

func fileNodeID(path string) string { return "file:" + filepath.ToSlash(path) }

func nodeFlags(path string) []string {
	path = filepath.ToSlash(path)
	var flags []string
	if strings.HasSuffix(path, "_test.go") || strings.Contains(path, "/__tests__/") {
		flags = append(flags, "test")
	}
	if strings.HasSuffix(path, ".pb.go") || strings.HasSuffix(path, "_generated.go") ||
		strings.HasPrefix(path, "bin/") || strings.HasPrefix(path, "dist/") || strings.HasPrefix(path, "build/") {
		flags = append(flags, "generated")
	}
	return flags
}

func edgeRole(confidence string) string {
	if confidence == "exact" || confidence == "name" {
		return "extracted"
	}
	return "inferred"
}
func symbolID(path string, n parser.Node) string {
	return fmt.Sprintf("symbol:%s:%s:%s:%d:%d", filepath.ToSlash(path), n.Type, n.Name, n.StartLine, n.StartCol)
}
func normalizeName(name string) string {
	parts := strings.FieldsFunc(name, func(r rune) bool { return r == '.' || r == ':' || r == '/' })
	if len(parts) == 0 {
		return strings.ToLower(name)
	}
	return strings.ToLower(parts[len(parts)-1])
}
func edgeLine(id string) int {
	i := strings.LastIndex(id, ":")
	if i < 0 {
		return 0
	}
	n, _ := strconv.Atoi(id[i+1:])
	if n == 0 {
		return 1
	}
	return n
}
func shouldSkipDir(name string) bool {
	return name == "node_modules" || name == ".git" || name == "vendor" || name == "dist" || name == "build" || name == "__pycache__" || name == "target" || strings.HasPrefix(name, ".")
}
func samePath(a, b string) bool { return strings.EqualFold(filepath.Clean(a), filepath.Clean(b)) }

func resolveImport(sourceFile, importSource string, files []string, goImports map[string]string) (string, string) {
	if strings.HasSuffix(sourceFile, ".go") {
		if target, ok := goImports[strings.TrimSuffix(importSource, "/")]; ok {
			return target, "exact"
		}
	}
	clean := strings.TrimPrefix(filepath.ToSlash(importSource), "./")
	clean = strings.TrimPrefix(clean, "/")
	withoutExt := strings.TrimSuffix(clean, filepath.Ext(clean))
	relative := filepath.ToSlash(filepath.Join(filepath.Dir(sourceFile), clean))
	relativeStem := strings.TrimSuffix(relative, filepath.Ext(relative))
	var exact []string
	for _, candidate := range files {
		stem := strings.TrimSuffix(candidate, filepath.Ext(candidate))
		if candidate == clean || candidate == relative || stem == withoutExt || stem == relativeStem {
			exact = append(exact, candidate)
		}
	}
	if len(exact) == 1 {
		return exact[0], "exact"
	}
	if len(exact) > 1 {
		return "", "ambiguous"
	}
	base := strings.ToLower(filepath.Base(withoutExt))
	var named []string
	for _, candidate := range files {
		if strings.ToLower(strings.TrimSuffix(filepath.Base(candidate), filepath.Ext(candidate))) == base {
			named = append(named, candidate)
		}
	}
	if len(named) == 1 {
		return named[0], "name"
	}
	return "", "unresolved"
}
