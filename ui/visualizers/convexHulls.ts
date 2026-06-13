import * as d3 from 'd3'
import { stringHash, FOLDER_COLORS, calculateCentroid as calcCentroid } from '../utils/visualizationUtils.js'

export interface NodePosition {
  id: string
  x: number
  y: number
  folder: string
}

export class ConvexHullVisualizer {
  /**
   * Calculate convex hulls for each folder
   * Returns hull polygons grouped by folder
   */
  calculateHulls(nodes: NodePosition[], padding: number = 30): Map<string, [number, number][]> {
    // Group nodes by folder
    const nodesByFolder = new Map<string, NodePosition[]>()
    for (const node of nodes) {
      if (!nodesByFolder.has(node.folder)) {
        nodesByFolder.set(node.folder, [])
      }
      nodesByFolder.get(node.folder)!.push(node)
    }

    // Calculate convex hull for each folder
    const hulls = new Map<string, [number, number][]>()
    for (const [folder, folderNodes] of nodesByFolder.entries()) {
      if (folderNodes.length < 3) {
        // For small groups, use bounding box instead
        const hull = this.calculateBoundingBox(folderNodes, padding)
        hulls.set(folder, hull)
      } else {
        // Use d3 convex hull
        const points = folderNodes.map((n) => [n.x, n.y] as [number, number])
        const hull = d3.polygonHull(points)
        if (hull && hull.length > 0) {
          // Expand hull by padding
          const expandedHull = this.expandHull(hull, padding)
          hulls.set(folder, expandedHull)
        }
      }
    }

    return hulls
  }

  /**
   * Render convex hulls to SVG
   */
  renderHulls(svg: any, hulls: Map<string, [number, number][]>, colors: Map<string, string> = new Map()): void {
    const g = svg.append('g').attr('class', 'convex-hulls').attr('fill-opacity', 0.08)

    for (const [folder, hull] of hulls.entries()) {
      if (!hull || hull.length === 0) continue

      const color = colors.get(folder) || this.getDefaultColor(folder)

      // Create path for hull
      const pathData = this.hullToPath(hull)

      g.append('path')
        .attr('class', `hull-${this.sanitizeClassName(folder)}`)
        .attr('d', pathData)
        .attr('fill', color)
        .attr('stroke', color)
        .attr('stroke-width', 2)
        .attr('stroke-opacity', 0.5)
        .style('pointer-events', 'none')

      // Add folder label
      const centroid = this.calculateCentroid(hull)
      if (centroid) {
        g.append('text')
          .attr('x', centroid[0])
          .attr('y', centroid[1])
          .attr('text-anchor', 'middle')
          .attr('dy', '0.35em')
          .attr('font-size', 12)
          .attr('font-weight', 'bold')
          .attr('fill', color)
          .attr('fill-opacity', 0.6)
          .attr('pointer-events', 'none')
          .text(folder || 'root')
      }
    }
  }

  /**
   * Convert hull points to SVG path
   */
  private hullToPath(hull: [number, number][]): string {
    if (hull.length === 0) return ''

    // Use Catmull-Rom spline for smooth curves
    const points = hull.map((p) => p.join(','))
    const path = [`M${points[0]}`]

    for (let i = 0; i < points.length; i++) {
      const p0 = hull[(i - 1 + hull.length) % hull.length]
      const p1 = hull[i]
      const p2 = hull[(i + 1) % hull.length]
      const p3 = hull[(i + 2) % hull.length]

      // Catmull-Rom curve control points
      const cp1x = p1[0] + (p2[0] - p0[0]) / 6
      const cp1y = p1[1] + (p2[1] - p0[1]) / 6
      const cp2x = p2[0] - (p3[0] - p1[0]) / 6
      const cp2y = p2[1] - (p3[1] - p1[1]) / 6

      path.push(`C${cp1x},${cp1y} ${cp2x},${cp2y} ${p2[0]},${p2[1]}`)
    }

    path.push('Z')
    return path.join(' ')
  }

  /**
   * Calculate centroid of polygon
   * Uses shared utility
   */
  private calculateCentroid(hull: [number, number][]): [number, number] | null {
    return calcCentroid(hull)
  }

  /**
   * Expand hull by padding distance
   */
  private expandHull(hull: [number, number][], padding: number): [number, number][] {
    const centroid = this.calculateCentroid(hull)
    if (!centroid) return hull

    return hull.map((p) => {
      const dx = p[0] - centroid[0]
      const dy = p[1] - centroid[1]
      const dist = Math.sqrt(dx * dx + dy * dy)

      if (dist === 0) return p

      const scale = (dist + padding) / dist
      return [centroid[0] + dx * scale, centroid[1] + dy * scale] as [number, number]
    })
  }

  /**
   * Calculate bounding box for small groups
   */
  private calculateBoundingBox(nodes: NodePosition[], padding: number): [number, number][] {
    if (nodes.length === 0) return []

    let minX = Infinity,
      maxX = -Infinity,
      minY = Infinity,
      maxY = -Infinity

    for (const node of nodes) {
      minX = Math.min(minX, node.x)
      maxX = Math.max(maxX, node.x)
      minY = Math.min(minY, node.y)
      maxY = Math.max(maxY, node.y)
    }

    // Add padding and create rounded rectangle
    minX -= padding
    maxX += padding
    minY -= padding
    maxY += padding

    const radius = 10
    return [
      [minX + radius, minY],
      [maxX - radius, minY],
      [maxX, minY + radius],
      [maxX, maxY - radius],
      [maxX - radius, maxY],
      [minX + radius, maxY],
      [minX, maxY - radius],
      [minX, minY + radius],
    ]
  }

  /**
   * Get default color for folder based on hash
   * Uses shared utility for consistent coloring
   */
  private getDefaultColor(folder: string): string {
    const hash = stringHash(folder)
    return FOLDER_COLORS[hash % FOLDER_COLORS.length]
  }

  /**
   * Sanitize folder name for CSS class
   */
  private sanitizeClassName(folder: string): string {
    return folder.replace(/[^a-zA-Z0-9-_]/g, '_')
  }

  /**
   * Update hull positions as nodes move (for animation)
   */
  updateHullPositions(svg: any, nodes: NodePosition[], padding: number = 30): void {
    const hulls = this.calculateHulls(nodes, padding)
    const hullToPath = this.hullToPath.bind(this)

    svg.selectAll('.convex-hulls path').each(function (this: SVGPathElement, _d: unknown, i: number) {
      const hull = Array.from(hulls.values())[i]
      if (hull) {
        const path = d3.select(this)
        path.transition().duration(300).attr('d', hullToPath(hull))
      }
    })
  }
}
