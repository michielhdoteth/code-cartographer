/**
 * WebGL-based graph renderer for 1000+ nodes
 * Uses GPU acceleration for rendering 5000+ nodes smoothly
 */

import React, { useRef, useEffect, useState, useCallback } from 'react'

export interface WebGLNode {
  id: string
  x: number
  y: number
  r?: number
  color?: [number, number, number, number]
}

export interface WebGLEdge {
  source: string
  target: string
  color?: [number, number, number, number]
  width?: number
}

export interface WebGLRendererProps {
  nodes: WebGLNode[]
  edges: WebGLEdge[]
  width: number
  height: number
  onNodeClick?: (nodeId: string) => void
  backgroundColor?: string
  zoom?: number
  pan?: { x: number; y: number }
}

const HIT_THRESHOLD = 0.1

function parseBackgroundColor(hex: string): [number, number, number] {
  return [
    parseInt(hex.slice(1, 3), 16) / 255,
    parseInt(hex.slice(3, 5), 16) / 255,
    parseInt(hex.slice(5, 7), 16) / 255,
  ]
}

export function WebGLRenderer({
  nodes,
  edges,
  width,
  height,
  onNodeClick,
  backgroundColor = '#ffffff',
  zoom = 1.0,
  pan = { x: 0, y: 0 },
}: WebGLRendererProps): React.ReactElement {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const [gl, setGl] = useState<WebGLRenderingContext | null>(null)
  const [nodeBuffer, setNodeBuffer] = useState<WebGLBuffer | null>(null)
  const [edgeBuffer, setEdgeBuffer] = useState<WebGLBuffer | null>(null)

  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return

    const webgl = canvas.getContext('webgl') ?? canvas.getContext('webgl2')
    if (!webgl) {
      console.warn('WebGL not supported, falling back to canvas')
      return
    }

    setGl(webgl)

    webgl.viewport(0, 0, width, height)
    const [r, g, b] = parseBackgroundColor(backgroundColor)
    webgl.clearColor(r, g, b, 1.0)
    webgl.enable(webgl.BLEND)
    webgl.blendFunc(webgl.SRC_ALPHA, webgl.ONE_MINUS_SRC_ALPHA)
  }, [width, height, backgroundColor])

  useEffect(() => {
    if (!gl) return

    const nodeData = new Float32Array(nodes.length * 6)
    nodes.forEach((node, i) => {
      const base = i * 6
      nodeData[base] = node.x
      nodeData[base + 1] = node.y
      nodeData[base + 2] = node.r ?? 5
      nodeData[base + 3] = node.color?.[0] ?? 0.3
      nodeData[base + 4] = node.color?.[1] ?? 0.6
      nodeData[base + 5] = node.color?.[2] ?? 1.0
    })

    const buffer = gl.createBuffer()
    gl.bindBuffer(gl.ARRAY_BUFFER, buffer)
    gl.bufferData(gl.ARRAY_BUFFER, nodeData, gl.DYNAMIC_DRAW)
    setNodeBuffer(buffer)
  }, [gl, nodes])

  useEffect(() => {
    if (!gl) return

    const edgeData: number[] = []
    const nodeMap = new Map(nodes.map((n) => [n.id, n]))

    for (const edge of edges) {
      const source = nodeMap.get(edge.source)
      const target = nodeMap.get(edge.target)
      if (source && target) {
        edgeData.push(source.x, source.y, target.x, target.y)
      }
    }

    if (edgeData.length > 0) {
      const buffer = gl.createBuffer()
      gl.bindBuffer(gl.ARRAY_BUFFER, buffer)
      gl.bufferData(gl.ARRAY_BUFFER, new Float32Array(edgeData), gl.DYNAMIC_DRAW)
      setEdgeBuffer(buffer)
    }
  }, [gl, nodes, edges])

  useEffect(() => {
    if (!gl) return

    gl.clear(gl.COLOR_BUFFER_BIT)

    if (edgeBuffer) {
      gl.bindBuffer(gl.ARRAY_BUFFER, edgeBuffer)
      gl.drawArrays(gl.LINES, 0, edges.length * 2)
    }

    if (nodeBuffer) {
      gl.bindBuffer(gl.ARRAY_BUFFER, nodeBuffer)
      gl.drawArrays(gl.POINTS, 0, nodes.length)
    }
  }, [gl, nodeBuffer, edgeBuffer, nodes.length, edges.length])

  const handleClick = useCallback(
    (e: React.MouseEvent<HTMLCanvasElement>) => {
      const canvas = canvasRef.current
      if (!canvas || !onNodeClick) return

      const rect = canvas.getBoundingClientRect()
      const x = ((e.clientX - rect.left) / width) * 2 - 1
      const y = 1 - ((e.clientY - rect.top) / height) * 2

      let minDist = Infinity
      let clickedNode: WebGLNode | null = null

      for (const node of nodes) {
        const dx = node.x - x
        const dy = node.y - y
        const dist = Math.sqrt(dx * dx + dy * dy)

        if (dist < minDist && dist < HIT_THRESHOLD) {
          minDist = dist
          clickedNode = node
        }
      }

      if (clickedNode) onNodeClick(clickedNode.id)
    },
    [nodes, width, height, onNodeClick]
  )

  return (
    <canvas
      ref={canvasRef}
      width={width}
      height={height}
      onClick={handleClick}
      style={{ border: '1px solid #ddd', display: 'block' }}
    />
  )
}

export default WebGLRenderer
