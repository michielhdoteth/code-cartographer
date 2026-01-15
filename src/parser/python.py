"""
Python AST Parser for Code Cartographer

Extracts code structure from Python files using the ast module.
Supports functions, classes, methods, imports, and relationships.
"""

import ast
import os
import uuid
from dataclasses import dataclass
from datetime import datetime
from typing import Any, Dict, List, Optional, Tuple

from src.models.code_map import CodeMap, CodeNode, CodeEdge, CodeFile, Position
from src.parser.base import ASTParser, ParseOptions, ParserRegistry


def _get_position(node: ast.AST) -> Position:
    return Position(
        line=node.lineno,
        column=node.col_offset,
        end_line=node.end_lineno if hasattr(node, "end_lineno") else node.lineno,
        end_column=node.end_col_offset if hasattr(node, "end_col_offset") else node.col_offset,
    )


def _get_docstring(node: ast.AST) -> Optional[str]:
    if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef, ast.ClassDef)):
        if node.body and isinstance(node.body[0], ast.Expr):
            if isinstance(node.body[0].value, ast.Constant) and isinstance(node.body[0].value.s, str):
                return node.body[0].value.s
            if isinstance(node.body[0].value, ast.Str):
                return node.body[0].value.s
    return None


def _get_decorators(node: ast.AST) -> List[str]:
    if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef, ast.ClassDef)):
        return [ast.unparse(d) for d in node.decorator_list]
    return []


def _get_visibility(name: str) -> str:
    if name.startswith("__") and not name.endswith("__"):
        return "private"
    if name.startswith("__"):
        return "dunder"
    if name.startswith("_"):
        return "protected"
    return "public"


def _is_static_method(node: ast.FunctionDef) -> bool:
    return any(
        (isinstance(d, ast.Name) and d.id == "staticmethod") or
        (isinstance(d, ast.Attribute) and d.attr == "staticmethod")
        for d in node.decorator_list
    )


def _is_class_method(node: ast.FunctionDef) -> bool:
    return any(
        (isinstance(d, ast.Name) and d.id == "classmethod") or
        (isinstance(d, ast.Attribute) and d.attr == "classmethod")
        for d in node.decorator_list
    )


def _is_abstract(node: ast.AST) -> bool:
    if isinstance(node, ast.FunctionDef):
        for decorator in node.decorator_list:
            if isinstance(decorator, ast.Name) and decorator.id == "abstractmethod":
                return True
            if isinstance(decorator, ast.Attribute) and decorator.attr == "abstractmethod":
                return True
    return False


def _get_type_name(annotation: Optional[Any]) -> Optional[str]:
    if annotation is None:
        return None
    try:
        if isinstance(annotation, ast.Name):
            return annotation.id
        if isinstance(annotation, ast.Attribute):
            parts = []
            current = annotation
            while isinstance(current, ast.Attribute):
                parts.append(current.attr)
                current = current.value
            if isinstance(current, ast.Name):
                parts.append(current.id)
            return ".".join(reversed(parts))
        if isinstance(annotation, ast.Subscript):
            base = _get_type_name(annotation.value)
            index = _get_type_name(annotation.slice)
            if base and index:
                return f"{base}[{index}]"
            return base or index
        return ast.unparse(annotation)
    except:
        return None


def _get_parameters(node: ast.FunctionDef) -> List[Dict]:
    params = []
    
    defaults_start = len(node.args.args) - len(node.args.defaults)
    for i, arg in enumerate(node.args.args):
        param_info = {
            "name": getattr(arg, 'arg', str(arg)),
            "type": _get_type_name(arg.annotation) if arg.annotation else None,
            "default": None,
            "kind": "positional",
        }
        if i >= defaults_start:
            default_idx = i - defaults_start
            default = node.args.defaults[default_idx]
            if isinstance(default, ast.Constant):
                param_info["default"] = str(default.value)
            elif hasattr(default, 's'):
                param_info["default"] = default.s
            else:
                param_info["default"] = "..."
        params.append(param_info)
    if node.args.vararg:
        params.append({
            "name": f"*{getattr(node.args.vararg, 'arg', str(node.args.vararg))}",
            "type": None,
            "default": None,
            "kind": "vararg",
        })
    if node.args.kwarg:
        params.append({
            "name": f"**{getattr(node.args.kwarg, 'arg', str(node.args.kwarg))}",
            "type": None,
            "default": None,
            "kind": "kwarg",
        })
    return params


def _calculate_complexity(node: ast.AST) -> int:
    complexity = 1
    for child in ast.walk(node):
        if isinstance(child, (ast.If, ast.IfExp)):
            complexity += 1
        elif isinstance(child, (ast.For, ast.AsyncFor, ast.While)):
            complexity += 1
        elif isinstance(child, ast.Try):
            for handler in child.handlers:
                complexity += 1
        elif isinstance(child, ast.BoolOp):
            if isinstance(child.op, ast.And):
                complexity += len(child.values) - 1
            elif isinstance(child.op, ast.Or):
                complexity += len(child.values) - 1
        elif isinstance(child, ast.comprehension):
            complexity += 1
    return complexity


