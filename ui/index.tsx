import React, { useState, useEffect } from 'react'
import ReactDOM from 'react-dom/client'
import { CodeMap } from '@models/codeMap'
import { VisualizationConfig } from '@models/types'
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
import { ForceDirectedGraph } from '@visualizers/forceDirectedGraph'
import { RadialLayout } from '@visualizers/radialLayout'
import { TreeVisualization } from '@visualizers/treeVisualization'
import { TreemapVisualization } from '@visualizers/treemapVisualization'
import { MatrixVisualization } from '@visualizers/matrixVisualization'
import { FlowVisualization } from '@visualizers/flowVisualization'
import { FileExplorer } from '@components/FileExplorer'
import { LanguagePieChart } from '@components/LanguagePieChart'
import { HealthScoreDisplay } from '@components/HealthScoreDisplay'
import { ArchitectureIssuesPanel } from '@components/ArchitectureIssuesPanel'
import './styles.css'

// Initialize parser registry
const initializeParsers = (): ParserRegistry => {
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

interface AppState {
  codeMap: CodeMap | null
  selectedNodeId: string | null
  visualizationConfig: VisualizationConfig
  parsers: ParserRegistry
  securityScanner: SecurityScanner
  patternDetector: PatternDetector
  healthScorer: HealthScorer
  visualizationType: 'graph' | 'tree' | 'treemap' | 'matrix' | 'flow'
}

const App: React.FC = () => {
  const [state, setState] = useState<AppState>({
    codeMap: null,
    selectedNodeId: null,
    visualizationConfig: {
      layout: 'force',
      colorMode: 'type',
      showEdges: true,
      showLabels: true,
      nodeSize: 'lines',
      linkDistance: 100,
      chargeStrength: -300,
    },
    parsers: initializeParsers(),
    securityScanner: new SecurityScanner(),
    patternDetector: new PatternDetector(),
    healthScorer: new HealthScorer(),
    visualizationType: 'graph',
  })

  const handleLoadExample = async () => {
    try {
      const newCodeMap = new CodeMap('map_1', './examples/sample-project', './examples/sample-project')

      // Mock data: Parse example files structure with proper folder hierarchy
      const exampleFiles = [
        {
          name: 'Button.tsx',
          path: './examples/sample-project/src/components/Button.tsx',
          language: 'typescript',
          content: `import React from 'react'

interface ButtonProps {
  label: string
  onClick: () => void
  variant?: 'primary' | 'secondary'
}

export const Button: React.FC<ButtonProps> = ({ label, onClick, variant = 'primary' }) => {
  return (
    <button onClick={onClick} className={\`btn btn-\${variant}\`}>
      {label}
    </button>
  )
}

function handleClick() {
  console.log('Button clicked')
}

export function getButtonStyles() {
  return { padding: '8px 16px', borderRadius: '4px' }
}`,
        },
        {
          name: 'Card.tsx',
          path: './examples/sample-project/src/components/Card.tsx',
          language: 'typescript',
          content: `import React from 'react'

interface CardProps {
  title: string
  children: React.ReactNode
  footer?: React.ReactNode
}

export const Card: React.FC<CardProps> = ({ title, children, footer }) => {
  return (
    <div className="card">
      <div className="card-header">
        <h2>{title}</h2>
      </div>
      <div className="card-body">{children}</div>
      {footer && <div className="card-footer">{footer}</div>}
    </div>
  )
}

function renderCardContent() {
  return <div>Card content here</div>
}`,
        },
        {
          name: 'useAuth.ts',
          path: './examples/sample-project/src/hooks/useAuth.ts',
          language: 'typescript',
          content: `import { useState, useCallback } from 'react'

interface User {
  id: string
  email: string
  name: string
}

export function useAuth() {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(false)

  const login = useCallback(async (email: string, password: string) => {
    setLoading(true)
    try {
      const response = await fetch('/api/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
      })
      const data = await response.json()
      setUser(data.user)
      return data
    } finally {
      setLoading(false)
    }
  }, [])

  const logout = useCallback(() => {
    setUser(null)
  }, [])

  return { user, loading, login, logout }
}`,
        },
        {
          name: 'useForm.ts',
          path: './examples/sample-project/src/hooks/useForm.ts',
          language: 'typescript',
          content: `import { useState, useCallback } from 'react'

export function useForm<T>(initialValues: T) {
  const [values, setValues] = useState<T>(initialValues)
  const [errors, setErrors] = useState<Record<string, string>>({})

  const handleChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target
    setValues((prev) => ({ ...prev, [name]: value }))
  }, [])

  const handleSubmit = useCallback((onSubmit: (values: T) => void) => {
    return (e: React.FormEvent) => {
      e.preventDefault()
      onSubmit(values)
    }
  }, [values])

  return { values, errors, handleChange, handleSubmit }
}`,
        },
        {
          name: 'api.ts',
          path: './examples/sample-project/src/services/api.ts',
          language: 'typescript',
          content: `interface APIResponse<T> {
  data?: T
  error?: string
  status: number
}

export class APIClient {
  private baseURL: string

  constructor(baseURL: string) {
    this.baseURL = baseURL
  }

  async get<T>(endpoint: string): Promise<APIResponse<T>> {
    try {
      const response = await fetch(\`\${this.baseURL}\${endpoint}\`)
      const data = await response.json()
      return { data, status: response.status }
    } catch (error) {
      return { error: String(error), status: 500 }
    }
  }

  async post<T>(endpoint: string, payload: unknown): Promise<APIResponse<T>> {
    try {
      const response = await fetch(\`\${this.baseURL}\${endpoint}\`, {
        method: 'POST',
        body: JSON.stringify(payload),
      })
      const data = await response.json()
      return { data, status: response.status }
    } catch (error) {
      return { error: String(error), status: 500 }
    }
  }
}

export const apiClient = new APIClient(process.env.REACT_APP_API_URL || 'http://localhost:3000')`,
        },
        {
          name: 'auth.ts',
          path: './examples/sample-project/src/services/auth.ts',
          language: 'typescript',
          content: `import { apiClient } from './api'

export class AuthService {
  async login(email: string, password: string) {
    return apiClient.post('/auth/login', { email, password })
  }

  async logout() {
    return apiClient.post('/auth/logout', {})
  }

  async register(email: string, password: string, name: string) {
    return apiClient.post('/auth/register', { email, password, name })
  }

  async getProfile() {
    return apiClient.get('/auth/profile')
  }

  getToken(): string | null {
    return localStorage.getItem('token')
  }

  setToken(token: string): void {
    localStorage.setItem('token', token)
  }
}

export const authService = new AuthService()`,
        },
        {
          name: 'helpers.ts',
          path: './examples/sample-project/src/utils/helpers.ts',
          language: 'typescript',
          content: `export function formatDate(date: Date): string {
  return date.toLocaleDateString('en-US')
}

export function parseJSON(json: string) {
  try {
    return JSON.parse(json)
  } catch {
    return null
  }
}

export function debounce(fn: Function, delay: number) {
  let timeout: NodeJS.Timeout
  return function (...args: any[]) {
    clearTimeout(timeout)
    timeout = setTimeout(() => fn(...args), delay)
  }
}

export function capitalize(str: string): string {
  return str.charAt(0).toUpperCase() + str.slice(1)
}`,
        },
        {
          name: 'constants.ts',
          path: './examples/sample-project/config/constants.ts',
          language: 'typescript',
          content: `export const API_TIMEOUT = 5000
export const MAX_RETRIES = 3
export const CACHE_DURATION = 3600

export enum UserRole {
  Admin = 'admin',
  User = 'user',
  Guest = 'guest',
}

export const DEFAULT_PAGINATION = {
  page: 1,
  pageSize: 20,
}

export const ERROR_MESSAGES = {
  NETWORK_ERROR: 'Network connection failed',
  TIMEOUT: 'Request timeout',
  UNAUTHORIZED: 'Unauthorized access',
}`,
        },
        {
          name: 'helpers.test.ts',
          path: './examples/sample-project/tests/unit/helpers.test.ts',
          language: 'typescript',
          content: `import { formatDate, capitalize, debounce } from '../../src/utils/helpers'

describe('Helpers', () => {
  test('formatDate returns correct format', () => {
    const date = new Date('2024-01-15')
    expect(formatDate(date)).toContain('2024')
  })

  test('capitalize converts first letter to uppercase', () => {
    expect(capitalize('hello')).toBe('Hello')
  })

  test('debounce delays function execution', async () => {
    const mockFn = jest.fn()
    const debounced = debounce(mockFn, 100)
    debounced()
    debounced()
    expect(mockFn).not.toHaveBeenCalled()
    await new Promise((resolve) => setTimeout(resolve, 150))
    expect(mockFn).toHaveBeenCalledTimes(1)
  })
})`,
        },
      ]

      // Parse example files
      for (const file of exampleFiles) {
        const parser = state.parsers.getParser(file.language as any)
        if (parser) {
          parser.parseAndIntegrate(file.path, file.content, newCodeMap)
        }
      }

      // Run analyses
      const codeContent = exampleFiles.map((f) => f.content).join('\n')
      const securityIssues = state.securityScanner.scan(codeContent)
      const patterns = state.patternDetector.detectPatterns(newCodeMap)
      const health = state.healthScorer.calculateHealth(newCodeMap, codeContent)

      setState((prev) => ({
        ...prev,
        codeMap: newCodeMap,
      }))
    } catch (error) {
      console.error('Failed to load example:', error)
    }
  }

  // Auto-load example on mount
  useEffect(() => {
    handleLoadExample()
  }, [])

  return (
    <div className='app'>
      <header className='app-header'>
        <div className='header-content'>
          <div className='header-left'>
            <h1>Code Cartographer</h1>
            <p>Map, analyze, and visualize your codebase instantly</p>
          </div>
          <div className='header-right'>
            <button className='header-btn analyze-btn' onClick={handleLoadExample}>
              Analyze
            </button>
            <button className='header-btn' onClick={() => window.location.reload()}>
              Refresh
            </button>
            <button
              className='header-btn'
              onClick={() => {
                setState({
                  codeMap: null,
                  selectedNodeId: null,
                  visualizationConfig: {
                    layout: 'force',
                    colorMode: 'type',
                    showEdges: true,
                    showLabels: true,
                    nodeSize: 'lines',
                    linkDistance: 100,
                    chargeStrength: -300,
                  },
                  parsers: state.parsers,
                  securityScanner: state.securityScanner,
                  patternDetector: state.patternDetector,
                  healthScorer: state.healthScorer,
                  visualizationType: 'graph',
                })
              }}
            >
              Reset
            </button>
          </div>
        </div>
      </header>

      {!state.codeMap ? (
        <div className='upload-area'>
          <div className='upload-content'>
            <h2>Code Cartographer</h2>
            <p>Live analyzing folder structure...</p>
            <div className='loading-spinner'></div>
            <p className='supported-formats' style={{ marginTop: '1rem' }}>Supports: JS, TS, Python, Java, Go, Rust, C++, Ruby, PHP</p>
            <p className='supported-formats' style={{ fontSize: '0.8rem', marginTop: '2rem', color: '#9ca3af' }}>
              Use Claude Code plugin commands:<br />
              /carto:carto-scan → /carto:carto-parse → /carto:carto-graph → /carto:carto-detect
            </p>
          </div>
        </div>
      ) : (
        <div className='analyzer-container'>
          <div className='sidebar'>
            <div className='sidebar-section language-section'>
              <h4>Languages</h4>
              <LanguagePieChart codeMap={state.codeMap} />
            </div>

            <div className='sidebar-section color-by'>
              <h4>COLOR BY</h4>
              <div className='color-options'>
                <button
                  className={`color-option ${state.visualizationConfig.colorMode === 'type' ? 'active' : ''}`}
                  onClick={() => setState((prev) => ({ ...prev, visualizationConfig: { ...prev.visualizationConfig, colorMode: 'type' } }))}
                >
                  <span className='color-dot' style={{ backgroundColor: '#8b5cf6' }}></span>
                  Type
                </button>
                <button
                  className={`color-option ${state.visualizationConfig.colorMode === 'language' ? 'active' : ''}`}
                  onClick={() => setState((prev) => ({ ...prev, visualizationConfig: { ...prev.visualizationConfig, colorMode: 'language' } }))}
                >
                  <span className='color-dot' style={{ backgroundColor: '#f1e05a' }}></span>
                  Language
                </button>
                <button
                  className={`color-option ${state.visualizationConfig.colorMode === 'complexity' ? 'active' : ''}`}
                  onClick={() => setState((prev) => ({ ...prev, visualizationConfig: { ...prev.visualizationConfig, colorMode: 'complexity' } }))}
                >
                  <span className='color-dot' style={{ backgroundColor: '#ef4444' }}></span>
                  Complexity
                </button>
                <button
                  className={`color-option ${state.visualizationConfig.colorMode === 'layer' ? 'active' : ''}`}
                  onClick={() => setState((prev) => ({ ...prev, visualizationConfig: { ...prev.visualizationConfig, colorMode: 'layer' } }))}
                >
                  <span className='color-dot' style={{ backgroundColor: '#4d9fff' }}></span>
                  Layer
                </button>
                <button
                  className={`color-option ${state.visualizationConfig.colorMode === 'churn' ? 'active' : ''}`}
                  onClick={() => setState((prev) => ({ ...prev, visualizationConfig: { ...prev.visualizationConfig, colorMode: 'churn' } }))}
                >
                  <span className='color-dot' style={{ backgroundColor: '#ff6b6b' }}></span>
                  Churn
                </button>
                <button
                  className={`color-option ${state.visualizationConfig.colorMode === 'blast' ? 'active' : ''}`}
                  onClick={() => setState((prev) => ({ ...prev, visualizationConfig: { ...prev.visualizationConfig, colorMode: 'blast' } }))}
                >
                  <span className='color-dot' style={{ backgroundColor: '#ff5f5f' }}></span>
                  Blast
                </button>
                <button
                  className={`color-option ${state.visualizationConfig.colorMode === 'folder' ? 'active' : ''}`}
                  onClick={() => setState((prev) => ({ ...prev, visualizationConfig: { ...prev.visualizationConfig, colorMode: 'folder' } }))}
                >
                  <span className='color-dot' style={{ backgroundColor: '#22d3ee' }}></span>
                  Folder
                </button>
              </div>
            </div>

            <div className='sidebar-section stats-section'>
              <div className='stat-box'>
                <div className='stat-number' style={{ color: '#10b981' }}>{state.codeMap.getFiles().length}</div>
                <div className='stat-label'>FILES</div>
              </div>
              <div className='stat-box'>
                <div className='stat-number' style={{ color: '#f59e0b' }}>{state.codeMap.getNodes().length}</div>
                <div className='stat-label'>FUNCTIONS</div>
              </div>
              <div className='stat-box'>
                <div className='stat-number' style={{ color: '#06b6d4' }}>{state.codeMap.getEdges().length}</div>
                <div className='stat-label'>BUNDLES</div>
              </div>
              <div className='stat-box'>
                <div className='stat-number' style={{ color: '#10b981' }}>
                  {state.codeMap.getNodes().reduce((sum, n) => sum + (n.data.metrics?.linesOfCode || 0), 0)}
                </div>
                <div className='stat-label'>LINES OF CODE</div>
              </div>
            </div>

            <div className='sidebar-section explorer-section'>
              <h4>EXPLORER</h4>
              <FileExplorer
                codeMap={state.codeMap}
                selectedNodeId={state.selectedNodeId}
                onNodeSelect={(nodeId) => setState((prev) => ({ ...prev, selectedNodeId: nodeId }))}
              />
            </div>
          </div>

          <div className='main-content'>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: '1rem', padding: '0.75rem 1rem', backgroundColor: 'var(--bg-secondary)', borderBottom: '1px solid var(--border)', flexShrink: 0 }}>
              <div style={{ display: 'flex', gap: '0.5rem' }}>
                <button
                  className={`viz-tab ${state.visualizationType === 'graph' ? 'active' : ''}`}
                  onClick={() => setState((prev) => ({ ...prev, visualizationType: 'graph' }))}
                >
                  Graph
                </button>
                <button
                  className={`viz-tab ${state.visualizationType === 'tree' ? 'active' : ''}`}
                  onClick={() => setState((prev) => ({ ...prev, visualizationType: 'tree' }))}
                >
                  Tree
                </button>
                <button
                  className={`viz-tab ${state.visualizationType === 'treemap' ? 'active' : ''}`}
                  onClick={() => setState((prev) => ({ ...prev, visualizationType: 'treemap' }))}
                >
                  Treemap
                </button>
                <button
                  className={`viz-tab ${state.visualizationType === 'flow' ? 'active' : ''}`}
                  onClick={() => setState((prev) => ({ ...prev, visualizationType: 'flow' }))}
                >
                  Flow
                </button>
              </div>

              <div style={{ display: 'flex', alignItems: 'center', gap: '1.5rem', marginLeft: 'auto' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                  <label style={{ fontSize: '0.85rem', fontWeight: '500', color: 'var(--text-secondary)' }}>Layout:</label>
                  <select
                    value={state.visualizationConfig.layout}
                    onChange={(e) =>
                      setState((prev) => ({
                        ...prev,
                        visualizationConfig: { ...prev.visualizationConfig, layout: e.target.value as 'force' | 'radial' | 'hierarchical' | 'grid' | 'metro' | 'treemap' | 'matrix' },
                      }))
                    }
                    style={{ padding: '0.4rem 0.6rem', borderRadius: '4px', border: '1px solid var(--border)', backgroundColor: 'var(--bg)', color: 'var(--text-primary)', fontSize: '0.8rem' }}
                  >
                    <option value='force'>Force-Directed</option>
                    <option value='radial'>Radial</option>
                    <option value='hierarchical'>Hierarchical</option>
                    <option value='grid'>Grid</option>
                    <option value='metro'>Metro</option>
                    <option value='treemap'>Treemap</option>
                    <option value='matrix'>Matrix</option>
                  </select>
                </div>

                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                  <label style={{ fontSize: '0.85rem', fontWeight: '500', color: 'var(--text-secondary)' }}>Node Size:</label>
                  <select
                    value={state.visualizationConfig.nodeSize}
                    onChange={(e) =>
                      setState((prev) => ({
                        ...prev,
                        visualizationConfig: { ...prev.visualizationConfig, nodeSize: e.target.value as 'size' | 'lines' | 'complexity' },
                      }))
                    }
                    style={{ padding: '0.4rem 0.6rem', borderRadius: '4px', border: '1px solid var(--border)', backgroundColor: 'var(--bg)', color: 'var(--text-primary)', fontSize: '0.8rem' }}
                  >
                    <option value='size'>Fixed</option>
                    <option value='lines'>Lines of Code</option>
                    <option value='complexity'>Complexity</option>
                  </select>
                </div>

                <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', fontSize: '0.85rem', fontWeight: '500', color: 'var(--text-secondary)', cursor: 'pointer', whiteSpace: 'nowrap' }}>
                  <input
                    type='checkbox'
                    checked={state.visualizationConfig.showLabels}
                    onChange={(e) =>
                      setState((prev) => ({
                        ...prev,
                        visualizationConfig: { ...prev.visualizationConfig, showLabels: e.target.checked },
                      }))
                    }
                    style={{ cursor: 'pointer' }}
                  />
                  Show Labels
                </label>
              </div>
            </div>

            <div className='visualization'>
              {state.codeMap && state.codeMap.getNodes().length > 0 ? (
                <>
                  {state.visualizationType === 'graph' && (
                    <ForceDirectedGraph
                      codeMap={state.codeMap}
                      config={state.visualizationConfig}
                      onNodeSelect={(nodeId) => setState((prev) => ({ ...prev, selectedNodeId: nodeId }))}
                    />
                  )}
                  {state.visualizationType === 'tree' && (
                    <TreeVisualization
                      codeMap={state.codeMap}
                      config={state.visualizationConfig}
                      onNodeSelect={(nodeId) => setState((prev) => ({ ...prev, selectedNodeId: nodeId }))}
                    />
                  )}
                  {state.visualizationType === 'treemap' && (
                    <TreemapVisualization
                      codeMap={state.codeMap}
                      config={state.visualizationConfig}
                      onNodeSelect={(nodeId) => setState((prev) => ({ ...prev, selectedNodeId: nodeId }))}
                    />
                  )}
                  {state.visualizationType === 'matrix' && (
                    <MatrixVisualization
                      codeMap={state.codeMap}
                      config={state.visualizationConfig}
                      onNodeSelect={(nodeId) => setState((prev) => ({ ...prev, selectedNodeId: nodeId }))}
                    />
                  )}
                  {state.visualizationType === 'flow' && (
                    <FlowVisualization
                      codeMap={state.codeMap}
                      config={state.visualizationConfig}
                      onNodeSelect={(nodeId) => setState((prev) => ({ ...prev, selectedNodeId: nodeId }))}
                    />
                  )}
                </>
              ) : (
                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100%' }}>
                  <p>No code map data available</p>
                </div>
              )}
            </div>
            <div className='visualization-footer'>
              <span>{state.codeMap.getFiles().length} files</span>
              <span>{state.codeMap.getEdges().length} links</span>
            </div>
          </div>

          <div className='right-panel'>
            <ArchitectureIssuesPanel
              issues={[
                {
                  id: 'unused',
                  type: 'warning',
                  title: 'Unused Functions',
                  count: 8,
                  description: 'Functions not called from other files',
                  details: ['api.unused()', 'parser.oldMethod()', 'utils.deprecated()'],
                },
                {
                  id: 'large-files',
                  type: 'info',
                  title: 'Large Files',
                  count: 2,
                  description: 'Files exceeding 10KB',
                  details: ['src/index.tsx (15KB)', 'src/visualizers/graph.tsx (12KB)'],
                },
                {
                  id: 'circular',
                  type: 'error',
                  title: 'Circular Dependencies',
                  count: 1,
                  description: 'Modules that depend on each other',
                  details: ['UserService <-> AuthService'],
                },
              ]}
            />
          </div>
        </div>
      )}

      <footer className='app-footer'>
        <p>Code Cartographer v2.0 - Browser-based Code Analysis & Visualization</p>
      </footer>
    </div>
  )
}

// Mount React app
const root = ReactDOM.createRoot(document.getElementById('root')!)
root.render(<App />)
