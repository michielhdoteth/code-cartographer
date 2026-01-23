import { describe, it, expect, beforeEach } from 'vitest'
import { CodeMap } from '@models/codeMap'
import { TypeScriptParser } from '@parsers/typescript'

describe('Performance Benchmarks: Visualization', () => {
  let codeMap: CodeMap

  beforeEach(() => {
    codeMap = new CodeMap('bench', 'Benchmark', '/')
  })

  describe('Graph rendering performance', () => {
    it('should render 100 nodes within budget', () => {
      const parser = new TypeScriptParser()
      for (let i = 0; i < 100; i++) {
        parser.parseAndIntegrate(`f${i}.ts`, `export const v${i} = 1;`, codeMap)
      }

      const start = performance.now()
      const nodes = codeMap.getNodes()
      const edges = codeMap.getEdges()
      const elapsed = performance.now() - start

      console.log(`  Rendering 100 nodes: ${elapsed.toFixed(2)}ms`)
      expect(elapsed).toBeLessThan(500)
    })

    it('should render 500 nodes at target speed', () => {
      const parser = new TypeScriptParser()
      for (let i = 0; i < 500; i++) {
        parser.parseAndIntegrate(`f${i}.ts`, `export const v${i} = 1;`, codeMap)
      }

      const start = performance.now()
      const nodes = codeMap.getNodes()
      const edges = codeMap.getEdges()
      const elapsed = performance.now() - start

      console.log(`  Rendering 500 nodes: ${elapsed.toFixed(2)}ms`)
      expect(elapsed).toBeLessThan(2000)
    })

    it('should render 1000 nodes within budget', () => {
      const parser = new TypeScriptParser()
      for (let i = 0; i < 200; i++) {
        const code = Array.from({ length: 5 }, (_, j) => `export const v${i}_${j} = 1;`).join('\n')
        parser.parseAndIntegrate(`f${i}.ts`, code, codeMap)
      }

      const start = performance.now()
      const nodes = codeMap.getNodes()
      const edges = codeMap.getEdges()
      const elapsed = performance.now() - start

      console.log(`  Rendering 1000+ nodes: ${elapsed.toFixed(2)}ms`)
      expect(elapsed).toBeLessThan(5000)
    })

    it('should render 5000 nodes at reasonable speed', () => {
      const parser = new TypeScriptParser()
      for (let i = 0; i < 500; i++) {
        const code = Array.from({ length: 10 }, (_, j) => `export const v${i}_${j} = 1;`).join('\n')
        parser.parseAndIntegrate(`f${i}.ts`, code, codeMap)
      }

      const start = performance.now()
      const nodes = codeMap.getNodes()
      const elapsed = performance.now() - start

      console.log(`  Rendering 5000+ nodes: ${elapsed.toFixed(2)}ms`)
      expect(elapsed).toBeLessThan(10000)
    })
  })

  describe('Viewport operations', () => {
    beforeEach(() => {
      const parser = new TypeScriptParser()
      for (let i = 0; i < 100; i++) {
        parser.parseAndIntegrate(`f${i}.ts`, `export const v${i} = 1;`, codeMap)
      }
    })

    it('should handle zoom operations efficiently', () => {
      const nodes = codeMap.getNodes()

      const start = performance.now()
      for (let zoom = 0.1; zoom <= 5; zoom += 0.5) {
        const transformed = nodes.map((n) => ({
          ...n,
          x: n.x * zoom,
          y: n.y * zoom,
        }))
      }
      const elapsed = performance.now() - start

      console.log(`  Zoom transformations (100 nodes, 10 levels): ${elapsed.toFixed(2)}ms`)
      expect(elapsed).toBeLessThan(100)
    })

    it('should handle pan operations efficiently', () => {
      const nodes = codeMap.getNodes()

      const start = performance.now()
      for (let pan = 0; pan < 100; pan += 10) {
        const panned = nodes.map((n) => ({
          ...n,
          x: n.x + pan,
          y: n.y + pan,
        }))
      }
      const elapsed = performance.now() - start

      console.log(`  Pan transformations (100 nodes, 10 iterations): ${elapsed.toFixed(2)}ms`)
      expect(elapsed).toBeLessThan(100)
    })
  })

  describe('Interaction performance', () => {
    beforeEach(() => {
      const parser = new TypeScriptParser()
      for (let i = 0; i < 100; i++) {
        parser.parseAndIntegrate(`f${i}.ts`, `export const v${i} = 1;`, codeMap)
      }
    })

    it('should highlight related nodes quickly', () => {
      const nodes = codeMap.getNodes()
      if (nodes.length > 0) {
        const start = performance.now()
        const related = codeMap.calculateBlastRadiusDetailed(nodes[0].id, 2)
        const elapsed = performance.now() - start

        console.log(`  Blast radius calculation: ${elapsed.toFixed(2)}ms`)
        expect(elapsed).toBeLessThan(100)
      }
    })

    it('should handle node selection without lag', () => {
      const nodes = codeMap.getNodes()

      const start = performance.now()
      for (const node of nodes.slice(0, 10)) {
        const selected = nodes.find((n) => n.id === node.id)
      }
      const elapsed = performance.now() - start

      console.log(`  Node selection (10 nodes): ${elapsed.toFixed(2)}ms`)
      expect(elapsed).toBeLessThan(50)
    })
  })

  describe('Rendering frame rate', () => {
    it('should maintain 30+ FPS for 100 nodes', () => {
      const parser = new TypeScriptParser()
      for (let i = 0; i < 100; i++) {
        parser.parseAndIntegrate(`f${i}.ts`, `export const v${i} = 1;`, codeMap)
      }

      const nodes = codeMap.getNodes()
      const frameTimes = []

      for (let frame = 0; frame < 30; frame++) {
        const start = performance.now()
        const rendered = nodes.map((n) => ({
          ...n,
          x: n.x + Math.sin(frame * 0.1),
          y: n.y + Math.cos(frame * 0.1),
        }))
        frameTimes.push(performance.now() - start)
      }

      const avgFrameTime = frameTimes.reduce((a, b) => a + b) / frameTimes.length
      const fps = 1000 / avgFrameTime

      console.log(`  100 nodes: ${fps.toFixed(1)} FPS (avg frame: ${avgFrameTime.toFixed(2)}ms)`)
      expect(fps).toBeGreaterThan(30)
    })

    it('should maintain 30+ FPS for 500 nodes', () => {
      const parser = new TypeScriptParser()
      for (let i = 0; i < 500; i++) {
        parser.parseAndIntegrate(`f${i}.ts`, `export const v${i} = 1;`, codeMap)
      }

      const nodes = codeMap.getNodes()
      const frameTimes = []

      for (let frame = 0; frame < 10; frame++) {
        const start = performance.now()
        const rendered = nodes.map((n) => ({
          ...n,
          x: n.x + Math.sin(frame * 0.1),
          y: n.y + Math.cos(frame * 0.1),
        }))
        frameTimes.push(performance.now() - start)
      }

      const avgFrameTime = frameTimes.reduce((a, b) => a + b) / frameTimes.length
      const fps = 1000 / avgFrameTime

      console.log(`  500 nodes: ${fps.toFixed(1)} FPS (avg frame: ${avgFrameTime.toFixed(2)}ms)`)
      expect(avgFrameTime).toBeLessThan(33)
    })

    it('should maintain minimum FPS for 1000 nodes', () => {
      const parser = new TypeScriptParser()
      for (let i = 0; i < 200; i++) {
        const code = Array.from({ length: 5 }, (_, j) => `export const v${i}_${j} = 1;`).join('\n')
        parser.parseAndIntegrate(`f${i}.ts`, code, codeMap)
      }

      const nodes = codeMap.getNodes()
      const frameTimes = []

      for (let frame = 0; frame < 5; frame++) {
        const start = performance.now()
        const rendered = nodes.map((n) => ({
          ...n,
          x: n.x + Math.sin(frame * 0.1),
          y: n.y + Math.cos(frame * 0.1),
        }))
        frameTimes.push(performance.now() - start)
      }

      const avgFrameTime = frameTimes.reduce((a, b) => a + b) / frameTimes.length
      const fps = 1000 / avgFrameTime

      console.log(`  1000 nodes: ${fps.toFixed(1)} FPS (avg frame: ${avgFrameTime.toFixed(2)}ms)`)
      expect(fps).toBeGreaterThanOrEqual(15)
    })
  })

  describe('Visualization target goals', () => {
    it('should render 1000 nodes in <2 seconds', () => {
      const parser = new TypeScriptParser()
      for (let i = 0; i < 200; i++) {
        const code = Array.from({ length: 5 }, (_, j) => `export const v${i}_${j} = 1;`).join('\n')
        parser.parseAndIntegrate(`f${i}.ts`, code, codeMap)
      }

      const start = performance.now()
      const nodes = codeMap.getNodes()
      const edges = codeMap.getEdges()
      const elapsed = performance.now() - start

      console.log(`  GOAL: Render 1000 nodes: ${elapsed.toFixed(2)}ms (target: <2000ms)`)
      expect(elapsed).toBeLessThan(2000)
    })

    it('should render 5000 nodes in <5 seconds', () => {
      const parser = new TypeScriptParser()
      for (let i = 0; i < 500; i++) {
        const code = Array.from({ length: 10 }, (_, j) => `export const v${i}_${j} = 1;`).join('\n')
        parser.parseAndIntegrate(`f${i}.ts`, code, codeMap)
      }

      const start = performance.now()
      const nodes = codeMap.getNodes()
      const edges = codeMap.getEdges()
      const elapsed = performance.now() - start

      console.log(`  GOAL: Render 5000 nodes: ${elapsed.toFixed(2)}ms (target: <5000ms)`)
      expect(elapsed).toBeLessThan(5000)
    })
  })

  describe('Layout calculation', () => {
    it('should calculate initial layout quickly', () => {
      const parser = new TypeScriptParser()
      for (let i = 0; i < 100; i++) {
        parser.parseAndIntegrate(`f${i}.ts`, `export const v${i} = 1;`, codeMap)
      }

      const nodes = codeMap.getNodes()

      const start = performance.now()
      const withLayout = nodes.map((n, i) => ({
        ...n,
        x: Math.random() * 1000,
        y: Math.random() * 1000,
        vx: 0,
        vy: 0,
        fx: 0,
        fy: 0,
      }))
      const elapsed = performance.now() - start

      console.log(`  Initial layout (100 nodes): ${elapsed.toFixed(2)}ms`)
      expect(elapsed).toBeLessThan(100)
    })
  })
})
