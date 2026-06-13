/**
 * Skyline Visualization Renderer
 * 
 * Creates a city skyline metaphor for code visualization,
 * inspired by JordanCoin's codemap skyline mode.
 */

import { CodeNodeData as CodeNode, CodeEdgeData as CodeEdge, Language } from '../models/types';

export interface SkylineBuilding {
  id: string;
  name: string;
  x: number;
  y: number;
  width: number;
  height: number;
  language: Language;
  type: string;
  color: string;
  isHub: boolean;
  hubScore?: number;
  windows: SkylineWindow[];
}

export interface SkylineWindow {
  x: number;
  y: number;
  width: number;
  height: number;
  lit: boolean;
  color?: string;
}

export interface SkylineConfig {
  buildingWidth: number;
  maxHeight: number;
  spacing: number;
  groundLevel: number;
  windowSize: number;
  windowSpacing: number;
  hubHighlight: boolean;
  animate: boolean;
  colorMode: 'language' | 'type' | 'complexity' | 'hub';
}

export interface SkylineData {
  buildings: SkylineBuilding[];
  stats: {
    totalBuildings: number;
    avgHeight: number;
    maxHeight: number;
    hubCount: number;
    languageDistribution: Record<Language, number>;
  };
  config: SkylineConfig;
}

export class SkylineRenderer {
  private readonly DEFAULT_CONFIG: SkylineConfig = {
    buildingWidth: 40,
    maxHeight: 300,
    spacing: 10,
    groundLevel: 50,
    windowSize: 4,
    windowSpacing: 2,
    hubHighlight: true,
    animate: false,
    colorMode: 'language'
  };

  private readonly LANGUAGE_COLORS: Record<Language, string> = {
    javascript: '#f7df1e',
    typescript: '#3178c6',
    python: '#3776ab',
    java: '#007396',
    go: '#00add8',
    rust: '#ce4e2a',
    cpp: '#00599c',
    ruby: '#cc342d',
    php: '#777bb4',
    jsx: '#61dafb',
    tsx: '#61dafb'
  };

  private readonly TYPE_COLORS: Record<string, string> = {
    module: '#6366f1',
    class: '#8b5cf6',
    function: '#10b981',
    method: '#06b6d4',
    property: '#f59e0b',
    field: '#ef4444',
    interface: '#ec4899',
    enum: '#84cc16',
    constant: '#f97316',
    variable: '#64748b'
  };

  /**
   * Generate skyline visualization data from nodes and edges
   */
  generateSkyline(
    nodes: CodeNode[],
    edges: CodeEdge[],
    config: Partial<SkylineConfig> = {}
  ): SkylineData {
    const finalConfig = { ...this.DEFAULT_CONFIG, ...config };
    
    // Calculate building metrics
    const buildingMetrics = this.calculateBuildingMetrics(nodes, edges);
    
    // Generate buildings
    const buildings = this.createBuildings(nodes, buildingMetrics, finalConfig);
    
    // Calculate statistics
    const stats = this.calculateStats(buildings);
    
    return {
      buildings,
      stats,
      config: finalConfig
    };
  }

  /**
   * Calculate metrics for each building
   */
  private calculateBuildingMetrics(nodes: CodeNode[], edges: CodeEdge[]) {
    const metrics = new Map<string, {
      complexity: number;
      connections: number;
      imports: number;
      exports: number;
      isHub: boolean;
      hubScore: number;
    }>();

    // Calculate connections for each node
    const connections = new Map<string, number>();
    nodes.forEach(node => {
      const nodeConnections = edges.filter(edge => 
        edge.source === node.id || edge.target === node.id
      ).length;
      connections.set(node.id, nodeConnections);
    });

    // Calculate hub scores
    const hubScores = new Map<string, number>();
    nodes.forEach(node => {
      const importers = edges.filter(edge => edge.target === node.id).length;
      hubScores.set(node.id, importers);
    });

    // Calculate metrics for each node
    nodes.forEach(node => {
      const complexity = node.metrics.cyclomatic || 1;
      const nodeConnections = connections.get(node.id) || 0;
      const imports = edges.filter(edge => edge.source === node.id && edge.type === 'imports').length;
      const exports = edges.filter(edge => edge.target === node.id && edge.type === 'exports').length;
      const hubScore = hubScores.get(node.id) || 0;
      const isHub = hubScore >= 3; // Hub threshold

      metrics.set(node.id, {
        complexity,
        connections: nodeConnections,
        imports,
        exports,
        isHub,
        hubScore
      });
    });

    return metrics;
  }

