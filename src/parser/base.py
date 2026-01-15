"""
AST Parser Base for Code Cartographer

Abstract base class and utilities for AST-based code parsing.
Supports multiple languages through extensible parser implementations.
"""

import os
import hashlib
from abc import ABC, abstractmethod
from typing import Dict, List, Optional, Any, Type

from src.models.code_map import CodeMap, CodeNode, CodeEdge, CodeFile, Position


class ParseOptions:
    def __init__(
        self,
        include_private: bool = False,
        include_tests: bool = True,
        include_generated: bool = False,
        extract_docstrings: bool = True,
        extract_comments: bool = False,
        calculate_complexity: bool = True,
        follow_imports: bool = False,
        max_depth: int = 10,
    ):
        self.include_private = include_private
        self.include_tests = include_tests
        self.include_generated = include_generated
        self.extract_docstrings = extract_docstrings
        self.extract_comments = extract_comments
        self.calculate_complexity = calculate_complexity
        self.follow_imports = follow_imports
        self.max_depth = max_depth


class ParseResult:
    def __init__(
        self,
        code_map: CodeMap,
        files_processed: int = 0,
        nodes_extracted: int = 0,
        edges_created: int = 0,
        errors: Optional[List[str]] = None,
        warnings: Optional[List[str]] = None,
    ):
        self.code_map = code_map
        self.files_processed = files_processed
        self.nodes_extracted = nodes_extracted
        self.edges_created = edges_created
        self.errors = errors or []
        self.warnings = warnings or []

    def to_dict(self) -> Dict:
        return {
            "files_processed": self.files_processed,
            "nodes_extracted": self.nodes_extracted,
            "edges_created": self.edges_created,
            "errors": self.errors,
            "warnings": self.warnings,
            "statistics": self.code_map.statistics,
        }


class ASTParser(ABC):
    LANGUAGE_NAME: str = ""
    FILE_EXTENSIONS: List[str] = []
    LANGUAGE_ID: str = ""

    def __init__(self, options: Optional[ParseOptions] = None):
        self.options = options or ParseOptions()
        self._node_map: Dict[str, CodeNode] = {}
        self._current_file_id: Optional[str] = None
        self._current_qualified_name: str = ""
        self._current_scope: List[str] = []
        self._current_file: Optional[CodeFile] = None

    @abstractmethod
    def parse_file(self, file_path: str, content: str) -> CodeFile:
        pass

    @abstractmethod
    def extract_nodes(self, file_id: str, content: str, tree: Any) -> List[CodeNode]:
        pass

    @abstractmethod
    def extract_edges(self, nodes: List[CodeNode], tree: Any) -> List[CodeEdge]:
        pass

    @abstractmethod
    def get_tree_sitter_language(self) -> str:
        pass

    def can_parse(self, file_path: str) -> bool:
        ext = os.path.splitext(file_path)[1].lower()
        return ext in self.FILE_EXTENSIONS

    def should_skip_file(self, file_path: str) -> bool:
        filename = os.path.basename(file_path)
        if not self.options.include_tests:
            if filename.endswith("_test.py") or filename.startswith("test_"):
                return True
        if not self.options.include_generated:
            if "generated" in file_path.lower() or file_path.lower().startswith("gen_"):
                return True
        if not self.options.include_private:
            if filename.startswith("_") and not filename.startswith("__"):
                return True
        return False

    def _get_position(self, node: Any) -> Position:
        start_point = node.start_point if hasattr(node, "start_point") else (0, 0)
        end_point = node.end_point if hasattr(node, "end_point") else (0, 0)
        return Position(
            line=start_point[0] + 1,
            column=start_point[1],
            end_line=end_point[0] + 1,
            end_column=end_point[1],
        )

    def _calculate_complexity(self, node: Any) -> int:
        if not self.options.calculate_complexity:
            return 1
        if not hasattr(node, "type"):
            return 1
        node_type = node.type
        complexity_map = {"if": 1, "elif": 1, "for": 1, "while": 1, "case": 1, "when": 1, "and": 1, "or": 1}
        return 1 + sum(1 for key in complexity_map if key in node_type)

    def _extract_docstring(self, node: Any) -> Optional[str]:
        if not self.options.extract_docstrings:
            return None
        return None

    def _make_qualified_name(self, name: str) -> str:
        if self._current_qualified_name:
            return f"{self._current_qualified_name}.{name}"
        return name

    def _push_scope(self, name: str):
        self._current_scope.append(name)
        self._current_qualified_name = self._make_qualified_name(name)

    def _pop_scope(self):
        if self._current_scope:
            self._current_scope.pop()
        if self._current_scope:
            self._current_qualified_name = ".".join(self._current_scope)
        else:
            self._current_qualified_name = ""

    def _generate_id(self, node_type: str, name: str) -> str:
        content = f"{self._current_file_id}:{self._current_qualified_name}:{name}"
        return hashlib.md5(content.encode()).hexdigest()[:16]

    def _create_basic_node(
        self,
        node_type: str,
        name: str,
        position: Position,
        additional_attrs: Optional[Dict] = None,
    ) -> CodeNode:
        qualified_name = self._make_qualified_name(name)
        node_id = self._generate_id(node_type, name)
        file_id = self._current_file_id or "unknown"
        node = CodeNode(
            id=node_id,
            node_type=node_type,
            name=name,
            qualified_name=qualified_name,
            file_id=file_id,
            start_position=position,
            parent_id=self._current_scope[-1] if self._current_scope else None,
            complexity=self._calculate_complexity(None),
        )
        if additional_attrs:
            for key, value in additional_attrs.items():
                setattr(node, key, value)
        self._node_map[node_id] = node
        return node

    def _extract_visibility(self, name: str) -> str:
        if name.startswith("__") and name.endswith("__"):
            return "dunder"
        if name.startswith("_"):
            return "protected"
        if name.isupper():
            return "constant"
        return "public"

    def _extract_decorators(self, node: Any) -> List[str]:
        return []

    def _extract_annotations(self, node: Any) -> List[str]:
        return []

    def _extract_parameters(self, node: Any) -> List[Dict]:
        return []

    def _extract_return_type(self, node: Any) -> Optional[str]:
        return None

    def _extract_code_snippet(self, content: str, position: Position, max_lines: int = 10) -> Optional[str]:
        lines = content.split("\n")
        start_line = max(0, position.line - 1)
        end_line = min(len(lines), start_line + max_lines)
        snippet_lines = lines[start_line:end_line]
        if len(lines) > start_line + max_lines:
            snippet_lines.append("...")
        return "\n".join(snippet_lines) if snippet_lines else None


