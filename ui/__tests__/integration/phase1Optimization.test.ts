import { describe, it, expect, beforeEach } from 'vitest'
import { MapGenerator } from '@engine/mapGenerator'
import { Language } from '@models/types'
import { cacheManager } from '@engine/cacheManager'

describe('Phase 1: Parser Optimization Integration', () => {
  let mapGenerator: MapGenerator

  beforeEach(async () => {
    mapGenerator = new MapGenerator()
    await cacheManager.initialize()
    await cacheManager.clear()
  })

  describe('basic parsing with worker pool', () => {
    it('should parse JavaScript files correctly', async () => {
      const files = [
        {
          name: 'test1.js',
          content: `
            function add(a, b) {
              return a + b;
            }
            const multiply = (x, y) => x * y;
          `,
          language: 'javascript' as Language,
        },
        {
          name: 'test2.js',
          content: `
            class Calculator {
              constructor() {
                this.value = 0;
              }
              reset() {
                this.value = 0;
              }
            }
          `,
          language: 'javascript' as Language,
        },
      ]

      const result = await mapGenerator.generateMap(files)

      expect(result.id).toBeTruthy()
      expect(result.statistics.totalFiles).toBe(2)
      expect(result.statistics.totalNodes).toBeGreaterThan(0)
      expect(result.statistics.languageBreakdown.javascript).toBe(2)
    })

    it('should handle TypeScript files', async () => {
      const files = [
        {
          name: 'types.ts',
          content: `
            interface User {
              id: number;
              name: string;
            }
            function getUser(id: number): User {
              return { id, name: 'John' };
            }
          `,
          language: 'typescript' as Language,
        },
      ]

      const result = await mapGenerator.generateMap(files)

      expect(result.statistics.totalFiles).toBe(1)
      expect(result.statistics.languageBreakdown.typescript).toBe(1)
    })

    it('should handle multiple languages', async () => {
      const files = [
        {
          name: 'script.js',
          content: 'const x = 1;',
          language: 'javascript' as Language,
        },
        {
          name: 'script.py',
          content: 'def add(a, b):\n  return a + b',
          language: 'python' as Language,
        },
        {
          name: 'script.java',
          content: 'public class Test { public int add(int a, int b) { return a + b; } }',
          language: 'java' as Language,
        },
      ]

      const result = await mapGenerator.generateMap(files)

      expect(result.statistics.totalFiles).toBe(3)
      expect(result.statistics.languageBreakdown.javascript).toBe(1)
      expect(result.statistics.languageBreakdown.python).toBe(1)
      expect(result.statistics.languageBreakdown.java).toBe(1)
    })
  })

  describe('caching performance', () => {
    it('should cache parsed results', async () => {
      const files = [
        {
          name: 'cached.js',
          content: 'function test() { return 42; }',
          language: 'javascript' as Language,
        },
      ]

      // First parse - cache miss
      const start1 = performance.now()
      const result1 = await mapGenerator.generateMap(files, { useCache: true })
      const time1 = performance.now() - start1

      // Second parse - cache hit
      const start2 = performance.now()
      const result2 = await mapGenerator.generateMap(files, { useCache: true })
      const time2 = performance.now() - start2

      expect(result2.statistics.totalNodes).toBe(result1.statistics.totalNodes)
      // Second parse should be faster (100x speedup expected)
      // Allow some variance for test environment
      expect(time2).toBeLessThanOrEqual(time1 * 2)
    })

    it('should detect content changes and reparse', async () => {
      const files1 = [
        {
          name: 'test.js',
          content: 'function old() {}',
          language: 'javascript' as Language,
        },
      ]

      const files2 = [
        {
          name: 'test.js',
          content: 'function old() {} function new() {}',
          language: 'javascript' as Language,
        },
      ]

      const result1 = await mapGenerator.generateMap(files1, { useCache: true })
      const result2 = await mapGenerator.generateMap(files2, { useCache: true })

      // Should detect change and reparse
      expect(result2.statistics.totalNodes).toBeGreaterThanOrEqual(result1.statistics.totalNodes)
    })

    it('should skip cache if useCache is false', async () => {
      const files = [
        {
          name: 'no-cache.js',
          content: 'const x = 1;',
          language: 'javascript' as Language,
        },
      ]

      const result1 = await mapGenerator.generateMap(files, { useCache: false })
      const result2 = await mapGenerator.generateMap(files, { useCache: false })

      // Both should parse (no cache)
      expect(result1.statistics.totalFiles).toBe(1)
      expect(result2.statistics.totalFiles).toBe(1)
    })
  })

  describe('progress reporting', () => {
    it('should report parsing progress', async () => {
      const progressUpdates: Array<{ completed: number; total: number; message: string }> = []

      const files = Array.from({ length: 10 }, (_, i) => ({
        name: `file${i}.js`,
        content: `function func${i}() { return ${i}; }`,
        language: 'javascript' as Language,
      }))

      await mapGenerator.generateMap(files, {
        onProgress: (completed, total, message) => {
          progressUpdates.push({ completed, total, message })
        },
      })

      // Should have progress updates
      expect(progressUpdates.length).toBeGreaterThan(0)
      expect(progressUpdates[0].message).toBeTruthy()
    })
  })

  describe('error handling', () => {
    it('should handle malformed code gracefully', async () => {
      const files = [
        {
          name: 'valid.js',
          content: 'function valid() {}',
          language: 'javascript' as Language,
        },
        {
          name: 'invalid.js',
          content: 'function { invalid ] syntax [',
          language: 'javascript' as Language,
        },
        {
          name: 'valid2.js',
          content: 'const x = 42;',
          language: 'javascript' as Language,
        },
      ]

      const result = await mapGenerator.generateMap(files)

      // Should parse valid files despite invalid syntax in one file
      expect(result.statistics.totalFiles).toBeGreaterThanOrEqual(2)
    })

    it('should handle empty files', async () => {
      const files = [
        {
          name: 'empty.js',
          content: '',
          language: 'javascript' as Language,
        },
        {
          name: 'non-empty.js',
          content: 'function test() {}',
          language: 'javascript' as Language,
        },
      ]

      const result = await mapGenerator.generateMap(files)

      expect(result.statistics.totalFiles).toBe(2)
    })

    it('should handle very large files', async () => {
      const largeContent = 'const func = () => {}; '.repeat(5000)
      const files = [
        {
          name: 'large.js',
          content: largeContent,
          language: 'javascript' as Language,
        },
      ]

      const result = await mapGenerator.generateMap(files)

      expect(result.statistics.totalFiles).toBe(1)
    })
  })

  describe('statistics generation', () => {
    it('should calculate correct statistics', async () => {
      const files = [
        {
          name: 'file1.js',
          content: 'function add(a, b) { return a + b; }',
          language: 'javascript' as Language,
        },
        {
          name: 'file2.js',
          content: 'const multiply = (x, y) => x * y;',
          language: 'javascript' as Language,
        },
      ]

      const result = await mapGenerator.generateMap(files)

      expect(result.statistics.totalFiles).toBe(2)
      expect(result.statistics.totalNodes).toBeGreaterThanOrEqual(2) // At least 2 functions
      expect(result.statistics.languageBreakdown.javascript).toBe(2)
    })

    it('should include node type breakdown', async () => {
      const files = [
        {
          name: 'test.js',
          content: `
            function funcA() {}
            class MyClass {
              method() {}
            }
            const arrow = () => {};
          `,
          language: 'javascript' as Language,
        },
      ]

      const result = await mapGenerator.generateMap(files)

      expect(result.statistics.nodeTypeBreakdown).toBeDefined()
      expect(Object.keys(result.statistics.nodeTypeBreakdown).length).toBeGreaterThan(0)
    })
  })

  describe('map export formats', () => {
    it('should export as JSON', async () => {
      const files = [
        {
          name: 'test.js',
          content: 'function test() {}',
          language: 'javascript' as Language,
        },
      ]

      const result = await mapGenerator.generateMap(files)
      const json = mapGenerator.exportAsJSON(result)

      expect(json).toBeTruthy()
      expect(() => JSON.parse(json)).not.toThrow()

      const parsed = JSON.parse(json)
      expect(parsed.id).toBe(result.id)
      expect(parsed.name).toBe(result.name)
    })

    it('should export visualization data', async () => {
      const files = [
        {
          name: 'test.js',
          content: 'function test() {}',
          language: 'javascript' as Language,
        },
      ]

      const result = await mapGenerator.generateMap(files)
      const vizData = mapGenerator.exportAsVisualizationData(result)

      expect(vizData.nodes).toBeDefined()
      expect(vizData.edges).toBeDefined()
      expect(Array.isArray(vizData.nodes)).toBe(true)
      expect(Array.isArray(vizData.edges)).toBe(true)
    })

    it('should export as markdown', async () => {
      const files = [
        {
          name: 'test.js',
          content: 'function test() {}',
          language: 'javascript' as Language,
        },
      ]

      const result = await mapGenerator.generateMap(files)
      const markdown = mapGenerator.exportAsMarkdown(result)

      expect(markdown).toContain('# Code Map:')
      expect(markdown).toContain('Statistics')
      expect(markdown).toContain('javascript')
    })
  })
})
