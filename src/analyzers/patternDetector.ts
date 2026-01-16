import { PatternMatch } from '@models/types'
import { CodeMap, CodeNode } from '@models/codeMap'

export class PatternDetector {
  detectPatterns(codeMap: CodeMap): PatternMatch[] {
    const patterns: PatternMatch[] = []

    patterns.push(...this.detectDesignPatterns(codeMap))
    patterns.push(...this.detectAntiPatterns(codeMap))
    patterns.push(...this.detectLayerViolations(codeMap))

    return patterns
  }

  private detectDesignPatterns(codeMap: CodeMap): PatternMatch[] {
    const patterns: PatternMatch[] = []

    // Singleton pattern: class with private constructor and static instance
    for (const node of codeMap.getNodes()) {
      if (node.data.type !== 'class') continue

      const methods = codeMap.getNodes().filter(n => n.data.parent === node.data.id && n.data.type === 'method')
      const hasPrivateConstructor = methods.some(m => m.data.name === 'constructor' || m.data.name === '__init__')
      const hasStaticInstance = methods.some(m => m.data.modifiers?.includes('static'))

      if (hasPrivateConstructor && hasStaticInstance) {
        patterns.push({
          type: 'design_pattern',
          name: 'Singleton',
          description: 'Class appears to implement the Singleton pattern',
          severity: 'info',
          nodes: [node.data.id],
          message: `${node.data.name} likely implements the Singleton pattern`,
        })
      }
    }

    // Factory pattern: class with static factory methods
    for (const node of codeMap.getNodes()) {
      if (node.data.type !== 'class') continue

      const methods = codeMap.getNodes().filter(n => n.data.parent === node.data.id && n.data.type === 'method')
      const factoryMethods = methods.filter(m => m.data.name?.includes('create') && m.data.modifiers?.includes('static'))

      if (factoryMethods.length > 0) {
        patterns.push({
          type: 'design_pattern',
          name: 'Factory',
          description: 'Class appears to implement a Factory pattern',
          severity: 'info',
          nodes: [node.data.id],
          message: `${node.data.name} has factory methods: ${factoryMethods.map(m => m.data.name).join(', ')}`,
        })
      }
    }

    // Observer/Event pattern: classes with subscribe/on methods
    for (const node of codeMap.getNodes()) {
      if (node.data.type !== 'class') continue

      const methods = codeMap.getNodes().filter(n => n.data.parent === node.data.id && n.data.type === 'method')
      const observerMethods = methods.filter(m => {
        const name = m.data.name?.toLowerCase() || ''
        return name.includes('subscribe') || name.includes('observe') || name.includes('on') || name.includes('listen')
      })

      if (observerMethods.length > 0) {
        patterns.push({
          type: 'design_pattern',
          name: 'Observer/Event Emitter',
          description: 'Class appears to implement an Observer or Event pattern',
          severity: 'info',
          nodes: [node.data.id],
          message: `${node.data.name} has observer methods: ${observerMethods.map(m => m.data.name).join(', ')}`,
        })
      }
    }

    return patterns
  }

  private detectAntiPatterns(codeMap: CodeMap): PatternMatch[] {
    const patterns: PatternMatch[] = []

    // God Object: class with too many methods/properties
    for (const node of codeMap.getNodes()) {
      if (node.data.type !== 'class') continue

      const children = codeMap.getNodes().filter(n => n.data.parent === node.data.id)
      if (children.length > 15) {
        patterns.push({
          type: 'anti_pattern',
          name: 'God Object',
          description: 'Class with too many methods/properties (Low cohesion)',
          severity: 'high',
          nodes: [node.data.id],
          message: `${node.data.name} has ${children.length} methods/properties - consider splitting into smaller classes`,
        })
      }
    }

    // Long files: files with too many lines
    for (const file of codeMap.getFiles()) {
      if (file.data.lines > 500) {
        patterns.push({
          type: 'anti_pattern',
          name: 'Long File',
          description: 'File with excessive lines of code',
          severity: 'medium',
          nodes: [file.data.id],
          message: `${file.data.path} has ${file.data.lines} lines - consider splitting into smaller modules`,
        })
      }
    }

    // Complex functions: functions with high cyclomatic complexity
    for (const node of codeMap.getNodes()) {
      if ((node.data.type === 'function' || node.data.type === 'method') && node.data.metrics.cyclomatic > 10) {
        patterns.push({
          type: 'anti_pattern',
          name: 'Complex Function',
          description: 'Function with high cyclomatic complexity',
          severity: 'high',
          nodes: [node.data.id],
          message: `${node.data.name} has cyclomatic complexity of ${node.data.metrics.cyclomatic} - consider refactoring`,
        })
      }
    }

    // Dead code: functions/methods that are never called
    const callEdges = codeMap.getEdges().filter(e => e.data.type === 'calls')
    const calledNodes = new Set(callEdges.map(e => e.data.target))

    for (const node of codeMap.getNodes()) {
      if ((node.data.type === 'function' || node.data.type === 'method') && !node.data.isExported && !calledNodes.has(node.data.id)) {
        patterns.push({
          type: 'anti_pattern',
          name: 'Dead Code',
          description: 'Function/method that appears to be unused',
          severity: 'medium',
          nodes: [node.data.id],
          message: `${node.data.name} appears to be unused - consider removing it`,
        })
      }
    }

    return patterns
  }

  private detectLayerViolations(codeMap: CodeMap): PatternMatch[] {
    const patterns: PatternMatch[] = []

    // Detect layer violations: lower layer depending on higher layer
    // Layers: Utils < Services < Controllers
    const layers: Record<string, number> = {
      util: 1,
      helper: 1,
      service: 2,
      controller: 3,
      view: 4,
      model: 2,
    }

    const getLayer = (nodeName: string): number => {
      const name = nodeName.toLowerCase()
      for (const [keyword, layer] of Object.entries(layers)) {
        if (name.includes(keyword)) return layer
      }
      return 0
    }

    for (const edge of codeMap.getEdges()) {
      if (edge.data.type === 'depends_on' || edge.data.type === 'imports') {
        const sourceLayer = getLayer(edge.data.source)
        const targetLayer = getLayer(edge.data.target)

        if (sourceLayer > 0 && targetLayer > 0 && sourceLayer < targetLayer) {
          patterns.push({
            type: 'architecture_violation',
            name: 'Layer Violation',
            description: 'Lower layer depends on higher layer',
            severity: 'high',
            nodes: [edge.data.source, edge.data.target],
            message: `Potential layer violation: Layer ${sourceLayer} depending on Layer ${targetLayer}`,
          })
        }
      }
    }

    return patterns
  }
}
