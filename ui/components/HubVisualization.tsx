import React, { useMemo } from 'react';
import { HubAnalyzer, HubFile, HubAnalysisResult } from '../analyzers/hubAnalyzer';
import { CodeNodeData as CodeNode, CodeEdgeData as CodeEdge } from '../models/types';
import { AlertTriangle, Target, Zap, Shield } from 'lucide-react';

interface HubVisualizationProps {
  nodes: CodeNode[];
  edges: CodeEdge[];
  onHubSelect?: (hub: HubFile) => void;
  className?: string;
}

export const HubVisualization: React.FC<HubVisualizationProps> = ({
  nodes,
  edges,
  onHubSelect,
  className = ''
}) => {
  const hubAnalyzer = useMemo(() => new HubAnalyzer(), []);
  
  const hubAnalysis = useMemo(() => {
    return hubAnalyzer.analyzeHubFiles(nodes, edges);
  }, [nodes, edges, hubAnalyzer]);

  const getImpactIcon = (impact: HubFile['impact']) => {
    switch (impact) {
      case 'critical':
        return <AlertTriangle className="w-4 h-4 text-red-500" />;
      case 'high':
        return <Zap className="w-4 h-4 text-orange-500" />;
      case 'medium':
        return <Target className="w-4 h-4 text-yellow-500" />;
      case 'low':
        return <Shield className="w-4 h-4 text-green-500" />;
    }
  };

  const getImpactColor = (impact: HubFile['impact']) => {
    switch (impact) {
      case 'critical':
        return 'border-red-200 bg-red-50 hover:bg-red-100';
      case 'high':
        return 'border-orange-200 bg-orange-50 hover:bg-orange-100';
      case 'medium':
        return 'border-yellow-200 bg-yellow-50 hover:bg-yellow-100';
      case 'low':
        return 'border-green-200 bg-green-50 hover:bg-green-100';
    }
  };

  const getImpactBadgeColor = (impact: HubFile['impact']) => {
    switch (impact) {
      case 'critical':
        return 'bg-red-500 text-white';
      case 'high':
        return 'bg-orange-500 text-white';
      case 'medium':
        return 'bg-yellow-500 text-white';
      case 'low':
        return 'bg-green-500 text-white';
    }
  };

  if (hubAnalysis.hubs.length === 0) {
    return (
      <div className={`text-center py-8 text-gray-500 ${className}`}>
        <Shield className="w-12 h-12 mx-auto mb-4 text-gray-300" />
        <p className="text-lg font-medium mb-2">No Hub Files Detected</p>
        <p className="text-sm">
          Hub files are imported by 3+ other files. Your project has no high-impact files.
        </p>
      </div>
    );
  }

  return (
    <div className={`space-y-6 ${className}`}>
      {/* Summary Stats */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        {Object.entries(hubAnalysis.summary).map(([impact, count]) => (
          <div key={impact} className="text-center p-4 bg-gray-50 rounded-lg">
            {getImpactIcon(impact as HubFile['impact'])}
            <div className="text-2xl font-bold mt-2">{count}</div>
            <div className="text-sm text-gray-600 capitalize">{impact}</div>
          </div>
        ))}
      </div>

      {/* Analysis Info */}
      <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="font-semibold text-blue-900">Hub Analysis</h3>
            <p className="text-sm text-blue-700">
              Found {hubAnalysis.hubs.length} hub files out of {hubAnalysis.totalFiles} total files 
              (threshold: {hubAnalysis.hubThreshold}+ importers)
            </p>
          </div>
          <div className="text-sm text-blue-600">
            Analysis time: {hubAnalysis.analysisTime.toFixed(2)}ms
          </div>
        </div>
      </div>

      {/* Hub Files List */}
      <div className="space-y-3">
        <h3 className="text-lg font-semibold text-gray-900">Hub Files</h3>
        <div className="space-y-2">
          {hubAnalysis.hubs.map((hub, index) => (
            <div
              key={hub.id}
              className={`
                border rounded-lg p-4 cursor-pointer transition-all duration-200
                ${getImpactColor(hub.impact)}
              `}
              onClick={() => onHubSelect?.(hub)}
            >
              <div className="flex items-start justify-between">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center space-x-2 mb-2">
                    {getImpactIcon(hub.impact)}
                    <span className="font-semibold text-gray-900 truncate">
                      {hub.name}
                    </span>
                    <span className={`px-2 py-1 rounded-full text-xs font-medium ${getImpactBadgeColor(hub.impact)}`}>
                      {hub.impact}
                    </span>
                  </div>
                  
                  <div className="text-sm text-gray-600 space-y-1">
                    <div className="flex items-center space-x-4">
                      <span>Language: <span className="font-medium">{hub.language}</span></span>
                      <span>Score: <span className="font-medium">{hub.score}</span></span>
                      <span>Blast Radius: <span className="font-medium">{hub.blastRadius}</span></span>
                    </div>
                    
                    <div className="text-xs">
                      <div>Imported by {hub.importers.length} files</div>
                      {hub.imports.length > 0 && (
                        <div>Imports {hub.imports.length} files</div>
                      )}
                    </div>
                    
                    {hub.path !== hub.id && (
                      <div className="text-xs text-gray-500 truncate">
                        Path: {hub.path}
                      </div>
                    )}
                  </div>
                </div>
                
                <div className="text-2xl font-bold text-gray-400 ml-4">
                  #{index + 1}
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Impact Legend */}
      <div className="border-t pt-4">
        <h4 className="text-sm font-medium text-gray-700 mb-3">Impact Levels</h4>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3 text-xs">
          <div className="flex items-center space-x-2">
            <AlertTriangle className="w-4 h-4 text-red-500" />
            <span>Critical: 20+ importers</span>
          </div>
          <div className="flex items-center space-x-2">
            <Zap className="w-4 h-4 text-orange-500" />
            <span>High: 10-19 importers</span>
          </div>
          <div className="flex items-center space-x-2">
            <Target className="w-4 h-4 text-yellow-500" />
            <span>Medium: 5-9 importers</span>
          </div>
          <div className="flex items-center space-x-2">
            <Shield className="w-4 h-4 text-green-500" />
            <span>Low: 3-4 importers</span>
          </div>
        </div>
      </div>
    </div>
  );
};

export default HubVisualization;