class ParserRegistry:
    _parsers: Dict[str, Type[ASTParser]] = {}
    _language_map: Dict[str, Type[ASTParser]] = {}

    @classmethod
    def register(cls, parser_class: Type[ASTParser]):
        cls._parsers[parser_class.LANGUAGE_NAME.lower()] = parser_class
        for ext in parser_class.FILE_EXTENSIONS:
            cls._language_map[ext.lower()] = parser_class
        return parser_class

    @classmethod
    def get_parser(cls, file_path: str) -> Optional[ASTParser]:
        ext = os.path.splitext(file_path)[1].lower()
        if ext in cls._language_map:
            parser_class = cls._language_map[ext]
            return parser_class()
        return None

    @classmethod
    def get_parser_by_language(cls, language_id: str) -> Optional[ASTParser]:
        language_id = language_id.lower()
        if language_id in cls._parsers:
            parser_class = cls._parsers[language_id]
            return parser_class()
        return None

    @classmethod
    def list_languages(cls) -> List[str]:
        return list(cls._parsers.keys())

    @classmethod
    def list_extensions(cls) -> List[str]:
        return list(cls._language_map.keys())


def get_language_from_extension(extension: str) -> Optional[str]:
    extension = extension.lower()
    mapping = {
        ".py": "python",
        ".js": "javascript",
        ".ts": "typescript",
        ".jsx": "tsx",
        ".tsx": "tsx",
        ".java": "java",
        ".cpp": "cpp",
        ".cc": "cpp",
        ".c": "c",
        ".h": "c",
        ".hpp": "cpp",
        ".cs": "csharp",
        ".go": "go",
        ".rs": "rust",
        ".rb": "ruby",
        ".php": "php",
        ".swift": "swift",
        ".kt": "kotlin",
        ".scala": "scala",
    }
    return mapping.get(extension)
