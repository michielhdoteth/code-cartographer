/**
 * Shared test fixtures and factory functions
 * Centralizes test data creation for consistency across all tests
 */

import type { CodeNode, CodeEdge, ChurnMetrics } from '@models/types'

// =============================================================================
// Node Fixtures
// =============================================================================

export interface CreateNodeOptions {
  id?: string
  name?: string
  type?: 'function' | 'class' | 'variable' | 'interface' | 'module' | 'file' | 'directory'
  language?: 'javascript' | 'typescript' | 'python' | 'java' | 'go' | 'rust' | 'cpp' | 'ruby' | 'php'
  file?: string
  line?: number
  endLine?: number
  cyclomatic?: number
  cognitive?: number
  linesOfCode?: number
  nestedDepth?: number
  numberOfParameters?: number
}

export function createTestNode(options: CreateNodeOptions = {}): CodeNode {
  const id = options.id ?? `node_${Math.random().toString(36).slice(2, 9)}`
  return {
    id,
    name: options.name ?? id,
    type: options.type ?? 'function',
    language: options.language ?? 'typescript',
    file: options.file ?? 'test.ts',
    range: {
      start: { line: options.line ?? 0, column: 0 },
      end: { line: options.endLine ?? 10, column: 0 },
    },
    metrics: {
      cyclomatic: options.cyclomatic ?? 1,
      cognitive: options.cognitive ?? 1,
      linesOfCode: options.linesOfCode ?? 10,
      nestedDepth: options.nestedDepth ?? 0,
      numberOfParameters: options.numberOfParameters ?? 0,
    },
  }
}

export function createTestNodes(count: number, baseOptions: CreateNodeOptions = {}): CodeNode[] {
  return Array.from({ length: count }, (_, i) =>
    createTestNode({
      ...baseOptions,
      id: baseOptions.id ? `${baseOptions.id}_${i}` : `node_${i}`,
      name: baseOptions.name ? `${baseOptions.name}_${i}` : `node_${i}`,
      file: baseOptions.file ?? `file_${i % 10}.ts`,
      type: baseOptions.type ?? (['function', 'class', 'variable'] as const)[i % 3],
    })
  )
}

// =============================================================================
// Edge Fixtures
// =============================================================================

export interface CreateEdgeOptions {
  id?: string
  source?: string
  target?: string
  type?: 'calls' | 'imports' | 'extends' | 'implements' | 'depends'
  weight?: number
}

export function createTestEdge(options: CreateEdgeOptions = {}): CodeEdge {
  const source = options.source ?? 'source_node'
  const target = options.target ?? 'target_node'
  return {
    id: options.id ?? `${source}_${target}`,
    source,
    target,
    type: options.type ?? 'calls',
    weight: options.weight ?? 1,
  }
}

export function createTestEdges(
  nodes: CodeNode[],
  edgeCount: number,
  options: Omit<CreateEdgeOptions, 'source' | 'target'> = {}
): CodeEdge[] {
  return Array.from({ length: edgeCount }, (_, i) => ({
    id: options.id ? `${options.id}_${i}` : `edge_${i}`,
    source: nodes[i % nodes.length].id,
    target: nodes[(i + 1) % nodes.length].id,
    type: options.type ?? 'calls',
    weight: options.weight ?? 1,
  }))
}

// =============================================================================
// Churn Metrics Fixtures
// =============================================================================

export function createTestChurnMetrics(options: Partial<ChurnMetrics> = {}): ChurnMetrics {
  return {
    commits: options.commits ?? 10,
    additions: options.additions ?? 500,
    deletions: options.deletions ?? 300,
    lastModified: options.lastModified ?? new Date(),
  }
}

// =============================================================================
// Simulation Node Fixtures (for visualizers)
// =============================================================================

export interface SimulationNode {
  id: string
  x: number
  y: number
  vx: number
  vy: number
  fx: number
  fy: number
  type?: string
  color?: string
}

export function createSimulationNode(options: Partial<SimulationNode> = {}): SimulationNode {
  return {
    id: options.id ?? `sim_node_${Math.random().toString(36).slice(2, 9)}`,
    x: options.x ?? Math.random() * 1000,
    y: options.y ?? Math.random() * 1000,
    vx: options.vx ?? 0,
    vy: options.vy ?? 0,
    fx: options.fx ?? 0,
    fy: options.fy ?? 0,
    type: options.type,
    color: options.color,
  }
}

export function createSimulationNodes(count: number, options: Partial<SimulationNode> = {}): SimulationNode[] {
  return Array.from({ length: count }, (_, i) =>
    createSimulationNode({
      ...options,
      id: options.id ? `${options.id}_${i}` : `node_${i}`,
      type: options.type ?? (i % 3 === 0 ? 'directory' : 'file'),
    })
  )
}

// =============================================================================
// Simulation Edge Fixtures
// =============================================================================

export interface SimulationEdge {
  id: string
  source: SimulationNode
  target: SimulationNode
  color?: string
  width?: number
}

export function createSimulationEdges(
  nodes: SimulationNode[],
  count: number,
  options: { color?: string; width?: number } = {}
): SimulationEdge[] {
  return Array.from({ length: count }, (_, i) => ({
    id: `edge_${i}`,
    source: nodes[i % nodes.length],
    target: nodes[(i + 1) % nodes.length],
    color: options.color,
    width: options.width,
  }))
}

// =============================================================================
// Code Samples
// =============================================================================

