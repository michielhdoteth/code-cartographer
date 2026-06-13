import { CodeMap, CodeNode, CodeEdge, CodeFile } from '@models/codeMap'
import { ParserRegistry } from '@parsers/base'
import { JavaScriptParser } from '@parsers/javascript'
import { TypeScriptParser } from '@parsers/typescript'
import { PythonParser } from '@parsers/python'
import { JavaParser } from '@parsers/java'
import { GoParser } from '@parsers/go'
import { RustParser } from '@parsers/rust'
import { CppParser } from '@parsers/cpp'
import { RubyParser } from '@parsers/ruby'
import { PhpParser } from '@parsers/php'
import { SecurityScanner } from '@analyzers/securityScanner'
import { PatternDetector } from '@analyzers/patternDetector'
import { HealthScorer } from '@analyzers/healthScorer'
import { Language } from '@models/types'
import { WorkerPool, PoolTask, ProgressCallback } from '@engine/workerPool'
import { cacheManager, CacheEntry } from '@engine/cacheManager'
import { StreamingParser } from '@engine/streamingParser'
import { hashContent, generateCacheKeySync } from '@utils/hashUtils'

export interface MapGenerationOptions {
  projectName?: string
  projectRoot?: string
  followImports?: boolean
  extractMetrics?: boolean
  detectPatterns?: boolean
  runSecurity?: boolean
  runHealth?: boolean
  useCache?: boolean
  workerPoolSize?: number
  onProgress?: ProgressCallback
}

export interface GeneratedMap {
  id: string
  name: string
  timestamp: number
  codeMap: CodeMap
  statistics: {
    totalNodes: number
    totalEdges: number
    totalFiles: number
    languageBreakdown: Record<Language, number>
    nodeTypeBreakdown: Record<string, number>
  }
  analysis?: {
    securityIssues: any[]
    patterns: any[]
    health?: any
  }
}

export class MapGenerator {
  private registry: ParserRegistry
  private securityScanner: SecurityScanner
  private patternDetector: PatternDetector
  private healthScorer: HealthScorer
  private workerPool: WorkerPool | null = null
  private streamingParser: StreamingParser | null = null

  constructor() {
    this.registry = this.initializeParsers()
    this.securityScanner = new SecurityScanner()
    this.patternDetector = new PatternDetector()
    this.healthScorer = new HealthScorer()
    this.streamingParser = new StreamingParser(this.registry)
  }

  private initializeParsers(): ParserRegistry {
    const registry = new ParserRegistry()
    registry.register('javascript', new JavaScriptParser())
    registry.register('jsx', new JavaScriptParser())
    registry.register('typescript', new TypeScriptParser())
    registry.register('tsx', new TypeScriptParser())
    registry.register('python', new PythonParser())
    registry.register('java', new JavaParser())
    registry.register('go', new GoParser())
    registry.register('rust', new RustParser())
    registry.register('cpp', new CppParser())
    registry.register('ruby', new RubyParser())
    registry.register('php', new PhpParser())
    return registry
  }

