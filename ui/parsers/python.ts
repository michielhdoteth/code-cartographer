import { BaseParser, ParserOptions } from './base'
import { CodeNodeData, CodeEdgeData } from '@models/types'

export class PythonParser extends BaseParser {
  constructor(options: ParserOptions = {}) {
    super('python', options)
  }

  parse(code: string, filePath: string): CodeNodeData[] {
    const nodes: CodeNodeData[] = []
    const lines = code.split('\n')
    let nodeId = 0

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i]
      const indent = line.search(/\S/)

      // Class definitions
      const classMatch = line.match(/^\s*class\s+(\w+)\s*(?:\(([^)]*)\))?:/)
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
        if (classMatch[2]) {
          nodeData.metrics.numberOfParameters = classMatch[2].split(',').length
        }
        nodes.push(nodeData)
        continue
      }

      // Function definitions
      const funcMatch = line.match(/^\s*(?:async\s+)?def\s+(\w+)\s*\(([^)]*)\)/)
      if (funcMatch) {
        const nodeData = this.createNode(
          `${filePath}:func:${nodeId++}`,
          funcMatch[1],
          indent === 0 ? 'function' : 'method',
          filePath,
          i,
          indent,
          i,
          line.length
        )
        // Count parameters
        const params = funcMatch[2].split(',').filter(p => p.trim() && p.trim() !== 'self')
        nodeData.metrics.numberOfParameters = params.length
        nodes.push(nodeData)
        continue
      }

      // Decorators
      const decorMatch = line.match(/^\s*@(\w+)/)
      if (decorMatch) {
        const nodeData = this.createNode(
          `${filePath}:decorator:${nodeId++}`,
          decorMatch[1],
          'decorator',
          filePath,
          i,
          indent,
          i,
          line.length
        )
        nodeData.decorators = [decorMatch[1]]
        nodes.push(nodeData)
        continue
      }

      // Module-level variables/constants
      const varMatch = line.match(/^\s*([A-Z_][A-Z0-9_]*)\s*=/)
      if (varMatch && indent === 0) {
        const nodeData = this.createNode(
          `${filePath}:const:${nodeId++}`,
          varMatch[1],
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
    }

    return nodes
  }

  extractEdges(code: string, filePath: string, nodes: CodeNodeData[]): CodeEdgeData[] {
    const edges: CodeEdgeData[] = []
    const lines = code.split('\n')

    // Import statements
    for (const line of lines) {
      const importMatch = line.match(/^\s*(?:from\s+(\S+)\s+)?import\s+(.+)/)
      if (importMatch) {
        const module = importMatch[1] || importMatch[2]
        edges.push({
          id: `edge:${filePath}->import:${module}`,
          source: `file:${filePath}`,
          target: module,
          type: 'imports',
        })
      }
    }

    // Class inheritance
    for (const line of lines) {
      const inheritMatch = line.match(/^\s*class\s+(\w+)\s*\(\s*([^)]+)\s*\)/)
      if (inheritMatch) {
        const baseClasses = inheritMatch[2].split(',').map(c => c.trim())
        for (const base of baseClasses) {
          edges.push({
            id: `edge:${inheritMatch[1]}->inherits:${base}`,
            source: `${filePath}:${inheritMatch[1]}`,
            target: `${filePath}:${base}`,
            type: 'inherits',
          })
        }
      }
    }

    // Function calls
    const callPattern = /(\w+)\s*\(/g
    let match
    const calls = new Set<string>()

    while ((match = callPattern.exec(code)) !== null) {
      calls.add(match[1])
    }

    for (const call of calls) {
      for (const node of nodes) {
        if (node.name === call && (node.type === 'function' || node.type === 'method')) {
          edges.push({
            id: `edge:${filePath}->call:${call}`,
            source: `file:${filePath}`,
            target: node.id,
            type: 'calls',
          })
          break
        }
      }
    }

    return edges
  }
}
