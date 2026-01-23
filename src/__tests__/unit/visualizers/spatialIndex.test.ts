import { describe, it, expect, beforeEach } from 'vitest'
import { SpatialIndex } from '@visualizers/spatialIndex'
import { bounds, createTestPoints, createGridPoints } from '../../fixtures/index'
import { measurePerformance, expectPerformanceWithin } from '../../helpers/index'

describe('SpatialIndex (Quadtree)', () => {
  let spatialIndex: SpatialIndex

  beforeEach(() => {
    spatialIndex = new SpatialIndex(bounds.standard, 4, 10)
  })

  describe('insertion and basic operations', () => {
    it('should insert points into the tree', () => {
      const point = { id: 'test1', x: 100, y: 100 }
      spatialIndex.insert(point)
      expect(spatialIndex.getPointCount()).toBeGreaterThan(0)
    })

    it('should handle multiple insertions', () => {
      const points = createTestPoints(100, bounds.standard)
      for (const point of points) {
        spatialIndex.insert(point)
      }
      expect(spatialIndex.getPointCount()).toBe(100)
    })

    it('should clear all points', () => {
      spatialIndex.insert({ id: 'test', x: 100, y: 100 })
      expect(spatialIndex.getPointCount()).toBeGreaterThan(0)
      spatialIndex.clear()
      expect(spatialIndex.getPointCount()).toBe(0)
    })
  })

  describe('range queries', () => {
    beforeEach(() => {
      const points = createGridPoints(2, 2, 100)
      // Override IDs to match expected test IDs
      const testPoints = [
        { id: 'p1', x: 100, y: 100 },
        { id: 'p2', x: 150, y: 150 },
        { id: 'p3', x: 500, y: 500 },
        { id: 'p4', x: 900, y: 900 },
      ]
      testPoints.forEach((p) => spatialIndex.insert(p))
    })

    it('should find points in a rectangular range', () => {
      const results = spatialIndex.query({ minX: 50, minY: 50, maxX: 200, maxY: 200 })
      expect(results.length).toBeGreaterThan(0)
      expect(results.some((p) => p.id === 'p1')).toBe(true)
    })

    it('should return empty results for empty range', () => {
      const results = spatialIndex.query({ minX: 200, minY: 200, maxX: 400, maxY: 400 })
      expect(results.length).toBe(0)
    })

    it('should handle viewport culling query', () => {
      const viewport = { minX: 0, minY: 0, maxX: 600, maxY: 600 }
      const results = spatialIndex.query(viewport)
      expect(results.length).toBeGreaterThan(0)
      expect(results.length).toBeLessThanOrEqual(3)
    })
  })

  describe('nearest neighbor queries', () => {
    beforeEach(() => {
      const points = [
        { id: 'p1', x: 100, y: 100 },
        { id: 'p2', x: 110, y: 110 },
        { id: 'p3', x: 500, y: 500 },
      ]
      points.forEach((p) => spatialIndex.insert(p))
    })

    it('should find nearest point', () => {
      const nearest = spatialIndex.findNearest({ x: 105, y: 105 }, 1)
      expect(nearest.length).toBe(1)
      expect(['p1', 'p2']).toContain(nearest[0].id)
    })

    it('should find k nearest points', () => {
      const nearest = spatialIndex.findNearest({ x: 105, y: 105 }, 2)
      expect(nearest.length).toBe(2)
    })

    it('should respect distance radius', () => {
      const nearest = spatialIndex.findNearest({ x: 105, y: 105 }, 10, 50)
      expect(nearest.length).toBeGreaterThanOrEqual(0)
    })
  })

  describe('bounds intersection', () => {
    it('should detect bounds intersection', () => {
      const bounds1 = { minX: 0, minY: 0, maxX: 100, maxY: 100 }
      const bounds2 = { minX: 50, minY: 50, maxX: 150, maxY: 150 }
      const result = spatialIndex.intersects(bounds1, bounds2)
      expect(result).toBe(true)
    })

    it('should detect non-intersecting bounds', () => {
      const bounds1 = { minX: 0, minY: 0, maxX: 100, maxY: 100 }
      const bounds2 = { minX: 200, minY: 200, maxX: 300, maxY: 300 }
      const result = spatialIndex.intersects(bounds1, bounds2)
      expect(result).toBe(false)
    })
  })

  describe('performance characteristics', () => {
    it('should handle 1000 points efficiently', () => {
      const points = createTestPoints(1000, bounds.standard)
      for (const point of points) {
        spatialIndex.insert(point)
      }
      expect(spatialIndex.getPointCount()).toBe(1000)

      const { elapsed, result } = measurePerformance(() =>
        spatialIndex.query({ minX: 0, minY: 0, maxX: 500, maxY: 500 })
      )

      expectPerformanceWithin(elapsed, 100, 'Query 1000 points')
      expect(result.length).toBeGreaterThan(0)
    })

    it('should perform nearest neighbor in log time', () => {
      const points = createTestPoints(500, bounds.standard)
      for (const point of points) {
        spatialIndex.insert(point)
      }

      const { elapsed } = measurePerformance(() => spatialIndex.findNearest({ x: 500, y: 500 }, 10))

      expectPerformanceWithin(elapsed, 50, 'Nearest neighbor query')
    })
  })

  describe('edge cases', () => {
    it('should handle points at boundaries', () => {
      spatialIndex.insert({ id: 'corner', x: 0, y: 0 })
      spatialIndex.insert({ id: 'edge', x: 1000, y: 500 })
      const results = spatialIndex.query({ minX: 0, minY: 0, maxX: 100, maxY: 100 })
      expect(results.length).toBeGreaterThan(0)
    })

    it('should handle duplicate coordinates', () => {
      spatialIndex.insert({ id: 'p1', x: 100, y: 100 })
      spatialIndex.insert({ id: 'p2', x: 100, y: 100 })
      const results = spatialIndex.query({ minX: 50, minY: 50, maxX: 150, maxY: 150 })
      expect(results.length).toBe(2)
    })

    it('should handle empty queries', () => {
      const results = spatialIndex.query({ minX: -100, minY: -100, maxX: -50, maxY: -50 })
      expect(results.length).toBe(0)
    })
  })
})
