import React, { useEffect, useRef, useMemo } from 'react'
import { CodeMap } from '@models/codeMap'
import { VisualizationConfig } from '@models/types'

interface MatrixVisualizationProps {
  codeMap: CodeMap
  config: VisualizationConfig
  onNodeSelect: (nodeId: string) => void
}

export const MatrixVisualization: React.FC<MatrixVisualizationProps> = ({ codeMap, config, onNodeSelect }) => {
  const containerRef = useRef<HTMLDivElement>(null)
  const canvasRef = useRef<HTMLCanvasElement>(null)

  const nodes = useMemo(() => codeMap.getNodes(), [codeMap])
  const edges = useMemo(() => codeMap.getEdges(), [codeMap])

  useEffect(() => {
    if (!containerRef.current || !canvasRef.current || nodes.length === 0) return

    const canvas = canvasRef.current
    const ctx = canvas.getContext('2d')
    if (!ctx) return

    const containerWidth = containerRef.current.clientWidth
    const containerHeight = containerRef.current.clientHeight

    canvas.width = containerWidth
    canvas.height = containerHeight

    const cellSize = Math.max(4, Math.min(20, containerWidth / (nodes.length + 3)))
    const margin = 80
    const startX = margin
    const startY = margin

    // Clear canvas
    ctx.fillStyle = '#1a1a1a'
    ctx.fillRect(0, 0, canvas.width, canvas.height)

    // Build edge map for quick lookup
    const edgeMap = new Map<string, any>()
    edges.forEach((edge) => {
      const key = `${edge.data.source},${edge.data.target}`
      edgeMap.set(key, edge)
    })

    // Draw grid
    ctx.strokeStyle = '#333'
    ctx.lineWidth = 0.5

    // Vertical lines
    for (let i = 0; i <= nodes.length; i++) {
      const x = startX + i * cellSize
      ctx.beginPath()
      ctx.moveTo(x, startY)
      ctx.lineTo(x, startY + nodes.length * cellSize)
      ctx.stroke()
    }

    // Horizontal lines
    for (let i = 0; i <= nodes.length; i++) {
      const y = startY + i * cellSize
      ctx.beginPath()
      ctx.moveTo(startX, y)
      ctx.lineTo(startX + nodes.length * cellSize, y)
      ctx.stroke()
    }

    // Draw cells
    nodes.forEach((source, i) => {
      nodes.forEach((target, j) => {
        const key = `${source.data.id},${target.data.id}`
        const edge = edgeMap.get(key)

        const x = startX + j * cellSize
        const y = startY + i * cellSize

        if (edge) {
          ctx.fillStyle = '#06b6d4'
          ctx.fillRect(x, y, cellSize, cellSize)
        } else if (i === j) {
          ctx.fillStyle = '#8b5cf6'
          ctx.fillRect(x, y, cellSize, cellSize)
        }
      })
    })

    // Draw row labels (Y-axis)
    ctx.fillStyle = '#fff'
    ctx.font = 'bold 10px sans-serif'
    ctx.textAlign = 'right'
    ctx.textBaseline = 'middle'

    nodes.forEach((node, i) => {
      const y = startY + i * cellSize + cellSize / 2
      const label = node.data.name.substring(0, 10)
      ctx.fillText(label, startX - 10, y)
    })

    // Draw column labels (X-axis)
    ctx.textAlign = 'center'
    ctx.textBaseline = 'top'

    nodes.forEach((node, j) => {
      const x = startX + j * cellSize + cellSize / 2
      const label = node.data.name.substring(0, 8)
      ctx.save()
      ctx.translate(x, startY - 15)
      ctx.rotate(-Math.PI / 4)
      ctx.fillText(label, 0, 0)
      ctx.restore()
    })

    // Legend
    const legendX = startX
    const legendY = startY + nodes.length * cellSize + 30

    ctx.fillStyle = '#8b5cf6'
    ctx.fillRect(legendX, legendY, 12, 12)
    ctx.fillStyle = '#fff'
    ctx.font = '11px sans-serif'
    ctx.textAlign = 'left'
    ctx.fillText('Self', legendX + 20, legendY + 6)

    ctx.fillStyle = '#06b6d4'
    ctx.fillRect(legendX + 100, legendY, 12, 12)
    ctx.fillStyle = '#fff'
    ctx.fillText('Dependency', legendX + 120, legendY + 6)
  }, [nodes, edges, config])

  return (
    <div ref={containerRef} className='visualizer-container' style={{ width: '100%', height: '100%', overflow: 'auto' }}>
      <canvas ref={canvasRef} style={{ display: 'block' }}></canvas>
    </div>
  )
}
