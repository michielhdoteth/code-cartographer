import { CodeMap } from '@models/codeMap'
import { TypeScriptParser } from '@parsers/typescript'

describe('E2E: Large Codebase (1000+ files)', () => {
  let codeMap: CodeMap

  beforeEach(() => {
    codeMap = new CodeMap('large-project', 'Large Project', '/project')
  })

  describe('parsing performance benchmarks', () => {
    it('should parse 100 files in <2 seconds', () => {
      const parser = new TypeScriptParser()
      const start = performance.now()

      for (let i = 0; i < 100; i++) {
        const code = `
          export interface Data${i} {
            id: number;
            name: string;
          }

          export class Service${i} {
            getData(): Data${i} {
              return { id: 1, name: 'test' };
            }
          }

          export function process${i}(data: Data${i}): string {
            return data.name;
          }
        `
        parser.parseAndIntegrate(`src/file${i}.ts`, code, codeMap)
      }

      const elapsed = performance.now() - start
      expect(elapsed).toBeLessThan(2000)
      expect(codeMap.getNodes().length).toBeGreaterThan(200)
    })

    it('should parse 500 files efficiently', () => {
      const parser = new TypeScriptParser()
      const start = performance.now()

      for (let i = 0; i < 500; i++) {
        const code = `
          export const value${i} = ${i};
          export function func${i}() { return ${i}; }
        `
        parser.parseAndIntegrate(`src/module${Math.floor(i / 10)}/file${i}.ts`, code, codeMap)
      }

      const elapsed = performance.now() - start
      expect(elapsed).toBeLessThan(5000)
      expect(codeMap.getNodes().length).toBeGreaterThan(500)
    })
  })

  describe('memory efficiency', () => {
    it('should handle 1000+ nodes in memory', () => {
      const parser = new TypeScriptParser()

      for (let i = 0; i < 200; i++) {
        const functions = Array.from({ length: 5 }, (_, j) => `
          export function func${i}_${j}(x: number): number {
            return x * ${i + j};
          }
        `).join('\n')

        parser.parseAndIntegrate(`src/file${i}.ts`, functions, codeMap)
      }

      const nodes = codeMap.getNodes()
      expect(nodes.length).toBeGreaterThan(1000)
    })

    it('should maintain index performance with large dataset', () => {
      const parser = new TypeScriptParser()

      for (let i = 0; i < 100; i++) {
        parser.parseAndIntegrate(
          `src/file${i}.ts`,
          `export class Class${i} {}
           export function func${i}() {}
           export const var${i} = 1;`,
          codeMap
        )
      }

      const start = performance.now()
      const functions = codeMap.getNodesByType('function')
      const classes = codeMap.getNodesByType('class')
      const elapsed = performance.now() - start

      expect(elapsed).toBeLessThan(1)
      expect(functions.length).toBeGreaterThan(0)
      expect(classes.length).toBeGreaterThan(0)
    })
  })

  describe('query performance at scale', () => {
    beforeEach(() => {
      const parser = new TypeScriptParser()
      for (let i = 0; i < 100; i++) {
        parser.parseAndIntegrate(
          `src/module${i % 10}/file${i}.ts`,
          `export class Cls${i} {}`,
          codeMap
        )
      }
    })

    it('should query by file in constant time', () => {
      const start = performance.now()
      const nodes = codeMap.getNodesByFile('src/module0/file0.ts')
      const elapsed = performance.now() - start

      expect(elapsed).toBeLessThan(1)
    })

    it('should query by type in constant time', () => {
      const start = performance.now()
      const nodes = codeMap.getNodesByType('class')
      const elapsed = performance.now() - start

      expect(elapsed).toBeLessThan(1)
      expect(nodes.length).toBeGreaterThan(0)
    })

    it('should query edges by source efficiently', () => {
      if (codeMap.getEdges().length > 0) {
        const firstEdge = codeMap.getEdges()[0]
        const start = performance.now()
        const edges = codeMap.getEdgesBySource(firstEdge.source)
        const elapsed = performance.now() - start

        expect(elapsed).toBeLessThan(1)
      }
    })
  })

  describe('blast radius calculation at scale', () => {
    beforeEach(() => {
      const parser = new TypeScriptParser()

      for (let i = 0; i < 50; i++) {
        parser.parseAndIntegrate(
          `src/service${i}.ts`,
          `
          export class Service${i} {
            method() { return 'result'; }
          }
          `,
          codeMap
        )
      }

      for (let i = 0; i < 49; i++) {
        codeMap.addEdge({
          id: `edge${i}`,
          source: `Service${i}`,
          target: `Service${i + 1}`,
          type: 'depends',
          weight: 1,
        })
      }
    })

    it('should calculate blast radius in <100ms', () => {
      const nodes = codeMap.getNodes()
      if (nodes.length > 0) {
        const start = performance.now()
        codeMap.calculateBlastRadiusDetailed(nodes[0].id, 2)
        const elapsed = performance.now() - start

        expect(elapsed).toBeLessThan(100)
      }
    })

    it('should calculate cascading blast radius', () => {
      const nodes = codeMap.getNodes()
      if (nodes.length > 0) {
        const radius = codeMap.calculateBlastRadiusDetailed(nodes[0].id, 3)
        expect(radius.size).toBeGreaterThanOrEqual(1)
      }
    })
  })

  describe('visualization readiness', () => {
    beforeEach(() => {
      const parser = new TypeScriptParser()

      for (let i = 0; i < 200; i++) {
        parser.parseAndIntegrate(
          `src/components/component${i}.ts`,
          `
          export const Component${i} = {
            name: 'Component${i}',
            render() { return 'rendered'; }
          };
          `,
          codeMap
        )
      }
    })

    it('should prepare graph data structure quickly', () => {
      const start = performance.now()
      const nodes = codeMap.getNodes()
      const edges = codeMap.getEdges()
      const elapsed = performance.now() - start

      expect(elapsed).toBeLessThan(100)
      expect(nodes.length).toBeGreaterThan(0)
    })

    it('should organize nodes by file', () => {
      const fileGroups = new Map<string, number>()
      codeMap.getNodes().forEach((node) => {
        const file = node.file
        fileGroups.set(file, (fileGroups.get(file) || 0) + 1)
      })

      expect(fileGroups.size).toBeGreaterThan(0)
    })

    it('should provide statistics for visualization', () => {
      const stats = {
        totalNodes: codeMap.getNodes().length,
        totalEdges: codeMap.getEdges().length,
        nodesByType: new Map<string, number>(),
        fileCount: new Set(codeMap.getNodes().map((n) => n.file)).size,
      }

      codeMap.getNodes().forEach((node) => {
        stats.nodesByType.set(node.type, (stats.nodesByType.get(node.type) || 0) + 1)
      })

      expect(stats.totalNodes).toBeGreaterThan(0)
      expect(stats.fileCount).toBeGreaterThan(0)
    })
  })

  describe('concurrent operations', () => {
    it('should handle sequential file parsing', async () => {
      const parser = new TypeScriptParser()
      const start = performance.now()

      const promises = Array.from({ length: 50 }, (_, i) =>
        Promise.resolve(
          parser.parseAndIntegrate(
            `src/async${i}.ts`,
            `export const x${i} = ${i};`,
            codeMap
          )
        )
      )

      await Promise.all(promises)
      const elapsed = performance.now() - start

      expect(elapsed).toBeLessThan(5000)
      expect(codeMap.getNodes().length).toBeGreaterThan(0)
    })
  })

  describe('serialization of large graphs', () => {
    it('should serialize 500 node graph', () => {
      const parser = new TypeScriptParser()

      for (let i = 0; i < 50; i++) {
        parser.parseAndIntegrate(
          `src/file${i}.ts`,
          Array.from({ length: 10 }, (_, j) => `export const x${i}_${j} = 1;`).join('\n'),
          codeMap
        )
      }

      const start = performance.now()
      const json = codeMap.toJSON()
      const elapsed = performance.now() - start

      expect(elapsed).toBeLessThan(1000)
      expect(typeof json).toBe('string')
      expect(json.length).toBeGreaterThan(0)
    })

    it('should deserialize large graph', () => {
      const parser = new TypeScriptParser()

      for (let i = 0; i < 20; i++) {
        parser.parseAndIntegrate(`src/file${i}.ts`, `export const x = ${i};`, codeMap)
      }

      const json = codeMap.toJSON()
      const start = performance.now()
      const restored = CodeMap.fromJSON(json)
      const elapsed = performance.now() - start

      expect(elapsed).toBeLessThan(1000)
      expect(restored.getNodes().length).toBe(codeMap.getNodes().length)
    })
  })

  describe('stress testing', () => {
    it('should handle extreme node additions', () => {
      for (let i = 0; i < 5000; i++) {
        codeMap.addNode({
          id: `node${i}`,
          name: `node${i}`,
          type: 'function' as const,
          language: 'javascript' as const,
          file: `file${i % 100}.js`,
          range: { start: { line: 0, column: 0 }, end: { line: 10, column: 0 } },
          metrics: {
            cyclomatic: 1,
            cognitive: 1,
            linesOfCode: 10,
            nestedDepth: 0,
            numberOfParameters: 0,
          },
        })
      }

      expect(codeMap.getNodes().length).toBe(5000)
    })

    it('should maintain performance with 5000 edges', () => {
      const nodes = Array.from({ length: 100 }, (_, i) =>
        codeMap.addNode({
          id: `n${i}`,
          name: `n${i}`,
          type: 'function' as const,
          language: 'javascript' as const,
          file: `f${i}.js`,
          range: { start: { line: 0, column: 0 }, end: { line: 10, column: 0 } },
          metrics: {
            cyclomatic: 1,
            cognitive: 1,
            linesOfCode: 10,
            nestedDepth: 0,
            numberOfParameters: 0,
          },
        })
      )

      const start = performance.now()
      for (let i = 0; i < 5000; i++) {
        codeMap.addEdge({
          id: `e${i}`,
          source: `n${i % 100}`,
          target: `n${(i + 1) % 100}`,
          type: 'calls',
          weight: 1,
        })
      }
      const elapsed = performance.now() - start

      expect(elapsed).toBeLessThan(500)
      expect(codeMap.getEdges().length).toBe(5000)
    })
  })
})
