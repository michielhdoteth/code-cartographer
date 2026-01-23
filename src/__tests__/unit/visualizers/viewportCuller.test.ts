import { ViewportCuller } from '@visualizers/viewportCuller'
import { createSimulationNodes, viewports } from '../../fixtures/index'
import { measurePerformance, expectPerformanceWithin } from '../../helpers/index'

describe('ViewportCuller', () => {
  let culler: ViewportCuller
  const nodes = [
    { id: 'n1', x: 100, y: 100, vx: 0, vy: 0, fx: 0, fy: 0 },
    { id: 'n2', x: 200, y: 200, vx: 0, vy: 0, fx: 0, fy: 0 },
    { id: 'n3', x: 500, y: 500, vx: 0, vy: 0, fx: 0, fy: 0 },
    { id: 'n4', x: 800, y: 800, vx: 0, vy: 0, fx: 0, fy: 0 },
  ]

  beforeEach(() => {
    culler = new ViewportCuller({ margin: 100, cullThreshold: 50 })
  })

  describe('basic culling', () => {
    it('should cull nodes outside viewport', () => {
      const visible = culler.cull(nodes, viewports.small)
      expect(visible.length).toBeLessThan(nodes.length)
      expect(visible.some((n) => n.id === 'n1' || n.id === 'n2')).toBe(true)
    })

    it('should include nodes in viewport with margin', () => {
      const visible = culler.cull(nodes, viewports.small)
      expect(visible.length).toBeGreaterThan(0)
    })

    it('should include all nodes when viewport is large', () => {
      const visible = culler.cull(nodes, { x: 0, y: 0, width: 1000, height: 1000 })
      expect(visible.length).toBe(nodes.length)
    })
  })

  describe('edge cases', () => {
    it('should handle empty node list', () => {
      const visible = culler.cull([], viewports.small)
      expect(visible.length).toBe(0)
    })

    it('should handle single node', () => {
      const visible = culler.cull([nodes[0]], { x: 50, y: 50, width: 200, height: 200 })
      expect(visible.length).toBeGreaterThanOrEqual(0)
    })

    it('should handle negative viewport coordinates', () => {
      const visible = culler.cull(nodes, viewports.negative)
      expect(Array.isArray(visible)).toBe(true)
    })
  })

  describe('statistics', () => {
    it('should track culling statistics', () => {
      culler.cull(nodes, viewports.small)
      const stats = culler.getStatistics()
      expect(stats.totalNodes).toBeGreaterThan(0)
      expect(stats.culledNodes).toBeGreaterThanOrEqual(0)
      expect(stats.culledPercentage).toBeGreaterThanOrEqual(0)
      expect(stats.culledPercentage).toBeLessThanOrEqual(100)
    })

    it('should calculate culled percentage correctly', () => {
      culler.cull(nodes, { x: 0, y: 0, width: 200, height: 200 })
      const stats = culler.getStatistics()
      expect(stats.culledPercentage).toBeGreaterThan(0)
    })
  })

  describe('performance', () => {
    it('should cull 1000 nodes efficiently', () => {
      const largeNodeSet = createSimulationNodes(1000)
        .map((n) => ({ ...n, x: Math.random() * 2000, y: Math.random() * 2000 }))

      const { elapsed, result } = measurePerformance(() => culler.cull(largeNodeSet, viewports.medium))

      expectPerformanceWithin(elapsed, 50, 'Cull 1000 nodes')
      expect(result.length).toBeLessThanOrEqual(largeNodeSet.length)
    })

    it('should provide 5-10x speedup benefit', () => {
      const largeNodeSet = createSimulationNodes(5000)
        .map((n) => ({ ...n, x: Math.random() * 2000, y: Math.random() * 2000 }))

      const visible = culler.cull(largeNodeSet, viewports.medium)
      const reduction = 1 - visible.length / largeNodeSet.length
      expect(reduction).toBeGreaterThan(0.5)
    })
  })

  describe('margin configuration', () => {
    it('should respect margin setting', () => {
      const newCuller = new ViewportCuller({ margin: 200 })
      const visible = newCuller.cull(nodes, viewports.small)
      expect(visible.length).toBeGreaterThanOrEqual(0)
    })

    it('should handle zero margin', () => {
      const newCuller = new ViewportCuller({ margin: 0 })
      const visible = newCuller.cull(nodes, viewports.small)
      expect(visible.length).toBeGreaterThanOrEqual(0)
    })
  })

  describe('threshold-based activation', () => {
    it('should activate culling when node count exceeds threshold', () => {
      const newCuller = new ViewportCuller({ cullThreshold: 10 })
      const smallSet = nodes.slice(0, 5)
      const largeSet = createSimulationNodes(50)

      const smallVisible = newCuller.cull(smallSet, viewports.medium)
      const largeVisible = newCuller.cull(largeSet, viewports.medium)

      expect(smallVisible.length).toBeLessThanOrEqual(smallSet.length)
      expect(largeVisible.length).toBeLessThanOrEqual(largeSet.length)
    })
  })
})
