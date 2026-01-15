"""
Code Cartographer Models

Core data structures for representing code structure and relationships.
Supports AST-based code mapping with full relationship tracking.
"""

import uuid
from dataclasses import dataclass, field
from typing import Dict, List, Optional, Any
from datetime import datetime


@dataclass
class Position:
    line: int
    column: int
    end_line: Optional[int] = None
    end_column: Optional[int] = None

    def to_dict(self) -> Dict:
        return {
            "line": self.line,
            "column": self.column,
            "end_line": self.end_line,
            "end_column": self.end_column,
        }

    @classmethod
    def from_dict(cls, data: Dict) -> "Position":
        return cls(
            line=data.get("line", 0),
            column=data.get("column", 0),
            end_line=data.get("end_line"),
            end_column=data.get("end_column"),
        )


@dataclass
class CodeLocation:
    file_id: str
    position: Position
    full_path: str

    def to_dict(self) -> Dict:
        return {
            "file_id": self.file_id,
            "position": self.position.to_dict(),
            "full_path": self.full_path,
        }

    @classmethod
    def from_dict(cls, data: Dict) -> "CodeLocation":
        return cls(
            file_id=data["file_id"],
            position=Position.from_dict(data["position"]),
            full_path=data["full_path"],
        )


NODE_TYPES = frozenset({
    "module", "package", "class", "interface", "function", "method",
    "property", "attribute", "variable", "constant", "enum", "enum_value",
    "type_alias", "import", "export", "struct", "trait", "impl",
    "namespace", "file", "comment", "decorator",
})

EDGE_TYPES = frozenset({
    "contains", "inherits", "implements", "imports", "exports", "calls",
    "references", "assigns", "returns", "throws", "param", "type_alias",
    "decorates", "depends_on", "extends", "uses", "contains_type",
    "instanceof", "test_for", "configuration", "extends_type",
})

INHERITANCE_EDGES = frozenset({"inherits", "implements", "extends", "extends_type"})
DEPENDENCY_EDGES = frozenset({"imports", "depends_on", "uses", "references"})
NESTING_EDGES = frozenset({"contains", "contains_type"})

SUPPORTED_LANGUAGES = {
    "python": {".py", ".pyi"},
    "javascript": {".js", ".jsx"},
    "typescript": {".ts", ".tsx"},
    "java": {".java"},
    "rust": {".rs"},
    "go": {".go"},
    "cpp": {".cpp", ".cc", ".cxx", ".h", ".hpp", ".hxx", ".c", ".c++"},
    "ruby": {".rb", ".erb"},
    "php": {".php"},
}

LANGUAGE_FILE_EXTENSIONS = {
    ext: lang for lang, exts in SUPPORTED_LANGUAGES.items() for ext in exts
}


