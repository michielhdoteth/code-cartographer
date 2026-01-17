import React from 'react'
import '../styles/visualizationTabs.css'

interface VisualizationTab {
  id: 'graph' | 'tree' | 'treemap' | 'matrix' | 'flow'
  label: string
  icon: string
}

interface VisualizationTabsProps {
  activeTab: string
  onTabChange: (tab: string) => void
}

const tabs: VisualizationTab[] = [
  { id: 'graph', label: 'Graph', icon: '◉' },
  { id: 'tree', label: 'Tree', icon: '▲' },
  { id: 'treemap', label: 'Treemap', icon: '■' },
  { id: 'matrix', label: 'Matrix', icon: '▦' },
  { id: 'flow', label: 'Flow', icon: '→' },
]

export const VisualizationTabs: React.FC<VisualizationTabsProps> = ({ activeTab, onTabChange }) => {
  return (
    <div className='visualization-tabs'>
      {tabs.map((tab) => (
        <button
          key={tab.id}
          className={`viz-tab ${activeTab === tab.id ? 'active' : ''}`}
          onClick={() => onTabChange(tab.id)}
          title={tab.label}
        >
          <span className='viz-tab-icon'>{tab.icon}</span>
          <span className='viz-tab-label'>{tab.label}</span>
        </button>
      ))}
    </div>
  )
}
