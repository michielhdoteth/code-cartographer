import { CodeMap, CodeNode, CodeFile } from '@models/codeMap'
import { AnalysisResult, BlastRadiusLevel } from '@models/types'
import { LayerAnalyzer } from '@analyzers/layerAnalyzer'

export type OutputFormat = 'tree' | 'deps' | 'health' | 'coverage'

export class AsciiFormatter {
  private layerAnalyzer = new LayerAnalyzer()

  /**
   * Format codebase as tree view with metrics
   * Example output:
   * src/
   *   api.ts (45 LOC, 8 functions) [HIGH BLAST RADIUS]
   *   components/
   *     Header.tsx (120 LOC, 3 components)
   */
  formatAsTree(codeMap: CodeMap, churnMap?: Map<string, number>, blastMap?: Map<string, number>): string {
    const lines: string[] = ['# Code Structure\n']
    const files = codeMap.getFiles()

    // Group files by directory
    const filesByDir = new Map<string, CodeFile[]>()
    for (const file of files) {
      const dir = this.getDirFromPath(file.data.path)
      if (!filesByDir.has(dir)) {
        filesByDir.set(dir, [])
      }
      filesByDir.get(dir)!.push(file)
    }

    // Sort and format
    const dirs = Array.from(filesByDir.keys()).sort()
    for (const dir of dirs) {
      const depth = dir.split('/').filter((p) => p).length
      const indent = '  '.repeat(depth)
      lines.push(`${indent}${dir || 'src/'}/`)

      const filesInDir = filesByDir.get(dir) || []
      for (const file of filesInDir.sort((a, b) => a.data.path.localeCompare(b.data.path))) {
        const fileName = file.data.path.split('/').pop() || file.data.path
        const nodes = codeMap.getNodesByFile(file.data.path)
        const nodeCount = nodes.length
        const churn = churnMap?.get(file.data.path) || 0
        const blastScore = blastMap?.get(file.data.id) || 0

        let metaStr = `(${file.data.lines} LOC, ${nodeCount} items)`
        if (churn > 0) {
          metaStr += ` [churn: ${Math.round(churn)}]`
        }
        if (blastScore > 5) {
          metaStr += ` [HIGH IMPACT]`
        }

        lines.push(`${indent}  ${fileName} ${metaStr}`)
      }
    }

    return lines.join('\n')
  }

  /**
   * Format as dependency graph
   * Example output:
   * api.ts -> auth.ts (imports)
   * Header.tsx -> api.ts (calls getUsers)
   * components/ -> services/api.ts (3 imports)
   */
  formatAsDependencies(codeMap: CodeMap): string {
    const lines: string[] = ['# Dependency Graph\n']
    const edges = codeMap.getEdges()
    const nodes = new Map(codeMap.getNodes().map((n) => [n.data.id, n]))

    // Group by edge type
    const edgesByType = new Map<string, string[]>()
    for (const edge of edges) {
      const source = nodes.get(edge.data.source)
      const target = nodes.get(edge.data.target)

      if (!source || !target) continue

      const sourceName = source.data.name
      const targetName = target.data.name
      const edgeType = edge.data.type
      const metadata = edge.data.metadata?.fn ? ` (${edge.data.metadata.fn})` : ''

      const edgeStr = `${sourceName} -> ${targetName} (${edgeType})${metadata}`

      if (!edgesByType.has(edgeType)) {
        edgesByType.set(edgeType, [])
      }
      edgesByType.get(edgeType)!.push(edgeStr)
    }

    // Format by type
    for (const [type, edges_] of edgesByType) {
      lines.push(`\n## ${type.toUpperCase()}`)
      for (const edge of edges_.sort()) {
        lines.push(`  ${edge}`)
      }
    }

    return lines.join('\n')
  }

