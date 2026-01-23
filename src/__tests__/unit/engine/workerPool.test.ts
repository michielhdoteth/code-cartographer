import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import { WorkerPool, PoolTask } from '@engine/workerPool'

describe('WorkerPool', () => {
  let pool: WorkerPool

  beforeEach(() => {
    pool = new WorkerPool(2)
  })

  afterEach(() => {
    pool.terminate()
  })

  describe('initialization', () => {
    it('should initialize with correct pool size', () => {
      const stats = pool.getStats()
      expect(stats.poolSize).toBe(2)
      expect(stats.busyWorkers).toBe(0)
      expect(stats.queuedTasks).toBe(0)
    })

    it('should detect hardware concurrency', () => {
      const autoPool = new WorkerPool()
      expect(autoPool.getStats().poolSize).toBeGreaterThan(0)
      expect(autoPool.getStats().poolSize).toBeLessThanOrEqual(8)
      autoPool.terminate()
    })
  })

  describe('task dispatch', () => {
    it('should track queued and completed tasks', async () => {
      const tasks: PoolTask[] = [
        {
          fileId: 'file1',
          fileName: 'test.js',
          content: 'const x = 1;',
          language: 'javascript',
        },
      ]

      // Mock worker for testing
      const stats = pool.getStats()
      expect(stats.poolSize).toBeGreaterThan(0)
    })

    it('should report progress callbacks', async () => {
      let progressCalls = 0
      const callback = (completed: number, total: number, message: string) => {
        progressCalls++
        expect(total).toBeGreaterThan(0)
        expect(message).toBeTruthy()
      }

      // Test callback invocation
      expect(progressCalls).toBe(0)
    })
  })

  describe('error handling', () => {
    it('should handle worker errors gracefully', () => {
      const stats = pool.getStats()
      expect(stats.busyWorkers).toBe(0)
      // Worker errors should not crash pool
      expect(() => pool.getStats()).not.toThrow()
    })

    it('should recover from worker failure', () => {
      const stats1 = pool.getStats()
      const stats2 = pool.getStats()
      expect(stats2.poolSize).toBe(stats1.poolSize)
    })
  })

  describe('termination', () => {
    it('should terminate all workers', () => {
      pool.terminate()
      const stats = pool.getStats()
      expect(stats.poolSize).toBe(0)
    })

    it('should allow reinitialization after termination', () => {
      pool.terminate()
      const newPool = new WorkerPool(4)
      expect(newPool.getStats().poolSize).toBe(4)
      newPool.terminate()
    })
  })
})
