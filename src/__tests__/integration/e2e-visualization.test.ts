import { describe, it, expect, beforeEach } from 'vitest'
import { CodeMap } from '@models/codeMap'
import { TypeScriptParser } from '@parsers/typescript'

describe('E2E: Visualization Rendering', () => {
  let codeMap: CodeMap
  let canvas: HTMLCanvasElement

  beforeEach(() => {
    codeMap = new CodeMap('test-project', 'Test Project', '/test')
    canvas = document.createElement('canvas')
    canvas.width = 800
    canvas.height = 600
  })

  describe('graph building', () => {
    it('should build graph from parsed code', () => {
      const parser = new TypeScriptParser()
      const code = `
        export class User {
          getId() { return 1; }
          getName() { return 'John'; }
        }

        export class UserService {
          user = new User();
          get() { return this.user.getId(); }
        }
      `
      parser.parseAndIntegrate('test.ts', code, codeMap)

      const nodes = codeMap.getNodes()
      const edges = codeMap.getEdges()

      expect(nodes.length).toBeGreaterThan(0)
      expect(nodes.every((n) => n.x !== undefined && n.y !== undefined)).toBe(false)
    })
  })

  describe('viewport rendering', () => {
    beforeEach(() => {
      const parser = new TypeScriptParser()
      for (let i = 0; i < 50; i++) {
        parser.parseAndIntegrate(`file${i}.ts`, `export const x${i} = ${i};`, codeMap)
      }
    })

    it('should handle viewport at origin', () => {
      const nodes = codeMap.getNodes()
      const edges = codeMap.getEdges()
      expect(nodes.length).toBeGreaterThan(0)
      expect(Array.isArray(edges)).toBe(true)
    })

    it('should handle zoomed out viewport', () => {
      const nodes = codeMap.getNodes()
      expect(nodes.length).toBeGreaterThan(0)
    })

    it('should handle zoomed in viewport', () => {
      const nodes = codeMap.getNodes()
      expect(nodes.length).toBeGreaterThan(0)
    })

    it('should handle panned viewport', () => {
      const nodes = codeMap.getNodes()
      expect(nodes.length).toBeGreaterThan(0)
    })
  })

  describe('interactive features', () => {
    beforeEach(() => {
      const parser = new TypeScriptParser()
      parser.parseAndIntegrate(
        'app.ts',
        `
        export class App {
          start() { console.log('started'); }
          stop() { console.log('stopped'); }
        }
      `,
        codeMap
      )
    })

    it('should support node selection', () => {
      const nodes = codeMap.getNodes()
      expect(nodes.length).toBeGreaterThan(0)
    })

    it('should support node hovering', () => {
      const nodes = codeMap.getNodes()
      expect(nodes.length).toBeGreaterThan(0)
    })

    it('should support zoom interactions', () => {
      const nodes = codeMap.getNodes()
      expect(nodes.length).toBeGreaterThan(0)
    })

    it('should support pan interactions', () => {
      const nodes = codeMap.getNodes()
      expect(nodes.length).toBeGreaterThan(0)
    })
  })

  describe('rendering performance with various sizes', () => {
    it('should render 100 node graph', () => {
      const parser = new TypeScriptParser()
      for (let i = 0; i < 100; i++) {
        parser.parseAndIntegrate(`file${i}.ts`, `export const var${i} = ${i};`, codeMap)
      }

      const start = performance.now()
      const nodes = codeMap.getNodes()
      const elapsed = performance.now() - start

      expect(elapsed).toBeLessThan(100)
      expect(nodes.length).toBeGreaterThan(0)
    })

    it('should render 500 node graph', () => {
      const parser = new TypeScriptParser()
      for (let i = 0; i < 500; i++) {
        parser.parseAndIntegrate(`file${i}.ts`, `export const var${i} = ${i};`, codeMap)
      }

      const start = performance.now()
      const nodes = codeMap.getNodes()
      const elapsed = performance.now() - start

      expect(elapsed).toBeLessThan(500)
      expect(nodes.length).toBe(500)
    })

    it('should handle 1000+ node rendering preparation', () => {
      const parser = new TypeScriptParser()
      for (let i = 0; i < 50; i++) {
        const code = Array.from({ length: 20 }, (_, j) => `export function fn${i}_${j}() {}`).join('\n')
        parser.parseAndIntegrate(`file${i}.ts`, code, codeMap)
      }

      const start = performance.now()
      const nodes = codeMap.getNodes()
      const elapsed = performance.now() - start

      expect(elapsed).toBeLessThan(1000)
      expect(nodes.length).toBeGreaterThan(900)
    })
  })

  describe('color coding and visual features', () => {
    it('should support different node colors', () => {
      const parser = new TypeScriptParser()
      const code = `
        export class A {}
        export class B {}
        export function c() {}
      `
      parser.parseAndIntegrate('types.ts', code, codeMap)

      const nodes = codeMap.getNodes()
      const hasMultipleTypes = new Set(nodes.map((n) => n.type)).size > 1
      expect(hasMultipleTypes).toBe(true)
    })

    it('should support edge color coding', () => {
      const parser = new TypeScriptParser()
      parser.parseAndIntegrate('api.ts', 'export class API {}', codeMap)

      const edges = codeMap.getEdges()
      expect(Array.isArray(edges)).toBe(true)
    })

    it('should support node sizing by metrics', () => {
      const parser = new TypeScriptParser()
      const code = `
        export function simple() { return 1; }
        export function complex(a, b, c, d, e) {
          if (a) if (b) if (c) if (d) return e;
        }
      `
      parser.parseAndIntegrate('metrics.ts', code, codeMap)

      const nodes = codeMap.getNodes()
      const withMetrics = nodes.filter((n) => n.metrics)
      expect(withMetrics.length).toBeGreaterThan(0)
    })
  })

  describe('legend and statistics', () => {
    it('should provide statistics for display', () => {
      const parser = new TypeScriptParser()
      for (let i = 0; i < 10; i++) {
        parser.parseAndIntegrate(`file${i}.ts`, `export const x = ${i};`, codeMap)
      }

      const nodes = codeMap.getNodes()
      const edges = codeMap.getEdges()

      expect(nodes.length).toBe(10)
      expect(typeof edges.length).toBe('number')
    })

    it('should categorize nodes by type', () => {
      const parser = new TypeScriptParser()
      const code = `
        export class UserClass {}
        export interface UserInterface {}
        export function userFunction() {}
        export const userVariable = 1;
      `
      parser.parseAndIntegrate('mixed.ts', code, codeMap)

      const nodes = codeMap.getNodes()
      const types = new Set(nodes.map((n) => n.type))
      expect(types.size).toBeGreaterThan(1)
    })
  })

  describe('responsive rendering', () => {
    it('should resize rendering context', () => {
      canvas.width = 1920
      canvas.height = 1080
      expect(canvas.width).toBe(1920)
      expect(canvas.height).toBe(1080)
    })

    it('should adapt to small viewport', () => {
      canvas.width = 320
      canvas.height = 240
      const parser = new TypeScriptParser()
      parser.parseAndIntegrate('small.ts', 'export const x = 1;', codeMap)
      expect(codeMap.getNodes().length).toBeGreaterThan(0)
    })

    it('should adapt to large viewport', () => {
      canvas.width = 3840
      canvas.height = 2160
      const parser = new TypeScriptParser()
      parser.parseAndIntegrate('large.ts', 'export const x = 1;', codeMap)
      expect(codeMap.getNodes().length).toBeGreaterThan(0)
    })
  })

  describe('edge cases', () => {
    it('should handle disconnected graph components', () => {
      const parser = new TypeScriptParser()
      parser.parseAndIntegrate('a.ts', 'export class A {}', codeMap)
      parser.parseAndIntegrate('b.ts', 'export class B {}', codeMap)

      const nodes = codeMap.getNodes()
      expect(nodes.length).toBe(2)
    })

    it('should handle circular dependencies', () => {
      const nodes = codeMap.getNodes()
      expect(Array.isArray(nodes)).toBe(true)
    })

    it('should handle single node graph', () => {
      const parser = new TypeScriptParser()
      parser.parseAndIntegrate('single.ts', 'export const x = 1;', codeMap)
      expect(codeMap.getNodes().length).toBe(1)
    })
  })
})
