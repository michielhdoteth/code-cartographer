export type Language = 'python' | 'javascript' | 'typescript' | 'jsx' | 'tsx' | 'java' | 'go' | 'rust' | 'cpp' | 'ruby' | 'php' | 'c' | 'csharp' | 'swift' | 'kotlin' | 'scala' | 'html' | 'css' | 'sql' | 'shell' | 'yaml' | 'json' | 'toml' | 'markdown'

export type NodeType =
  | 'module'
  | 'package'
  | 'class'
  | 'interface'
  | 'function'
  | 'method'
  | 'property'
  | 'field'
  | 'enum'
  | 'constant'
  | 'variable'
  | 'struct'
  | 'trait'
  | 'decorator'
  | 'import'
  | 'export'

export type EdgeType =
  | 'contains'
  | 'inherits'
  | 'implements'
  | 'imports'
  | 'exports'
  | 'calls'
  | 'references'
  | 'extends'
  | 'uses'
  | 'depends_on'

export type Severity = 'critical' | 'high' | 'medium' | 'low' | 'info'

export type HealthGrade = 'A' | 'B' | 'C' | 'D' | 'F'

export type ArchitectureLayer = 'ui' | 'components' | 'services' | 'data' | 'utils' | 'config' | 'other'

export type BlastRadiusLevel = 'low' | 'medium' | 'high' | 'critical'

export interface Position {
  line: number
  column: number
}

export interface Range {
  start: Position
  end: Position
}

export interface SecurityIssue {
  id: string
  type: 'secret' | 'sql_injection' | 'xss' | 'eval' | 'command_execution' | 'other'
  severity: Severity
  message: string
  location: Range
  suggestion?: string
  code?: string
}

export interface PatternMatch {
  type: 'design_pattern' | 'anti_pattern' | 'architecture_violation'
  name: string
  description: string
  severity: Severity
  nodes: string[] // nodeIds
  message: string
}

export interface ComplexityMetrics {
  cyclomatic: number
  cognitive: number
  linesOfCode: number
  nestedDepth: number
  numberOfParameters: number
}

export interface CodeNodeData {
  id: string
  name: string
  type: NodeType
  language: Language
  description?: string
  range: Range
  file: string
  parent?: string
  children?: string[]
  metrics: ComplexityMetrics
  isExported?: boolean
  isAbstract?: boolean
  decorators?: string[]
  modifiers?: string[]
  tags?: string[]
}

export interface CodeEdgeData {
  id: string
  source: string
  target: string
  type: EdgeType
  weight?: number
  metadata?: Record<string, any>
}

export interface ChurnMetrics {
  commits: number
  additions: number
  deletions: number
  lastModified: Date
}

export interface BlastRadiusResult {
  affected: string[]
  transitive: string[]
  count: number
  transitiveCount: number
  level: BlastRadiusLevel
  depth: number
  fnsUsed: number
  totalCalls: number
  dependencies: string[]
  impactScore: number
  centrality: number
}

export interface CodeFileData {
  id: string
  path: string
  language: Language
  size: number
  lines: number
  checksum?: string
  lastModified?: number
  imports?: string[]
  exports?: string[]
  dependencies?: string[]
  tags?: string[]
  layer?: ArchitectureLayer
  churn?: ChurnMetrics
}

export interface AnalysisResult {
  codeMap: string // CodeMap id
  timestamp: number
  securityIssues: SecurityIssue[]
  patterns: PatternMatch[]
  deadCode: string[] // nodeIds
  duplicates: string[][] // pairs of similar code
  circularDependencies: string[][] // node pairs
  layerViolations: string[]
  metrics: {
    totalNodes: number
    totalEdges: number
    avgCoupling: number
    deadCodePercentage: number
    healthGrade: HealthGrade
    securityScore: number
  }
}

export interface VisualizationConfig {
  layout: 'force' | 'radial' | 'hierarchical' | 'grid' | 'metro' | 'treemap' | 'matrix'
  colorMode: 'type' | 'language' | 'complexity' | 'layer' | 'churn' | 'blast' | 'folder'
  showEdges: boolean
  showLabels: boolean
  nodeSize: 'size' | 'lines' | 'complexity'
  linkDistance?: number
  chargeStrength?: number
  filter?: {
    language?: Language[]
    nodeType?: NodeType[]
    minComplexity?: number
    maxComplexity?: number
    hasIssues?: boolean
  }
}