class PythonParser(ASTParser):
    LANGUAGE_NAME = "python"
    FILE_EXTENSIONS = [".py", ".pyi"]
    LANGUAGE_ID = "python"

    def __init__(self, options: Optional[ParseOptions] = None):
        super().__init__(options)
        self._current_file_path = ""
        self._imports: List[Dict] = []
        self._pending_edges: List[Tuple[str, str, str]] = []

    def get_tree_sitter_language(self) -> str:
        return "python"

    def parse_file(self, file_path: str, content: str) -> CodeFile:
        self._current_file_path = file_path
        self._imports = []
        self._pending_edges = []

        rel_path = os.path.basename(file_path)
        file_id = str(uuid.uuid4())

        file = CodeFile(
            id=file_id,
            path=file_path,
            relative_path=rel_path,
            language="python",
            size_bytes=len(content.encode("utf-8")),
            line_count=len(content.split("\n")),
            last_modified=datetime.now().isoformat(),
            is_test=file_path.endswith("_test.py") or "test_" in file_path,
        )

        self._current_file_id = file_id
        self._current_file = file
        self._current_qualified_name = ""

        return file

    def extract_nodes(self, file_id: str, content: str, tree: ast.AST) -> List[CodeNode]:
        nodes = []

        module_node = self._create_module_node(file_id, content)
        nodes.append(module_node)

        for child in ast.iter_child_nodes(tree):
            if isinstance(child, ast.ClassDef):
                node = self._create_class_node(child, file_id, content)
                nodes.append(node)
                nodes.extend(self._extract_class_children(child, file_id, content))
            elif isinstance(child, (ast.FunctionDef, ast.AsyncFunctionDef)):
                node = self._create_function_node(child, file_id, content)
                nodes.append(node)
            elif isinstance(child, (ast.Assign, ast.AnnAssign)):
                if isinstance(child.targets[0], ast.Name):
                    var_name = child.targets[0].id
                    if not var_name.startswith("_"):
                        node = self._create_variable_node(var_name, child, file_id, content)
                        nodes.append(node)
            elif isinstance(child, ast.Import):
                for alias in child.names:
                    self._imports.append({
                        "name": alias.name,
                        "asname": alias.asname,
                        "type": "import",
                    })
            elif isinstance(child, ast.ImportFrom):
                module = child.module or ""
                for alias in child.names:
                    self._imports.append({
                        "name": f"{module}.{alias.name}" if module else alias.name,
                        "asname": alias.asname,
                        "type": "from_import",
                    })

        return nodes

    def extract_edges(self, nodes: List[CodeNode], tree: ast.AST) -> List[CodeEdge]:
        edges = []

        for node in nodes:
            if node.parent_id:
                edges.append(CodeEdge(
                    id=str(uuid.uuid4()),
                    edge_type="contains",
                    source_id=node.parent_id,
                    target_id=node.id,
                ))

        for node in nodes:
            if node.node_type in ("class", "function", "method"):
                call_edges = self._extract_calls(node, tree)
                edges.extend(call_edges)

        for node in nodes:
            if node.node_type == "class":
                inheritance_edges = self._extract_inheritance(node, tree)
                edges.extend(inheritance_edges)

        return edges

    def _create_module_node(self, file_id: str, content: str) -> CodeNode:
        module_name = os.path.splitext(os.path.basename(self._current_file_path))[0]
        position = Position(line=1, column=0, end_line=1, end_column=0)

        node = CodeNode(
            id=file_id,
            node_type="module",
            name=module_name,
            qualified_name=module_name,
            file_id=file_id,
            start_position=position,
            docstring=_get_docstring(ast.parse(content)) if self.options.extract_docstrings else None,
        )

        return node

    def _create_class_node(self, node: ast.ClassDef, file_id: str, content: str) -> CodeNode:
        position = _get_position(node)
        qualified_name = self._make_qualified_name(node.name)

        bases = []
        for base in node.bases:
            if isinstance(base, ast.Name):
                bases.append(base.id)
            elif isinstance(base, ast.Attribute):
                bases.append(ast.unparse(base))

        class_node = CodeNode(
            id=self._generate_id("class", node.name),
            node_type="class",
            name=node.name,
            qualified_name=qualified_name,
            file_id=file_id,
            start_position=position,
            end_position=_get_position(node),
            parent_id=self._current_file_id if not self._current_qualified_name else None,
            docstring=_get_docstring(node) if self.options.extract_docstrings else None,
            decorators=_get_decorators(node),
            visibility=_get_visibility(node.name),
            is_abstract=_is_abstract(node),
            attributes={"bases": bases},
        )

        return class_node

    def _create_function_node(self, node: ast.FunctionDef, file_id: str, content: str) -> CodeNode:
        position = _get_position(node)
        qualified_name = self._make_qualified_name(node.name)

        func_node = CodeNode(
            id=self._generate_id("function", node.name),
            node_type="method" if self._current_qualified_name else "function",
            name=node.name,
            qualified_name=qualified_name,
            file_id=file_id,
            start_position=position,
            end_position=_get_position(node),
            parent_id=self._current_scope[-1] if self._current_scope else None,
            docstring=_get_docstring(node) if self.options.extract_docstrings else None,
            decorators=_get_decorators(node),
            visibility=_get_visibility(node.name),
            is_static=_is_static_method(node),
            is_async=isinstance(node, ast.AsyncFunctionDef),
            return_type=_get_type_name(node.returns),
            parameters=_get_parameters(node),
            complexity=_calculate_complexity(node),
        )

        return func_node

    def _create_variable_node(self, name: str, node: ast.Assign, file_id: str, content: str) -> CodeNode:
        position = _get_position(node)

        var_type = None
        if isinstance(node, ast.AnnAssign) and node.annotation:
            var_type = _get_type_name(node.annotation)

        return CodeNode(
            id=self._generate_id("variable", name),
            node_type="constant" if name.isupper() else "attribute",
            name=name,
            qualified_name=self._make_qualified_name(name),
            file_id=file_id,
            start_position=position,
            parent_id=self._current_scope[-1] if self._current_scope else None,
            visibility=_get_visibility(name),
            return_type=var_type,
        )

    def _extract_class_children(self, class_node: ast.ClassDef, file_id: str, content: str) -> List[CodeNode]:
        nodes = []
        self._push_scope(class_node.name)

        for child in ast.iter_child_nodes(class_node):
            if isinstance(child, ast.FunctionDef):
                node = self._create_function_node(child, file_id, content)
                nodes.append(node)
            elif isinstance(child, (ast.FunctionDef,)):
                for nested in ast.iter_child_nodes(child):
                    if isinstance(nested, ast.ClassDef):
                        nested_node = self._create_class_node(nested, file_id, content)
                        nodes.append(nested_node)
                        nodes.extend(self._extract_class_children(nested, file_id, content))
            elif isinstance(child, ast.Assign):
                if isinstance(child.targets[0], ast.Name):
                    var_name = child.targets[0].id
                    if not var_name.startswith("_"):
                        node = self._create_variable_node(var_name, child, file_id, content)
                        nodes.append(node)

        self._pop_scope()
        return nodes

    def _extract_calls(self, node: CodeNode, tree: ast.AST) -> List[CodeEdge]:
        return []

    def _extract_inheritance(self, node: CodeNode, tree: ast.AST) -> List[CodeEdge]:
        return []


