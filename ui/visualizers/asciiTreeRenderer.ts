/**
 * ASCII Tree Renderer
 * 
 * Provides tree visualization capabilities for CLI output,
 * inspired by JordanCoin's codemap tree rendering.
 */

import { CodeNodeData as CodeNode, CodeEdgeData as CodeEdge, Language } from '../models/types';

export interface ASCIITreeNode {
  id: string;
  name: string;
  type: string;
  language?: Language;
  children: ASCIITreeNode[];
  depth: number;
  isLast: boolean;
  path: string;
  size?: number;
  metrics?: any;
  isHub?: boolean;
  hubScore?: number;
}

export interface ASCIITreeConfig {
  showIcons: boolean;
  showLanguage: boolean;
  showSize: boolean;
  showMetrics: boolean;
  showHubs: boolean;
  colorOutput: boolean;
  maxDepth: number;
  indentSize: number;
  icons: Record<string, string>;
  colors: Record<string, string>;
}

export interface ASCIITreeResult {
  tree: string;
  stats: {
    totalNodes: number;
    maxDepth: number;
    hubCount: number;
    languageDistribution: Record<string, number>;
  };
}

export class ASCIITreeRenderer {
  private readonly DEFAULT_CONFIG: ASCIITreeConfig = {
    showIcons: true,
    showLanguage: true,
    showSize: false,
    showMetrics: false,
    showHubs: true,
    colorOutput: true,
    maxDepth: 0, // 0 = unlimited
    indentSize: 2,
    icons: {
      folder: '📁',
      file: '📄',
      class: '🏗️',
      function: '⚡',
      method: '🔧',
      interface: '📋',
      module: '📦',
      hub: '⭐',
      default: '📄'
    },
    colors: {
      reset: '\x1b[0m',
      bright: '\x1b[1m',
      dim: '\x1b[2m',
      red: '\x1b[31m',
      green: '\x1b[32m',
      yellow: '\x1b[33m',
      blue: '\x1b[34m',
      magenta: '\x1b[35m',
      cyan: '\x1b[36m',
      white: '\x1b[37m',
      gray: '\x1b[90m'
    }
  };

  private readonly LANGUAGE_COLORS: Record<Language, string> = {
    javascript: 'yellow',
    typescript: 'blue',
    python: 'green',
    java: 'red',
    go: 'cyan',
    rust: 'orange',
    cpp: 'magenta',
    ruby: 'red',
    php: 'purple',
    jsx: 'blue',
    tsx: 'blue',
    c: 'white',
    csharp: 'green',
    swift: 'orange',
    kotlin: 'purple',
    scala: 'red',
    html: 'red',
    css: 'blue',
    sql: 'cyan',
    shell: 'green',
    yaml: 'yellow',
    json: 'white',
    toml: 'cyan',
    markdown: 'white',
  };

  /**
   * Render nodes and edges as ASCII tree
   */
  renderTree(
    nodes: CodeNode[],
    edges: CodeEdge[],
    config: Partial<ASCIITreeConfig> = {}
  ): ASCIITreeResult {
    const finalConfig = { ...this.DEFAULT_CONFIG, ...config };
    
    // Build tree structure
    const treeData = this.buildTreeStructure(nodes, edges);
    
    // Render tree
    const tree = this.renderTreeNode(treeData, finalConfig);
    
    // Calculate statistics
    const stats = this.calculateTreeStats(treeData);
    
    return { tree, stats };
  }

  /**
   * Build hierarchical tree structure from flat nodes
   */
  private buildTreeStructure(nodes: CodeNode[], edges: CodeEdge[]): ASCIITreeNode {
    // Create a map of nodes by ID
    const nodeMap = new Map<string, ASCIITreeNode>();
    
    // Convert CodeNodes to ASCIITreeNodes
    nodes.forEach(node => {
      const treeNode: ASCIITreeNode = {
        id: node.id,
        name: node.name,
        type: node.type,
        language: node.language,
        children: [],
        depth: 0,
        isLast: false,
        path: node.file || node.id,
        size: node.metrics?.linesOfCode,
        metrics: node.metrics,
        isHub: false,
        hubScore: 0
      };
      nodeMap.set(node.id, treeNode);
    });

    // Build hierarchy using 'contains' relationships
    const rootNodes: ASCIITreeNode[] = [];
    
    nodes.forEach(node => {
      const treeNode = nodeMap.get(node.id);
      if (!treeNode) return;

      // Find parent using 'contains' edge or file path hierarchy
      const containsEdge = edges.find(edge => edge.type === 'contains' && edge.target === node.id);
      
      if (containsEdge) {
        const parent = nodeMap.get(containsEdge.source);
        if (parent) {
          parent.children.push(treeNode);
        }
      } else {
        // Use file path as fallback
        const parent = this.findParentByPath(node, nodeMap, edges);
        if (parent) {
          parent.children.push(treeNode);
        } else {
          rootNodes.push(treeNode);
        }
      }
    });

    // Calculate hub scores
    this.calculateHubScores(rootNodes, edges);

    // Create artificial root if multiple roots
    if (rootNodes.length === 1) {
      return rootNodes[0];
    } else {
      return {
        id: 'root',
        name: 'root',
        type: 'folder',
        children: rootNodes,
        depth: 0,
        isLast: false,
        path: '',
        isHub: false,
        hubScore: 0
      };
    }
  }

