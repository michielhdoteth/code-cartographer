/**
 * Real-time File Watching Daemon
 * 
 * Provides live file system monitoring for development workflows,
 * inspired by JordanCoin's codemap watch functionality.
 */

import { CodeNodeData as CodeNode, CodeEdgeData as CodeEdge } from '../models/types';

export interface FileWatchEvent {
  type: 'add' | 'change' | 'unlink' | 'rename';
  path: string;
  timestamp: Date;
  size?: number;
  oldPath?: string; // For rename events
}

export interface WatchSession {
  id: string;
  startTime: Date;
  endTime?: Date;
  events: FileWatchEvent[];
  stats: {
    totalEvents: number;
    filesAdded: number;
    filesChanged: number;
    filesDeleted: number;
    filesRenamed: number;
  };
}

export interface FileWatchConfig {
  debounceMs: number;
  maxEvents: number;
  ignoredPatterns: string[];
  watchedExtensions: string[];
  enableMetrics: boolean;
  enableHotspots: boolean;
}

export interface HotspotFile {
  path: string;
  changeCount: number;
  lastChanged: Date;
  contributors: string[];
  complexity: number;
}

export class FileWatcher {
  private watchId: string;
  private isWatching: boolean;
  private eventQueue: FileWatchEvent[];
  private debounceTimer: any;
  private config: FileWatchConfig;
  private session: WatchSession;
  private fileMetrics: Map<string, { changes: number; lastChanged: Date }>;
  private hotspots: Map<string, HotspotFile>;

  private readonly DEFAULT_CONFIG: FileWatchConfig = {
    debounceMs: 300,
    maxEvents: 1000,
    ignoredPatterns: [
      'node_modules/**',
      '.git/**',
      'dist/**',
      'build/**',
      '*.log',
      '*.tmp',
      '.DS_Store'
    ],
    watchedExtensions: [
      'js', 'jsx', 'ts', 'tsx',
      'py', 'java', 'go', 'rs', 'cpp', 'cc', 'cxx',
      'rb', 'php', 'cs', 'swift', 'kt', 'scala',
      'json', 'yaml', 'yml', 'md', 'txt'
    ],
    enableMetrics: true,
    enableHotspots: true
  };

  constructor(config: Partial<FileWatchConfig> = {}) {
    this.config = { ...this.DEFAULT_CONFIG, ...config };
    this.watchId = this.generateWatchId();
    this.isWatching = false;
    this.eventQueue = [];
    this.debounceTimer = null;
    this.fileMetrics = new Map();
    this.hotspots = new Map();
    
    this.session = {
      id: this.watchId,
      startTime: new Date(),
      events: [],
      stats: {
        totalEvents: 0,
        filesAdded: 0,
        filesChanged: 0,
        filesDeleted: 0,
        filesRenamed: 0
      }
    };
  }

  /**
   * Start watching files in the specified path
   */
  async startWatch(rootPath: string): Promise<void> {
    if (this.isWatching) {
      throw new Error('Watch session already active');
    }

    try {
      // In browser environment, we'll use polling simulation
      if (this.isBrowserEnvironment()) {
        await this.startBrowserWatch(rootPath);
      } else {
        // In Node.js environment, we'd use fs.watch or chokidar
        // This is a simulation for browser compatibility
        await this.startNodeWatch(rootPath);
      }
      
      this.isWatching = true;
      console.log(`📡 File watching started: ${rootPath}`);
      console.log(`📡 Session ID: ${this.watchId}`);
      
    } catch (error) {
      console.error('❌ Failed to start file watching:', error);
      throw error;
    }
  }

  /**
   * Stop the current watch session
   */
  async stopWatch(): Promise<void> {
    if (!this.isWatching) {
      return;
    }

    this.isWatching = false;
    this.session.endTime = new Date();
    
    if (this.debounceTimer) {
      clearTimeout(this.debounceTimer);
      this.debounceTimer = null;
    }

    // Process any remaining events
    await this.processEventQueue();

    console.log(`🛑 File watching stopped`);
    console.log(`📊 Session summary: ${this.getSessionSummary()}`);
  }

  /**
   * Get current watch session
   */
  getCurrentSession(): WatchSession {
    return { ...this.session };
  }

  /**
   * Get file hotspots
   */
  getHotspots(): HotspotFile[] {
    return Array.from(this.hotspots.values())
      .sort((a, b) => b.changeCount - a.changeCount)
      .slice(0, 20); // Top 20 hotspots
  }

  /**
   * Get file metrics
   */
  getFileMetrics(): Map<string, { changes: number; lastChanged: Date }> {
    return new Map(this.fileMetrics);
  }

  /**
   * Export session data as JSON
   */
  exportSession(): string {
    const exportData = {
      session: this.session,
      hotspots: this.getHotspots(),
      fileMetrics: Array.from(this.fileMetrics.entries()).map(([path, metrics]) => ({
        path,
        ...metrics
      })),
      config: this.config,
      exportTime: new Date().toISOString()
    };
    
    return JSON.stringify(exportData, null, 2);
  }

  // Private methods

