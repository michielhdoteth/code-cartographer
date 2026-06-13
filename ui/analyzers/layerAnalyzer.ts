import { ArchitectureLayer } from '@models/types'

export class LayerAnalyzer {
  private layerOrder: Record<ArchitectureLayer, number> = {
    ui: 0,
    components: 1,
    services: 2,
    data: 3,
    utils: 4,
    config: 4,
    other: 5,
  }

  private layerPatterns: Array<[ArchitectureLayer, RegExp]> = [
    ['ui', /\b(ui|views|pages|screens|layouts?)\b/i],
    ['components', /\b(components?|widgets?)\b/i],
    ['services', /\b(services?|api|controllers?|handlers?|endpoints?)\b/i],
    ['data', /\b(data|models?|stores?|db|database|repositories?)\b/i],
    ['utils', /\b(utils?|helpers?|libs?|constants?|types?)\b/i],
    ['config', /\b(config|settings|env|constants?)\b/i],
  ]

  /**
   * Detect architectural layer from file path
   * Based on codeflow implementation
   */
  detectLayer(filePath: string): ArchitectureLayer {
    const pathLower = filePath.toLowerCase()

    // Check each pattern in order
    for (const [layer, pattern] of this.layerPatterns) {
      if (pattern.test(pathLower)) {
        return layer
      }
    }

    return 'other'
  }

  /**
   * Get all layer patterns for a given layer
   */
  getLayerPatterns(layer: ArchitectureLayer): RegExp[] {
    return this.layerPatterns.filter(([l]) => l === layer).map(([, pattern]) => pattern)
  }

  /**
   * Check if an import violates layer boundaries
   * Lower layers (higher order number) should not import from higher layers (lower order number)
   */
  isLayerViolation(sourcePath: string, targetPath: string): boolean {
    const sourceLayer = this.detectLayer(sourcePath)
    const targetLayer = this.detectLayer(targetPath)

    const sourceOrder = this.layerOrder[sourceLayer]
    const targetOrder = this.layerOrder[targetLayer]

    // Violation: lower layer (higher number) importing from higher layer (lower number)
    return sourceOrder > targetOrder
  }

  /**
   * Describe layer violation
   */
  describeViolation(sourcePath: string, targetPath: string): string {
    const sourceLayer = this.detectLayer(sourcePath)
    const targetLayer = this.detectLayer(targetPath)

    return `Layer violation: ${sourceLayer} (${sourcePath}) importing from ${targetLayer} (${targetPath})`
  }

  /**
   * Get layer display name with color
   */
  getLayerInfo(layer: ArchitectureLayer): { name: string; color: string; order: number } {
    const colors: Record<ArchitectureLayer, string> = {
      ui: '#4d9fff',
      components: '#22d3ee',
      services: '#a78bfa',
      data: '#ff9f43',
      utils: '#00ff9d',
      config: '#ec4899',
      other: '#64748b',
    }

    return {
      name: layer,
      color: colors[layer],
      order: this.layerOrder[layer],
    }
  }

  /**
   * Classify multiple files by layer
   */
  classifyFiles(filePaths: string[]): Map<ArchitectureLayer, string[]> {
    const layerMap = new Map<ArchitectureLayer, string[]>()

    for (const layer of Object.keys(this.layerOrder) as ArchitectureLayer[]) {
      layerMap.set(layer, [])
    }

    for (const filePath of filePaths) {
      const layer = this.detectLayer(filePath)
      const files = layerMap.get(layer) || []
      files.push(filePath)
      layerMap.set(layer, files)
    }

    return layerMap
  }

  /**
   * Detect all layer violations in a set of imports
   */
  findLayerViolations(imports: Array<{ from: string; to: string }>): Array<{ from: string; to: string; violation: string }> {
    return imports
      .filter(({ from, to }) => this.isLayerViolation(from, to))
      .map(({ from, to }) => ({
        from,
        to,
        violation: this.describeViolation(from, to),
      }))
  }
}
