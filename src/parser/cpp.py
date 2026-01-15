"""
C/C++ Parser for Code Cartographer

Extracts code structure from C/C++ files using regex-based parsing.
Supports classes, structs, functions, methods, templates, and relationships.
"""

import os
import re
import uuid
from dataclasses import dataclass, field
from typing import Any, Dict, List, Optional, Tuple

from src.models.code_map import CodeMap, CodeNode, CodeEdge, CodeFile, Position


@dataclass
class CppNode:
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
    base_classes: List[str] = field(default_factory=list)
    templates: List[str] = field(default_factory=list)
    visibility: str = "private"
    is_virtual: bool = False
    is_static: bool = False
    is_const: bool = False
    is_extern: bool = False
    is_inline: bool = False
    is_template: bool = False


class CppParser:
    LANGUAGE_NAME = "cpp"
    FILE_EXTENSIONS = [".cpp", ".cc", ".cxx", ".h", ".hpp", ".hxx", ".c", ".c++"]

    RESERVED_KEYWORDS = frozenset({
        "alignas", "alignof", "and", "and_eq", "asm", "auto", "bitand", "bitor",
        "bool", "break", "case", "catch", "char", "class", "compl", "const",
        "constexpr", "const_cast", "continue", "decltype", "default", "delete",
        "do", "double", "dynamic_cast", "else", "enum", "explicit", "export",
        "extern", "false", "float", "for", "friend", "goto", "if", "inline",
        "int", "long", "mutable", "namespace", "new", "noexcept", "not",
        "not_eq", "operator", "or", "or_eq", "private", "protected", "public",
        "register", "reinterpret_cast", "return", "short", "signed", "sizeof",
        "static", "static_assert", "static_cast", "struct", "switch", "template",
        "this", "thread_local", "throw", "true", "try", "typedef", "typeid",
        "typename", "union", "unsigned", "using", "virtual", "void", "volatile",
        "wchar_t", "while", "xor", "xor_eq", "override", "final",
    })

    FRAMEWORK_MARKERS = {
        "qt": ["Q_OBJECT", "QCoreApplication", "QWidget", "qlayout"],
        "boost": ["boost::", "#include <boost/"],
        "gtk": ["GtkWidget", "gtk_", "#include <gtk"],
        "opencv": ["cv::", "cv::Mat", "#include <opencv2"],
        "eigen": ["Eigen::", "#include <Eigen"],
        "fmt": ["fmt::", "format.h", "#include <fmt"],
        "spdlog": ["spdlog::", "#include <spdlog"],
        "nlohmann": ["nlohmann::json", "#include <nlohmann"],
        "grpc": ["grpc::", "#include <grpc"],
        "protobuf": ["google::protobuf", "#include <google/protobuf"],
    }

    def __init__(self):
        self._current_file_id = ""
        self._current_file_path = ""
        self._nodes: List[CppNode] = []
        self._edges: List[Dict] = []
        self._scope: List[str] = []
        self._framework_detected: List[str] = []

    def parse(self, file_id: str, content: str) -> Tuple[List[CodeNode], List[CodeEdge]]:
        self._current_file_id = file_id
        self._current_file_path = ""
        self._nodes = []
        self._edges = []
        self._scope = []
        self._framework_detected = []

        lines = content.split("\n")
        self._extract_includes(content)
        self._extract_nodes(content, lines)
        self._extract_edges()

        return self._to_code_map_objects()

    def _extract_includes(self, content: str):
        for include_match in re.finditer(r'#include\s*(<[^>]+>|"[^"]+")', content):
            include_path = include_match.group(1).strip("<>\"")
            self._includes.append(include_path)

            for framework, markers in self.FRAMEWORK_MARKERS.items():
                for marker in markers:
                    if marker in include_path:
                        if framework not in self._framework_detected:
                            self._framework_detected.append(framework)

    def _extract_nodes(self, content: str, lines: List[str]):
        class_pattern = r'(?:export\s+)?(?:class|struct)\s+(\w+)(?:\s*:\s*public\s+([\w,\s]+))?(?:\s*:\s*private\s+([\w,\s]+))?'
        template_class_pattern = r'template\s*<[^>]*>\s*(?:class|struct)\s+(\w+)'
        function_pattern = r'(?:export\s+)?(?:inline\s+)?(?:virtual\s+)?(?:static\s+)?(?:extern\s+)?(?:void|int|long|double|float|char|bool|auto|const|auto|[\w:<>\[\]*&]+)\s+(?:\*\s*)?(\w+)\s*\('
        method_pattern = r'(?:public|private|protected):\s*(\w+)\s*\([^)]*\)\s*(?:const)?(?:\s*override)?(?:\s*final)?(?:\s*\{|\s*;)'
        template_func_pattern = r'template\s*<[^>]*>\s*(?:inline\s+)?(?:virtual\s+)?(?:static\s+)?(?:void|int|[\w:<>\[\]*&]+)\s+(?:\*\s*)?(\w+)\s*\('
        namespace_pattern = r'^namespace\s+(\w+)\s*\{'

        current_file_id = self._current_file_id

        for i, line in enumerate(lines, 1):
            if i > 1 and lines[i-2].strip().startswith("//"):
                docstring = lines[i-2].strip()[2:].strip()
            else:
                docstring = None

            template_class_match = re.search(template_class_pattern, line)
            if template_class_match:
                name = template_class_match.group(1)
                if name not in self.RESERVED_KEYWORDS:
                    node = self._create_class_node(name, current_file_id, i, line, docstring, is_template=True)
                    self._nodes.append(node)
                    continue

            class_match = re.search(class_pattern, line)
            if class_match:
                name = class_match.group(1)
                if name not in self.RESERVED_KEYWORDS:
                    base_classes = []
                    if class_match.lastindex and class_match.lastindex >= 2:
                        bases = class_match.group(2)
                        if bases:
                            base_classes = [b.strip() for b in bases.split(",")]
                    node = self._create_class_node(name, current_file_id, i, line, docstring, base_classes)
                    self._nodes.append(node)
                    self._scope.append(node.id)
                    continue

            namespace_match = re.search(namespace_pattern, line)
            if namespace_match:
                name = namespace_match.group(1)
                if name not in self.RESERVED_KEYWORDS:
                    node = self._create_namespace_node(name, current_file_id, i)
                    self._nodes.append(node)
                    continue

            template_func_match = re.search(template_func_pattern, line)
            if template_func_match and 'class ' not in line and 'struct ' not in line:
                name = template_func_match.group(1)
                if name not in self.RESERVED_KEYWORDS:
                    node = self._create_function_node(name, current_file_id, i, line, docstring, is_template=True)
                    if node:
                        self._nodes.append(node)
                        continue

            func_match = re.search(function_pattern, line)
            if func_match and 'class ' not in line and 'struct ' not in line and 'for ' not in line and 'while ' not in line:
                name = func_match.group(1)
                if name not in self.RESERVED_KEYWORDS:
                    node = self._create_function_node(name, current_file_id, i, line, docstring, is_template=False)
                    if node:
                        self._nodes.append(node)
                        continue

        self._extract_class_methods(content, lines)

    def _extract_class_methods(self, content: str, lines: List[str]):
        visibility = "private"
        in_class = False
        class_indent = 0

        method_pattern = r'(?:virtual\s+)?(?:static\s+)?(?:void|int|long|double|float|char|bool|auto|const|[\w:<>\[\]*&]+)\s+(?:\*\s*)?(\w+)\s*\([^)]*\)\s*(?:const)?(?:\s*override)?(?:\s*final)?(?:\s*\{|\s*;)'

        current_file_id = self._current_file_id

        for i, line in enumerate(lines, 1):
            if re.search(r'(class|struct)\s+\w+', line):
                in_class = True
                class_indent = len(line) - len(line.lstrip())
                continue

            if in_class:
                if 'public:' in line:
                    visibility = "public"
                    continue
                elif 'private:' in line:
                    visibility = "private"
                    continue
                elif 'protected:' in line:
                    visibility = "protected"
                    continue

                current_indent = len(line) - len(line.lstrip())
                if line.strip() and current_indent <= class_indent:
                    in_class = False
                    visibility = "private"
                    continue

                method_match = re.search(method_pattern, line)
                if method_match:
                    method_name = method_match.group(1)
                    if method_name not in self.RESERVED_KEYWORDS:
                        node = self._create_method_node(method_name, current_file_id, i, line, visibility)
                        self._nodes.append(node)

    def _create_class_node(self, name: str, file_id: str, line: int, line_content: str, docstring: str = None, base_classes: List[str] = None, is_template: bool = False) -> CppNode:
        templates = []
        if is_template:
            template_match = re.search(r'template\s*<([^>]+)>', line_content)
            if template_match:
                templates = [t.strip() for t in template_match.group(1).split(",")]

        return CppNode(
            id=self._generate_id("class", name),
            node_type="class",
            name=name,
            qualified_name=name,
            file_id=file_id,
            start_line=line,
            docstring=docstring if docstring else None,
            base_classes=base_classes or [],
            templates=templates,
            is_template=is_template,
        )

    def _create_namespace_node(self, name: str, file_id: str, line: int) -> CppNode:
        return CppNode(
            id=self._generate_id("namespace", name),
            node_type="namespace",
            name=name,
            qualified_name=name,
            file_id=file_id,
            start_line=line,
        )

    def _create_function_node(self, name: str, file_id: str, line: int, line_content: str, docstring: str = None, is_template: bool = False) -> Optional[CppNode]:
        if name.startswith("operator"):
            return None

        return_type_match = re.search(r'(?:virtual\s+)?(?:static\s+)?(?:inline\s+)?(?:extern\s+)?(?:void|int|long|double|float|char|bool|auto|const|[\w:<>\[\]*&]+)', line_content)
        return_type = return_type_match.group(0) if return_type_match else "void"

        is_virtual = "virtual " in line_content
        is_static = "static " in line_content
        is_extern = "extern " in line_content
        is_inline = "inline " in line_content

        return CppNode(
            id=self._generate_id("function", name),
            node_type="function",
            name=name,
            qualified_name=name,
            file_id=file_id,
            start_line=line,
            docstring=docstring if docstring else None,
            return_type=return_type,
            is_virtual=is_virtual,
            is_static=is_static,
            is_extern=is_extern,
            is_inline=is_inline,
            is_template=is_template,
        )

    def _create_method_node(self, name: str, file_id: str, line: int, line_content: str, visibility: str) -> CppNode:
        return_type_match = re.search(r'(?:virtual\s+)?(?:static\s+)?(?:void|int|long|double|float|char|bool|auto|const|[\w:<>\[\]*&]+)', line_content)
        return_type = return_type_match.group(0) if return_type_match else "void"

        is_virtual = "virtual " in line_content
        is_const = " const" in line_content

        return CppNode(
            id=self._generate_id("method", name),
            node_type="method",
            name=name,
            qualified_name=name,
            file_id=file_id,
            start_line=line,
            parent_id=self._scope[-1] if self._scope else None,
            return_type=return_type,
            visibility=visibility,
            is_virtual=is_virtual,
            is_const=is_const,
        )

    def _extract_edges(self):
        for node in self._nodes:
            if node.node_type in ("class", "struct", "namespace"):
                self._edges.append({
                    "edge_type": "contains",
                    "source_id": node.id,
                    "target_id": node.id,
                })

            if node.node_type == "class" and node.base_classes:
                for base in node.base_classes:
                    self._edges.append({
                        "edge_type": "inherits",
                        "source_id": node.id,
                        "target_id": f"class_{base.strip()}",
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
                annotations=node.templates,
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

    _includes: List[str] = []

    def get_frameworks(self) -> List[str]:
        return self._framework_detected


def parse_cpp_code(file_path: str, content: str) -> Tuple[CodeFile, List[CodeNode], List[CodeEdge]]:
    parser = CppParser()

    rel_path = os.path.basename(file_path)
    file_id = str(uuid.uuid4())[:16]

    file = CodeFile(
        id=file_id,
        path=file_path,
        relative_path=rel_path,
        language="cpp",
        size_bytes=len(content.encode("utf-8")),
        line_count=len(content.split("\n")),
    )

    nodes, edges = parser.parse(file_id, content)

    return file, nodes, edges