ParserRegistry.register(PythonParser)


@dataclass
class StateInfo:
    name: str
    states: List[str]
    transitions: List[Tuple[str, str, str]]
    variable_name: Optional[str] = None
    class_name: Optional[str] = None
    pattern_type: str = "unknown"


def detect_state_machines(content: str) -> List[StateInfo]:
    state_machines: List[StateInfo] = []

    try:
        tree = ast.parse(content)
    except:
        return state_machines

    lines = content.split("\n")

    for node in ast.walk(tree):
        if isinstance(node, ast.ClassDef):
            sm = _detect_enum_states(node, tree)
            if sm:
                state_machines.append(sm)

            sm = _detect_class_methods(node, tree)
            if sm:
                state_machines.append(sm)

    sm = _detect_state_variable(tree, lines)
    if sm:
        state_machines.append(sm)

    sm = _detect_match_case(tree, lines)
    if sm:
        state_machines.append(sm)

    sm = _detect_conditional_chain(tree, lines)
    if sm:
        state_machines.append(sm)

    return state_machines


def _detect_enum_states(class_node: ast.ClassDef, tree: ast.AST) -> Optional[StateInfo]:
    bases = [ast.unparse(b) for b in class_node.bases]
    if not any("Enum" in b for b in bases):
        return None

    members = []
    for child in ast.iter_child_nodes(class_node):
        if isinstance(child, ast.Assign):
            for target in child.targets:
                if isinstance(target, ast.Name):
                    members.append(target.id)
        elif isinstance(child, ast.FunctionDef):
            if child.name in ("__members__", "__ignore__"):
                continue
            members.append(child.name)

    if not members:
        return None

    return StateInfo(
        name=f"EnumStateMachine:{class_node.name}",
        states=members,
        transitions=[],
        class_name=class_node.name,
        pattern_type="enum",
    )


