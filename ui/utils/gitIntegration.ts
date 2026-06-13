/**
 * Git Integration Module
 * 
 * Provides Git operations for diff visualization, branch comparison,
 * and change tracking within the browser environment.
 */

import { CodeNodeData as CodeNode, CodeEdgeData as CodeEdge, Language } from '../models/types';

export interface GitDiff {
  branch: string;
  commit?: string;
  files: GitFileDiff[];
  stats: GitStats;
  timestamp: Date;
}

export interface GitFileDiff {
  path: string;
  type: 'added' | 'modified' | 'deleted' | 'renamed';
  additions: number;
  deletions: number;
  oldPath?: string; // For renamed files
  content?: string; // New content (for browser-based analysis)
  oldContent?: string; // Old content (for browser-based analysis)
}

export interface GitStats {
  files: number;
  additions: number;
  deletions: number;
  changes: number;
}

export interface GitCommit {
  hash: string;
  author: string;
  message: string;
  timestamp: Date;
  files: string[];
}

export interface GitBranch {
  name: string;
  current: boolean;
  ahead?: number;
  behind?: number;
  lastCommit?: GitCommit;
}

export interface GitBlame {
  path: string;
  lines: GitBlameLine[];
}

export interface GitBlameLine {
  lineNumber: number;
  commit: GitCommit;
  originalLine: string;
}

export interface GitHistory {
  path: string;
  commits: GitCommit[];
}

/**
 * Git Integration Browser API
 * 
 * Since we're in a browser environment, we can't directly execute git commands.
 * This class provides utilities for working with git data that can be:
 * 1. Provided by the CLI tool
 * 2. Fetched from Git APIs (GitHub, GitLab, etc.)
 * 3. Analyzed from uploaded patch files
 */
export class GitIntegration {
  private readonly MAX_HISTORY_ENTRIES = 100;
  private readonly DIFF_PATTERNS = {
    added: /^\+\+\+ b\/(.+)$/,
    deleted: /^--- a\/(.+)$/,
    hunk: /^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@/,
    addition: /^\+(.*)$/,
    deletion: /^-(.*)$/,
    context: /^ (.*)$/
  };

  /**
   * Parse a unified diff string into structured data
   */
  parseDiff(diffText: string, branch: string = 'main'): GitDiff {
    const lines = diffText.split('\n');
    const files: GitFileDiff[] = [];
    let currentFile: Partial<GitFileDiff> | null = null;
    let stats: GitStats = { files: 0, additions: 0, deletions: 0, changes: 0 };

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i];

      // New file header
      const newFileMatch = line.match(this.DIFF_PATTERNS.added);
      if (newFileMatch) {
        if (currentFile) {
          files.push(currentFile as GitFileDiff);
        }
        currentFile = {
          path: newFileMatch[1],
          type: 'modified',
          additions: 0,
          deletions: 0
        };
        stats.files++;
        continue;
      }

      // Old file header
      const oldFileMatch = line.match(this.DIFF_PATTERNS.deleted);
      if (oldFileMatch && currentFile) {
        const oldPath = oldFileMatch[1];
        if (oldPath === '/dev/null') {
          currentFile.type = 'added';
        } else if (currentFile.path !== oldPath) {
          currentFile.type = 'renamed';
          currentFile.oldPath = oldPath;
        }
        continue;
      }

      // File deletion
      if (line === '+++ /dev/null' && currentFile) {
        currentFile.type = 'deleted';
        continue;
      }

