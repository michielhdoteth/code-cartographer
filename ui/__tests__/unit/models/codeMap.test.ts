import { describe, it, expect, beforeEach } from 'vitest'
import { CodeMap } from '@models/codeMap'
import { createTestNode, createTestEdge } from '../../fixtures/index'
import { createTestCodeMap, createPopulatedCodeMap, measurePerformance, expectPerformanceWithin } from '../../helpers/index'

describe('CodeMap (Indexed Data Structures)', () => {
  let codeMap: CodeMap

  beforeEach(() => {
    codeMap = createTestCodeMap('test-project', 'Test Project', '/test')
  })

  describe('node operations', () => {
    it('should add nodes', () => {
      codeMap.addNode(createTestNode({ id: 'fn1', name: 'myFunction' }))
      expect(codeMap.getNodes().length).toBe(1)
    })

    it('should retrieve nodes by ID', () => {
      codeMap.addNode(createTestNode({ id: 'fn1', name: 'myFunction' }))
      const retrieved = codeMap.getNodeById('fn1')
      expect(retrieved?.id).toBe('fn1')
    })

    it('should handle multiple nodes', () => {
      for (let i = 0; i < 100; i++) {
        codeMap.addNode(createTestNode({ id: `fn${i}`, name: `function${i}`, file: `file${i % 10}.js` }))
      }
      expect(codeMap.getNodes().length).toBe(100)
    })
  })

  describe('indexed queries (O(1) performance)', () => {
    beforeEach(() => {
      for (let i = 0; i < 50; i++) {
        codeMap.addNode(
          createTestNode({
            id: `fn${i}`,
            name: `function${i}`,
            type: i % 2 === 0 ? 'function' : 'class',
            file: `file${i % 5}.js`,
          })
        )
      }
    })

    it('should query nodes by type in O(1)', () => {
      const { elapsed, result: functions } = measurePerformance(() => codeMap.getNodesByType('function'))
      expectPerformanceWithin(elapsed, 5, 'Query by type')
      expect(functions.length).toBeGreaterThan(0)
    })

    it('should query nodes by file in O(1)', () => {
      const { elapsed, result: nodes } = measurePerformance(() => codeMap.getNodesByFile('file0.js'))
      expectPerformanceWithin(elapsed, 5, 'Query by file')
      expect(nodes.length).toBeGreaterThan(0)
    })

    it('should return correct node counts', () => {
      const functions = codeMap.getNodesByType('function')
      const classes = codeMap.getNodesByType('class')
      expect(functions.length + classes.length).toBe(50)
    })
  })

  describe('edge operations', () => {
    beforeEach(() => {
      codeMap.addNode(createTestNode({ id: 'fn1', name: 'func1', file: 'a.js' }))
      codeMap.addNode(createTestNode({ id: 'fn2', name: 'func2', file: 'b.js' }))
    })

    it('should add edges', () => {
      codeMap.addEdge(createTestEdge({ id: 'e1', source: 'fn1', target: 'fn2' }))
      expect(codeMap.getEdges().length).toBe(1)
    })

    it('should retrieve edges by source', () => {
      codeMap.addEdge(createTestEdge({ id: 'e1', source: 'fn1', target: 'fn2' }))
      const edges = codeMap.getEdgesBySource('fn1')
      expect(edges.length).toBe(1)
      expect(edges[0].target).toBe('fn2')
    })

    it('should retrieve edges by target', () => {
      codeMap.addEdge(createTestEdge({ id: 'e1', source: 'fn1', target: 'fn2' }))
      const edges = codeMap.getEdgesByTarget('fn2')
      expect(edges.length).toBe(1)
      expect(edges[0].source).toBe('fn1')
    })
  })

  describe('blast radius (adjacency cache)', () => {
    beforeEach(() => {
      for (let i = 0; i < 10; i++) {
        codeMap.addNode(createTestNode({ id: `fn${i}`, name: `function${i}`, file: `file${i}.js` }))
      }

      for (let i = 0; i < 9; i++) {
        codeMap.addEdge(createTestEdge({ id: `e${i}`, source: `fn${i}`, target: `fn${i + 1}` }))
      }
    })

    it('should calculate blast radius efficiently', () => {
      const { elapsed, result: radius } = measurePerformance(() =>
        codeMap.calculateBlastRadiusDetailed('fn5', 2)
      )
      expectPerformanceWithin(elapsed, 50, 'Blast radius calculation')
      expect(radius.size).toBeGreaterThan(0)
    })

    it('should cache blast radius results', () => {
      codeMap.calculateBlastRadiusDetailed('fn5', 2)
      const { elapsed } = measurePerformance(() => codeMap.calculateBlastRadiusDetailed('fn5', 2))
      expectPerformanceWithin(elapsed, 10, 'Cached blast radius')
    })

    it('should find imported nodes', () => {
      codeMap.addEdge(createTestEdge({ id: 'e_import', source: 'fn0', target: 'fn5', type: 'imports' }))
      const radius = codeMap.calculateBlastRadiusDetailed('fn5', 1)
      expect(radius.has('fn0')).toBe(true)
    })
  })

  describe('serialization', () => {
    it('should serialize to JSON', () => {
      codeMap.addNode(createTestNode({ id: 'fn1', name: 'func1' }))
      const json = codeMap.toJSON()
      expect(json).toBeDefined()
      expect(typeof json).toBe('string')
    })

    it('should deserialize from JSON', () => {
      codeMap.addNode(createTestNode({ id: 'fn1', name: 'func1' }))
      const json = codeMap.toJSON()
      const newMap = CodeMap.fromJSON(json)
      expect(newMap.getNodes().length).toBe(1)
    })
  })

  describe('performance with large codebases', () => {
    it('should handle 1000 nodes efficiently', () => {
      for (let i = 0; i < 1000; i++) {
        codeMap.addNode(
          createTestNode({
            id: `fn${i}`,
            name: `function${i}`,
            type: (['function', 'class', 'variable'] as const)[i % 3],
            file: `file${i % 100}.js`,
          })
        )
      }
      expect(codeMap.getNodes().length).toBe(1000)
    })

    it('should query 1000 nodes in <1ms', () => {
      for (let i = 0; i < 1000; i++) {
        codeMap.addNode(
          createTestNode({
            id: `fn${i}`,
            name: `function${i}`,
            type: i % 2 === 0 ? 'function' : 'class',
            file: `file${i % 50}.js`,
          })
        )
      }

      const { elapsed, result: functions } = measurePerformance(() => codeMap.getNodesByType('function'))
      expectPerformanceWithin(elapsed, 1, 'Query 1000 nodes')
      expect(functions.length).toBeGreaterThan(0)
    })

    it('should add 1000 edges efficiently', () => {
      for (let i = 0; i < 100; i++) {
        codeMap.addNode(createTestNode({ id: `fn${i}`, name: `function${i}` }))
      }

      const { elapsed } = measurePerformance(() => {
        for (let i = 0; i < 1000; i++) {
          codeMap.addEdge(
            createTestEdge({ id: `e${i}`, source: `fn${i % 100}`, target: `fn${(i + 1) % 100}` })
          )
        }
      })

      expectPerformanceWithin(elapsed, 100, 'Add 1000 edges')
      expect(codeMap.getEdges().length).toBe(1000)
    })
  })

  describe('edge cases', () => {
    it('should handle circular references', () => {
      codeMap.addNode(createTestNode({ id: 'fn1', name: 'func1' }))
      codeMap.addEdge(createTestEdge({ id: 'e1', source: 'fn1', target: 'fn1' }))
      expect(codeMap.getEdges().length).toBe(1)
    })

    it('should handle duplicate node additions', () => {
      const node = createTestNode({ id: 'fn1', name: 'func1' })
      codeMap.addNode(node)
      codeMap.addNode(node)
      expect(codeMap.getNodes().length).toBe(1)
    })
  })
})
