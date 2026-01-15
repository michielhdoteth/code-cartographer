"""
Code Cartographer Agent API for Claude Code

Interface for Claude Code agents to map codebases.
Creates .code-map/ folder with code-map.toon and code-graph.md.

Usage:
    from src.agent import CodeCartographerAgent

    agent = CodeCartographerAgent()
    agent.init("/path/to/project")
    agent.scan()
    agent.parse()
    agent.generate_graph()
    agent.full()

    agent.query_node("ClassName")
    agent.query_edges("calls")
"""

import os
import sys
import ast
import uuid
import hashlib
import json
from typing import Dict, List, Optional, Any, Tuple
from datetime import datetime
from concurrent.futures import ThreadPoolExecutor, as_completed
from functools import partial

from src.models.code_map import CodeMap, CodeNode, CodeEdge, CodeFile, Position
from src.formatter import to_toon, from_toon


def _get_parser_for_language(language: str):
    """Lazy import parsers to avoid import overhead in worker processes."""
    from src.parser.python import parse_python_code
    from src.parser.javascript import parse_javascript_code
    from src.parser.typescript import parse_typescript_code
    from src.parser.java import parse_java_code
    from src.parser.rust import parse_rust_code
    from src.parser.go import parse_go_code
    from src.parser.cpp import parse_cpp_code
    from src.parser.ruby import parse_ruby_code
    from src.parser.php import parse_php_code

    PARSERS = {
        "python": parse_python_code,
        "javascript": parse_javascript_code,
        "typescript": parse_typescript_code,
        "java": parse_java_code,
        "rust": parse_rust_code,
        "go": parse_go_code,
        "cpp": parse_cpp_code,
        "ruby": parse_ruby_code,
        "php": parse_php_code,
    }
    return PARSERS.get(language)


def _parse_file_task(file_path: str, file_id: str, language: str) -> Optional[Tuple[str, List[CodeNode], List[CodeEdge]]]:
    """Worker function to parse a single file. Returns (file_id, nodes, edges)."""
    try:
        parser = _get_parser_for_language(language)
        if parser is None:
            return None

        with open(file_path, "r", encoding="utf-8", errors="ignore") as f:
            content = f.read()

        _, nodes, edges = parser(file_path, content)
        return (file_id, nodes, edges)
    except Exception:
        return None


