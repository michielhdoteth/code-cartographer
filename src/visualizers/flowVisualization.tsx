import React, { useEffect, useRef } from 'react'
import * as d3 from 'd3'
import { CodeMap } from '@models/codeMap'
import { VisualizationConfig } from '@models/types'

interface FlowVisualizationProps {
  codeMap: CodeMap
  config: VisualizationConfig
  onNodeSelect: (nodeId: string) => void
}

export const FlowVisualization: React.FC<FlowVisualizationProps> = ({ codeMap, config, onNodeSelect }) => {
  const containerRef = useRef<HTMLDivElement>(null)
  const svgRef = useRef<SVGSVGElement>(null)

  useEffect(() => {
    if (!containerRef.current || !codeMap) return

    const width = containerRef.current.clientWidth
    const height = containerRef.current.clientHeight

    const nodes = codeMap.getNodes()
    const edges = codeMap.getEdges()

    // Group nodes by type for layered layout
    const layers: Map<string, any[]> = new Map()
    const typeOrder = ['file', 'class', 'function', 'method', 'variable']

    nodes.forEach((node) => {
      const type = node.data.type || 'unknown'
      if (!layers.has(type)) {
        layers.set(type, [])
      }
      layers.get(type)!.push({
        id: node.data.id,
        name: node.data.name,
        type: type,
        language: node.data.language,
        metrics: node.data.metrics,
      })
    })

    // Clear previous
    if (svgRef.current) {
      svgRef.current.innerHTML = ''
    }

    const svg = d3
      .select(svgRef.current)
      .attr('width', width)
      .attr('height', height)

    try {
      const g = svg.append('g').attr('class', 'flow-container')

      const layerWidth = width / (layers.size + 1)
      let layerIndex = 0

      // Draw each layer
      const layerNodeMap = new Map<string, { x: number; y: number }>()

      for (const [type, layerNodes] of layers) {
        const x = (layerIndex + 1) * layerWidth
        const nodeHeight = height / (layerNodes.length + 1)

        layerNodes.forEach((node, i) => {
          const y = (i + 1) * nodeHeight

          layerNodeMap.set(node.id, { x, y })

          // Add circle
          g.append('circle')
            .attr('cx', x)
            .attr('cy', y)
            .attr('r', 6)
            .attr('fill', getFlowNodeColor(node, config.colorMode))
            .attr('stroke', '#000')
            .attr('stroke-width', 2)
            .style('cursor', 'pointer')
            .on('click', () => onNodeSelect(node.id))

          // Add label
          if (config.showLabels) {
            g.append('text')
              .attr('x', x + 12)
              .attr('y', y + 4)
              .text(node.name)
              .attr('font-size', 10)
              .attr('font-weight', 'bold')
              .attr('fill', '#fff')
              .attr('pointer-events', 'none')
              .style('text-shadow', '0 0 2px rgba(0,0,0,0.8)')
          }

          // Add type label
          g.append('text')
            .attr('x', x)
            .attr('y', y - 18)
            .text(type)
            .attr('font-size', 9)
            .attr('text-anchor', 'middle')
            .attr('fill', '#999')
            .attr('pointer-events', 'none')
        })

        layerIndex++
      }

      // Draw edges
      edges.forEach((edge) => {
        const source = layerNodeMap.get(edge.data.source)
        const target = layerNodeMap.get(edge.data.target)

        if (source && target) {
          // Create curved path
          const midX = (source.x + target.x) / 2

          g.append('path')
            .attr('d', `M ${source.x} ${source.y} Q ${midX} ${(source.y + target.y) / 2} ${target.x} ${target.y}`)
            .attr('fill', 'none')
            .attr('stroke', '#999')
            .attr('stroke-opacity', 0.4)
            .attr('stroke-width', 1)
            .attr('marker-end', 'url(#arrowhead)')

          // Add arrow marker definition
          if (!document.querySelector('#arrowhead')) {
            svg
              .append('defs')
              .append('marker')
              .attr('id', 'arrowhead')
              .attr('markerWidth', 10)
              .attr('markerHeight', 10)
              .attr('refX', 9)
              .attr('refY', 3)
              .attr('orient', 'auto')
              .append('polygon')
              .attr('points', '0 0, 10 3, 0 6')
              .attr('fill', '#999')
          }
        }
      })

      // Add zoom
      const zoom = d3.zoom<SVGSVGElement, unknown>().on('zoom', (event: any) => {
        g.attr('transform', event.transform)
      })

      svg.call(zoom as any)
    } catch (error) {
      console.error('Error rendering flow:', error)
    }
  }, [codeMap, config, onNodeSelect])

  return (
    <div ref={containerRef} className='visualizer-container' style={{ width: '100%', height: '100%' }}>
      <svg ref={svgRef} className='graph'></svg>
    </div>
  )
}

function getFlowNodeColor(node: any, colorMode: string): string {
  const typeColors: Record<string, string> = {
    file: '#3b82f6',
    class: '#8b5cf6',
    function: '#06b6d4',
    method: '#10b981',
    variable: '#f59e0b',
  }

  if (colorMode === 'type') {
    return typeColors[node.type] || '#94a3b8'
  }

  if (colorMode === 'language') {
    const langColors: Record<string, string> = {
      python: '#3572a5',
      javascript: '#f1e05a',
      typescript: '#3178c6',
      java: '#b07219',
      go: '#00add8',
      rust: '#ce422b',
      cpp: '#f34b7d',
      ruby: '#cc342d',
      php: '#777bb4',
    }
    return langColors[node.language] || '#94a3b8'
  }

  if (colorMode === 'complexity') {
    if (node.metrics) {
      const complexity = node.metrics.cyclomatic
      if (complexity > 10) return '#ef4444'
      if (complexity > 5) return '#f59e0b'
      return '#10b981'
    }
    return '#94a3b8'
  }

  return '#94a3b8'
}
