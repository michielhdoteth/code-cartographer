import { BaseParser, ParserOptions } from './base'
import { CodeNodeData, CodeEdgeData, Language } from '@models/types'
import * as acorn from 'acorn'

export class JavaScriptParser extends BaseParser {
  constructor(language: Language = 'javascript', options: ParserOptions = {}) {
    super(language, options)
  }

  parse(code: string, filePath: string): CodeNodeData[] {
    const nodes: CodeNodeData[] = []
    const nodeId = new Map<string, number>()

    try {
      const ast = acorn.parse(code, {
        ecmaVersion: 2022,
        sourceType: 'module',
        locations: true,
      })

      const walkAst = (node: any, parentId?: string, depth = 0) => {
        if (depth > (this.options.maxDepth || 10)) return

        const id = `${filePath}:${node.type}:${(nodeId.get(node.type) || 0) + 1}`
        nodeId.set(node.type, (nodeId.get(node.type) || 0) + 1)

        switch (node.type) {
          case 'FunctionDeclaration': {
            const nodeData = this.createNode(
              id,
              node.id?.name || 'anonymous',
              'function',
              filePath,
              node.loc?.start.line || 0,
              node.loc?.start.column || 0,
              node.loc?.end.line || 0,
              node.loc?.end.column || 0
            )
            nodeData.parent = parentId
            nodeData.metrics.numberOfParameters = node.params?.length || 0
            nodeData.metrics.cyclomatic = this.calculateCyclomaticComplexity(code.substring(node.start, node.end))
            nodes.push(nodeData)
            break
          }

          case 'ArrowFunctionExpression': {
            const nodeData = this.createNode(
              id,
              'arrow_function',
              'function',
              filePath,
              node.loc?.start.line || 0,
              node.loc?.start.column || 0,
              node.loc?.end.line || 0,
              node.loc?.end.column || 0
            )
            nodeData.parent = parentId
            nodeData.metrics.numberOfParameters = node.params?.length || 0
            nodes.push(nodeData)
            break
          }

          case 'ClassDeclaration': {
            const nodeData = this.createNode(
              id,
              node.id?.name || 'anonymous',
              'class',
              filePath,
              node.loc?.start.line || 0,
              node.loc?.start.column || 0,
              node.loc?.end.line || 0,
              node.loc?.end.column || 0
            )
            nodeData.parent = parentId
            nodes.push(nodeData)

            // Walk class methods
            if (node.body?.body) {
              for (const method of node.body.body) {
                walkAst(method, id, depth + 1)
              }
            }
            return // Skip default recursion
          }

          case 'MethodDefinition': {
            const nodeData = this.createNode(
              id,
              node.key?.name || 'unknown',
              'method',
              filePath,
              node.loc?.start.line || 0,
              node.loc?.start.column || 0,
              node.loc?.end.line || 0,
              node.loc?.end.column || 0
            )
            nodeData.parent = parentId
            nodeData.isExported = node.key?.exported || false
            nodes.push(nodeData)
            break
          }

          case 'VariableDeclarator': {
            if (node.id?.name) {
              const nodeData = this.createNode(
                id,
                node.id.name,
                'variable',
                filePath,
                node.loc?.start.line || 0,
                node.loc?.start.column || 0,
                node.loc?.end.line || 0,
                node.loc?.end.column || 0
              )
              nodeData.parent = parentId
              nodes.push(nodeData)
            }
            break
          }

          case 'ExportNamedDeclaration':
          case 'ExportDefaultDeclaration': {
            if (node.declaration) {
              walkAst(node.declaration, parentId, depth + 1)
            }
            return
          }
        }

        // Recursively walk children
        for (const key in node) {
          if (key.startsWith('_')) continue
          const child = node[key]
          if (child && typeof child === 'object') {
            if (Array.isArray(child)) {
              for (const item of child) {
                if (item?.type) {
                  walkAst(item, parentId, depth + 1)
                }
              }
            } else if (child.type) {
              walkAst(child, parentId, depth + 1)
            }
          }
        }
      }

      // Start walking from module level
      for (const node of ast.body) {
        walkAst(node, undefined, 0)
      }
    } catch (error) {
      console.error(`Failed to parse ${filePath}:`, error)
      // Fallback to regex-based parsing
      return this.parseRegex(code, filePath)
    }

    return nodes
  }

  extractEdges(code: string, filePath: string, nodes: CodeNodeData[]): CodeEdgeData[] {
    const edges: CodeEdgeData[] = []

    // Import statements
    const importPattern = /import\s+(?:{[^}]*}|\*\s+as\s+\w+|\w+)\s+from\s+['"]([^'"]+)['"]/g
    let match
    while ((match = importPattern.exec(code)) !== null) {
      const importPath = match[1]
      edges.push({
        id: `edge:${filePath}->import:${importPath}`,
        source: `file:${filePath}`,
        target: importPath,
        type: 'imports',
      })
    }

    // Require statements
    const requirePattern = /require\(['"]([^'"]+)['"]\)/g
    while ((match = requirePattern.exec(code)) !== null) {
      const importPath = match[1]
      edges.push({
        id: `edge:${filePath}->require:${importPath}`,
        source: `file:${filePath}`,
        target: importPath,
        type: 'imports',
      })
    }

    // Class inheritance
    const extendsPattern = /class\s+(\w+)\s+extends\s+(\w+)/g
    while ((match = extendsPattern.exec(code)) !== null) {
      edges.push({
        id: `edge:${match[1]}->extends:${match[2]}`,
        source: `${filePath}:${match[1]}`,
        target: `${filePath}:${match[2]}`,
        type: 'inherits',
      })
    }

    // Method calls
    const callPattern = /\.(\w+)\s*\(/g
    const functionCalls = new Set<string>()
    while ((match = callPattern.exec(code)) !== null) {
      functionCalls.add(match[1])
    }

    for (const func of functionCalls) {
      for (const node of nodes) {
        if (node.name === func && node.type === 'method') {
          edges.push({
            id: `edge:${filePath}->call:${func}`,
            source: `file:${filePath}`,
            target: node.id,
            type: 'calls',
          })
        }
      }
    }

    return edges
  }

  private parseRegex(code: string, filePath: string): CodeNodeData[] {
    const nodes: CodeNodeData[] = []
    const lines = code.split('\n')

    // Function declarations
    const funcPattern = /^\s*(?:async\s+)?function\s+(\w+)\s*\(/gm
    let match
    while ((match = funcPattern.exec(code)) !== null) {
      const line = code.substring(0, match.index).split('\n').length - 1
      nodes.push(
        this.createNode(
          `${filePath}:func:${match[1]}`,
          match[1],
          'function',
          filePath,
          line,
          0,
          line,
          match[0].length
        )
      )
    }

    // Class declarations
    const classPattern = /^\s*class\s+(\w+)/gm
    while ((match = classPattern.exec(code)) !== null) {
      const line = code.substring(0, match.index).split('\n').length - 1
      nodes.push(
        this.createNode(
          `${filePath}:class:${match[1]}`,
          match[1],
          'class',
          filePath,
          line,
          0,
          line,
          match[0].length
        )
      )
    }

    return nodes
  }
}
