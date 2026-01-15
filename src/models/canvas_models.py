"""
Canvas Models for Code Cartographer

Visualization models for the infinite canvas.
Used to render code maps as interactive visualizations.
"""

import json
import os
from dataclasses import dataclass, field, asdict
from typing import List, Dict, Tuple, Optional, Any
from datetime import datetime


@dataclass
class Block:
    id: str
    type: str
    x: float
    y: float
    width: float
    height: float
    content: str
    node_type: str = "concept"
    node_id: Optional[str] = None
    chunk_id: Optional[str] = None
    render_priority: int = 5

    def intersects_bounds(self, x_min: float, x_max: float, y_min: float, y_max: float) -> bool:
        return not (
            self.x + self.width < x_min or
            self.x > x_max or
            self.y + self.height < y_min or
            self.y > y_max
        )

    def get_center(self) -> Tuple[float, float]:
        return (self.x + self.width / 2, self.y + self.height / 2)

    def to_dict(self) -> Dict[str, Any]:
        return asdict(self)

    @classmethod
    def from_dict(cls, data: Dict) -> 'Block':
        return cls(**data)


@dataclass
class Connector:
    id: str
    from_id: str
    to_id: str
    label: str = ""
    edge_type: str = "arrow"
    chunk_id: Optional[str] = None

    def to_dict(self) -> Dict[str, Any]:
        return asdict(self)

    @classmethod
    def from_dict(cls, data: Dict) -> 'Connector':
        return cls(**data)


@dataclass
class BoardDocument:
    blocks: List[Block] = field(default_factory=list)
    connectors: List[Connector] = field(default_factory=list)
    groups: List[str] = field(default_factory=list)
    meta: Dict[str, str] = field(default_factory=dict)
    canvas_width: float = 10000.0
    canvas_height: float = 10000.0
    version: str = "1.0"
    generated_at: str = field(default_factory=lambda: datetime.now().isoformat())

    def to_json(self) -> str:
        return json.dumps(asdict(self), separators=(',', ':'))

    def to_dict(self) -> Dict[str, Any]:
        return asdict(self)

    @classmethod
    def from_dict(cls, data: Dict) -> 'BoardDocument':
        blocks = [Block(**b) for b in data.get('blocks', [])]
        connectors = [Connector(**c) for c in data.get('connectors', [])]
        return cls(
            blocks=blocks,
            connectors=connectors,
            groups=data.get('groups', []),
            meta=data.get('meta', {}),
            canvas_width=data.get('canvas_width', 10000.0),
            canvas_height=data.get('canvas_height', 10000.0),
            version=data.get('version', '1.0'),
            generated_at=data.get('generated_at', datetime.now().isoformat()),
        )

    def get_block_by_id(self, block_id: str) -> Optional[Block]:
        for b in self.blocks:
            if b.id == block_id:
                return b
        return None

    def get_block_for_node(self, node_id: str) -> Optional[Block]:
        for b in self.blocks:
            if b.node_id == node_id:
                return b
        return None

    def get_blocks_in_bounds(self, x_min: float, x_max: float, y_min: float, y_max: float) -> List[Block]:
        return [b for b in self.blocks if b.intersects_bounds(x_min, x_max, y_min, y_max)]

    def get_connectors_for_blocks(self, block_ids: set) -> List[Connector]:
        return [c for c in self.connectors if c.from_id in block_ids and c.to_id in block_ids]

    def get_blocks_by_chunk(self, chunk_id: str) -> List[Block]:
        return [b for b in self.blocks if b.chunk_id == chunk_id]

    def get_connectors_by_chunk(self, chunk_id: str) -> List[Connector]:
        return [c for c in self.connectors if c.chunk_id == chunk_id]

    def compute_bounds(self) -> Dict[str, float]:
        if not self.blocks:
            return {'x_min': 0, 'x_max': 0, 'y_min': 0, 'y_max': 0}
        x_min = min(b.x for b in self.blocks)
        x_max = max(b.x + b.width for b in self.blocks)
        y_min = min(b.y for b in self.blocks)
        y_max = max(b.y + b.height for b in self.blocks)
        return {'x_min': x_min, 'x_max': x_max, 'y_min': y_min, 'y_max': y_max}

    def add_block(self, block: Block):
        self.blocks.append(block)

    def add_connector(self, connector: Connector):
        self.connectors.append(connector)

    def remove_block(self, block_id: str):
        self.blocks = [b for b in self.blocks if b.id != block_id]
        self.connectors = [c for c in self.connectors if c.from_id != block_id and c.to_id != block_id]

    def remove_connector(self, connector_id: str):
        self.connectors = [c for c in self.connectors if c.id != connector_id]


