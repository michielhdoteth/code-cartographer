import { BaseParser, ParserOptions } from './base'
import { CodeNodeData, CodeEdgeData } from '@models/types'

export class SqlParser extends BaseParser {
  constructor(options: ParserOptions = {}) {
    super('sql', options)
  }

  parse(code: string, filePath: string): CodeNodeData[] {
    const nodes: CodeNodeData[] = []
    const lines = code.split('\n')
    let nodeId = 0

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i]
      const indent = line.search(/\S/)

      // CREATE TABLE
      const tableMatch = line.match(/^\s*CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?(\w+)/i)
      if (tableMatch) {
        const nodeData = this.createNode(`${filePath}:table:${nodeId++}`, tableMatch[1], 'class', filePath, i, indent, i, line.length)
        nodeData.modifiers = ['table']
        nodes.push(nodeData)
        continue
      }

      // CREATE VIEW
      const viewMatch = line.match(/^\s*CREATE\s+VIEW\s+(?:IF\s+NOT\s+EXISTS\s+)?(\w+)/i)
      if (viewMatch) {
        const nodeData = this.createNode(`${filePath}:view:${nodeId++}`, viewMatch[1], 'class', filePath, i, indent, i, line.length)
        nodeData.modifiers = ['view']
        nodes.push(nodeData)
        continue
      }

      // CREATE FUNCTION
      const funcMatch = line.match(/^\s*CREATE\s+(?:OR\s+REPLACE\s+)?FUNCTION\s+(\w+)/i)
      if (funcMatch) {
        const nodeData = this.createNode(`${filePath}:func:${nodeId++}`, funcMatch[1], 'function', filePath, i, indent, i, line.length)
        nodes.push(nodeData)
        continue
      }

      // CREATE PROCEDURE
      const procMatch = line.match(/^\s*CREATE\s+(?:OR\s+REPLACE\s+)?PROCEDURE\s+(\w+)/i)
      if (procMatch) {
        const nodeData = this.createNode(`${filePath}:proc:${nodeId++}`, procMatch[1], 'function', filePath, i, indent, i, line.length)
        nodes.push(nodeData)
        continue
      }

      // CREATE INDEX
      const indexMatch = line.match(/^\s*CREATE\s+(?:UNIQUE\s+)?INDEX\s+(?:IF\s+NOT\s+EXISTS\s+)?(\w+)/i)
      if (indexMatch) {
        const nodeData = this.createNode(`${filePath}:index:${nodeId++}`, indexMatch[1], 'property', filePath, i, indent, i, line.length)
        nodes.push(nodeData)
      }
    }

    return nodes
  }

  extractEdges(code: string, filePath: string, nodes: CodeNodeData[]): CodeEdgeData[] {
    const edges: CodeEdgeData[] = []

    // JOIN clauses
    const joinPattern = /(?:JOIN|INNER\s+JOIN|LEFT\s+JOIN|RIGHT\s+JOIN|FULL\s+JOIN)\s+(\w+)/gi
    let match
    while ((match = joinPattern.exec(code)) !== null) {
      edges.push({
        id: `edge:${filePath}->join:${match[1]}`,
        source: `file:${filePath}`,
        target: match[1],
        type: 'depends_on',
      })
    }

    // FROM clauses
    const fromPattern = /FROM\s+(\w+)/gi
    while ((match = fromPattern.exec(code)) !== null) {
      edges.push({
        id: `edge:${filePath}->from:${match[1]}`,
        source: `file:${filePath}`,
        target: match[1],
        type: 'depends_on',
      })
    }

    // REFERENCES
    const refPattern = /REFERENCES\s+(\w+)/gi
    while ((match = refPattern.exec(code)) !== null) {
      edges.push({
        id: `edge:${filePath}->ref:${match[1]}`,
        source: `file:${filePath}`,
        target: match[1],
        type: 'depends_on',
      })
    }

    return edges
  }
}
