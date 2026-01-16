import { ChurnMetrics } from '@models/types'
import { execSync } from 'child_process'

export class ChurnAnalyzer {
  /**
   * Parse git log to calculate churn metrics per file
   * Returns a map of file paths to their churn metrics
   */
  analyzeChurn(repoPath: string, authorFilter?: string): Map<string, ChurnMetrics> {
    const churnMap = new Map<string, ChurnMetrics>()

    try {
      // Get git log with numstat format
      let command = 'git -C "{0}" log --numstat --follow --pretty=format:"%H %aI"'
        .replace('{0}', repoPath)

      if (authorFilter) {
        command += ` --author="${authorFilter}"`
      }

      const output = execSync(command, { encoding: 'utf-8', maxBuffer: 10 * 1024 * 1024 })

      // Parse git log output
      const lines = output.split('\n')
      let currentCommit = ''
      let currentDate = new Date()

      for (const line of lines) {
        if (!line.trim()) continue

        // Check if this is a commit header line (hash and date)
        if (line.match(/^[a-f0-9]{40}\s/)) {
          const parts = line.split(' ')
          currentCommit = parts[0]
          currentDate = new Date(parts.slice(1).join(' '))
          continue
        }

        // Parse numstat line (additions deletions filename)
        const parts = line.split('\t')
        if (parts.length >= 3) {
          const additions = parts[0] === '-' ? 0 : parseInt(parts[0], 10) || 0
          const deletions = parts[1] === '-' ? 0 : parseInt(parts[1], 10) || 0
          const filePath = parts[2]

          if (!churnMap.has(filePath)) {
            churnMap.set(filePath, {
              commits: 0,
              additions: 0,
              deletions: 0,
              lastModified: currentDate,
            })
          }

          const metrics = churnMap.get(filePath)!
          metrics.commits += 1
          metrics.additions += additions
          metrics.deletions += deletions
          metrics.lastModified = new Date(Math.max(metrics.lastModified.getTime(), currentDate.getTime()))
        }
      }
    } catch (error) {
      console.error('Error analyzing git churn:', error)
    }

    return churnMap
  }

  /**
   * Get churn score for a file (weighted by commits and changes)
   * Higher score = more active file
   */
  getChurnScore(metrics: ChurnMetrics): number {
    return metrics.commits + (metrics.additions + metrics.deletions) / 100
  }

  /**
   * Classify files by churn level
   */
  getChurnLevel(metrics: ChurnMetrics): 'critical' | 'high' | 'medium' | 'low' | 'none' {
    const commits = metrics.commits
    const changes = metrics.additions + metrics.deletions

    if (commits > 50 || changes > 1000) return 'critical'
    if (commits > 20 || changes > 500) return 'high'
    if (commits > 10 || changes > 200) return 'medium'
    if (commits > 0) return 'low'
    return 'none'
  }

  /**
   * Find hotspot files - files with high churn
   */
  findHotspots(churnMap: Map<string, ChurnMetrics>, limit: number = 10): Array<{ file: string; metrics: ChurnMetrics; score: number }> {
    const hotspots = Array.from(churnMap.entries())
      .map(([file, metrics]) => ({
        file,
        metrics,
        score: this.getChurnScore(metrics),
      }))
      .sort((a, b) => b.score - a.score)
      .slice(0, limit)

    return hotspots
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
  getFileAuthors(repoPath: string, filePath: string): string[] {
    try {
      const command = `git -C "${repoPath}" log --pretty=format:"%an" -- "${filePath}"`
      const output = execSync(command, { encoding: 'utf-8' })
      const authors = [...new Set(output.split('\n').filter((a) => a.trim()))]
      return authors
    } catch (error) {
      return []
    }
  }

  /**
   * Get commit count for a file
   */
  getFileCommitCount(repoPath: string, filePath: string): number {
    try {
      const command = `git -C "${repoPath}" log --oneline -- "${filePath}" | wc -l`
      const output = execSync(command, { encoding: 'utf-8', shell: '/bin/bash' })
      return parseInt(output.trim(), 10) || 0
    } catch (error) {
      return 0
    }
  }

  /**
   * Get current git status to identify uncommitted changes
   */
  getUncommittedChanges(repoPath: string): Map<string, 'modified' | 'added' | 'deleted' | 'renamed'> {
    const statusMap = new Map<string, 'modified' | 'added' | 'deleted' | 'renamed'>()

    try {
      const command = `git -C "${repoPath}" status --porcelain`
      const output = execSync(command, { encoding: 'utf-8' })

      for (const line of output.split('\n')) {
        if (!line.trim()) continue

        const status = line.substring(0, 2).trim()
        const filePath = line.substring(3).trim()

        let statusType: 'modified' | 'added' | 'deleted' | 'renamed' = 'modified'
        if (status.includes('A')) statusType = 'added'
        else if (status.includes('D')) statusType = 'deleted'
        else if (status.includes('R')) statusType = 'renamed'

        statusMap.set(filePath, statusType)
      }
    } catch (error) {
      console.error('Error getting git status:', error)
    }

    return statusMap
  }
}
