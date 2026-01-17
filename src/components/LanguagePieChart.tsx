import React, { useMemo } from 'react'
import { CodeMap } from '@models/codeMap'

interface LanguagePieChartProps {
  codeMap: CodeMap | null
}

interface LanguageData {
  language: string
  lines: number
  color: string
}

const LANGUAGE_COLORS: Record<string, string> = {
  javascript: '#f1e05a',
  typescript: '#3178c6',
  python: '#3572a5',
  java: '#b07219',
  go: '#00add8',
  rust: '#ce422b',
  cpp: '#f34b7d',
  ruby: '#cc342d',
  php: '#777bb4',
}

export const LanguagePieChart: React.FC<LanguagePieChartProps> = ({ codeMap }) => {
  const languageStats = useMemo(() => {
    if (!codeMap) return []

    const stats: Record<string, number> = {}
    const nodes = codeMap.getNodes()
    const files = codeMap.getFiles()

    // Create a map of file IDs to languages
    const fileLanguages: Record<string, string> = {}
    files.forEach((file) => {
      const lang = (file.data.language || 'unknown').toLowerCase()
      fileLanguages[file.data.id] = lang
    })

    // Count lines by language from nodes
    nodes.forEach((node) => {
      let lang = (node.data.language || 'unknown').toLowerCase()

      // If node doesn't have language, try to infer from parent file
      if (!node.data.language && node.data.file) {
        const parentFile = files.find((f) => f.data.path === node.data.file)
        if (parentFile) {
          lang = (parentFile.data.language || 'unknown').toLowerCase()
        }
      }

      const lines = node.data.metrics?.linesOfCode || 1
      stats[lang] = (stats[lang] || 0) + lines
    })

    return Object.entries(stats)
      .map(([language, lines]) => ({
        language,
        lines,
        color: LANGUAGE_COLORS[language] || '#94a3b8',
      }))
      .sort((a, b) => b.lines - a.lines)
  }, [codeMap])

  const total = useMemo(() => {
    return languageStats.reduce((sum, stat) => sum + stat.lines, 0)
  }, [languageStats])

  if (!codeMap || languageStats.length === 0) {
    return <div style={{ fontSize: '0.8rem', color: 'var(--text-tertiary)' }}>No data</div>
  }

  // Calculate pie slices
  const slices = languageStats.map((stat) => {
    const percentage = (stat.lines / total) * 100
    return { ...stat, percentage }
  })

  // Create SVG path data for donut chart
  let currentAngle = 0
  const outerRadius = 40
  const innerRadius = 25
  const centerX = 50
  const centerY = 50

  const paths = slices.map((slice) => {
    const sliceAngle = (slice.percentage / 100) * 360
    const startAngle = currentAngle
    const endAngle = currentAngle + sliceAngle
    const midAngle = startAngle + sliceAngle / 2

    // Convert angles to radians
    const startRad = (startAngle - 90) * (Math.PI / 180)
    const endRad = (endAngle - 90) * (Math.PI / 180)
    const midRad = (midAngle - 90) * (Math.PI / 180)

    // Outer arc points
    const x1 = centerX + outerRadius * Math.cos(startRad)
    const y1 = centerY + outerRadius * Math.sin(startRad)
    const x2 = centerX + outerRadius * Math.cos(endRad)
    const y2 = centerY + outerRadius * Math.sin(endRad)

    // Inner arc points
    const x3 = centerX + innerRadius * Math.cos(endRad)
    const y3 = centerY + innerRadius * Math.sin(endRad)
    const x4 = centerX + innerRadius * Math.cos(startRad)
    const y4 = centerY + innerRadius * Math.sin(startRad)

    // Determine if we need the large arc flag
    const largeArc = sliceAngle > 180 ? 1 : 0

    // Create donut path
    const pathData = `M ${x1} ${y1} A ${outerRadius} ${outerRadius} 0 ${largeArc} 1 ${x2} ${y2} L ${x3} ${y3} A ${innerRadius} ${innerRadius} 0 ${largeArc} 0 ${x4} ${y4} Z`

    // Calculate label position (middle of the slice)
    const labelRadius = (outerRadius + innerRadius) / 2
    const labelX = centerX + labelRadius * Math.cos(midRad)
    const labelY = centerY + labelRadius * Math.sin(midRad)

    currentAngle = endAngle

    return { pathData, labelX, labelY, ...slice }
  })

  return (
    <div className='language-pie-chart'>
      <div className='pie-chart-container'>
        <svg viewBox='0 0 100 100' preserveAspectRatio='xMidYMid meet'>
          {paths.map((slice, idx) => (
            <g key={idx}>
              <path
                d={slice.pathData}
                fill={slice.color}
                stroke='var(--bg-secondary)'
                strokeWidth='0.5'
              />
              {slice.percentage >= 8 && (
                <text
                  x={slice.labelX}
                  y={slice.labelY}
                  textAnchor='middle'
                  dominantBaseline='middle'
                  fill='white'
                  fontSize='7'
                  fontWeight='bold'
                  pointerEvents='none'
                >
                  {slice.language.substring(0, 3).toUpperCase()}
                </text>
              )}
            </g>
          ))}
        </svg>
      </div>

      <div className='language-list'>
        {slices.map((slice) => (
          <div key={slice.language} className='language-item'>
            <span className='language-dot' style={{ backgroundColor: slice.color }}></span>
            <span className='language-name'>{slice.language}</span>
            <span className='language-percent'>{slice.percentage.toFixed(0)}%</span>
          </div>
        ))}
      </div>
    </div>
  )
}
