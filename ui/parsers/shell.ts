import { BaseParser, ParserOptions } from './base'
import { CodeNodeData, CodeEdgeData } from '@models/types'

export class ShellParser extends BaseParser {
  constructor(options: ParserOptions = {}) {
    super('shell', options)
  }

  parse(code: string, filePath: string): CodeNodeData[] {
    const nodes: CodeNodeData[] = []
    const lines = code.split('\n')
    let nodeId = 0

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i]
      const indent = line.search(/\S/)

      // Shebang line as module marker
      if (i === 0 && line.startsWith('#!')) {
        const nodeData = this.createNode(`${filePath}:module:${nodeId++}`, filePath.split('/').pop() || filePath, 'module', filePath, i, indent, i, line.length)
        nodes.push(nodeData)
      }

      // function name() { or function name {
      const funcMatch = line.match(/^\s*function\s+(\w+)\s*(?:\(\s*\))?\s*\{?/)
      if (funcMatch) {
        const nodeData = this.createNode(`${filePath}:func:${nodeId++}`, funcMatch[1], 'function', filePath, i, indent, i, line.length)
        nodes.push(nodeData)
        continue
      }

      // name() function definition pattern
      const namedFuncMatch = line.match(/^(\w+)\s*\(\s*\)\s*\{/)
      if (namedFuncMatch && !['if', 'for', 'while', 'case', 'until', 'select'].includes(namedFuncMatch[1])) {
        const nodeData = this.createNode(`${filePath}:func:${nodeId++}`, namedFuncMatch[1], 'function', filePath, i, indent, i, line.length)
        nodes.push(nodeData)
      }
    }

    return nodes
  }

  extractEdges(code: string, filePath: string, nodes: CodeNodeData[]): CodeEdgeData[] {
    const edges: CodeEdgeData[] = []

    // source / . statements
    const sourcePattern = /(?:source|\.\s+)\s+["']?([^"'\s;]+)["']?/g
    let match
    while ((match = sourcePattern.exec(code)) !== null) {
      edges.push({
        id: `edge:${filePath}->source:${match[1]}`,
        source: `file:${filePath}`,
        target: match[1],
        type: 'imports',
      })
    }

    return edges
  }
}
