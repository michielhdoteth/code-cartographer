import React, { useState, useCallback, useMemo } from 'react'

interface SearchNode {
  id: string
  name: string
  type?: string
  file?: string
}

export interface SearchHighlightProps {
  nodes: SearchNode[]
  onSelect?: (nodeId: string) => void
  onFilterChange?: (nodeIds: string[]) => void
}

type MatchType = 'name' | 'type' | 'file'

interface HighlightRange {
  start: number
  end: number
}

export interface SearchResult {
  nodeId: string
  nodeName: string
  matchType: MatchType
  highlights: HighlightRange[]
}

const MAX_VISIBLE_RESULTS = 10

type SearchMode = 'name' | 'advanced'

/**
 * Search and highlight component for graph visualization.
 * Supports finding nodes by name, type, and file.
 * Provides keyboard navigation and filtering.
 */
export function SearchHighlight({
  nodes,
  onSelect,
  onFilterChange,
}: SearchHighlightProps): React.ReactElement {
  const [searchQuery, setSearchQuery] = useState('')
  const [searchResults, setSearchResults] = useState<SearchResult[]>([])
  const [selectedResultIndex, setSelectedResultIndex] = useState(0)
  const [searchMode, setSearchMode] = useState<SearchMode>('name')

  const performSearch = useCallback(
    (query: string) => {
      if (!query.trim()) {
        setSearchResults([])
        onFilterChange?.(nodes.map((n) => n.id))
        return
      }

      const results = searchNodes(nodes, query)
      setSearchResults(results)
      setSelectedResultIndex(0)
      onFilterChange?.(results.map((r) => r.nodeId))

      if (results.length > 0) {
        onSelect?.(results[0].nodeId)
      }
    },
    [nodes, onFilterChange, onSelect]
  )

  const handleSearchChange = useCallback(
    (event: React.ChangeEvent<HTMLInputElement>) => {
      const query = event.target.value
      setSearchQuery(query)
      performSearch(query)
    },
    [performSearch]
  )

  const handleSelectResult = useCallback(
    (index: number) => {
      setSelectedResultIndex(index)
      onSelect?.(searchResults[index].nodeId)
    },
    [searchResults, onSelect]
  )

  const handleKeyDown = useCallback(
    (event: React.KeyboardEvent<HTMLInputElement>) => {
      if (searchResults.length === 0) return

      if (event.key === 'ArrowDown') {
        event.preventDefault()
        handleSelectResult(Math.min(selectedResultIndex + 1, searchResults.length - 1))
      } else if (event.key === 'ArrowUp') {
        event.preventDefault()
        handleSelectResult(Math.max(selectedResultIndex - 1, 0))
      } else if (event.key === 'Enter') {
        event.preventDefault()
        onSelect?.(searchResults[selectedResultIndex].nodeId)
      }
    },
    [searchResults, selectedResultIndex, handleSelectResult, onSelect]
  )

  const resultsSummary = useMemo(() => {
    if (searchResults.length === 0) return null
    return `${selectedResultIndex + 1} of ${searchResults.length} results`
  }, [searchResults.length, selectedResultIndex])

  const visibleResults = searchResults.slice(0, MAX_VISIBLE_RESULTS)
  const remainingCount = searchResults.length - MAX_VISIBLE_RESULTS

  return (
    <div style={containerStyle}>
      <div style={searchSectionStyle}>
        <div style={inputContainerStyle}>
          <input
            type="text"
            placeholder="Search nodes by name, type, or file..."
            value={searchQuery}
            onChange={handleSearchChange}
            onKeyDown={handleKeyDown}
            style={inputStyle}
          />
        </div>

        <div style={modeContainerStyle}>
          <ModeButton
            label="Simple"
            isActive={searchMode === 'name'}
            onClick={() => setSearchMode('name')}
          />
          <ModeButton
            label="Advanced"
            isActive={searchMode === 'advanced'}
            onClick={() => setSearchMode('advanced')}
          />
        </div>

        {searchQuery && <div style={summaryStyle}>{resultsSummary}</div>}
      </div>

      {searchResults.length > 0 && (
        <div style={resultsContainerStyle}>
          {visibleResults.map((result, index) => (
            <ResultItem
              key={result.nodeId}
              result={result}
              isSelected={index === selectedResultIndex}
              onClick={() => handleSelectResult(index)}
            />
          ))}

          {remainingCount > 0 && (
            <div style={moreResultsStyle}>+{remainingCount} more results</div>
          )}
        </div>
      )}

      {searchQuery && searchResults.length === 0 && (
        <div style={emptyStateStyle}>No matches found</div>
      )}

      {!searchQuery && (
        <div style={helpStyle}>
          <div style={helpTitleStyle}>
            <strong>Keyboard shortcuts:</strong>
          </div>
          <div>Up/Down Navigate | Enter Select</div>
        </div>
      )}
    </div>
  )
}

// Sub-components

interface ModeButtonProps {
  label: string
  isActive: boolean
  onClick: () => void
}

function ModeButton({ label, isActive, onClick }: ModeButtonProps): React.ReactElement {
  const style: React.CSSProperties = {
    flex: 1,
    padding: '6px 8px',
    fontSize: '12px',
    border: `1px solid ${isActive ? '#2196F3' : '#ddd'}`,
    borderRadius: '4px',
    backgroundColor: isActive ? '#e3f2fd' : '#f9f9f9',
    cursor: 'pointer',
  }
  return (
    <button onClick={onClick} style={style}>
      {label}
    </button>
  )
}

interface ResultItemProps {
  result: SearchResult
  isSelected: boolean
  onClick: () => void
}

