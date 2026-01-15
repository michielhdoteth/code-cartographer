"""
JavaScript AST Parser for Code Cartographer

Extracts code structure from JavaScript files using regex-based parsing.
Supports functions, classes, methods, imports, and relationships.
"""

import os
import re
import uuid
from dataclasses import dataclass, field
from typing import Any, Dict, List, Optional, Tuple

from src.models.code_map import CodeMap, CodeNode, CodeEdge, CodeFile, Position


@dataclass
class JSNode:
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


class JavaScriptParser:
    LANGUAGE_NAME = "javascript"
    FILE_EXTENSIONS = [".js", ".jsx"]

    RESERVED_NAMES = frozenset({'if', 'else', 'for', 'while', 'switch', 'catch', 'return'})

    def __init__(self):
        self._current_file_id = ""
        self._current_file_path = ""
        self._nodes: List[JSNode] = []
        self._edges: List[Dict] = []
        self._scope: List[str] = []

    def parse(self, file_id: str, content: str) -> Tuple[List[CodeNode], List[CodeEdge]]:
        self._current_file_id = file_id
        self._current_file_path = ""
        self._nodes = []
        self._edges = []
        self._scope = []

        lines = content.split("\n")
        self._extract_nodes(content, lines)
        self._extract_edges()

        return self._to_code_map_objects()

    def _extract_nodes(self, content: str, lines: List[str]):
        class_pat = r'class\s+(\w+)\s*(?:extends\s+(\w+))?'
        func_pat = r'(?:function\s+)?(\w+)\s*\([^)]*\)\s*\{'
        arrow_pat = r'(\w+)\s*=\s*\([^)]*\)\s*=>'
        const_arrow_pat = r'const\s+(\w+)\s*=\s*\([^)]*\)\s*=>'

        current_file_id = self._current_file_id

        for i, line in enumerate(lines, 1):
            class_match = re.search(class_pat, line)
            if class_match:
                node = self._create_class_node(class_match, current_file_id, i)
                self._nodes.append(node)
                self._scope.append(node.id)
                continue

            func_match = re.search(func_pat, line)
            if func_match and 'class ' not in line and not line.strip().startswith('//'):
                node = self._create_function_node(func_match, current_file_id, i, is_arrow=False)
                if node:
                    self._nodes.append(node)
                    self._scope.append(node.id)
                    continue

            arrow_match = re.search(arrow_pat, line)
            const_match = re.search(const_arrow_pat, line)
            if (arrow_match or const_match) and 'class ' not in line:
                name = None
                if arrow_match:
                    name = arrow_match.group(1)
                elif const_match:
                    name = const_match.group(1)
                if name and not name[0].isupper():
                    node = self._create_arrow_function_node(name, current_file_id, i)
                    if node:
                        self._nodes.append(node)
                        continue

        self._extract_class_methods(content, lines)

    def _extract_class_methods(self, content: str, lines: List[str]):
        in_class = False
        class_indent = 0

        for i, line in enumerate(lines, 1):
            if 'class ' in line and '{' in line:
                in_class = True
                class_indent = len(line) - len(line.lstrip())
                continue

            if in_class:
                current_indent = len(line) - len(line.lstrip())
                if current_indent <= class_indent and line.strip():
                    in_class = False
                    continue

                method_match = re.search(r'(\w+)\s*\([^)]*\)\s*\{', line)
                if method_match:
                    method_name = method_match.group(1)
                    if not method_name.startswith('_'):
                        node = self._create_method_node(method_name, self._current_file_id, i)
                        self._nodes.append(node)

    def _create_class_node(self, match, file_id: str, line: int) -> JSNode:
        name = match.group(1)
        base_class = match.group(2) if match.lastindex and match.lastindex >= 2 else None

        return JSNode(
            id=self._generate_id("class", name),
            node_type="class",
            name=name,
            qualified_name=name,
            file_id=file_id,
            start_line=line,
        )

    def _create_function_node(self, match, file_id: str, line: int, is_arrow: bool = False) -> Optional[JSNode]:
        name = match.group(1)

        if name in self.RESERVED_NAMES:
            return None
        if name and name[0].isupper():
            return None

        return JSNode(
            id=self._generate_id("function", name),
            node_type="function",
            name=name,
            qualified_name=name,
            file_id=file_id,
            start_line=line,
        )

    def _create_arrow_function_node(self, name: str, file_id: str, line: int) -> Optional[JSNode]:
        if name in self.RESERVED_NAMES:
            return None
        if name and name[0].isupper():
            return None

        return JSNode(
            id=self._generate_id("function", name),
            node_type="function",
            name=name,
            qualified_name=name,
            file_id=file_id,
            start_line=line,
        )

    def _create_method_node(self, name: str, file_id: str, line: int) -> JSNode:
        return JSNode(
            id=self._generate_id("method", name),
            node_type="method",
            name=name,
            qualified_name=name,
            file_id=file_id,
            start_line=line,
            parent_id=self._scope[-1] if self._scope else None,
        )

    def _extract_edges(self):
        for node in self._nodes:
            if node.node_type == "class":
                self._edges.append({
                    "edge_type": "contains",
                    "source_id": node.id,
                    "target_id": node.id,
                })

    def _to_code_map_objects(self) -> Tuple[List[CodeNode], List[CodeEdge]]:
        code_nodes = []
        code_edges = []

        for node in self._nodes:
            position = Position(line=node.start_line, column=0, end_line=node.end_line or node.start_line, end_column=0)

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
        return f"{node_type}_{name}_{uuid.uuid4().hex[:8]}"


def parse_javascript_code(file_path: str, content: str) -> Tuple[CodeFile, List[CodeNode], List[CodeEdge]]:
    parser = JavaScriptParser()

    rel_path = os.path.basename(file_path)
    file_id = str(uuid.uuid4())[:16]

    file = CodeFile(
        id=file_id,
        path=file_path,
        relative_path=rel_path,
        language="javascript",
        size_bytes=len(content.encode("utf-8")),
        line_count=len(content.split("\n")),
    )

    nodes, edges = parser.parse(file_id, content)

    return file, nodes, edges
