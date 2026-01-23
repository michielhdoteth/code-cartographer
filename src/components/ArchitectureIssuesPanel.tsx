import React, { useState, useCallback } from 'react'
import '../styles/issuesPanel.css'

interface Issue {
  id: string
  type: 'warning' | 'error' | 'info' | 'success'
  title: string
  count: number
  description: string
  details?: string[]
  severity?: 'critical' | 'high' | 'medium' | 'low'
}

interface VerificationCheck {
  name: string
  passed: boolean
  message?: string
}

interface ArchitectureIssuesPanelProps {
  issues: Issue[]
  verificationChecks?: VerificationCheck[]
  onExportReport?: () => void
}

export const ArchitectureIssuesPanel: React.FC<ArchitectureIssuesPanelProps> = ({
  issues,
  verificationChecks,
  onExportReport,
}) => {
  const [expandedId, setExpandedId] = useState<string | null>(issues.length > 0 ? issues[0].id : null)
  const [activeTab, setActiveTab] = useState<'issues' | 'verification'>('issues')

  const getIconType = (type: string) => {
    switch (type) {
      case 'error':
        return '⚠️'
      case 'warning':
        return '⚡'
      case 'info':
        return 'ℹ️'
      case 'success':
        return '✓'
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
      case 'success':
        return '#10b981'
      default:
        return '#6b7280'
    }
  }

  const getSeverityBadge = (severity?: string) => {
    const colors: Record<string, { bg: string; text: string }> = {
      critical: { bg: '#7f1d1d', text: '#fecaca' },
      high: { bg: '#991b1b', text: '#fecaca' },
      medium: { bg: '#92400e', text: '#fef3c7' },
      low: { bg: '#1e3a5f', text: '#bfdbfe' },
    }
    const style = colors[severity || 'low'] || colors.low
    return severity ? (
      <span
        className="severity-badge"
        style={{ backgroundColor: style.bg, color: style.text }}
      >
        {severity.toUpperCase()}
      </span>
    ) : null
  }

  const errorCount = issues.filter((i) => i.type === 'error').length
  const warningCount = issues.filter((i) => i.type === 'warning').length
  const infoCount = issues.filter((i) => i.type === 'info').length

  const defaultChecks: VerificationCheck[] = verificationChecks || [
    { name: 'No Circular Dependencies', passed: errorCount === 0 || !issues.some((i) => i.id === 'circular') },
    { name: 'No Dead Code', passed: !issues.some((i) => i.id === 'unused') },
    { name: 'Complexity Under Threshold', passed: !issues.some((i) => i.id === 'complexity') },
    { name: 'Security Patterns', passed: !issues.some((i) => i.type === 'error' && i.id.includes('security')) },
  ]

  const exportReport = useCallback(() => {
    const report = {
      generated: new Date().toISOString(),
      summary: {
        totalIssues: issues.length,
        errors: errorCount,
        warnings: warningCount,
        info: infoCount,
      },
      verificationChecks: defaultChecks,
      issues: issues.map((i) => ({
        type: i.type,
        title: i.title,
        count: i.count,
        description: i.description,
        details: i.details,
        severity: i.severity,
      })),
    }

    const blob = new Blob([JSON.stringify(report, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.download = `code-cartographer-report-${Date.now()}.json`
    link.href = url
    link.click()
    URL.revokeObjectURL(url)

    onExportReport?.()
  }, [issues, errorCount, warningCount, infoCount, defaultChecks, onExportReport])

  return (
    <div className="issues-panel">
      <div className="issues-header">
        <h3>Analysis Results</h3>
        <div className="issues-header-badges">
          {errorCount > 0 && (
            <span className="count-badge error-badge">{errorCount}</span>
          )}
          {warningCount > 0 && (
            <span className="count-badge warning-badge">{warningCount}</span>
          )}
          {infoCount > 0 && (
            <span className="count-badge info-badge">{infoCount}</span>
          )}
        </div>
      </div>

      <div className="panel-tabs">
        <button
          className={`panel-tab ${activeTab === 'issues' ? 'active' : ''}`}
          onClick={() => setActiveTab('issues')}
        >
          Issues
        </button>
        <button
          className={`panel-tab ${activeTab === 'verification' ? 'active' : ''}`}
          onClick={() => setActiveTab('verification')}
        >
          Checks
        </button>
      </div>

      {activeTab === 'verification' && (
        <div className="verification-panel">
          <div className="verification-list">
            {defaultChecks.map((check, idx) => (
              <div key={idx} className={`verification-item ${check.passed ? 'passed' : 'failed'}`}>
                <span className={`verification-badge ${check.passed ? 'pass' : 'fail'}`}>
                  {check.passed ? 'PASS' : 'FAIL'}
                </span>
                <span className="verification-name">{check.name}</span>
              </div>
            ))}
          </div>
        </div>
      )}

      {activeTab === 'issues' && (
        <>
          {issues.length === 0 ? (
            <div className="issues-empty">
              <div className="success-icon">✓</div>
              <p>No architectural issues detected!</p>
              <p className="issues-empty-subtitle">Your codebase is looking clean.</p>
            </div>
          ) : (
            <div className="issues-list">
              {issues.map((issue) => (
                <div
                  key={issue.id}
                  className={`issue-item ${expandedId === issue.id ? 'expanded' : ''}`}
                  style={{ borderLeftColor: getIssueColor(issue.type) }}
                >
                  <div
                    className="issue-header"
                    onClick={() => setExpandedId(expandedId === issue.id ? null : issue.id)}
                  >
                    <div className="issue-icon" style={{ color: getIssueColor(issue.type) }}>
                      {getIconType(issue.type)}
                    </div>
                    <div className="issue-title-section">
                      <div className="issue-title-row">
                        <div className="issue-title">{issue.title}</div>
                        {getSeverityBadge(issue.severity)}
                      </div>
                      <div className="issue-count">{issue.count} items</div>
                    </div>
                    <div className="issue-toggle">{expandedId === issue.id ? '▼' : '▶'}</div>
                  </div>
                  {expandedId === issue.id && (
                    <div className="issue-details">
                      <p className="issue-description">{issue.description}</p>
                      {issue.details && issue.details.length > 0 && (
                        <div className="issue-items-list">
                          {issue.details.slice(0, 5).map((detail, idx) => (
                            <div key={idx} className="detail-item">
                              <span className="detail-bullet">•</span>
                              {detail}
                            </div>
                          ))}
                          {issue.details.length > 5 && (
                            <div className="detail-more">+{issue.details.length - 5} more</div>
                          )}
                        </div>
                      )}
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </>
      )}

      <div className="panel-footer">
        <button className="export-report-btn" onClick={exportReport}>
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4M7 10l5 5 5-5M12 15V3" />
          </svg>
          Export Report
        </button>
      </div>
    </div>
  )
}