  /**
   * Create skyline buildings from nodes
   */
  private createBuildings(
    nodes: CodeNode[],
    metrics: Map<string, any>,
    config: SkylineConfig
  ): SkylineBuilding[] {
    const buildings: SkylineBuilding[] = [];
    
    // Sort nodes by some metric (e.g., connections, complexity)
    const sortedNodes = [...nodes].sort((a, b) => {
      const aMetric = metrics.get(a.id);
      const bMetric = metrics.get(b.id);
      return (bMetric?.connections || 0) - (aMetric?.connections || 0);
    });

    sortedNodes.forEach((node, index) => {
      const metric = metrics.get(node.id);
      if (!metric) return;

      // Calculate building height based on metrics
      const height = this.calculateBuildingHeight(metric, config);
      
      // Calculate position
      const x = index * (config.buildingWidth + config.spacing);
      const y = config.groundLevel + config.maxHeight - height;

      // Get color based on color mode
      const color = this.getBuildingColor(node, metric, config);

      // Create windows
      const windows = this.createWindows(x, y, config.buildingWidth, height, config, metric);

      const building: SkylineBuilding = {
        id: node.id,
        name: node.name,
        x,
        y,
        width: config.buildingWidth,
        height,
        language: node.language,
        type: node.type,
        color,
        isHub: metric.isHub,
        hubScore: metric.hubScore,
        windows
      };

      buildings.push(building);
    });

    return buildings;
  }

  /**
   * Calculate building height based on metrics
   */
  private calculateBuildingHeight(metric: any, config: SkylineConfig): number {
    const baseHeight = 50;
    const complexityFactor = Math.min((metric.complexity || 1) * 10, 150);
    const connectionFactor = Math.min(metric.connections * 5, 100);
    
    return Math.min(baseHeight + complexityFactor + connectionFactor, config.maxHeight);
  }

  /**
   * Get building color based on color mode
   */
  private getBuildingColor(node: CodeNode, metric: any, config: SkylineConfig): string {
    switch (config.colorMode) {
      case 'language':
        return this.LANGUAGE_COLORS[node.language] || '#64748b';
      
      case 'type':
        return this.TYPE_COLORS[node.type] || '#64748b';
      
      case 'complexity':
        if (metric.complexity <= 5) return '#10b981'; // Green
        if (metric.complexity <= 10) return '#f59e0b'; // Yellow
        if (metric.complexity <= 20) return '#ef4444'; // Red
        return '#7c3aed'; // Purple
      
      case 'hub':
        if (metric.isHub) return '#f59e0b'; // Yellow for hubs
        return '#64748b'; // Gray for non-hubs
      
      default:
        return '#64748b';
    }
  }

  /**
   * Create windows for a building
   */
  private createWindows(
    buildingX: number,
    buildingY: number,
    buildingWidth: number,
    buildingHeight: number,
    config: SkylineConfig,
    metric: any
  ): SkylineWindow[] {
    const windows: SkylineWindow[] = [];
    
    const windowCols = Math.floor((buildingWidth - config.windowSpacing) / (config.windowSize + config.windowSpacing));
    const windowRows = Math.floor((buildingHeight - config.windowSpacing) / (config.windowSize + config.windowSpacing));

    for (let row = 0; row < windowRows; row++) {
      for (let col = 0; col < windowCols; col++) {
        const x = buildingX + config.windowSpacing + col * (config.windowSize + config.windowSpacing);
        const y = buildingY + config.windowSpacing + row * (config.windowSize + config.windowSpacing);
        
        // Randomly light windows based on some metric
        const lit = Math.random() > 0.3; // 70% of windows are lit
        
        windows.push({
          x,
          y,
          width: config.windowSize,
          height: config.windowSize,
          lit,
          color: lit ? '#fbbf24' : '#1f2937'
        });
      }
    }

    return windows;
  }