export const codeSamples = {
  typescript: {
    simple: 'export const x = 1;',
    classWithMethods: `
      export class Calculator {
        add(a: number, b: number): number { return a + b; }
        multiply(x: number, y: number): number { return x * y; }
      }
    `,
    interfaceAndClass: `
      export interface User {
        id: number;
        name: string;
      }
      export class UserService {
        getUser(id: number): User { return { id, name: 'John' }; }
      }
    `,
    withGenerics: `
      function identity<T>(arg: T): T { return arg; }
      class Box<T> { value: T; }
    `,
    withDecorators: `
      @Component
      class MyComponent {
        @Input() value: string;
      }
    `,
    deeplyNested: `
      function outer() {
        function middle() {
          function inner() {
            function veryInner() { return 42; }
          }
        }
      }
    `,
  },
  javascript: {
    simple: 'export const x = 1;',
    functionDeclaration: `
      function add(a, b) { return a + b; }
    `,
    arrowFunction: 'const multiply = (x, y) => x * y;',
    classDeclaration: `
      class Calculator {
        add(a, b) { return a + b; }
      }
    `,
    asyncFunction: `
      async function fetchData() {
        const data = await fetch('/api/data');
        return data;
      }
    `,
  },
  python: {
    simple: 'x = 1',
    functionDef: `
def add(a, b):
    return a + b
    `,
    classDef: `
class Calculator:
    def add(self, a, b):
        return a + b
    `,
    withDecorators: `
@property
def value(self):
    return self._value
    `,
  },
  java: {
    simple: 'public class Main {}',
    classWithMethods: `
      public class Calculator {
        public int add(int a, int b) { return a + b; }
        private double multiply(double x, double y) { return x * y; }
      }
    `,
  },
  go: {
    simple: 'package main',
    functionAndStruct: `
      package main
      import "fmt"
      func Add(a, b int) int { return a + b }
      type Calculator struct {}
      func (c *Calculator) Multiply(x, y int) int { return x * y }
    `,
  },
  rust: {
    simple: 'fn main() {}',
    structAndImpl: `
      pub fn add(a: i32, b: i32) -> i32 { a + b }
      pub struct Calculator;
      impl Calculator {
        pub fn multiply(x: i32, y: i32) -> i32 { x * y }
      }
    `,
  },
  malformed: 'function { invalid ] syntax [',
  empty: '',
  commentsOnly: '// This is a comment\n/* Multi-line comment */',
  unicode: 'const emoji = "Hello World";',
}

// =============================================================================
// Generate Large Code Samples
// =============================================================================

export function generateLargeFunctionList(count: number, language: 'typescript' | 'javascript' = 'typescript'): string {
  return Array.from({ length: count }, (_, i) => `export function func${i}() { return ${i}; }`).join('\n')
}

export function generateLargeClassList(count: number): string {
  return Array.from(
    { length: count },
    (_, i) => `
    export interface Interface${i} { value: number; }
    export class Class${i} { method() { return 1; } }
  `
  ).join('\n')
}

// =============================================================================
// Viewport and Transform Fixtures
// =============================================================================

export interface Viewport {
  x: number
  y: number
  width: number
  height: number
}

export interface Transform {
  x: number
  y: number
  k: number
}

export const viewports = {
  small: { x: 0, y: 0, width: 300, height: 300 },
  medium: { x: 0, y: 0, width: 800, height: 600 },
  large: { x: 0, y: 0, width: 1920, height: 1080 },
  ultrawide: { x: 0, y: 0, width: 3840, height: 2160 },
  offset: { x: 100, y: 100, width: 500, height: 500 },
  negative: { x: -100, y: -100, width: 300, height: 300 },
}

export const transforms = {
  identity: { x: 0, y: 0, k: 1 },
  zoomedIn: { x: 0, y: 0, k: 2 },
  zoomedOut: { x: 0, y: 0, k: 0.5 },
  panned: { x: 100, y: 100, k: 1 },
  combined: { x: 200, y: 150, k: 1.5 },
}

// =============================================================================
// Bounds Fixtures
// =============================================================================

export interface Bounds {
  minX: number
  minY: number
  maxX: number
  maxY: number
}

export const bounds = {
  standard: { minX: 0, minY: 0, maxX: 1000, maxY: 1000 },
  small: { minX: 0, minY: 0, maxX: 100, maxY: 100 },
  large: { minX: 0, minY: 0, maxX: 10000, maxY: 10000 },
  offset: { minX: 500, minY: 500, maxX: 1500, maxY: 1500 },
}

// =============================================================================
// Point Fixtures
// =============================================================================

export interface Point {
  id: string
  x: number
  y: number
}

export function createTestPoints(count: number, bounds: Bounds = { minX: 0, minY: 0, maxX: 1000, maxY: 1000 }): Point[] {
  return Array.from({ length: count }, (_, i) => ({
    id: `point_${i}`,
    x: bounds.minX + Math.random() * (bounds.maxX - bounds.minX),
    y: bounds.minY + Math.random() * (bounds.maxY - bounds.minY),
  }))
}

export function createGridPoints(rows: number, cols: number, spacing: number = 100): Point[] {
  const points: Point[] = []
  for (let row = 0; row < rows; row++) {
    for (let col = 0; col < cols; col++) {
      points.push({
        id: `point_${row}_${col}`,
        x: col * spacing,
        y: row * spacing,
      })
    }
  }
  return points
}
