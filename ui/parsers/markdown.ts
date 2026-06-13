import { BaseParser, ParserOptions } from './base'
import { CodeNodeData, CodeEdgeData } from '@models/types'

export class MarkdownParser extends BaseParser {
  constructor(options: ParserOptions = {}) {
    super('markdown', options)
  }

  parse(code: string, filePath: string): CodeNodeData[] {
    const nodes: CodeNodeData[] = []
    const lines = code.split('\n')
    let nodeId = 0

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i]
      const indent = line.search(/\S/)

      // Headings: # Title -> function node by level
      const headingMatch = line.match(/^(#{1,6})\s+(.+)/)
      if (headingMatch) {
        const level = headingMatch[1].length
        const nodeType = level === 1 ? 'module' : 'function'
        const nodeData = this.createNode(`${filePath}:heading:${nodeId++}`, headingMatch[2].trim(), nodeType, filePath, i, indent, i, line.length)
        nodeData.modifiers = [`h${level}`]
        nodes.push(nodeData)
      }

      // Code blocks with language tag
      const codeBlockMatch = line.match(/^```(\w+)/)
      if (codeBlockMatch) {
        const nodeData = this.createNode(`${filePath}:codeblock:${nodeId++}`, codeBlockMatch[1], 'import', filePath, i, indent, i, line.length)
        nodeData.modifiers = ['codeblock']
        nodes.push(nodeData)
      }
    }

    return nodes
  }

  extractEdges(code: string, filePath: string, nodes: CodeNodeData[]): CodeEdgeData[] {
    const edges: CodeEdgeData[] = []

    // Markdown links: [text](url)
    const linkPattern = /\[([^\]]+)\]\(([^)]+)\)/g
    let match
    while ((match = linkPattern.exec(code)) !== null) {
      edges.push({
        id: `edge:${filePath}->link:${match[2]}`,
        source: `file:${filePath}`,
        target: match[2],
        type: 'imports',
      })
    }

    return edges
  }
}
