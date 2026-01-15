"""
ASCII Graph Output for Code Cartographer

Generates ASCII-based representations of code maps for efficient LLM consumption.
Token-efficient alternative to visual graphs or TOON for quick overview.
"""

from typing import Dict, List, Optional, Set, Tuple
from src.models.code_map import CodeMap, CodeNode, CodeEdge


class ASCIIGraphGenerator:
    def __init__(self, code_map: CodeMap):
        self.code_map = code_map
        self.node_width = 40
        self.max_depth = 5

    def generate_overview(self) -> str:
        stats = self.code_map.statistics
        lines = [
            "=" * 60,
            f"CODE MAP: {self.code_map.name}",
            f"Version: {self.code_map.version}",
            "=" * 60,
            "",
            "STATISTICS",
            "-" * 20,
            f"Files:      {stats.get('total_files', 0)}",
            f"Nodes:      {stats.get('total_nodes', 0)}",
            f"Edges:      {stats.get('total_edges', 0)}",
            "",
            "BY TYPE",
            "-" * 20,
        ]

        by_type = stats.get("by_type", {})
        for node_type, count in sorted(by_type.items()):
            lines.append(f"  {node_type:<15} {count:>5}")

        lines.extend(["", "FILES", "-" * 20])
        for file_id, file in self.code_map.files.items():
            node_count = len(file.node_ids)
            lines.append(f"  [{file.language:<8}] {file.relative_path:<40} ({node_count} nodes)")

        return "\n".join(lines)

    def generate_tree(self, root_name: Optional[str] = None, max_depth: int = 5) -> str:
        lines = ["CODE STRUCTURE TREE", "=" * 40, ""]

        if root_name:
            root = self.code_map.get_node_by_name(root_name)
            if root:
                lines.extend(self._render_tree_node(root.id, 0, max_depth, set()))
            else:
                lines.append(f"Node not found: {root_name}")
        else:
            modules = [n for n in self.code_map.nodes.values() if n.node_type == "module"]
            for module in modules:
                lines.extend(self._render_tree_node(module.id, 0, max_depth, set()))

        return "\n".join(lines)

    def _render_tree_node(self, node_id: str, depth: int, max_depth: int, visited: Set[str]) -> List[str]:
        if depth > max_depth or node_id in visited:
            return []

        visited.add(node_id)
        node = self.code_map.nodes.get(node_id)
        if not node:
            return []

        lines = []
        prefix = "  " * depth
        node_type = node.node_type[:3].upper()

        if node.node_type in ("class", "interface"):
            lines.append(f"{prefix}[{node_type}] {node.name}")
            bases = node.attributes.get("bases", [])
            if bases:
                lines.append(f"{prefix}    extends: {', '.join(bases)}")
        elif node.node_type in ("function", "method"):
            ret_type = node.return_type or "->"
            params = ", ".join(p["name"] for p in node.parameters[:3])
            if len(node.parameters) > 3:
                params += ", ..."
            lines.append(f"{prefix}[{node_type}] {node.name}({params}) {ret_type}")
        elif node.node_type == "module":
            lines.append(f"{prefix}[MOD] {node.name}/")
        else:
            lines.append(f"{prefix}[{node_type}] {node.name}")

        children = [e.target_id for e in self.code_map.edges.values()
                    if e.source_id == node_id and e.edge_type == "contains"]
        for child_id in children:
            lines.extend(self._render_tree_node(child_id, depth + 1, max_depth, visited))

        return lines

    def generate_relationships(self, node_name: Optional[str] = None) -> str:
        lines = ["RELATIONSHIPS", "=" * 40, ""]

        if node_name:
            node = self.code_map.get_node_by_name(node_name)
            if node:
                lines.append(f"NODE: {node.qualified_name}")
                lines.append("-" * 40)
                lines.extend(self._render_node_relationships(node))
            else:
                lines.append(f"Node not found: {node_name}")
        else:
            inheritance_edges = [e for e in self.code_map.edges.values() if e.edge_type in ("inherits", "implements")]
            call_edges = [e for e in self.code_map.edges.values() if e.edge_type == "calls"]
            import_edges = [e for e in self.code_map.edges.values() if e.edge_type == "imports"]

            lines.append("INHERITANCE")
            lines.append("-" * 40)
            for edge in inheritance_edges:
                src = self.code_map.nodes.get(edge.source_id)
                tgt = self.code_map.nodes.get(edge.target_id)
                if src and tgt:
                    lines.append(f"  {src.name} --{edge.edge_type}--> {tgt.name}")

            lines.append("")
            lines.append("CALLS (top 20)")
            lines.append("-" * 40)
            for edge in call_edges[:20]:
                src = self.code_map.nodes.get(edge.source_id)
                tgt = self.code_map.nodes.get(edge.target_id)
                if src and tgt:
                    lines.append(f"  {src.name}() -> {tgt.name}()")

            lines.append("")
            lines.append("IMPORTS")
            lines.append("-" * 40)
            for edge in import_edges[:30]:
                src = self.code_map.nodes.get(edge.source_id)
                tgt = self.code_map.nodes.get(edge.target_id)
                if src and tgt:
                    module = edge.attributes.get("module", "")
                    lines.append(f"  {src.name} imports {tgt.name} ({module})")

        return "\n".join(lines)

    def _render_node_relationships(self, node: CodeNode) -> List[str]:
        lines = []

        related = self.code_map.get_related_nodes(node.id)
        ancestors = self.code_map.get_ancestors(node.id)
        descendants = self.code_map.get_descendants(node.id)

        lines.append(f"Type:       {node.node_type}")
        lines.append(f"File:       {node.file_id}")
        lines.append(f"Visibility: {node.visibility}")
        if node.docstring:
            doc = node.docstring[:100] + "..." if len(node.docstring) > 100 else node.docstring
            lines.append(f"Doc:        {doc}")
        lines.append("")

        lines.append(f"RELATED ({len(related)})")
        lines.append("-" * 20)
        for r in related[:10]:
            lines.append(f"  {r.node_type}: {r.name}")

        lines.append(f"ANCESTORS ({len(ancestors)})")
        lines.append("-" * 20)
        for a in ancestors:
            lines.append(f"  {a.node_type}: {a.name}")

        lines.append(f"CHILDREN ({len(descendants)})")
        lines.append("-" * 20)
        for d in descendants[:10]:
            lines.append(f"  {d.node_type}: {d.name}")

        return lines

    def generate_dependency_graph(self, max_nodes: int = 50) -> str:
        lines = ["DEPENDENCY GRAPH", "=" * 40, ""]

        import_edges = [e for e in self.code_map.edges.values() if e.edge_type in ("imports", "depends_on")]
        call_edges = [e for e in self.code_map.edges.values() if e.edge_type == "calls"]

        node_imports: Dict[str, List[str]] = {}
        for edge in import_edges:
            src = self.code_map.nodes.get(edge.source_id)
            if src:
                if src.id not in node_imports:
                    node_imports[src.id] = []
                tgt = self.code_map.nodes.get(edge.target_id)
                if tgt:
                    node_imports[src.id].append(tgt.name)

        sorted_nodes = sorted(node_imports.items(), key=lambda x: len(x[1]), reverse=True)

        for node_id, imports in sorted_nodes[:max_nodes]:
            node = self.code_map.nodes.get(node_id)
            if node:
                imports_str = ", ".join(imports[:5])
                if len(imports) > 5:
                    imports_str += f" ... +{len(imports) - 5} more"
                lines.append(f"{node.name:<30} -> {imports_str}")

        return "\n".join(lines)

    def generate_call_graph(self, max_nodes: int = 30) -> str:
        lines = ["CALL GRAPH", "=" * 40, ""]

        call_edges = [e for e in self.code_map.edges.values() if e.edge_type == "calls"]

        callers: Dict[str, List[str]] = {}
        callees: Dict[str, List[str]] = {}

        for edge in call_edges:
            src = self.code_map.nodes.get(edge.source_id)
            tgt = self.code_map.nodes.get(edge.target_id)
            if src and tgt:
                callers.setdefault(src.id, []).append(tgt.name)
                callees.setdefault(tgt.id, []).append(src.name)

        lines.append("FUNCTIONS CALLING MOST OTHERS")
        lines.append("-" * 40)
        for node_id, called in sorted(callers.items(), key=lambda x: len(x[1]), reverse=True)[:max_nodes]:
            node = self.code_map.nodes.get(node_id)
            if node:
                called_str = ", ".join(called[:5])
                if len(called) > 5:
                    called_str += f" ... +{len(called) - 5}"
                lines.append(f"{node.name:<30} calls: {called_str}")

        return "\n".join(lines)

    def generate_class_hierarchy(self) -> str:
        lines = ["CLASS HIERARCHY", "=" * 40, ""]

        class_nodes = {n.id: n for n in self.code_map.nodes.values() if n.node_type == "class"}

        inheritance_edges = [e for e in self.code_map.edges.values() if e.edge_type in ("inherits", "imherits")]

        child_to_parent: Dict[str, str] = {}
        for edge in inheritance_edges:
            child_to_parent[edge.source_id] = edge.target_id

        roots = [nid for nid in class_nodes.keys() if nid not in child_to_parent]

        def render_class(node_id: str, depth: int, visited: Set[str]) -> List[str]:
            if node_id in visited:
                return [f"{'  ' * depth}  (circular)"]
            visited.add(node_id)

            node = class_nodes.get(node_id)
            if not node:
                return []

            lines = []
            prefix = "  " * depth
            bases = node.attributes.get("bases", [])
            base_str = f" : {', '.join(bases)}" if bases else ""
            lines.append(f"{prefix}{node.name}{base_str}")

            children = [e.source_id for e in inheritance_edges if e.target_id == node_id]
            for child_id in children:
                lines.extend(render_class(child_id, depth + 1, visited))

            return lines

        for root_id in roots:
            lines.extend(render_class(root_id, 0, set()))
            lines.append("")

        return "\n".join(lines)

    def generate_summary_for_llm(self) -> str:
        stats = self.code_map.statistics
        classes = self.code_map.get_class_nodes()
        functions = self.code_map.get_function_nodes()

        lines = [
            f"CODEBASE: {self.code_map.name}",
            f"Files: {stats.get('total_files', 0)} | Classes: {len(classes)} | Functions: {len(functions)}",
            "",
            "TOP CLASSES",
            "-" * 20,
        ]

        for cls in sorted(classes, key=lambda c: len(c.children_ids), reverse=True)[:10]:
            methods = sum(1 for c in functions if c.parent_id == cls.id)
            lines.append(f"  {cls.name}: {methods} methods, {len(cls.children_ids)} children")

        lines.extend(["", "TOP FUNCTIONS", "-" * 20])

        for func in sorted(functions, key=lambda f: f.complexity, reverse=True)[:10]:
            lines.append(f"  {func.name} (complexity: {func.complexity})")

        lines.extend(["", "KEY RELATIONSHIPS", "-" * 20])

        inheritance = [e for e in self.code_map.edges.values() if e.edge_type == "inherits"]
        imports = [e for e in self.code_map.edges.values() if e.edge_type == "imports"]

        lines.append(f"  Inheritance chains: {len(inheritance)}")
        lines.append(f"  Import dependencies: {len(imports)}")

        return "\n".join(lines)

    def generate_full_report(self) -> str:
        parts = [
            self.generate_overview(),
            "",
            self.generate_tree(max_depth=4),
            "",
            self.generate_class_hierarchy(),
            "",
            self.generate_summary_for_llm(),
        ]

        return "\n".join(parts)


def generate_ascii(code_map: CodeMap, style: str = "overview", **kwargs) -> str:
    generator = ASCIIGraphGenerator(code_map)

    styles = {
        "overview": generator.generate_overview,
        "tree": lambda: generator.generate_tree(kwargs.get("root")),
        "relationships": lambda: generator.generate_relationships(kwargs.get("node")),
        "dependencies": generator.generate_dependency_graph,
        "calls": generator.generate_call_graph,
        "hierarchy": generator.generate_class_hierarchy,
        "summary": generator.generate_summary_for_llm,
        "full": generator.generate_full_report,
    }

    func = styles.get(style, styles["overview"])
    return func()
