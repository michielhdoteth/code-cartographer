import { BaseParser, ParserOptions } from './base'
import { CodeNodeData, CodeEdgeData } from '@models/types'

export class CppParser extends BaseParser {
  constructor(options: ParserOptions = {}) {
    super('cpp', options)
  }

  parse(code: string, filePath: string): CodeNodeData[] {
    const nodes: CodeNodeData[] = []
    const lines = code.split('\n')
    let nodeId = 0

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i]
      const indent = line.search(/\S/)

      // Class definitions
      const classMatch = line.match(/^\s*(?:class|struct)\s+(\w+)(?:\s*[:]\s*(?:public|private|protected)\s+(\w+))?/)
      if (classMatch && line.includes('{')) {
        const nodeData = this.createNode(
          `${filePath}:class:${nodeId++}`,
          classMatch[1],
          line.includes('class') ? 'class' : 'struct',
          filePath,
          i,
          indent,
          i,
          line.length
        )
        nodes.push(nodeData)
      }

      // Function definitions
      const funcMatch = line.match(/\w+\s+(\w+)\s*\([^)]*\)\s*(?:const\s*)?{/)
      if (funcMatch && !line.includes('if') && !line.includes('for') && !line.includes('while')) {
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
        nodes.push(nodeData)
      }
    }

    return nodes
  }

  extractEdges(code: string, filePath: string, nodes: CodeNodeData[]): CodeEdgeData[] {
    const edges: CodeEdgeData[] = []

    // Include statements
    const includePattern = /#include\s+[<"]([^>"]+)[>"]*/g
    let match
    while ((match = includePattern.exec(code)) !== null) {
      edges.push({
        id: `edge:${filePath}->include:${match[1]}`,
        source: `file:${filePath}`,
        target: match[1],
        type: 'imports',
      })
    }

    return edges
  }
}