@dataclass
class CodeNode:
    id: str
    node_type: str
    name: str
    qualified_name: str
    file_id: str
    start_position: Position
    end_position: Optional[Position] = None
    parent_id: Optional[str] = None
    children_ids: List[str] = field(default_factory=list)
    attributes: Dict[str, Any] = field(default_factory=dict)
    docstring: Optional[str] = None
    decorators: List[str] = field(default_factory=list)
    annotations: List[str] = field(default_factory=list)
    visibility: str = "public"
    is_abstract: bool = False
    is_static: bool = False
    is_async: bool = False
    return_type: Optional[str] = None
    parameters: List[Dict] = field(default_factory=list)
    signatures: List[str] = field(default_factory=list)
    complexity: int = 1
    code_snippet: Optional[str] = None

    def __post_init__(self):
        if self.node_type not in NODE_TYPES:
            raise ValueError(f"Invalid node_type: {self.node_type}")

    def to_dict(self) -> Dict:
        return {
            "id": self.id,
            "type": self.node_type,
            "name": self.name,
            "qualified_name": self.qualified_name,
            "file_id": self.file_id,
            "start_position": self.start_position.to_dict(),
            "end_position": self.end_position.to_dict() if self.end_position else None,
            "parent_id": self.parent_id,
            "children_ids": self.children_ids,
            "attributes": self.attributes,
            "docstring": self.docstring,
            "decorators": self.decorators,
            "annotations": self.annotations,
            "visibility": self.visibility,
            "is_abstract": self.is_abstract,
            "is_static": self.is_static,
            "is_async": self.is_async,
            "return_type": self.return_type,
            "parameters": self.parameters,
            "signatures": self.signatures,
            "complexity": self.complexity,
            "code_snippet": self.code_snippet,
        }

    @classmethod
    def from_dict(cls, data: Dict) -> "CodeNode":
        return cls(
            id=data.get("id", str(uuid.uuid4())),
            node_type=data["type"],
            name=data["name"],
            qualified_name=data["qualified_name"],
            file_id=data["file_id"],
            start_position=Position.from_dict(data["start_position"]),
            end_position=Position.from_dict(data["end_position"]) if data.get("end_position") else None,
            parent_id=data.get("parent_id"),
            children_ids=data.get("children_ids", []),
            attributes=data.get("attributes", {}),
            docstring=data.get("docstring"),
            decorators=data.get("decorators", []),
            annotations=data.get("annotations", []),
            visibility=data.get("visibility", "public"),
            is_abstract=data.get("is_abstract", False),
            is_static=data.get("is_static", False),
            is_async=data.get("is_async", False),
            return_type=data.get("return_type"),
            parameters=data.get("parameters", []),
            signatures=data.get("signatures", []),
            complexity=data.get("complexity", 1),
            code_snippet=data.get("code_snippet"),
        )

    def add_child(self, child_id: str):
        if child_id not in self.children_ids:
            self.children_ids.append(child_id)

    def update_complexity(self, complexity: int):
        self.complexity = max(self.complexity, complexity)


@dataclass
class CodeEdge:
    id: str
    edge_type: str
    source_id: str
    target_id: str
    attributes: Dict[str, Any] = field(default_factory=dict)
    label: Optional[str] = None
    weight: float = 1.0

    def __post_init__(self):
        if self.edge_type not in EDGE_TYPES:
            raise ValueError(f"Invalid edge_type: {self.edge_type}")

    def to_dict(self) -> Dict:
        return {
            "id": self.id,
            "type": self.edge_type,
            "source": self.source_id,
            "target": self.target_id,
            "attributes": self.attributes,
            "label": self.label,
            "weight": self.weight,
        }

    @classmethod
    def from_dict(cls, data: Dict) -> "CodeEdge":
        return cls(
            id=data.get("id", str(uuid.uuid4())),
            edge_type=data["type"],
            source_id=data["source"],
            target_id=data["target"],
            attributes=data.get("attributes", {}),
            label=data.get("label"),
            weight=data.get("weight", 1.0),
        )

    def is_inheritance(self) -> bool:
        return self.edge_type in INHERITANCE_EDGES

    def is_dependency(self) -> bool:
        return self.edge_type in DEPENDENCY_EDGES

    def is_nesting(self) -> bool:
        return self.edge_type in NESTING_EDGES


@dataclass
class CodeFile:
    id: str
    path: str
    relative_path: str
    language: str
    size_bytes: int = 0
    line_count: int = 0
    last_modified: Optional[str] = None
    checksum: Optional[str] = None
    node_ids: List[str] = field(default_factory=list)
    is_test: bool = False
    is_generated: bool = False
    is_ignored: bool = False
    metadata: Dict[str, Any] = field(default_factory=dict)

    def __post_init__(self):
        if self.id is None:
            self.id = str(uuid.uuid4())

    def to_dict(self) -> Dict:
        return {
            "id": self.id,
            "path": self.path,
            "relative_path": self.relative_path,
            "language": self.language,
            "size_bytes": self.size_bytes,
            "line_count": self.line_count,
            "last_modified": self.last_modified,
            "checksum": self.checksum,
            "node_ids": self.node_ids,
            "is_test": self.is_test,
            "is_generated": self.is_generated,
            "is_ignored": self.is_ignored,
            "metadata": self.metadata,
        }

    @classmethod
    def from_dict(cls, data: Dict) -> "CodeFile":
        return cls(
            id=data.get("id", str(uuid.uuid4())),
            path=data["path"],
            relative_path=data["relative_path"],
            language=data["language"],
            size_bytes=data.get("size_bytes", 0),
            line_count=data.get("line_count", 0),
            last_modified=data.get("last_modified"),
            checksum=data.get("checksum"),
            node_ids=data.get("node_ids", []),
            is_test=data.get("is_test", False),
            is_generated=data.get("is_generated", False),
            is_ignored=data.get("is_ignored", False),
            metadata=data.get("metadata", {}),
        )

    def add_node(self, node_id: str):
        if node_id not in self.node_ids:
            self.node_ids.append(node_id)


