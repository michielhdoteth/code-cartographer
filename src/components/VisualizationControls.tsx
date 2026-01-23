import React, { useState, useCallback } from 'react'

interface VisualizationControlsProps {
  svgRef: React.RefObject<SVGSVGElement>
  onZoomIn: () => void
  onZoomOut: () => void
  onZoomFit: () => void
  currentLayout: string
}

const layoutDescriptions: Record<string, string> = {
  force: 'Physics-based layout where nodes repel each other and edges act as springs. Best for exploring relationships.',
  radial: 'Circular arrangement with central nodes at the core. Best for showing hierarchy depth.',
  hierarchical: 'Top-to-bottom tree structure. Best for visualizing module dependencies.',
  grid: 'Organized grid layout. Best for comparing nodes side by side.',
  metro: 'Transit map style with clean lines. Best for tracing paths between nodes.',
  treemap: 'Space-filling rectangles sized by code metrics. Best for seeing relative sizes.',
  matrix: 'Adjacency matrix showing all connections. Best for dense dependency analysis.',
}

export const VisualizationControls: React.FC<VisualizationControlsProps> = ({
  svgRef,
  onZoomIn,
  onZoomOut,
  onZoomFit,
  currentLayout,
}) => {
  const [exporting, setExporting] = useState(false)

  const exportToPNG = useCallback(async () => {
    if (!svgRef.current) return
    setExporting(true)

    try {
      const svg = svgRef.current
      const svgData = new XMLSerializer().serializeToString(svg)
      const svgBlob = new Blob([svgData], { type: 'image/svg+xml;charset=utf-8' })
      const svgUrl = URL.createObjectURL(svgBlob)

      const img = new Image()
      img.onload = () => {
        const canvas = document.createElement('canvas')
        const scale = 2 // High DPI export
        canvas.width = svg.clientWidth * scale
        canvas.height = svg.clientHeight * scale

        const ctx = canvas.getContext('2d')
        if (ctx) {
          ctx.scale(scale, scale)
          ctx.fillStyle = '#1f2937' // Dark background
          ctx.fillRect(0, 0, canvas.width, canvas.height)
          ctx.drawImage(img, 0, 0)

          canvas.toBlob((blob) => {
            if (blob) {
              const url = URL.createObjectURL(blob)
              const link = document.createElement('a')
              link.download = `code-cartographer-${currentLayout}-${Date.now()}.png`
              link.href = url
              link.click()
              URL.revokeObjectURL(url)
            }
            setExporting(false)
          }, 'image/png')
        }
        URL.revokeObjectURL(svgUrl)
      }
      img.src = svgUrl
    } catch (error) {
      console.error('Export failed:', error)
      setExporting(false)
    }
  }, [svgRef, currentLayout])

  const exportToSVG = useCallback(() => {
    if (!svgRef.current) return

    try {
      const svg = svgRef.current.cloneNode(true) as SVGSVGElement

      // Add background rect
      const rect = document.createElementNS('http://www.w3.org/2000/svg', 'rect')
      rect.setAttribute('width', '100%')
      rect.setAttribute('height', '100%')
      rect.setAttribute('fill', '#1f2937')
      svg.insertBefore(rect, svg.firstChild)

      const svgData = new XMLSerializer().serializeToString(svg)
      const svgBlob = new Blob([svgData], { type: 'image/svg+xml;charset=utf-8' })
      const url = URL.createObjectURL(svgBlob)

      const link = document.createElement('a')
      link.download = `code-cartographer-${currentLayout}-${Date.now()}.svg`
      link.href = url
      link.click()
      URL.revokeObjectURL(url)
    } catch (error) {
      console.error('SVG export failed:', error)
    }
  }, [svgRef, currentLayout])

  return (
    <div className="visualization-controls-bar">
      <div className="zoom-controls">
        <button
          className="control-btn"
          onClick={onZoomIn}
          title="Zoom In"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <circle cx="11" cy="11" r="8" />
            <path d="M21 21l-4.35-4.35M11 8v6M8 11h6" />
          </svg>
        </button>
        <button
          className="control-btn"
          onClick={onZoomOut}
          title="Zoom Out"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <circle cx="11" cy="11" r="8" />
            <path d="M21 21l-4.35-4.35M8 11h6" />
          </svg>
        </button>
        <button
          className="control-btn"
          onClick={onZoomFit}
          title="Fit to View"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M15 3h6v6M9 21H3v-6M21 3l-7 7M3 21l7-7" />
          </svg>
        </button>
      </div>

      <div className="export-controls">
        <button
          className="control-btn export-btn"
          onClick={exportToPNG}
          disabled={exporting}
          title="Export as PNG"
        >
          {exporting ? (
            <span className="export-spinner"></span>
          ) : (
            <>
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4M7 10l5 5 5-5M12 15V3" />
              </svg>
              PNG
            </>
          )}
        </button>
        <button
          className="control-btn export-btn"
          onClick={exportToSVG}
          title="Export as SVG"
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4M7 10l5 5 5-5M12 15V3" />
          </svg>
          SVG
        </button>
      </div>

      <div className="layout-description">
        {layoutDescriptions[currentLayout] || 'Select a layout to see description'}
      </div>
    </div>
  )
}
