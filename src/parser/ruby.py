"""
Ruby Parser for Code Cartographer

Extracts code structure from Ruby files using regex-based parsing.
Supports classes, modules, methods, and relationships.
"""

import os
import re
import uuid
from dataclasses import dataclass, field
from typing import Any, Dict, List, Optional, Tuple

from src.models.code_map import CodeMap, CodeNode, CodeEdge, CodeFile, Position


@dataclass
class RubyNode:
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
    includes: List[str] = field(default_factory=list)
    extends: List[str] = field(default_factory=list)
    visibility: str = "public"
    is_class_method: bool = False
    is_module_function: bool = False


class RubyParser:
    LANGUAGE_NAME = "ruby"
    FILE_EXTENSIONS = [".rb", ".erb"]

    RESERVED_KEYWORDS = frozenset({
        "BEGIN", "END", "alias", "and", "begin", "break", "case", "class",
        "def", "defined?", "do", "else", "elsif", "end", "ensure", "false",
        "for", "if", "in", "module", "next", "nil", "not", "or", "redo",
        "rescue", "retry", "return", "self", "super", "then", "true",
        "undef", "unless", "until", "when", "while", "yield", "__FILE__",
        "__LINE__", "__ENCODING__",
    })

    FRAMEWORK_MARKERS = {
        "rails": ["Rails::", "ActiveRecord", "ActionController", "ApplicationController"],
        "sinatra": ["Sinatra::", "require 'sinatra'"],
        "rack": ["Rack::", "require 'rack'"],
        "rspec": ["describe", "it ", "expect", "RSpec"],
        "minitest": ["require 'minitest'", "describe ", "test "],
        "sidekiq": ["Sidekiq::", "include Sidekiq::Worker"],
        "puma": ["Puma::", "require 'puma'"],
        "capistrano": ["Capistrano::", "require 'capistrano'"],
    }

    def __init__(self):
        self._current_file_id = ""
        self._current_file_path = ""
        self._nodes: List[RubyNode] = []
        self._edges: List[Dict] = []
        self._scope: List[str] = []
        self._current_class: str = ""
        self._current_module: str = ""
        self._framework_detected: List[str] = []

    def parse(self, file_id: str, content: str) -> Tuple[List[CodeNode], List[CodeEdge]]:
        self._current_file_id = file_id
        self._current_file_path = ""
        self._nodes = []
        self._edges = []
        self._scope = []
        self._current_class = ""
        self._current_module = ""
        self._framework_detected = []

        lines = content.split("\n")
        self._extract_requires_and_includes(content)
        self._extract_nodes(content, lines)
        self._extract_edges()

        return self._to_code_map_objects()

    def _extract_requires_and_includes(self, content: str):
        for require_match in re.finditer(r'^(?:require|require_relative)\s+[\'"]([^\'"]+)[\'"]', content, re.MULTILINE):
            req_path = require_match.group(1)

            for framework, markers in self.FRAMEWORK_MARKERS.items():
                for marker in markers:
                    if marker in req_path:
                        if framework not in self._framework_detected:
                            self._framework_detected.append(framework)

    def _extract_nodes(self, content: str, lines: List[str]):
        class_pattern = r'^(?:private|protected|public\s+)?\s*class\s+(?:<<\s*(\w+)|(\w+))(?:\s*<\s*([\w:]+))?'
        module_pattern = r'^module\s+(\w+)'
        def_pattern = r'^(?:private|protected|public)?\s*def\s+((?:[\w]+\.)?#{?\w+}?|self\.\w+|(\w+))(?:\s*\([^)]*\))?'
        class_method_pattern = r'^def\s+self\.(\w+)'
        attr_pattern = r'^(?:attr_reader|attr_writer|attr_accessor)\s+:(\w+)'

        current_file_id = self._current_file_id

        for i, line in enumerate(lines, 1):
            if i > 1 and lines[i-2].strip().startswith("#"):
                docstring = lines[i-2].strip()[1:].strip()
            else:
                docstring = None

            class_match = re.search(class_pattern, line)
            if class_match:
                name = class_match.group(1) or class_match.group(2)
                if name and name not in self.RESERVED_KEYWORDS:
                    parent_class = class_match.group(3)
                    node = self._create_class_node(name, current_file_id, i, docstring, parent_class)
                    self._nodes.append(node)
                    self._scope.append(node.id)
                    self._current_class = name
                    continue

            module_match = re.search(module_pattern, line)
            if module_match:
                name = module_match.group(1)
                if name not in self.RESERVED_KEYWORDS:
                    node = self._create_module_node(name, current_file_id, i, docstring)
                    self._nodes.append(node)
                    self._scope.append(node.id)
                    self._current_module = name
                    continue

            class_method_match = re.search(class_method_pattern, line)
            if class_method_match:
                name = class_method_match.group(1)
                if name not in self.RESERVED_KEYWORDS:
                    node = self._create_method_node(f"self.{name}", current_file_id, i, line, docstring, is_class_method=True)
                    if node:
                        self._nodes.append(node)
                        continue

            def_match = re.search(def_pattern, line)
            if def_match and 'class ' not in line and 'module ' not in line:
                full_name = def_match.group(1)
                name = full_name.split('.')[-1] if '.' in full_name else full_name
                if name and name not in self.RESERVED_KEYWORDS:
                    node = self._create_method_node(full_name, current_file_id, i, line, docstring, is_class_method=False)
                    if node:
                        self._nodes.append(node)
                        continue

            attr_match = re.search(attr_pattern, line)
            if attr_match:
                attr_name = attr_match.group(1)
                node = self._create_attr_node(attr_name, current_file_id, i, line)
                self._nodes.append(node)
                continue

    def _create_class_node(self, name: str, file_id: str, line: int, docstring: str = None, parent_class: str = None) -> RubyNode:
        includes = self._extract_includes_or_extends(content, "include") if False else []
        extends = self._extract_includes_or_extends(content, "extend") if False else []

        return RubyNode(
            id=self._generate_id("class", name),
            node_type="class",
            name=name,
            qualified_name=name,
            file_id=file_id,
            start_line=line,
            docstring=docstring if docstring else None,
            includes=includes,
            extends=extends,
        )

    def _create_module_node(self, name: str, file_id: str, line: int, docstring: str = None) -> RubyNode:
        return RubyNode(
            id=self._generate_id("module", name),
            node_type="module",
            name=name,
            qualified_name=name,
            file_id=file_id,
            start_line=line,
            docstring=docstring if docstring else None,
        )

    def _create_method_node(self, name: str, file_id: str, line: int, line_content: str, docstring: str = None, is_class_method: bool = False) -> Optional[RubyNode]:
        if name.startswith("_"):
            return None

        params_match = re.search(r'def\s+[\w.]+\s*(\([^)]*\))?', line_content)
        parameters = []
        if params_match:
            params_str = params_match.group(1)
            if params_str and params_str.strip():
                params_str = params_str.strip("() ")
                for param in params_str.split(","):
                    param = param.strip()
                    if param:
                        default_match = re.search(r'(\w+)(?:\s*=\s*(.+))?', param)
                        if default_match:
                            parameters.append({
                                "name": default_match.group(1),
                                "default": default_match.group(2) if default_match.lastindex and default_match.group(2) else None,
                            })

        visibility = "private" if "private" in self._get_visibility_context([], line) else "public"

        return RubyNode(
            id=self._generate_id("method", name.replace('.', '_')),
            node_type="method",
            name=name,
            qualified_name=name,
            file_id=file_id,
            start_line=line,
            parent_id=self._scope[-1] if self._scope else None,
            docstring=docstring if docstring else None,
            parameters=parameters,
            visibility=visibility,
            is_class_method=is_class_method,
        )

    def _create_attr_node(self, name: str, file_id: str, line: int, line_content: str) -> RubyNode:
        attr_type = "reader"
        if "attr_writer" in line_content:
            attr_type = "writer"
        elif "attr_accessor" in line_content:
            attr_type = "accessor"

        return RubyNode(
            id=self._generate_id("attribute", name),
            node_type="attribute",
            name=name,
            qualified_name=name,
            file_id=file_id,
            start_line=line,
            parent_id=self._scope[-1] if self._scope else None,
        )

    def _get_visibility_context(self, lines: List[str], current_line: int) -> str:
        return "public"

    def _extract_includes_or_extends(self, content: str, keyword: str) -> List[str]:
        includes = []
        pattern = rf'^{keyword}\s+([\w:]+)'
        for match in re.finditer(pattern, content, re.MULTILINE):
            includes.append(match.group(1))
        return includes

    def _extract_edges(self):
        for node in self._nodes:
            if node.node_type in ("class", "module"):
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
                annotations=node.includes,
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


def parse_ruby_code(file_path: str, content: str) -> Tuple[CodeFile, List[CodeNode], List[CodeEdge]]:
    parser = RubyParser()

    rel_path = os.path.basename(file_path)
    file_id = str(uuid.uuid4())[:16]

    file = CodeFile(
        id=file_id,
        path=file_path,
        relative_path=rel_path,
        language="ruby",
        size_bytes=len(content.encode("utf-8")),
        line_count=len(content.split("\n")),
    )

    nodes, edges = parser.parse(file_id, content)

    return file, nodes, edges
