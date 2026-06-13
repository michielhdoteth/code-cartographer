import React, { useState } from 'react'
import '../styles/issuesPanel.css'

interface Issue {
  id: string
  type: 'warning' | 'error' | 'info'
  title: string
  count: number
  description: string
  details?: string[]
}

interface ArchitectureIssuesPanelProps {
  issues: Issue[]
}

export const ArchitectureIssuesPanel: React.FC<ArchitectureIssuesPanelProps> = ({ issues }) => {
  const [expandedId, setExpandedId] = useState<string | null>(issues.length > 0 ? issues[0].id : null)

  const getIconType = (type: string) => {
    switch (type) {
      case 'error':
        return '⚠️'
      case 'warning':
        return '⚡'
      case 'info':
        return 'ℹ️'
      default:
        return '○'
    }
  }

  const getIssueColor = (type: string) => {
    switch (type) {
      case 'error':
        return '#ef4444'
      case 'warning':
        return '#f59e0b'
      case 'info':
        return '#3b82f6'
      default:
        return '#6b7280'
    }
  }

  if (issues.length === 0) {
    return (
      <div className='issues-panel'>
        <div className='issues-header'>
          <h3>Architecture Issues</h3>
          <span className='issues-badge'>0</span>
        </div>
        <div className='issues-empty'>
          <p>No architectural issues detected!</p>
          <p className='issues-empty-subtitle'>Your codebase is looking clean.</p>
        </div>
      </div>
    )
  }

  return (
    <div className='issues-panel'>
      <div className='issues-header'>
        <h3>Architecture Issues</h3>
        <span className='issues-badge'>{issues.length}</span>
      </div>
      <div className='issues-list'>
        {issues.map((issue) => (
          <div
            key={issue.id}
            className={`issue-item ${expandedId === issue.id ? 'expanded' : ''}`}
            style={{ borderLeftColor: getIssueColor(issue.type) }}
          >
            <div
              className='issue-header'
              onClick={() => setExpandedId(expandedId === issue.id ? null : issue.id)}
            >
              <div className='issue-icon' style={{ color: getIssueColor(issue.type) }}>
                {getIconType(issue.type)}
              </div>
              <div className='issue-title-section'>
                <div className='issue-title'>{issue.title}</div>
                <div className='issue-count'>{issue.count} items</div>
              </div>
              <div className='issue-toggle'>{expandedId === issue.id ? '▼' : '▶'}</div>
            </div>
            {expandedId === issue.id && (
              <div className='issue-details'>
                <p className='issue-description'>{issue.description}</p>
                {issue.details && issue.details.length > 0 && (
                  <div className='issue-items-list'>
                    {issue.details.slice(0, 3).map((detail, idx) => (
                      <div key={idx} className='detail-item'>
                        {detail}
                      </div>
                    ))}
                    {issue.details.length > 3 && (
                      <div className='detail-more'>+{issue.details.length - 3} more</div>
                    )}
                  </div>
                )}
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
