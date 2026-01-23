import { BaseParser, ParserOptions } from './base'
import { CodeNodeData, CodeEdgeData } from '@models/types'

export class PhpParser extends BaseParser {
  constructor(options: ParserOptions = {}) {
    super('php', options)
  }

  parse(code: string, filePath: string): CodeNodeData[] {
    const nodes: CodeNodeData[] = []
    const lines = code.split('\n')
    let nodeId = 0

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i]
      const indent = line.search(/\S/)

      // Class definitions
      const classMatch = line.match(/^\s*(?:abstract\s+)?(?:final\s+)?class\s+(\w+)(?:\s+extends\s+(\w+))?(?:\s+implements\s+([^{]+))?/)
      if (classMatch) {
        const nodeData = this.createNode(
          `${filePath}:class:${nodeId++}`,
          classMatch[1],
          'class',
          filePath,
          i,
          indent,
          i,
          line.length
        )
        nodeData.isAbstract = line.includes('abstract')
        nodes.push(nodeData)
      }

      // Interface definitions
      const interfaceMatch = line.match(/^\s*interface\s+(\w+)/)
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

      // Trait definitions
      const traitMatch = line.match(/^\s*trait\s+(\w+)/)
      if (traitMatch) {
        const nodeData = this.createNode(
          `${filePath}:trait:${nodeId++}`,
          traitMatch[1],
          'trait',
          filePath,
          i,
          indent,
          i,
          line.length
        )
        nodes.push(nodeData)
      }

      // Function definitions
      const funcMatch = line.match(/^\s*(?:public|private|protected)?\s*(?:static\s+)?(?:function\s+)?(\w+)\s*\(/)
      if (funcMatch && (line.includes('function') || indent === 0)) {
        const nodeData = this.createNode(
          `${filePath}:func:${nodeId++}`,
          funcMatch[1],
          indent > 0 ? 'method' : 'function',
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

    // Use statements (namespaces)
    const usePattern = /use\s+([^;]+);/g
    let match
    while ((match = usePattern.exec(code)) !== null) {
      edges.push({
        id: `edge:${filePath}->use:${match[1]}`,
        source: `file:${filePath}`,
        target: match[1],
        type: 'imports',
      })
    }

    return edges
  }
}
