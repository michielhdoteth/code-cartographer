import React, { useState } from 'react'
import { CodeMap, CodeNode } from '@models/codeMap'
import { ChevronDown, ChevronRight, File, Folder, FolderOpen, Code2, FunctionSquare, Search, X } from 'lucide-react'

interface FileExplorerProps {
  codeMap: CodeMap | null
  selectedNodeId: string | null
  onNodeSelect: (nodeId: string) => void
}

interface TreeNode {
  id: string
  name: string
  type: 'folder' | 'file' | 'class' | 'function' | 'method'
  children: TreeNode[]
  isExpanded: boolean
  nodeData?: any
}

export const FileExplorer: React.FC<FileExplorerProps> = ({ codeMap, selectedNodeId, onNodeSelect }) => {
  // Initialize with all folders and files expanded
  const [expandedFolders, setExpandedFolders] = useState<Set<string>>(() => {
    if (!codeMap) return new Set()

    const expanded = new Set<string>()

    // Add all file IDs
    codeMap.getFiles().forEach((f) => {
      expanded.add(f.data.id)
    })

    // Add all folder IDs (folder:path format)
    const files = codeMap.getFiles()
    const folderPaths = new Set<string>()

    files.forEach((file) => {
      const pathParts = file.data.path.split('/')
      let currentPath = ''
      for (let i = 0; i < pathParts.length - 1; i++) {
        currentPath = (currentPath ? currentPath + '/' : '') + pathParts[i]
        folderPaths.add(currentPath)
      }
    })

    folderPaths.forEach((path) => {
      expanded.add(`folder:${path}`)
    })

    return expanded
  })
  const [filterText, setFilterText] = useState('')

  if (!codeMap) {
    return (
      <div className='file-explorer'>
        <h3>Code Structure</h3>
        <p style={{ color: 'var(--text-tertiary)', fontSize: '0.85rem' }}>No code map loaded</p>
      </div>
    )
  }

  const buildTree = (): TreeNode[] => {
    const files = codeMap.getFiles()
    const nodes = codeMap.getNodes()
    const filter = filterText.toLowerCase()

    // Create a file map for quick lookup
    const fileMap = new Map<string, any>()
    files.forEach((file) => {
      fileMap.set(file.data.path, file)
    })

    // Build folder and file hierarchy
    const folderTree: Map<string, TreeNode> = new Map()
    const fileNodes: Map<string, TreeNode> = new Map()

    // Helper to get or create folder node
    const getOrCreateFolder = (folderPath: string): TreeNode => {
      if (folderTree.has(folderPath)) {
        return folderTree.get(folderPath)!
      }

      const folderName = folderPath.split('/').pop() || folderPath
      const folderId = `folder:${folderPath}`
      const folderNode: TreeNode = {
        id: folderId,
        name: folderName,
        type: 'folder',
        children: [],
        isExpanded: expandedFolders.has(folderId),
      }

      folderTree.set(folderPath, folderNode)

      // Add to parent folder
      const parentPath = folderPath.substring(0, folderPath.lastIndexOf('/'))
      if (parentPath && parentPath !== folderPath) {
        const parentFolder = getOrCreateFolder(parentPath)
        if (!parentFolder.children.find((c) => c.id === folderId)) {
          parentFolder.children.push(folderNode)
        }
      }

      return folderNode
    }

    // Create file and code nodes
    files.forEach((file) => {
      const fileName = file.data.path.split('/').pop() || file.data.path
      const filePath = file.data.path
      const folderPath = filePath.substring(0, filePath.lastIndexOf('/'))

      const fileNode: TreeNode = {
        id: file.data.id,
        name: fileName,
        type: 'file',
        children: [],
        isExpanded: expandedFolders.has(file.data.id),
        nodeData: file.data,
      }

      fileNodes.set(filePath, fileNode)

      // Add to parent folder
      if (folderPath) {
        const parentFolder = getOrCreateFolder(folderPath)
        parentFolder.children.push(fileNode)
      }
    })

    // Add code nodes to files
    nodes.forEach((node) => {
      if (['class', 'function', 'method'].includes(node.data.type)) {
        const filePath = node.data.file
        const fileNode = fileNodes.get(filePath)

        if (fileNode) {
          fileNode.children.push({
            id: node.data.id,
            name: node.data.name,
            type: node.data.type as 'class' | 'function' | 'method',
            children: [],
            isExpanded: false,
            nodeData: node.data,
          })
        }
      }
    })

    // Sort all children
    const sortNode = (node: TreeNode) => {
      node.children.sort((a, b) => {
        // Folders first, then files, then code elements
        const aIsFolderOrFile = a.type === 'folder' || a.type === 'file'
        const bIsFolderOrFile = b.type === 'folder' || b.type === 'file'

        if (aIsFolderOrFile !== bIsFolderOrFile) {
          return aIsFolderOrFile ? -1 : 1
        }

        // Within same type, sort by name
        return a.name.localeCompare(b.name)
      })
      node.children.forEach(sortNode)
    }

    // Get root folders
    let rootNodes: TreeNode[] = []
    folderTree.forEach((node) => {
      const pathParts = node.name.split('/')
      const isRoot = !Array.from(folderTree.values()).some((f) => f.id !== node.id && node.id.includes(f.id))

      // Check if this folder should be at root level
      const parentPath = node.id.replace('folder:', '').substring(0, node.id.replace('folder:', '').lastIndexOf('/'))
      if (!parentPath || !folderTree.has(parentPath)) {
        rootNodes.push(node)
      }
    })

    // Add root-level files (if any)
    fileNodes.forEach((fileNode) => {
      if (!fileNode.nodeData || !fileNode.nodeData.path.includes('/')) {
        rootNodes.push(fileNode)
      }
    })

    rootNodes.forEach(sortNode)

    // Apply filter
    if (!filter) {
      return rootNodes
    }

    const filterNode = (node: TreeNode): TreeNode | null => {
      const nameMatches = node.name.toLowerCase().includes(filter)
      const childrenAfterFilter = node.children
        .map(filterNode)
        .filter((c) => c !== null) as TreeNode[]

      if (nameMatches || childrenAfterFilter.length > 0) {
        return {
          ...node,
          children: childrenAfterFilter,
          isExpanded: childrenAfterFilter.length > 0 || nameMatches,
        }
      }

      return null
    }

    return rootNodes
      .map(filterNode)
      .filter((n) => n !== null) as TreeNode[]
  }

  const toggleExpand = (id: string) => {
    const newExpanded = new Set(expandedFolders)
    if (newExpanded.has(id)) {
      newExpanded.delete(id)
    } else {
      newExpanded.add(id)
    }
    setExpandedFolders(newExpanded)
  }

  const getTypeIcon = (type: string, isExpanded: boolean = false) => {
    switch (type) {
      case 'file':
        return <File size={16} className='text-blue-400' />
      case 'class':
        return <Code2 size={16} className='text-yellow-400' />
      case 'function':
        return <FunctionSquare size={16} className='text-green-400' />
      case 'method':
        return <FunctionSquare size={16} className='text-green-400' />
      case 'folder':
        return isExpanded ? (
          <FolderOpen size={16} className='text-orange-400' />
        ) : (
          <Folder size={16} className='text-orange-400' />
        )
      default:
        return <File size={16} className='text-gray-400' />
    }
  }

  const renderNode = (node: TreeNode, depth: number = 0) => {
    const isExpanded = expandedFolders.has(node.id)
    const isSelected = selectedNodeId === node.id

    return (
      <div key={node.id}>
        <div
          className={`explorer-item ${isSelected ? 'selected' : ''}`}
          style={{ paddingLeft: `${depth * 0.75}rem` }}
          onClick={(e) => {
            e.stopPropagation()
            if (node.children.length > 0) {
              toggleExpand(node.id)
            }
            onNodeSelect(node.id)
          }}
        >
          {node.children.length > 0 && (
            <button
              className='expand-toggle hover:bg-gray-700 p-1 rounded'
              onClick={(e) => {
                e.stopPropagation()
                toggleExpand(node.id)
              }}
              aria-label={isExpanded ? 'Collapse' : 'Expand'}
            >
              {isExpanded ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
            </button>
          )}
          {!node.children.length && <span className='expand-toggle w-6' />}

          <span className='node-icon'>{getTypeIcon(node.type, isExpanded)}</span>
          <span className='node-name'>{node.name}</span>

          {node.children.length > 0 && (
            <span className='node-count'>{node.children.length}</span>
          )}

          {node.nodeData?.metrics && (
            <span className='node-metrics'>
              {node.nodeData.metrics.linesOfCode > 0 && (
                <span className='metric'>{node.nodeData.metrics.linesOfCode}L</span>
              )}
              {node.nodeData.metrics.cyclomatic > 0 && (
                <span className='metric'>C{node.nodeData.metrics.cyclomatic}</span>
              )}
            </span>
          )}
        </div>

        {isExpanded && node.children.length > 0 && (
          <div className='explorer-children'>
            {node.children.map((child) => renderNode(child, depth + 1))}
          </div>
        )}
      </div>
    )
  }

  const tree = buildTree()

  return (
    <div className='file-explorer'>
      <h3>Code Structure</h3>
      <div style={{ padding: '0.75rem 0.5rem', borderBottom: '1px solid var(--border)' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', position: 'relative' }}>
          <Search size={14} style={{ color: 'var(--text-tertiary)', position: 'absolute', left: '0.6rem', pointerEvents: 'none' }} />
          <input
            type='text'
            placeholder='Filter code structure...'
            value={filterText}
            onChange={(e) => setFilterText(e.target.value)}
            style={{
              width: '100%',
              padding: '0.5rem 0.75rem 0.5rem 2rem',
              backgroundColor: 'var(--bg)',
              border: '1px solid var(--border)',
              borderRadius: '4px',
              fontSize: '0.8rem',
              color: 'var(--text-primary)',
              outline: 'none',
            }}
            onFocus={(e) => (e.currentTarget.style.borderColor = 'var(--primary)')}
            onBlur={(e) => (e.currentTarget.style.borderColor = 'var(--border)')}
          />
          {filterText && (
            <button
              onClick={() => setFilterText('')}
              style={{
                position: 'absolute',
                right: '0.6rem',
                background: 'none',
                border: 'none',
                color: 'var(--text-tertiary)',
                cursor: 'pointer',
                padding: '0.25rem',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
              }}
              aria-label='Clear filter'
            >
              <X size={14} />
            </button>
          )}
        </div>
        {filterText && (
          <div style={{ fontSize: '0.7rem', color: 'var(--text-tertiary)', marginTop: '0.4rem', paddingLeft: '0.5rem' }}>
            Filtered by: <strong>{filterText}</strong>
          </div>
        )}
      </div>
      <div className='explorer-tree'>{tree.map((node) => renderNode(node))}</div>
    </div>
  )
}
