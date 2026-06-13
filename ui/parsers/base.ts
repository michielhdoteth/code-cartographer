import { CodeNode, CodeFile, CodeEdge, CodeMap } from '@models/codeMap'
import { CodeNodeData, CodeEdgeData, CodeFileData, Language, NodeType } from '@models/types'

export interface ParserOptions {
  followImports?: boolean
  extractMetrics?: boolean
  detectPatterns?: boolean
  includeDocstrings?: boolean
  maxDepth?: number
}

export abstract class BaseParser {
  protected language: Language
  protected options: ParserOptions

  constructor(language: Language, options: ParserOptions = {}) {
    this.language = language
    this.options = {
      followImports: true,
      extractMetrics: true,
      detectPatterns: true,
      includeDocstrings: true,
      maxDepth: 10,
      ...options,
    }
  }

  abstract parse(code: string, filePath: string): CodeNodeData[]

  abstract extractEdges(code: string, filePath: string, nodes: CodeNodeData[]): CodeEdgeData[]

  protected createNode(
    id: string,
    name: string,
    type: NodeType,
    filePath: string,
    line: number = 0,
    column: number = 0,
    endLine: number = 0,
    endColumn: number = 0
  ): CodeNodeData {
    return {
      id,
      name,
      type,
      language: this.language,
      file: filePath,
      range: {
        start: { line, column },
        end: { line: endLine, column: endColumn },
      },
      metrics: {
        cyclomatic: 1,
        cognitive: 0,
        linesOfCode: endLine - line,
        nestedDepth: 0,
        numberOfParameters: 0,
      },
    }
  }

  protected calculateLinesOfCode(code: string): number {
    return code.split('\n').length
  }

  protected calculateCyclomaticComplexity(code: string): number {
    // Simple heuristic: count control flow keywords
    const patterns = [/\bif\b/g, /\belse\b/g, /\bfor\b/g, /\bwhile\b/g, /\bcase\b/g, /\bcatch\b/g, /\b&&\b/g, /\b\|\|\b/g]
    let complexity = 1
    for (const pattern of patterns) {
      complexity += (code.match(pattern) || []).length
    }
    return complexity
  }

  protected calculateCognitiveComplexity(code: string): number {
    // Similar to cyclomatic but with different weighting
    let complexity = 0
    const patterns: [RegExp, number][] = [
      [/\bif\b/g, 1],
      [/\belse\s+if\b/g, 1],
      [/\belse\b/g, 0],
      [/\bfor\b/g, 1],
      [/\bwhile\b/g, 1],
      [/\bswitch\b/g, 0],
      [/\bcase\b/g, 1],
      [/\bcatch\b/g, 1],
      [/\?.*:/g, 1],
      [/\bnested\b/g, 1],
    ]

    for (const [pattern, weight] of patterns) {
      complexity += (code.match(pattern) || []).length * weight
    }

    return complexity
  }

  protected extractDocstring(code: string): string | undefined {
    // Try common docstring patterns
    const patterns = [
      /\/\*\*\s*([\s\S]*?)\*\//,
      /\/\/\/\s*(.*?)(?:\n|$)/,
      /"""\s*([\s\S]*?)"""/,
      /'''\s*([\s\S]*?)'''/,
    ]

    for (const pattern of patterns) {
      const match = code.match(pattern)
      if (match) {
        return match[1] || match[0]
      }
    }

    return undefined
  }

  protected extractDecorators(code: string): string[] {
    const decorators: string[] = []
    const pattern = /@\w+/g
    let match
    while ((match = pattern.exec(code)) !== null) {
      decorators.push(match[0])
    }
    return decorators
  }

  protected getNestedDepth(code: string): number {
    let maxDepth = 0
    let currentDepth = 0
    for (const char of code) {
      if (char === '{' || char === '[' || char === '(') {
        currentDepth++
        maxDepth = Math.max(maxDepth, currentDepth)
      } else if (char === '}' || char === ']' || char === ')') {
        currentDepth--
      }
    }
    return maxDepth
  }

  parseFile(filePath: string, code: string, codeMap: CodeMap): CodeFile {
    const fileId = `file:${filePath}`
    const lines = code.split('\n').length

    return new CodeFile({
      id: fileId,
      path: filePath,
      language: this.language,
      size: code.length,
      lines,
      checksum: this.calculateChecksum(code),
      lastModified: Date.now(),
    })
  }

  private calculateChecksum(str: string): string {
    let hash = 0
    for (let i = 0; i < str.length; i++) {
      const char = str.charCodeAt(i)
      hash = (hash << 5) - hash + char
      hash = hash & hash // Convert to 32bit integer
    }
    return hash.toString(16)
  }

  parseAndIntegrate(filePath: string, code: string, codeMap: CodeMap): void {
    // Parse nodes
    const nodes = this.parse(code, filePath)
    for (const nodeData of nodes) {
      codeMap.addNode(new CodeNode(nodeData))
    }

    // Extract edges
    const edges = this.extractEdges(code, filePath, nodes)
    for (const edgeData of edges) {
      codeMap.addEdge(new CodeEdge(edgeData))
    }

    // Add file
    const file = this.parseFile(filePath, code, codeMap)
    codeMap.addFile(file)
  }
}

export class ParserRegistry {
  private parsers: Map<Language, BaseParser> = new Map()

  register(language: Language, parser: BaseParser): void {
    this.parsers.set(language, parser)
  }

  getParser(language: Language): BaseParser | undefined {
    return this.parsers.get(language)
  }

  getAllParsers(): Map<Language, BaseParser> {
    return this.parsers
  }
}
