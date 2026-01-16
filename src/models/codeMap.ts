import {
  CodeNodeData,
  CodeEdgeData,
  CodeFileData,
  AnalysisResult,
  NodeType,
  EdgeType,
} from './types'

export class CodeNode {
  constructor(public data: CodeNodeData) {}

  getHierarchyPath(): string[] {
    const path: string[] = [this.data.name]
    let current = this.data.parent
    // Note: Will need to resolve through CodeMap when integrated
    return path
  }

  getMetrics() {
    return this.data.metrics
  }

  isLeaf(): boolean {
    return !this.data.children || this.data.children.length === 0
  }
}

export class CodeEdge {
  constructor(public data: CodeEdgeData) {}

  getWeight(): number {
    return this.data.weight || 1
  }

  isDirectional(): boolean {
    return ['calls', 'imports', 'depends_on', 'inherits', 'implements'].includes(this.data.type)
  }
}

export class CodeFile {
  constructor(public data: CodeFileData) {}

  getLinesPerNode(nodeCount: number): number {
    return Math.round(this.data.lines / Math.max(nodeCount, 1))
  }

  getComplexity(): number {
    return this.data.size / Math.max(this.data.lines, 1)
  }
}

export class CodeMap {
  private nodes: Map<string, CodeNode> = new Map()
  private edges: Map<string, CodeEdge> = new Map()
  private files: Map<string, CodeFile> = new Map()
  private analysis: AnalysisResult | null = null
  private churnMap: Map<string, number> = new Map() // filePath -> churn score

  constructor(
    public id: string,
    public name: string,
    public root: string
  ) {}

  addNode(node: CodeNode): void {
    this.nodes.set(node.data.id, node)
  }

  addEdge(edge: CodeEdge): void {
    this.edges.set(edge.data.id, edge)
  }

  addFile(file: CodeFile): void {
    this.files.set(file.data.id, file)
  }

  getNode(id: string): CodeNode | undefined {
    return this.nodes.get(id)
  }

  getEdge(id: string): CodeEdge | undefined {
    return this.edges.get(id)
  }

  getFile(id: string): CodeFile | undefined {
    return this.files.get(id)
  }

  getNodes(): CodeNode[] {
    return Array.from(this.nodes.values())
  }

  getEdges(): CodeEdge[] {
    return Array.from(this.edges.values())
  }

  getFiles(): CodeFile[] {
    return Array.from(this.files.values())
  }

  getNodesByType(type: NodeType): CodeNode[] {
    return this.getNodes().filter(n => n.data.type === type)
  }

  getNodesByFile(filePath: string): CodeNode[] {
    return this.getNodes().filter(n => n.data.file === filePath)
  }

  getEdgesBySource(nodeId: string): CodeEdge[] {
    return this.getEdges().filter(e => e.data.source === nodeId)
  }

  getEdgesByTarget(nodeId: string): CodeEdge[] {
    return this.getEdges().filter(e => e.data.target === nodeId)
  }

  getEdgesByType(type: EdgeType): CodeEdge[] {
    return this.getEdges().filter(e => e.data.type === type)
  }

  // Dependency analysis
  getDependencies(nodeId: string): string[] {
    return this.getEdgesBySource(nodeId).map(e => e.data.target)
  }

  getDependents(nodeId: string): string[] {
    return this.getEdgesByTarget(nodeId).map(e => e.data.source)
  }

  getBlastRadius(nodeId: string): Set<string> {
    const visited = new Set<string>()
    const queue = [nodeId]

    while (queue.length > 0) {
      const current = queue.shift()!
      if (visited.has(current)) continue

      visited.add(current)
      const deps = this.getDependents(current)
      queue.push(...deps.filter(d => !visited.has(d)))
    }

    visited.delete(nodeId)
    return visited
  }

