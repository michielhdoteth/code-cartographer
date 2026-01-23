/**
 * Git churn analyzer for identifying code hotspots
 * Analyzes commit history to find frequently changed files
 */

import type { ChurnMetrics } from '@models/types'
import { execGit, getRepositoryStatus } from '@utils/gitUtils.js'

// =============================================================================
// Types
// =============================================================================

export type ChurnLevel = 'critical' | 'high' | 'medium' | 'low' | 'none'

export interface ChurnHotspot {
  file: string
  metrics: ChurnMetrics
  score: number
}

type ProgressCallback = (percent: number, message: string) => void

// =============================================================================
// Churn Analyzer
// =============================================================================

export class ChurnAnalyzer {
  /**
   * Parse git log to calculate churn metrics per file
   */
  async analyzeChurn(
    repoPath: string,
    authorFilter?: string,
    onProgress?: ProgressCallback
  ): Promise<Map<string, ChurnMetrics>> {
    onProgress?.(0, 'Analyzing git churn...')

    const output = await this.fetchGitLog(repoPath, authorFilter, onProgress)
    const churnMap = this.parseGitLogOutput(output, onProgress)

    onProgress?.(100, 'Churn analysis complete')
    return churnMap
  }

  private async fetchGitLog(
    repoPath: string,
    authorFilter?: string,
    onProgress?: ProgressCallback
  ): Promise<string> {
    const args = ['log', '--numstat', '--follow', '--pretty=format:%H %aI']

    if (authorFilter) {
      args.push(`--author=${authorFilter.replace(/['"]/g, '')}`)
    }

    return execGit(args, {
      cwd: repoPath,
      onProgress: (percent, msg) => onProgress?.(Math.min(50, percent / 2), msg),
    })
  }

  private parseGitLogOutput(output: string, onProgress?: ProgressCallback): Map<string, ChurnMetrics> {
    const churnMap = new Map<string, ChurnMetrics>()
    const lines = output.split('\n')

    let currentDate = new Date()
    let lineIndex = 0

    for (const line of lines) {
      if (!line.trim()) continue

      if (lineIndex++ % 100 === 0) {
        onProgress?.(50 + (lineIndex / lines.length) * 50, `Parsed ${lineIndex} lines`)
      }

      if (this.isCommitHeader(line)) {
        currentDate = this.parseCommitDate(line)
        continue
      }

      this.processNumstatLine(line, currentDate, churnMap)
    }

    return churnMap
  }

  private isCommitHeader(line: string): boolean {
    return /^[a-f0-9]{40}\s/.test(line)
  }

  private parseCommitDate(line: string): Date {
    const dateStr = line.split(' ').slice(1).join(' ')
    return new Date(dateStr)
  }

  private processNumstatLine(line: string, currentDate: Date, churnMap: Map<string, ChurnMetrics>): void {
    const parts = line.split('\t')
    if (parts.length < 3) return

    const additions = parts[0] === '-' ? 0 : parseInt(parts[0], 10) || 0
    const deletions = parts[1] === '-' ? 0 : parseInt(parts[1], 10) || 0
    const filePath = parts[2]

    this.updateChurnMetrics(churnMap, filePath, additions, deletions, currentDate)
  }

  private updateChurnMetrics(
    churnMap: Map<string, ChurnMetrics>,
    filePath: string,
    additions: number,
    deletions: number,
    date: Date
  ): void {
    if (!churnMap.has(filePath)) {
      churnMap.set(filePath, {
        commits: 0,
        additions: 0,
        deletions: 0,
        lastModified: date,
      })
    }

    const metrics = churnMap.get(filePath)!
    metrics.commits += 1
    metrics.additions += additions
    metrics.deletions += deletions
    metrics.lastModified = new Date(Math.max(metrics.lastModified.getTime(), date.getTime()))
  }

  /**
   * Get churn score for a file (weighted by commits and changes)
   */
  getChurnScore(metrics: ChurnMetrics): number {
    return metrics.commits + (metrics.additions + metrics.deletions) / 100
  }

  /**
   * Classify files by churn level
   */
  getChurnLevel(metrics: ChurnMetrics): ChurnLevel {
    const { commits, additions, deletions } = metrics
    const totalChanges = additions + deletions

    if (commits > 50 || totalChanges > 1000) return 'critical'
    if (commits > 20 || totalChanges > 500) return 'high'
    if (commits > 10 || totalChanges > 200) return 'medium'
    if (commits > 0) return 'low'
    return 'none'
  }

  /**
   * Find hotspot files with highest churn
   */
  findHotspots(churnMap: Map<string, ChurnMetrics>, limit: number = 10): ChurnHotspot[] {
    return Array.from(churnMap.entries())
      .map(([file, metrics]) => ({
        file,
        metrics,
        score: this.getChurnScore(metrics),
      }))
      .sort((a, b) => b.score - a.score)
      .slice(0, limit)
  }

  /**
   * Get files modified in the last N days
   */
  getRecentFiles(churnMap: Map<string, ChurnMetrics>, daysBack: number = 30): Map<string, ChurnMetrics> {
    const cutoffDate = new Date()
    cutoffDate.setDate(cutoffDate.getDate() - daysBack)

    const recentFiles = new Map<string, ChurnMetrics>()
    for (const [file, metrics] of churnMap.entries()) {
      if (metrics.lastModified >= cutoffDate) {
        recentFiles.set(file, metrics)
      }
    }

    return recentFiles
  }

  /**
   * Get authors who have modified a file
   */
  async getFileAuthors(repoPath: string, filePath: string): Promise<string[]> {
    try {
      const output = await execGit(['log', '--pretty=format:%an', '--', filePath], { cwd: repoPath })
      return [...new Set(output.split('\n').filter((author) => author.trim()))]
    } catch {
      return []
    }
  }

  /**
   * Get commit count for a file
   */
  async getFileCommitCount(repoPath: string, filePath: string): Promise<number> {
    try {
      const output = await execGit(['log', '--oneline', '--', filePath], { cwd: repoPath })
      return output.split('\n').filter((line) => line.trim()).length
    } catch {
      return 0
    }
  }

  /**
   * Get uncommitted changes in the repository
   */
  async getUncommittedChanges(
    repoPath: string,
    onProgress?: ProgressCallback
  ): Promise<Map<string, 'modified' | 'added' | 'deleted' | 'renamed'>> {
    onProgress?.(0, 'Getting git status...')

    const output = await execGit(['status', '--porcelain'], {
      cwd: repoPath,
      onProgress,
    })

    const statusMap = this.parseStatusOutput(output, onProgress)

    onProgress?.(100, 'Git status complete')
    return statusMap
  }

  private parseStatusOutput(
    output: string,
    onProgress?: ProgressCallback
  ): Map<string, 'modified' | 'added' | 'deleted' | 'renamed'> {
    const statusMap = new Map<string, 'modified' | 'added' | 'deleted' | 'renamed'>()
    const lines = output.split('\n')

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i]
      if (!line.trim()) continue

      if (i % 10 === 0) {
        onProgress?.(50 + (i / lines.length) * 50, `Processed ${i} changes`)
      }

      const status = line.substring(0, 2).trim()
      const filePath = line.substring(3).trim()

      statusMap.set(filePath, this.mapStatusCode(status))
    }

    return statusMap
  }

  private mapStatusCode(status: string): 'modified' | 'added' | 'deleted' | 'renamed' {
    if (status.includes('A')) return 'added'
    if (status.includes('D')) return 'deleted'
    if (status.includes('R')) return 'renamed'
    return 'modified'
  }
}
