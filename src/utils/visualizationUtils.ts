/**
 * Shared visualization utilities
 * Consolidates common functions used across visualization components
 */

// Color palettes for different visualization modes

export const TYPE_COLORS: Record<string, string> = {
  file: '#1f2937',
  class: '#a855f7',
  function: '#0ea5e9',
  method: '#10b981',
  variable: '#f59e0b',
  module: '#3b82f6',
  package: '#8b5cf6',
  interface: '#ec4899',
  enum: '#f97316',
  constant: '#06b6d4',
}

export const LAYER_COLORS: Record<string, string> = {
  ui: '#4d9fff',
  components: '#22d3ee',
  services: '#a78bfa',
  data: '#ff9f43',
  utils: '#00ff9d',
  config: '#ec4899',
  other: '#64748b',
}

export const LANGUAGE_COLORS: Record<string, string> = {
  python: '#3b82f6',
  javascript: '#fbbf24',
  typescript: '#3b82f6',
  java: '#ea580c',
  go: '#00d9ff',
  rust: '#f14e28',
  cpp: '#ed1c24',
  ruby: '#cc342d',
  php: '#777bb4',
}

export const FOLDER_COLORS = [
  '#4d9fff',
  '#22d3ee',
  '#a78bfa',
  '#ff9f43',
  '#00ff9d',
  '#ec4899',
  '#84cc16',
  '#f43f5e',
]

export const DEFAULT_COLOR = '#64748b'

/**
 * Generate a hash from a string (DJB2 variant)
 * Used for consistent color assignment based on folder names
 */
export function stringHash(str: string): number {
  let hash = 0
  for (let i = 0; i < str.length; i++) {
    hash = ((hash << 5) - hash) + str.charCodeAt(i)
    hash = hash & hash // Convert to 32-bit integer
  }
  return Math.abs(hash)
}

/**
 * Get color for a folder based on its name hash
 */
export function getFolderColor(folder: string): string {
  const hash = stringHash(folder)
  return FOLDER_COLORS[hash % FOLDER_COLORS.length]
}

/**
 * Get color based on node type
 */
export function getTypeColor(type: string): string {
  return TYPE_COLORS[type] ?? DEFAULT_COLOR
}

/**
 * Get color based on complexity score
 */
export function getComplexityColor(complexity: number): string {
  if (complexity > 10) return '#dc2626' // Red - high
  if (complexity > 5) return '#f97316' // Orange
  if (complexity > 3) return '#eab308' // Yellow
  return '#16a34a' // Green - low
}

/**
 * Get color based on churn (git activity) score
 */
export function getChurnColor(churn: number): string {
  if (churn > 20) return '#dc2626' // Red - high churn
  if (churn > 10) return '#f97316' // Orange
  if (churn > 5) return '#eab308' // Yellow
  if (churn > 0) return '#22c55e' // Green
  return DEFAULT_COLOR // Gray - no changes
}

// Bounds calculation utilities

export interface Bounds {
  minX: number
  minY: number
  maxX: number
  maxY: number
}

export interface RectBounds {
  x: number
  y: number
  width: number
  height: number
}

export interface Point2D {
  x: number
  y: number
}

/**
 * Calculate bounding box for a set of points
 */
export function calculateBounds(points: Point2D[], padding: number = 0): Bounds {
  let minX = Infinity
  let minY = Infinity
  let maxX = -Infinity
  let maxY = -Infinity

  for (const point of points) {
    minX = Math.min(minX, point.x)
    minY = Math.min(minY, point.y)
    maxX = Math.max(maxX, point.x)
    maxY = Math.max(maxY, point.y)
  }

  return {
    minX: minX - padding,
    minY: minY - padding,
    maxX: maxX + padding,
    maxY: maxY + padding,
  }
}

/**
 * Calculate uniform scale factor to fit bounds within dimensions
 */
export function calculateUniformScale(bounds: Bounds, width: number, height: number): number {
  const boundsWidth = bounds.maxX - bounds.minX
  const boundsHeight = bounds.maxY - bounds.minY

  if (boundsWidth === 0 || boundsHeight === 0) return 1

  return Math.min(width / boundsWidth, height / boundsHeight)
}

/**
 * Calculate centroid of a polygon
 */
export function calculateCentroid(points: Array<[number, number]>): [number, number] | null {
  if (points.length === 0) return null

  let sumX = 0
  let sumY = 0

  for (const [x, y] of points) {
    sumX += x
    sumY += y
  }

  return [sumX / points.length, sumY / points.length]
}

/**
 * Check if two bounds rectangles intersect
 */
export function boundsIntersect(a: Bounds, b: Bounds): boolean {
  return !(
    a.maxX < b.minX ||
    a.minX > b.maxX ||
    a.maxY < b.minY ||
    a.minY > b.maxY
  )
}

/**
 * Calculate Euclidean distance between two points
 */
export function distance(x1: number, y1: number, x2: number, y2: number): number {
  const dx = x2 - x1
  const dy = y2 - y1
  return Math.sqrt(dx * dx + dy * dy)
}

/**
 * Calculate squared distance (faster when exact distance not needed)
 */
export function distanceSquared(x1: number, y1: number, x2: number, y2: number): number {
  const dx = x2 - x1
  const dy = y2 - y1
  return dx * dx + dy * dy
}

/**
 * Clamp a value between min and max
 */
export function clamp(value: number, min: number, max: number): number {
  return Math.max(min, Math.min(max, value))
}

/**
 * Linear interpolation between two values
 */
export function lerp(a: number, b: number, t: number): number {
  return a + (b - a) * t
}

/**
 * Transform world coordinates to screen coordinates
 */
export function worldToScreen(
  worldX: number,
  worldY: number,
  bounds: Bounds,
  scale: number,
  offsetX: number = 0,
  offsetY: number = 0
): { x: number; y: number } {
  return {
    x: (worldX - bounds.minX) * scale + offsetX,
    y: (worldY - bounds.minY) * scale + offsetY,
  }
}

/**
 * Transform screen coordinates to world coordinates
 */
export function screenToWorld(
  screenX: number,
  screenY: number,
  bounds: Bounds,
  scale: number,
  offsetX: number = 0,
  offsetY: number = 0
): { x: number; y: number } {
  return {
    x: (screenX - offsetX) / scale + bounds.minX,
    y: (screenY - offsetY) / scale + bounds.minY,
  }
}
