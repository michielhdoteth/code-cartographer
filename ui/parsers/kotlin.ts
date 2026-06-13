import { BaseParser, ParserOptions } from './base'
import { CodeNodeData, CodeEdgeData } from '@models/types'

export class KotlinParser extends BaseParser {
  constructor(options: ParserOptions = {}) {
    super('kotlin', options)
  }

  parse(code: string, filePath: string): CodeNodeData[] {
    const nodes: CodeNodeData[] = []
    const lines = code.split('\n')
    let nodeId = 0

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i]
      const indent = line.search(/\S/)

      // Function definitions
      const funcMatch = line.match(/^\s*(?:(?:public|private|protected|internal|open|override|suspend)\s+)*fun\s+(?:<[^>]+>\s+)?(\w+)\s*\(/)
      if (funcMatch) {
        const nodeData = this.createNode(
          `${filePath}:func:${nodeId++}`,
          funcMatch[1],
          'function',
          filePath, i, indent, i, line.length
        )
        nodeData.modifiers = []
        if (line.includes('suspend')) nodeData.modifiers.push('suspend')
        if (line.includes('override')) nodeData.modifiers.push('override')
        nodes.push(nodeData)
        continue
      }

      // Data class
      const dataClassMatch = line.match(/^\s*(?:(?:public|private|protected|internal)\s+)*data\s+class\s+(\w+)/)
      if (dataClassMatch) {
        const nodeData = this.createNode(`${filePath}:class:${nodeId++}`, dataClassMatch[1], 'class', filePath, i, indent, i, line.length)
        nodeData.modifiers = ['data']
        nodes.push(nodeData)
        continue
      }

      // Sealed class
      const sealedClassMatch = line.match(/^\s*(?:(?:public|private|protected|internal)\s+)*sealed\s+class\s+(\w+)/)
      if (sealedClassMatch) {
        const nodeData = this.createNode(`${filePath}:class:${nodeId++}`, sealedClassMatch[1], 'class', filePath, i, indent, i, line.length)
        nodeData.modifiers = ['sealed']
        nodes.push(nodeData)
        continue
      }

      // Class definitions
      const classMatch = line.match(/^\s*(?:(?:public|private|protected|internal|open|abstract)\s+)*class\s+(\w+)/)
      if (classMatch) {
        const nodeData = this.createNode(`${filePath}:class:${nodeId++}`, classMatch[1], 'class', filePath, i, indent, i, line.length)
        nodeData.isAbstract = line.includes('abstract')
        nodes.push(nodeData)
        continue
      }

      // Interface definitions
      const interfaceMatch = line.match(/^\s*(?:(?:public|private|protected|internal)\s+)*interface\s+(\w+)/)
      if (interfaceMatch) {
        const nodeData = this.createNode(`${filePath}:interface:${nodeId++}`, interfaceMatch[1], 'interface', filePath, i, indent, i, line.length)
        nodes.push(nodeData)
        continue
      }

      // Object definitions
      const objectMatch = line.match(/^\s*(?:(?:public|private|protected|internal)\s+)*object\s+(\w+)/)
      if (objectMatch) {
        const nodeData = this.createNode(`${filePath}:class:${nodeId++}`, objectMatch[1], 'class', filePath, i, indent, i, line.length)
        nodeData.modifiers = ['object']
        nodes.push(nodeData)
        continue
      }

      // Top-level val/var
      const varMatch = line.match(/^\s*(?:(?:public|private|protected|internal)\s+)*(?:val|var)\s+(\w+)\s*[=:]/)
      if (varMatch && indent === 0) {
        const nodeData = this.createNode(`${filePath}:var:${nodeId++}`, varMatch[1], 'variable', filePath, i, indent, i, line.length)
        nodes.push(nodeData)
      }
    }

    return nodes
  }

  extractEdges(code: string, filePath: string, nodes: CodeNodeData[]): CodeEdgeData[] {
    const edges: CodeEdgeData[] = []

    // Import statements
    const importPattern = /import\s+([\w.]+)/g
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
    const extendsPattern = /class\s+(\w+)(?:<[^>]*>)?\s*(?::\s*(\w+))?/g
    while ((match = extendsPattern.exec(code)) !== null) {
      if (match[2]) {
        edges.push({
          id: `edge:${match[1]}->inherits:${match[2]}`,
          source: `${filePath}:${match[1]}`,
          target: `${filePath}:${match[2]}`,
          type: 'inherits',
        })
      }
    }

    return edges
  }
}