def create_block_for_node(node_id: str, node_name: str, node_type: str,
                          x: float, y: float, width: float = 200.0, height: float = 80.0) -> Block:
    content = f"{node_type}: {node_name}"
    block_type = "class" if node_type == "class" else "function"

    return Block(
        id=f"block_{node_id}",
        type=block_type,
        x=x,
        y=y,
        width=width,
        height=height,
        content=content,
        node_type=node_type,
        node_id=node_id,
        render_priority=0 if node_type == "class" else 5,
    )


def create_connector_for_edge(source_id: str, target_id: str, edge_type: str) -> Connector:
    return Connector(
        id=f"conn_{source_id}_{target_id}",
        from_id=f"block_{source_id}",
        to_id=f"block_{target_id}",
        edge_type=edge_type,
        label=edge_type,
    )


EDGE_TYPE_MAP = {
    "calls": "arrow",
    "inherits": "dashed",
    "implements": "dashed",
    "imports": "dotted",
    "contains": "nested",
    "references": "arrow",
    "param": "arrow",
}


def code_map_to_board_document(code_map) -> BoardDocument:
    doc = BoardDocument()
    doc.meta = {
        "name": code_map.name,
        "version": code_map.version,
    }

    node_positions: Dict[str, Tuple[float, float]] = {}
    start_x, start_y = 5000.0, 5000.0
    row_height = 180.0
    col_width = 320.0

    classes = code_map.get_class_nodes()
    methods = [n for n in code_map.nodes.values() if n.node_type == "method"]
    functions = [n for n in code_map.nodes.values() if n.node_type == "function"]
    modules = [n for n in code_map.nodes.values() if n.node_type == "module"]

    class_positions: Dict[str, Tuple[float, float]] = {}

    for i, cls in enumerate(classes):
        bx = start_x + (i % 3) * col_width
        by = start_y + (i // 3) * row_height * 2
        class_positions[cls.id] = (bx, by)

        content_lines = [
            f"[CLASS] {cls.name}",
        ]
        bases = cls.attributes.get("bases", [])
        if bases:
            content_lines.append(f"extends: {', '.join(bases)}")
        if cls.docstring:
            content_lines.append("")
            for line in cls.docstring.split("\n")[:3]:
                content_lines.append(line.strip())
        if len(cls.children_ids) > 0:
            content_lines.append(f"  {len(cls.children_ids)} members")

        block = Block(
            id=f"block_{cls.id}",
            type="class",
            x=bx,
            y=by,
            width=280.0,
            height=140.0,
            content="\n".join(content_lines),
            node_type="class",
            node_id=cls.id,
            render_priority=0,
        )
        doc.add_block(block)
        node_positions[cls.id] = (bx, by)

    method_offset = len(classes) * 2
    for i, method in enumerate(methods):
        parent_id = method.parent_id
        if parent_id and parent_id in class_positions:
            px, py = class_positions[parent_id]
            mx, my = px + 300, py + (i % 5) * 100
        else:
            mx = start_x + (method_offset + i) % 5 * col_width
            my = start_y + 800 + (method_offset + i) // 5 * row_height

        content = f"[METH] {method.name}("
        params = ", ".join(p["name"] for p in method.parameters[:3])
        if len(method.parameters) > 3:
            params += ", ..."
        content += f"{params})"
        if method.return_type:
            content += f" -> {method.return_type}"

        block = Block(
            id=f"block_{method.id}",
            type="method",
            x=mx,
            y=my,
            width=260.0,
            height=70.0,
            content=content,
            node_type="method",
            node_id=method.id,
            render_priority=5,
        )
        doc.add_block(block)
        node_positions[method.id] = (mx, my)

    func_offset = method_offset + len(methods)
    for i, func in enumerate(functions):
        fx = start_x + (func_offset + i) % 4 * col_width
        fy = start_y + 1200 + (func_offset + i) // 4 * row_height

        content = f"[FUNC] {func.name}("
        params = ", ".join(p["name"] for p in func.parameters[:3])
        if len(func.parameters) > 3:
            params += ", ..."
        content += f"{params})"
        if func.return_type:
            content += f" -> {func.return_type}"

        block = Block(
            id=f"block_{func.id}",
            type="function",
            x=fx,
            y=fy,
            width=260.0,
            height=70.0,
            content=content,
            node_type="function",
            node_id=func.id,
            render_priority=5,
        )
        doc.add_block(block)
        node_positions[func.id] = (fx, fy)

    id_to_block: Dict[str, str] = {}
    for node_id in node_positions:
        id_to_block[node_id] = f"block_{node_id}"
    for node_id, node in code_map.nodes.items():
        id_to_block[node.name] = f"block_{node_id}"
        id_to_block[node.qualified_name] = f"block_{node_id}" if node_id in node_positions else id_to_block.get(node.name, "")

    for edge in code_map.edges.values():
        source_block = id_to_block.get(edge.source_id)
        target_block = id_to_block.get(edge.target_id)
        if source_block and target_block:
            edge_style = EDGE_TYPE_MAP.get(edge.edge_type, "arrow")
            connector = Connector(
                id=f"conn_{edge.source_id}_{edge.target_id}",
                from_id=source_block,
                to_id=target_block,
                edge_type=edge_style,
                label=edge.edge_type,
            )
            doc.add_connector(connector)

    bounds = doc.compute_bounds()
    if bounds['x_max'] > 0 and bounds['y_max'] > 0:
        center_x = (bounds['x_min'] + bounds['x_max']) / 2
        center_y = (bounds['y_min'] + bounds['y_max']) / 2
        offset_x = 5000.0 - center_x
        offset_y = 5000.0 - center_y

        for block in doc.blocks:
            block.x += offset_x
            block.y += offset_y

        doc.canvas_width = (bounds['x_max'] - bounds['x_min']) + 2000
        doc.canvas_height = (bounds['y_max'] - bounds['y_min']) + 2000

    return doc


def generate_call_graph(code_map, focus_node_id: Optional[str] = None) -> BoardDocument:
    doc = BoardDocument()
    doc.meta = {
        "name": code_map.name,
        "version": code_map.version,
        "type": "call_graph",
    }

    node_positions: Dict[str, Tuple[float, float]] = {}
    start_x, start_y = 5000.0, 5000.0
    node_spacing = 280.0
    level_height = 150.0

    call_edges = [e for e in code_map.edges.values() if e.edge_type == "calls"]
    function_nodes = {n.id: n for n in code_map.nodes.values() if n.node_type in ("function", "method")}

    if focus_node_id and focus_node_id in function_nodes:
        focus_func = function_nodes[focus_node_id]
        call_depths = {focus_node_id: 0}
        queue = [(focus_node_id, 0)]

        while queue:
            node_id, depth = queue.pop(0)
            for edge in call_edges:
                if edge.source_id == node_id and edge.target_id in function_nodes:
                    if edge.target_id not in call_depths:
                        call_depths[edge.target_id] = depth + 1
                        queue.append((edge.target_id, depth + 1))

        relevant_nodes = {n for n in function_nodes.keys() if n in call_depths}
    else:
        call_depths = {}
        for edge in call_edges:
            if edge.source_id in function_nodes:
                call_depths[edge.source_id] = call_depths.get(edge.source_id, 0)
            if edge.target_id in function_nodes:
                call_depths[edge.target_id] = call_depths.get(edge.target_id, 0)

        sorted_by_depth = sorted(call_depths.items(), key=lambda x: x[1], reverse=True)
        relevant_nodes = {n[0] for n in sorted_by_depth}

    nodes_by_depth: Dict[int, List[str]] = {}
    for node_id, depth in call_depths.items():
        if node_id in relevant_nodes:
            nodes_by_depth.setdefault(depth, []).append(node_id)

    max_depth = max(nodes_by_depth.keys()) if nodes_by_depth else 0

    for depth in range(max_depth + 1):
        if depth not in nodes_by_depth:
            continue
        nodes_at_depth = nodes_by_depth[depth]
        total_width = len(nodes_at_depth) * node_spacing
        x_offset = start_x - total_width / 2

        for i, node_id in enumerate(nodes_at_depth):
            if node_id not in function_nodes:
                continue
            node = function_nodes[node_id]
            x = x_offset + i * node_spacing
            y = start_y + depth * level_height

            content = f"{node.name}("
            params = ", ".join(p.get("name", "") for p in node.parameters[:2])
            if len(node.parameters) > 2:
                params += ", ..."
            content += f"{params})"

            block = Block(
                id=f"block_{node_id}",
                type="function",
                x=x,
                y=y,
                width=220.0,
                height=60.0,
                content=content,
                node_type="function",
                node_id=node_id,
                render_priority=5 if depth == 0 else 10,
            )
            doc.add_block(block)
            node_positions[node_id] = (x, y)

    for edge in call_edges:
        if edge.source_id in node_positions and edge.target_id in node_positions:
            connector = Connector(
                id=f"conn_{edge.source_id}_{edge.target_id}",
                from_id=f"block_{edge.source_id}",
                to_id=f"block_{edge.target_id}",
                edge_type="arrow",
                label="calls",
            )
            doc.add_connector(connector)

    bounds = doc.compute_bounds()
    if bounds['x_max'] > 0:
        center_x = (bounds['x_min'] + bounds['x_max']) / 2
        offset_x = 5000.0 - center_x
        for block in doc.blocks:
            block.x += offset_x
        doc.canvas_width = (bounds['x_max'] - bounds['x_min']) + 2000
        doc.canvas_height = (max_depth + 2) * level_height + 1000

    return doc


def generate_dependency_graph(code_map) -> BoardDocument:
    doc = BoardDocument()
    doc.meta = {
        "name": code_map.name,
        "version": code_map.version,
        "type": "dependency_graph",
    }

    node_positions: Dict[str, Tuple[float, float]] = {}
    start_x, start_y = 5000.0, 5000.0
    node_spacing = 300.0
    level_height = 200.0

    import_edges = [e for e in code_map.edges.values() if e.edge_type == "imports"]
    file_nodes = {n.id: n for n in code_map.nodes.values() if n.node_type in ("module", "file")}

    file_by_name: Dict[str, str] = {}
    for node_id, node in file_nodes.items():
        file_by_name[node.name] = node_id

    dependency_levels: Dict[str, int] = {}
    for edge in import_edges:
        if edge.target_id in file_by_name.values():
            dependency_levels[edge.target_id] = dependency_levels.get(edge.target_id, 0)
        if edge.source_id in file_by_name.values():
            dependency_levels[edge.source_id] = max(
                dependency_levels.get(edge.source_id, 0),
                dependency_levels.get(edge.target_id, 0) + 1
            )

    nodes_by_level: Dict[int, List[str]] = {}
    for node_id, level in dependency_levels.items():
        if node_id in file_by_name.values():
            nodes_by_level.setdefault(level, []).append(node_id)

    max_level = max(nodes_by_level.keys()) if nodes_by_level else 0

    for level in range(max_level + 1):
        if level not in nodes_by_level:
            continue
        nodes_at_level = nodes_by_level[level]
        total_width = len(nodes_at_level) * node_spacing
        x_offset = start_x - total_width / 2

        for i, node_id in enumerate(nodes_at_level):
            if node_id not in file_nodes:
                continue
            node = file_nodes[node_id]
            x = x_offset + i * node_spacing
            y = start_y + level * level_height

            short_name = node.name.split("/")[-1].split("\\")[-1]
            if len(short_name) > 20:
                short_name = short_name[:17] + "..."

            block = Block(
                id=f"block_{node_id}",
                type="module",
                x=x,
                y=y,
                width=250.0,
                height=50.0,
                content=f"[FILE] {short_name}",
                node_type="module",
                node_id=node_id,
                render_priority=0,
            )
            doc.add_block(block)
            node_positions[node_id] = (x, y)

    for edge in import_edges:
        if edge.source_id in node_positions and edge.target_id in node_positions:
            connector = Connector(
                id=f"conn_{edge.source_id}_{edge.target_id}",
                from_id=f"block_{edge.source_id}",
                to_id=f"block_{edge.target_id}",
                edge_type="dotted",
                label="imports",
            )
            doc.add_connector(connector)

    bounds = doc.compute_bounds()
    if bounds['x_max'] > 0:
        center_x = (bounds['x_min'] + bounds['x_max']) / 2
        offset_x = 5000.0 - center_x
        for block in doc.blocks:
            block.x += offset_x
        doc.canvas_width = (bounds['x_max'] - bounds['x_min']) + 2000
        doc.canvas_height = (max_level + 2) * level_height + 1000

    return doc


def generate_class_hierarchy(code_map) -> BoardDocument:
    doc = BoardDocument()
    doc.meta = {
        "name": code_map.name,
        "version": code_map.version,
        "type": "class_hierarchy",
    }

    node_positions: Dict[str, Tuple[float, float]] = {}
    start_x, start_y = 5000.0, 5000.0
    node_spacing = 350.0
    level_height = 200.0

    inherit_edges = [e for e in code_map.edges.values() if e.edge_type in ("inherits", "extends")]
    class_nodes = {n.id: n for n in code_map.nodes.values() if n.node_type in ("class", "interface")}

    inheritance_tree: Dict[str, List[str]] = {}
    child_classes: Dict[str, List[str]] = {}

    for edge in inherit_edges:
        if edge.target_id not in inheritance_tree:
            inheritance_tree[edge.target_id] = []
        inheritance_tree[edge.target_id].append(edge.source_id)
        if edge.source_id not in child_classes:
            child_classes[edge.source_id] = []
        child_classes[edge.source_id].append(edge.target_id)

    root_classes = [n for n in class_nodes.keys() if n not in child_classes]

    def calculate_depth(node_id: str, current_depth: int, max_depths: Dict[str, int]):
        max_depths[node_id] = max(max_depths.get(node_id, 0), current_depth)
        for child in inheritance_tree.get(node_id, []):
            calculate_depth(child, current_depth + 1, max_depths)

    max_depths: Dict[str, int] = {}
    for root in root_classes:
        calculate_depth(root, 0, max_depths)

    nodes_by_depth: Dict[int, List[str]] = {}
    for node_id, depth in max_depths.items():
        if node_id in class_nodes:
            nodes_by_depth.setdefault(depth, []).append(node_id)

    for depth, nodes_at_depth in nodes_by_depth.items():
        total_width = len(nodes_at_depth) * node_spacing
        x_offset = start_x - total_width / 2

        for i, node_id in enumerate(nodes_at_depth):
            if node_id not in class_nodes:
                continue
            node = class_nodes[node_id]
            x = x_offset + i * node_spacing
            y = start_y + depth * level_height

            bases = node.attributes.get("bases", [])
            base_str = f" : {', '.join(bases)}" if bases else ""

            content = f"[CLASS] {node.name}{base_str}"

            block = Block(
                id=f"block_{node_id}",
                type="class",
                x=x,
                y=y,
                width=300.0,
                height=80.0,
                content=content,
                node_type="class",
                node_id=node_id,
                render_priority=0,
            )
            doc.add_block(block)
            node_positions[node_id] = (x, y)

    for edge in inherit_edges:
        if edge.source_id in node_positions and edge.target_id in node_positions:
            connector = Connector(
                id=f"conn_{edge.source_id}_{edge.target_id}",
                from_id=f"block_{edge.source_id}",
                to_id=f"block_{edge.target_id}",
                edge_type="dashed",
                label="extends",
            )
            doc.add_connector(connector)

    bounds = doc.compute_bounds()
    if bounds['x_max'] > 0:
        center_x = (bounds['x_min'] + bounds['x_max']) / 2
        offset_x = 5000.0 - center_x
        for block in doc.blocks:
            block.x += offset_x
        doc.canvas_width = (bounds['x_max'] - bounds['x_min']) + 2000
        doc.canvas_height = (max(nodes_by_depth.keys()) + 2) * level_height + 1000

    return doc


def generate_state_machine_diagram(state_machines: List[Dict]) -> BoardDocument:
    doc = BoardDocument()
    doc.meta = {
        "name": "State Machines",
        "version": "1.0",
        "type": "state_machine",
    }

    start_x, start_y = 5000.0, 5000.0
    machine_spacing = 500.0
    state_spacing = 200.0
    state_size = 120.0

    for mi, sm in enumerate(state_machines):
        machine_x = start_x + mi * machine_spacing
        machine_y = start_y

        machine_block = Block(
            id=f"sm_{mi}_name",
            type="state_machine",
            x=machine_x - 150,
            y=machine_y - 80,
            width=300.0,
            height=50.0,
            content=f"[STATE MACHINE] {sm.get('name', f'Machine {mi+1}')}",
            node_type="state_machine",
            node_id=f"sm_{mi}",
            render_priority=0,
        )
        doc.add_block(machine_block)

        states = sm.get("states", [])
        transitions = sm.get("transitions", [])

        state_positions: Dict[str, Tuple[float, float]] = {}

        for si, state in enumerate(states):
            sx = machine_x + (si % 4) * state_spacing - 180
            sy = machine_y + 50 + (si // 4) * state_spacing

            is_initial = si == 0
            block_type = "state_initial" if is_initial else "state"

            content = state
            if is_initial:
                content = "[INIT] " + state

            block = Block(
                id=f"sm_{mi}_state_{si}",
                type=block_type,
                x=sx,
                y=sy,
                width=state_size,
                height=state_size,
                content=content,
                node_type="state",
                node_id=f"sm_{mi}_state_{si}",
                render_priority=5,
            )
            doc.add_block(block)
            state_positions[state] = (sx + state_size/2, sy + state_size/2)

        for trans in transitions:
            from_state = trans[0] if len(trans) > 0 else None
            to_state = trans[2] if len(trans) > 2 else None
            event = trans[1] if len(trans) > 1 else ""

            if from_state in state_positions and to_state in state_positions:
                from_pos = state_positions[from_state]
                to_pos = state_positions[to_state]

                connector = Connector(
                    id=f"sm_{mi}_trans_{from_state}_{to_state}",
                    from_id=f"sm_{mi}_state_{states.index(from_state)}",
                    to_id=f"sm_{mi}_state_{states.index(to_state)}",
                    edge_type="arrow",
                    label=event if event else "transition",
                )
                doc.add_connector(connector)

    bounds = doc.compute_bounds()
    if bounds['x_max'] > 0:
        center_x = (bounds['x_min'] + bounds['x_max']) / 2
        offset_x = 5000.0 - center_x
        center_y = (bounds['y_min'] + bounds['y_max']) / 2
        offset_y = 5000.0 - center_y
        for block in doc.blocks:
            block.x += offset_x
            block.y += offset_y
        doc.canvas_width = (bounds['x_max'] - bounds['x_min']) + 2000
        doc.canvas_height = (bounds['y_max'] - bounds['y_min']) + 2000

    return doc


ChunkSize = 1000
ChunkOverlap = 200


def compute_chunk_bounds(blocks: List[Block], chunk_size: float = ChunkSize, overlap: float = ChunkOverlap) -> List[Dict]:
    if not blocks:
        return []

    min_x = min(b.x for b in blocks)
    max_x = max(b.x + b.width for b in blocks)
    min_y = min(b.y for b in blocks)
    max_y = max(b.y + b.height for b in blocks)

    cols = max(1, int((max_x - min_x) / (chunk_size - overlap))) + 1
    rows = max(1, int((max_y - min_y) / (chunk_size - overlap))) + 1

    chunks = []
    for row in range(rows):
        for col in range(cols):
            cx_min = min_x + col * (chunk_size - overlap)
            cx_max = cx_min + chunk_size
            cy_min = min_y + row * (chunk_size - overlap)
            cy_max = cy_min + chunk_size

            chunk_blocks = [b for b in blocks if b.x >= cx_min - overlap and b.x < cx_max and b.y >= cy_min - overlap and b.y < cy_max]

            if chunk_blocks:
                chunks.append({
                    "chunk_id": f"chunk_{row:03d}_{col:03d}",
                    "bounds": {
                        "x_min": cx_min,
                        "x_max": cx_max,
                        "y_min": cy_min,
                        "y_max": cy_max,
                    },
                    "block_count": len(chunk_blocks),
                })

    return chunks


def assign_blocks_to_chunks(blocks: List[Block], chunks: List[Dict]) -> Dict[str, List[Block]]:
    chunk_map: Dict[str, List[Block]] = {}
    for chunk in chunks:
        chunk_id = chunk["chunk_id"]
        bounds = chunk["bounds"]
        x_min, x_max = bounds["x_min"], bounds["x_max"]
        y_min, y_max = bounds["y_min"], bounds["y_max"]

        for block in blocks:
            bx = block.x
            by = block.y
            if bx >= x_min - ChunkOverlap and bx < x_max and by >= y_min - ChunkOverlap and by < y_max:
                if chunk_id not in chunk_map:
                    chunk_map[chunk_id] = []
                chunk_map[chunk_id].append(block)

    return chunk_map


def get_connectors_for_blocks(connectors: List[Connector], block_ids: set) -> List[Connector]:
    return [c for c in connectors if c.from_id in block_ids and c.to_id in block_ids]


def export_chunks(doc: BoardDocument, output_dir: str, chunk_size: float = ChunkSize, overlap: float = ChunkOverlap) -> Dict[str, Any]:
    os.makedirs(output_dir, exist_ok=True)
    os.makedirs(os.path.join(output_dir, "chunks"), exist_ok=True)

    blocks = doc.blocks
    connectors = doc.connectors

    chunks = compute_chunk_bounds(blocks, chunk_size, overlap)
    chunk_block_map = assign_blocks_to_chunks(blocks, chunks)

    chunk_data_files = []
    for chunk in chunks:
        chunk_id = chunk["chunk_id"]
        chunk_blocks = chunk_block_map.get(chunk_id, [])
        if not chunk_blocks:
            continue

        block_ids = {b.id for b in chunk_blocks}
        chunk_connectors = get_connectors_for_blocks(connectors, block_ids)

        chunk_content = {
            "chunk_id": chunk_id,
            "bounds": chunk["bounds"],
            "blocks": [b.to_dict() for b in chunk_blocks],
            "connectors": [c.to_dict() for c in chunk_connectors],
        }

        chunk_path = os.path.join(output_dir, "chunks", f"{chunk_id}.json")
        with open(chunk_path, "w", encoding="utf-8") as f:
            json.dump(chunk_content, f, separators=(',', ':'))
        chunk_data_files.append(chunk_id)

    index_data = {
        "version": "1.0",
        "chunk_size": chunk_size,
        "overlap": overlap,
        "total_chunks": len(chunks),
        "canvas_width": doc.canvas_width,
        "canvas_height": doc.canvas_height,
        "chunks": [
            {
                "chunk_id": c["chunk_id"],
                "bounds": c["bounds"],
                "block_count": c["block_count"],
            }
            for c in chunks
        ],
    }

    index_path = os.path.join(output_dir, "chunks", "index.json")
    with open(index_path, "w", encoding="utf-8") as f:
        json.dump(index_data, f, indent=2)

    return {
        "status": "exported",
        "chunks_dir": os.path.join(output_dir, "chunks"),
        "total_chunks": len(chunks),
        "index_file": index_path,
    }
