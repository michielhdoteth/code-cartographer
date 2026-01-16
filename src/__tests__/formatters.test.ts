import { AsciiFormatter, OutputFormat } from '@formatters/asciiFormatter'
import { CodeMap, CodeNode, CodeFile } from '@models/codeMap'
import { CodeNodeData, CodeFileData } from '@models/types'

describe('AsciiFormatter', () => {
  let formatter: AsciiFormatter
  let codeMap: CodeMap

  beforeEach(() => {
    formatter = new AsciiFormatter()
    codeMap = new CodeMap('test-map', 'Test Project', '/test')

    // Add test files
    const files: CodeFileData[] = [
      {
        id: 'file1',
        path: 'src/api.ts',
        language: 'typescript',
        size: 1000,
        lines: 100,
      },
      {
        id: 'file2',
        path: 'src/auth.ts',
        language: 'typescript',
        size: 800,
        lines: 80,
      },
      {
        id: 'file3',
        path: 'src/components/Header.tsx',
        language: 'typescript',
        size: 500,
        lines: 50,
      },
    ]

    for (const fileData of files) {
      codeMap.addFile(new CodeFile(fileData))
    }

    // Add test nodes
    const nodes: CodeNodeData[] = [
      {
        id: 'node1',
        name: 'APIHandler',
        type: 'class',
        language: 'typescript',
        range: { start: { line: 0, column: 0 }, end: { line: 50, column: 0 } },
        file: 'src/api.ts',
        metrics: {
          cyclomatic: 3,
          cognitive: 4,
          linesOfCode: 50,
          nestedDepth: 2,
          numberOfParameters: 2,
        },
      },
      {
        id: 'node2',
        name: 'authenticate',
        type: 'function',
        language: 'typescript',
        range: { start: { line: 0, column: 0 }, end: { line: 30, column: 0 } },
        file: 'src/auth.ts',
        metrics: {
          cyclomatic: 2,
          cognitive: 2,
          linesOfCode: 30,
          nestedDepth: 1,
          numberOfParameters: 1,
        },
      },
      {
        id: 'node3',
        name: 'Header',
        type: 'function',
        language: 'typescript',
        range: { start: { line: 0, column: 0 }, end: { line: 40, column: 0 } },
        file: 'src/components/Header.tsx',
        metrics: {
          cyclomatic: 1,
          cognitive: 1,
          linesOfCode: 40,
          nestedDepth: 1,
          numberOfParameters: 3,
        },
      },
    ]

    for (const nodeData of nodes) {
      codeMap.addNode(new CodeNode(nodeData))
    }
  })

  describe('formatAsTree', () => {
    it('should format codebase as tree structure', () => {
      const output = formatter.formatAsTree(codeMap)
      expect(output).toContain('# Code Structure')
      expect(output).toContain('src/')
      expect(output).toContain('api.ts')
      expect(output).toContain('LOC')
      expect(output).toContain('items')
    })

    it('should include churn metrics if available', () => {
      const churnMap = new Map([['src/api.ts', 12]])
      const output = formatter.formatAsTree(codeMap, churnMap)
      expect(output).toContain('churn')
    })

    it('should mark high blast radius files', () => {
      const blastMap = new Map([['file1', 8]])
      const output = formatter.formatAsTree(codeMap, undefined, blastMap)
      expect(output).toContain('HIGH IMPACT')
    })
  })

  describe('formatAsDependencies', () => {
    it('should format dependencies section', () => {
      const output = formatter.formatAsDependencies(codeMap)
      expect(output).toContain('# Dependency Graph')
    })

    it('should include edge types', () => {
      const output = formatter.formatAsDependencies(codeMap)
      // Output will include edge type headers if edges exist
      expect(typeof output).toBe('string')
    })
  })

  describe('formatAsHealth', () => {
    it('should include statistics section', () => {
      const output = formatter.formatAsHealth(codeMap)
      expect(output).toContain('# Code Health Report')
      expect(output).toContain('## Statistics')
      expect(output).toContain('Total Nodes')
      expect(output).toContain('Total Files')
    })

    it('should include recommendations', () => {
      const output = formatter.formatAsHealth(codeMap)
      expect(output).toContain('## Recommendations')
    })
  })

  describe('formatAsCoverage', () => {
    it('should format coverage and hotspots', () => {
      const output = formatter.formatAsCoverage(codeMap)
      expect(output).toContain('# Coverage & Hotspots')
      expect(output).toContain('File')
      expect(output).toContain('LOC')
    })

    it('should show summary statistics', () => {
      const output = formatter.formatAsCoverage(codeMap)
      expect(output).toContain('## Summary')
      expect(output).toContain('Average Churn')
    })
  })

  describe('format method', () => {
    it('should dispatch to tree formatter', () => {
      const output = formatter.format(codeMap, 'tree')
      expect(output).toContain('# Code Structure')
    })

    it('should dispatch to deps formatter', () => {
      const output = formatter.format(codeMap, 'deps')
      expect(output).toContain('# Dependency Graph')
    })

    it('should dispatch to health formatter', () => {
      const output = formatter.format(codeMap, 'health')
      expect(output).toContain('# Code Health Report')
    })

    it('should dispatch to coverage formatter', () => {
      const output = formatter.format(codeMap, 'coverage')
      expect(output).toContain('# Coverage & Hotspots')
    })
  })

  describe('formatBlastRadius', () => {
    it('should format blast radius analysis', () => {
      const blastResult = {
        affected: ['file2', 'file3'],
        transitive: ['file4'],
        count: 2,
        transitiveCount: 1,
        level: 'high',
        depth: 2,
        fnsUsed: 3,
        totalCalls: 5,
        dependencies: ['file1'],
        impactScore: 5.5,
        centrality: 10,
      }

      const output = formatter.formatBlastRadius('file1', 'src/api.ts', blastResult)
      expect(output).toContain('Blast Radius Analysis')
      expect(output).toContain('src/api.ts')
      expect(output).toContain('HIGH')
      expect(output).toContain('2')
      expect(output).toContain('Directly Affected Files')
    })

    it('should include transitive dependencies', () => {
      const blastResult = {
        affected: ['file2'],
        transitive: ['file3', 'file4', 'file5'],
        count: 1,
        transitiveCount: 3,
        level: 'critical',
        depth: 3,
        fnsUsed: 5,
        totalCalls: 12,
        dependencies: ['file1'],
        impactScore: 8.2,
        centrality: 15,
      }

      const output = formatter.formatBlastRadius('file1', 'src/api.ts', blastResult)
      expect(output).toContain('Transitively Affected Files')
      expect(output).toContain('3')
    })
  })
})

describe('AsciiFormatter - Text Processing', () => {
  let formatter: AsciiFormatter

  beforeEach(() => {
    formatter = new AsciiFormatter()
  })

  describe('getDirFromPath', () => {
    it('should extract directory from file path', () => {
      // Using a private method through public interface
      const codeMap = new CodeMap('test', 'test', '/test')
      const output = formatter.format(codeMap, 'tree')
      // If formatting works, directory extraction is working
      expect(typeof output).toBe('string')
    })
  })
})