def _detect_class_methods(class_node: ast.ClassDef, tree: ast.AST) -> Optional[StateInfo]:
    state_methods = []
    transitions = []

    for child in ast.iter_child_nodes(class_node):
        if isinstance(child, ast.FunctionDef):
            method_name = child.name
            if method_name.lower().startswith("on_") or method_name.lower().startswith("handle_"):
                state = method_name
                if method_name.lower().startswith("on_"):
                    state = method_name[3:]
                elif method_name.lower().startswith("handle_"):
                    state = method_name[7:]
                state_methods.append(state)

                trans = _extract_transitions_from_method(child, state)
                transitions.extend(trans)

    if not state_methods:
        return None

    return StateInfo(
        name=f"StateMachine:{class_node.name}",
        states=state_methods,
        transitions=transitions,
        class_name=class_node.name,
        pattern_type="state_methods",
    )


def _extract_transitions_from_method(func_node: ast.FunctionDef, from_state: str) -> List[Tuple[str, str, str]]:
    transitions = []

    for node in ast.walk(func_node):
        if isinstance(node, ast.Assign):
            for target in node.targets:
                if isinstance(target, ast.Attribute):
                    if hasattr(target.value, "id") and target.value.id == "self":
                        if target.attr == "state":
                            if isinstance(node.value, ast.Constant):
                                to_state = node.value.value
                                if isinstance(to_state, str):
                                    transitions.append((from_state, "transition", to_state))

    return transitions


def _detect_state_variable(tree: ast.AST, lines: List[str]) -> Optional[StateInfo]:
    state_var = None
    states = set()

    for node in ast.walk(tree):
        if isinstance(node, ast.ClassDef):
            for child in ast.iter_child_nodes(node):
                if isinstance(child, ast.Assign):
                    for target in child.targets:
                        if isinstance(target, ast.Attribute):
                            if isinstance(target.value, ast.Name) and target.value.id == "self":
                                if target.attr == "state":
                                    if isinstance(child.value, ast.Constant):
                                        state_val = child.value.value
                                        if isinstance(state_val, str):
                                            states.add(state_val)
                                            if state_var is None:
                                                state_var = f"self.state"

    if not state_var or not states:
        return None

    return StateInfo(
        name=f"StateVariable:{state_var}",
        states=list(states),
        transitions=[],
        variable_name=state_var,
        pattern_type="state_variable",
    )


def _detect_match_case(tree: ast.AST, lines: List[str]) -> Optional[StateInfo]:
    states = set()
    transitions = []

    for node in ast.walk(tree):
        if isinstance(node, ast.Match):
            match_subject = ast.unparse(node.subject) if hasattr(node, 'subject') else "value"
            for case in node.cases:
                if isinstance(case.pattern, ast.MatchValue):
                    if isinstance(case.pattern.value, ast.Constant):
                        state = case.pattern.value.value
                        if isinstance(state, str):
                            states.add(state)
                            transitions.append((match_subject, "match", state))

    if not states:
        return None

    return StateInfo(
        name=f"MatchCaseStateMachine",
        states=list(states),
        transitions=transitions,
        pattern_type="match_case",
    )


def _detect_conditional_chain(tree: ast.AST, lines: List[str]) -> Optional[StateInfo]:
    states = set()
    state_vars = set()

    for node in ast.walk(tree):
        if isinstance(node, ast.If):
            if_state = None
            elif_states = []

            if isinstance(node.test, ast.Compare):
                if isinstance(node.test.left, ast.Name):
                    state_vars.add(node.test.left.id)
                    if isinstance(node.test.comparators[0], ast.Constant):
                        if isinstance(node.test.comparators[0].value, str):
                            if_state = node.test.comparators[0].value
                            states.add(if_state)

            for orelse in node.orelse:
                if isinstance(orelse, ast.If):
                    if isinstance(orelse.test, ast.Compare):
                        if isinstance(orelse.test.left, ast.Name):
                            state_vars.add(orelse.test.left.id)
                            if isinstance(orelse.test.comparators[0], ast.Constant):
                                if isinstance(orelse.test.comparators[0].value, str):
                                    elif_states.append(orelse.test.comparators[0].value)
                                    states.add(orelse.test.comparators[0].value)

    if not states:
        return None

    return StateInfo(
        name=f"ConditionalChainStateMachine",
        states=list(states),
        transitions=[],
        variable_name=list(state_vars)[0] if state_vars else None,
        pattern_type="conditional_chain",
    )


def parse_python_code(file_path: str, content: str, options: Optional[ParseOptions] = None) -> Tuple[CodeFile, List[CodeNode], List[CodeEdge]]:
    parser = PythonParser(options)
    file = parser.parse_file(file_path, content)
    tree = ast.parse(content)
    nodes = parser.extract_nodes(file.id, content, tree)
    edges = parser.extract_edges(nodes, tree)
    return file, nodes, edges