      // Hunk header
      const hunkMatch = line.match(this.DIFF_PATTERNS.hunk);
      if (hunkMatch && currentFile) {
        // Process hunk for line changes
        const oldStart = parseInt(hunkMatch[1]);
        const oldLines = parseInt(hunkMatch[2] || '1');
        const newStart = parseInt(hunkMatch[3]);
        const newLines = parseInt(hunkMatch[4] || '1');

        // Count additions and deletions in this hunk
        let lineIndex = i + 1;
        let addedCount = 0;
        let deletedCount = 0;

        while (lineIndex < lines.length && !lines[lineIndex].startsWith('diff')) {
          const hunkLine = lines[lineIndex];
          if (hunkLine.startsWith('+') && !hunkLine.startsWith('+++')) {
            addedCount++;
          } else if (hunkLine.startsWith('-') && !hunkLine.startsWith('---')) {
            deletedCount++;
          } else if (hunkLine.startsWith('diff')) {
            break;
          }
          lineIndex++;
        }

        currentFile.additions = (currentFile.additions || 0) + addedCount;
        currentFile.deletions = (currentFile.deletions || 0) + deletedCount;
        stats.additions += addedCount;
        stats.deletions += deletedCount;
        stats.changes += addedCount + deletedCount;
        i = lineIndex - 1; // Skip processed lines
        continue;
      }
    }

    if (currentFile) {
      files.push(currentFile as GitFileDiff);
    }

    return {
      branch,
      files,
      stats,
      timestamp: new Date()
    };
  }

  /**
   * Parse git log output into structured commits
   */
  parseLog(logText: string): GitCommit[] {
    const commits: GitCommit[] = [];
    const commitBlocks = logText.split('\n\n');

    for (const block of commitBlocks) {
      if (!block.trim()) continue;

      const lines = block.split('\n');
      const headerMatch = lines[0].match(/^commit (\w+)$/);
      if (!headerMatch) continue;

      const hash = headerMatch[1];
      let author = '';
      let message = '';
      let timestamp = new Date();

      for (let i = 1; i < lines.length; i++) {
        const line = lines[i];
        if (line.startsWith('Author:')) {
          author = line.replace('Author:', '').trim();
        } else if (line.startsWith('Date:')) {
          const dateStr = line.replace('Date:', '').trim();
          timestamp = new Date(dateStr);
        } else if (line.trim() && !line.startsWith('    ')) {
          // Skip empty lines and continue parsing message
          continue;
        } else if (line.startsWith('    ')) {
          message += line.substring(4) + '\n';
        }
      }

      if (author && message) {
        commits.push({
          hash,
          author,
          message: message.trim(),
          timestamp,
          files: [] // Files would need to be parsed separately
        });
      }
    }

    return commits.slice(0, this.MAX_HISTORY_ENTRIES);
  }

  /**
   * Parse git blame output
   */
  parseBlame(blameText: string, filePath: string): GitBlame {
    const lines = blameText.split('\n');
    const blameLines: GitBlameLine[] = [];

    const blameRegex = /^(\w+)\s+\((.+?)\s+(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} \+\d{4})\s+(\d+)\)\s*(.*)$/;

    for (const line of lines) {
      const match = line.match(blameRegex);
      if (!match) continue;

      const [, hash, author, dateStr, lineNumber, content] = match;
      
      // Create a minimal commit object
      const commit: GitCommit = {
        hash,
        author: author.trim(),
        message: '', // Not available in blame output
        timestamp: new Date(dateStr),
        files: [filePath]
      };

      blameLines.push({
        lineNumber: parseInt(lineNumber),
        commit,
        originalLine: content
      });
    }

    return {
      path: filePath,
      lines: blameLines
    };
  }

  /**
   * Parse git branch information
   */
  parseBranches(branchText: string): GitBranch[] {
    const lines = branchText.split('\n');
    const branches: GitBranch[] = [];

    for (const line of lines) {
      if (!line.trim()) continue;

      const current = line.startsWith('*');
      const name = line.replace(/^\*\s*/, '').trim();

      // Parse ahead/behind info if available
      const aheadMatch = line.match(/ahead (\d+)/);
      const behindMatch = line.match(/behind (\d+)/);

      branches.push({
        name,
        current,
        ahead: aheadMatch ? parseInt(aheadMatch[1]) : undefined,
        behind: behindMatch ? parseInt(behindMatch[1]) : undefined
      });
    }

    return branches;
  }

  /**
   * Filter nodes and edges based on git diff
   */
  filterByDiff(nodes: CodeNode[], edges: CodeEdge[], diff: GitDiff): {
    filteredNodes: CodeNode[];
    filteredEdges: CodeEdge[];
    changedFiles: Set<string>;
  } {
    const changedFiles = new Set(
      diff.files.map(file => file.path)
    );

    // Filter nodes that are in changed files
    const filteredNodes = nodes.filter(node => 
      changedFiles.has(node.file) || 
      changedFiles.has(node.id)
    );

    // Get all node IDs in filtered set
    const nodeIds = new Set(filteredNodes.map(node => node.id));

    // Filter edges where both endpoints are in the filtered node set
    const filteredEdges = edges.filter(edge =>
      nodeIds.has(edge.source) && nodeIds.has(edge.target)
    );

    return {
      filteredNodes,
      filteredEdges,
      changedFiles
    };
  }

  /**
   * Calculate change impact based on git diff
   */
  calculateChangeImpact(nodes: CodeNode[], edges: CodeEdge[], diff: GitDiff): {
    highImpact: CodeNode[];
    mediumImpact: CodeNode[];
    lowImpact: CodeNode[];
    totalChanges: number;
  } {
    const changedFiles = new Set(
      diff.files.map(file => file.path)
    );

    const highImpact: CodeNode[] = [];
    const mediumImpact: CodeNode[] = [];
    const lowImpact: CodeNode[] = [];

    // Group nodes by file
    const nodesByFile = new Map<string, CodeNode[]>();
    nodes.forEach(node => {
      const file = node.file || node.id;
      const fileNodes = nodesByFile.get(file) || [];
      fileNodes.push(node);
      nodesByFile.set(file, fileNodes);
    });

    // Analyze each changed file
    for (const [filePath, fileNodes] of nodesByFile.entries()) {
      if (!changedFiles.has(filePath)) continue;

      const diffFile = diff.files.find(f => f.path === filePath);
      if (!diffFile) continue;

      // Calculate complexity impact
      const changeDensity = (diffFile.additions + diffFile.deletions) / Math.max(fileNodes.length, 1);
      
      // Count connections
      const connections = edges.filter(edge =>
        fileNodes.some(node => node.id === edge.source || node.id === edge.target)
      ).length;

      fileNodes.forEach(node => {
        if (connections >= 10 || changeDensity >= 5) {
          highImpact.push(node);
        } else if (connections >= 5 || changeDensity >= 2) {
          mediumImpact.push(node);
        } else {
          lowImpact.push(node);
        }
      });
    }

    return {
      highImpact,
      mediumImpact,
      lowImpact,
      totalChanges: diff.stats.changes
    };
  }

  /**
   * Generate a summary of git changes
   */
  generateChangeSummary(diff: GitDiff): string {
    const { files, additions, deletions, changes } = diff.stats;
    
    if (files === 0) {
      return 'No changes detected';
    }

    const fileTypes = new Map<string, number>();
    diff.files.forEach(file => {
      const ext = file.path.split('.').pop() || 'no-ext';
      fileTypes.set(ext, (fileTypes.get(ext) || 0) + 1);
    });

    const topTypes = Array.from(fileTypes.entries())
      .sort(([, a], [, b]) => b - a)
      .slice(0, 3)
      .map(([ext, count]) => `${ext} (${count})`)
      .join(', ');

    return `${files} file${files !== 1 ? 's' : ''} changed with ${additions} addition${additions !== 1 ? 's' : ''} and ${deletions} deletion${deletions !== 1 ? 's' : ''}. Most changed: ${topTypes}`;
  }

  /**
   * Detect hotspots from git history
   */
  detectHotspots(commits: GitCommit[], nodes: CodeNode[]): {
    files: Array<{ path: string; changeCount: number; authors: string[] }>;
    authors: Array<{ name: string; commitCount: number; files: string[] }>;
  } {
    const fileChanges = new Map<string, { count: number; authors: Set<string> }>();
    const authorCommits = new Map<string, { count: number; files: Set<string> }>();

    commits.forEach(commit => {
      // Track author commits
      const authorData = authorCommits.get(commit.author) || { count: 0, files: new Set() };
      authorData.count++;
      commit.files.forEach(file => authorData.files.add(file));
      authorCommits.set(commit.author, authorData);

      // Track file changes
      commit.files.forEach(file => {
        const fileData = fileChanges.get(file) || { count: 0, authors: new Set() };
        fileData.count++;
        fileData.authors.add(commit.author);
        fileChanges.set(file, fileData);
      });
    });

    const files = Array.from(fileChanges.entries())
      .map(([path, data]) => ({
        path,
        changeCount: data.count,
        authors: Array.from(data.authors)
      }))
      .sort((a, b) => b.changeCount - a.changeCount)
      .slice(0, 10);

    const authors = Array.from(authorCommits.entries())
      .map(([name, data]) => ({
        name,
        commitCount: data.count,
        files: Array.from(data.files)
      }))
      .sort((a, b) => b.commitCount - a.commitCount)
      .slice(0, 10);

    return { files, authors };
  }
}