@dataclass
class CodeMap:
    id: str
    name: str
    version: str = "1.0.0"
    created_at: str = field(default_factory=lambda: datetime.now().isoformat())
    updated_at: str = field(default_factory=lambda: datetime.now().isoformat())
    root_path: str = ""
    files: Dict[str, CodeFile] = field(default_factory=dict)
    nodes: Dict[str, CodeNode] = field(default_factory=dict)
    edges: Dict[str, CodeEdge] = field(default_factory=dict)
    metadata: Dict[str, Any] = field(default_factory=dict)
    statistics: Dict[str, Any] = field(default_factory=dict)

    def __post_init__(self):
        if self.id is None:
            self.id = str(uuid.uuid4())

    def to_dict(self) -> Dict:
        return {
            "id": self.id,
            "name": self.name,
            "version": self.version,
            "created_at": self.created_at,
            "updated_at": self.updated_at,
            "root_path": self.root_path,
            "files": {k: v.to_dict() for k, v in self.files.items()},
            "nodes": {k: v.to_dict() for k, v in self.nodes.items()},
            "edges": {k: v.to_dict() for k, v in self.edges.items()},
            "metadata": self.metadata,
            "statistics": self.statistics,
        }

    @classmethod
    def from_dict(cls, data: Dict) -> "CodeMap":
        return cls(
            id=data.get("id", str(uuid.uuid4())),
            name=data["name"],
            version=data.get("version", "1.0.0"),
            created_at=data.get("created_at", datetime.now().isoformat()),
            updated_at=data.get("updated_at", datetime.now().isoformat()),
            root_path=data.get("root_path", ""),
            files={k: CodeFile.from_dict(v) for k, v in data.get("files", {}).items()},
            nodes={k: CodeNode.from_dict(v) for k, v in data.get("nodes", {}).items()},
            edges={k: CodeEdge.from_dict(v) for k, v in data.get("edges", {}).items()},
            metadata=data.get("metadata", {}),
            statistics=data.get("statistics", {}),
        )

    def add_file(self, file: CodeFile):
        self.files[file.id] = file

    def add_node(self, node: CodeNode, file_id: Optional[str] = None):
        self.nodes[node.id] = node
        if file_id and file_id in self.files:
            self.files[file_id].add_node(node.id)

    def add_edge(self, edge: CodeEdge):
        self.edges[edge.id] = edge

    def add_contains_edge(self, parent_id: str, child_id: str):
        edge = CodeEdge(
            id=str(uuid.uuid4()),
            edge_type="contains",
            source_id=parent_id,
            target_id=child_id,
        )
        self.add_edge(edge)
        if parent_id in self.nodes:
            self.nodes[parent_id].add_child(child_id)

    def add_import_edge(self, importer_id: str, imported_id: str, module: str):
        edge = CodeEdge(
            id=str(uuid.uuid4()),
            edge_type="imports",
            source_id=importer_id,
            target_id=imported_id,
            attributes={"module": module},
        )
        self.add_edge(edge)

    def add_inheritance_edge(self, child_id: str, parent_id: str, is_interface: bool = False):
        edge_type = "implements" if is_interface else "inherits"
        edge = CodeEdge(
            id=str(uuid.uuid4()),
            edge_type=edge_type,
            source_id=child_id,
            target_id=parent_id,
        )
        self.add_edge(edge)

    def add_call_edge(self, caller_id: str, callee_id: str, line_number: Optional[int] = None):
        edge = CodeEdge(
            id=str(uuid.uuid4()),
            edge_type="calls",
            source_id=caller_id,
            target_id=callee_id,
            attributes={"line_number": line_number} if line_number else {},
        )
        self.add_edge(edge)

    def get_node_by_name(self, name: str) -> Optional[CodeNode]:
        for node in self.nodes.values():
            if node.name == name:
                return node
        return None

    def get_nodes_by_type(self, node_type: str) -> List[CodeNode]:
        return [n for n in self.nodes.values() if n.node_type == node_type]

    def get_file_by_path(self, path: str) -> Optional[CodeFile]:
        for file in self.files.values():
            if file.path == path or file.relative_path == path:
                return file
        return None

    def get_edges_from(self, node_id: str) -> List[CodeEdge]:
        return [e for e in self.edges.values() if e.source_id == node_id]

    def get_edges_to(self, node_id: str) -> List[CodeEdge]:
        return [e for e in self.edges.values() if e.target_id == node_id]

    def get_related_nodes(self, node_id: str) -> List[CodeNode]:
        related_ids = {
            edge.target_id for edge in self.edges.values() if edge.source_id == node_id
        } | {
            edge.source_id for edge in self.edges.values() if edge.target_id == node_id
        }
        return [self.nodes[nid] for nid in related_ids if nid in self.nodes]

    def get_module_nodes(self) -> List[CodeNode]:
        return self.get_nodes_by_type("module")

    def get_class_nodes(self) -> List[CodeNode]:
        return self.get_nodes_by_type("class")

    def get_function_nodes(self) -> List[CodeNode]:
        return [n for n in self.nodes.values() if n.node_type in ("function", "method")]

    def update_statistics(self):
        stats = {
            "total_files": len(self.files),
            "total_nodes": len(self.nodes),
            "total_edges": len(self.edges),
            "by_type": {},
            "by_file": {},
        }
        for node_type in NODE_TYPES:
            count = len(self.get_nodes_by_type(node_type))
            if count > 0:
                stats["by_type"][node_type] = count

        for file_id, file in self.files.items():
            stats["by_file"][file_id] = {
                "path": file.relative_path,
                "node_count": len(file.node_ids),
            }

        self.statistics = stats

    def find_path(self, from_id: str, to_id: str) -> Optional[List[str]]:
        from collections import deque

        queue = deque([(from_id, [from_id])])
        visited = {from_id}

        while queue:
            current, path = queue.popleft()
            if current == to_id:
                return path

            for edge in self.edges.values():
                if edge.source_id == current and edge.target_id not in visited:
                    visited.add(edge.target_id)
                    queue.append((edge.target_id, path + [edge.target_id]))
                elif edge.target_id == current and edge.source_id not in visited:
                    visited.add(edge.source_id)
                    queue.append((edge.source_id, path + [edge.source_id]))

        return None

    def get_ancestors(self, node_id: str, max_depth: int = 10) -> List[CodeNode]:
        ancestors = []
        current = node_id
        depth = 0

        while depth < max_depth:
            edges_to_parent = [e for e in self.edges.values() if e.target_id == current and e.is_nesting()]
            if not edges_to_parent:
                break
            parent_id = edges_to_parent[0].source_id
            if parent_id not in self.nodes:
                break
            ancestors.append(self.nodes[parent_id])
            current = parent_id
            depth += 1

        return ancestors

    def get_descendants(self, node_id: str, max_depth: int = 10) -> List[CodeNode]:
        descendants = []
        to_visit = [(node_id, 0)]
        visited = {node_id}

        while to_visit:
            current_id, depth = to_visit.pop(0)
            if depth >= max_depth:
                continue

            for edge in self.edges.values():
                if edge.source_id == current_id and edge.target_id not in visited:
                    if edge.is_nesting() and edge.target_id in self.nodes:
                        descendants.append(self.nodes[edge.target_id])
                        visited.add(edge.target_id)
                        to_visit.append((edge.target_id, depth + 1))

        return descendants
