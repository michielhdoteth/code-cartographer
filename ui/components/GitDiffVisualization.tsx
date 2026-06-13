import React, { useState, useMemo } from 'react';
import { GitIntegration, GitDiff, GitFileDiff } from '../utils/gitIntegration';
import { CodeNodeData as CodeNode, CodeEdgeData as CodeEdge } from '../models/types';

interface GitDiffVisualizationProps {
  diffText: string;
  branch?: string;
  nodes?: CodeNode[];
  edges?: CodeEdge[];
  onFileSelect?: (file: GitFileDiff) => void;
  className?: string;
}

interface FileDiffViewProps {
  fileDiff: GitFileDiff;
  content?: string;
}

const FileDiffView: React.FC<FileDiffViewProps> = ({ fileDiff, content }) => {
  const [expanded, setExpanded] = useState(false);

  const getTypeIcon = (type: GitFileDiff['type']) => {
    switch (type) {
      case 'added':
        return <span className="text-green-600">➕</span>;
      case 'deleted':
        return <span className="text-red-600">➖</span>;
      case 'modified':
        return <span className="text-yellow-600">✏️</span>;
      case 'renamed':
        return <span className="text-blue-600">🔄</span>;
    }
  };

  const getTypeColor = (type: GitFileDiff['type']) => {
    switch (type) {
      case 'added':
        return 'border-green-200 bg-green-50';
      case 'deleted':
        return 'border-red-200 bg-red-50';
      case 'modified':
        return 'border-yellow-200 bg-yellow-50';
      case 'renamed':
        return 'border-blue-200 bg-blue-50';
    }
  };

  return (
    <div className={`border rounded-lg p-4 ${getTypeColor(fileDiff.type)}`}>
      <div className="flex items-center justify-between cursor-pointer" onClick={() => setExpanded(!expanded)}>
        <div className="flex items-center space-x-3">
          {getTypeIcon(fileDiff.type)}
          <div>
            <div className="font-medium">{fileDiff.path}</div>
            {fileDiff.oldPath && (
              <div className="text-sm text-gray-600">from: {fileDiff.oldPath}</div>
            )}
          </div>
        </div>
        
        <div className="flex items-center space-x-4 text-sm">
          {fileDiff.additions > 0 && (
            <span className="text-green-600">+{fileDiff.additions}</span>
          )}
          {fileDiff.deletions > 0 && (
            <span className="text-red-600">-{fileDiff.deletions}</span>
          )}
          <span className="text-gray-500">
            {expanded ? '▼' : '▶'}
          </span>
        </div>
      </div>

      {expanded && content && (
        <div className="mt-4 border-t pt-4">
          <div className="bg-gray-900 text-gray-100 p-4 rounded-lg overflow-x-auto text-sm font-mono">
            <pre className="whitespace-pre-wrap">{content}</pre>
          </div>
        </div>
      )}
    </div>
  );
};

