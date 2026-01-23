/**
 * Worker pool manager for parallel parsing
 * Manages workers and distributes parsing tasks with progress reporting
 */

import { CodeNodeData, Language } from '@models/types'
import { CodeEdgeData } from '@engine/cacheManager'

export interface PoolTask {
  fileId: string
  fileName: string
  content: string
  language: Language
}

export interface PoolResult {
  fileId: string
  success: boolean
  nodes?: CodeNodeData[]
  edges?: CodeEdgeData[]
  error?: string
}

export type ProgressCallback = (completed: number, total: number, message: string) => void

interface WorkerWrapper {
  worker: Worker
  busy: boolean
  currentTaskId?: string
}

const DEFAULT_POOL_SIZE = Math.min(navigator.hardwareConcurrency || 4, 8)

export class WorkerPool {
  private workers: WorkerWrapper[] = []
  private taskQueue: PoolTask[] = []
  private results: Map<string, PoolResult> = new Map()
  private resolvers: Map<string, (result: PoolResult) => void> = new Map()
  private poolSize: number
  private initialized = false
  private onProgress?: ProgressCallback
  private totalTasks = 0

  constructor(poolSize: number = DEFAULT_POOL_SIZE) {
    this.poolSize = poolSize
  }

  async initialize(workerUrl: string): Promise<void> {
    if (this.initialized) return

    for (let i = 0; i < this.poolSize; i++) {
      const wrapper: WorkerWrapper = { worker: new Worker(workerUrl), busy: false }

      wrapper.worker.onmessage = (event: MessageEvent<PoolResult>) => {
        this.handleWorkerResult(event.data, wrapper)
      }

      wrapper.worker.onerror = (error: ErrorEvent) => {
        console.error(`Worker ${i} error:`, error.message)
        this.handleWorkerError(wrapper)
      }

      this.workers.push(wrapper)
    }

    this.initialized = true
  }

  private handleWorkerError(wrapper: WorkerWrapper): void {
    wrapper.busy = false
    if (wrapper.currentTaskId) {
      const task = this.taskQueue.find((t) => t.fileId === wrapper.currentTaskId)
      if (task) this.taskQueue.unshift(task)
    }
    this.processQueue()
  }

  async dispatch(tasks: PoolTask[], onProgress?: ProgressCallback): Promise<PoolResult[]> {
    if (!this.initialized) {
      throw new Error('Worker pool not initialized. Call initialize() first.')
    }

    this.taskQueue = [...tasks]
    this.totalTasks = tasks.length
    this.results.clear()
    this.onProgress = onProgress

    return new Promise((resolve) => {
      this.processQueue()

      const checkCompletion = setInterval(() => {
        if (this.results.size === tasks.length) {
          clearInterval(checkCompletion)
          resolve(tasks.map((t) => this.results.get(t.fileId)!))
        }
      }, 50)
    })
  }

  private processQueue(): void {
    for (const worker of this.workers) {
      if (worker.busy || this.taskQueue.length === 0) continue

      const task = this.taskQueue.shift()!
      worker.busy = true
      worker.currentTaskId = task.fileId
      worker.worker.postMessage(task)

      this.onProgress?.(this.results.size, this.totalTasks, `Parsing: ${task.fileName}`)
    }
  }

  private handleWorkerResult(result: PoolResult, wrapper: WorkerWrapper): void {
    wrapper.busy = false
    wrapper.currentTaskId = undefined
    this.results.set(result.fileId, result)

    const resolver = this.resolvers.get(result.fileId)
    if (resolver) {
      resolver(result)
      this.resolvers.delete(result.fileId)
    }

    this.processQueue()
  }

  terminate(): void {
    this.workers.forEach((w) => w.worker.terminate())
    this.workers = []
    this.initialized = false
  }

  getStats(): { poolSize: number; busyWorkers: number; queuedTasks: number; completedTasks: number } {
    return {
      poolSize: this.poolSize,
      busyWorkers: this.workers.filter((w) => w.busy).length,
      queuedTasks: this.taskQueue.length,
      completedTasks: this.results.size,
    }
  }
}
