import { BaseParser, ParserOptions } from './base'
import { CodeNodeData, CodeEdgeData } from '@models/types'

export class RustParser extends BaseParser {
  constructor(options: ParserOptions = {}) {
    super('rust', options)
  }

  parse(code: string, filePath: string): CodeNodeData[] {
    const nodes: CodeNodeData[] = []
    const lines = code.split('\n')
    let nodeId = 0

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i]
      const indent = line.search(/\S/)

      // Function definitions
      const funcMatch = line.match(/^\s*(?:pub\s+)?(?:async\s+)?fn\s+(\w+)\s*\(/)
      if (funcMatch) {
        const nodeData = this.createNode(
          `${filePath}:fn:${nodeId++}`,
          funcMatch[1],
          'function',
          filePath,
          i,
          indent,
          i,
          line.length
        )
        nodes.push(nodeData)
      }

      // Struct definitions
      const structMatch = line.match(/^\s*(?:pub\s+)?struct\s+(\w+)/)
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

      // Trait definitions
      const traitMatch = line.match(/^\s*(?:pub\s+)?trait\s+(\w+)/)
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

      // Enum definitions
      const enumMatch = line.match(/^\s*(?:pub\s+)?enum\s+(\w+)/)
      if (enumMatch) {
        const nodeData = this.createNode(
          `${filePath}:enum:${nodeId++}`,
          enumMatch[1],
          'enum',
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

    // Use statements
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
