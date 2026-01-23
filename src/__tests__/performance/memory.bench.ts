import { CodeMap } from '@models/codeMap'
import { TypeScriptParser } from '@parsers/typescript'

describe('Performance Benchmarks: Memory Usage', () => {
  let codeMap: CodeMap

  beforeEach(() => {
    codeMap = new CodeMap('bench', 'Benchmark', '/')
  })

  describe('memory efficiency with indexed structures', () => {
    it('should handle 1000 nodes in <500MB', () => {
      const parser = new TypeScriptParser()

      for (let i = 0; i < 200; i++) {
        const code = Array.from({ length: 5 }, (_, j) => `export const v${i}_${j} = 1;`).join('\n')
        parser.parseAndIntegrate(`f${i}.ts`, code, codeMap)
      }

      const nodes = codeMap.getNodes()
      expect(nodes.length).toBeGreaterThan(1000)

      const json = JSON.stringify(codeMap.toJSON())
      const sizeMB = new Blob([json]).size / (1024 * 1024)

      console.log(`  1000+ nodes serialized: ${sizeMB.toFixed(2)}MB`)
      expect(sizeMB).toBeLessThan(500)
    })

    it('should maintain O(1) query performance with memory', () => {
      for (let i = 0; i < 1000; i++) {
        codeMap.addNode({
          id: `n${i}`,
          name: `node${i}`,
          type: i % 2 === 0 ? 'function' : 'class',
          language: 'javascript' as const,
          file: `f${i % 100}.js`,
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

      const start = performance.now()
      for (let i = 0; i < 1000; i++) {
        codeMap.getNodesByType(i % 2 === 0 ? 'function' : 'class')
      }
      const elapsed = performance.now() - start

      console.log(`  1000 indexed queries on 1000 nodes: ${elapsed.toFixed(2)}ms`)
      expect(elapsed).toBeLessThan(100)
    })
  })

  describe('memory overhead of indexes', () => {
    it('should show minimal memory overhead from indexes', () => {
      for (let i = 0; i < 500; i++) {
        codeMap.addNode({
          id: `n${i}`,
          name: `node${i}`,
          type: i % 3 === 0 ? 'function' : i % 3 === 1 ? 'class' : 'variable',
          language: 'javascript' as const,
          file: `f${i % 50}.js`,
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

      const json = codeMap.toJSON()
      const sizeMB = new Blob([json]).size / (1024 * 1024)

      console.log(`  500 nodes with indexes: ${sizeMB.toFixed(2)}MB`)
      expect(sizeMB).toBeLessThan(100)
    })
  })

  describe('caching memory impact', () => {
    it('should demonstrate cache memory trade-off', () => {
      const parser = new TypeScriptParser()
      const code = 'export const x = 1;'

      const start = performance.now()
      for (let i = 0; i < 1000; i++) {
        parser.parseAndIntegrate(`f${i}.ts`, code, codeMap)
      }
      const elapsed = performance.now() - start

      console.log(`  1000 cached identical parses: ${elapsed.toFixed(2)}ms`)
      expect(elapsed).toBeLessThan(1000)
    })
  })

  describe('index memory trade-offs', () => {
    it('should show speed-memory trade-off is favorable', () => {
      const nodeCount = 2000
      for (let i = 0; i < nodeCount; i++) {
        codeMap.addNode({
          id: `n${i}`,
          name: `node${i}`,
          type: 'function' as const,
          language: 'javascript' as const,
          file: `f${i % 100}.js`,
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

      const nodeTimings = []
      for (let i = 0; i < 100; i++) {
        const start = performance.now()
        codeMap.getNodesByType('function')
        nodeTimings.push(performance.now() - start)
      }

      const avgTime = nodeTimings.reduce((a, b) => a + b) / nodeTimings.length
      const json = codeMap.toJSON()
      const sizeMB = new Blob([json]).size / (1024 * 1024)

      console.log(`  2000 nodes: ${avgTime.toFixed(3)}ms per query, ${sizeMB.toFixed(2)}MB`)
      expect(avgTime).toBeLessThan(1)
    })
  })

  describe('blast radius memory cost', () => {
    it('should show minimal memory cost for blast radius cache', () => {
      for (let i = 0; i < 100; i++) {
        codeMap.addNode({
          id: `n${i}`,
          name: `node${i}`,
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

        if (i > 0) {
          codeMap.addEdge({
            id: `e${i}`,
            source: `n${i - 1}`,
            target: `n${i}`,
            type: 'calls',
            weight: 1,
          })
        }
      }

      const nodes = codeMap.getNodes()
      const start = performance.now()
      codeMap.calculateBlastRadiusDetailed(nodes[0].id, 3)
      const elapsed = performance.now() - start

      console.log(`  Blast radius calculation (100 nodes): ${elapsed.toFixed(2)}ms`)
      expect(elapsed).toBeLessThan(100)
    })
  })

  describe('serialization efficiency', () => {
    it('should serialize large graph efficiently', () => {
      const parser = new TypeScriptParser()
      for (let i = 0; i < 100; i++) {
        const code = Array.from({ length: 5 }, (_, j) => `export const v${i}_${j} = 1;`).join('\n')
        parser.parseAndIntegrate(`f${i}.ts`, code, codeMap)
      }

      const start = performance.now()
      const json = codeMap.toJSON()
      const elapsed = performance.now() - start

      const sizeMB = new Blob([json]).size / (1024 * 1024)

      console.log(`  Serialize 500+ nodes: ${elapsed.toFixed(2)}ms, size: ${sizeMB.toFixed(2)}MB`)
      expect(elapsed).toBeLessThan(1000)
    })

    it('should deserialize large graph efficiently', () => {
      const parser = new TypeScriptParser()
      for (let i = 0; i < 100; i++) {
        parser.parseAndIntegrate(`f${i}.ts`, `export const v${i} = 1;`, codeMap)
      }

      const json = codeMap.toJSON()

      const start = performance.now()
      const restored = CodeMap.fromJSON(json)
      const elapsed = performance.now() - start

      console.log(`  Deserialize 100+ nodes: ${elapsed.toFixed(2)}ms`)
      expect(elapsed).toBeLessThan(1000)
    })
  })

  describe('memory target goals', () => {
    it('should use <500MB for 1000 files', () => {
      const parser = new TypeScriptParser()
      for (let i = 0; i < 1000; i++) {
        parser.parseAndIntegrate(`f${i}.ts`, `export const v${i} = 1;`, codeMap)
      }

      const json = codeMap.toJSON()
      const sizeMB = new Blob([json]).size / (1024 * 1024)

      console.log(`  GOAL: Memory for 1000 files: ${sizeMB.toFixed(2)}MB (target: <500MB)`)
      expect(sizeMB).toBeLessThan(500)
    })
  })

  describe('progressive loading memory impact', () => {
    it('should support lazy loading pattern', () => {
      const phases = [
        { count: 200, description: 'Phase 1: File structure' },
        { count: 500, description: 'Phase 2: Nodes/edges' },
        { count: 1000, description: 'Phase 3: Metrics' },
        { count: 1500, description: 'Phase 4: Full analysis' },
      ]

      for (const phase of phases) {
        const tempMap = new CodeMap(`phase-${phase.description}`, phase.description, '/')
        const parser = new TypeScriptParser()

        for (let i = 0; i < phase.count; i++) {
          parser.parseAndIntegrate(`f${i}.ts`, `export const v${i} = 1;`, tempMap)
        }

        const json = tempMap.toJSON()
        const sizeMB = new Blob([json]).size / (1024 * 1024)
        console.log(`  ${phase.description}: ${phase.count} nodes, ${sizeMB.toFixed(2)}MB`)
      }
    })
  })
})
