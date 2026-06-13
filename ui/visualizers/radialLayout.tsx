import React, { useEffect, useRef } from 'react'
import * as d3 from 'd3'
import { CodeMap } from '@models/codeMap'
import { VisualizationConfig } from '@models/types'

interface RadialLayoutProps {
  codeMap: CodeMap
  config: VisualizationConfig
  onNodeSelect: (nodeId: string) => void
}

export const RadialLayout: React.FC<RadialLayoutProps> = ({ codeMap, config, onNodeSelect }) => {
  const containerRef = useRef<HTMLDivElement>(null)
  const svgRef = useRef<SVGSVGElement>(null)

  useEffect(() => {
    if (!containerRef.current || !svgRef.current || !codeMap) return

    const width = containerRef.current.clientWidth
    const height = containerRef.current.clientHeight
    const centerX = width / 2
    const centerY = height / 2

    // Get data
    const nodes = codeMap.getNodes().map((n) => n.data)
    const edges = codeMap.getEdges().map((e) => ({
      source: e.data.source,
      target: e.data.target,
      type: e.data.type,
    }))

    // Identify root nodes (no incoming edges or service/api nodes)
    const targetSet = new Set(edges.map((e) => e.target))
    const rootNodes = nodes.filter((n) => !targetSet.has(n.id) || n.type === 'module' || n.type === 'package')

    // Calculate radial positions
    const nodePositions = new Map<string, { angle: number; radius: number }>()
    const visited = new Set<string>()
    const queue: Array<{ nodeId: string; layer: number }> = rootNodes.map((n) => ({ nodeId: n.id, layer: 0 }))
    const layerMap = new Map<number, string[]>()

    // BFS to assign layers
    while (queue.length > 0) {
      const { nodeId, layer } = queue.shift()!
      if (visited.has(nodeId)) continue

      visited.add(nodeId)
      if (!layerMap.has(layer)) {
        layerMap.set(layer, [])
      }
      layerMap.get(layer)!.push(nodeId)

      // Add dependencies to queue
      const deps = edges.filter((e) => e.source === nodeId).map((e) => e.target)
      for (const dep of deps) {
        if (!visited.has(dep)) {
          queue.push({ nodeId: dep, layer: layer + 1 })
        }
      }
    }

    // Add unvisited nodes to outer layer
    const maxLayer = Math.max(...Array.from(layerMap.keys()), 0) + 1
    for (const node of nodes) {
      if (!visited.has(node.id)) {
        if (!layerMap.has(maxLayer)) {
          layerMap.set(maxLayer, [])
        }
        layerMap.get(maxLayer)!.push(node.id)
      }
    }

    // Calculate positions based on layer and angle distribution
    const radiiIncrement = Math.min(width, height) / (2 * (layerMap.size + 1))
    for (const [layer, nodeIds] of layerMap.entries()) {
      const radius = radiiIncrement * (layer + 1)
      const angleStep = (2 * Math.PI) / nodeIds.length

      for (let i = 0; i < nodeIds.length; i++) {
        const angle = angleStep * i - Math.PI / 2
        nodePositions.set(nodeIds[i], { angle, radius })
      }
    }

    // Convert to cartesian coordinates
    const positionedNodes = nodes.map((n) => {
      const pos = nodePositions.get(n.id) || { angle: 0, radius: radiiIncrement }
      const x = centerX + pos.radius * Math.cos(pos.angle)
      const y = centerY + pos.radius * Math.sin(pos.angle)
      return { ...n, x, y, layer: layerMap.get(0)?.includes(n.id) ? 0 : 1 }
    })

    // Clear SVG
    if (svgRef.current) {
      svgRef.current.innerHTML = ''
    }

    // Create SVG
    const svg = d3
      .select(svgRef.current)
      .attr('width', width)
      .attr('height', height)

    // Add background
    svg
      .append('rect')
      .attr('width', width)
      .attr('height', height)
      .attr('fill', '#1a1a1a')

    // Create group
    const g = svg.append('g').attr('class', 'radial-container')

    // Draw concentric circles (guides)
    for (let i = 1; i <= layerMap.size; i++) {
      g.append('circle')
        .attr('cx', centerX)
        .attr('cy', centerY)
        .attr('r', radiiIncrement * i)
        .attr('fill', 'none')
        .attr('stroke', '#333')
        .attr('stroke-width', 1)
        .attr('stroke-dasharray', '2,4')
    }

    // Draw edges
    const links = g
      .selectAll('line.link')
      .data(edges)
      .enter()
      .append('line')
      .attr('class', 'link')
      .attr('x1', (d: any) => {
        const source = positionedNodes.find((n) => n.id === d.source)
        return source?.x || centerX
      })
      .attr('y1', (d: any) => {
        const source = positionedNodes.find((n) => n.id === d.source)
        return source?.y || centerY
      })
      .attr('x2', (d: any) => {
        const target = positionedNodes.find((n) => n.id === d.target)
        return target?.x || centerX
      })
      .attr('y2', (d: any) => {
        const target = positionedNodes.find((n) => n.id === d.target)
        return target?.y || centerY
      })
      .attr('stroke', (d: any) => {
        if (d.type === 'calls' || d.type === 'imports') return '#3b82f6'
        if (d.type === 'inherits' || d.type === 'implements') return '#8b5cf6'
        return '#9ca3af'
      })
      .attr('stroke-opacity', 0.4)
      .attr('stroke-width', 1)
      .attr('marker-end', 'url(#radial-arrow)')

    // Add arrow marker
    svg
      .append('defs')
      .append('marker')
      .attr('id', 'radial-arrow')
      .attr('markerWidth', 10)
      .attr('markerHeight', 10)
      .attr('refX', 9)
      .attr('refY', 3)
      .attr('orient', 'auto')
      .append('polygon')
      .attr('points', '0 0, 10 3, 0 6')
      .attr('fill', '#999')

    // Draw nodes
    const nodeCircles = g
      .selectAll('circle.node')
      .data(positionedNodes)
      .enter()
      .append('circle')
      .attr('class', 'node')
      .attr('cx', (d: any) => d.x)
      .attr('cy', (d: any) => d.y)
      .attr('r', 6)
      .attr('fill', (d: any) => getRadialNodeColor(d, config.colorMode))
      .attr('stroke', '#fff')
      .attr('stroke-width', 2)
      .style('cursor', 'pointer')
      .on('click', (event: any, d: any) => {
        event.stopPropagation()
        onNodeSelect(d.id)
      })
      .on('mouseover', function () {
        d3.select(this).transition().duration(200).attr('r', 8).attr('stroke-width', 3)
      })
      .on('mouseout', function () {
        d3.select(this).transition().duration(200).attr('r', 6).attr('stroke-width', 2)
      })

    // Add labels
    if (config.showLabels) {
      g.selectAll('text.label')
        .data(positionedNodes)
        .enter()
        .append('text')
        .attr('class', 'label')
        .attr('x', (d: any) => d.x)
        .attr('y', (d: any) => d.y)
        .attr('text-anchor', 'middle')
        .attr('dy', '0.35em')
        .attr('font-size', 10)
        .attr('font-weight', '600')
        .attr('fill', '#fff')
        .attr('pointer-events', 'none')
        .style('text-shadow', '0 0 3px rgba(0,0,0,0.95)')
        .text((d: any) => (d.name.length > 10 ? d.name.substring(0, 8) + '..' : d.name))
    }

    // Add zoom
    const zoom = d3.zoom<SVGSVGElement, unknown>().on('zoom', (event: any) => {
      g.attr('transform', event.transform)
    })
    svg.call(zoom as any)
  }, [codeMap, config, onNodeSelect])

  return (
    <div ref={containerRef} className="visualizer-container" style={{ width: '100%', height: '100%' }}>
      <svg ref={svgRef} className="radial-graph" style={{ background: '#1a1a1a' }}></svg>
    </div>
  )
}

function getRadialNodeColor(node: any, colorMode: string): string {
  const typeColors: Record<string, string> = {
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

  if (colorMode === 'type') {
    return typeColors[node.type] || '#64748b'
  }

  if (colorMode === 'language') {
    const langColors: Record<string, string> = {
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
    return langColors[node.language] || '#64748b'
  }

  if (colorMode === 'complexity') {
    if (node.metrics) {
      const complexity = node.metrics.cyclomatic
      if (complexity > 10) return '#dc2626'
      if (complexity > 5) return '#f97316'
      if (complexity > 3) return '#eab308'
      return '#16a34a'
    }
    return '#64748b'
  }

  return '#64748b'
}
