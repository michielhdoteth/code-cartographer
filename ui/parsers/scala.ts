import { BaseParser, ParserOptions } from './base'
import { CodeNodeData, CodeEdgeData } from '@models/types'

export class ScalaParser extends BaseParser {
  constructor(options: ParserOptions = {}) {
    super('scala', options)
  }

  parse(code: string, filePath: string): CodeNodeData[] {
    const nodes: CodeNodeData[] = []
    const lines = code.split('\n')
    let nodeId = 0

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i]
      const indent = line.search(/\S/)

      // Function definitions
      const funcMatch = line.match(/^\s*(?:(?:private|protected|override|abstract|final|sealed)\s+)*def\s+(\w+)\s*[\(\[]/)
      if (funcMatch) {
        const nodeData = this.createNode(`${filePath}:func:${nodeId++}`, funcMatch[1], 'function', filePath, i, indent, i, line.length)
        nodes.push(nodeData)
        continue
      }

      // Case class definitions
      const caseClassMatch = line.match(/^\s*(?:(?:private|protected|case)\s+)*case\s+class\s+(\w+)/)
      if (caseClassMatch) {
        const nodeData = this.createNode(`${filePath}:class:${nodeId++}`, caseClassMatch[1], 'class', filePath, i, indent, i, line.length)
        nodeData.modifiers = ['case']
        nodes.push(nodeData)
        continue
      }

      // Class definitions
      const classMatch = line.match(/^\s*(?:(?:private|protected|abstract|final|sealed)\s+)*class\s+(\w+)/)
      if (classMatch) {
        const nodeData = this.createNode(`${filePath}:class:${nodeId++}`, classMatch[1], 'class', filePath, i, indent, i, line.length)
        nodeData.isAbstract = line.includes('abstract')
        nodes.push(nodeData)
        continue
      }

      // Trait definitions
      const traitMatch = line.match(/^\s*(?:(?:private|protected)\s+)*trait\s+(\w+)/)
      if (traitMatch) {
        const nodeData = this.createNode(`${filePath}:trait:${nodeId++}`, traitMatch[1], 'interface', filePath, i, indent, i, line.length)
        nodes.push(nodeData)
        continue
      }

      // Object definitions
      const objectMatch = line.match(/^\s*(?:(?:private|protected)\s+)*object\s+(\w+)/)
      if (objectMatch) {
        const nodeData = this.createNode(`${filePath}:class:${nodeId++}`, objectMatch[1], 'class', filePath, i, indent, i, line.length)
        nodeData.modifiers = ['object']
        nodes.push(nodeData)
        continue
      }

      // val/var definitions
      const varMatch = line.match(/^\s*(?:(?:private|protected|override)\s+)*(?:val|var)\s+(\w+)\s*[=:]/)
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

    // Class/trait extends
    const extendsPattern = /(?:class|trait|object)\s+(\w+)(?:\[[^\]]*\])?(?:\([^)]*\))?\s+extends\s+(\w+)/g
    while ((match = extendsPattern.exec(code)) !== null) {
      edges.push({
        id: `edge:${match[1]}->inherits:${match[2]}`,
        source: `${filePath}:${match[1]}`,
        target: `${filePath}:${match[2]}`,
        type: 'inherits',
      })
    }

    // with (trait mixin)
    const withPattern = /(?:class|object)\s+(\w+).*with\s+(\w+)/g
    while ((match = withPattern.exec(code)) !== null) {
      edges.push({
        id: `edge:${match[1]}->implements:${match[2]}`,
        source: `${filePath}:${match[1]}`,
        target: `${filePath}:${match[2]}`,
        type: 'implements',
      })
    }

    return edges
  }
}
