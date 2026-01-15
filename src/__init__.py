"""
Code Cartographer - Claude Code Agent Tool

A tool for mapping codebases with AST parsing and TOON-formatted output.
Designed for Claude Code agents to easily identify code structure and connections.

Usage by Claude Code agents:
    from src.agent import CodeCartographerAgent
    
    agent = CodeCartographerAgent()
    agent.init("/path/to/project")           # Creates .code-map/ folder
    agent.scan()                             # Scans files → code-map.toon
    agent.parse()                            # Extract nodes/edges
    agent.generate_graph()                   # → code-graph.md
    agent.full()                             # All-in-one
    
    # Query the map
    agent.query_node("ClassName")
    agent.query_edges("calls")
    agent.diff()                             # Git-style diff
"""

__version__ = "1.0.0"
__author__ = "Code Cartographer"

from src.models import CodeMap, CodeNode, CodeEdge, CodeFile
from src.agent.api import CodeCartographerAgent, create_agent
from src.formatter import to_toon, from_toon, generate_ascii

__all__ = [
    "CodeMap",
    "CodeNode",
    "CodeEdge",
    "CodeFile",
    "CodeCartographerAgent",
    "create_agent",
    "to_toon",
    "from_toon",
    "generate_ascii",
    "__version__",
]
