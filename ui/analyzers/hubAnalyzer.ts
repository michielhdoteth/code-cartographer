/**
 * Hub File Detection Algorithm
 * 
 * Hub files are files that are imported/used by 3+ other files.
 * They have high impact when modified and should be handled with care.
 */

import { CodeNodeData as CodeNode, CodeEdgeData as CodeEdge, Language, EdgeType } from '../models/types';

export interface HubFile {
  id: string;
  name: string;
  path: string;
  language: Language;
  score: number;
  importers: string[]; // Files that import this hub
  imports: string[];   // Files this hub imports
  impact: 'low' | 'medium' | 'high' | 'critical';
  blastRadius: number; // Number of files potentially affected
  isHub: boolean;    // Hub flag for convenience
}

export interface HubAnalysisResult {
  hubs: HubFile[];
  totalFiles: number;
  hubThreshold: number;
  analysisTime: number;
  summary: {
    critical: number;
    high: number;
    medium: number;
    low: number;
  };
}

export class HubAnalyzer {
  private readonly HUB_THRESHOLD = 3; // Files imported by 3+ others are hubs
  
  /**
   * Analyze code nodes and edges to identify hub files
   */
  analyzeHubFiles(nodes: CodeNode[], edges: CodeEdge[]): HubAnalysisResult {
    const startTime = performance.now();
    
    // Build dependency graph
    const dependencyMap = this.buildDependencyMap(nodes, edges);
    
    // Calculate hub scores
    const hubs = this.calculateHubScores(nodes, dependencyMap, edges);
    
    // Categorize hubs by impact level
    const categorizedHubs = this.categorizeHubs(hubs);
    
    const endTime = performance.now();
    
    return {
      hubs: categorizedHubs,
      totalFiles: nodes.length,
      hubThreshold: this.HUB_THRESHOLD,
      analysisTime: endTime - startTime,
      summary: this.generateSummary(categorizedHubs)
    };
  }
  
  /**
   * Build a dependency map from nodes and edges
   */
  private buildDependencyMap(nodes: CodeNode[], edges: CodeEdge[]): Map<string, Set<string>> {
    const dependencyMap = new Map<string, Set<string>>();
    
    // Initialize map for all nodes
    nodes.forEach(node => {
      dependencyMap.set(node.id, new Set<string>());
    });
    
    // Populate with import relationships
    edges
      .filter(edge => edge.type === 'imports' || edge.type === 'uses')
      .forEach(edge => {
        const importers = dependencyMap.get(edge.target) || new Set();
        importers.add(edge.source);
        dependencyMap.set(edge.target, importers);
      });
    
    return dependencyMap;
  }
  
  /**
   * Calculate hub scores for each node
   */
  private calculateHubScores(
    nodes: CodeNode[], 
    dependencyMap: Map<string, Set<string>>, 
    edges: CodeEdge[]
  ): HubFile[] {
    const hubs: HubFile[] = [];
    
    nodes.forEach(node => {
      const importers = Array.from(dependencyMap.get(node.id) || new Set<string>());
      const score = importers.length;
      
      if (score >= this.HUB_THRESHOLD) {
        // Get imports for this node
        const imports: string[] = edges
          .filter((edge: CodeEdge) => edge.source === node.id && (edge.type === 'imports' || edge.type === 'uses'))
          .map((edge: CodeEdge) => edge.target);
        
        const hub: HubFile = {
          id: node.id,
          name: node.name,
          path: node.file || node.id,
          language: node.language,
          score,
          importers,
          imports: imports as string[],
          impact: this.calculateImpact(score),
          blastRadius: this.calculateBlastRadius(node, dependencyMap, edges),
          isHub: true
        };
        
        hubs.push(hub);
      }
    });
    
    return hubs.sort((a, b) => b.score - a.score); // Sort by score descending
  }
  
  /**
   * Calculate impact level based on score
   */
  private calculateImpact(score: number): 'low' | 'medium' | 'high' | 'critical' {
    if (score >= 20) return 'critical';
    if (score >= 10) return 'high';
    if (score >= 5) return 'medium';
    return 'low';
  }
  