  /**
   * Calculate skyline statistics
   */
  private calculateStats(buildings: SkylineBuilding[]) {
    const totalBuildings = buildings.length;
    const heights = buildings.map(b => b.height);
    const avgHeight = heights.reduce((sum, h) => sum + h, 0) / totalBuildings;
    const maxHeight = Math.max(...heights);
    const hubCount = buildings.filter(b => b.isHub).length;

    const languageDistribution: Record<Language, number> = {} as any;
    buildings.forEach(building => {
      languageDistribution[building.language] = (languageDistribution[building.language] || 0) + 1;
    });

    return {
      totalBuildings,
      avgHeight,
      maxHeight,
      hubCount,
      languageDistribution
    };
  }

  /**
   * Render skyline as SVG
   */
  renderAsSVG(skylineData: SkylineData, width: number, height: number): string {
    const { buildings, config } = skylineData;
    
    // Calculate total width needed
    const totalWidth = buildings.length * (config.buildingWidth + config.spacing) - config.spacing;
    const svgWidth = Math.max(width, totalWidth);
    
    let svg = `<svg width="${svgWidth}" height="${height}" xmlns="http://www.w3.org/2000/svg">`;
    
    // Add sky gradient
    svg += `
      <defs>
        <linearGradient id="sky" x1="0%" y1="0%" x2="0%" y2="100%">
          <stop offset="0%" style="stop-color:#1e293b;stop-opacity:1" />
          <stop offset="100%" style="stop-color:#475569;stop-opacity:1" />
        </linearGradient>
      </defs>
      <rect width="${svgWidth}" height="${height}" fill="url(#sky)" />
    `;

    // Add ground
    svg += `<rect x="0" y="${config.groundLevel + config.maxHeight}" width="${svgWidth}" height="${height - config.groundLevel - config.maxHeight}" fill="#1f2937" />`;

    // Add buildings
    buildings.forEach(building => {
      // Building shadow
      svg += `<rect x="${building.x + 2}" y="${building.y + 2}" width="${building.width}" height="${building.height}" fill="#000000" opacity="0.2" />`;
      
      // Building
      svg += `<rect x="${building.x}" y="${building.y}" width="${building.width}" height="${building.height}" fill="${building.color}" stroke="#000000" stroke-width="1" />`;
      
      // Hub highlight
      if (config.hubHighlight && building.isHub) {
        svg += `<rect x="${building.x - 2}" y="${building.y - 2}" width="${building.width + 4}" height="${building.height + 4}" fill="none" stroke="#f59e0b" stroke-width="2" stroke-dasharray="5,5" />`;
      }
      
      // Windows
      building.windows.forEach(window => {
        svg += `<rect x="${window.x}" y="${window.y}" width="${window.width}" height="${window.height}" fill="${window.color}" />`;
      });
    });

    svg += '</svg>';
    return svg;
  }

  /**
   * Render skyline as ASCII art
   */
  renderAsASCII(skylineData: SkylineData): string {
    const { buildings, config } = skylineData;
    
    // Create ASCII canvas
    const canvasHeight = config.groundLevel + config.maxHeight + 10;
    const canvasWidth = buildings.length * (config.buildingWidth + config.spacing) + 10;
    const canvas: string[][] = Array(canvasHeight).fill(null).map(() => Array(canvasWidth).fill(' '));

    // Draw buildings
    buildings.forEach(building => {
      for (let y = building.y; y < building.y + building.height; y++) {
        for (let x = building.x; x < building.x + building.width; x++) {
          if (y < canvasHeight && x < canvasWidth) {
            canvas[Math.floor(y)][Math.floor(x)] = building.isHub ? '█' : '▓';
          }
        }
      }
    });

    // Draw ground
    for (let x = 0; x < canvasWidth; x++) {
      canvas[config.groundLevel + config.maxHeight][x] = '═';
    }

    // Convert to string
    return canvas.map(row => row.join('')).join('\n');
  }
}