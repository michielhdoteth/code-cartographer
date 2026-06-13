import React, { useEffect, useRef } from 'react'
import * as d3 from 'd3'
import { CodeMap, CodeNode } from '@models/codeMap'
import { VisualizationConfig, BlastRadiusResult } from '@models/types'

interface ForceDirectedGraphProps {
  codeMap: CodeMap
  config: VisualizationConfig
  onNodeSelect: (nodeId: string) => void
}

export const ForceDirectedGraph: React.FC<ForceDirectedGraphProps> = ({ codeMap, config, onNodeSelect }) => {
  const containerRef = useRef<HTMLDivElement>(null)
  const svgRef = useRef<SVGSVGElement>(null)
  const selectedNodeRef = useRef<string | null>(null)
  const blastRadiusRef = useRef<any | null>(null)

  useEffect(() => {
    if (!containerRef.current || !codeMap) return

    const width = containerRef.current.clientWidth
    const height = containerRef.current.clientHeight

    // Prepare data
    const allNodes = codeMap.getNodes()
    const files = codeMap.getFiles()

    // Create combined node list (code nodes + file nodes)
    const nodes = [
      ...allNodes.map((n) => ({
        ...n.data,
      })),
      ...files.map((f) => ({
        ...f.data,
        name: f.data.path.split('/').pop() || f.data.path,
        type: 'file' as const,
      })),
    ]

    // Create valid edges (only include edges where both nodes exist)
    const nodeIds = new Set(nodes.map(n => n.id))
    const edges = codeMap.getEdges()
      .filter(e => nodeIds.has(e.data.source) && nodeIds.has(e.data.target))
      .map((e) => ({
        source: e.data.source,
        target: e.data.target,
        type: e.data.type,
      }))

    // Clear previous
    if (svgRef.current) {
      svgRef.current.innerHTML = ''
    }

    // Add defs for gradients and markers
    const defs = d3.select(svgRef.current).append('defs')

    // Modern arrow marker for edges
    defs
      .append('marker')
      .attr('id', 'arrowhead')
      .attr('markerWidth', 8)
      .attr('markerHeight', 8)
      .attr('refX', 6)
      .attr('refY', 4)
      .attr('orient', 'auto')
      .append('polygon')
      .attr('points', '0 0, 8 4, 0 8')
      .attr('fill', '#6366f1')
      .attr('opacity', 0.7)

    // Glow filter for selected nodes
    const glow = defs.append('filter').attr('id', 'glow')
    glow.append('feGaussianBlur').attr('stdDeviation', '3').attr('result', 'coloredBlur')
    const feMerge = glow.append('feMerge')
    feMerge.append('feMergeNode').attr('in', 'coloredBlur')
    feMerge.append('feMergeNode').attr('in', 'SourceGraphic')

    // Create SVG
    const svg = d3
      .select(svgRef.current)
      .attr('width', width)
      .attr('height', height)

    // Create force simulation
    try {
      const simulation = d3
        .forceSimulation(nodes as any)
        .force(
          'link',
          d3.forceLink<any, any>(edges as any)
            .id((d: any) => d.id)
            .distance(config.linkDistance || 120)
            .strength(0.4)
        )
        .force('charge', d3.forceManyBody().strength(config.chargeStrength || -800).distanceMax(500))
        .force('center', d3.forceCenter(width / 2, height / 2))
        .force('collide', d3.forceCollide().radius((d: any) => {
          const baseSize = 8
          if (d.metrics) {
            if (config.nodeSize === 'lines') {
              return Math.sqrt(d.metrics.linesOfCode) / 8 + baseSize + 10
            } else if (config.nodeSize === 'complexity') {
              return d.metrics.cyclomatic + baseSize + 10
            }
          }
          return baseSize + 10
        }))
        .alphaDecay(0.02)

      // Create groups
      const g = svg.append('g').attr('class', 'graph-container')

      // Add edges with modern curved paths
      const links = g
        .selectAll('path.link')
        .data(edges)
        .enter()
        .append('path')
        .attr('class', 'link')
        .attr('fill', 'none')
        .attr('stroke', (d: any) => {
          if (d.type === 'calls' || d.type === 'imports') return '#6366f1'
          if (d.type === 'inherits' || d.type === 'implements') return '#8b5cf6'
          return '#94a3b8'
        })
        .attr('stroke-opacity', 0.4)
        .attr('stroke-width', 1.5)
        .attr('stroke-dasharray', (d: any) => d.type === 'imports' ? '3,3' : 'none')
        .attr('marker-end', 'url(#arrowhead)')

      // Add nodes
      const nodeGroups = g
        .selectAll('g.node')
        .data(nodes)
        .enter()
        .append('g')
        .attr('class', 'node')
        .call(d3.drag<any, any>().on('start', dragStarted).on('drag', dragged).on('end', dragEnded))

      // Add circles with modern styling
      const circles = nodeGroups
        .append('circle')
        .attr('r', (d: any) => {
          const baseSize = 8
          if (d.metrics) {
            if (config.nodeSize === 'lines') {
              return Math.sqrt(d.metrics.linesOfCode) / 8 + baseSize
            } else if (config.nodeSize === 'complexity') {
              return d.metrics.cyclomatic + baseSize
            }
          }
          return baseSize
        })
        .attr('fill', (d: any) =>
          getNodeColor(d, config.colorMode, selectedNodeRef.current || undefined, blastRadiusRef.current || undefined)
        )
        .attr('fill-opacity', (d: any) => {
          // Dim non-affected nodes when blast radius is active
          if (config.colorMode === 'blast' && blastRadiusRef.current) {
            const affected = new Set([
              selectedNodeRef.current,
              ...(blastRadiusRef.current.affected || []),
              ...(blastRadiusRef.current.transitive || [])
            ])
            return affected.has(d.id) ? 1.0 : 0.15
          }
          return 1.0
        })
        .attr('stroke', '#fff')
        .attr('stroke-width', 2)
        .style('cursor', 'pointer')
        .style('transition', 'all 0.2s ease')
        .on('click', (event: any, d: any) => {
          event.stopPropagation()
          selectedNodeRef.current = d.id

          // Calculate blast radius for selected node
          if (config.colorMode === 'blast') {
            const blastResult = codeMap.calculateBlastRadiusDetailed(d.id)
            blastRadiusRef.current = blastResult

            // Update circle colors with transition
            circles
              .transition()
              .duration(300)
              .attr('fill', (nodeData: any) =>
                getNodeColor(nodeData, config.colorMode, selectedNodeRef.current || undefined, blastRadiusRef.current || undefined)
              )
              .attr('fill-opacity', (nodeData: any) => {
                const affected = new Set([
                  selectedNodeRef.current,
                  ...(blastRadiusRef.current?.affected || []),
                  ...(blastRadiusRef.current?.transitive || [])
                ])
                return affected.has(nodeData.id) ? 1.0 : 0.2
              })
          }

          onNodeSelect(d.id)
        })
        .on('mouseover', function (event: any, d: any) {
          const isAffected = config.colorMode === 'blast' && blastRadiusRef.current
            ? new Set([
                selectedNodeRef.current,
                ...(blastRadiusRef.current?.affected || []),
                ...(blastRadiusRef.current?.transitive || [])
              ]).has(d.id)
            : true

          if (!isAffected && config.colorMode === 'blast') return // Don't hover dimmed nodes

          d3.select(this)
            .transition()
            .duration(200)
            .attr('stroke-width', 4)
            .attr('r', (d: any) => {
              const baseSize = 6
              if (d.metrics) {
                if (config.nodeSize === 'lines') {
                  return Math.sqrt(d.metrics.linesOfCode) / 8 + baseSize + 3
                } else if (config.nodeSize === 'complexity') {
                  return d.metrics.cyclomatic + baseSize + 3
                }
              }
              return baseSize + 3
            })
        })
        .on('mouseout', function (event: any, d: any) {
          d3.select(this)
            .transition()
            .duration(200)
            .attr('stroke-width', 2.5)
            .attr('r', (d: any) => {
              const baseSize = 6
              if (d.metrics) {
                if (config.nodeSize === 'lines') {
                  return Math.sqrt(d.metrics.linesOfCode) / 8 + baseSize
                } else if (config.nodeSize === 'complexity') {
                  return d.metrics.cyclomatic + baseSize
                }
              }
              return baseSize
            })
        })

      // Add labels if enabled
      if (config.showLabels) {
        nodeGroups
          .append('text')
          .text((d: any) => d.name.length > 12 ? d.name.substring(0, 10) + '..' : d.name)
          .attr('text-anchor', 'middle')
          .attr('dy', '0.35em')
          .attr('font-size', 11)
          .attr('font-weight', '600')
          .attr('fill', '#fff')
          .attr('pointer-events', 'none')
          .style('text-shadow', '0 0 4px rgba(0,0,0,0.95), 0 0 8px rgba(0,0,0,0.6)')
          .style('font-family', 'JetBrains Mono, monospace')
      }

      // Update positions
      simulation.on('tick', () => {
        links.attr('d', (d: any) => {
          const dx = d.target.x - d.source.x
          const dy = d.target.y - d.source.y
          const dr = Math.sqrt(dx * dx + dy * dy) * 1.5
          return `M${d.source.x},${d.source.y}A${dr},${dr} 0 0,1 ${d.target.x},${d.target.y}`
        })

        nodeGroups.attr('transform', (d: any) => `translate(${d.x},${d.y})`)
      })

      // Zoom
      const zoom = d3.zoom<SVGSVGElement, unknown>().on('zoom', (event: any) => {
        g.attr('transform', event.transform)
      })

      svg.call(zoom as any)

      // Drag functions
      function dragStarted(event: any, d: any) {
        if (!event.active) simulation.alphaTarget(0.3).restart()
        d.fx = d.x
        d.fy = d.y
      }

      function dragged(event: any, d: any) {
        d.fx = event.x
        d.fy = event.y
      }

      function dragEnded(event: any, d: any) {
        if (!event.active) simulation.alphaTarget(0)
        d.fx = null
        d.fy = null
      }

      return () => {
        simulation.stop()
      }
    } catch (error) {
      console.error('Error rendering graph:', error)
      return () => {}
    }
  }, [codeMap, config, onNodeSelect])

  return (
    <div ref={containerRef} className='visualizer-container' style={{ width: '100%', height: '100%' }}>
      <svg ref={svgRef} className='graph'></svg>
    </div>
  )
}

