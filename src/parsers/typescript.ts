import { JavaScriptParser } from './javascript'
import { CodeNodeData, CodeEdgeData } from '@models/types'
import { ParserOptions } from './base'

export class TypeScriptParser extends JavaScriptParser {
  constructor(options: ParserOptions = {}) {
    super('typescript', options)
  }

  parse(code: string, filePath: string): CodeNodeData[] {
    // TypeScript is a superset of JavaScript, so we use the JS parser
    // But we might want to strip type annotations first for better parsing
    const strippedCode = this.stripTypeAnnotations(code)
    return super.parse(strippedCode, filePath)
  }

  extractEdges(code: string, filePath: string, nodes: CodeNodeData[]): CodeEdgeData[] {
    const strippedCode = this.stripTypeAnnotations(code)
    return super.extractEdges(strippedCode, filePath, nodes)
  }

  private stripTypeAnnotations(code: string): string {
    let stripped = code

    // Remove type annotations from function parameters
    stripped = stripped.replace(/:\s*[^=,)]+/g, '')

    // Remove type annotations from variables
    stripped = stripped.replace(/<[^>]+>/g, '')

    // Remove interface and type declarations (but keep the structure)
    stripped = stripped.replace(/interface\s+/g, 'class ')
    stripped = stripped.replace(/type\s+\w+\s*=\s*[^;]+;/g, '')

    return stripped
  }
}