  /**
   * Format health report with issues and recommendations
   */
  formatAsHealth(codeMap: CodeMap, analysis?: AnalysisResult): string {
    const lines: string[] = ['# Code Health Report\n']
    const stats = codeMap.getStatistics()

    lines.push('## Statistics')
    lines.push(`  Total Nodes: ${stats.totalNodes}`)
    lines.push(`  Total Edges: ${stats.totalEdges}`)
    lines.push(`  Total Files: ${stats.totalFiles}`)
    lines.push(`  Total Lines: ${stats.totalLines}`)
    lines.push(`  Avg Fan-In: ${stats.avgFanIn}`)
    lines.push(`  Avg Fan-Out: ${stats.avgFanOut}`)
    lines.push(`  Avg Node Size: ${stats.avgNodeSize} LOC`)

    // Circular dependencies
    const cycles = codeMap.findCircularDependencies()
    lines.push(`\n## Circular Dependencies: ${cycles.length}`)
    if (cycles.length > 0) {
      for (const cycle of cycles.slice(0, 10)) {
        lines.push(`  ${cycle.join(' -> ')}`)
      }
      if (cycles.length > 10) {
        lines.push(`  ... and ${cycles.length - 10} more`)
      }
    } else {
      lines.push('  None detected')
    }

    // Layer violations
    if (analysis?.layerViolations) {
      lines.push(`\n## Architecture Violations: ${analysis.layerViolations.length}`)
      for (const violation of analysis.layerViolations.slice(0, 10)) {
        lines.push(`  ${violation}`)
      }
      if (analysis.layerViolations.length > 10) {
        lines.push(`  ... and ${analysis.layerViolations.length - 10} more`)
      }
    }

    // Security issues
    if (analysis?.securityIssues) {
      lines.push(`\n## Security Issues: ${analysis.securityIssues.length}`)
      for (const issue of analysis.securityIssues.slice(0, 10)) {
        lines.push(`  [${issue.severity}] ${issue.type}: ${issue.message}`)
      }
      if (analysis.securityIssues.length > 10) {
        lines.push(`  ... and ${analysis.securityIssues.length - 10} more`)
      }
    }

    // Dead code
    if (analysis?.deadCode) {
      lines.push(`\n## Dead Code: ${analysis.deadCode.length} items`)
      const deadCodeNames = analysis.deadCode
        .slice(0, 10)
        .map((id) => codeMap.getNode(id)?.data.name || id)
      for (const name of deadCodeNames) {
        lines.push(`  ${name}`)
      }
      if (analysis.deadCode.length > 10) {
        lines.push(`  ... and ${analysis.deadCode.length - 10} more`)
      }
    }

    // Recommendations
    lines.push('\n## Recommendations')
    if (cycles.length > 0) {
      lines.push('  - Refactor circular dependencies')
    }
    if (stats.avgFanIn > 10) {
      lines.push('  - High fan-in indicates tightly coupled modules')
    }
    if (stats.avgFanOut > 10) {
      lines.push('  - High fan-out indicates modules with many dependencies')
    }
    if (analysis?.deadCode && analysis.deadCode.length > stats.totalNodes * 0.1) {
      lines.push('  - Consider removing or refactoring dead code')
    }

    return lines.join('\n')
  }

