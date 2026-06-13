import { HealthGrade } from '@models/types'
import { CodeMap } from '@models/codeMap'
import { SecurityScanner } from './securityScanner'
import { PatternDetector } from './patternDetector'

export interface HealthMetrics {
  score: number
  grade: HealthGrade
  deadCodePercentage: number
  circularDependencies: number
  avgCoupling: number
  securityIssues: number
  criticalIssues: number
  complexFunctions: number
  godObjects: number
}

export class HealthScorer {
  private securityScanner: SecurityScanner
  private patternDetector: PatternDetector

  constructor() {
    this.securityScanner = new SecurityScanner()
    this.patternDetector = new PatternDetector()
  }

  calculateHealth(codeMap: CodeMap, code: string): HealthMetrics {
    const metrics = {
      score: 100,
      grade: 'A' as HealthGrade,
      deadCodePercentage: 0,
      circularDependencies: 0,
      avgCoupling: 0,
      securityIssues: 0,
      criticalIssues: 0,
      complexFunctions: 0,
      godObjects: 0,
    }

    // Calculate dead code percentage
    const { deadCodePercentage, deadCodeNodes } = this.calculateDeadCode(codeMap)
    metrics.deadCodePercentage = deadCodePercentage

    // Deduct points for dead code (up to 20 points)
    metrics.score -= Math.min(deadCodePercentage * 0.2, 20)

    // Calculate circular dependencies
    const cycles = codeMap.findCircularDependencies()
    metrics.circularDependencies = cycles.length

    // Deduct points for circular dependencies (2 points each, up to 15 points)
    metrics.score -= Math.min(cycles.length * 2, 15)

    // Calculate coupling
    const { avgCoupling } = this.calculateCoupling(codeMap)
    metrics.avgCoupling = avgCoupling

    // Deduct points for high coupling
    if (avgCoupling > 5) {
      metrics.score -= Math.min((avgCoupling - 5) * 2, 15)
    }

    // Scan for security issues
    const securityIssues = this.securityScanner.scan(code)
    metrics.securityIssues = securityIssues.length
    metrics.criticalIssues = securityIssues.filter(i => i.severity === 'critical').length

    // Deduct points for security issues
    const criticalDeduction = metrics.criticalIssues * 5
    const highDeduction = securityIssues.filter(i => i.severity === 'high').length * 2
    metrics.score -= Math.min(criticalDeduction + highDeduction, 30)

    // Detect anti-patterns
    const patterns = this.patternDetector.detectPatterns(codeMap)
    const antiPatterns = patterns.filter(p => p.type === 'anti_pattern')
    const godObjects = antiPatterns.filter(p => p.name === 'God Object').length
    const complexFunctions = antiPatterns.filter(p => p.name === 'Complex Function').length

    metrics.godObjects = godObjects
    metrics.complexFunctions = complexFunctions

    // Deduct points for anti-patterns
    metrics.score -= Math.min(godObjects * 3, 10)
    metrics.score -= Math.min(complexFunctions * 1, 10)

    // Ensure score doesn't go below 0
    metrics.score = Math.max(metrics.score, 0)

    // Convert score to grade
    metrics.grade = this.scoreToGrade(metrics.score)

    return metrics
  }

  private calculateDeadCode(codeMap: CodeMap): { deadCodePercentage: number; deadCodeNodes: string[] } {
    const callEdges = codeMap.getEdges().filter(e => e.data.type === 'calls')
    const calledNodes = new Set(callEdges.map(e => e.data.target))

    const allFunctions = codeMap.getNodes().filter(n => n.data.type === 'function' || n.data.type === 'method')
    const deadFunctions = allFunctions.filter(n => !n.data.isExported && !calledNodes.has(n.data.id))

    const deadCodePercentage = allFunctions.length > 0 ? (deadFunctions.length / allFunctions.length) * 100 : 0

    return {
      deadCodePercentage,
      deadCodeNodes: deadFunctions.map(n => n.data.id),
    }
  }

  private calculateCoupling(codeMap: CodeMap): { fanIn: number; fanOut: number; avgCoupling: number } {
    const coupling = codeMap.calculateCoupling()

    const fanInArray = Array.from(coupling.fanIn.values())
    const fanOutArray = Array.from(coupling.fanOut.values())

    const avgFanIn = fanInArray.length > 0 ? fanInArray.reduce((a, b) => a + b, 0) / fanInArray.length : 0
    const avgFanOut = fanOutArray.length > 0 ? fanOutArray.reduce((a, b) => a + b, 0) / fanOutArray.length : 0

    return {
      fanIn: Math.round(avgFanIn * 100) / 100,
      fanOut: Math.round(avgFanOut * 100) / 100,
      avgCoupling: Math.round((avgFanIn + avgFanOut) * 100) / 100,
    }
  }

  private scoreToGrade(score: number): HealthGrade {
    if (score >= 90) return 'A'
    if (score >= 80) return 'B'
    if (score >= 70) return 'C'
    if (score >= 60) return 'D'
    return 'F'
  }

  getHealthRecommendations(metrics: HealthMetrics): string[] {
    const recommendations: string[] = []

    if (metrics.deadCodePercentage > 10) {
      recommendations.push(`Remove ${Math.round(metrics.deadCodePercentage)}% of dead code`)
    }

    if (metrics.circularDependencies > 0) {
      recommendations.push(`Break ${metrics.circularDependencies} circular dependency cycles`)
    }

    if (metrics.avgCoupling > 5) {
      recommendations.push(`Reduce average coupling from ${metrics.avgCoupling.toFixed(2)} (target: < 3)`)
    }

    if (metrics.criticalIssues > 0) {
      recommendations.push(`Fix ${metrics.criticalIssues} critical security issues immediately`)
    }

    if (metrics.godObjects > 0) {
      recommendations.push(`Refactor ${metrics.godObjects} God Object(s) by splitting responsibilities`)
    }

    if (metrics.complexFunctions > 0) {
      recommendations.push(`Simplify ${metrics.complexFunctions} overly complex functions`)
    }

    return recommendations
  }
}
