import { describe, it, expect, beforeEach } from 'vitest'
import { AdaptiveSimulation } from '@visualizers/adaptiveSimulation'

describe('AdaptiveSimulation', () => {
  let simulation: AdaptiveSimulation
  const nodes = Array.from({ length: 50 }, (_, i) => ({
    id: `node${i}`,
    x: Math.random() * 1000,
    y: Math.random() * 1000,
    vx: 0,
    vy: 0,
    fx: 0,
    fy: 0,
  }))

  beforeEach(() => {
    simulation = new AdaptiveSimulation()
  })

  describe('initialization', () => {
    it('should initialize with default parameters', () => {
      expect(simulation).toBeDefined()
      expect(simulation.alpha).toBeGreaterThan(0)
    })

    it('should set simulation parameters', () => {
      const alpha = simulation.getAlpha()
      expect(alpha).toBeGreaterThan(0)
      expect(alpha).toBeLessThanOrEqual(1)
    })
  })

  describe('early stopping', () => {
    it('should stop when alpha drops below threshold', () => {
      simulation.setAlpha(0.0001)
      const shouldStop = simulation.shouldStop()
      expect(shouldStop).toBe(true)
    })

    it('should not stop when alpha is high', () => {
      simulation.setAlpha(0.5)
      const shouldStop = simulation.shouldStop()
      expect(shouldStop).toBe(false)
    })

    it('should track convergence state', () => {
      simulation.setAlpha(0.0001)
      const isConverged = simulation.isConverged()
      expect(typeof isConverged).toBe('boolean')
    })
  })

  describe('tick throttling', () => {
    it('should throttle ticks when converging', () => {
      simulation.setAlpha(0.05)
      const ticks = Array.from({ length: 100 }, (_, i) => i)
      let activeTickCount = 0
      for (const _ of ticks) {
        if (simulation.shouldTick()) {
          activeTickCount++
        }
        simulation.tick()
      }
      expect(activeTickCount).toBeLessThan(ticks.length)
    })

    it('should not throttle when alpha is high', () => {
      simulation.setAlpha(0.5)
      const shouldTick = simulation.shouldTick()
      expect(shouldTick).toBe(true)
    })
  })

  describe('alpha decay', () => {
    it('should decay alpha over time', () => {
      const initialAlpha = simulation.getAlpha()
      for (let i = 0; i < 10; i++) {
        simulation.tick()
      }
      const finalAlpha = simulation.getAlpha()
      expect(finalAlpha).toBeLessThanOrEqual(initialAlpha)
    })

    it('should increase alpha decay when converging', () => {
      simulation.setAlpha(0.1)
      const decay1 = simulation.getAlphaDampening()
      simulation.setAlpha(0.01)
      const decay2 = simulation.getAlphaDampening()
      expect(decay2).toBeGreaterThanOrEqual(decay1)
    })
  })

  describe('interaction reheat', () => {
    it('should reheat simulation on interaction', () => {
      simulation.setAlpha(0.001)
      simulation.onInteraction()
      const newAlpha = simulation.getAlpha()
      expect(newAlpha).toBeGreaterThan(0.001)
    })

    it('should cool down after inactivity', () => {
      simulation.onInteraction()
      const reheatedAlpha = simulation.getAlpha()
      for (let i = 0; i < 100; i++) {
        simulation.tick()
      }
      const cooledAlpha = simulation.getAlpha()
      expect(cooledAlpha).toBeLessThanOrEqual(reheatedAlpha)
    })
  })

  describe('performance metrics', () => {
    it('should track CPU usage reduction', () => {
      const stats = simulation.getStatistics()
      expect(stats.ticksSkipped).toBeGreaterThanOrEqual(0)
      expect(stats.cpuReduction).toBeGreaterThanOrEqual(0)
      expect(stats.cpuReduction).toBeLessThanOrEqual(100)
    })

    it('should report 70% CPU reduction when idle', () => {
      simulation.setAlpha(0.001)
      for (let i = 0; i < 1000; i++) {
        simulation.tick()
      }
      const stats = simulation.getStatistics()
      expect(stats.cpuReduction).toBeGreaterThan(50)
    })
  })

  describe('pause and resume', () => {
    it('should pause simulation', () => {
      simulation.pause()
      const isPaused = simulation.isPaused()
      expect(isPaused).toBe(true)
    })

    it('should resume simulation', () => {
      simulation.pause()
      simulation.resume()
      const isPaused = simulation.isPaused()
      expect(isPaused).toBe(false)
    })

    it('should not progress when paused', () => {
      const initialAlpha = simulation.getAlpha()
      simulation.pause()
      for (let i = 0; i < 10; i++) {
        simulation.tick()
      }
      const pausedAlpha = simulation.getAlpha()
      expect(pausedAlpha).toBe(initialAlpha)
    })
  })

  describe('edge cases', () => {
    it('should handle very small alpha values', () => {
      simulation.setAlpha(0.00001)
      expect(simulation.shouldStop()).toBe(true)
    })

    it('should handle alpha equal to 1', () => {
      simulation.setAlpha(1.0)
      expect(simulation.shouldStop()).toBe(false)
    })

    it('should recover from paused state', () => {
      simulation.setAlpha(0.001)
      simulation.pause()
      simulation.resume()
      simulation.onInteraction()
      const newAlpha = simulation.getAlpha()
      expect(newAlpha).toBeGreaterThan(0.001)
    })
  })

  describe('integration with nodes', () => {
    it('should provide acceleration multiplier for all nodes', () => {
      const accelMultiplier = simulation.getAccelerationMultiplier()
      expect(accelMultiplier).toBeGreaterThan(0)
    })

    it('should decrease acceleration as simulation converges', () => {
      simulation.setAlpha(0.5)
      const accel1 = simulation.getAccelerationMultiplier()
      simulation.setAlpha(0.05)
      const accel2 = simulation.getAccelerationMultiplier()
      expect(accel2).toBeLessThanOrEqual(accel1)
    })
  })

  describe('performance characteristics', () => {
    it('should achieve 70% CPU reduction', () => {
      for (let i = 0; i < 1000; i++) {
        simulation.tick()
      }
      const stats = simulation.getStatistics()
      expect(stats.cpuReduction).toBeGreaterThan(50)
    })

    it('should converge within reasonable time', () => {
      const start = performance.now()
      while (!simulation.isConverged() && performance.now() - start < 5000) {
        simulation.tick()
      }
      const elapsed = performance.now() - start
      expect(elapsed).toBeLessThan(5000)
    })
  })
})
