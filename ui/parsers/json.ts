import { BaseParser, ParserOptions } from './base'
import { CodeNodeData, CodeEdgeData } from '@models/types'

export class JsonParser extends BaseParser {
  constructor(options: ParserOptions = {}) {
    super('json', options)
  }

  parse(code: string, filePath: string): CodeNodeData[] {
    const nodes: CodeNodeData[] = []
    const lines = code.split('\n')
    let nodeId = 0

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i]
      const indent = line.search(/\S/)

      // Top-level keys: "key": (indent === 0 or minimal)
      if (indent <= 2) {
        const keyMatch = line.match(/^\s*"(\w[\w-]*)"\s*:\s*[{[]/)
        if (keyMatch) {
          const nodeData = this.createNode(`${filePath}:key:${nodeId++}`, keyMatch[1], 'module', filePath, i, indent, i, line.length)
          nodes.push(nodeData)
        }
      }

      // Property keys
      const propMatch = line.match(/^\s+"(\w[\w-]*)"\s*:/)
      if (propMatch) {
        const nodeData = this.createNode(`${filePath}:prop:${nodeId++}`, propMatch[1], 'property', filePath, i, indent, i, line.length)
        nodes.push(nodeData)
      }
    }

    return nodes
  }

  extractEdges(code: string, filePath: string, nodes: CodeNodeData[]): CodeEdgeData[] {
    const edges: CodeEdgeData[] = []

    // $ref references
    const refPattern = /"\$ref"\s*:\s*"([^"]+)"/g
    let match
    while ((match = refPattern.exec(code)) !== null) {
      edges.push({
        id: `edge:${filePath}->ref:${match[1]}`,
        source: `file:${filePath}`,
        target: match[1],
        type: 'references',
      })
    }

    return edges
  }
}
