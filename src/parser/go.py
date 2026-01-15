"""
Go Parser for Code Cartographer

Extracts code structure from Go files using regex-based parsing.
Supports functions, methods, structs, interfaces, and relationships.
"""

import os
import re
import uuid
from dataclasses import dataclass, field
from typing import Any, Dict, List, Optional, Tuple

from src.models.code_map import CodeMap, CodeNode, CodeEdge, CodeFile, Position


@dataclass
class GoNode:
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
    interfaces: List[str] = field(default_factory=list)
    is_method: bool = False
    is_pointer: bool = False
    is_struct: bool = False


class GoParser:
    LANGUAGE_NAME = "go"
    FILE_EXTENSIONS = [".go"]

    RESERVED_KEYWORDS = frozenset({
        "break", "case", "chan", "const", "continue", "default", "defer", "else",
        "fallthrough", "for", "func", "go", "goto", "if", "import", "interface",
        "map", "package", "range", "return", "select", "struct", "switch", "type",
        "var", "true", "false", "nil", "iota", "append", "cap", "close", "complex",
        "copy", "delete", "imag", "len", "make", "new", "panic", "print", "println",
        "real", "recover",
    })

    FRAMEWORK_MARKERS = {
        "gin": ["gin-gonic/gin", "gin.Context", "gin.Engine"],
        "echo": ["echo-labs/echo", "echo.Context", "echo.Echo"],
        "fiber": ["gofiber/fiber", "fiber.Ctx", "fiber.App"],
        "gorilla": ["gorilla/mux", "mux.Router"],
        "beego": ["beego/beego", "beego.Controller"],
        "chi": ["go-chi/chi", "chi.Router"],
        "gorm": ["gorm.io/gorm", "gorm.DB"],
        "sqlx": ["jmoiron/sqlx", "sqlx.DB"],
        "grpc": ["google.golang.org/grpc", "grpc.Server"],
    }

    def __init__(self):
        self._current_file_id = ""
        self._current_file_path = ""
        self._nodes: List[GoNode] = []
        self._edges: List[Dict] = []
        self._scope: List[str] = []
        self._package_name: str = ""
        self._framework_detected: List[str] = []

    def parse(self, file_id: str, content: str) -> Tuple[List[CodeNode], List[CodeEdge]]:
        self._current_file_id = file_id
        self._current_file_path = ""
        self._nodes = []
        self._edges = []
        self._scope = []
        self._package_name = ""
        self._framework_detected = []

        lines = content.split("\n")
        self._extract_package_and_imports(content)
        self._extract_nodes(content, lines)
        self._extract_edges()

        return self._to_code_map_objects()

    def _extract_package_and_imports(self, content: str):
        package_match = re.search(r'^package\s+(\w+)', content, re.MULTILINE)
        if package_match:
            self._package_name = package_match.group(1)

        for import_match in re.finditer(r'^import\s*\(([^)]*)\)', content, re.MULTILINE | re.DOTALL):
            import_block = import_match.group(1)
            for single_import in re.finditer(r'(?:"([^"]+)"|`([^`]+)`)', import_block):
                import_path = single_import.group(1) or single_import.group(2)

                for framework, markers in self.FRAMEWORK_MARKERS.items():
                    for marker in markers:
                        if marker in import_path:
                            if framework not in self._framework_detected:
                                self._framework_detected.append(framework)

    def _extract_nodes(self, content: str, lines: List[str]):
        struct_pattern = r'type\s+(\w+)\s+struct'
        interface_pattern = r'type\s+(\w+)\s+interface'
        func_pattern = r'^func\s+(?:\([^)]*\)\s+)?(\w+)\s*\('
        top_level_func_pattern = r'^func\s+([A-Z]\w*)\s*\('
        method_pattern = r'^func\s+\((?:\*\s*)?(\w+)\)\s+(\w+)\s*\('

        current_file_id = self._current_file_id

        for i, line in enumerate(lines, 1):
            docstring = lines[i-2].strip()[2:].strip() if i > 1 and lines[i-2].strip().startswith("//") else ""
            docstring = docstring if docstring else ""

            if struct_match := re.search(struct_pattern, line):
                name = struct_match.group(1)
                if name not in self.RESERVED_KEYWORDS:
                    node = self._create_struct_node(name, current_file_id, i, docstring)
                    self._nodes.append(node)
                    self._scope.append(node.id)
                    continue

            if interface_match := re.search(interface_pattern, line):
                name = interface_match.group(1)
                if name not in self.RESERVED_KEYWORDS:
                    node = self._create_interface_node(name, current_file_id, i, docstring)
                    self._nodes.append(node)
                    continue

            if method_match := re.search(method_pattern, line):
                receiver, method_name = method_match.groups()
                if method_name not in self.RESERVED_KEYWORDS:
                    node = self._create_method_node(receiver, method_name, current_file_id, i, line, docstring)
                    if node:
                        self._nodes.append(node)
                        continue

            if top_func_match := re.search(top_level_func_pattern, line):
                name = top_func_match.group(1)
                if name not in self.RESERVED_KEYWORDS and name[0].isupper():
                    node = self._create_func_node(name, current_file_id, i, line, docstring, is_exported=True)
                    if node:
                        self._nodes.append(node)
                        continue

            if func_match := re.search(func_pattern, line):
                if not ('type ' in line or 'struct' in line or 'interface' in line):
                    name = func_match.group(1)
                    if name not in self.RESERVED_KEYWORDS:
                        node = self._create_func_node(name, current_file_id, i, line, docstring, is_exported=False)
                        if node:
                            self._nodes.append(node)
                            continue

    def _create_struct_node(self, name: str, file_id: str, line: int, docstring: str) -> GoNode:
        interfaces = self._extract_interfaces(content, line) if False else []

        return GoNode(
            id=self._generate_id("struct", name),
            node_type="struct",
            name=name,
            qualified_name=f"{self._package_name}.{name}",
            file_id=file_id,
            start_line=line,
            docstring=docstring if docstring else None,
            interfaces=interfaces,
            is_struct=True,
        )

    def _create_interface_node(self, name: str, file_id: str, line: int, docstring: str) -> GoNode:
        return GoNode(
            id=self._generate_id("interface", name),
            node_type="interface",
            name=name,
            qualified_name=f"{self._package_name}.{name}",
            file_id=file_id,
            start_line=line,
            docstring=docstring if docstring else None,
        )

    def _create_method_node(self, receiver: str, name: str, file_id: str, line: int, line_content: str, docstring: str) -> Optional[GoNode]:
        is_pointer = "*" in line_content[:line_content.find("(")]

        return_type_match = re.search(r'\)\s*([\w\[\]*\s<>]+)(?:\s*\{)?$', line_content)
        return_type = return_type_match.group(1).strip() if return_type_match else None

        return GoNode(
            id=self._generate_id("method", f"{receiver}_{name}"),
            node_type="method",
            name=name,
            qualified_name=f"{self._package_name}.{receiver}.{name}",
            file_id=file_id,
            start_line=line,
            parent_id=f"struct_{receiver}",
            docstring=docstring if docstring else None,
            return_type=return_type,
            is_method=True,
            is_pointer=is_pointer,
        )

    def _create_func_node(self, name: str, file_id: str, line: int, line_content: str, docstring: str, is_exported: bool) -> Optional[GoNode]:
        if name.startswith("_"):
            return None

        return_type_match = re.search(r'\)\s*([\w\[\]*\s<>]+)(?:\s*\{)?$', line_content)
        return_type = return_type_match.group(1).strip() if return_type_match else None

        return GoNode(
            id=self._generate_id("function", name),
            node_type="function",
            name=name,
            qualified_name=f"{self._package_name}.{name}",
            file_id=file_id,
            start_line=line,
            docstring=docstring if docstring else None,
            return_type=return_type,
        )

    def _extract_interfaces(self, content: str, line_num: int) -> List[str]:
        return []

    def _extract_edges(self):
        for node in self._nodes:
            if node.node_type in ("struct", "interface"):
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
                annotations=node.interfaces,
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

    def get_frameworks(self) -> List[str]:
        return self._framework_detected


def parse_go_code(file_path: str, content: str) -> Tuple[CodeFile, List[CodeNode], List[CodeEdge]]:
    parser = GoParser()

    rel_path = os.path.basename(file_path)
    file_id = str(uuid.uuid4())[:16]

    file = CodeFile(
        id=file_id,
        path=file_path,
        relative_path=rel_path,
        language="go",
        size_bytes=len(content.encode("utf-8")),
        line_count=len(content.split("\n")),
    )

    nodes, edges = parser.parse(file_id, content)

    return file, nodes, edges
