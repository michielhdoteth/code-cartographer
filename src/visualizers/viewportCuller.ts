/**
 * Viewport culling for D3 force simulation
 * Only renders nodes visible in current viewport (5-10x speedup)
 */

import { SpatialIndex, Bounds } from '@visualizers/spatialIndex'

export interface ViewportCullerOptions {
  margin?: number
  enableCulling?: boolean
  cullThreshold?: number
}

export interface CullResult {
  visibleNodeIds: Set<string>
  culledNodeIds: Set<string>
  visibleCount: number
  culledCount: number
  culledPercentage: number
}

const DEFAULT_OPTIONS: Required<ViewportCullerOptions> = {
  margin: 100,
  enableCulling: true,
  cullThreshold: 200,
}

export class ViewportCuller {
  private spatialIndex: SpatialIndex
  private options: Required<ViewportCullerOptions>
  private lastCullResult: CullResult | null = null

  constructor(graphBounds: Bounds, options: ViewportCullerOptions = {}) {
    this.options = { ...DEFAULT_OPTIONS, ...options }
    this.spatialIndex = new SpatialIndex(graphBounds)
  }

  updatePositions(nodes: Array<{ id: string; x: number; y: number }>): void {
    this.spatialIndex.clear()
    nodes.forEach((node) => this.spatialIndex.add({ id: node.id, x: node.x, y: node.y }))
  }

  cull(viewport: Bounds): CullResult {
    const allNodes = this.spatialIndex.getAll()
    const allNodeIds = new Set(allNodes.map((p) => p.id))

    if (!this.options.enableCulling || allNodes.length < this.options.cullThreshold) {
      return this.createResult(allNodeIds, new Set())
    }

    const expandedViewport: Bounds = {
      x: viewport.x - this.options.margin,
      y: viewport.y - this.options.margin,
      width: viewport.width + this.options.margin * 2,
      height: viewport.height + this.options.margin * 2,
    }

    const visibleNodeIds = new Set(this.spatialIndex.query(expandedViewport).map((p) => p.id))
    const culledNodeIds = new Set([...allNodeIds].filter((id) => !visibleNodeIds.has(id)))

    this.lastCullResult = this.createResult(visibleNodeIds, culledNodeIds)
    return this.lastCullResult
  }

  private createResult(visible: Set<string>, culled: Set<string>): CullResult {
    const total = visible.size + culled.size
    return {
      visibleNodeIds: visible,
      culledNodeIds: culled,
      visibleCount: visible.size,
      culledCount: culled.size,
      culledPercentage: total > 0 ? (culled.size / total) * 100 : 0,
    }
  }

  getNodesNear(x: number, y: number, distance: number): string[] {
    return this.spatialIndex.findWithinDistance(x, y, distance).map((p) => p.id)
  }

  findNearest(x: number, y: number): string | null {
    return this.spatialIndex.findNearest(x, y, 500)?.id ?? null
  }

  getStats(): {
    lastCullPercentage: number
    lastVisibleCount: number
    lastCulledCount: number
    spatialIndexStats: { pointCount: number; estimatedDepth: number }
  } {
    return {
      lastCullPercentage: this.lastCullResult?.culledPercentage ?? 0,
      lastVisibleCount: this.lastCullResult?.visibleCount ?? 0,
      lastCulledCount: this.lastCullResult?.culledCount ?? 0,
      spatialIndexStats: this.spatialIndex.getStats(),
    }
  }

  shouldUpdateNode(nodeId: string, cullResult: CullResult): boolean {
    return cullResult.visibleNodeIds.has(nodeId)
  }

  clear(): void {
    this.spatialIndex.clear()
    this.lastCullResult = null
  }
}
