import { BaseParser, ParserOptions } from './base'
import { CodeNodeData, CodeEdgeData } from '@models/types'

export class JavaParser extends BaseParser {
  constructor(options: ParserOptions = {}) {
    super('java', options)
  }

  parse(code: string, filePath: string): CodeNodeData[] {
    const nodes: CodeNodeData[] = []
    const lines = code.split('\n')
    let nodeId = 0

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i]
      const indent = line.search(/\S/)

      // Class definitions
      const classMatch = line.match(/^\s*(?:public\s+)?(?:abstract\s+)?class\s+(\w+)\s*(?:extends\s+(\w+))?\s*(?:implements\s+([^{]+))?{/)
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
        nodeData.isAbstract = line.includes('abstract')
        nodes.push(nodeData)
        continue
      }

      // Interface definitions
      const interfaceMatch = line.match(/^\s*(?:public\s+)?interface\s+(\w+)\s*(?:extends\s+([^{]+))?\s*{/)
      if (interfaceMatch) {
        const nodeData = this.createNode(
          `${filePath}:interface:${nodeId++}`,
          interfaceMatch[1],
          'interface',
          filePath,
          i,
          indent,
          i,
          line.length
        )
        nodes.push(nodeData)
        continue
      }

      // Method definitions
      const methodMatch = line.match(/^\s*(?:public|private|protected)?\s*(?:static\s+)?(?:synchronized\s+)?(?:\w+<[^>]+>\s+|\w+\s+)(\w+)\s*\(([^)]*)\)/)
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
        const params = methodMatch[2].split(',').filter(p => p.trim())
        nodeData.metrics.numberOfParameters = params.length
        nodeData.modifiers = line.includes('public')
          ? ['public']
          : line.includes('private')
            ? ['private']
            : line.includes('protected')
              ? ['protected']
              : []
        if (line.includes('static')) nodeData.modifiers?.push('static')
        nodes.push(nodeData)
        continue
      }

      // Enum definitions
      const enumMatch = line.match(/^\s*(?:public\s+)?enum\s+(\w+)\s*{/)
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
        continue
      }
    }

    return nodes
  }

  extractEdges(code: string, filePath: string, nodes: CodeNodeData[]): CodeEdgeData[] {
    const edges: CodeEdgeData[] = []

    // Import statements
    const importPattern = /import\s+(?:static\s+)?([^;]+);/g
    let match
    while ((match = importPattern.exec(code)) !== null) {
      edges.push({
        id: `edge:${filePath}->import:${match[1]}`,
        source: `file:${filePath}`,
        target: match[1],
        type: 'imports',
      })
    }

    // Class extends
    const extendsPattern = /class\s+(\w+)\s+extends\s+(\w+)/g
    while ((match = extendsPattern.exec(code)) !== null) {
      edges.push({
        id: `edge:${match[1]}->extends:${match[2]}`,
        source: `${filePath}:${match[1]}`,
        target: `${filePath}:${match[2]}`,
        type: 'inherits',
      })
    }

    // Class implements
    const implementsPattern = /class\s+(\w+).*implements\s+([^{]+)/g
    while ((match = implementsPattern.exec(code)) !== null) {
      const interfaces = match[2].split(',').map(i => i.trim())
      for (const iface of interfaces) {
        edges.push({
          id: `edge:${match[1]}->implements:${iface}`,
          source: `${filePath}:${match[1]}`,
          target: `${filePath}:${iface}`,
          type: 'implements',
        })
      }
    }

    return edges
  }
}
