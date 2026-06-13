import { BaseParser, ParserOptions } from './base'
import { CodeNodeData, CodeEdgeData } from '@models/types'

export class HtmlParser extends BaseParser {
  constructor(options: ParserOptions = {}) {
    super('html', options)
  }

  parse(code: string, filePath: string): CodeNodeData[] {
    const nodes: CodeNodeData[] = []
    const lines = code.split('\n')
    let nodeId = 0

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i]
      const indent = line.search(/\S/)

      // Title tag
      const titleMatch = line.match(/<title>([^<]+)<\/title>/i)
      if (titleMatch) {
        const nodeData = this.createNode(`${filePath}:title:${nodeId++}`, titleMatch[1], 'module', filePath, i, indent, i, line.length)
        nodes.push(nodeData)
      }

      // Heading tags as sections
      const headingMatch = line.match(/<h([1-6])[^>]*>([^<]+)<\/h\1>/i)
      if (headingMatch) {
        const nodeData = this.createNode(`${filePath}:heading:${nodeId++}`, headingMatch[2], 'function', filePath, i, indent, i, line.length)
        nodeData.modifiers = [`h${headingMatch[1]}`]
        nodes.push(nodeData)
      }

      // Container tags: div, section, nav, header, footer, main
      const containerMatch = line.match(/<(div|section|nav|header|footer|main)\s+[^>]*id=["']([^"']+)["']/i)
      if (containerMatch) {
        const nodeData = this.createNode(`${filePath}:${containerMatch[1]}:${nodeId++}`, containerMatch[2], 'class', filePath, i, indent, i, line.length)
        nodeData.modifiers = [containerMatch[1]]
        nodes.push(nodeData)
      }

      // Script blocks
      const scriptMatch = line.match(/<script\s+[^>]*src=["']([^"']+)["']/i)
      if (scriptMatch) {
        const nodeData = this.createNode(`${filePath}:script:${nodeId++}`, scriptMatch[1], 'import', filePath, i, indent, i, line.length)
        nodes.push(nodeData)
      }
    }

    return nodes
  }

  extractEdges(code: string, filePath: string, nodes: CodeNodeData[]): CodeEdgeData[] {
    const edges: CodeEdgeData[] = []

    // Link href
    const linkPattern = /<link\s+[^>]*href=["']([^"']+)["']/gi
    let match
    while ((match = linkPattern.exec(code)) !== null) {
      edges.push({
        id: `edge:${filePath}->link:${match[1]}`,
        source: `file:${filePath}`,
        target: match[1],
        type: 'imports',
      })
    }

    // Script src
    const scriptPattern = /<script\s+[^>]*src=["']([^"']+)["']/gi
    while ((match = scriptPattern.exec(code)) !== null) {
      edges.push({
        id: `edge:${filePath}->script:${match[1]}`,
        source: `file:${filePath}`,
        target: match[1],
        type: 'imports',
      })
    }

    // Image src
    const imgPattern = /<img\s+[^>]*src=["']([^"']+)["']/gi
    while ((match = imgPattern.exec(code)) !== null) {
      edges.push({
        id: `edge:${filePath}->img:${match[1]}`,
        source: `file:${filePath}`,
        target: match[1],
        type: 'imports',
      })
    }

    return edges
  }
}
