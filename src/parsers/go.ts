import { BaseParser, ParserOptions } from './base'
import { CodeNodeData, CodeEdgeData } from '@models/types'

export class GoParser extends BaseParser {
  constructor(options: ParserOptions = {}) {
    super('go', options)
  }

  parse(code: string, filePath: string): CodeNodeData[] {
    const nodes: CodeNodeData[] = []
    const lines = code.split('\n')
    let nodeId = 0

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i]
      const indent = line.search(/\S/)

      // Function definitions
      const funcMatch = line.match(/^\s*func\s+(?:\(([^)]*)\)\s+)?(\w+)\s*\(/)
      if (funcMatch) {
        const nodeData = this.createNode(
          `${filePath}:func:${nodeId++}`,
          funcMatch[2],
          funcMatch[1] ? 'method' : 'function',
          filePath,
          i,
          indent,
          i,
          line.length
        )
        nodes.push(nodeData)
      }

      // Interface definitions
      const interfaceMatch = line.match(/^\s*type\s+(\w+)\s+interface\s*{/)
      if (interfaceMatch) {
        const nodeData = this.createNode(
          `${filePath}:interface:${nodeId++}`,
          interfaceMatch[1],
          'interface',
          filePath,
          i,
          indent,
          i,
          line.length
        )
        nodes.push(nodeData)
      }

      // Struct definitions
      const structMatch = line.match(/^\s*type\s+(\w+)\s+struct\s*{/)
      if (structMatch) {
        const nodeData = this.createNode(
          `${filePath}:struct:${nodeId++}`,
          structMatch[1],
          'struct',
          filePath,
          i,
          indent,
          i,
          line.length
        )
        nodes.push(nodeData)
      }
    }

    return nodes
  }

  extractEdges(code: string, filePath: string, nodes: CodeNodeData[]): CodeEdgeData[] {
    const edges: CodeEdgeData[] = []

    // Import statements
    const importPattern = /import\s+(?:\(\s*)?["']([^"']+)["']/g
    let match
    while ((match = importPattern.exec(code)) !== null) {
      edges.push({
        id: `edge:${filePath}->import:${match[1]}`,
        source: `file:${filePath}`,
        target: match[1],
        type: 'imports',
      })
    }

    return edges
  }
}
