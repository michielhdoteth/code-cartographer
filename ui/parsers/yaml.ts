import { BaseParser, ParserOptions } from './base'
import { CodeNodeData, CodeEdgeData } from '@models/types'

export class YamlParser extends BaseParser {
  constructor(options: ParserOptions = {}) {
    super('yaml', options)
  }

  parse(code: string, filePath: string): CodeNodeData[] {
    const nodes: CodeNodeData[] = []
    const lines = code.split('\n')
    let nodeId = 0

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i]
      const indent = line.search(/\S/)

      // Top-level keys: key: (indent === 0)
      if (indent === 0) {
        const topKeyMatch = line.match(/^(\w[\w-]*)\s*:/)
        if (topKeyMatch) {
          const nodeData = this.createNode(`${filePath}:key:${nodeId++}`, topKeyMatch[1], 'module', filePath, i, indent, i, line.length)
          nodes.push(nodeData)
        }
      }

      // Nested keys (indent > 0)
      if (indent > 0) {
        const nestedKeyMatch = line.match(/^(\w[\w-]*)\s*:/)
        if (nestedKeyMatch) {
          const nodeData = this.createNode(`${filePath}:prop:${nodeId++}`, nestedKeyMatch[1], 'property', filePath, i, indent, i, line.length)
          nodes.push(nodeData)
        }
      }
    }

    return nodes
  }

  extractEdges(code: string, filePath: string, nodes: CodeNodeData[]): CodeEdgeData[] {
    const edges: CodeEdgeData[] = []

    // $ref references
    const refPattern = /\$ref\s*:\s*["']?([^"'\s]+)["']?/g
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