function getNodeColor(
  node: any,
  colorMode: string,
  selectedNodeId?: string,
  blastRadius?: any,
  layerMap?: Map<string, string>,
  churnMap?: Map<string, number>
): string {
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

  // Blast radius highlighting (RED for selected, ORANGE for affected, YELLOW for transitive)
  if (colorMode === 'blast' && selectedNodeId && blastRadius) {
    if (node.id === selectedNodeId) return '#ff5f5f' // RED - selected
    if (blastRadius.affected?.includes(node.id)) return '#ff9f43' // ORANGE - direct
    if (blastRadius.transitive?.includes(node.id)) return '#fbbf24' // YELLOW - transitive
    return typeColors[node.type] || '#64748b' // Default color
  }

  // Layer-based coloring
  if (colorMode === 'layer') {
    const layerColors: Record<string, string> = {
      ui: '#4d9fff',
      components: '#22d3ee',
      services: '#a78bfa',
      data: '#ff9f43',
      utils: '#00ff9d',
      config: '#ec4899',
      other: '#64748b',
    }
    const layer = layerMap?.get(node.id) || node.layer || 'other'
    return layerColors[layer] || '#64748b'
  }

  // Churn-based coloring (git activity)
  if (colorMode === 'churn') {
    const churn = churnMap?.get(node.id) || node.churn?.commits || 0
    if (churn > 20) return '#dc2626' // RED - high churn
    if (churn > 10) return '#f97316' // ORANGE
    if (churn > 5) return '#eab308' // YELLOW
    if (churn > 0) return '#22c55e' // GREEN
    return '#64748b' // GRAY - no changes
  }

  // Folder-based coloring (using path hash)
  if (colorMode === 'folder') {
    const folderColors = ['#4d9fff', '#22d3ee', '#a78bfa', '#ff9f43', '#00ff9d', '#ec4899', '#84cc16', '#f43f5e']
    const folder = node.path?.split('/')[0] || 'root'
    let hash = 0
    for (let i = 0; i < folder.length; i++) {
      hash = ((hash << 5) - hash) + folder.charCodeAt(i)
      hash = hash & hash // Convert to 32bit integer
    }
    return folderColors[Math.abs(hash) % folderColors.length]
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
