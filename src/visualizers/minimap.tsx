import React, { useRef, useEffect, useCallback } from 'react'
import {
  calculateBounds as calcBounds,
  calculateUniformScale,
  TYPE_COLORS,
  DEFAULT_COLOR,
  type Bounds,
} from '../utils/visualizationUtils.js'

interface MinimapNode {
  id: string
  x: number
  y: number
  type?: string
}

interface MinimapEdge {
  source: { id: string } | string
  target: { id: string } | string
}

interface ViewportBounds {
  x0: number
  y0: number
  x1: number
  y1: number
}

interface ViewportChangeEvent {
  x: number
  y: number
  width: number
  height: number
}

export interface MinimapProps {
  nodes: MinimapNode[]
  edges: MinimapEdge[]
  width?: number
  height?: number
  viewportBounds?: ViewportBounds
  onViewportChange?: (bounds: ViewportChangeEvent) => void
}

const DEFAULT_VIEWPORT_SIZE = { width: 400, height: 300 }
const PADDING = 50

const MINIMAP_NODE_COLORS: Record<string, string> = {
  class: TYPE_COLORS.class ?? '#2196F3',
  function: TYPE_COLORS.function ?? '#4CAF50',
  method: TYPE_COLORS.method ?? '#4CAF50',
  interface: TYPE_COLORS.interface ?? '#FF9800',
  variable: TYPE_COLORS.variable ?? '#9C27B0',
}

const NODE_RADII: Record<string, number> = {
  class: 1.5,
  interface: 1.5,
  function: 1,
  method: 1,
}

const DEFAULT_NODE_RADIUS = 0.8

/**
 * Minimap component for large graph visualization.
 * Shows overview of entire graph with viewport indicator.
 * Allows quick navigation by clicking on minimap.
 */
export function Minimap({
  nodes,
  edges,
  width = 200,
  height = 200,
  viewportBounds,
  onViewportChange,
}: MinimapProps): React.ReactElement {
  const canvasRef = useRef<HTMLCanvasElement>(null)

  const getNodeId = useCallback((ref: { id: string } | string): string => {
    return typeof ref === 'string' ? ref : ref.id
  }, [])

  useEffect(() => {
    if (!canvasRef.current || nodes.length === 0) return

    const canvas = canvasRef.current
    const ctx = canvas.getContext('2d')
    if (!ctx) return

    const bounds = calculateBounds(nodes)
    const scale = calculateScale(bounds, width, height)

    renderCanvas(ctx, {
      nodes,
      edges,
      bounds,
      scale,
      viewportBounds,
      width,
      height,
      getNodeId,
    })
  }, [nodes, edges, width, height, viewportBounds, getNodeId])

  const handleCanvasClick = useCallback(
    (event: React.MouseEvent<HTMLCanvasElement>) => {
      if (!canvasRef.current || !onViewportChange || nodes.length === 0) return

      const canvas = canvasRef.current
      const rect = canvas.getBoundingClientRect()
      const clickX = event.clientX - rect.left
      const clickY = event.clientY - rect.top

      const bounds = calculateBounds(nodes)
      const scale = calculateScale(bounds, width, height)

      const worldX = clickX / scale + bounds.minX
      const worldY = clickY / scale + bounds.minY

      onViewportChange({
        x: worldX - DEFAULT_VIEWPORT_SIZE.width / 2,
        y: worldY - DEFAULT_VIEWPORT_SIZE.height / 2,
        width: DEFAULT_VIEWPORT_SIZE.width,
        height: DEFAULT_VIEWPORT_SIZE.height,
      })
    },
    [nodes, width, height, onViewportChange]
  )

  return (
    <div style={containerStyle}>
      <canvas
        ref={canvasRef}
        width={width}
        height={height}
        onClick={handleCanvasClick}
        style={canvasStyle}
      />
      <div style={labelStyle}>Overview</div>
    </div>
  )
}

