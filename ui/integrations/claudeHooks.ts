/**
 * Claude Code Hooks System
 * 
 * Provides integration hooks for Claude Code to give LLMs
 * rich context about codebases, inspired by JordanCoin's codemap hooks.
 */

import { HubAnalyzer } from '../analyzers/hubAnalyzer';
import { GitIntegration } from '../utils/gitIntegration';
import { CodeNodeData as CodeNode, CodeEdgeData as CodeEdge } from '../models/types';

export interface HookContext {
  sessionId: string;
  projectPath: string;
  nodes: CodeNode[];
  edges: CodeEdge[];
  gitDiff?: string;
  branch?: string;
  lastEditTime?: Date;
  editedFiles: Set<string>;
  sessionStartTime: Date;
}

export interface HookResult {
  context: string;
  level: 'info' | 'warning' | 'error';
  metadata?: Record<string, any>;
  suggestions?: string[];
}

export interface HookConfig {
  enableHubs: boolean;
  enableGit: boolean;
  enableDiff: boolean;
  hubThreshold: number;
  maxFilesToShow: number;
  showMetrics: boolean;
  showSuggestions: boolean;
}

export class ClaudeHooks {
  private hubAnalyzer: HubAnalyzer;
  private gitIntegration: GitIntegration;
  private context: HookContext;

  constructor(config: HookConfig = {
    enableHubs: true,
    enableGit: true,
    enableDiff: true,
    hubThreshold: 3,
    maxFilesToShow: 50,
    showMetrics: true,
    showSuggestions: true
  }) {
    this.hubAnalyzer = new HubAnalyzer();
    this.gitIntegration = new GitIntegration();
    
    this.context = {
      sessionId: this.generateSessionId(),
      projectPath: '',
      nodes: [],
      edges: [],
      editedFiles: new Set(),
      sessionStartTime: new Date()
    };
  }

  /**
   * Session Start Hook
   * Called when Claude Code session starts
   */
  async sessionStart(context: Partial<HookContext>): Promise<HookResult> {
    this.context = { ...this.context, ...context };
    
    const sections: string[] = [];
    
    // Project overview
    sections.push(this.generateProjectOverview());
    
    // Hub analysis
    if (this.context.nodes.length > 0) {
      sections.push(this.generateHubContext());
    }
    
    // Git diff
    if (this.context.gitDiff) {
      sections.push(this.generateDiffContext());
    }
    
    const contextText = sections.join('\n\n');
    
    return {
      context: contextText,
      level: 'info',
      metadata: {
        hookType: 'session-start',
        nodeCount: this.context.nodes.length,
        fileCount: this.context.editedFiles.size
      }
    };
  }

  /**
   * Pre-Edit Hook
   * Called before Claude makes an edit
   */
  async preEdit(filePath: string, editType: 'write' | 'edit' = 'write'): Promise<HookResult> {
    // Update context
    this.context.lastEditTime = new Date();
    
    const sections: string[] = [];
    sections.push(`📍 File Context: ${filePath}`);
    
    // Hub analysis for specific file
    const fileHub = this.analyzeFileHub(filePath);
    if (fileHub) {
      sections.push(this.generateFileHubContext(fileHub));
    }
    
    // Dependencies analysis
    const dependencies = this.analyzeFileDependencies(filePath);
    if (dependencies.length > 0) {
      sections.push(this.generateDependencyContext(dependencies));
    }
    
    // Warning if file is a hub
    if (fileHub) {
      sections.push(this.generateHubWarning(fileHub));
    }
    
    return {
      context: sections.join('\n\n'),
      level: fileHub ? 'warning' : 'info',
      metadata: {
        hookType: 'pre-edit',
        filePath,
        editType,
        isHub: fileHub?.isHub || false
      },
      suggestions: fileHub?.isHub ? [
        'This is a hub file - changes may affect multiple files',
        'Consider running tests after modification',
        'Check for circular dependencies'
      ] : []
    };
  }

