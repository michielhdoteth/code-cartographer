import {
  CodeNodeData,
  CodeEdgeData,
  CodeFileData,
  AnalysisResult,
  NodeType,
  EdgeType,
} from './types'

const DIRECTIONAL_EDGE_TYPES = new Set(['calls', 'imports', 'depends_on', 'inherits', 'implements'])

function getOrCreateSet<K, V>(map: Map<K, Set<V>>, key: K): Set<V> {
  let set = map.get(key)
  if (!set) {
    set = new Set()
    map.set(key, set)
  }
  return set
}

export class CodeNode {
  constructor(public data: CodeNodeData) {}

  getHierarchyPath(): string[] {
    return [this.data.name]
  }

  getMetrics(): CodeNodeData['metrics'] {
    return this.data.metrics
  }

  isLeaf(): boolean {
    return !this.data.children?.length
  }
}

export class CodeEdge {
  constructor(public data: CodeEdgeData) {}

  getWeight(): number {
    return this.data.weight ?? 1
  }

  isDirectional(): boolean {
    return DIRECTIONAL_EDGE_TYPES.has(this.data.type)
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

  // Indexes for O(1) queries (Phase 3 optimization)
  private nodesByType: Map<NodeType, Set<string>> = new Map()
  private nodesByFile: Map<string, Set<string>> = new Map()
  private edgesBySource: Map<string, Set<string>> = new Map()
  private edgesByTarget: Map<string, Set<string>> = new Map()
  private edgesByType: Map<EdgeType, Set<string>> = new Map()
  private adjacencyCache: {
    exportedTo: Map<string, Set<string>>
    importedFrom: Map<string, Set<string>>
    exportedFns: Map<string, Map<string, number>>
    lastInvalidation: number
  } | null = null

  constructor(
    public id: string,
    public name: string,
    public root: string
  ) {}

  addNode(node: CodeNode): void {
    this.nodes.set(node.data.id, node)
    getOrCreateSet(this.nodesByType, node.data.type).add(node.data.id)
    getOrCreateSet(this.nodesByFile, node.data.file).add(node.data.id)
    this.adjacencyCache = null
  }

  addEdge(edge: CodeEdge): void {
    this.edges.set(edge.data.id, edge)
    getOrCreateSet(this.edgesBySource, edge.data.source).add(edge.data.id)
    getOrCreateSet(this.edgesByTarget, edge.data.target).add(edge.data.id)
    getOrCreateSet(this.edgesByType, edge.data.type).add(edge.data.id)
    this.adjacencyCache = null
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

  private resolveNodes(ids: Set<string> | undefined): CodeNode[] {
    return ids ? [...ids].map((id) => this.nodes.get(id)).filter(Boolean) as CodeNode[] : []
  }

  private resolveEdges(ids: Set<string> | undefined): CodeEdge[] {
    return ids ? [...ids].map((id) => this.edges.get(id)).filter(Boolean) as CodeEdge[] : []
  }

  getNodesByType(type: NodeType): CodeNode[] {
    return this.resolveNodes(this.nodesByType.get(type))
  }

  getNodesByFile(filePath: string): CodeNode[] {
    return this.resolveNodes(this.nodesByFile.get(filePath))
  }

  getEdgesBySource(nodeId: string): CodeEdge[] {
    return this.resolveEdges(this.edgesBySource.get(nodeId))
  }

  getEdgesByTarget(nodeId: string): CodeEdge[] {
    return this.resolveEdges(this.edgesByTarget.get(nodeId))
  }

  getEdgesByType(type: EdgeType): CodeEdge[] {
    return this.resolveEdges(this.edgesByType.get(type))
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
    const { exportedTo, importedFrom, exportedFns } = this.ensureAdjacencyCache()

    const directDeps = [...(exportedTo.get(fileId) ?? [])]
    const transitive = this.computeTransitiveDeps(fileId, directDeps, exportedTo)

    const fnUsage = exportedFns.get(fileId) ?? new Map<string, number>()
    const fnsUsed = fnUsage.size
    const totalCalls = [...fnUsage.values()].reduce((sum, cnt) => sum + cnt, 0)
    const dependencies = [...(importedFrom.get(fileId) ?? [])]

    let impactScore = directDeps.length
    transitive.forEach((depth) => {
      if (depth > 1) impactScore += 1 / depth
    })

    const centrality = directDeps.length + dependencies.length + fnsUsed
    const level = this.determineImpactLevel(directDeps.length, fnsUsed)

    return {
      affected: directDeps,
      transitive: [...transitive.keys()],
      count: directDeps.length,
      transitiveCount: transitive.size,
      level,
      depth: transitive.size > 0 ? Math.max(...transitive.values()) : 0,
      fnsUsed,
      totalCalls,
      dependencies,
      impactScore: Math.round(impactScore * 10) / 10,
      centrality,
    }
  }

  private computeTransitiveDeps(
    fileId: string,
    directDeps: string[],
    exportedTo: Map<string, Set<string>>
  ): Map<string, number> {
    const transitive = new Map<string, number>()
    const visited = new Set([fileId, ...directDeps])
    const queue = directDeps.map((f) => ({ file: f, depth: 1 }))

    while (queue.length > 0) {
      const { file, depth } = queue.shift()!
      if (depth > 3) continue

      transitive.set(file, depth)

      for (const f of exportedTo.get(file) ?? []) {
        if (!visited.has(f)) {
          visited.add(f)
          queue.push({ file: f, depth: depth + 1 })
        }
      }
    }

    return transitive
  }

  private determineImpactLevel(depCount: number, fnsUsed: number): 'low' | 'medium' | 'high' | 'critical' {
    if (depCount >= 8 || fnsUsed >= 5) return 'critical'
    if (depCount >= 4 || fnsUsed >= 3) return 'high'
    if (depCount >= 2 || fnsUsed >= 1) return 'medium'
    return 'low'
  }

  private ensureAdjacencyCache(): {
    exportedTo: Map<string, Set<string>>
    importedFrom: Map<string, Set<string>>
    exportedFns: Map<string, Map<string, number>>
  } {
    if (this.adjacencyCache) return this.adjacencyCache

    const exportedTo = new Map<string, Set<string>>()
    const importedFrom = new Map<string, Set<string>>()
    const exportedFns = new Map<string, Map<string, number>>()

    for (const edge of this.getEdges()) {
      const { source: src, target: tgt, metadata, weight } = edge.data

      getOrCreateSet(exportedTo, src).add(tgt)
      getOrCreateSet(importedFrom, tgt).add(src)

      let fnMap = exportedFns.get(src)
      if (!fnMap) {
        fnMap = new Map()
        exportedFns.set(src, fnMap)
      }
      const fnName = metadata?.fn ?? 'unknown'
      fnMap.set(fnName, (fnMap.get(fnName) ?? 0) + (weight ?? 1))
    }

    this.adjacencyCache = { exportedTo, importedFrom, exportedFns, lastInvalidation: Date.now() }
    return this.adjacencyCache
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