  /**
   * Calculate blast radius (total files potentially affected)
   */
  private calculateBlastRadius(
    node: CodeNode, 
    dependencyMap: Map<string, Set<string>>, 
    edges: CodeEdge[]
  ): number {
    // Direct importers
    const directImporters = dependencyMap.get(node.id) || new Set<string>();
    
    // For a more sophisticated blast radius, we could traverse the dependency graph
    // to find indirect dependencies. For now, we'll use direct importers + imports
    const imports: string[] = edges
      .filter((edge: CodeEdge) => edge.source === node.id && (edge.type === 'imports' || edge.type === 'uses'))
      .map((edge: CodeEdge) => edge.target);
    
    return directImporters.size + imports.length;
  }
  
  /**
   * Categorize hubs and add additional metadata
   */
  private categorizeHubs(hubs: HubFile[]): HubFile[] {
    return hubs.map(hub => ({
      ...hub,
      // Add hub file type detection
      name: this.enrichHubName(hub)
    }));
  }
  
  /**
   * Enrich hub name with type information
   */
  private enrichHubName(hub: HubFile): string {
    const { name, language } = hub;
    
    // Common patterns for hub files
    const hubPatterns: Record<string, string[]> = {
      javascript: ['index.js', 'app.js', 'main.js', 'utils.js', 'config.js', 'types.js'],
      typescript: ['index.ts', 'app.ts', 'main.ts', 'utils.ts', 'config.ts', 'types.ts'],
      python: ['__init__.py', 'main.py', 'app.py', 'utils.py', 'config.py'],
      go: ['main.go', 'utils.go', 'config.go', 'types.go'],
      java: ['Main.java', 'Application.java', 'Config.java', 'Utils.java'],
    };
    
    const patterns = hubPatterns[language] || [];
    if (patterns.some((pattern: string) => name.includes(pattern))) {
      return `${name} 🎯`;
    }
    
    return name;
  }
  
  /**
   * Generate summary statistics
   */
  private generateSummary(hubs: HubFile[]) {
    return hubs.reduce(
      (summary, hub) => {
        summary[hub.impact]++;
        return summary;
      },
      { critical: 0, high: 0, medium: 0, low: 0 }
    );
  }
  
  /**
   * Get blast radius for a specific file
   */
  getBlastRadius(fileId: string, nodes: CodeNode[], edges: CodeEdge[]): Set<string> {
    const affected = new Set<string>();
    const visited = new Set<string>();
    
    const traverse = (currentId: string, depth: number = 0) => {
      if (visited.has(currentId) || depth > 3) return; // Limit depth to prevent infinite recursion
      
      visited.add(currentId);
      affected.add(currentId);
      
      // Find all files that depend on current file
      const dependents = edges
        .filter((edge: CodeEdge) => edge.target === currentId && (edge.type === 'imports' || edge.type === 'uses'))
        .map((edge: CodeEdge) => edge.source);
      
      dependents.forEach(dependent => {
        traverse(dependent, depth + 1);
      });
    };
    
    traverse(fileId);
    return affected;
  }
  
  /**
   * Check if a file is a hub file
   */
  isHubFile(fileId: string, nodes: CodeNode[], edges: CodeEdge[]): boolean {
    const importers = edges
      .filter((edge: CodeEdge) => edge.target === fileId && (edge.type === 'imports' || edge.type === 'uses'))
      .map((edge: CodeEdge) => edge.source);
    
    return importers.length >= this.HUB_THRESHOLD;
  }
  
  /**
   * Get hub files by language
   */
  getHubsByLanguage(hubs: HubFile[]): Map<Language, HubFile[]> {
    const hubsByLanguage = new Map<Language, HubFile[]>();
    
    hubs.forEach(hub => {
      const languageHubs = hubsByLanguage.get(hub.language) || [];
      languageHubs.push(hub);
      hubsByLanguage.set(hub.language, languageHubs);
    });
    
    return hubsByLanguage;
  }
}