  /**
   * Post-Edit Hook
   * Called after Claude makes an edit
   */
  async postEdit(filePath: string, linesAdded: number, linesRemoved: number): Promise<HookResult> {
    // Update context
    this.context.editedFiles.add(filePath);
    this.context.lastEditTime = new Date();
    
    const sections: string[] = [];
    sections.push(`✅ Edit Complete: ${filePath}`);
    sections.push(`Changes: +${linesAdded} -${linesRemoved} lines`);
    
    // Impact analysis
    const impact = this.analyzeEditImpact(filePath, linesAdded + linesRemoved);
    if (impact.affectedFiles > 1) {
      sections.push(this.generateImpactContext(impact));
    }
    
    // Recommendations
    const recommendations = this.generateEditRecommendations(filePath, impact);
    if (recommendations.length > 0) {
      sections.push(`💡 Recommendations:\n${recommendations.map(r => `• ${r}`).join('\n')}`);
    }
    
    return {
      context: sections.join('\n\n'),
      level: impact.affectedFiles > 5 ? 'warning' : 'info',
      metadata: {
        hookType: 'post-edit',
        filePath,
        linesAdded,
        linesRemoved,
        impact
      }
    };
  }

  /**
   * Prompt Submit Hook
   * Called when user submits a prompt
   */
  async promptSubmit(prompt: string): Promise<HookResult> {
    // Extract mentioned files from prompt
    const mentionedFiles = this.extractMentionedFiles(prompt);
    
    const sections: string[] = [];
    
    // Session progress
    sections.push(this.generateSessionProgress());
    
    // Context for mentioned files
    if (mentionedFiles.length > 0) {
      sections.push(this.generateMentionedFilesContext(mentionedFiles));
    }
    
    return {
      context: sections.join('\n\n'),
      level: 'info',
      metadata: {
        hookType: 'prompt-submit',
        mentionedFiles,
        promptLength: prompt.length
      }
    };
  }

  /**
   * Pre-Compact Hook
   * Called before context is compacted
   */
  async preCompact(): Promise<HookResult> {
    // Save hub state for later recovery
    const hubState = this.saveHubState();
    
    return {
      context: '💾 Saving hub state for context recovery...',
      level: 'info',
      metadata: {
        hookType: 'pre-compact',
        hubState
      }
    };
  }

  /**
   * Session Stop Hook
   * Called when Claude Code session ends
   */
  async sessionStop(): Promise<HookResult> {
    const sections: string[] = [];
    
    // Session summary
    sections.push(this.generateSessionSummary());
    
    // Edit timeline
    const timeline = this.generateEditTimeline();
    if (timeline.length > 0) {
      sections.push(`📝 Edit Timeline:\n${timeline.join('\n')}`);
    }
    
    return {
      context: sections.join('\n\n'),
      level: 'info',
      metadata: {
        hookType: 'session-stop',
        duration: Date.now() - this.context.sessionStartTime.getTime(),
        filesEdited: this.context.editedFiles.size,
        hubState: this.saveHubState()
      }
    };
  }

  /**
   * Update context with new data
   */
  updateContext(data: Partial<HookContext>): void {
    this.context = { ...this.context, ...data };
  }

  // Private helper methods

  private generateSessionId(): string {
    return Math.random().toString(36).substring(2, 15) + 
           Math.random().toString(36).substring(2, 15);
  }

  private generateProjectOverview(): string {
    const { nodes, projectPath, branch } = this.context;
    const fileCount = nodes.length;
    const languages = this.countLanguages(nodes);
    
    let overview = `📍 Project Context:\n`;
    overview += `Path: ${projectPath}\n`;
    overview += `Files: ${fileCount}`;
    
    if (branch) {
      overview += ` | Branch: ${branch}`;
    }
    
    if (languages.size > 0) {
      const langList = Array.from(languages.entries())
        .map(([lang, count]) => `${lang} (${count})`)
        .join(', ');
      overview += `\nLanguages: ${langList}`;
    }
    
    return overview;
  }

