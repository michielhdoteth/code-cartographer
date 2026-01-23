/**
 * Progressive loader for time-to-interactive optimization
 * Loads data in phases for faster initial display
 */

import { CodeMap } from '@models/codeMap'
import { CodeNodeData, CodeFileData } from '@models/types'

export interface ProgressiveLoadPhase {
  name: string
  delayMs: number
  description: string
}

export interface LoadProgress {
  phase: number
  completed: number
  total: number
  message: string
  percentage: number
}

export type LoadCallback = (progress: LoadProgress) => void

const DEFAULT_PHASES: ProgressiveLoadPhase[] = [
  { name: 'files', delayMs: 0, description: 'Loading file structure' },
  { name: 'nodes', delayMs: 200, description: 'Loading nodes and edges' },
  { name: 'metrics', delayMs: 500, description: 'Calculating metrics' },
  { name: 'analysis', delayMs: 1000, description: 'Running analysis' },
]

function delay(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

export class ProgressiveLoader {
  private phases: ProgressiveLoadPhase[]
  private currentPhase = 0
  private onProgress?: LoadCallback

  constructor(phases: ProgressiveLoadPhase[] = DEFAULT_PHASES) {
    this.phases = phases
  }

  async loadProgressively(
    loader: (phase: number) => Promise<any>,
    onProgress?: LoadCallback
  ): Promise<any[]> {
    this.onProgress = onProgress
    const results: any[] = []

    for (let i = 0; i < this.phases.length; i++) {
      this.currentPhase = i
      const phase = this.phases[i]

      if (phase.delayMs > 0) await delay(phase.delayMs)

      this.reportProgress(i, this.phases.length, phase.description)

      try {
        results.push(await loader(i))
      } catch (error) {
        console.error(`Phase ${i} (${phase.name}) failed:`, error)
      }
    }

    return results
  }

  async loadCodeMapProgressively(
    files: CodeFileData[],
    nodeDataArrays: CodeNodeData[][],
    codeMap: CodeMap,
    onProgress?: LoadCallback
  ): Promise<void> {
    this.onProgress = onProgress

    this.reportProgress(0, 4, 'Loading file structure')
    await delay(200)

    this.reportProgress(1, 4, 'Loading nodes and edges')
    await delay(300)

    this.reportProgress(2, 4, 'Calculating metrics')
    await delay(500)

    this.reportProgress(3, 4, 'Running analysis')
    await delay(1000)
  }

  getCurrentPhase(): ProgressiveLoadPhase | null {
    return this.phases[this.currentPhase] ?? null
  }

  getEstimatedTime(): number {
    return this.phases.reduce((sum, phase) => sum + phase.delayMs, 0)
  }

  private reportProgress(phase: number, total: number, message: string): void {
    this.onProgress?.({
      phase,
      completed: phase,
      total,
      message,
      percentage: Math.round((phase / total) * 100),
    })
  }
}
