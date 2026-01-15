"""
Rust Parser for Code Cartographer

Extracts code structure from Rust files using regex-based parsing.
Supports structs, functions, impl blocks, traits, and relationships.
"""

import os
import re
import uuid
from dataclasses import dataclass, field
from typing import Any, Dict, List, Optional, Tuple

from src.models.code_map import CodeMap, CodeNode, CodeEdge, CodeFile, Position


@dataclass
class RustNode:
    id: str
    node_type: str
    name: str
    qualified_name: str
    file_id: str
    start_line: int
    end_line: Optional[int] = None
    parent_id: Optional[str] = None
    children_ids: List[str] = field(default_factory=list)
    docstring: Optional[str] = None
    parameters: List[Dict] = field(default_factory=list)
    return_type: Optional[str] = None
    traits: List[str] = field(default_factory=list)
    visibility: str = "pub"
    is_async: bool = False
    is_const: bool = False
    is_static: bool = False


class RustParser:
    LANGUAGE_NAME = "rust"
    FILE_EXTENSIONS = [".rs"]

    RESERVED_KEYWORDS = frozenset({
        "as", "break", "const", "continue", "crate", "else", "enum", "extern",
        "false", "fn", "for", "if", "impl", "in", "let", "loop", "match", "mod",
        "move", "mut", "pub", "ref", "return", "self", "Self", "static", "struct",
        "super", "trait", "true", "type", "unsafe", "use", "where", "while",
        "async", "await", "dyn", "abstract", "become", "box", "do", "final",
        "macro", "override", "priv", "typeof", "unsized", "virtual", "union",
    })

    CRATE_MARKERS = ["Cargo.toml", "src/lib.rs", "src/main.rs"]

    FRAMEWORK_MARKERS = {
        "tokio": ["tokio::", "async_tokio"],
        "actix": ["actix::", "actix_web"],
        "serde": ["serde::", "Serialize", "Deserialize"],
        "warp": ["warp::"],
        "rocket": ["rocket::"],
        "axum": ["axum::"],
        "diesel": ["diesel::"],
        "sqlx": ["sqlx::"],
    }

    def __init__(self):
        self._current_file_id = ""
        self._current_file_path = ""
        self._nodes: List[RustNode] = []
        self._edges: List[Dict] = []
        self._scope: List[str] = []
        self._current_mod: str = ""
        self._framework_detected: List[str] = []

    def parse(self, file_id: str, content: str) -> Tuple[List[CodeNode], List[CodeEdge]]:
        self._current_file_id = file_id
        self._current_file_path = ""
        self._nodes = []
        self._edges = []
        self._scope = []
        self._current_mod = ""
        self._framework_detected = []

        lines = content.split("\n")
        self._extract_mod_and_use(content)
        self._extract_nodes(content, lines)
        self._extract_edges()

        return self._to_code_map_objects()

    def _extract_mod_and_use(self, content: str):
        mod_match = re.search(r'^mod\s+(\w+)', content, re.MULTILINE)
        if mod_match:
            self._current_mod = mod_match.group(1)

        for use_match in re.finditer(r'^use\s+([\w:;]+(?:\{[^}]*\})?)', content, re.MULTILINE):
            use_path = use_match.group(1)

            for framework, markers in self.FRAMEWORK_MARKERS.items():
                for marker in markers:
                    if marker in use_path:
                        if framework not in self._framework_detected:
                            self._framework_detected.append(framework)

    def _extract_nodes(self, content: str, lines: List[str]):
        struct_pattern = r'(?:pub\s+)?(?:(?:pub\([^)]*\)\s+)?struct)\s+(\w+)(?:\s*<[^>]*>)?(?:\s*\(.*?\))?(?:\s*where\s*[^;]+)?'
        enum_pattern = r'(?:pub\s+)?enum\s+(\w+)(?:\s*<[^>]*>)?(?:\s*where\s*[^;]+)?'
        trait_pattern = r'(?:pub\s+)?trait\s+(\w+)(?:\s*<[^>]*>)?(?:\s+where\s*[^;]+)?'
        impl_pattern = r'(?:pub\s+)?impl(?:\s*<[^>]*>)?\s+(?:[\w:]+)(?:\s*for\s+[\w:]+)?(?:\s*where\s*[^;]+)?\s*\{'
        fn_pattern = r'(?:pub\s+)?(?:pub\([^)]*\)\s+)?(?:async\s+)?(?:const\s+)?(?:unsafe\s+)?(?:extern\s+(?:"[^"]*"\s+)?)?fn\s+(\w+)'
        impl_fn_pattern = r'(?:pub\s+)?(?:pub\([^)]*\)\s+)?(?:async\s+)?(?:const\s+)?(?:unsafe\s+)?fn\s+(\w+)'

        current_file_id = self._current_file_id

        for i, line in enumerate(lines, 1):
            doc_comment = ""
            if i > 1 and lines[i-2].strip().startswith("///"):
                doc_comment = lines[i-2].strip()[3:].strip()

            struct_match = re.search(struct_pattern, line)
            if struct_match:
                name = struct_match.group(1)
                if name not in self.RESERVED_KEYWORDS:
                    node = self._create_struct_node(name, current_file_id, i, doc_comment)
                    self._nodes.append(node)
                    self._scope.append(node.id)
                    continue

            enum_match = re.search(enum_pattern, line)
            if enum_match:
                name = enum_match.group(1)
                if name not in self.RESERVED_KEYWORDS:
                    node = self._create_enum_node(name, current_file_id, i, doc_comment)
                    self._nodes.append(node)
                    self._scope.append(node.id)
                    continue

            trait_match = re.search(trait_pattern, line)
            if trait_match:
                name = trait_match.group(1)
                if name not in self.RESERVED_KEYWORDS:
                    node = self._create_trait_node(name, current_file_id, i, doc_comment)
                    self._nodes.append(node)
                    continue

            fn_match = re.search(fn_pattern, line)
            if fn_match and 'struct ' not in line and 'impl ' not in line and 'trait ' not in line:
                name = fn_match.group(1)
                if name not in self.RESERVED_KEYWORDS:
                    node = self._create_fn_node(name, current_file_id, i, line, doc_comment, is_toplevel=True)
                    if node:
                        self._nodes.append(node)
                        continue

            impl_match = re.search(impl_pattern, line)
            if impl_match:
                impl_target = line[line.find("impl")+4:line.find("{")].strip()
                node = self._create_impl_block(impl_target, current_file_id, i)
                self._nodes.append(node)
                continue

        self._extract_impl_methods(content, lines)

    def _extract_impl_methods(self, content: str, lines: List[str]):
        in_impl = False
        impl_indent = 0
        impl_target = ""

        fn_pattern = r'(?:pub\s+)?(?:pub\([^)]*\)\s+)?(?:async\s+)?fn\s+(\w+)'

        for i, line in enumerate(lines, 1):
            impl_match = re.search(r'(?:pub\s+)?impl(?:\s*<[^>]*>)?\s+([\w:]+)', line)
            if impl_match:
                in_impl = True
                impl_target = impl_match.group(1)
                impl_indent = len(line) - len(line.lstrip())
                continue

            if in_impl:
                current_indent = len(line) - len(line.lstrip())
                if line.strip() and current_indent <= impl_indent:
                    in_impl = False
                    continue

                fn_match = re.search(fn_pattern, line)
                if fn_match:
                    name = fn_match.group(1)
                    if name not in self.RESERVED_KEYWORDS:
                        node = self._create_fn_node(name, self._current_file_id, i, line, "", is_toplevel=False)
                        if node:
                            node.parent_id = f"impl_{impl_target}"
                            self._nodes.append(node)

    def _create_struct_node(self, name: str, file_id: str, line: int, docstring: str) -> RustNode:
        visibility = "pub" if "pub " in self._get_line_before(line) else "private"

        return RustNode(
            id=self._generate_id("struct", name),
            node_type="struct",
            name=name,
            qualified_name=name,
            file_id=file_id,
            start_line=line,
            docstring=docstring if docstring else None,
            visibility=visibility,
        )

    def _create_enum_node(self, name: str, file_id: str, line: int, docstring: str) -> RustNode:
        visibility = "pub" if "pub " in self._get_line_before(line) else "private"

        return RustNode(
            id=self._generate_id("enum", name),
            node_type="enum",
            name=name,
            qualified_name=name,
            file_id=file_id,
            start_line=line,
            docstring=docstring if docstring else None,
            visibility=visibility,
        )

    def _create_trait_node(self, name: str, file_id: str, line: int, docstring: str) -> RustNode:
        visibility = "pub" if "pub " in self._get_line_before(line) else "private"

        return RustNode(
            id=self._generate_id("trait", name),
            node_type="trait",
            name=name,
            qualified_name=name,
            file_id=file_id,
            start_line=line,
            docstring=docstring if docstring else None,
            visibility=visibility,
        )

    def _create_fn_node(self, name: str, file_id: str, line: int, line_content: str, docstring: str, is_toplevel: bool) -> Optional[RustNode]:
        if name.startswith("_") and not name.startswith("__"):
            return None

        visibility = "pub" if "pub " in line_content or "pub(" in line_content else "private"

        return_type_match = re.search(r'->\s*([\w<>,\s\[\]]+)', line_content)
        return_type = return_type_match.group(1).strip() if return_type_match else None

        is_async = "async " in line_content
        is_const = "const " in line_content
        is_static = "static " in line_content

        return RustNode(
            id=self._generate_id("function", name),
            node_type="function",
            name=name,
            qualified_name=name,
            file_id=file_id,
            start_line=line,
            parent_id=None if is_toplevel else self._scope[-1] if self._scope else None,
            docstring=docstring if docstring else None,
            return_type=return_type,
            visibility=visibility,
            is_async=is_async,
            is_const=is_const,
            is_static=is_static,
        )

    def _create_impl_block(self, target: str, file_id: str, line: int) -> RustNode:
        return RustNode(
            id=self._generate_id("impl", f"impl_{target}"),
            node_type="impl",
            name=f"impl {target}",
            qualified_name=f"impl_{target}",
            file_id=file_id,
            start_line=line,
        )

    def _get_line_before(self, line_num: int) -> str:
        return ""

    def _extract_edges(self):
        for node in self._nodes:
            if node.node_type in ("struct", "enum"):
                self._edges.append({
                    "edge_type": "contains",
                    "source_id": node.id,
                    "target_id": node.id,
                })

    def _to_code_map_objects(self) -> Tuple[List[CodeNode], List[CodeEdge]]:
        code_nodes = []
        code_edges = []

        for node in self._nodes:
            position = Position(
                line=node.start_line,
                column=0,
                end_line=node.end_line or node.start_line,
                end_column=0,
            )

            code_node = CodeNode(
                id=node.id,
                node_type=node.node_type,
                name=node.name,
                qualified_name=node.qualified_name,
                file_id=node.file_id,
                start_position=position,
                end_position=position,
                parent_id=node.parent_id,
                children_ids=node.children_ids,
                docstring=node.docstring,
                parameters=node.parameters,
                return_type=node.return_type,
                annotations=node.traits,
            )
            code_nodes.append(code_node)

        for edge in self._edges:
            code_edge = CodeEdge(
                id=str(uuid.uuid4()),
                edge_type=edge["edge_type"],
                source_id=edge["source_id"],
                target_id=edge["target_id"],
            )
            code_edges.append(code_edge)

        return code_nodes, code_edges

    def _generate_id(self, node_type: str, name: str) -> str:
        clean_name = re.sub(r'[^a-zA-Z0-9_]', '', name)
        return f"{node_type}_{clean_name}_{uuid.uuid4().hex[:8]}"

    def get_frameworks(self) -> List[str]:
        return self._framework_detected


def parse_rust_code(file_path: str, content: str) -> Tuple[CodeFile, List[CodeNode], List[CodeEdge]]:
    parser = RustParser()

    rel_path = os.path.basename(file_path)
    file_id = str(uuid.uuid4())[:16]

    file = CodeFile(
        id=file_id,
        path=file_path,
        relative_path=rel_path,
        language="rust",
        size_bytes=len(content.encode("utf-8")),
        line_count=len(content.split("\n")),
    )

    nodes, edges = parser.parse(file_id, content)

    return file, nodes, edges
