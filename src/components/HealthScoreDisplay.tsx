import React from 'react'
import '../styles/healthScore.css'

interface HealthScoreDisplayProps {
  score: number
  maxScore?: number
}

export const HealthScoreDisplay: React.FC<HealthScoreDisplayProps> = ({ score, maxScore = 100 }) => {
  const percentage = (score / maxScore) * 100
  const radius = 45
  const circumference = 2 * Math.PI * radius
  const offset = circumference - (percentage / 100) * circumference

  const getColor = () => {
    if (score >= 80) return '#10b981'
    if (score >= 60) return '#f59e0b'
    if (score >= 40) return '#f97316'
    return '#ef4444'
  }

  const getGrade = () => {
    if (score >= 90) return 'A'
    if (score >= 80) return 'B'
    if (score >= 70) return 'C'
    if (score >= 60) return 'D'
    return 'F'
  }

  return (
    <div className='health-score-container'>
      <svg className='health-score-ring' viewBox='0 0 100 100'>
        <circle cx='50' cy='50' r={radius} className='health-score-bg' />
        <circle
          cx='50'
          cy='50'
          r={radius}
          className='health-score-ring-fill'
          style={{
            strokeDashoffset: offset,
            strokeDasharray: circumference,
            stroke: getColor(),
          }}
        />
      </svg>
      <div className='health-score-content'>
        <div className='health-score-number'>{score}</div>
        <div className='health-score-label'>Health Score</div>
        <div className='health-score-grade' style={{ color: getColor() }}>
          Grade: {getGrade()}
        </div>
      </div>
    </div>
  )
}
