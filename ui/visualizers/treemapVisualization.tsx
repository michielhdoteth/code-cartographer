import React, { useEffect, useRef } from 'react'
import * as d3 from 'd3'
import { CodeMap } from '@models/codeMap'
import { VisualizationConfig } from '@models/types'

interface TreemapVisualizationProps {
  codeMap: CodeMap
  config: VisualizationConfig
  onNodeSelect: (nodeId: string) => void
}

export const TreemapVisualization: React.FC<TreemapVisualizationProps> = ({ codeMap, config, onNodeSelect }) => {
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
      value: 0,
      children: [] as any[],
    }

    const fileMap = new Map<string, any>()
    const files = codeMap.getFiles()
    const nodes = codeMap.getNodes()

    files.forEach((file) => {
      const fileNode = {
        id: file.data.id,
        name: file.data.path.split('/').pop() || file.data.path,
        type: 'file',
        language: file.data.language,
        value: file.data.lines || 1,
        children: [] as any[],
      }
      fileMap.set(file.data.id, fileNode)
    })

    nodes.forEach((node) => {
      if (node.data.file) {
        const fileId = codeMap.getFiles().find((f) => f.data.path === node.data.file)?.data.id
        if (fileId && fileMap.has(fileId)) {
          fileMap.get(fileId)!.children.push({
            id: node.data.id,
            name: node.data.name,
            type: node.data.type,
            value: node.data.metrics?.linesOfCode || 1,
            metrics: node.data.metrics,
          })
        }
      }
    })

    root.children = Array.from(fileMap.values())
    root.value = root.children.reduce((sum, f) => sum + f.value, 0)

    // Clear previous
    if (svgRef.current) {
      svgRef.current.innerHTML = ''
    }

    const svg = d3
      .select(svgRef.current)
      .attr('width', width)
      .attr('height', height)

    try {
      // Create hierarchy
      const hierarchy = d3
        .hierarchy(root as any)
        .sum((d) => d.value)
        .sort((a, b) => (b.value || 0) - (a.value || 0))

      const treemap = d3.treemap<any>().size([width, height]).paddingTop(0).paddingRight(2).paddingBottom(2).paddingLeft(2)

      const treemapRoot = treemap(hierarchy)

      // Create main group
      const g = svg.append('g')

      // Add rectangles
      g.selectAll('rect')
        .data(treemapRoot.leaves())
        .enter()
        .append('rect')
        .attr('x', (d: any) => d.x0)
        .attr('y', (d: any) => d.y0)
        .attr('width', (d: any) => d.x1 - d.x0)
        .attr('height', (d: any) => d.y1 - d.y0)
        .attr('fill', (d: any) => getTreemapNodeColor(d, config.colorMode))
        .attr('stroke', '#fff')
        .attr('stroke-width', 2)
        .style('cursor', 'pointer')
        .on('click', (event: any, d: any) => {
          event.stopPropagation()
          onNodeSelect(d.data.id)
        })

      // Add labels if enabled
      if (config.showLabels) {
        g.selectAll('text')
          .data(treemapRoot.leaves())
          .enter()
          .append('text')
          .attr('x', (d: any) => d.x0 + 4)
          .attr('y', (d: any) => d.y0 + 20)
          .text((d: any) => d.data.name)
          .attr('font-size', 11)
          .attr('font-weight', 'bold')
          .attr('fill', '#fff')
          .attr('pointer-events', 'none')
          .style('text-shadow', '0 0 2px rgba(0,0,0,0.8)')
          .each(function (d: any) {
            const width = d.x1 - d.x0
            const height = d.y1 - d.y0
            if (width < 50 || height < 30) {
              d3.select(this).style('opacity', 0)
            }
          })
      }
    } catch (error) {
      console.error('Error rendering treemap:', error)
    }
  }, [codeMap, config, onNodeSelect])

  return (
    <div ref={containerRef} className='visualizer-container' style={{ width: '100%', height: '100%' }}>
      <svg ref={svgRef} className='graph'></svg>
    </div>
  )
}

function getTreemapNodeColor(node: any, colorMode: string): string {
  const typeColors: Record<string, string> = {
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
