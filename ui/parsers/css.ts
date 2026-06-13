import { BaseParser, ParserOptions } from './base'
import { CodeNodeData, CodeEdgeData } from '@models/types'

export class CssParser extends BaseParser {
  constructor(options: ParserOptions = {}) {
    super('css', options)
  }

  parse(code: string, filePath: string): CodeNodeData[] {
    const nodes: CodeNodeData[] = []
    const lines = code.split('\n')
    let nodeId = 0

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i]
      const indent = line.search(/\S/)

      // Class selectors: .name {
      const classMatch = line.match(/^\s*\.([a-zA-Z_][\w-]*)\s*\{/)
      if (classMatch) {
        const nodeData = this.createNode(`${filePath}:class:${nodeId++}`, classMatch[1], 'class', filePath, i, indent, i, line.length)
        nodes.push(nodeData)
      }

      // ID selectors: #name {
      const idMatch = line.match(/^\s*#([a-zA-Z_][\w-]*)\s*\{/)
      if (idMatch) {
        const nodeData = this.createNode(`${filePath}:id:${nodeId++}`, idMatch[1], 'property', filePath, i, indent, i, line.length)
        nodes.push(nodeData)
      }

      // @media blocks
      const mediaMatch = line.match(/^\s*@media\s+(.+)\s*\{/)
      if (mediaMatch) {
        const nodeData = this.createNode(`${filePath}:media:${nodeId++}`, mediaMatch[1].trim(), 'module', filePath, i, indent, i, line.length)
        nodes.push(nodeData)
      }

      // @keyframes blocks
      const keyframesMatch = line.match(/^\s*@keyframes\s+(\w+)\s*\{/)
      if (keyframesMatch) {
        const nodeData = this.createNode(`${filePath}:keyframes:${nodeId++}`, keyframesMatch[1], 'function', filePath, i, indent, i, line.length)
        nodes.push(nodeData)
      }
    }

    return nodes
  }

  extractEdges(code: string, filePath: string, nodes: CodeNodeData[]): CodeEdgeData[] {
    const edges: CodeEdgeData[] = []

    // @import statements
    const importPattern = /@import\s+(?:url\()?["']([^"')]+)["']\)?/g
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
