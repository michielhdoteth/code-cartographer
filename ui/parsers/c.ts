import { BaseParser, ParserOptions } from './base'
import { CodeNodeData, CodeEdgeData } from '@models/types'

export class CParser extends BaseParser {
  constructor(options: ParserOptions = {}) {
    super('c', options)
  }

  parse(code: string, filePath: string): CodeNodeData[] {
    const nodes: CodeNodeData[] = []
    const lines = code.split('\n')
    let nodeId = 0

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i]
      const indent = line.search(/\S/)

      // Function definitions: type name(params) {
      const funcMatch = line.match(
        /^\s*(?:(?:static|inline|extern|const|unsigned|signed)\s+)*(?:\w+[\s*]+)+(\w+)\s*\([^)]*\)\s*(?:const\s*)?{/
      )
      if (funcMatch && !line.includes('#') && !line.includes('if') && !line.includes('for') && !line.includes('while')) {
        const nodeData = this.createNode(
          `${filePath}:fn:${nodeId++}`,
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

      // Function-like macros: #define NAME(...)
      const macroMatch = line.match(/^\s*#define\s+(\w+)\s*\(/)
      if (macroMatch) {
        const nodeData = this.createNode(
          `${filePath}:macro:${nodeId++}`,
          macroMatch[1],
          'constant',
          filePath,
          i,
          indent,
          i,
          line.length
        )
        nodes.push(nodeData)
        continue
      }

      // Simple value macros: #define NAME value
      const valueMacroMatch = line.match(/^\s*#define\s+(\w+)\s+\S/)
      if (valueMacroMatch) {
        const nodeData = this.createNode(
          `${filePath}:macro:${nodeId++}`,
          valueMacroMatch[1],
          'constant',
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
      const structMatch = line.match(/^\s*(?:typedef\s+)?struct\s+(\w+)/)
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
      }

      // Enum definitions
      const enumMatch = line.match(/^\s*(?:typedef\s+)?enum\s+(\w+)/)
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
      }
    }

    return nodes
  }

  extractEdges(code: string, filePath: string, nodes: CodeNodeData[]): CodeEdgeData[] {
    const edges: CodeEdgeData[] = []

    // Include statements
    const includePattern = /#include\s+[<"]([^>"]+)[>"]/g
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