  private generateHubContext(): string {
    const hubAnalysis = this.hubAnalyzer.analyzeHubFiles(this.context.nodes, this.context.edges);
    
    if (hubAnalysis.hubs.length === 0) {
      return '⚠️  No hub files detected in this project.';
    }
    
    let context = `⚠️  High-impact files (hubs):\n`;
    
    hubAnalysis.hubs.slice(0, 5).forEach((hub, index) => {
      context += `   ${index + 1}. ${hub.name} (${hub.language}) - imported by ${hub.score} files\n`;
    });
    
    if (hubAnalysis.hubs.length > 5) {
      context += `   ... and ${hubAnalysis.hubs.length - 5} more\n`;
    }
    
    return context;
  }

  private generateDiffContext(): string {
    if (!this.context.gitDiff) return '';
    
    const diff = this.gitIntegration.parseDiff(this.context.gitDiff, this.context.branch);
    const summary = this.gitIntegration.generateChangeSummary(diff);
    
    return `📝 Changes on branch '${diff.branch}':\n${summary}`;
  }

  private analyzeFileHub(filePath: string) {
    // Find nodes for this file
    const fileNodes = this.context.nodes.filter(node => 
      node.file === filePath || node.id.includes(filePath)
    );
    
    if (fileNodes.length === 0) return null;
    
    // Check if any nodes in this file are hubs
    const hubAnalysis = this.hubAnalyzer.analyzeHubFiles(fileNodes, this.context.edges);
    return hubAnalysis.hubs.length > 0 ? hubAnalysis.hubs[0] : null;
  }

  private generateFileHubContext(fileHub: any): string {
    return `   ⚠️  HUB FILE: ${fileHub.name} (imported by ${fileHub.score} files)\n` +
           `   Changes have wide impact across ${fileHub.blastRadius} files`;
  }

  private analyzeFileDependencies(filePath: string): string[] {
    const fileNodes = this.context.nodes.filter(node => 
      node.file === filePath || node.id.includes(filePath)
    );
    
    const fileIds = new Set(fileNodes.map(node => node.id));
    
    return this.context.edges
      .filter(edge => fileIds.has(edge.source) || fileIds.has(edge.target))
      .map(edge => edge.type === 'imports' ? edge.target : edge.source)
      .filter((id, index, array) => array.indexOf(id) === index)
      .slice(0, 10);
  }

  private generateDependencyContext(dependencies: string[]): string {
    return `🔗 Dependencies (${dependencies.length}):\n` +
           dependencies.slice(0, 5).map(dep => `   • ${dep}`).join('\n') +
           (dependencies.length > 5 ? `\n   ... and ${dependencies.length - 5} more` : '');
  }

  private generateHubWarning(fileHub: any): string {
    return `⚠️  WARNING: This is a hub file!\n` +
           `   ${fileHub.score} files depend on this code.\n` +
           `   Changes will ripple through the codebase.`;
  }

  private analyzeEditImpact(filePath: string, changeCount: number) {
    // Simplified impact calculation
    const fileHub = this.analyzeFileHub(filePath);
    const affectedFiles = fileHub ? fileHub.blastRadius : 1;
    
    return {
      filePath,
      changeCount,
      affectedFiles,
      impactLevel: affectedFiles >= 10 ? 'high' : 
                   affectedFiles >= 5 ? 'medium' : 'low'
    };
  }

  private generateImpactContext(impact: any): string {
    return `🎯 Impact Analysis:\n` +
           `   Files potentially affected: ${impact.affectedFiles}\n` +
           `   Impact level: ${impact.impactLevel}`;
  }

  private generateEditRecommendations(filePath: string, impact: any): string[] {
    const recommendations: string[] = [];
    
    if (impact.impactLevel === 'high') {
      recommendations.push('Run comprehensive tests after this change');
      recommendations.push('Consider creating a separate branch for this modification');
    }
    
    if (impact.changeCount > 50) {
      recommendations.push('Large change detected - consider breaking into smaller commits');
    }
    
    const ext = filePath.split('.').pop()?.toLowerCase();
    if (ext && ['js', 'ts', 'jsx', 'tsx'].includes(ext)) {
      recommendations.push('Consider running type checking after this change');
    }
    
    return recommendations;
  }

