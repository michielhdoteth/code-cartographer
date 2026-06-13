/**
 * Shared test helpers and utility functions
 * Centralizes common test patterns for DRY compliance
 */

import { CodeMap } from '@models/codeMap'
import { createTestNode, createTestEdge, type CreateNodeOptions, type CreateEdgeOptions } from '../fixtures/index'

// =============================================================================
// CodeMap Factory
// =============================================================================

export function createTestCodeMap(id: string = 'test', name: string = 'Test', root: string = '/'): CodeMap {
  return new CodeMap(id, name, root)
}

export function createPopulatedCodeMap(
  nodeCount: number,
  edgeCount: number = 0,
  options: { projectId?: string; projectName?: string; root?: string } = {}
): CodeMap {
  const codeMap = createTestCodeMap(options.projectId ?? 'test', options.projectName ?? 'Test', options.root ?? '/')

  for (let i = 0; i < nodeCount; i++) {
    codeMap.addNode(
      createTestNode({
        id: `node_${i}`,
        name: `node_${i}`,
        type: (['function', 'class', 'variable'] as const)[i % 3],
        file: `file_${i % 10}.ts`,
      })
    )
  }

  for (let i = 0; i < edgeCount; i++) {
    codeMap.addEdge(
      createTestEdge({
        id: `edge_${i}`,
        source: `node_${i % nodeCount}`,
        target: `node_${(i + 1) % nodeCount}`,
      })
    )
  }

  return codeMap
}

// =============================================================================
// Performance Measurement Helpers
// =============================================================================

export interface PerformanceResult {
  elapsed: number
  result: unknown
}

export function measurePerformance<T>(fn: () => T): { elapsed: number; result: T } {
  const start = performance.now()
  const result = fn()
  const elapsed = performance.now() - start
  return { elapsed, result }
}

export async function measureAsyncPerformance<T>(fn: () => Promise<T>): Promise<{ elapsed: number; result: T }> {
  const start = performance.now()
  const result = await fn()
  const elapsed = performance.now() - start
  return { elapsed, result }
}

export function measureMultipleRuns<T>(fn: () => T, iterations: number): { avgElapsed: number; results: T[] } {
  const timings: number[] = []
  const results: T[] = []

  for (let i = 0; i < iterations; i++) {
    const { elapsed, result } = measurePerformance(fn)
    timings.push(elapsed)
    results.push(result)
  }

  const avgElapsed = timings.reduce((a, b) => a + b, 0) / timings.length
  return { avgElapsed, results }
}

export function calculateFPS(frameTimes: number[]): number {
  const avgFrameTime = frameTimes.reduce((a, b) => a + b, 0) / frameTimes.length
  return 1000 / avgFrameTime
}

// =============================================================================
// Canvas Test Helpers
// =============================================================================

export function createTestCanvas(width: number = 800, height: number = 600): HTMLCanvasElement {
  const canvas = document.createElement('canvas')
  canvas.width = width
  canvas.height = height
  return canvas
}

// =============================================================================
// Assertion Helpers
// =============================================================================

export function expectPerformanceWithin(elapsed: number, maxMs: number, description?: string): void {
  const msg = description ? `${description}: ${elapsed.toFixed(2)}ms (max: ${maxMs}ms)` : undefined
  if (msg) console.log(`  ${msg}`)
  expect(elapsed).toBeLessThan(maxMs)
}

export function expectFPSAbove(frameTimes: number[], minFPS: number, description?: string): void {
  const fps = calculateFPS(frameTimes)
  const avgFrameTime = frameTimes.reduce((a, b) => a + b, 0) / frameTimes.length
  const msg = description ? `${description}: ${fps.toFixed(1)} FPS (avg frame: ${avgFrameTime.toFixed(2)}ms)` : undefined
  if (msg) console.log(`  ${msg}`)
  expect(fps).toBeGreaterThan(minFPS)
}

// =============================================================================
// Data Generation Helpers
// =============================================================================

export function generateRandomCoordinates(
  count: number,
  bounds: { minX: number; minY: number; maxX: number; maxY: number }
): Array<{ x: number; y: number }> {
  return Array.from({ length: count }, () => ({
    x: bounds.minX + Math.random() * (bounds.maxX - bounds.minX),
    y: bounds.minY + Math.random() * (bounds.maxY - bounds.minY),
  }))
}

// =============================================================================
// Test Suite Helpers
// =============================================================================

export function describePerformance(
  name: string,
  fn: () => void,
  options: { skip?: boolean; only?: boolean } = {}
): void {
  if (options.skip) {
    describe.skip(name, fn)
  } else if (options.only) {
    describe.only(name, fn)
  } else {
    describe(name, fn)
  }
}

// =============================================================================
// Error Testing Helpers
// =============================================================================

export function expectNoThrow(fn: () => void, description?: string): void {
  expect(() => fn()).not.toThrow()
  if (description) console.log(`  ${description}: passed`)
}

export async function expectAsyncNoThrow(fn: () => Promise<unknown>, description?: string): Promise<void> {
  await expect(fn()).resolves.not.toThrow()
  if (description) console.log(`  ${description}: passed`)
}

// =============================================================================
// Rendering Test Patterns
// =============================================================================

export interface RenderOptions {
  nodes: Array<{ id: string; x: number; y: number; [key: string]: unknown }>
  edges: Array<{ id: string; source: unknown; target: unknown; [key: string]: unknown }>
  transform: { x: number; y: number; k: number }
}

export function createRenderOptions(
  nodeCount: number,
  edgeCount: number,
  transform: { x: number; y: number; k: number } = { x: 0, y: 0, k: 1 }
): RenderOptions {
  const nodes = Array.from({ length: nodeCount }, (_, i) => ({
    id: `node_${i}`,
    x: Math.random() * 1000,
    y: Math.random() * 1000,
    vx: 0,
    vy: 0,
    fx: 0,
    fy: 0,
  }))

  const edges = Array.from({ length: edgeCount }, (_, i) => ({
    id: `edge_${i}`,
    source: nodes[i % nodeCount],
    target: nodes[(i + 1) % nodeCount],
  }))

  return { nodes, edges, transform }
}

// =============================================================================
// Benchmark Logging
// =============================================================================

export function logBenchmark(name: string, elapsed: number, target?: number): void {
  const targetStr = target ? ` (target: <${target}ms)` : ''
  console.log(`  ${name}: ${elapsed.toFixed(2)}ms${targetStr}`)
}

export function logGoal(name: string, elapsed: number, target: number): void {
  console.log(`  GOAL: ${name}: ${elapsed.toFixed(2)}ms (target: <${target}ms)`)
}
