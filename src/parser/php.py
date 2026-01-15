"""
PHP Parser for Code Cartographer

Extracts code structure from PHP files using regex-based parsing.
Supports classes, interfaces, traits, functions, methods, and relationships.
"""

import os
import re
import uuid
from dataclasses import dataclass, field
from typing import Any, Dict, List, Optional, Tuple

from src.models.code_map import CodeMap, CodeNode, CodeEdge, CodeFile, Position


@dataclass
class PhpNode:
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
    traits: List[str] = field(default_factory=list)
    visibility: str = "private"
    is_static: bool = False
    is_final: bool = False
    is_abstract: bool = False
    is_magic: bool = False


class PhpParser:
    LANGUAGE_NAME = "php"
    FILE_EXTENSIONS = [".php"]

    RESERVED_KEYWORDS = frozenset({
        "__halt_compiler", "abstract", "and", "array", "as", "break", "callable",
        "case", "catch", "class", "clone", "const", "continue", "declare",
        "default", "die", "do", "echo", "else", "elseif", "empty", "enddeclare",
        "endfor", "endforeach", "endif", "endswitch", "endwhile", "eval", "exit",
        "extends", "final", "finally", "for", "foreach", "function", "global",
        "goto", "if", "implements", "include", "include_once", "instanceof",
        "insteadof", "interface", "isset", "list", "namespace", "new", "or",
        "print", "private", "protected", "public", "require", "require_once",
        "return", "static", "switch", "throw", "trait", "try", "unset", "use",
        "var", "while", "xor", "yield", "int", "float", "string", "bool", "iterable",
        "object", "null", "false", "true", "parent", "self", "void", "never",
    })

    MAGIC_METHODS = frozenset({
        "__construct", "__destruct", "__call", "__callStatic", "__get", "__set",
        "__isset", "__unset", "__sleep", "__wakeup", "__serialize", "__unserialize",
        "__toString", "__invoke", "__clone", "__debugInfo",
    })

    FRAMEWORK_MARKERS = {
        "laravel": ["Illuminate\\", "App\\", "Route::", " artisan "],
        "symfony": ["Symfony\\", "Controller\\", "Service\\"],
        " Yii ": ["Yii\\", "yii\\"],
        "codeigniter": ["CodeIgniter\\", "Controller\\"],
        "cakephp": ["Cake\\", "Controller\\", "Model\\"],
        "drupal": ["Drupal\\", "\\Controller\\"],
        "wordpress": ["WP_", "add_action", "add_filter"],
        "doctrine": ["Doctrine\\", "Entity\\", "Repository\\"],
        "monolog": ["Monolog\\", "Logger::"],
        "guzzle": ["GuzzleHttp\\", "Client::"],
    }

    def __init__(self):
        self._current_file_id = ""
        self._current_file_path = ""
        self._nodes: List[PhpNode] = []
        self._edges: List[Dict] = []
        self._scope: List[str] = []
        self._namespace: str = ""
        self._framework_detected: List[str] = []

    def parse(self, file_id: str, content: str) -> Tuple[List[CodeNode], List[CodeEdge]]:
        self._current_file_id = file_id
        self._current_file_path = ""
        self._nodes = []
        self._edges = []
        self._scope = []
        self._namespace = ""
        self._framework_detected = []

        lines = content.split("\n")
        self._extract_namespace_and_use(content)
        self._extract_nodes(content, lines)
        self._extract_edges()

        return self._to_code_map_objects()

    def _extract_namespace_and_use(self, content: str):
        ns_match = re.search(r'^namespace\s+([\w\\\\]+);', content, re.MULTILINE)
        if ns_match:
            self._namespace = ns_match.group(1)

        for use_match in re.finditer(r'^use\s+([\w\\\\]+)(?:\s+as\s+(\w+))?;', content, re.MULTILINE):
            use_path = use_match.group(1)
            use_alias = use_match.group(2) or use_path.split("\\")[-1]
            self._use_statements[use_alias] = use_path

            for framework, markers in self.FRAMEWORK_MARKERS.items():
                for marker in markers:
                    if marker in use_path:
                        if framework not in self._framework_detected:
                            self._framework_detected.append(framework)

    def _extract_nodes(self, content: str, lines: List[str]):
        class_pattern = r'(?:abstract\s+)?(?:final\s+)?(?:class|interface|trait)\s+(\w+)(?:\s+extends\s+([\w\\\\]+))?(?:\s+implements\s+([\w,\s\\\\]+))?'
        function_pattern = r'(?:public|private|protected|static|final)?\s*(?:function)\s+(\w+)\s*\('
        arrow_function_pattern = r'(?:fn|function)\s*\(([^)]*)\)\s*=>'
        method_pattern = r'(?:public|private|protected|static|final|abstract)?\s*function\s+(\w+)\s*\('

        current_file_id = self._current_file_id

        for i, line in enumerate(lines, 1):
            docstring = None
            if i > 1:
                if lines[i-2].strip().startswith("/**"):
                    docstring = self._parse_phpdoc(lines[i-2])
                elif lines[i-2].strip().startswith("//"):
                    docstring = lines[i-2].strip()[2:].strip()

            class_match = re.search(class_pattern, line)
            if class_match:
                name = class_match.group(1)
                if name not in self.RESERVED_KEYWORDS:
                    node_type = "class"
                    if "interface" in line:
                        node_type = "interface"
                    elif "trait" in line:
                        node_type = "trait"

                    parent = class_match.group(2) if class_match.lastindex and class_match.lastindex >= 2 else None
                    interfaces = []
                    if class_match.lastindex and class_match.lastindex >= 3:
                        if class_match.group(3):
                            interfaces = [iface.strip() for iface in class_match.group(3).split(",")]

                    node = self._create_class_node(name, current_file_id, i, docstring, node_type, parent, interfaces)
                    self._nodes.append(node)
                    self._scope.append(node.id)
                    continue

            method_match = re.search(method_pattern, line)
            if method_match and 'class ' not in line and 'interface ' not in line and 'trait ' not in line:
                method_name = method_match.group(1)
                if method_name not in self.RESERVED_KEYWORDS:
                    node = self._create_method_node(method_name, current_file_id, i, line, docstring)
                    if node:
                        self._nodes.append(node)
                        continue

            func_match = re.search(function_pattern, line)
            if func_match and 'function ' in line:
                func_name = func_match.group(1)
                if func_name not in self.RESERVED_KEYWORDS and not func_name.startswith("_"):
                    node = self._create_function_node(func_name, current_file_id, i, line, docstring)
                    if node:
                        self._nodes.append(node)
                        continue

    def _parse_phpdoc(self, line: str) -> Optional[str]:
        match = re.search(r'/\*\*(.*?)\*/', line)
        if match:
            desc = match.group(1).strip()
            desc = re.sub(r'^\*', '', desc).strip()
            return desc
        return None

    def _create_class_node(self, name: str, file_id: str, line: int, docstring: str, node_type: str, parent: str = None, interfaces: List[str] = None) -> PhpNode:
        is_abstract = "abstract " in self._get_line_context(line)
        is_final = "final " in self._get_line_context(line)

        qualified_name = f"{self._namespace}\\{name}" if self._namespace else name

        return PhpNode(
            id=self._generate_id(node_type, name),
            node_type=node_type,
            name=name,
            qualified_name=qualified_name,
            file_id=file_id,
            start_line=line,
            docstring=docstring if docstring else None,
            interfaces=interfaces or [],
            is_abstract=is_abstract,
            is_final=is_final,
        )

    def _create_method_node(self, name: str, file_id: str, line: int, line_content: str, docstring: str = None) -> Optional[PhpNode]:
        if name in self.MAGIC_METHODS:
            is_magic = True

        visibility = "private"
        if "public " in line_content:
            visibility = "public"
        elif "protected " in line_content:
            visibility = "protected"

        is_static = "static " in line_content
        is_final = "final " in line_content
        is_abstract = "abstract " in line_content

        return_type_match = re.search(r':\s*([\w?\\\\\[\]]+)', line_content)
        return_type = return_type_match.group(1) if return_type_match else None

        return PhpNode(
            id=self._generate_id("method", name),
            node_type="method",
            name=name,
            qualified_name=name,
            file_id=file_id,
            start_line=line,
            parent_id=self._scope[-1] if self._scope else None,
            docstring=docstring if docstring else None,
            return_type=return_type,
            visibility=visibility,
            is_static=is_static,
            is_final=is_final,
            is_abstract=is_abstract,
            is_magic=name in self.MAGIC_METHODS,
        )

    def _create_function_node(self, name: str, file_id: str, line: int, line_content: str, docstring: str = None) -> Optional[PhpNode]:
        if name.startswith("_"):
            return None

        return_type_match = re.search(r':\s*([\w?\\\\\[\]]+)', line_content)
        return_type = return_type_match.group(1) if return_type_match else None

        return PhpNode(
            id=self._generate_id("function", name),
            node_type="function",
            name=name,
            qualified_name=name,
            file_id=file_id,
            start_line=line,
            docstring=docstring if docstring else None,
            return_type=return_type,
        )

    def _get_line_context(self, line_num: int) -> str:
        return ""

    def _extract_edges(self):
        for node in self._nodes:
            if node.node_type in ("class", "interface", "trait"):
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

    _use_statements: Dict[str, str] = {}

    def get_frameworks(self) -> List[str]:
        return self._framework_detected


def parse_php_code(file_path: str, content: str) -> Tuple[CodeFile, List[CodeNode], List[CodeEdge]]:
    parser = PhpParser()

    rel_path = os.path.basename(file_path)
    file_id = str(uuid.uuid4())[:16]

    file = CodeFile(
        id=file_id,
        path=file_path,
        relative_path=rel_path,
        language="php",
        size_bytes=len(content.encode("utf-8")),
        line_count=len(content.split("\n")),
    )

    nodes, edges = parser.parse(file_id, content)

    return file, nodes, edges