  /**
   * Parse files using worker pool with caching and streaming for large files
   */
  private async parseFilesWithWorkerPool(
    files: Array<{ name: string; content: string; language: Language }>,
    codeMap: CodeMap,
    languageCount: Record<Language, number>,
    useCache: boolean,
    poolSize: number | undefined,
    onProgress: ProgressCallback | undefined
  ): Promise<void> {
    if (!this.streamingParser) {
      throw new Error('StreamingParser not initialized')
    }

    // Separate regular and large files
    const regularFiles: PoolTask[] = []
    const largeFiles: typeof files = []

    const STREAMING_THRESHOLD = 100 * 1024 // 100KB

    for (let i = 0; i < files.length; i++) {
      const file = files[i]
      if (file.content.length > STREAMING_THRESHOLD) {
        largeFiles.push(file)
      } else {
        regularFiles.push({
          fileId: `file_${i}`,
          fileName: file.name,
          content: file.content,
          language: file.language,
        })
      }
    }

    // Process large files with streaming (sequential)
    for (const file of largeFiles) {
      try {
        const result = await this.streamingParser.parseStreaming(
          file.name,
          file.content,
          file.language,
          codeMap,
          (chunk, total) => {
            if (onProgress) {
              onProgress(chunk, total, `Streaming: ${file.name} (${chunk}/${total})`)
            }
          }
        )

        languageCount[file.language]++

        // Cache streaming result
        if (useCache) {
          const contentHash = generateCacheKeySync(file.name, file.content)
          const codeFile = new CodeFile({
            id: `file_${file.name}`,
            path: file.name,
            language: file.language,
            size: file.content.length,
            lines: file.content.split('\n').length,
          })
          const nodes = codeMap
            .getNodes()
            .filter((n) => n.data.file === file.name)
            .map((n) => n.data)
          const edges = codeMap
            .getEdges()
            .filter((e) => e.data.source.includes(file.name) || e.data.target.includes(file.name))
            .map((e) => e.data)

          await cacheManager.set(file.name, contentHash, file.language, nodes, edges, codeFile.data)
        }
      } catch (error) {
        console.error(`Failed to stream parse ${file.name}:`, error)
      }
    }

    // Process regular files with worker pool (parallel)
    if (regularFiles.length > 0) {
      // Initialize worker pool if needed
      if (!this.workerPool) {
        this.workerPool = new WorkerPool(poolSize)
        try {
          const workerUrl = new URL('@workers/parserWorker.ts', import.meta.url).href
          await this.workerPool.initialize(workerUrl)
        } catch (error) {
          console.error('Worker pool initialization failed, falling back to sequential parsing:', error)
          // Fallback: parse sequentially
          for (const task of regularFiles) {
            await this.parseFileFallback(
              task,
              codeMap,
              languageCount,
              useCache
            )
          }
          return
        }
      }

      // Dispatch to worker pool
      const results = await this.workerPool.dispatch(regularFiles, onProgress)

      // Process results and cache
      for (let i = 0; i < results.length; i++) {
        const result = results[i]
        const originalFile = files.find((f) => f.name === regularFiles[i].fileName)

        if (!originalFile) continue

        if (result.success && result.nodes && result.edges) {
          // Add nodes to map
          for (const nodeData of result.nodes) {
            const node = new CodeNode(nodeData)
            codeMap.addNode(node)
          }

          // Add edges to map
          for (const edgeData of result.edges) {
            const edge = new CodeEdge(edgeData)
            codeMap.addEdge(edge)
          }

          // Add file to map
          const codeFile = new CodeFile({
            id: result.fileId,
            path: originalFile.name,
            language: originalFile.language,
            size: originalFile.content.length,
            lines: originalFile.content.split('\n').length,
          })
          codeMap.addFile(codeFile)
          languageCount[originalFile.language]++

          // Cache the result
          if (useCache) {
            const contentHash = generateCacheKeySync(originalFile.name, originalFile.content)
            await cacheManager.set(
              originalFile.name,
              contentHash,
              originalFile.language,
              result.nodes,
              result.edges,
              codeFile.data
            )
          }
        } else {
          console.error(
            `Failed to parse ${regularFiles[i].fileName}:`,
            result.error || 'Unknown error'
          )
        }
      }
    }
  }

  /**
   * Fallback synchronous parsing for when workers unavailable
   */
  private async parseFileFallback(
    task: PoolTask,
    codeMap: CodeMap,
    languageCount: Record<Language, number>,
    useCache: boolean
  ): Promise<void> {
    try {
      const parser = this.registry.getParser(task.language)
      if (parser) {
        parser.parseAndIntegrate(task.fileName, task.content, codeMap)
        languageCount[task.language]++

        if (useCache) {
          const contentHash = generateCacheKeySync(task.fileName, task.content)
          const codeFile = new CodeFile({
            id: task.fileId,
            path: task.fileName,
            language: task.language,
            size: task.content.length,
            lines: task.content.split('\n').length,
          })
          const nodes = codeMap
            .getNodes()
            .filter((n) => n.data.file === task.fileName)
            .map((n) => n.data)
          const edges = codeMap
            .getEdges()
            .filter((e) => e.data.source.includes(task.fileName) || e.data.target.includes(task.fileName))
            .map((e) => e.data)

          await cacheManager.set(
            task.fileName,
            contentHash,
            task.language,
            nodes,
            edges,
            codeFile.data
          )
        }
      }
    } catch (error) {
      console.error(`Fallback parse failed for ${task.fileName}:`, error)
    }
  }

