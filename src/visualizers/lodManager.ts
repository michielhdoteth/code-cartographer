/**
 * Level-of-Detail (LOD) management for visualization
 * Shows different node counts based on zoom level (10x speedup through aggregation)
 */

export interface LODConfig {
  zoomThresholds?: {
    veryZoomedOut: number
    zoomedOut: number
    normal: number
    zoomedIn: number
  }
  aggregationThreshold?: number
}

export interface NodeAggregate {
  id: string
  type: 'aggregate'
  label: string
  childCount: number
  childNodeIds: string[]
  x: number
  y: number
  r?: number
}

export interface LODNode {
  id: string
  type: 'node' | 'aggregate'
  label: string
  visible: boolean
  aggregate?: NodeAggregate
}

type ZoomLevel = 'veryZoomedOut' | 'zoomedOut' | 'normal' | 'zoomedIn'

const DEFAULT_CONFIG: Required<LODConfig> = {
  zoomThresholds: {
    veryZoomedOut: 0.1,
    zoomedOut: 0.3,
    normal: 1.0,
    zoomedIn: 2.0,
  },
  aggregationThreshold: 300,
}

export class LODManager {
  private config: Required<LODConfig>
  private currentZoom = 1.0
  private nodeMap: Map<string, LODNode> = new Map()
  private aggregates: Map<string, NodeAggregate> = new Map()

  constructor(config: LODConfig = {}) {
    this.config = { ...DEFAULT_CONFIG, ...config }
  }

  setZoom(zoom: number): {
    levelChanged: boolean
    previousLevel: string
    currentLevel: string
    nodesToShow: string[]
    nodesToHide: string[]
  } {
    const previousLevel = this.getZoomLevel(this.currentZoom)
    const currentLevel = this.getZoomLevel(zoom)
    this.currentZoom = zoom

    if (previousLevel === currentLevel) {
      return { levelChanged: false, previousLevel, currentLevel, nodesToShow: [], nodesToHide: [] }
    }

    const nodesToShow: string[] = []
    const nodesToHide: string[] = []

    for (const [id, node] of this.nodeMap) {
      const shouldBeVisible = this.shouldBeVisible(id, currentLevel)
      if (shouldBeVisible && !node.visible) {
        nodesToShow.push(id)
        node.visible = true
      } else if (!shouldBeVisible && node.visible) {
        nodesToHide.push(id)
        node.visible = false
      }
    }

    return { levelChanged: true, previousLevel, currentLevel, nodesToShow, nodesToHide }
  }

  addNode(id: string, label: string, type: 'node' | 'aggregate' = 'node'): void {
    this.nodeMap.set(id, { id, type, label, visible: true })
  }

  createAggregate(
    aggregateId: string,
    label: string,
    childNodeIds: string[],
    centerX: number,
    centerY: number,
    radius?: number
  ): void {
    const aggregate: NodeAggregate = {
      id: aggregateId,
      type: 'aggregate',
      label,
      childCount: childNodeIds.length,
      childNodeIds,
      x: centerX,
      y: centerY,
      r: radius,
    }

    this.aggregates.set(aggregateId, aggregate)
    this.addNode(aggregateId, label, 'aggregate')
  }

  getVisibleNodes(): LODNode[] {
    const level = this.getZoomLevel(this.currentZoom)
    return [...this.nodeMap.values()].filter((node) => this.shouldBeVisible(node.id, level))
  }

  getVisibleNodeCount(): number {
    return this.getVisibleNodes().length
  }

  getStats(): {
    zoom: number
    level: string
    totalNodes: number
    visibleNodes: number
    aggregates: number
    estimatedReduction: number
  } {
    const visible = this.getVisibleNodeCount()
    const total = this.nodeMap.size

    return {
      zoom: this.currentZoom,
      level: this.getZoomLevel(this.currentZoom),
      totalNodes: total,
      visibleNodes: visible,
      aggregates: this.aggregates.size,
      estimatedReduction: total > 0 ? ((total - visible) / total) * 100 : 0,
    }
  }

  getNodeSize(id: string): number {
    const node = this.nodeMap.get(id)
    if (!node) return 5

    if (node.type === 'aggregate') {
      const aggregate = this.aggregates.get(id)
      if (aggregate) return Math.sqrt(aggregate.childCount) * 2
    }

    return 5
  }

  shouldShowLabel(id: string): boolean {
    const level = this.getZoomLevel(this.currentZoom)
    if (level === 'zoomedIn') return true
    if (level === 'veryZoomedOut') return false

    const node = this.nodeMap.get(id)
    return node?.type === 'aggregate'
  }

  clear(): void {
    this.nodeMap.clear()
    this.aggregates.clear()
  }

  private getZoomLevel(zoom: number): ZoomLevel {
    const { zoomThresholds } = this.config

    if (zoom < zoomThresholds.veryZoomedOut) return 'veryZoomedOut'
    if (zoom < zoomThresholds.zoomedOut) return 'zoomedOut'
    if (zoom < zoomThresholds.normal) return 'normal'
    return 'zoomedIn'
  }

  private shouldBeVisible(nodeId: string, level: ZoomLevel): boolean {
    const node = this.nodeMap.get(nodeId)
    if (!node) return false

    const isAggregate = node.type === 'aggregate'
    const showAggregatesOnly = level === 'veryZoomedOut' || level === 'zoomedOut'

    if (isAggregate) return showAggregatesOnly
    return !showAggregatesOnly
  }
}
