import React, { useState, useCallback } from 'react'
import { CodeMap, CodeNode } from '@models/codeMap'

interface TraceStep {
  nodeId: string
  nodeName: string
  nodeType: string
  action: 'enter' | 'call' | 'return' | 'import'
  depth: number
}

interface TraceSimulatorProps {
  codeMap: CodeMap
  startNodeId?: string
  onStepSelect?: (nodeId: string) => void
  onHighlightPath?: (nodeIds: string[]) => void
}

export const TraceSimulator: React.FC<TraceSimulatorProps> = ({
  codeMap,
  startNodeId,
  onStepSelect,
  onHighlightPath,
}) => {
  const [trace, setTrace] = useState<TraceStep[]>([])
  const [currentStep, setCurrentStep] = useState(-1)
  const [isPlaying, setIsPlaying] = useState(false)
  const [selectedStartNode, setSelectedStartNode] = useState(startNodeId || '')

  const availableEntryPoints = codeMap
    .getNodes()
    .filter((n) => n.data.type === 'function' || n.data.type === 'method' || n.data.type === 'class')
    .slice(0, 20)

  const buildTrace = useCallback(
    (nodeId: string, visited: Set<string> = new Set(), depth: number = 0): TraceStep[] => {
      if (visited.has(nodeId) || depth > 10) return []

      const node = codeMap.getNode(nodeId)
      if (!node) return []

      visited.add(nodeId)
      const steps: TraceStep[] = []

      steps.push({
        nodeId: node.data.id,
        nodeName: node.data.name,
        nodeType: node.data.type,
        action: depth === 0 ? 'enter' : 'call',
        depth,
      })

      // Find outgoing calls/references
      const outgoingEdges = codeMap.getEdges().filter(
        (e) => e.data.source === nodeId && (e.data.type === 'calls' || e.data.type === 'references')
      )

      for (const edge of outgoingEdges.slice(0, 5)) {
        const childSteps = buildTrace(edge.data.target, visited, depth + 1)
        steps.push(...childSteps)
      }

      if (depth > 0) {
        steps.push({
          nodeId: node.data.id,
          nodeName: node.data.name,
          nodeType: node.data.type,
          action: 'return',
          depth,
        })
      }

      return steps
    },
    [codeMap]
  )

  const startSimulation = useCallback(() => {
    if (!selectedStartNode) return

    const newTrace = buildTrace(selectedStartNode)
    setTrace(newTrace)
    setCurrentStep(-1)
    setIsPlaying(false)

    // Highlight all nodes in the trace
    const pathNodes = [...new Set(newTrace.map((s) => s.nodeId))]
    onHighlightPath?.(pathNodes)
  }, [selectedStartNode, buildTrace, onHighlightPath])

  const stepForward = useCallback(() => {
    if (currentStep < trace.length - 1) {
      const nextStep = currentStep + 1
      setCurrentStep(nextStep)
      onStepSelect?.(trace[nextStep].nodeId)
    }
  }, [currentStep, trace, onStepSelect])

  const stepBackward = useCallback(() => {
    if (currentStep > 0) {
      const prevStep = currentStep - 1
      setCurrentStep(prevStep)
      onStepSelect?.(trace[prevStep].nodeId)
    }
  }, [currentStep, trace, onStepSelect])

  const playSimulation = useCallback(() => {
    setIsPlaying(true)
    let step = currentStep

    const interval = setInterval(() => {
      if (step < trace.length - 1) {
        step++
        setCurrentStep(step)
        onStepSelect?.(trace[step].nodeId)
      } else {
        setIsPlaying(false)
        clearInterval(interval)
      }
    }, 800)

    return () => clearInterval(interval)
  }, [currentStep, trace, onStepSelect])

  const resetSimulation = useCallback(() => {
    setCurrentStep(-1)
    setIsPlaying(false)
    setTrace([])
    onHighlightPath?.([])
  }, [onHighlightPath])

  const getActionIcon = (action: string) => {
    switch (action) {
      case 'enter':
        return '▶'
      case 'call':
        return '→'
      case 'return':
        return '←'
      case 'import':
        return '⬇'
      default:
        return '•'
    }
  }

  const getActionColor = (action: string) => {
    switch (action) {
      case 'enter':
        return '#10b981'
      case 'call':
        return '#3b82f6'
      case 'return':
        return '#8b5cf6'
      case 'import':
        return '#f59e0b'
      default:
        return '#6b7280'
    }
  }

  return (
    <div className="trace-simulator">
      <div className="trace-header">
        <h4>Flow Tracer</h4>
        <span className="trace-badge">{trace.length > 0 ? `${trace.length} steps` : 'Ready'}</span>
      </div>

      <div className="trace-controls">
        <select
          className="trace-select"
          value={selectedStartNode}
          onChange={(e) => setSelectedStartNode(e.target.value)}
        >
          <option value="">Select entry point...</option>
          {availableEntryPoints.map((node) => (
            <option key={node.data.id} value={node.data.id}>
              {node.data.name} ({node.data.type})
            </option>
          ))}
        </select>

        <div className="trace-buttons">
          <button
            className="trace-btn primary"
            onClick={startSimulation}
            disabled={!selectedStartNode}
          >
            Trace
          </button>
          <button
            className="trace-btn"
            onClick={stepBackward}
            disabled={currentStep <= 0 || isPlaying}
          >
            ◀
          </button>
          <button
            className="trace-btn"
            onClick={isPlaying ? () => setIsPlaying(false) : playSimulation}
            disabled={trace.length === 0}
          >
            {isPlaying ? '⏸' : '▶'}
          </button>
          <button
            className="trace-btn"
            onClick={stepForward}
            disabled={currentStep >= trace.length - 1 || isPlaying}
          >
            ▶
          </button>
          <button className="trace-btn" onClick={resetSimulation}>
            ↺
          </button>
        </div>
      </div>

      {trace.length > 0 && (
        <div className="trace-timeline">
          {trace.map((step, idx) => (
            <div
              key={idx}
              className={`trace-step ${idx === currentStep ? 'active' : ''} ${idx < currentStep ? 'visited' : ''}`}
              style={{ paddingLeft: `${step.depth * 12 + 8}px` }}
              onClick={() => {
                setCurrentStep(idx)
                onStepSelect?.(step.nodeId)
              }}
            >
              <span className="trace-icon" style={{ color: getActionColor(step.action) }}>
                {getActionIcon(step.action)}
              </span>
              <span className="trace-name">{step.nodeName}</span>
              <span className="trace-type">{step.nodeType}</span>
            </div>
          ))}
        </div>
      )}

      {trace.length === 0 && (
        <div className="trace-empty">
          <p>Select an entry point and click "Trace" to simulate code flow.</p>
        </div>
      )}
    </div>
  )
}
