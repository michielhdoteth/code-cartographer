/**
 * Quadtree spatial indexing for O(log n) proximity queries
 * Used for viewport culling and collision detection
 */

export interface Point {
  x: number
  y: number
  id: string
}

export interface Bounds {
  x: number
  y: number
  width: number
  height: number
}

const MAX_POINTS = 4
const MAX_DEPTH = 8

class QuadTreeNode {
  bounds: Bounds
  points: Point[] = []
  children: QuadTreeNode[] | null = null
  private depth: number

  constructor(bounds: Bounds, depth = 0) {
    this.bounds = bounds
    this.depth = depth
  }

  private containsPoint(point: Point): boolean {
    return (
      point.x >= this.bounds.x &&
      point.x <= this.bounds.x + this.bounds.width &&
      point.y >= this.bounds.y &&
      point.y <= this.bounds.y + this.bounds.height
    )
  }

  private intersectsBounds(bounds: Bounds): boolean {
    return !(
      bounds.x + bounds.width < this.bounds.x ||
      bounds.x > this.bounds.x + this.bounds.width ||
      bounds.y + bounds.height < this.bounds.y ||
      bounds.y > this.bounds.y + this.bounds.height
    )
  }

  insert(point: Point): boolean {
    if (!this.containsPoint(point)) return false

    if (this.points.length < MAX_POINTS || this.depth >= MAX_DEPTH) {
      this.points.push(point)
      return true
    }

    if (!this.children) this.subdivide()

    for (const child of this.children!) {
      if (child.insert(point)) return true
    }

    return false
  }

  query(bounds: Bounds, found: Point[] = []): Point[] {
    if (!this.intersectsBounds(bounds)) return found

    for (const point of this.points) {
      if (this.pointInBounds(point, bounds)) {
        found.push(point)
      }
    }

    this.children?.forEach((child) => child.query(bounds, found))
    return found
  }

  private pointInBounds(point: Point, bounds: Bounds): boolean {
    return (
      point.x >= bounds.x &&
      point.x <= bounds.x + bounds.width &&
      point.y >= bounds.y &&
      point.y <= bounds.y + bounds.height
    )
  }

  findNearest(x: number, y: number, maxDistance = Infinity): Point | null {
    let nearest: Point | null = null
    let minDist = maxDistance

    this.traverseNear(x, y, maxDistance, (point, dist) => {
      if (dist < minDist) {
        minDist = dist
        nearest = point
      }
    })

    return nearest
  }

  findWithinDistance(x: number, y: number, distance: number): Point[] {
    const bounds: Bounds = {
      x: x - distance,
      y: y - distance,
      width: distance * 2,
      height: distance * 2,
    }

    const distSq = distance * distance
    return this.query(bounds).filter((point) => {
      const dx = point.x - x
      const dy = point.y - y
      return dx * dx + dy * dy <= distSq
    })
  }

  clear(): void {
    this.points = []
    this.children = null
  }

  private subdivide(): void {
    const { x, y, width, height } = this.bounds
    const w = width / 2
    const h = height / 2
    const d = this.depth + 1

    this.children = [
      new QuadTreeNode({ x, y, width: w, height: h }, d),
      new QuadTreeNode({ x: x + w, y, width: w, height: h }, d),
      new QuadTreeNode({ x, y: y + h, width: w, height: h }, d),
      new QuadTreeNode({ x: x + w, y: y + h, width: w, height: h }, d),
    ]

    const oldPoints = this.points
    this.points = []
    oldPoints.forEach((p) => this.children!.some((child) => child.insert(p)))
  }

  private traverseNear(
    x: number,
    y: number,
    maxDist: number,
    callback: (point: Point, distance: number) => void
  ): void {
    for (const point of this.points) {
      const dx = point.x - x
      const dy = point.y - y
      const dist = Math.sqrt(dx * dx + dy * dy)
      if (dist <= maxDist) callback(point, dist)
    }

    this.children?.forEach((child) => {
      if (this.boundsIntersectCircle(child.bounds, x, y, maxDist)) {
        child.traverseNear(x, y, maxDist, callback)
      }
    })
  }

  private boundsIntersectCircle(bounds: Bounds, cx: number, cy: number, radius: number): boolean {
    const closestX = Math.max(bounds.x, Math.min(cx, bounds.x + bounds.width))
    const closestY = Math.max(bounds.y, Math.min(cy, bounds.y + bounds.height))
    const dx = cx - closestX
    const dy = cy - closestY
    return dx * dx + dy * dy < radius * radius
  }
}

export class SpatialIndex {
  private root: QuadTreeNode
  private allPoints: Map<string, Point> = new Map()

  constructor(bounds: Bounds) {
    this.root = new QuadTreeNode(bounds)
  }

  add(point: Point): void {
    this.root.insert(point)
    this.allPoints.set(point.id, point)
  }

  addAll(points: Point[]): void {
    points.forEach((p) => this.add(p))
  }

  query(bounds: Bounds): Point[] {
    return this.root.query(bounds)
  }

  findNearest(x: number, y: number, maxDistance?: number): Point | null {
    return this.root.findNearest(x, y, maxDistance)
  }

  findWithinDistance(x: number, y: number, distance: number): Point[] {
    return this.root.findWithinDistance(x, y, distance)
  }

  getAll(): Point[] {
    return Array.from(this.allPoints.values())
  }

  clear(): void {
    this.root.clear()
    this.allPoints.clear()
  }

  getStats(): { pointCount: number; estimatedDepth: number } {
    return {
      pointCount: this.allPoints.size,
      estimatedDepth: Math.ceil(Math.log2(this.allPoints.size / MAX_POINTS)) + 1,
    }
  }
}
