import React, { useEffect, useRef } from 'react'
import * as d3 from 'd3'
import { CodeMap } from '@models/codeMap'
import { VisualizationConfig } from '@models/types'

interface TreeVisualizationProps {
  codeMap: CodeMap
  config: VisualizationConfig
  onNodeSelect: (nodeId: string) => void
}

export const TreeVisualization: React.FC<TreeVisualizationProps> = ({ codeMap, config, onNodeSelect }) => {
  const containerRef = useRef<HTMLDivElement>(null)
  const svgRef = useRef<SVGSVGElement>(null)

  useEffect(() => {
    if (!containerRef.current || !codeMap) return

    const width = containerRef.current.clientWidth
    const height = containerRef.current.clientHeight

    // Build hierarchical data
    const root = {
      id: 'root',
      name: 'Codebase',
      type: 'root',
      children: [] as any[],
    }

    // Group nodes by file
    const fileMap = new Map<string, any>()
    const files = codeMap.getFiles()
    const nodes = codeMap.getNodes()

    files.forEach((file) => {
      fileMap.set(file.data.id, {
        id: file.data.id,
        name: file.data.path.split('/').pop() || file.data.path,
        type: 'file',
        language: file.data.language,
        children: [] as any[],
      })
    })

    // Add nodes to their files
    nodes.forEach((node) => {
      if (node.data.file) {
        const fileId = codeMap.getFiles().find((f) => f.data.path === node.data.file)?.data.id
        if (fileId && fileMap.has(fileId)) {
          fileMap.get(fileId)!.children.push({
            id: node.data.id,
            name: node.data.name,
            type: node.data.type,
            metrics: node.data.metrics,
          })
        }
      }
    })

    root.children = Array.from(fileMap.values())

    // Clear previous
    if (svgRef.current) {
      svgRef.current.innerHTML = ''
    }

    // Create SVG
    const svg = d3
      .select(svgRef.current)
      .attr('width', width)
      .attr('height', height)

    try {
      // Create hierarchy
      const hierarchy = d3.hierarchy(root as any)
      const treeLayout = d3.tree<any>().size([width, height - 40])
      const treeData = treeLayout(hierarchy)

      // Create main group
      const g = svg.append('g').attr('class', 'tree-container').attr('transform', 'translate(0,20)')

      // Add links
      g.selectAll('line')
        .data(treeData.links())
        .enter()
        .append('line')
        .attr('x1', (d: any) => d.source.x)
        .attr('y1', (d: any) => d.source.y)
        .attr('x2', (d: any) => d.target.x)
        .attr('y2', (d: any) => d.target.y)
        .attr('stroke', '#999')
        .attr('stroke-opacity', 0.6)
        .attr('stroke-width', 1)

      // Add nodes
      const nodeGroups = g
        .selectAll('g.tree-node')
        .data(treeData.descendants())
        .enter()
        .append('g')
        .attr('class', 'tree-node')
        .attr('transform', (d: any) => `translate(${d.x},${d.y})`)

      // Add circles
      nodeGroups
        .append('circle')
        .attr('r', (d: any) => {
          if (d.data.type === 'root') return 6
          if (d.data.type === 'file') return 5
          return 3
        })
        .attr('fill', (d: any) => getTreeNodeColor(d, config.colorMode))
        .attr('stroke', '#000')
        .attr('stroke-width', 1.5)
        .style('cursor', 'pointer')
        .on('click', (event: any, d: any) => {
          event.stopPropagation()
          if (d.data.id !== 'root') {
            onNodeSelect(d.data.id)
          }
        })

      // Add labels if enabled
      if (config.showLabels) {
        nodeGroups
          .append('text')
          .text((d: any) => d.data.name)
          .attr('text-anchor', 'middle')
          .attr('dy', '-8px')
          .attr('font-size', (d: any) => (d.data.type === 'root' ? 11 : 9))
          .attr('font-weight', 'bold')
          .attr('fill', '#fff')
          .attr('pointer-events', 'none')
          .style('text-shadow', '0 0 3px rgba(0,0,0,0.8)')
      }

      // Add zoom
      const zoom = d3.zoom<SVGSVGElement, unknown>().on('zoom', (event: any) => {
        g.attr('transform', event.transform)
      })

      svg.call(zoom as any)
    } catch (error) {
      console.error('Error rendering tree:', error)
    }
  }, [codeMap, config, onNodeSelect])

  return (
    <div ref={containerRef} className='visualizer-container' style={{ width: '100%', height: '100%' }}>
      <svg ref={svgRef} className='graph'></svg>
    </div>
  )
}

function getTreeNodeColor(node: any, colorMode: string): string {
  const typeColors: Record<string, string> = {
    root: '#6366f1',
    file: '#3b82f6',
    class: '#8b5cf6',
    function: '#06b6d4',
    method: '#10b981',
  }

  if (colorMode === 'type') {
    return typeColors[node.data.type] || '#94a3b8'
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
    return langColors[node.data.language] || '#94a3b8'
  }

  if (colorMode === 'complexity') {
    if (node.data.metrics) {
      const complexity = node.data.metrics.cyclomatic
      if (complexity > 10) return '#ef4444'
      if (complexity > 5) return '#f59e0b'
      return '#10b981'
    }
    return '#94a3b8'
  }

  return '#94a3b8'
}