class CodeCartographerAgent:
    FOLDER_NAME = ".code-map"
    TOON_FILE = "code-map.toon"
    MD_FILE = "code-graph.md"
    SUPPORTED_EXTENSIONS = {".py", ".js", ".ts", ".tsx", ".java", ".go", ".rs"}
    DEFAULT_MAX_WORKERS = min(8, (os.cpu_count() or 1) + 4)

    def __init__(self, max_workers: Optional[int] = None):
        self._root_path: Optional[str] = None
        self._code_map: Optional[CodeMap] = None
        self._max_workers = max_workers or self.DEFAULT_MAX_WORKERS

    @property
    def folder_path(self) -> str:
        if not self._root_path:
            raise ValueError("Not initialized. Call init() first.")
        return os.path.join(self._root_path, self.FOLDER_NAME)

    @property
    def toon_path(self) -> str:
        return os.path.join(self.folder_path, self.TOON_FILE)

    @property
    def md_path(self) -> str:
        return os.path.join(self.folder_path, self.MD_FILE)

    def init(self, root_path: str, name: str = "code-map") -> Dict[str, Any]:
        self._root_path = os.path.abspath(root_path)
        os.makedirs(self.folder_path, exist_ok=True)

        self._code_map = CodeMap(
            id="",
            name=name,
            root_path=self._root_path,
            created_at=datetime.now().isoformat(),
            updated_at=datetime.now().isoformat(),
        )

        self._save_toon()
        self._save_md()

        return {
            "status": "initialized",
            "folder": self.folder_path,
            "toon_file": self.toon_path,
            "md_file": self.md_path,
        }

    def scan(self) -> Dict[str, Any]:
        if not self._code_map:
            raise ValueError("Not initialized. Call init() first.")

        files_found = 0
        for file_path in self._scan_files():
            ext = os.path.splitext(file_path)[1].lower()
            if ext not in self.SUPPORTED_EXTENSIONS:
                continue

            try:
                with open(file_path, "r", encoding="utf-8") as f:
                    content = f.read()
                file = self._create_file_record(file_path, content)
                self._code_map.add_file(file)
                files_found += 1
            except Exception:
                continue

        self._code_map.update_statistics()
        self._save_toon()

        return {
            "status": "scanned",
            "files_found": files_found,
            "total_files": len(self._code_map.files),
        }

    def parse(self, parallel: bool = True) -> Dict[str, Any]:
        if not self._code_map:
            raise ValueError("Not initialized. Call init() first.")

        self._load_toon()

        if parallel and len(self._code_map.files) > 10:
            return self._parse_parallel()
        else:
            return self._parse_sequential()

    def _parse_parallel(self) -> Dict[str, Any]:
        """Parse files in parallel using ThreadPoolExecutor."""
        from src.parser.python import parse_python_code
        from src.parser.javascript import parse_javascript_code
        from src.parser.typescript import parse_typescript_code
        from src.parser.java import parse_java_code
        from src.parser.rust import parse_rust_code
        from src.parser.go import parse_go_code
        from src.parser.cpp import parse_cpp_code
        from src.parser.ruby import parse_ruby_code
        from src.parser.php import parse_php_code

        PARSERS = {
            "python": parse_python_code,
            "javascript": parse_javascript_code,
            "typescript": parse_typescript_code,
            "java": parse_java_code,
            "rust": parse_rust_code,
            "go": parse_go_code,
            "cpp": parse_cpp_code,
            "ruby": parse_ruby_code,
            "php": parse_php_code,
        }

        files_to_parse = [
            (file_id, file.path, file.language)
            for file_id, file in self._code_map.files.items()
            if file.language in PARSERS
        ]

        nodes_added = 0
        edges_added = 0
        files_processed = 0
        errors = 0

        def parse_task(file_id: str, path: str, lang: str):
            try:
                parser = PARSERS.get(lang)
                if parser is None:
                    return None

                with open(path, "r", encoding="utf-8", errors="ignore") as f:
                    content = f.read()

                _, nodes, edges = parser(path, content)
                return (file_id, nodes, edges)
            except Exception:
                return None

        with ThreadPoolExecutor(max_workers=self._max_workers) as executor:
            futures = [
                executor.submit(parse_task, file_id, path, lang)
                for file_id, path, lang in files_to_parse
            ]

            for future in as_completed(futures):
                try:
                    result = future.result()
                    if result:
                        file_id, nodes, edges = result
                        for node in nodes:
                            self._code_map.add_node(node, file_id)
                            nodes_added += 1
                        for edge in edges:
                            self._code_map.add_edge(edge)
                            edges_added += 1
                        files_processed += 1
                    else:
                        errors += 1
                except Exception:
                    errors += 1

        self._code_map.update_statistics()
        self._save_toon()

        return {
            "status": "parsed",
            "mode": "parallel",
            "workers": self._max_workers,
            "nodes_added": nodes_added,
            "edges_added": edges_added,
            "files_processed": files_processed,
            "errors": errors,
            "total_nodes": len(self._code_map.nodes),
            "total_edges": len(self._code_map.edges),
        }

    def _parse_sequential(self) -> Dict[str, Any]:
        """Parse files sequentially (fallback for small codebases)."""
        from src.parser.python import parse_python_code
        from src.parser.javascript import parse_javascript_code
        from src.parser.typescript import parse_typescript_code
        from src.parser.java import parse_java_code
        from src.parser.rust import parse_rust_code
        from src.parser.go import parse_go_code
        from src.parser.cpp import parse_cpp_code
        from src.parser.ruby import parse_ruby_code
        from src.parser.php import parse_php_code

        PARSERS = {
            "python": parse_python_code,
            "javascript": parse_javascript_code,
            "typescript": parse_typescript_code,
            "java": parse_java_code,
            "rust": parse_rust_code,
            "go": parse_go_code,
            "cpp": parse_cpp_code,
            "ruby": parse_ruby_code,
            "php": parse_php_code,
        }

        nodes_added = 0
        edges_added = 0
        files_processed = 0

        for file_id, file in list(self._code_map.files.items()):
            try:
                if file.language not in PARSERS:
                    continue

                with open(file.path, "r", encoding="utf-8", errors="ignore") as f:
                    content = f.read()

                parser = PARSERS[file.language]
                _, nodes, edges = parser(file.path, content)

                for node in nodes:
                    self._code_map.add_node(node, file_id)
                    nodes_added += 1

                for edge in edges:
                    self._code_map.add_edge(edge)
                    edges_added += 1

                files_processed += 1

            except Exception:
                continue

        self._code_map.update_statistics()
        self._save_toon()

        return {
            "status": "parsed",
            "mode": "sequential",
            "nodes_added": nodes_added,
            "edges_added": edges_added,
            "files_processed": files_processed,
            "total_nodes": len(self._code_map.nodes),
            "total_edges": len(self._code_map.edges),
        }

    def generate_graph(self, force: bool = False) -> Dict[str, Any]:
        if not self._code_map:
            raise ValueError("Not initialized. Call init() first.")

        self._load_toon()

        if os.path.exists(self.md_path) and not force:
            return {
                "status": "skipped",
                "reason": "code-graph.md exists, use force=True to overwrite",
                "file": self.md_path,
            }

        md_content = self._generate_md()
        with open(self.md_path, "w", encoding="utf-8") as f:
            f.write(md_content)

        return {
            "status": "generated",
            "file": self.md_path,
            "lines": len(md_content.split("\n")),
        }

    def full(self) -> Dict[str, Any]:
        if not self._root_path:
            return {"status": "error", "message": "Not initialized"}

        init_result = self.init(self._root_path)
        scan_result = self.scan()
        parse_result = self.parse()
        graph_result = self.generate_graph()

        return {
            "status": "complete",
            "init": init_result,
            "scan": scan_result,
            "parse": parse_result,
            "graph": graph_result,
            "statistics": self._code_map.statistics if self._code_map else {},
        }

    def status(self) -> Dict[str, Any]:
        toon_exists = os.path.exists(self.toon_path) if self._root_path else False
        md_exists = os.path.exists(self.md_path) if self._root_path else False

        if toon_exists:
            self._load_toon()
            stats = self._code_map.statistics if self._code_map else {}
            return {
                "initialized": True,
                "root_path": self._root_path,
                "folder_exists": os.path.exists(self.folder_path) if self._root_path else False,
                "toon_exists": toon_exists,
                "md_exists": md_exists,
                "files": stats.get("total_files", 0),
                "nodes": stats.get("total_nodes", 0),
                "edges": stats.get("total_edges", 0),
            }
        return {
            "initialized": False,
            "root_path": self._root_path,
            "folder_exists": os.path.exists(self.folder_path) if self._root_path else False,
            "toon_exists": toon_exists,
            "md_exists": md_exists,
        }

    def diff(self) -> Dict[str, Any]:
        toon_exists = os.path.exists(self.toon_path) if self._root_path else False

        if not toon_exists:
            return {
                "status": "no_map",
                "message": "No code map found. Run init() first.",
            }

        self._load_toon()

        if not self._code_map:
            return {
                "status": "error",
                "message": "Could not load code map.",
            }

        stats = self._code_map.statistics
        by_type = stats.get("by_type", {})

        lines = [
            f"diff --git a/.code-map/code-map.toon b/.code-map/code-map.toon",
            f"index {len(self._code_map.files)}..{len(self._code_map.nodes)} 100644",
            f"--- a/.code-map/code-map.toon",
            f"+++ b/.code-map/code-map.toon",
            f"@@ -1 +1,{stats.get('total_files', 0)} @@",
            f" Code Map: {self._code_map.name}",
            f"",
            f"@@ -0,0 +1,{stats.get('total_nodes', 0)} @@",
            f"+Files: {stats.get('total_files', 0)}",
            f"+Classes: {by_type.get('class', 0)}",
            f"+Functions: {by_type.get('function', 0) + by_type.get('method', 0)}",
            f"+Edges: {stats.get('total_edges', 0)}",
            f"",
        ]

        if by_type:
            lines.append("+By Type:")
            for node_type, count in sorted(by_type.items()):
                lines.append(f"+  {node_type}: {count}")

        lines.extend([
            f"",
            f"@@ --END OF DIFF--",
            f"",
            f"Statistics:",
            f"  Files: {stats.get('total_files', 0)}",
            f"  Nodes: {stats.get('total_nodes', 0)}",
            f"  Edges: {stats.get('total_edges', 0)}",
        ])

        return {
            "status": "diff",
            "diff_output": "\n".join(lines),
            "statistics": stats,
            "by_type": by_type,
        }

    def get_diff_text(self) -> str:
        result = self.diff()
        return result.get("diff_output", result.get("message", ""))

    def query_node(self, name: Optional[str] = None, node_id: Optional[str] = None,
                   node_type: Optional[str] = None) -> List[Dict[str, Any]]:
        if not self._code_map:
            raise ValueError("Not initialized. Call init() first.")

        self._load_toon()

        if not self._code_map:
            return []

        if name:
            node = self._code_map.get_node_by_name(name)
            return [node.to_dict()] if node else []
        elif node_id:
            node = self._code_map.nodes.get(node_id)
            return [node.to_dict()] if node else []
        elif node_type:
            return [n.to_dict() for n in self._code_map.get_nodes_by_type(node_type)]
        return []

    def query_files(self, language: Optional[str] = None) -> List[Dict[str, Any]]:
        if not self._code_map:
            raise ValueError("Not initialized. Call init() first.")

        self._load_toon()

        if not self._code_map:
            return []

        files = list(self._code_map.files.values())
        if language:
            files = [f for f in files if f.language == language]

        return [f.to_dict() for f in files]

    def query_edges(self, edge_type: Optional[str] = None, source_id: Optional[str] = None,
                    target_id: Optional[str] = None) -> List[Dict[str, Any]]:
        if not self._code_map:
            raise ValueError("Not initialized. Call init() first.")

        self._load_toon()

        if not self._code_map:
            return []

        edges = list(self._code_map.edges.values())

        if edge_type:
            edges = [e for e in edges if e.edge_type == edge_type]
        if source_id:
            edges = [e for e in edges if e.source_id == source_id]
        if target_id:
            edges = [e for e in edges if e.target_id == target_id]

        return [e.to_dict() for e in edges]

    def get_graph_content(self) -> str:
        if os.path.exists(self.md_path):
            with open(self.md_path, "r", encoding="utf-8") as f:
                return f.read()
        return ""

    def get_statistics(self) -> Dict[str, Any]:
        self._load_toon()
        return self._code_map.statistics if self._code_map else {}

    def _load_toon(self):
        if self._code_map is None and os.path.exists(self.toon_path):
            with open(self.toon_path, "r", encoding="utf-8") as f:
                content = f.read()
            data = from_toon(content)
            self._code_map = CodeMap.from_dict(data)

    def _save_toon(self):
        if self._code_map:
            data = self._code_map.to_dict()
            content = to_toon(data)
            with open(self.toon_path, "w", encoding="utf-8") as f:
                f.write(content)

    def _save_md(self):
        if self._code_map:
            md_content = self._generate_md()
            with open(self.md_path, "w", encoding="utf-8") as f:
                f.write(md_content)

    def _scan_files(self) -> List[str]:
        if not self._root_path:
            return []
        files = []
        for root, _, filenames in os.walk(self._root_path):
            for filename in filenames:
                filepath = os.path.join(root, filename)
                files.append(filepath)
        return files

    def _create_file_record(self, file_path: str, content: str) -> CodeFile:
        rel_path = os.path.relpath(file_path, self._root_path)

        ext = os.path.splitext(file_path)[1].lower()
        lang_map = {
            ".py": "python", ".pyi": "python",
            ".js": "javascript", ".jsx": "javascript",
            ".ts": "typescript", ".tsx": "typescript",
            ".java": "java", ".go": "go", ".rs": "rust",
            ".cpp": "cpp", ".cc": "cpp", ".cxx": "cpp", ".h": "cpp", ".hpp": "cpp",
            ".rb": "ruby", ".erb": "ruby",
            ".php": "php",
        }
        language = lang_map.get(ext, "unknown")
        checksum = hashlib.md5(content.encode()).hexdigest()

        return CodeFile(
            id=checksum[:16],
            path=file_path,
            relative_path=rel_path,
            language=language,
            size_bytes=len(content.encode("utf-8")),
            line_count=len(content.split("\n")),
            last_modified=datetime.now().isoformat(),
            checksum=checksum,
        )

    def _generate_md(self) -> str:
        if not self._code_map:
            return ""

        stats = self._code_map.statistics
        classes = self._code_map.get_class_nodes()
        functions = self._code_map.get_function_nodes()

        lines = [
            f"# Code Map: {self._code_map.name}",
            f"Generated: {datetime.now().isoformat()}",
            "",
            "## Statistics",
            f"Files: {stats.get('total_files', 0)} | Classes: {len(classes)} | Functions: {len(functions)} | Edges: {stats.get('total_edges', 0)}",
            "",
            "## Structure",
        ]

        for node in self._code_map.get_module_nodes():
            lines.append(f"[MOD] {node.name}")

        for cls in classes:
            bases = cls.attributes.get("bases", [])
            base_str = f" : {', '.join(bases)}" if bases else ""
            lines.append(f"[CLS] {cls.name}{base_str}")

        lines.extend(["", "## Relationships"])

        for edge in [e for e in self._code_map.edges.values() if e.edge_type == "inherits"][:20]:
            src = self._code_map.nodes.get(edge.source_id)
            tgt = self._code_map.nodes.get(edge.target_id)
            if src and tgt:
                lines.append(f"{src.name} --inherits--> {tgt.name}")

        lines.extend(["", "## Top Classes"])
        sorted_classes = sorted(classes, key=lambda c: len(c.children_ids), reverse=True)[:10]
        for cls in sorted_classes:
            methods = sum(1 for f in functions if f.parent_id == cls.id)
            lines.append(f"- {cls.name}: {methods} methods, {len(cls.children_ids)} children")

        lines.extend(["", "## Top Functions"])
        sorted_funcs = sorted(functions, key=lambda f: f.complexity, reverse=True)[:10]
        for func in sorted_funcs:
            lines.append(f"- {func.name}: complexity={func.complexity}")

        return "\n".join(lines)

    def _get_canvas_doc(self):
        from src.models.canvas_models import code_map_to_board_document

        if not self._root_path:
            raise ValueError("Not initialized. Call init() first.")

        self._load_toon()

        if not self._code_map:
            raise ValueError("No code map to export.")

        return code_map_to_board_document(self._code_map)

    def export_canvas(self, output_path: Optional[str] = None) -> Dict[str, Any]:
        board_doc = self._get_canvas_doc()

        if output_path is None:
            output_path = os.path.join(self.folder_path, "code-map-board.json")

        with open(output_path, "w", encoding="utf-8") as f:
            f.write(board_doc.to_json())

        return {
            "status": "exported",
            "file": output_path,
            "blocks": len(board_doc.blocks),
            "connectors": len(board_doc.connectors),
            "canvas_width": board_doc.canvas_width,
            "canvas_height": board_doc.canvas_height,
        }

    def export_canvas_files(self, output_dir: Optional[str] = None) -> Dict[str, Any]:
        board_doc = self._get_canvas_doc()
        output_dir = output_dir or self.folder_path

        code_map_board_path = os.path.join(output_dir, "code-map-board.json")
        with open(code_map_board_path, "w", encoding="utf-8") as f:
            f.write(board_doc.to_json())

        code_map = self._code_map
        assert code_map is not None, "Code map should be loaded by _get_canvas_doc"

        topics = [
            {
                "id": cls.id,
                "title": cls.name,
                "summary": cls.docstring or f"Class {cls.name}",
                "depth": 0,
                "chunk_ids": [],
                "board_refs": [f"block_{cls.id}"],
                "type": "topic"
            }
            for cls in code_map.get_class_nodes()
        ]

        concepts = [
            {
                "id": func.id,
                "title": func.name,
                "summary": f"Function in {func.file_id}",
                "prerequisites": [],
                "parts": [],
                "related": [],
                "chunk_ids": [],
                "board_refs": [f"block_{func.id}"],
                "type": "concept"
            }
            for func in code_map.nodes.values() if func.node_type == "function"
        ]

        edges = [
            {
                "id": edge.id,
                "type": edge.edge_type,
                "source": edge.source_id,
                "target": edge.target_id,
                "label": edge.label or "",
            }
            for edge in code_map.edges.values()
        ]

        kg_data = {
            "topics": topics,
            "concepts": concepts,
            "edges": edges,
            "scope_root": None,
            "scope_rules": {}
        }

        kg_path = os.path.join(output_dir, "infinite.kg.json")
        import json
        with open(kg_path, "w", encoding="utf-8") as f:
            json.dump(kg_data, f, indent=2)

        return {
            "status": "exported",
            "code_map_board_file": code_map_board_path,
            "kg_file": kg_path,
            "blocks": len(board_doc.blocks),
            "connectors": len(board_doc.connectors),
            "topics": len(topics),
            "concepts": len(concepts),
            "edges": len(edges),
        }

    def export_canvas_chunks(self, output_dir: Optional[str] = None, chunk_size: float = 1000, overlap: float = 200) -> Dict[str, Any]:
        from src.models.canvas_models import export_chunks

        board_doc = self._get_canvas_doc()
        output_dir = output_dir or self.folder_path

        chunk_result = export_chunks(board_doc, output_dir, chunk_size, overlap)

        code_map = self._code_map
        assert code_map is not None

        topics = [
            {
                "id": cls.id,
                "title": cls.name,
                "summary": cls.docstring or f"Class {cls.name}",
                "depth": 0,
                "chunk_ids": [],
                "board_refs": [f"block_{cls.id}"],
                "type": "topic"
            }
            for cls in code_map.get_class_nodes()
        ]

        concepts = [
            {
                "id": func.id,
                "title": func.name,
                "summary": f"Function in {func.file_id}",
                "prerequisites": [],
                "parts": [],
                "related": [],
                "chunk_ids": [],
                "board_refs": [f"block_{func.id}"],
                "type": "concept"
            }
            for func in code_map.nodes.values() if func.node_type == "function"
        ]

        edges = [
            {
                "id": edge.id,
                "type": edge.edge_type,
                "source": edge.source_id,
                "target": edge.target_id,
                "label": edge.label or "",
            }
            for edge in code_map.edges.values()
        ]

        kg_data = {
            "topics": topics,
            "concepts": concepts,
            "edges": edges,
            "scope_root": None,
            "scope_rules": {}
        }

        kg_path = os.path.join(output_dir, "infinite.kg.json")
        with open(kg_path, "w", encoding="utf-8") as f:
            json.dump(kg_data, f, indent=2)

        return {
            "status": "exported",
            "mode": "chunked",
            "code_map_board_file": os.path.join(output_dir, "code-map-board.json"),
            "kg_file": kg_path,
            "chunks_dir": chunk_result["chunks_dir"],
            "total_chunks": chunk_result["total_chunks"],
            "index_file": chunk_result["index_file"],
            "blocks": len(board_doc.blocks),
            "connectors": len(board_doc.connectors),
            "topics": len(topics),
            "concepts": len(concepts),
            "edges": len(edges),
            "canvas_width": board_doc.canvas_width,
            "canvas_height": board_doc.canvas_height,
        }

    def get_canvas_data(self) -> Dict[str, Any]:
        board_doc = self._get_canvas_doc()
        return board_doc.to_dict()

    def get_block_count(self) -> int:
        data = self.get_canvas_data()
        return len(data.get("blocks", []))

    def get_connector_count(self) -> int:
        data = self.get_canvas_data()
        return len(data.get("connectors", []))

    def canvas(self, port: int = 8080) -> Dict[str, Any]:
        import subprocess
        import webbrowser
        import threading
        import time

        if not self._root_path:
            return {
                "status": "error",
                "message": "Not initialized. Call init() first.",
            }

        canvas_board_path = os.path.join(self.folder_path, "code-map-board.json")

        if not os.path.exists(canvas_board_path):
            self.export_canvas_files()

        if not os.path.exists(canvas_board_path):
            return {
                "status": "error",
                "message": "Canvas data not available. Run scan() and parse() first.",
            }

        script_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
        serve_script = os.path.join(script_dir, "serve_canvas.py")

        if not os.path.exists(serve_script):
            return {
                "status": "error",
                "message": "serve_canvas.py not found.",
            }

        def run_server():
            try:
                subprocess.run([
                    sys.executable, serve_script,
                    "--port", str(port),
                    "--dir", self.folder_path
                ], check=True)
            except subprocess.CalledProcessError as e:
                print(f"Canvas server error: {e}")

        threading.Thread(target=run_server, daemon=True).start()

        url = f"http://localhost:{port}/infinite-canvas.html"

        def open_browser():
            time.sleep(1.5)
            webbrowser.open(url)

        threading.Thread(target=open_browser, daemon=True).start()

        return {
            "status": "started",
            "url": url,
            "port": port,
            "canvas_file": canvas_board_path,
            "message": f"Canvas server starting at {url}",
        }

    def extract_state_machines(self) -> Dict[str, Any]:
        if not self._code_map:
            return {
                "status": "error",
                "message": "Not initialized. Run scan() and parse() first.",
            }

        from src.parser.python import detect_state_machines

        state_machines = []
        for file_id, file in self._code_map.files.items():
            if file.language == "python":
                try:
                    with open(file.path, "r", encoding="utf-8") as f:
                        content = f.read()
                    machines = detect_state_machines(content)
                    for sm in machines:
                        state_machines.append({
                            "name": sm.name,
                            "class_name": sm.class_name,
                            "states": sm.states,
                            "transitions": sm.transitions,
                            "file": file.relative_path,
                        })
                except Exception:
                    continue

        return {
            "status": "success",
            "state_machines": state_machines,
            "count": len(state_machines),
        }

    def detect_frameworks(self) -> Dict[str, Any]:
        if not self._root_path:
            return {
                "status": "error",
                "message": "Not initialized. Call init() first.",
            }

        from src.detector.framework import detect_frameworks

        result = detect_frameworks(self._root_path)

        return {
            "status": "success",
            "frameworks": [
                {
                    "name": f.name,
                    "category": f.category,
                    "confidence": f.confidence,
                    "indicators": f.indicators,
                }
                for f in result.frameworks
            ],
            "build_tools": result.build_tools,
            "testing_frameworks": result.testing_frameworks,
            "total_frameworks": len(result.frameworks),
        }

    def search(self, pattern: str, file_types: Optional[List[str]] = None,
               context_lines: int = 2) -> Dict[str, Any]:
        import subprocess
        import re

        if not self._root_path:
            return {
                "status": "error",
                "message": "Not initialized. Call init() first.",
            }

        ext_map = {
            "python": [".py", ".pyi"],
            "javascript": [".js", ".jsx"],
            "typescript": [".ts", ".tsx"],
            "java": [".java"],
            "go": [".go"],
            "rust": [".rs"],
        }

        if file_types:
            file_extensions = [ext for ft in file_types for ext in ext_map.get(ft, [f".{ft}"])]
        else:
            file_extensions = [".py", ".js", ".ts", ".jsx", ".tsx", ".java", ".go", ".rs"]

        glob_patterns = " ".join(f"-g '*{ext}'" for ext in file_extensions)

        try:
            result = subprocess.run(
                f"cd '{self._root_path}' && rg --vimgrep -C {context_lines} {glob_patterns} '{pattern}' 2>/dev/null || true",
                shell=True,
                capture_output=True,
                text=True,
                timeout=30
            )

            matches = []
            for line in result.stdout.strip().split("\n"):
                if not line.strip():
                    continue
                parts = line.split(":")
                if len(parts) >= 3:
                    file_path = parts[0]
                    line_num = int(parts[1]) if parts[1].isdigit() else 1
                    col = int(parts[2]) if len(parts) > 2 and parts[2].isdigit() else 0
                    content = ":".join(parts[3:]) if len(parts) > 3 else ""

                    rel_path = os.path.relpath(file_path, self._root_path)
                    file = self._code_map.get_file_by_path(file_path) if self._code_map else None

                    matches.append({
                        "file": rel_path,
                        "file_id": file.id if file else None,
                        "line": line_num,
                        "column": col,
                        "content": content.strip(),
                    })

            return {
                "status": "success",
                "pattern": pattern,
                "count": len(matches),
                "matches": matches[:100],
            }

        except subprocess.TimeoutExpired:
            return {"status": "error", "message": "Search timed out"}
        except Exception as e:
            return {"status": "error", "message": str(e)}

    def generate_call_graph(self, focus_function: Optional[str] = None) -> Dict[str, Any]:
        if not self._code_map:
            return {"status": "error", "message": "Not initialized. Run scan() and parse() first."}

        if not self._root_path:
            return {"status": "error", "message": "Root path not set."}

        from src.models.canvas_models import generate_call_graph, BoardDocument

        doc = generate_call_graph(self._code_map, focus_function)

        output_path = os.path.join(self._root_path, ".code-map", "call-graph.json")
        os.makedirs(os.path.dirname(output_path), exist_ok=True)

        with open(output_path, "w", encoding="utf-8") as f:
            f.write(doc.to_json())

        return {
            "status": "generated",
            "type": "call_graph",
            "file": output_path,
            "blocks": len(doc.blocks),
            "connectors": len(doc.connectors),
            "canvas_width": doc.canvas_width,
            "canvas_height": doc.canvas_height,
        }

    def generate_dependency_graph(self) -> Dict[str, Any]:
        if not self._code_map:
            return {"status": "error", "message": "Not initialized. Run scan() and parse() first."}

        if not self._root_path:
            return {"status": "error", "message": "Root path not set."}

        from src.models.canvas_models import generate_dependency_graph

        doc = generate_dependency_graph(self._code_map)

        output_path = os.path.join(self._root_path, ".code-map", "dependency-graph.json")
        os.makedirs(os.path.dirname(output_path), exist_ok=True)

        with open(output_path, "w", encoding="utf-8") as f:
            f.write(doc.to_json())

        return {
            "status": "generated",
            "type": "dependency_graph",
            "file": output_path,
            "blocks": len(doc.blocks),
            "connectors": len(doc.connectors),
            "canvas_width": doc.canvas_width,
            "canvas_height": doc.canvas_height,
        }

    def generate_class_hierarchy(self) -> Dict[str, Any]:
        if not self._code_map:
            return {"status": "error", "message": "Not initialized. Run scan() and parse() first."}

        if not self._root_path:
            return {"status": "error", "message": "Root path not set."}

        from src.models.canvas_models import generate_class_hierarchy

        doc = generate_class_hierarchy(self._code_map)

        output_path = os.path.join(self._root_path, ".code-map", "class-hierarchy.json")
        os.makedirs(os.path.dirname(output_path), exist_ok=True)

        with open(output_path, "w", encoding="utf-8") as f:
            f.write(doc.to_json())

        return {
            "status": "generated",
            "type": "class_hierarchy",
            "file": output_path,
            "blocks": len(doc.blocks),
            "connectors": len(doc.connectors),
            "canvas_width": doc.canvas_width,
            "canvas_height": doc.canvas_height,
        }

    def generate_state_machine_diagrams(self) -> Dict[str, Any]:
        if not self._code_map:
            return {"status": "error", "message": "Not initialized. Run scan() and parse() first."}

        if not self._root_path:
            return {"status": "error", "message": "Root path not set."}

        from src.parser.python import detect_state_machines
        from src.models.canvas_models import generate_state_machine_diagram

        state_machines = []
        for file_id, file in self._code_map.files.items():
            if file.language == "python":
                try:
                    with open(file.path, "r", encoding="utf-8") as f:
                        content = f.read()
                    machines = detect_state_machines(content)
                    for sm in machines:
                        state_machines.append({
                            "name": sm.name,
                            "states": sm.states,
                            "transitions": sm.transitions,
                            "file": file.relative_path,
                        })
                except Exception:
                    continue

        if not state_machines:
            return {
                "status": "success",
                "message": "No state machines detected",
                "count": 0,
            }

        doc = generate_state_machine_diagram(state_machines)

        output_path = os.path.join(self._root_path, ".code-map", "state-machines.json")
        os.makedirs(os.path.dirname(output_path), exist_ok=True)

        with open(output_path, "w", encoding="utf-8") as f:
            f.write(doc.to_json())

        return {
            "status": "generated",
            "type": "state_machine",
            "file": output_path,
            "state_machines": len(state_machines),
            "blocks": len(doc.blocks),
            "connectors": len(doc.connectors),
            "canvas_width": doc.canvas_width,
            "canvas_height": doc.canvas_height,
        }

    def git_status(self) -> Dict[str, Any]:
        """Get git repository status."""
        from src.integration.git_integration import get_git_integration

        if not self._root_path:
            return {
                "status": "error",
                "message": "Not initialized. Call init() first.",
            }

        git = get_git_integration(self._root_path)
        if not git:
            return {
                "status": "not_git_repo",
                "message": "Not a git repository",
            }

        stats = git.get_stats_summary()
        return {
            "status": "success",
            "is_git_repo": True,
            **stats,
        }

    def git_commits(self, limit: int = 100, since: Optional[str] = None) -> Dict[str, Any]:
        """Get recent git commits."""
        from src.integration.git_integration import get_git_integration

        if not self._root_path:
            return {"status": "error", "message": "Not initialized"}

        git = get_git_integration(self._root_path)
        if not git:
            return {"status": "not_git_repo"}

        commits = git.get_commits(limit=limit, since=since)
        return {
            "status": "success",
            "count": len(commits),
            "commits": [
                {
                    "hash": c.hash,
                    "short_hash": c.short_hash,
                    "message": c.message,
                    "author": c.author,
                    "date": c.date,
                    "timestamp": c.timestamp,
                }
                for c in commits
            ],
        }

    def git_file_history(self, file_path: str, limit: int = 20) -> Dict[str, Any]:
        """Get commit history for a specific file."""
        from src.integration.git_integration import get_git_integration

        if not self._root_path:
            return {"status": "error", "message": "Not initialized"}

        git = get_git_integration(self._root_path)
        if not git:
            return {"status": "not_git_repo"}

        full_path = os.path.join(self._root_path, file_path)
        commits = git.get_file_history(full_path, limit=limit)
        return {
            "status": "success",
            "file": file_path,
            "count": len(commits),
            "commits": [
                {
                    "hash": c.hash,
                    "short_hash": c.short_hash,
                    "message": c.message,
                    "author": c.author,
                    "date": c.date,
                }
                for c in commits
            ],
        }

    def git_diff(self, from_commit: str, to_commit: str, 
                 file_path: Optional[str] = None) -> Dict[str, Any]:
        """Get diff between two commits."""
        from src.integration.git_integration import get_git_integration

        if not self._root_path:
            return {"status": "error", "message": "Not initialized"}

        git = get_git_integration(self._root_path)
        if not git:
            return {"status": "not_git_repo"}

        diff = git.get_diff_between(from_commit, to_commit, file_path)
        return {
            "status": "success",
            "from": from_commit,
            "to": to_commit,
            "file": file_path,
            "diff": diff[:50000] if diff else "",
            "truncated": len(diff) > 50000 if diff else False,
        }

    def git_heatmap(self, year: Optional[int] = None) -> Dict[str, Any]:
        """Get commit heatmap data."""
        from src.integration.git_integration import get_git_integration

        if not self._root_path:
            return {"status": "error", "message": "Not initialized"}

        git = get_git_integration(self._root_path)
        if not git:
            return {"status": "not_git_repo"}

        heatmap = git.get_heatmap_data(year)
        return {
            "status": "success",
            "year": year or datetime.now().year,
            "data": heatmap,
            "total_commits": sum(d["count"] for d in heatmap),
        }

    def git_churn(self, days: int = 30) -> Dict[str, Any]:
        """Get code churn statistics."""
        from src.integration.git_integration import get_git_integration

        if not self._root_path:
            return {"status": "error", "message": "Not initialized"}

        git = get_git_integration(self._root_path)
        if not git:
            return {"status": "not_git_repo"}

        churn = git.get_code_churn(days)
        return {
            "status": "success",
            "days": days,
            "authors": churn,
        }

    def export_with_git(self, output_dir: Optional[str] = None) -> Dict[str, Any]:
        """Export complete visualization including git data."""
        from src.integration.git_integration import get_git_integration

        output_dir = output_dir or self.folder_path
        os.makedirs(output_dir, exist_ok=True)

        result = self.export_canvas_chunks(output_dir)

        git = get_git_integration(self._root_path)
        if git:
            git_data_path = os.path.join(output_dir, "git-data.json")
            git.export_git_data(git_data_path)
            result["git_data_file"] = git_data_path

        return result


def create_agent() -> CodeCartographerAgent:
    return CodeCartographerAgent()