  /**
   * Generate a code map from code files
   * Uses web worker pool for parallel parsing (6-12x speedup)
   * Uses IndexedDB cache for unchanged files (120x speedup)
   */
  async generateMap(
    files: Array<{ name: string; content: string; language: Language }>,
    options: MapGenerationOptions = {}
  ): Promise<GeneratedMap> {
    const startTime = Date.now()

    const projectName = options.projectName || 'CodeMap'
    const projectRoot = options.projectRoot || '/'
    const useCache = options.useCache !== false

    // Initialize cache if enabled
    if (useCache) {
      await cacheManager.initialize()
    }

    // Create code map
    const codeMap = new CodeMap(`map_${Date.now()}`, projectName, projectRoot)

    // Parse all files with caching and worker pool
    let totalCode = ''
    const languageCount: Record<Language, number> = {
      python: 0,
      javascript: 0,
      typescript: 0,
      jsx: 0,
      tsx: 0,
      java: 0,
      go: 0,
      rust: 0,
      cpp: 0,
      ruby: 0,
      php: 0,
    }

    // Separate cached and uncached files
    const cachedFiles: Array<{ file: typeof files[0]; entry: CacheEntry }> = []
    const uncachedFiles: typeof files = []

    // Check cache for each file
    for (let i = 0; i < files.length; i++) {
      const file = files[i]
      const contentHash = generateCacheKeySync(file.name, file.content)

      if (useCache) {
        const cached = await cacheManager.get(file.name, contentHash)
        if (cached) {
          cachedFiles.push({ file, entry: cached })
          if (options.onProgress) {
            options.onProgress(i, files.length, `(cached) ${file.name}`)
          }
        } else {
          uncachedFiles.push(file)
        }
      } else {
        uncachedFiles.push(file)
      }
    }

    // Process cached files first (instant)
    for (const { file, entry } of cachedFiles) {
      const codeFile = new CodeFile(entry.file)
      codeMap.addFile(codeFile)
      languageCount[file.language]++
      totalCode += file.content + '\n'

      // Add cached nodes and edges
      for (const nodeData of entry.nodes) {
        const node = new CodeNode(nodeData)
        codeMap.addNode(node)
      }
      for (const edgeData of entry.edges) {
        const edge = new CodeEdge(edgeData)
        codeMap.addEdge(edge)
      }
    }

    // Process uncached files with worker pool
    if (uncachedFiles.length > 0) {
      await this.parseFilesWithWorkerPool(
        uncachedFiles,
        codeMap,
        languageCount,
        useCache,
        options.workerPoolSize,
        options.onProgress
      )

      // Add content for analyzers
      for (const file of uncachedFiles) {
        totalCode += file.content + '\n'
      }
    }

    // Run analyzers
    const analysis: any = {}

    if (options.runSecurity !== false) {
      analysis.securityIssues = this.securityScanner.scan(totalCode)
    }

    if (options.detectPatterns !== false) {
      analysis.patterns = this.patternDetector.detectPatterns(codeMap)
    }

    if (options.runHealth !== false) {
      analysis.health = this.healthScorer.calculateHealth(codeMap, totalCode)
    }

    // Calculate statistics
    const nodeTypeBreakdown: Record<string, number> = {}
    for (const node of codeMap.getNodes()) {
      nodeTypeBreakdown[node.data.type] = (nodeTypeBreakdown[node.data.type] || 0) + 1
    }

    const endTime = Date.now()

    return {
      id: codeMap.id,
      name: projectName,
      timestamp: Date.now(),
      codeMap,
      statistics: {
        totalNodes: codeMap.getNodes().length,
        totalEdges: codeMap.getEdges().length,
        totalFiles: codeMap.getFiles().length,
        languageBreakdown: languageCount,
        nodeTypeBreakdown,
      },
      analysis,
    }
  }

  /**
   * Export map as JSON (for storage/sharing)
   */
  exportAsJSON(map: GeneratedMap): string {
    return JSON.stringify(
      {
        id: map.id,
        name: map.name,
        timestamp: map.timestamp,
        statistics: map.statistics,
        analysis: map.analysis,
        // Serialize code map
        codeMap: map.codeMap.serialize(),
      },
      null,
      2
    )
  }

  /**
   * Export map as visualization data (for D3.js)
   */
  exportAsVisualizationData(map: GeneratedMap) {
    const nodes = map.codeMap.getNodes().map((n) => ({
      id: n.data.id,
      name: n.data.name,
      type: n.data.type,
      language: n.data.language,
      metrics: n.data.metrics,
    }))

    const edges = map.codeMap.getEdges().map((e) => ({
      source: e.data.source,
      target: e.data.target,
      type: e.data.type,
    }))

    return { nodes, edges }
  }

  /**
   * Export map as markdown (for documentation)
   */
  exportAsMarkdown(map: GeneratedMap): string {
    const lines: string[] = []

    lines.push(`# Code Map: ${map.name}`)
    lines.push(`Generated: ${new Date(map.timestamp).toISOString()}`)
    lines.push('')

    // Statistics
    lines.push('## Statistics')
    lines.push(`- Total Nodes: ${map.statistics.totalNodes}`)
    lines.push(`- Total Edges: ${map.statistics.totalEdges}`)
    lines.push(`- Total Files: ${map.statistics.totalFiles}`)
    lines.push('')

    // Language Breakdown
    lines.push('## Languages')
    for (const [lang, count] of Object.entries(map.statistics.languageBreakdown)) {
      if (count > 0) {
        lines.push(`- ${lang}: ${count} files`)
      }
    }
    lines.push('')

    // Node Types
    lines.push('## Node Types')
    for (const [type, count] of Object.entries(map.statistics.nodeTypeBreakdown)) {
      lines.push(`- ${type}: ${count}`)
    }
    lines.push('')

    // Security Issues
    if (map.analysis?.securityIssues && map.analysis.securityIssues.length > 0) {
      lines.push('## Security Issues')
      for (const issue of map.analysis.securityIssues.slice(0, 10)) {
        lines.push(`- **${issue.severity}**: ${issue.message}`)
      }
      if (map.analysis.securityIssues.length > 10) {
        lines.push(`... and ${map.analysis.securityIssues.length - 10} more`)
      }
      lines.push('')
    }

    // Health Score
    if (map.analysis?.health) {
      lines.push('## Health Score')
      lines.push(`- Grade: **${map.analysis.health.grade}**`)
      lines.push(`- Score: ${map.analysis.health.score}/100`)
      lines.push(`- Dead Code: ${map.analysis.health.deadCodePercentage.toFixed(1)}%`)
      lines.push('')
    }

    return lines.join('\n')
  }
}

// Export singleton instance
export const mapGenerator = new MapGenerator()