  /**
   * Format as coverage/hotspot view with ANSI colors for terminal
   * Uses ANSI color codes for terminal output
   */
  formatAsCoverage(codeMap: CodeMap, churnMap?: Map<string, number>, blastMap?: Map<string, number>): string {
    const lines: string[] = ['# Coverage & Hotspots\n']
    const files = codeMap.getFiles()

    // Sort by combined score (churn + blast radius)
    const fileScores = files
      .map((f) => {
        const churn = churnMap?.get(f.data.path) || 0
        const blast = blastMap?.get(f.data.id) || 0
        const score = churn + blast * 0.5 // Weight blast less than churn

        return { file: f, churn, blast, score }
      })
      .sort((a, b) => b.score - a.score)

    lines.push('## Hotspot Files (sorted by churn + impact)')
    lines.push('File | LOC | Churn | Impact | Risk Level')
    lines.push('-'.repeat(60))

    for (const { file, churn, blast, score } of fileScores.slice(0, 20)) {
      const fileName = file.data.path.padEnd(30)
      const loc = file.data.lines.toString().padStart(6)
      const churnStr = Math.round(churn).toString().padStart(6)
      const blastStr = Math.round(blast).toString().padStart(6)

      let risk = 'LOW'
      if (score > 30) risk = 'CRITICAL'
      else if (score > 15) risk = 'HIGH'
      else if (score > 5) risk = 'MEDIUM'

      lines.push(`${fileName} | ${loc} | ${churnStr} | ${blastStr} | ${risk}`)
    }

    if (fileScores.length > 20) {
      lines.push(`... and ${fileScores.length - 20} more files`)
    }

    // Summary stats
    lines.push('\n## Summary')
    const totalChurn = Array.from(churnMap?.values() || []).reduce((a, b) => a + b, 0)
    const totalBlast = Array.from(blastMap?.values() || []).reduce((a, b) => a + b, 0)

    lines.push(`  Average Churn: ${(totalChurn / files.length).toFixed(2)}`)
    lines.push(`  Total Impact Score: ${totalBlast.toFixed(2)}`)
    lines.push(`  Critical Files: ${fileScores.filter((f) => f.score > 30).length}`)
    lines.push(`  High Priority Files: ${fileScores.filter((f) => f.score > 15).length}`)

    return lines.join('\n')
  }

  /**
   * Main format method - dispatches to appropriate formatter
   */
  format(codeMap: CodeMap, outputFormat: OutputFormat, options?: { analysis?: AnalysisResult; churnMap?: Map<string, number>; blastMap?: Map<string, number> }): string {
    switch (outputFormat) {
      case 'tree':
        return this.formatAsTree(codeMap, options?.churnMap, options?.blastMap)
      case 'deps':
        return this.formatAsDependencies(codeMap)
      case 'health':
        return this.formatAsHealth(codeMap, options?.analysis)
      case 'coverage':
        return this.formatAsCoverage(codeMap, options?.churnMap, options?.blastMap)
      default:
        return ''
    }
  }

  /**
   * Extract directory path from file path
   */
  private getDirFromPath(path: string): string {
    const parts = path.split('/')
    return parts.slice(0, -1).join('/')
  }

  /**
   * Format a blast radius result as human-readable text
   */
  formatBlastRadius(fileId: string, fileName: string, blastResult: any): string {
    const lines: string[] = []

    lines.push(`# Blast Radius Analysis: ${fileName}\n`)
    lines.push(`## Risk Level: ${blastResult.level.toUpperCase()}`)
    lines.push(`## Direct Dependents: ${blastResult.count}`)
    lines.push(`## Transitive Dependents: ${blastResult.transitiveCount}`)
    lines.push(`## Impact Score: ${blastResult.impactScore}`)
    lines.push(`## Centrality Score: ${blastResult.centrality}`)
    lines.push(`## Functions Used Externally: ${blastResult.fnsUsed}`)
    lines.push(`## Total External Calls: ${blastResult.totalCalls}`)

    if (blastResult.affected.length > 0) {
      lines.push(`\n## Directly Affected Files (${blastResult.affected.length})`)
      for (const file of blastResult.affected.slice(0, 20)) {
        lines.push(`  - ${file}`)
      }
      if (blastResult.affected.length > 20) {
        lines.push(`  ... and ${blastResult.affected.length - 20} more`)
      }
    }

    if (blastResult.transitive.length > 0) {
      lines.push(`\n## Transitively Affected Files (${blastResult.transitive.length})`)
      for (const file of blastResult.transitive.slice(0, 20)) {
        lines.push(`  - ${file}`)
      }
      if (blastResult.transitive.length > 20) {
        lines.push(`  ... and ${blastResult.transitive.length - 20} more`)
      }
    }

    if (blastResult.dependencies.length > 0) {
      lines.push(`\n## Dependencies (Files This Imports From)`)
      for (const dep of blastResult.dependencies.slice(0, 20)) {
        lines.push(`  - ${dep}`)
      }
      if (blastResult.dependencies.length > 20) {
        lines.push(`  ... and ${blastResult.dependencies.length - 20} more`)
      }
    }

    return lines.join('\n')
  }
}