export const GitDiffVisualization: React.FC<GitDiffVisualizationProps> = ({
  diffText,
  branch = 'main',
  nodes,
  edges,
  onFileSelect,
  className = ''
}) => {
  const gitIntegration = useMemo(() => new GitIntegration(), []);
  const [selectedFile, setSelectedFile] = useState<GitFileDiff | null>(null);

  const diff = useMemo(() => {
    return gitIntegration.parseDiff(diffText, branch);
  }, [diffText, branch, gitIntegration]);

  const impactAnalysis = useMemo(() => {
    if (!nodes || !edges) return null;
    return gitIntegration.calculateChangeImpact(nodes, edges, diff);
  }, [nodes, edges, diff, gitIntegration]);

  const summary = useMemo(() => {
    return gitIntegration.generateChangeSummary(diff);
  }, [diff, gitIntegration]);

  const filesByType = useMemo(() => {
    const grouped = {
      added: diff.files.filter(f => f.type === 'added'),
      modified: diff.files.filter(f => f.type === 'modified'),
      deleted: diff.files.filter(f => f.type === 'deleted'),
      renamed: diff.files.filter(f => f.type === 'renamed')
    };
    return grouped;
  }, [diff.files]);

  const handleFileClick = (file: GitFileDiff) => {
    setSelectedFile(file);
    onFileSelect?.(file);
  };

  return (
    <div className={`space-y-6 ${className}`}>
      {/* Summary Header */}
      <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
        <h3 className="font-semibold text-blue-900 mb-2">Git Diff Summary</h3>
        <p className="text-blue-700 mb-3">{summary}</p>
        
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
          <div>
            <div className="font-medium">Files</div>
            <div className="text-2xl font-bold">{diff.stats.files}</div>
          </div>
          <div>
            <div className="font-medium text-green-600">Additions</div>
            <div className="text-2xl font-bold text-green-600">+{diff.stats.additions}</div>
          </div>
          <div>
            <div className="font-medium text-red-600">Deletions</div>
            <div className="text-2xl font-bold text-red-600">-{diff.stats.deletions}</div>
          </div>
          <div>
            <div className="font-medium">Total Changes</div>
            <div className="text-2xl font-bold">{diff.stats.changes}</div>
          </div>
        </div>
      </div>

      {/* Impact Analysis */}
      {impactAnalysis && (
        <div className="border rounded-lg p-4">
          <h3 className="font-semibold mb-3">Change Impact Analysis</h3>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="text-center p-3 bg-red-50 rounded-lg">
              <div className="text-2xl font-bold text-red-600">{impactAnalysis.highImpact.length}</div>
              <div className="text-sm text-red-700">High Impact</div>
            </div>
            <div className="text-center p-3 bg-yellow-50 rounded-lg">
              <div className="text-2xl font-bold text-yellow-600">{impactAnalysis.mediumImpact.length}</div>
              <div className="text-sm text-yellow-700">Medium Impact</div>
            </div>
            <div className="text-center p-3 bg-green-50 rounded-lg">
              <div className="text-2xl font-bold text-green-600">{impactAnalysis.lowImpact.length}</div>
              <div className="text-sm text-green-700">Low Impact</div>
            </div>
          </div>
        </div>
      )}

      {/* File Type Breakdown */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {Object.entries(filesByType).map(([type, files]) => (
          files.length > 0 && (
            <div key={type} className="border rounded-lg p-4">
              <div className="font-medium capitalize mb-2">
                {type} ({files.length})
              </div>
              <div className="space-y-1 text-sm">
                {files.slice(0, 3).map((file, index) => (
                  <div key={index} className="truncate text-gray-600">
                    {file.path.split('/').pop()}
                  </div>
                ))}
                {files.length > 3 && (
                  <div className="text-gray-400">
                    ... and {files.length - 3} more
                  </div>
                )}
              </div>
            </div>
          )
        ))}
      </div>

      {/* File List */}
      <div className="space-y-3">
        <h3 className="font-semibold">Changed Files</h3>
        <div className="space-y-2">
          {diff.files.map((file, index) => (
            <FileDiffView
              key={index}
              fileDiff={file}
              content={file.content}
            />
          ))}
        </div>
      </div>

      {/* Selected File Details */}
      {selectedFile && (
        <div className="border-t pt-4">
          <div className="flex items-center justify-between mb-3">
            <h3 className="font-semibold">File Details: {selectedFile.path}</h3>
            <button
              onClick={() => setSelectedFile(null)}
              className="text-gray-500 hover:text-gray-700"
            >
              ✕
            </button>
          </div>
          
          <div className="bg-gray-50 p-4 rounded-lg">
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
              <div>
                <div className="font-medium">Type</div>
                <div className="capitalize">{selectedFile.type}</div>
              </div>
              <div>
                <div className="font-medium">Additions</div>
                <div className="text-green-600">+{selectedFile.additions}</div>
              </div>
              <div>
                <div className="font-medium">Deletions</div>
                <div className="text-red-600">-{selectedFile.deletions}</div>
              </div>
              <div>
                <div className="font-medium">Net Change</div>
                <div className={selectedFile.additions - selectedFile.deletions >= 0 ? 'text-green-600' : 'text-red-600'}>
                  {(selectedFile.additions - selectedFile.deletions >= 0 ? '+' : '')}{selectedFile.additions - selectedFile.deletions}
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default GitDiffVisualization;