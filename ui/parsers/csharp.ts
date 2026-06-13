import { BaseParser, ParserOptions } from './base'
import { CodeNodeData, CodeEdgeData } from '@models/types'

export class CSharpParser extends BaseParser {
  constructor(options: ParserOptions = {}) {
    super('csharp', options)
  }

  parse(code: string, filePath: string): CodeNodeData[] {
    const nodes: CodeNodeData[] = []
    const lines = code.split('\n')
    let nodeId = 0

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i]
      const indent = line.search(/\S/)

      // Class definitions
      const classMatch = line.match(
        /^\s*(?:public|private|protected|internal)?\s*(?:abstract|sealed|static)?\s*class\s+(\w+)/
      )
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
      const interfaceMatch = line.match(/^\s*(?:public|private|protected|internal)?\s*interface\s+(\w+)/)
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

      // Struct definitions
      const structMatch = line.match(/^\s*(?:public|private|protected|internal)?\s*(?:readonly)?\s*struct\s+(\w+)/)
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
      const enumMatch = line.match(/^\s*(?:public|private|protected|internal)?\s*enum\s+(\w+)/)
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

      // Method definitions
      const methodMatch = line.match(
        /^\s*(?:public|private|protected|internal)?\s*(?:static|virtual|override|abstract|async)?\s*(?:\w+(?:<[^>]+>)?\s+)(\w+)\s*\(/
      )
      if (methodMatch && line.includes('(') && !line.includes('class ') && !line.includes('interface ')) {
        const nodeData = this.createNode(
          `${filePath}:method:${nodeId++}`,
          methodMatch[1],
          line.includes('(') ? 'method' : 'function',
          filePath,
          i,
          indent,
          i,
          line.length
        )
        nodeData.modifiers = []
        if (line.includes('public')) nodeData.modifiers.push('public')
        if (line.includes('private')) nodeData.modifiers.push('private')
        if (line.includes('protected')) nodeData.modifiers.push('protected')
        if (line.includes('static')) nodeData.modifiers.push('static')
        if (line.includes('async')) nodeData.modifiers.push('async')
        nodes.push(nodeData)
      }
    }

    return nodes
  }

  extractEdges(code: string, filePath: string, nodes: CodeNodeData[]): CodeEdgeData[] {
    const edges: CodeEdgeData[] = []

    // Using statements
    const usingPattern = /using\s+([\w.]+)\s*;/g
    let match
    while ((match = usingPattern.exec(code)) !== null) {
      edges.push({
        id: `edge:${filePath}->using:${match[1]}`,
        source: `file:${filePath}`,
        target: match[1],
        type: 'imports',
      })
    }

    // Namespace declarations
    const namespacePattern = /namespace\s+([\w.]+)/g
    while ((match = namespacePattern.exec(code)) !== null) {
      edges.push({
        id: `edge:${filePath}->namespace:${match[1]}`,
        source: `file:${filePath}`,
        target: match[1],
        type: 'imports',
      })
    }

    // Class extends
    const extendsPattern = /class\s+(\w+)\s*:\s*(\w+)/g
    while ((match = extendsPattern.exec(code)) !== null) {
      edges.push({
        id: `edge:${match[1]}->inherits:${match[2]}`,
        source: `${filePath}:${match[1]}`,
        target: `${filePath}:${match[2]}`,
        type: 'inherits',
      })
    }

    // Class implements (interface)
    const implementsPattern = /class\s+(\w+).*:\s*[\w,\s]+,\s*(I\w+)/g
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
