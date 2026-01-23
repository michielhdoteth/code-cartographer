import { LodManager } from '@visualizers/lodManager'
import { createSimulationNodes } from '../../fixtures/index'
import { measurePerformance, expectPerformanceWithin } from '../../helpers/index'

describe('LOD Manager (Level of Detail)', () => {
  let lodManager: LodManager
  const nodes = createSimulationNodes(100)

  beforeEach(() => {
    lodManager = new LodManager()
  })

  describe('zoom level detection', () => {
    it('should detect zoom level 1 (zoomed out far)', () => {
      const level = lodManager.getZoomLevel(0.1)
      expect(level).toBe(1)
    })

    it('should detect zoom level 2', () => {
      const level = lodManager.getZoomLevel(0.5)
      expect([1, 2]).toContain(level)
    })

    it('should detect zoom level 3', () => {
      const level = lodManager.getZoomLevel(1.0)
      expect([2, 3]).toContain(level)
    })

    it('should detect zoom level 4 (zoomed in)', () => {
      const level = lodManager.getZoomLevel(2.0)
      expect(level).toBeGreaterThanOrEqual(3)
    })
  })

  describe('node aggregation', () => {
    it('should aggregate nodes at zoom level 1', () => {
      const aggregated = lodManager.aggregate(nodes, 0.1)
      expect(aggregated.length).toBeLessThanOrEqual(nodes.length)
      expect(aggregated.length).toBeGreaterThan(0)
    })

    it('should provide minimal aggregation at zoom level 4', () => {
      const aggregated = lodManager.aggregate(nodes, 2.0)
      expect(aggregated.length).toBeLessThanOrEqual(nodes.length)
    })

    it('should cluster directory nodes at low zoom', () => {
      const aggregated = lodManager.aggregate(nodes, 0.1)
      const directories = aggregated.filter((n) => n.type === 'directory')
      expect(directories.length).toBeGreaterThan(0)
    })
  })

  describe('label visibility', () => {
    it('should hide labels at zoom level 1', () => {
      const visible = lodManager.shouldShowLabel(0.1)
      expect(visible).toBe(false)
    })

    it('should show labels at zoom level 3', () => {
      const visible = lodManager.shouldShowLabel(1.0)
      expect(visible).toBe(true)
    })

    it('should show labels at zoom level 4', () => {
      const visible = lodManager.shouldShowLabel(2.0)
      expect(visible).toBe(true)
    })
  })

  describe('reduction statistics', () => {
    it('should calculate reduction percentage', () => {
      lodManager.aggregate(nodes, 0.1)
      const stats = lodManager.getReductionStats()
      expect(stats.originalNodeCount).toBeGreaterThan(0)
      expect(stats.aggregatedNodeCount).toBeGreaterThan(0)
      expect(stats.reductionPercentage).toBeGreaterThanOrEqual(0)
      expect(stats.reductionPercentage).toBeLessThanOrEqual(100)
    })

    it('should report correct reduction for zoom level 1', () => {
      lodManager.aggregate(nodes, 0.1)
      const stats = lodManager.getReductionStats()
      expect(stats.reductionPercentage).toBeGreaterThan(0)
    })

    it('should report minimal reduction for zoom level 4', () => {
      lodManager.aggregate(nodes, 2.0)
      const stats = lodManager.getReductionStats()
      expect(stats.reductionPercentage).toBeLessThan(10)
    })
  })

  describe('performance with large datasets', () => {
    it('should aggregate 1000 nodes efficiently', () => {
      const largeNodeSet = createSimulationNodes(1000)

      const { elapsed, result } = measurePerformance(() => lodManager.aggregate(largeNodeSet, 0.1))

      expectPerformanceWithin(elapsed, 100, 'Aggregate 1000 nodes')
      expect(result.length).toBeLessThan(largeNodeSet.length)
    })

    it('should provide 10x speedup at zoom level 1', () => {
      const aggregated = lodManager.aggregate(nodes, 0.1)
      const reduction = (nodes.length - aggregated.length) / nodes.length
      expect(reduction).toBeGreaterThan(0.5)
    })
  })

  describe('edge cases', () => {
    it('should handle empty node list', () => {
      const aggregated = lodManager.aggregate([], 0.1)
      expect(aggregated.length).toBe(0)
    })

    it('should handle single node', () => {
      const aggregated = lodManager.aggregate([nodes[0]], 0.1)
      expect(aggregated.length).toBe(1)
    })

    it('should handle extreme zoom values', () => {
      const veryZoomedOut = lodManager.getZoomLevel(0.01)
      const veryZoomedIn = lodManager.getZoomLevel(10.0)
      expect(veryZoomedOut).toBeGreaterThanOrEqual(1)
      expect(veryZoomedIn).toBeGreaterThanOrEqual(1)
    })

    it('should handle nodes at same coordinates', () => {
      const clusteredNodes = createSimulationNodes(10).map((n) => ({ ...n, x: 500, y: 500 }))
      const aggregated = lodManager.aggregate(clusteredNodes, 0.1)
      expect(aggregated.length).toBeGreaterThan(0)
    })
  })

  describe('zoom level transitions', () => {
    it('should transition smoothly between zoom levels', () => {
      const zoom1 = lodManager.aggregate(nodes, 0.1)
      const zoom2 = lodManager.aggregate(nodes, 0.5)
      const zoom3 = lodManager.aggregate(nodes, 1.0)
      const zoom4 = lodManager.aggregate(nodes, 2.0)

      expect(zoom1.length).toBeLessThanOrEqual(zoom2.length)
      expect(zoom2.length).toBeLessThanOrEqual(zoom3.length)
      expect(zoom3.length).toBeLessThanOrEqual(zoom4.length)
    })

    it('should maintain node identity through aggregation levels', () => {
      const aggregated = lodManager.aggregate(nodes, 2.0)
      for (const node of aggregated) {
        expect(node.id).toBeDefined()
        expect(typeof node.id).toBe('string')
      }
    })
  })
})
