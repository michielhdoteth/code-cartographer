import { BaseParser, ParserOptions } from './base'
import { CodeNodeData, CodeEdgeData } from '@models/types'

export class TomlParser extends BaseParser {
  constructor(options: ParserOptions = {}) {
    super('toml', options)
  }

  parse(code: string, filePath: string): CodeNodeData[] {
    const nodes: CodeNodeData[] = []
    const lines = code.split('\n')
    let nodeId = 0

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i]
      const indent = line.search(/\S/)

      // Section headers: [section] or [section.subsection]
      const sectionMatch = line.match(/^\[([^\]]+)\]\s*$/)
      if (sectionMatch) {
        const nodeData = this.createNode(`${filePath}:section:${nodeId++}`, sectionMatch[1], 'module', filePath, i, indent, i, line.length)
        nodes.push(nodeData)
        continue
      }

      // Array of tables: [[section]]
      const arrayTableMatch = line.match(/^\[\[([^\]]+)\]\]\s*$/)
      if (arrayTableMatch) {
        const nodeData = this.createNode(`${filePath}:section:${nodeId++}`, arrayTableMatch[1], 'module', filePath, i, indent, i, line.length)
        nodeData.modifiers = ['array']
        nodes.push(nodeData)
        continue
      }

      // Key = value pairs
      const keyMatch = line.match(/^([\w.-]+)\s*=/)
      if (keyMatch) {
        const nodeData = this.createNode(`${filePath}:key:${nodeId++}`, keyMatch[1], 'property', filePath, i, indent, i, line.length)
        nodes.push(nodeData)
      }
    }

    return nodes
  }

  extractEdges(_code: string, _filePath: string, _nodes: CodeNodeData[]): CodeEdgeData[] {
    // TOML is a config format, typically no code dependencies
    return []
  }
}