  // Circular dependency detection
  findCircularDependencies(): string[][] {
    const cycles: string[][] = []
    const visited = new Set<string>()
    const recursionStack = new Set<string>()

    const dfs = (nodeId: string, path: string[]): void => {
      visited.add(nodeId)
      recursionStack.add(nodeId)
      path.push(nodeId)

      const deps = this.getDependencies(nodeId)
      for (const dep of deps) {
        if (!visited.has(dep)) {
          dfs(dep, [...path])
        } else if (recursionStack.has(dep)) {
          const cycleStart = path.indexOf(dep)
          if (cycleStart !== -1) {
            cycles.push([...path.slice(cycleStart), dep])
          }
        }
      }

      recursionStack.delete(nodeId)
    }

    for (const node of this.getNodes()) {
      if (!visited.has(node.data.id)) {
        dfs(node.data.id, [])
      }
    }

    return cycles
  }

  // Coupling metrics
  calculateCoupling(): { fanIn: Map<string, number>; fanOut: Map<string, number> } {
    const fanIn = new Map<string, number>()
    const fanOut = new Map<string, number>()

    for (const node of this.getNodes()) {
      fanIn.set(node.data.id, this.getEdgesByTarget(node.data.id).length)
      fanOut.set(node.data.id, this.getEdgesBySource(node.data.id).length)
    }

    return { fanIn, fanOut }
  }

  getDeadCode(callGraphEdges: string[] = []): string[] {
    const called = new Set(callGraphEdges)
    return this.getNodes()
      .filter(n => n.data.type === 'function' || n.data.type === 'method')
      .map(n => n.data.id)
      .filter(id => !called.has(id))
  }

  setAnalysis(analysis: AnalysisResult): void {
    this.analysis = analysis
  }

  getAnalysis(): AnalysisResult | null {
    return this.analysis
  }

  // Churn tracking methods
  setChurnScore(filePath: string, score: number): void {
    this.churnMap.set(filePath, score)
  }

  getChurnScore(filePath: string): number | undefined {
    return this.churnMap.get(filePath)
  }

  setChurnScores(churnScores: Map<string, number>): void {
    this.churnMap = new Map(churnScores)
  }

  getChurnScores(): Map<string, number> {
    return new Map(this.churnMap)
  }

  getStatistics() {
    const coupling = this.calculateCoupling()
    const avgFanIn = Array.from(coupling.fanIn.values()).reduce((a, b) => a + b, 0) / this.nodes.size
    const avgFanOut = Array.from(coupling.fanOut.values()).reduce((a, b) => a + b, 0) / this.nodes.size

    return {
      totalNodes: this.nodes.size,
      totalEdges: this.edges.size,
      totalFiles: this.files.size,
      avgFanIn: Math.round(avgFanIn * 100) / 100,
      avgFanOut: Math.round(avgFanOut * 100) / 100,
      avgNodeSize: this.files.size > 0
        ? Math.round((this.getFiles().reduce((sum, f) => sum + f.data.lines, 0) / this.nodes.size) * 100) / 100
        : 0,
      totalLines: this.files.size > 0
        ? this.getFiles().reduce((sum, f) => sum + f.data.lines, 0)
        : 0,
    }
  }

