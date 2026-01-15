"""
TypeScript AST Parser for Code Cartographer

Extracts code structure from TypeScript files using regex-based parsing.
Supports interfaces, types, classes, functions, generics, and relationships.
"""

import os
import re
import uuid
from dataclasses import dataclass, field
from typing import Any, Dict, List, Optional, Tuple

from src.models.code_map import CodeMap, CodeNode, CodeEdge, CodeFile, Position


@dataclass
class TSNode:
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
    type_params: List[str] = field(default_factory=list)
    type_annotation: Optional[str] = None
    extends: List[str] = field(default_factory=list)
    implements: List[str] = field(default_factory=list)


class TypeScriptParser:
    LANGUAGE_NAME = "typescript"
    FILE_EXTENSIONS = [".ts", ".tsx"]

    RESERVED_NAMES = frozenset({'if', 'else', 'for', 'while', 'switch', 'catch', 'return', 'import', 'export'})

    def __init__(self):
        self._current_file_id = ""
        self._current_file_path = ""
        self._current_content = ""
        self._nodes: List[TSNode] = []
        self._edges: List[Dict] = []
        self._scope: List[str] = []

    def parse(self, file_id: str, content: str) -> Tuple[List[CodeNode], List[CodeEdge]]:
        self._current_file_id = file_id
        self._current_file_path = ""
        self._current_content = content
        self._nodes = []
        self._edges = []
        self._scope = []

        lines = content.split("\n")
        self._extract_nodes(content, lines)
        self._extract_edges()

        return self._to_code_map_objects()

    def _extract_nodes(self, content: str, lines: List[str]):
        interface_pat = r'interface\s+(\w+)(?:\s*<\w+>)?\s*(?:extends\s+([^{]+))?\{'
        type_pat = r'type\s+(\w+)\s*=\s*'
        class_pat = r'class\s+(\w+)(?:\s*<\w+>)?\s*(?:extends\s+(\w+))?(?:\s+implements\s+([^{]+))?\{'
        func_pat = r'(?:export\s+)?(?:async\s+)?function\s+(\w+)\s*(?:<\w+>)?\s*\([^)]*\)'
        arrow_pat = r'(?:const|let|var)\s+(\w+)\s*(?:<\w+>)?\s*=\s*(?:async\s+)?\([^)]*\)\s*(?::\s*[\w\[\]<>,\s|]+)?\s*=>'

        current_file_id = self._current_file_id

        for i, line in enumerate(lines, 1):
            if interface_match := re.search(interface_pat, line):
                node = self._create_interface_node(interface_match, current_file_id, i)
                self._nodes.append(node)
                continue

            if type_match := re.search(type_pat, line):
                node = self._create_type_node(type_match, current_file_id, i)
                self._nodes.append(node)
                continue

            if class_match := re.search(class_pat, line):
                node = self._create_class_node(class_match, current_file_id, i)
                self._nodes.append(node)
                self._scope.append(node.id)
                continue

            if func_match := re.search(func_pat, line):
                if 'class ' not in line and 'interface ' not in line:
                    node = self._create_function_node(func_match, current_file_id, i)
                    if node:
                        self._nodes.append(node)
                        continue

            if arrow_match := re.search(arrow_pat, line):
                name = arrow_match.group(1)
                if not name[0].isupper() and name not in self.RESERVED_NAMES:
                    node = self._create_arrow_function_node(name, current_file_id, i)
                    if node:
                        self._nodes.append(node)
                        continue

        self._extract_ts_methods(content, lines)

    def _extract_ts_methods(self, content: str, lines: List[str]):
        in_class = False
        in_interface = False
        class_indent = 0

        for i, line in enumerate(lines, 1):
            if ('class ' in line or 'interface ' in line) and '{' in line:
                if 'class ' in line:
                    in_class = True
                else:
                    in_interface = True
                class_indent = len(line) - len(line.lstrip())
                continue

            if in_class or in_interface:
                current_indent = len(line) - len(line.lstrip())
                if current_indent <= class_indent and line.strip():
                    in_class = False
                    in_interface = False
                    continue

                method_match = re.search(r'(\w+)\s*(?:<\w+>)?\s*\([^)]*\)\s*(?::\s*[\w\[\]<>,\s|]+)?\s*\{', line)
                if method_match:
                    method_name = method_match.group(1)
                    if method_name and not method_name.startswith('_') and method_name[0].islower():
                        node = self._create_method_node(method_name, self._current_file_id, i)
                        self._nodes.append(node)

    def _create_interface_node(self, match, file_id: str, line: int) -> TSNode:
        name = match.group(1)
        extends = []
        if match.lastindex and match.lastindex >= 2 and match.group(2):
            extends = [e.strip() for e in match.group(2).split(',')]

        return TSNode(
            id=self._generate_id("interface", name),
            node_type="interface",
            name=name,
            qualified_name=name,
            file_id=file_id,
            start_line=line,
            extends=extends,
        )

    def _create_type_node(self, match, file_id: str, line: int) -> TSNode:
        name = match.group(1)

        type_annotation = None
        type_match = re.search(r'type\s+\w+\s*=\s*(.+?);', self._current_content)
        if type_match:
            type_annotation = type_match.group(1).strip()[:100]

        return TSNode(
            id=self._generate_id("type", name),
            node_type="type_alias",
            name=name,
            qualified_name=name,
            file_id=file_id,
            start_line=line,
            type_annotation=type_annotation,
        )

    def _create_class_node(self, match, file_id: str, line: int) -> TSNode:
        name = match.group(1)
        extends = []
        implements = []

        if match.lastindex and match.lastindex >= 2:
            base = match.group(2)
            if base:
                extends.append(base)

        if match.lastindex and match.lastindex >= 3:
            impl = match.group(3)
            if impl:
                implements = [e.strip() for e in impl.split(',')]

        type_params = []
        tp_match = re.search(r'<(\w+)>', str(line))
        if tp_match:
            type_params.append(tp_match.group(1))

        return TSNode(
            id=self._generate_id("class", name),
            node_type="class",
            name=name,
            qualified_name=name,
            file_id=file_id,
            start_line=line,
            type_params=type_params,
            extends=extends,
            implements=implements,
        )

    def _create_function_node(self, match, file_id: str, line: int) -> Optional[TSNode]:
        name = match.group(1)

        if name in self.RESERVED_NAMES:
            return None

        type_params = []
        tp_match = re.search(r'<(\w+)>', str(line))
        if tp_match:
            type_params.append(tp_match.group(1))

        return TSNode(
            id=self._generate_id("function", name),
            node_type="function",
            name=name,
            qualified_name=name,
            file_id=file_id,
            start_line=line,
            type_params=type_params,
        )

    def _create_arrow_function_node(self, name: str, file_id: str, line: int) -> Optional[TSNode]:
        if name in self.RESERVED_NAMES:
            return None
        if name and name[0].isupper():
            return None

        return TSNode(
            id=self._generate_id("function", name),
            node_type="function",
            name=name,
            qualified_name=name,
            file_id=file_id,
            start_line=line,
        )

    def _create_method_node(self, name: str, file_id: str, line: int) -> TSNode:
        type_params = []
        tp_match = re.search(r'<(\w+)>', str(line))
        if tp_match:
            type_params.append(tp_match.group(1))

        return TSNode(
            id=self._generate_id("method", name),
            node_type="method",
            name=name,
            qualified_name=name,
            file_id=file_id,
            start_line=line,
            parent_id=self._scope[-1] if self._scope else None,
            type_params=type_params,
        )

    def _extract_edges(self):
        for node in self._nodes:
            if node.node_type == "class":
                for base in node.extends:
                    self._edges.append({
                        "edge_type": "extends",
                        "source_id": node.id,
                        "target_id": f"class_{base}_",
                    })
                for impl in node.implements:
                    self._edges.append({
                        "edge_type": "implements",
                        "source_id": node.id,
                        "target_id": f"interface_{impl}_",
                    })

            if node.node_type in ("class", "interface"):
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
                annotations=node.type_params,
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


def parse_typescript_code(file_path: str, content: str) -> Tuple[CodeFile, List[CodeNode], List[CodeEdge]]:
    parser = TypeScriptParser()

    rel_path = os.path.basename(file_path)
    file_id = str(uuid.uuid4())[:16]

    file = CodeFile(
        id=file_id,
        path=file_path,
        relative_path=rel_path,
        language="typescript",
        size_bytes=len(content.encode("utf-8")),
        line_count=len(content.split("\n")),
    )

    nodes, edges = parser.parse(file_id, content)

    return file, nodes, edges