  private generateSessionProgress(): string {
    const duration = Date.now() - this.context.sessionStartTime.getTime();
    const filesEdited = this.context.editedFiles.size;
    const hours = Math.floor(duration / (1000 * 60 * 60));
    const minutes = Math.floor((duration % (1000 * 60 * 60)) / (1000 * 60));
    
    return `📊 Session Progress:\n` +
           `   Duration: ${hours}h ${minutes}m\n` +
           `   Files edited: ${filesEdited}\n` +
           `   Hub edits: ${this.countHubEdits()}`;
  }

  private extractMentionedFiles(prompt: string): string[] {
    // Simple regex to find file mentions - in real implementation, this would be more sophisticated
    const filePatterns = [
      /['"`]([^'"`]+\.(js|ts|py|java|go|rs|cpp|rb|php|css|html|json|yaml|yml|md))/g,
      /\b(\w+\.(js|ts|py|java|go|rs|cpp|rb|php|css|html|json|yaml|yml|md))\b/g
    ];
    
    const mentioned = new Set<string>();
    filePatterns.forEach(pattern => {
      const matches = prompt.match(pattern);
      if (matches) {
        matches.forEach(match => {
          const file = match?.replace(/['"`]/g, '') || '';
          mentioned.add(file);
        });
      }
    });
    
    return Array.from(mentioned);
  }

  private generateMentionedFilesContext(files: string[]): string {
    let context = `📍 Context for mentioned files:\n`;
    
    files.forEach(file => {
      const fileHub = this.analyzeFileHub(file);
      if (fileHub?.isHub) {
        context += `   ⚠️  ${file} is a HUB (imported by ${fileHub.score} files)\n`;
      }
    });
    
    return context;
  }

  private generateSessionSummary(): string {
    const duration = Date.now() - this.context.sessionStartTime.getTime();
    const filesEdited = this.context.editedFiles.size;
    const hubEdits = this.countHubEdits();
    
    return `📊 Session Summary:\n` +
           `   Duration: ${Math.round(duration / 1000 / 60)} minutes\n` +
           `   Files edited: ${filesEdited}\n` +
           `   Hub files modified: ${hubEdits}`;
  }

  private generateEditTimeline(): string[] {
    // Simplified timeline - in real implementation, this would track actual edits
    const timeline: string[] = [];
    const currentTime = new Date();
    
    this.context.editedFiles.forEach(file => {
      const timeStr = currentTime.toLocaleTimeString('en-US', { 
        hour12: false, 
        hour: '2-digit', 
        minute: '2-digit' 
      });
      const fileHub = this.analyzeFileHub(file);
      const hubMarker = fileHub?.isHub ? ' ⚠️HUB' : '';
      timeline.push(`  ${timeStr} EDIT ${file}${hubMarker}`);
    });
    
    return timeline;
  }

  private saveHubState(): any {
    const hubAnalysis = this.hubAnalyzer.analyzeHubFiles(this.context.nodes, this.context.edges);
    return {
      hubs: hubAnalysis.hubs,
      timestamp: Date.now(),
      sessionId: this.context.sessionId
    };
  }

  private countLanguages(nodes: CodeNode[]): Map<string, number> {
    const languages = new Map<string, number>();
    nodes.forEach(node => {
      const lang = languages.get(node.language) || 0;
      languages.set(node.language, lang + 1);
    });
    return languages;
  }

  private countHubEdits(): number {
    let hubEdits = 0;
    this.context.editedFiles.forEach(file => {
      const fileHub = this.analyzeFileHub(file);
      if (fileHub && fileHub.isHub) hubEdits++;
    });
    return hubEdits;
  }
}