function ResultItem({ result, isSelected, onClick }: ResultItemProps): React.ReactElement {
  const [isHovered, setIsHovered] = useState(false)

  const backgroundColor = isSelected ? '#f0f7ff' : isHovered ? '#f9f9f9' : 'transparent'

  const style: React.CSSProperties = {
    padding: '8px 12px',
    borderLeft: isSelected ? '3px solid #2196F3' : '3px solid transparent',
    backgroundColor,
    cursor: 'pointer',
    fontSize: '13px',
    borderBottom: '1px solid #f0f0f0',
    transition: 'background-color 0.2s',
  }

  return (
    <div
      onClick={onClick}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
      style={style}
    >
      <div style={resultNameStyle}>{highlightText(result.nodeName, result.highlights)}</div>
      <div style={resultMetaStyle}>{formatMatchType(result.matchType)}</div>
    </div>
  )
}

// Helper functions

function searchNodes(nodes: SearchNode[], query: string): SearchResult[] {
  const lowerQuery = query.toLowerCase()
  const results: SearchResult[] = []

  for (const node of nodes) {
    const nameMatch = node.name.toLowerCase().includes(lowerQuery)
    const typeMatch = node.type?.toLowerCase().includes(lowerQuery)
    const fileMatch = node.file?.toLowerCase().includes(lowerQuery)

    if (nameMatch) {
      results.push({
        nodeId: node.id,
        nodeName: node.name,
        matchType: 'name',
        highlights: findHighlights(node.name, query),
      })
    } else if (typeMatch) {
      results.push({
        nodeId: node.id,
        nodeName: node.name,
        matchType: 'type',
        highlights: findHighlights(node.type ?? '', query),
      })
    } else if (fileMatch) {
      results.push({
        nodeId: node.id,
        nodeName: node.name,
        matchType: 'file',
        highlights: findHighlights(node.file ?? '', query),
      })
    }
  }

  return results
}

function findHighlights(text: string, query: string): HighlightRange[] {
  const highlights: HighlightRange[] = []
  const lowerText = text.toLowerCase()
  const lowerQuery = query.toLowerCase()

  let startPos = 0
  let pos = lowerText.indexOf(lowerQuery, startPos)

  while (pos !== -1) {
    highlights.push({ start: pos, end: pos + query.length })
    startPos = pos + 1
    pos = lowerText.indexOf(lowerQuery, startPos)
  }

  return highlights
}

function highlightText(text: string, highlights: HighlightRange[]): React.ReactNode {
  if (highlights.length === 0) return text

  const segments: React.ReactNode[] = []
  let lastEnd = 0

  for (let i = 0; i < highlights.length; i++) {
    const highlight = highlights[i]

    if (highlight.start > lastEnd) {
      segments.push(<span key={`text-${i}`}>{text.substring(lastEnd, highlight.start)}</span>)
    }

    segments.push(
      <span key={`highlight-${i}`} style={highlightStyle}>
        {text.substring(highlight.start, highlight.end)}
      </span>
    )

    lastEnd = highlight.end
  }

  if (lastEnd < text.length) {
    segments.push(<span key="text-end">{text.substring(lastEnd)}</span>)
  }

  return <>{segments}</>
}

function formatMatchType(matchType: MatchType): string {
  if (matchType === 'name') return 'Match: Name'
  if (matchType === 'type') return 'Match: Type'
  return 'Match: File'
}

// Styles

const containerStyle: React.CSSProperties = {
  position: 'absolute',
  top: 10,
  left: 10,
  width: 350,
  backgroundColor: '#ffffff',
  borderRadius: '8px',
  boxShadow: '0 2px 8px rgba(0, 0, 0, 0.15)',
  zIndex: 1000,
}

const searchSectionStyle: React.CSSProperties = {
  padding: '12px',
}

const inputContainerStyle: React.CSSProperties = {
  marginBottom: '8px',
}

const inputStyle: React.CSSProperties = {
  width: '100%',
  padding: '8px 12px',
  fontSize: '14px',
  border: '1px solid #ddd',
  borderRadius: '4px',
  boxSizing: 'border-box',
  fontFamily: 'monospace',
}

const modeContainerStyle: React.CSSProperties = {
  display: 'flex',
  gap: '8px',
  marginBottom: '8px',
}

const summaryStyle: React.CSSProperties = {
  fontSize: '12px',
  color: '#666',
  marginBottom: '8px',
}

const resultsContainerStyle: React.CSSProperties = {
  maxHeight: '300px',
  overflowY: 'auto',
  borderTop: '1px solid #eee',
}

const resultNameStyle: React.CSSProperties = {
  fontFamily: 'monospace',
  fontWeight: 500,
  marginBottom: '2px',
}

const resultMetaStyle: React.CSSProperties = {
  fontSize: '12px',
  color: '#999',
}

const highlightStyle: React.CSSProperties = {
  backgroundColor: '#FFEB3B',
  fontWeight: 'bold',
}

const moreResultsStyle: React.CSSProperties = {
  padding: '8px 12px',
  fontSize: '12px',
  color: '#999',
  textAlign: 'center',
  borderTop: '1px solid #eee',
}

const emptyStateStyle: React.CSSProperties = {
  padding: '12px',
  fontSize: '12px',
  color: '#999',
  textAlign: 'center',
  borderTop: '1px solid #eee',
}

const helpStyle: React.CSSProperties = {
  padding: '12px',
  fontSize: '12px',
  color: '#999',
  borderTop: '1px solid #eee',
}

const helpTitleStyle: React.CSSProperties = {
  marginBottom: '4px',
}