  /**
   * Calculate detailed blast radius for a file
   * Based on codeflow implementation
   * Returns impact analysis showing affected files (direct and transitive)
   */
  calculateBlastRadiusDetailed(fileId: string): {
    affected: string[]
    transitive: string[]
    count: number
    transitiveCount: number
    level: 'low' | 'medium' | 'high' | 'critical'
    depth: number
    fnsUsed: number
    totalCalls: number
    dependencies: string[]
    impactScore: number
    centrality: number
  } {
    // Build adjacency lists for fast lookups
    const exportedTo = new Map<string, Set<string>>() // fileId -> Set of files that import from it
    const importedFrom = new Map<string, Set<string>>() // fileId -> Set of files it imports from
    const exportedFns = new Map<string, Map<string, number>>() // fileId -> Map of fn -> count of external calls

    // Process all edges
    for (const edge of this.getEdges()) {
      const src = edge.data.source
      const tgt = edge.data.target

      // src exports, tgt imports
      if (!exportedTo.has(src)) {
        exportedTo.set(src, new Set())
      }
      exportedTo.get(src)!.add(tgt)

      if (!importedFrom.has(tgt)) {
        importedFrom.set(tgt, new Set())
      }
      importedFrom.get(tgt)!.add(src)

      if (!exportedFns.has(src)) {
        exportedFns.set(src, new Map())
      }
      const fnMap = exportedFns.get(src)!
      const fnName = edge.data.metadata?.fn || 'unknown'
      fnMap.set(fnName, (fnMap.get(fnName) || 0) + (edge.data.weight || 1))
    }

    // 1. Direct dependents (files that directly import from this file)
    const directDeps = Array.from(exportedTo.get(fileId) || [])

    // 2. Transitive dependents (BFS with depth tracking, limit depth to 3)
    const transitive = new Map<string, number>() // fileId -> depth
    const queue = directDeps.map((f) => ({ file: f, depth: 1 }))
    const visited = new Set([fileId, ...directDeps])

    while (queue.length > 0) {
      const item = queue.shift()!
      if (item.depth > 3) continue // Limit depth to 3 for transitive
      transitive.set(item.file, item.depth)

      const nextDeps = exportedTo.get(item.file) || new Set<string>()
      for (const f of nextDeps) {
        if (!visited.has(f)) {
          visited.add(f)
          queue.push({ file: f, depth: item.depth + 1 })
        }
      }
    }

    // 3. Functions exported (how many of this file's functions are used)
    const fnUsage = exportedFns.get(fileId) || new Map<string, number>()
    const fnsUsed = fnUsage.size
    let totalCalls = 0
    fnUsage.forEach((cnt) => {
      totalCalls += cnt
    })

    // 4. Dependencies (files this file imports from - its risk)
    const dependencies = Array.from(importedFrom.get(fileId) || [])

    // 5. Calculate weighted impact score
    // Direct deps count fully, transitive count with decay
    let impactScore = directDeps.length
    transitive.forEach((depth) => {
      if (depth > 1) impactScore += 1 / depth // 0.5 for depth 2, 0.33 for depth 3
    })

    // 6. Calculate centrality (how connected is this file)
    const centrality = directDeps.length + dependencies.length + fnsUsed

    // Determine level based on direct dependents and functions used
    let level: 'low' | 'medium' | 'high' | 'critical' = 'low'
    const connectedFiles = this.getFiles().filter((f) => exportedTo.has(f.data.id) || importedFrom.has(f.data.id)).length
    const relativePct = connectedFiles > 0 ? Math.round((directDeps.length / connectedFiles) * 100) : 0

    if (directDeps.length >= 8 || fnsUsed >= 5) {
      level = 'critical'
    } else if (directDeps.length >= 4 || fnsUsed >= 3) {
      level = 'high'
    } else if (directDeps.length >= 2 || fnsUsed >= 1) {
      level = 'medium'
    }

    return {
      affected: directDeps,
      transitive: Array.from(transitive.keys()),
      count: directDeps.length,
      transitiveCount: transitive.size,
      level,
      depth: transitive.size > 0 ? Math.max(...Array.from(transitive.values())) : 0,
      fnsUsed,
      totalCalls,
      dependencies,
      impactScore: Math.round(impactScore * 10) / 10,
      centrality,
    }
  }

  serialize(): string {
    const data = {
      id: this.id,
      name: this.name,
      root: this.root,
      nodes: Array.from(this.nodes.values()).map(n => n.data),
      edges: Array.from(this.edges.values()).map(e => e.data),
      files: Array.from(this.files.values()).map(f => f.data),
      analysis: this.analysis,
      churnScores: Array.from(this.churnMap.entries()),
    }
    return JSON.stringify(data)
  }

  static deserialize(json: string): CodeMap {
    const data = JSON.parse(json)
    const map = new CodeMap(data.id, data.name, data.root)

    for (const nodeData of data.nodes) {
      map.addNode(new CodeNode(nodeData))
    }

    for (const edgeData of data.edges) {
      map.addEdge(new CodeEdge(edgeData))
    }

    for (const fileData of data.files) {
      map.addFile(new CodeFile(fileData))
    }

    if (data.analysis) {
      map.setAnalysis(data.analysis)
    }

    if (data.churnScores) {
      map.setChurnScores(new Map(data.churnScores))
    }

    return map
  }
}
