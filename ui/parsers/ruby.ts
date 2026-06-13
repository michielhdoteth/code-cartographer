import { BaseParser, ParserOptions } from './base'
import { CodeNodeData, CodeEdgeData } from '@models/types'

export class RubyParser extends BaseParser {
  constructor(options: ParserOptions = {}) {
    super('ruby', options)
  }

  parse(code: string, filePath: string): CodeNodeData[] {
    const nodes: CodeNodeData[] = []
    const lines = code.split('\n')
    let nodeId = 0

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i]
      const indent = line.search(/\S/)

      // Class definitions
      const classMatch = line.match(/^\s*class\s+(\w+)(?:\s*<\s*(\w+))?/)
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
        nodes.push(nodeData)
      }

      // Method definitions
      const methodMatch = line.match(/^\s*def\s+(\w+)/)
      if (methodMatch) {
        const nodeData = this.createNode(
          `${filePath}:method:${nodeId++}`,
          methodMatch[1],
          'method',
          filePath,
          i,
          indent,
          i,
          line.length
        )
        nodes.push(nodeData)
      }

      // Module definitions
      const moduleMatch = line.match(/^\s*module\s+(\w+)/)
      if (moduleMatch) {
        const nodeData = this.createNode(
          `${filePath}:module:${nodeId++}`,
          moduleMatch[1],
          'module',
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

    // Require statements
    const requirePattern = /require\s+['"]([^'"]+)['"]/g
    let match
    while ((match = requirePattern.exec(code)) !== null) {
      edges.push({
        id: `edge:${filePath}->require:${match[1]}`,
        source: `file:${filePath}`,
        target: match[1],
        type: 'imports',
      })
    }

    return edges
  }
}