  /**
   * Find parent by file path hierarchy
   */
  private findParentByPath(
    node: CodeNode,
    nodeMap: Map<string, ASCIITreeNode>,
    edges: CodeEdge[]
  ): ASCIITreeNode | null {
    const nodePath = node.file || node.id;
    
    // Find potential parents by path prefix
    let bestParent: ASCIITreeNode | null = null;
    let bestDepth = 0;

    for (const [id, candidate] of nodeMap.entries()) {
      if (id === node.id) continue;
      
      const candidatePath = candidate.path;
      if (candidatePath && nodePath.startsWith(candidatePath + '/') || nodePath.startsWith(candidatePath + '\\')) {
        const depth = candidatePath.split('/').length;
        if (depth > bestDepth) {
          bestDepth = depth;
          bestParent = candidate;
        }
      }
    }

    return bestParent;
  }

  /**
   * Calculate hub scores for all nodes
   */
  private calculateHubScores(nodes: ASCIITreeNode[], edges: CodeEdge[]) {
    // Count importers for each node
    const importCounts = new Map<string, number>();
    
    edges
      .filter(edge => edge.type === 'imports')
      .forEach(edge => {
        const count = importCounts.get(edge.target) || 0;
        importCounts.set(edge.target, count + 1);
      });

    // Apply hub scores to nodes
    const applyScores = (nodeList: ASCIITreeNode[]) => {
      nodeList.forEach(node => {
        const score = importCounts.get(node.id) || 0;
        node.hubScore = score;
        node.isHub = score >= 3; // Hub threshold
        applyScores(node.children);
      });
    };

    applyScores(nodes);
  }

  /**
   * Render tree node recursively
   */
  private renderTreeNode(node: ASCIITreeNode, config: ASCIITreeConfig, prefix = ''): string {
    if (config.maxDepth > 0 && node.depth >= config.maxDepth) {
      return '';
    }

    // Build node line
    let line = prefix;
    
    if (node.depth > 0) {
      const connector = node.isLast ? '└── ' : '├── ';
      line += this.colorize(connector, 'gray', config);
    }

    // Add icon
    if (config.showIcons) {
      const icon = this.getIcon(node, config);
      line += icon + ' ';
    }

    // Add name with hub highlighting
    let name = node.name;
    if (config.showHubs && node.isHub) {
      name = this.colorize(name, 'yellow', config);
      if (node.hubScore) {
        name += ` ${this.colorize(`[${node.hubScore}]`, 'bright', config)}`;
      }
    }
    line += name;

    // Add language
    if (config.showLanguage && node.language && node.type !== 'folder') {
      const langColor = this.LANGUAGE_COLORS[node.language] || 'white';
      line += ` ${this.colorize(`[${node.language}]`, langColor, config)}`;
    }

    // Add size
    if (config.showSize && node.size) {
      line += ` ${this.colorize(`(${node.size} lines)`, 'cyan', config)}`;
    }

    // Add metrics
    if (config.showMetrics && node.metrics) {
      line += ` ${this.colorize(`(complexity: ${node.metrics.cyclomatic || 'N/A'})`, 'magenta', config)}`;
    }

    line += '\n';

    // Render children
    if (node.children.length > 0) {
      const sortedChildren = [...node.children].sort((a, b) => {
        // Sort folders first, then by name
        if (a.type === 'folder' && b.type !== 'folder') return -1;
        if (a.type !== 'folder' && b.type === 'folder') return 1;
        return a.name.localeCompare(b.name);
      });

      sortedChildren.forEach((child, index) => {
        child.isLast = index === sortedChildren.length - 1;
        child.depth = node.depth + 1;
        
        const childPrefix = node.depth > 0 
          ? prefix + (node.isLast ? '    ' : '│   ')
          : '';
        
        line += this.renderTreeNode(child, config, childPrefix);
      });
    }

    return line;
  }

