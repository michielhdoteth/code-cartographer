import { BaseParser, ParserOptions } from './base'
import { CodeNodeData, CodeEdgeData } from '@models/types'

export class SwiftParser extends BaseParser {
  constructor(options: ParserOptions = {}) {
    super('swift', options)
  }

  parse(code: string, filePath: string): CodeNodeData[] {
    const nodes: CodeNodeData[] = []
    const lines = code.split('\n')
    let nodeId = 0

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i]
      const indent = line.search(/\S/)

      // Function definitions
      const funcMatch = line.match(
        /^\s*(?:(?:public|private|internal|fileprivate|open)\s+)?(?:(?:static|class|override|mutating)\s+)*func\s+(\w+)\s*\(/
      )
      if (funcMatch) {
        const nodeData = this.createNode(
          `${filePath}:func:${nodeId++}`,
          funcMatch[1],
          'function',
          filePath,
          i,
          indent,
          i,
          line.length
        )
        nodeData.modifiers = []
        if (line.includes('public')) nodeData.modifiers.push('public')
        if (line.includes('private')) nodeData.modifiers.push('private')
        if (line.includes('static')) nodeData.modifiers.push('static')
        if (line.includes('override')) nodeData.modifiers.push('override')
        nodes.push(nodeData)
        continue
      }

      // Class definitions
      const classMatch = line.match(/^\s*(?:(?:public|private|internal|open)\s+)?(?:final\s+)?class\s+(\w+)/)
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
        nodes.push(nodeData)
        continue
      }

      // Struct definitions
      const structMatch = line.match(/^\s*(?:(?:public|private|internal)\s+)?struct\s+(\w+)/)
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
        continue
      }

      // Enum definitions
      const enumMatch = line.match(/^\s*(?:(?:public|private|internal)\s+)?(?:case )?enum\s+(\w+)/)
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

      // Protocol definitions
      const protocolMatch = line.match(/^\s*(?:(?:public|private|internal)\s+)?protocol\s+(\w+)/)
      if (protocolMatch) {
        const nodeData = this.createNode(
          `${filePath}:trait:${nodeId++}`,
          protocolMatch[1],
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

      // Extension definitions
      const extensionMatch = line.match(/^\s*(?:(?:public|private|internal)\s+)?extension\s+(\w+)/)
      if (extensionMatch) {
        const nodeData = this.createNode(
          `${filePath}:extension:${nodeId++}`,
          extensionMatch[1],
          'class',
          filePath,
          i,
          indent,
          i,
          line.length
        )
        nodes.push(nodeData)
        continue
      }

      // Global variables: var/let name = ...
      const varMatch = line.match(/^\s*(?:(?:public|private|internal)\s+)?(?:let|var)\s+(\w+)\s*[=:]/)
      if (varMatch && indent === 0) {
        const nodeData = this.createNode(
          `${filePath}:var:${nodeId++}`,
          varMatch[1],
          'variable',
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

    // Import statements
    const importPattern = /import\s+([\w]+)/g
    let match
    while ((match = importPattern.exec(code)) !== null) {
      edges.push({
        id: `edge:${filePath}->import:${match[1]}`,
        source: `file:${filePath}`,
        target: match[1],
        type: 'imports',
      })
    }

    // Class/struct conforms to protocols
    const conformsPattern = /(?:class|struct)\s+(\w+)\s*:\s*([^{]+)/g
    while ((match = conformsPattern.exec(code)) !== null) {
      const protocols = match[2].split(',').map(p => p.trim())
      for (const protocol of protocols) {
        edges.push({
          id: `edge:${match[1]}->implements:${protocol}`,
          source: `${filePath}:${match[1]}`,
          target: `${filePath}:${protocol}`,
          type: 'implements',
        })
      }
    }

    // Class inherits
    const inheritsPattern = /class\s+(\w+)\s*:\s*(\w+)/g
    while ((match = inheritsPattern.exec(code)) !== null) {
      edges.push({
        id: `edge:${match[1]}->inherits:${match[2]}`,
        source: `${filePath}:${match[1]}`,
        target: `${filePath}:${match[2]}`,
        type: 'inherits',
      })
    }

    return edges
  }
}