interface RenderOptions {
  nodes: MinimapNode[]
  edges: MinimapEdge[]
  bounds: { minX: number; minY: number; maxX: number; maxY: number }
  scale: number
  viewportBounds?: ViewportBounds
  width: number
  height: number
  getNodeId: (ref: { id: string } | string) => string
}

function renderCanvas(ctx: CanvasRenderingContext2D, options: RenderOptions): void {
  const { nodes, edges, bounds, scale, viewportBounds, width, height, getNodeId } = options

  // Clear canvas
  ctx.fillStyle = '#ffffff'
  ctx.fillRect(0, 0, width, height)

  // Build node lookup for edge rendering
  const nodeMap = new Map(nodes.map((n) => [n.id, n]))

  // Draw edges
  ctx.strokeStyle = 'rgba(150, 150, 150, 0.3)'
  ctx.lineWidth = 0.5
  for (const edge of edges) {
    const source = nodeMap.get(getNodeId(edge.source))
    const target = nodeMap.get(getNodeId(edge.target))

    if (source && target) {
      ctx.beginPath()
      ctx.moveTo((source.x - bounds.minX) * scale, (source.y - bounds.minY) * scale)
      ctx.lineTo((target.x - bounds.minX) * scale, (target.y - bounds.minY) * scale)
      ctx.stroke()
    }
  }

  // Draw nodes
  for (const node of nodes) {
    const x = (node.x - bounds.minX) * scale
    const y = (node.y - bounds.minY) * scale
    const radius = NODE_RADII[node.type ?? ''] ?? DEFAULT_NODE_RADIUS

    ctx.fillStyle = MINIMAP_NODE_COLORS[node.type ?? ''] ?? DEFAULT_COLOR
    ctx.beginPath()
    ctx.arc(x, y, radius, 0, Math.PI * 2)
    ctx.fill()
  }

  // Draw viewport indicator
  if (viewportBounds) {
    const vx0 = (viewportBounds.x0 - bounds.minX) * scale
    const vy0 = (viewportBounds.y0 - bounds.minY) * scale
    const vx1 = (viewportBounds.x1 - bounds.minX) * scale
    const vy1 = (viewportBounds.y1 - bounds.minY) * scale

    ctx.strokeStyle = '#ff6b6b'
    ctx.lineWidth = 2
    ctx.strokeRect(vx0, vy0, vx1 - vx0, vy1 - vy0)

    ctx.fillStyle = 'rgba(255, 107, 107, 0.1)'
    ctx.fillRect(vx0, vy0, vx1 - vx0, vy1 - vy0)
  }

  // Draw border
  ctx.strokeStyle = '#cccccc'
  ctx.lineWidth = 1
  ctx.strokeRect(0, 0, width, height)
}

const containerStyle: React.CSSProperties = {
  position: 'absolute',
  bottom: 10,
  right: 10,
  border: '1px solid #ccc',
  backgroundColor: '#fff',
  borderRadius: '4px',
  boxShadow: '0 2px 4px rgba(0, 0, 0, 0.1)',
  cursor: 'pointer',
}

const canvasStyle: React.CSSProperties = {
  display: 'block',
  userSelect: 'none',
}

const labelStyle: React.CSSProperties = {
  fontSize: '12px',
  padding: '4px 6px',
  textAlign: 'center',
  borderTop: '1px solid #eee',
  backgroundColor: '#f9f9f9',
}

/**
 * Calculate bounding box of all nodes with padding
 * Uses shared utility with minimap-specific padding
 */
function calculateBounds(nodes: Array<{ x: number; y: number }>): Bounds {
  return calcBounds(nodes, PADDING)
}

/**
 * Calculate uniform scale factor to fit bounds into canvas
 * Uses shared utility
 */
function calculateScale(bounds: Bounds, width: number, height: number): number {
  return calculateUniformScale(bounds, width, height)
}