  /**
   * Get icon for node type
   */
  private getIcon(node: ASCIITreeNode, config: ASCIITreeConfig): string {
    if (config.showHubs && node.isHub) {
      return config.icons.hub;
    }

    const iconMap: Record<string, string> = {
      folder: config.icons.folder,
      module: config.icons.module,
      class: config.icons.class,
      interface: config.icons.interface,
      function: config.icons.function,
      method: config.icons.method
    };

    return iconMap[node.type] || config.icons.default;
  }

  /**
   * Apply color to text if enabled
   */
  private colorize(text: string, color: string, config: ASCIITreeConfig): string {
    if (!config.colorOutput) return text;
    
    const colorCode = config.colors[color];
    if (!colorCode) return text;
    
    return `${colorCode}${text}${config.colors.reset}`;
  }

  /**
   * Calculate tree statistics
   */
  private calculateTreeStats(root: ASCIITreeNode) {
    let totalNodes = 0;
    let maxDepth = 0;
    let hubCount = 0;
    const languageDistribution: Record<string, number> = {};

    const traverse = (node: ASCIITreeNode) => {
      totalNodes++;
      maxDepth = Math.max(maxDepth, node.depth);
      
      if (node.isHub) hubCount++;
      
      if (node.language) {
        languageDistribution[node.language] = (languageDistribution[node.language] || 0) + 1;
      }

      node.children.forEach(traverse);
    };

    traverse(root);

    return {
      totalNodes,
      maxDepth,
      hubCount,
      languageDistribution
    };
  }

  /**
   * Render file tree structure (simplified version for directories)
   */
  renderFileTree(files: Array<{ path: string; type: string; size?: number }>, config: Partial<ASCIITreeConfig> = {}): string {
    const finalConfig = { ...this.DEFAULT_CONFIG, ...config };
    
    // Build tree from file paths
    const root: ASCIITreeNode = {
      id: 'root',
      name: '.',
      type: 'folder',
      children: [],
      depth: 0,
      isLast: false,
      path: ''
    };

    files.forEach(file => {
      const parts = file.path.split(/[/\\]/);
      let current = root;

      parts.forEach((part, index) => {
        let child = current.children.find(c => c.name === part);
        
        if (!child) {
          child = {
            id: file.path,
            name: part,
            type: index === parts.length - 1 ? file.type : 'folder',
            children: [],
            depth: current.depth + 1,
            isLast: false,
            path: parts.slice(0, index + 1).join('/')
          };
          current.children.push(child);
        }
        
        current = child;
      });
    });

    // Mark last children
    const markLast = (node: ASCIITreeNode) => {
      if (node.children.length > 0) {
        node.children[node.children.length - 1].isLast = true;
        node.children.forEach(markLast);
      }
    };
    markLast(root);

    return this.renderTreeNode(root, finalConfig);
  }

  /**
   * Generate project summary ASCII art
   */
  generateSummary(stats: any, config: ASCIITreeConfig = this.DEFAULT_CONFIG): string {
    const { icons } = config;
    
    let summary = '\n';
    summary += config.colorOutput ? `${config.colors.bright}📊 Project Summary${config.colors.reset}\n` : '📊 Project Summary\n';
    summary += config.colorOutput ? `${config.colors.bright}${'═'.repeat(50)}${config.colors.reset}\n` : '═'.repeat(50) + '\n';
    
    if (stats.totalFiles !== undefined) {
      summary += `${icons.folder} Total Files: ${stats.totalFiles}\n`;
    }
    
    if (stats.totalSize !== undefined) {
      summary += `${icons.file} Total Size: ${this.formatBytes(stats.totalSize)}\n`;
    }
    
    if (stats.languages) {
      summary += `${icons.module} Languages: `;
      const langs = Object.entries(stats.languages)
        .map(([lang, count]) => `${lang} (${count})`)
        .join(', ');
      summary += `${langs}\n`;
    }
    
    if (stats.hubCount > 0) {
      summary += `${icons.hub} Hub Files: ${stats.hubCount}\n`;
    }

    return summary;
  }

  /**
   * Format bytes to human readable format
   */
  private formatBytes(bytes: number): string {
    const units = ['B', 'KB', 'MB', 'GB'];
    let size = bytes;
    let unitIndex = 0;
    
    while (size >= 1024 && unitIndex < units.length - 1) {
      size /= 1024;
      unitIndex++;
    }
    
    return `${size.toFixed(1)} ${units[unitIndex]}`;
  }
}