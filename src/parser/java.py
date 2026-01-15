"""
Java Parser for Code Cartographer

Extracts code structure from Java files using regex-based parsing.
Supports classes, methods, interfaces, and relationships.
"""

import os
import re
import uuid
from dataclasses import dataclass, field
from typing import Any, Dict, List, Optional, Tuple

from src.models.code_map import CodeMap, CodeNode, CodeEdge, CodeFile, Position


@dataclass
class JavaNode:
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
    extends: List[str] = field(default_factory=list)
    implements: List[str] = field(default_factory=list)
    annotations: List[str] = field(default_factory=list)
    modifiers: List[str] = field(default_factory=list)
    is_static: bool = False
    is_final: bool = False
    is_abstract: bool = False


class JavaParser:
    LANGUAGE_NAME = "java"
    FILE_EXTENSIONS = [".java"]

    RESERVED_TYPES = frozenset({
        "abstract", "assert", "boolean", "break", "byte", "case", "catch",
        "char", "class", "const", "continue", "default", "do", "double",
        "else", "enum", "extends", "final", "finally", "float", "for",
        "goto", "if", "implements", "import", "instanceof", "int",
        "interface", "long", "native", "new", "package", "private",
        "protected", "public", "return", "short", "static", "strictfp",
        "super", "switch", "synchronized", "this", "throw", "throws",
        "transient", "try", "void", "volatile", "while", "true", "false", "null"
    })

    FRAMEWORK_MARKERS = {
        "spring": ["org.springframework", "@RestController", "@Controller", "@Service", "@Repository"],
        "junit": ["org.junit", "@Test", "@Before", "@After"],
        "servlet": ["javax.servlet", "HttpServlet"],
        "android": ["android.", "Activity", "Fragment"],
    }

    def __init__(self):
        self._current_file_id = ""
        self._current_file_path = ""
        self._nodes: List[JavaNode] = []
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
        self._extract_package_and_imports(content)
        self._extract_nodes(content, lines)
        self._extract_edges()

        return self._to_code_map_objects()

    def _extract_package_and_imports(self, content: str):
        package_match = re.search(r'^package\s+([\w.]+);', content, re.MULTILINE)
        if package_match:
            self._package = package_match.group(1)
        else:
            self._package = ""

        for import_match in re.finditer(r'^import\s+(?:static\s+)?([\w.]+(?:\.\*)?);', content, re.MULTILINE):
            import_path = import_match.group(1)
            self._imports.append(import_path)

            for framework, markers in self.FRAMEWORK_MARKERS.items():
                for marker in markers:
                    if marker in import_path:
                        if framework not in self._framework_detected:
                            self._framework_detected.append(framework)

    def _extract_nodes(self, content: str, lines: List[str]):
        class_pattern = r'(?:public|private|protected|static|final|abstract|sealed|non-sealed)?\s*(?:class|interface|enum|record)\s+(\w+)'
        extends_implements = r'(?:\s+extends\s+([\w.,\s]+))?(?:\s+implements\s+([\w.,\s]+))?'

        method_pattern = r'(?:public|private|protected|static|final|abstract|synchronized|native)?\s*(?:void|boolean|int|long|double|float|char|byte|short|String|[\w.<>[\]]+)\s+(\w+)\s*\([^)]*\)'
        constructor_pattern = r'(?:public|private|protected)?\s+(?!void|boolean|int|long|double|float|char|byte|short|String|[\w.<>[\]]+)(\w+)\s*\([^)]*\)'

        annotation_pattern = r'@(\w+)(?:\([^)]*\))?'

        for i, line in enumerate(lines, 1):
            if i > 3 and lines[i-2].strip().startswith("//"):
                continue

            class_match = re.search(class_pattern + extends_implements, line)
            if class_match:
                name = class_match.group(1)
                if name not in self.RESERVED_TYPES:
                    node = self._create_class_node(name, class_match, self._current_file_id, i)
                    self._nodes.append(node)
                    self._scope.append(node.id)
                    continue

            method_match = re.search(method_pattern, line)
            if method_match and 'class ' not in line and 'interface ' not in line:
                method_name = method_match.group(1)
                if method_name not in self.RESERVED_TYPES and method_name[0].islower():
                    node = self._create_method_node(method_name, self._current_file_id, i, line)
                    if node:
                        self._nodes.append(node)
                        continue

            constructor_match = re.search(constructor_pattern, line)
            if constructor_match:
                cons_name = constructor_match.group(1)
                if cons_name not in self.RESERVED_TYPES:
                    node = self._create_constructor_node(cons_name, self._current_file_id, i)
                    self._nodes.append(node)
                    continue

        self._extract_fields(content, lines)

    def _extract_fields(self, content: str, lines: List[str]):
        field_pattern = r'(?:public|private|protected|static|final|transient|volatile)?\s*(?:final)?\s*(?:void|boolean|int|long|double|float|char|byte|short|String|[\w.<>\[\]]+)\s+(\w+)\s*(?:=|;|,)'

        for i, line in enumerate(lines, 1):
            field_match = re.search(field_pattern, line)
            if field_match:
                field_name = field_match.group(1)
                if field_name not in self.RESERVED_TYPES:
                    node = self._create_field_node(field_name, self._current_file_id, i, line)
                    self._nodes.append(node)

    def _create_class_node(self, name: str, match, file_id: str, line: int) -> JavaNode:
        node_type = "class"
        if "interface" in match.group(0):
            node_type = "interface"
        elif "enum" in match.group(0):
            node_type = "enum"
        elif "record" in match.group(0):
            node_type = "record"

        extends = []
        implements = []

        if match.lastindex and match.lastindex >= 2:
            extends_str = match.group(2)
            if extends_str:
                extends = [e.strip() for e in extends_str.split(",")]

        if match.lastindex and match.lastindex >= 3:
            implements_str = match.group(3)
            if implements_str:
                implements = [e.strip() for e in implements_str.split(",")]

        return JavaNode(
            id=self._generate_id(node_type, name),
            node_type=node_type,
            name=name,
            qualified_name=name,
            file_id=file_id,
            start_line=line,
            docstring=None,
            parameters=[],
            extends=extends,
            implements=implements,
        )

    def _create_method_node(self, name: str, file_id: str, line: int, line_content: str) -> Optional[JavaNode]:
        return_type_match = re.search(r'(?:public|private|protected|static|final|abstract|synchronized|native)?\s*((?:void|boolean|int|long|double|float|char|byte|short|String|[\w.<>\[\]]+))\s+\w+\s*\(', line_content)
        return_type = return_type_match.group(1) if return_type_match else "void"

        modifiers = []
        is_static = "static " in line_content
        is_final = "final " in line_content
        is_abstract = "abstract " in line_content

        if is_static:
            modifiers.append("static")
        if is_final:
            modifiers.append("final")
        if is_abstract:
            modifiers.append("abstract")

        return JavaNode(
            id=self._generate_id("method", name),
            node_type="method",
            name=name,
            qualified_name=name,
            file_id=file_id,
            start_line=line,
            parent_id=self._scope[-1] if self._scope else None,
            parameters=[],
            return_type=return_type,
            modifiers=modifiers,
            is_static=is_static,
            is_final=is_final,
            is_abstract=is_abstract,
        )

    def _create_constructor_node(self, name: str, file_id: str, line: int) -> JavaNode:
        return JavaNode(
            id=self._generate_id("constructor", name),
            node_type="constructor",
            name=name,
            qualified_name=name,
            file_id=file_id,
            start_line=line,
            parent_id=self._scope[-1] if self._scope else None,
            parameters=[],
        )

    def _create_field_node(self, name: str, file_id: str, line: int, line_content: str) -> JavaNode:
        type_match = re.search(r'(void|boolean|int|long|double|float|char|byte|short|String|[\w.<>\[\]]+)\s+' + name, line_content)
        field_type = type_match.group(1) if type_match else "Object"

        return JavaNode(
            id=self._generate_id("field", name),
            node_type="field",
            name=name,
            qualified_name=name,
            file_id=file_id,
            start_line=line,
            parent_id=self._scope[-1] if self._scope else None,
            return_type=field_type,
        )

    def _extract_edges(self):
        for node in self._nodes:
            if node.node_type in ("class", "interface"):
                self._edges.append({
                    "edge_type": "contains",
                    "source_id": node.id,
                    "target_id": node.id,
                })

                for base in node.extends:
                    self._edges.append({
                        "edge_type": "extends",
                        "source_id": node.id,
                        "target_id": f"class_{base}",
                    })

                for impl in node.implements:
                    self._edges.append({
                        "edge_type": "implements",
                        "source_id": node.id,
                        "target_id": f"interface_{impl}",
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
                annotations=node.annotations,
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

    _package: str = ""
    _imports: List[str] = []

    def get_frameworks(self) -> List[str]:
        return self._framework_detected


def parse_java_code(file_path: str, content: str) -> Tuple[CodeFile, List[CodeNode], List[CodeEdge]]:
    parser = JavaParser()

    rel_path = os.path.basename(file_path)
    file_id = str(uuid.uuid4())[:16]

    file = CodeFile(
        id=file_id,
        path=file_path,
        relative_path=rel_path,
        language="java",
        size_bytes=len(content.encode("utf-8")),
        line_count=len(content.split("\n")),
    )

    nodes, edges = parser.parse(file_id, content)

    return file, nodes, edges
