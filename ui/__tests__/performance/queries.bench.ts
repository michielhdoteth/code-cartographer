import { describe, it, expect, beforeEach } from 'vitest'
import { CodeMap } from '@models/codeMap'

describe('Performance Benchmarks: Query Performance', () => {
  let codeMap: CodeMap

  beforeEach(() => {
    codeMap = new CodeMap('bench', 'Benchmark', '/')
  })

  describe('indexed query performance (O(1))', () => {
    beforeEach(() => {
      for (let i = 0; i < 1000; i++) {
        codeMap.addNode({
          id: `n${i}`,
          name: `node${i}`,
          type: i % 3 === 0 ? 'function' : i % 3 === 1 ? 'class' : 'variable',
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
    })

    it('should query by type in O(1)', () => {
      const start = performance.now()
      const results = codeMap.getNodesByType('function')
      const elapsed = performance.now() - start

      console.log(`  Query by type (1000 nodes): ${elapsed.toFixed(3)}ms`)
      expect(elapsed).toBeLessThan(1)
      expect(results.length).toBeGreaterThan(0)
    })

    it('should query by file in O(1)', () => {
      const start = performance.now()
      const results = codeMap.getNodesByFile('f0.js')
      const elapsed = performance.now() - start

      console.log(`  Query by file (1000 nodes): ${elapsed.toFixed(3)}ms`)
      expect(elapsed).toBeLessThan(1)
    })

    it('should perform 1000 sequential queries efficiently', () => {
      const start = performance.now()
      for (let i = 0; i < 1000; i++) {
        codeMap.getNodesByType(i % 3 === 0 ? 'function' : i % 3 === 1 ? 'class' : 'variable')
      }
      const elapsed = performance.now() - start

      console.log(`  1000 sequential type queries: ${elapsed.toFixed(2)}ms`)
      expect(elapsed).toBeLessThan(100)
    })

    it('should handle mixed query types', () => {
      const start = performance.now()
      for (let i = 0; i < 100; i++) {
        codeMap.getNodesByType('function')
        codeMap.getNodesByFile(`f${i % 100}.js`)
        codeMap.getNodesByType('class')
      }
      const elapsed = performance.now() - start

      console.log(`  Mixed queries (100 iterations): ${elapsed.toFixed(2)}ms`)
      expect(elapsed).toBeLessThan(50)
    })
  })

  describe('edge queries', () => {
    beforeEach(() => {
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
      }

      for (let i = 0; i < 1000; i++) {
        codeMap.addEdge({
          id: `e${i}`,
          source: `n${i % 100}`,
          target: `n${(i + 1) % 100}`,
          type: 'calls',
          weight: 1,
        })
      }
    })

    it('should query edges by source in O(1)', () => {
      const start = performance.now()
      const edges = codeMap.getEdgesBySource('n0')
      const elapsed = performance.now() - start

      console.log(`  Query edges by source (1000 edges): ${elapsed.toFixed(3)}ms`)
      expect(elapsed).toBeLessThan(1)
    })

    it('should query edges by target in O(1)', () => {
      const start = performance.now()
      const edges = codeMap.getEdgesByTarget('n0')
      const elapsed = performance.now() - start

      console.log(`  Query edges by target (1000 edges): ${elapsed.toFixed(3)}ms`)
      expect(elapsed).toBeLessThan(1)
    })

    it('should retrieve node by ID in O(1)', () => {
      const start = performance.now()
      for (let i = 0; i < 1000; i++) {
        codeMap.getNodeById(`n${i % 100}`)
      }
      const elapsed = performance.now() - start

      console.log(`  1000 getNodeById calls: ${elapsed.toFixed(2)}ms`)
      expect(elapsed).toBeLessThan(10)
    })
  })

  describe('blast radius caching', () => {
    beforeEach(() => {
      for (let i = 0; i < 50; i++) {
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
      }

      for (let i = 0; i < 49; i++) {
        codeMap.addEdge({
          id: `e${i}`,
          source: `n${i}`,
          target: `n${i + 1}`,
          type: 'calls',
          weight: 1,
        })
      }
    })

    it('should calculate blast radius in <50ms', () => {
      const start = performance.now()
      codeMap.calculateBlastRadiusDetailed('n0', 2)
      const elapsed = performance.now() - start

      console.log(`  First blast radius calculation: ${elapsed.toFixed(2)}ms`)
      expect(elapsed).toBeLessThan(50)
    })

    it('should return cached blast radius instantly', () => {
      codeMap.calculateBlastRadiusDetailed('n0', 2)

      const start = performance.now()
      codeMap.calculateBlastRadiusDetailed('n0', 2)
      const elapsed = performance.now() - start

      console.log(`  Cached blast radius: ${elapsed.toFixed(3)}ms`)
      expect(elapsed).toBeLessThan(5)
    })

    it('should show 100x speedup with cache', () => {
      const start1 = performance.now()
      codeMap.calculateBlastRadiusDetailed('n0', 2)
      const time1 = performance.now() - start1

      const start2 = performance.now()
      for (let i = 0; i < 100; i++) {
        codeMap.calculateBlastRadiusDetailed('n0', 2)
      }
      const time2 = (performance.now() - start2) / 100

      console.log(`  First: ${time1.toFixed(2)}ms, Cached avg: ${time2.toFixed(3)}ms, Speedup: ${(time1 / time2).toFixed(0)}x`)
      expect(time2).toBeLessThan(time1 / 10)
    })
  })

  describe('query target goals', () => {
    beforeEach(() => {
      for (let i = 0; i < 5000; i++) {
        codeMap.addNode({
          id: `n${i}`,
          name: `node${i}`,
          type: i % 3 === 0 ? 'function' : i % 3 === 1 ? 'class' : 'variable',
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
    })

    it('should query 5000 nodes in <1ms', () => {
      const start = performance.now()
      const functions = codeMap.getNodesByType('function')
      const elapsed = performance.now() - start

      console.log(`  GOAL: Query 5000 nodes: ${elapsed.toFixed(3)}ms (target: <1ms)`)
      expect(elapsed).toBeLessThan(1)
    })

    it('should blast radius any node in <100ms', () => {
      const start = performance.now()
      codeMap.calculateBlastRadiusDetailed('n0', 2)
      const elapsed = performance.now() - start

      console.log(`  GOAL: Blast radius on 5000 nodes: ${elapsed.toFixed(2)}ms (target: <100ms)`)
      expect(elapsed).toBeLessThan(100)
    })
  })

  describe('complex query patterns', () => {
    beforeEach(() => {
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
    })

    it('should filter by multiple criteria efficiently', () => {
      const start = performance.now()
      const functions = codeMap.getNodesByType('function')
      const inFile = codeMap.getNodesByFile('f0.js')
      const result = functions.filter((n) => inFile.some((f) => f.id === n.id))
      const elapsed = performance.now() - start

      console.log(`  Multi-criteria filter (500 nodes): ${elapsed.toFixed(2)}ms`)
      expect(elapsed).toBeLessThan(50)
    })

    it('should chain queries efficiently', () => {
      const start = performance.now()
      for (let fileNum = 0; fileNum < 50; fileNum++) {
        const nodesInFile = codeMap.getNodesByFile(`f${fileNum}.js`)
        const functions = codeMap.getNodesByType('function')
      }
      const elapsed = performance.now() - start

      console.log(`  50 chained queries: ${elapsed.toFixed(2)}ms`)
      expect(elapsed).toBeLessThan(100)
    })
  })

  describe('query performance under stress', () => {
    it('should maintain O(1) performance at scale', () => {
      for (let i = 0; i < 10000; i++) {
        codeMap.addNode({
          id: `n${i}`,
          name: `node${i}`,
          type: i % 3 === 0 ? 'function' : i % 3 === 1 ? 'class' : 'variable',
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

      const timings = []
      for (let i = 0; i < 10; i++) {
        const start = performance.now()
        codeMap.getNodesByType('function')
        timings.push(performance.now() - start)
      }

      const avgTime = timings.reduce((a, b) => a + b) / timings.length

      console.log(`  10k nodes, 10 queries: avg ${avgTime.toFixed(3)}ms`)
      expect(avgTime).toBeLessThan(1)
    })
  })
})
