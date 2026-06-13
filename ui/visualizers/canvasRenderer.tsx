/**
 * Canvas-based graph renderer for 100-1000 nodes
 * 10-100x faster than SVG for large graphs
 */

import React, { useRef, useEffect, useState, useCallback } from 'react'

export interface CanvasNode {
  id: string
  x: number
  y: number
  r?: number
  color?: string
  label?: string
}

export interface CanvasEdge {
  source: string
  target: string
  color?: string
  width?: number
}

export interface CanvasRendererProps {
  nodes: CanvasNode[]
  edges: CanvasEdge[]
  width: number
  height: number
  onNodeClick?: (nodeId: string) => void
  onNodeHover?: (nodeId: string | null) => void
  backgroundColor?: string
  showLabels?: boolean
  zoom?: number
  pan?: { x: number; y: number }
}

interface RenderState {
  nodeMap: Map<string, CanvasNode>
  transform: { x: number; y: number; scale: number }
  hoveredNode: string | null
  selectedNode: string | null
}

const DEFAULT_NODE_RADIUS = 5
const DEFAULT_NODE_COLOR = '#4d9fff'
const HOVER_COLOR = '#ff6b6b'
const EDGE_COLOR = '#cccccc'
const HIT_THRESHOLD = 20

export function CanvasRenderer({
  nodes,
  edges,
  width,
  height,
  onNodeClick,
  onNodeHover,
  backgroundColor = '#ffffff',
  showLabels = true,
  zoom = 1.0,
  pan = { x: 0, y: 0 },
}: CanvasRendererProps): React.ReactElement {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const [state, setState] = useState<RenderState>({
    nodeMap: new Map(),
    transform: { x: pan.x, y: pan.y, scale: zoom },
    hoveredNode: null,
    selectedNode: null,
  })

  useEffect(() => {
    const nodeMap = new Map(nodes.map((n) => [n.id, n]))
    setState((prev) => ({
      ...prev,
      nodeMap,
      transform: { x: pan.x, y: pan.y, scale: zoom },
    }))
  }, [nodes, zoom, pan])

  useEffect(() => {
    const canvas = canvasRef.current
    const ctx = canvas?.getContext('2d')
    if (!canvas || !ctx) return

    const { transform, nodeMap, hoveredNode, selectedNode } = state

    ctx.fillStyle = backgroundColor
    ctx.fillRect(0, 0, width, height)

    ctx.strokeStyle = EDGE_COLOR
    ctx.lineWidth = 1
    for (const edge of edges) {
      const source = nodeMap.get(edge.source)
      const target = nodeMap.get(edge.target)
      if (!source || !target) continue

      ctx.beginPath()
      ctx.moveTo(source.x * transform.scale + transform.x, source.y * transform.scale + transform.y)
      ctx.lineTo(target.x * transform.scale + transform.x, target.y * transform.scale + transform.y)
      ctx.stroke()
    }

    for (const node of nodeMap.values()) {
      const x = node.x * transform.scale + transform.x
      const y = node.y * transform.scale + transform.y
      const r = (node.r ?? DEFAULT_NODE_RADIUS) * transform.scale

      ctx.fillStyle = hoveredNode === node.id ? HOVER_COLOR : (node.color ?? DEFAULT_NODE_COLOR)
      ctx.beginPath()
      ctx.arc(x, y, r, 0, Math.PI * 2)
      ctx.fill()

      if (selectedNode === node.id) {
        ctx.strokeStyle = '#000000'
        ctx.lineWidth = 2
        ctx.stroke()
      }

      if (showLabels && node.label && transform.scale > 0.5) {
        ctx.fillStyle = '#000000'
        ctx.font = `${Math.max(10, 12 * transform.scale)}px Arial`
        ctx.textAlign = 'center'
        ctx.textBaseline = 'middle'
        ctx.fillText(node.label.substring(0, 20), x, y + r + 20)
      }
    }
  }, [state, edges, backgroundColor, width, height, showLabels])

  const findNodeAt = useCallback(
    (clientX: number, clientY: number): string | null => {
      const canvas = canvasRef.current
      if (!canvas) return null

      const rect = canvas.getBoundingClientRect()
      const x = clientX - rect.left
      const y = clientY - rect.top
      const { transform, nodeMap } = state

      for (const node of nodeMap.values()) {
        const nodeX = node.x * transform.scale + transform.x
        const nodeY = node.y * transform.scale + transform.y
        const dist = Math.sqrt((x - nodeX) ** 2 + (y - nodeY) ** 2)
        if (dist < HIT_THRESHOLD) return node.id
      }

      return null
    },
    [state]
  )

  const handleMouseMove = useCallback(
    (e: React.MouseEvent<HTMLCanvasElement>) => {
      const nodeId = findNodeAt(e.clientX, e.clientY)
      if (nodeId !== state.hoveredNode) {
        setState((prev) => ({ ...prev, hoveredNode: nodeId }))
        onNodeHover?.(nodeId)
      }
    },
    [findNodeAt, state.hoveredNode, onNodeHover]
  )

  const handleClick = useCallback(
    (e: React.MouseEvent<HTMLCanvasElement>) => {
      const nodeId = findNodeAt(e.clientX, e.clientY)
      setState((prev) => ({ ...prev, selectedNode: nodeId }))
      if (nodeId) onNodeClick?.(nodeId)
    },
    [findNodeAt, onNodeClick]
  )

  const handleMouseLeave = useCallback(() => {
    setState((prev) => ({ ...prev, hoveredNode: null }))
    onNodeHover?.(null)
  }, [onNodeHover])

  return (
    <canvas
      ref={canvasRef}
      width={width}
      height={height}
      onMouseMove={handleMouseMove}
      onClick={handleClick}
      onMouseLeave={handleMouseLeave}
      style={{
        border: '1px solid #ddd',
        cursor: state.hoveredNode ? 'pointer' : 'default',
      }}
    />
  )
}

export default CanvasRenderer