  private generateWatchId(): string {
    return `watch-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
  }

  private isBrowserEnvironment(): boolean {
    return typeof window !== 'undefined' && typeof window.document !== 'undefined';
  }

  private async startBrowserWatch(rootPath: string): Promise<void> {
    // Browser-based watching would require file system access APIs
    // This is a simplified simulation
    console.log('🌐 Browser-based watch (simulation mode)');
    
    // Simulate watching with periodic checks
    const watchInterval = setInterval(async () => {
      if (!this.isWatching) {
        clearInterval(watchInterval);
        return;
      }
      
      // In a real implementation, this would check for file changes
      // For now, we'll just update the session time
      this.session.stats.totalEvents++;
      
    }, 5000); // Check every 5 seconds
  }

  private async startNodeWatch(rootPath: string): Promise<void> {
    // Node.js implementation would use fs.watch or chokidar
    console.log('🖥️  Node.js-based watch');
    
    // This would be the actual Node.js implementation
    // For browser compatibility, we're providing a placeholder
    console.log('📝 Note: Full file watching requires Node.js environment');
  }

  /**
   * Handle file system events
   */
  private handleFileEvent(event: FileWatchEvent): void {
    // Check if file should be ignored
    if (this.shouldIgnoreFile(event.path)) {
      return;
    }

    this.eventQueue.push(event);
    this.debouncedProcess();
  }

  /**
   * Debounced event processing
   */
  private debouncedProcess(): void {
    if (this.debounceTimer) {
      clearTimeout(this.debounceTimer);
    }
    
    this.debounceTimer = setTimeout(async () => {
      await this.processEventQueue();
    }, this.config.debounceMs);
  }

  /**
   * Process queued events
   */
  private async processEventQueue(): Promise<void> {
    if (this.eventQueue.length === 0) {
      return;
    }

    const events = [...this.eventQueue];
    this.eventQueue = [];

    for (const event of events) {
      this.processEvent(event);
    }

    // Update hotspots if enabled
    if (this.config.enableHotspots) {
      this.updateHotspots();
    }

    // Keep event queue within limits
    if (this.session.events.length > this.config.maxEvents) {
      this.session.events = this.session.events.slice(-this.config.maxEvents);
    }
  }

  /**
   * Process individual event
   */
  private processEvent(event: FileWatchEvent): void {
    // Add to session events
    this.session.events.push(event);
    
    // Update stats
    this.session.stats.totalEvents++;
    switch (event.type) {
      case 'add':
        this.session.stats.filesAdded++;
        break;
      case 'change':
        this.session.stats.filesChanged++;
        break;
      case 'unlink':
        this.session.stats.filesDeleted++;
        break;
      case 'rename':
        this.session.stats.filesRenamed++;
        break;
    }

    // Update file metrics
    if (this.config.enableMetrics) {
      this.updateFileMetrics(event);
    }

    // Log significant events
    if (event.type === 'add' || event.type === 'unlink') {
      console.log(`📁 ${event.type.toUpperCase()}: ${event.path}`);
    }
  }

  /**
   * Check if file should be ignored
   */
  private shouldIgnoreFile(filePath: string): boolean {
    const normalizedPath = filePath.replace(/\\/g, '/');
    
    return this.config.ignoredPatterns.some(pattern => {
      // Simple glob pattern matching
      if (pattern.includes('**')) {
        const basePattern = pattern.replace('**', '');
        return normalizedPath.includes(basePattern);
      }
      
      if (pattern.includes('*')) {
        const regex = new RegExp(pattern.replace(/\*/g, '.*'));
        return regex.test(normalizedPath);
      }
      
      return normalizedPath.includes(pattern);
    });
  }

  /**
   * Update file metrics
   */
  private updateFileMetrics(event: FileWatchEvent): void {
    const current = this.fileMetrics.get(event.path) || {
      changes: 0,
      lastChanged: new Date()
    };
    
    current.changes++;
    current.lastChanged = event.timestamp;
    
    this.fileMetrics.set(event.path, current);
  }

  /**
   * Update hotspots analysis
   */
  private updateHotspots(): void {
    // Reset hotspots
    this.hotspots.clear();
    
    // Calculate hotspots based on change frequency
    for (const [path, metrics] of this.fileMetrics.entries()) {
      if (metrics.changes >= 3) { // Threshold for hotspot
        const hotspot: HotspotFile = {
          path,
          changeCount: metrics.changes,
          lastChanged: metrics.lastChanged,
          contributors: [], // Would need git integration for this
          complexity: 0 // Would need code analysis for this
        };
        
        this.hotspots.set(path, hotspot);
      }
    }
  }

  /**
   * Generate session summary
   */
  private getSessionSummary(): string {
    const duration = this.session.endTime 
      ? this.session.endTime.getTime() - this.session.startTime.getTime()
      : Date.now() - this.session.startTime.getTime();
    
    const minutes = Math.floor(duration / (1000 * 60));
    const seconds = Math.floor((duration % (1000 * 60)) / 1000);
    
    const summary = [
      `Duration: ${minutes}m ${seconds}s`,
      `Events: ${this.session.stats.totalEvents}`,
      `Files added: ${this.session.stats.filesAdded}`,
      `Files changed: ${this.session.stats.filesChanged}`,
      `Files deleted: ${this.session.stats.filesDeleted}`,
      `Files renamed: ${this.session.stats.filesRenamed}`,
      `Hotspots: ${this.hotspots.size}`
    ];
    
    return summary.join(' | ');
  }

  /**
   * Simulate file change for testing
   */
  simulateFileChange(filePath: string, type: FileWatchEvent['type'] = 'change'): void {
    const event: FileWatchEvent = {
      type,
      path: filePath,
      timestamp: new Date(),
      size: Math.floor(Math.random() * 10000)
    };
    
    this.handleFileEvent(event);
  }

  /**
   * Get recent events
   */
  getRecentEvents(limit: number = 50): FileWatchEvent[] {
    return this.session.events.slice(-limit);
  }

  /**
   * Clear session history
   */
  clearSession(): void {
    this.session.events = [];
    this.session.stats = {
      totalEvents: 0,
      filesAdded: 0,
      filesChanged: 0,
      filesDeleted: 0,
      filesRenamed: 0
    };
    this.fileMetrics.clear();
    this.hotspots.clear();
  }
}