import { describe, it, expect, beforeEach } from 'vitest'
import { TypeScriptParser } from '@parsers/typescript'
import { PythonParser } from '@parsers/python'
import { JavaScriptParser } from '@parsers/javascript'
import { CodeMap } from '@models/codeMap'
import { codeSamples, generateLargeFunctionList, generateLargeClassList } from '../fixtures/index'
import { createTestCodeMap, measurePerformance, logBenchmark, logGoal } from '../helpers/index'

describe('Performance Benchmarks: Parsers', () => {
  let codeMap: CodeMap

  beforeEach(() => {
    codeMap = createTestCodeMap('bench', 'Benchmark')
  })

  describe('TypeScript Parser Benchmarks', () => {
    const parser = new TypeScriptParser()

    it('should parse 100 small TypeScript files in <5 seconds', () => {
      const { elapsed } = measurePerformance(() => {
        for (let i = 0; i < 100; i++) {
          parser.parseAndIntegrate(`test${i}.ts`, codeSamples.typescript.interfaceAndClass, codeMap)
        }
      })
      logBenchmark('100 TS files', elapsed, 5000)
      expect(elapsed).toBeLessThan(5000)
    })

    it('should parse single large TypeScript file in <1 second', () => {
      const code = generateLargeClassList(500)
      const { elapsed } = measurePerformance(() => {
        parser.parseAndIntegrate('large.ts', code, codeMap)
      })
      logBenchmark('Large TS file (500 types)', elapsed, 1000)
      expect(elapsed).toBeLessThan(1000)
    })

    it('should parse deeply nested TypeScript structure', () => {
      const { elapsed } = measurePerformance(() => {
        parser.parseAndIntegrate('deep.ts', codeSamples.typescript.deeplyNested, codeMap)
      })
      logBenchmark('Deeply nested structure', elapsed, 500)
      expect(elapsed).toBeLessThan(500)
    })
  })

  describe('Python Parser Benchmarks', () => {
    const parser = new PythonParser()

    it('should parse 100 small Python files in <5 seconds', () => {
      const { elapsed } = measurePerformance(() => {
        for (let i = 0; i < 100; i++) {
          parser.parseAndIntegrate(`test${i}.py`, codeSamples.python.classDef, codeMap)
        }
      })
      logBenchmark('100 Python files', elapsed, 5000)
      expect(elapsed).toBeLessThan(5000)
    })

    it('should parse Python file with many decorators', () => {
      const decorators = Array.from({ length: 100 }, (_, i) => `@decorator${i}`).join('\n')
      const code = `${decorators}\ndef decorated_function():\n    return 42`
      const { elapsed } = measurePerformance(() => {
        parser.parseAndIntegrate('decorated.py', code, codeMap)
      })
      logBenchmark('Heavily decorated Python', elapsed, 1000)
      expect(elapsed).toBeLessThan(1000)
    })
  })

  describe('JavaScript Parser Benchmarks', () => {
    const parser = new JavaScriptParser()

    it('should parse 100 small JavaScript files in <3 seconds', () => {
      const { elapsed } = measurePerformance(() => {
        for (let i = 0; i < 100; i++) {
          parser.parseAndIntegrate(`test${i}.js`, codeSamples.javascript.classDeclaration, codeMap)
        }
      })
      logBenchmark('100 JavaScript files', elapsed, 3000)
      expect(elapsed).toBeLessThan(3000)
    })

    it('should parse large JavaScript file with many functions', () => {
      const code = generateLargeFunctionList(1000, 'javascript')
      const { elapsed } = measurePerformance(() => {
        parser.parseAndIntegrate('large.js', code, codeMap)
      })
      logBenchmark('Large JS file (1000 functions)', elapsed, 2000)
      expect(elapsed).toBeLessThan(2000)
    })

    it('should parse JavaScript with complex AST', () => {
      const code = `
        export class ComplexClass {
          constructor(a, b, c, d, e) { this.a = a; this.b = b; this.c = c; }
          method1(x, y) { if (x > 0) { if (y > 0) { return x + y; } } return 0; }
          method2(data) { return data.map(d => d * 2).filter(d => d > 0).reduce((a, b) => a + b, 0); }
          async asyncMethod() { const result = await fetch('/api'); return result.json(); }
        }
      `
      const { elapsed } = measurePerformance(() => {
        parser.parseAndIntegrate('complex.js', code, codeMap)
      })
      logBenchmark('Complex JavaScript AST', elapsed, 1000)
      expect(elapsed).toBeLessThan(1000)
    })
  })

  describe('Cross-language parsing performance', () => {
    it('should parse mixed language codebase efficiently', () => {
      const tsParser = new TypeScriptParser()
      const pyParser = new PythonParser()
      const jsParser = new JavaScriptParser()

      const { elapsed } = measurePerformance(() => {
        for (let i = 0; i < 50; i++) {
          tsParser.parseAndIntegrate(`ts${i}.ts`, codeSamples.typescript.simple, codeMap)
          pyParser.parseAndIntegrate(`py${i}.py`, codeSamples.python.functionDef, codeMap)
          jsParser.parseAndIntegrate(`js${i}.js`, codeSamples.javascript.simple, codeMap)
        }
      })
      logBenchmark('Mixed language 150 files', elapsed, 10000)
      expect(elapsed).toBeLessThan(10000)
    })
  })

  describe('Cache effectiveness', () => {
    it('should show cache speedup on identical content', () => {
      const parser = new JavaScriptParser()
      const code = codeSamples.javascript.simple

      const { elapsed: time1 } = measurePerformance(() => {
        parser.parseAndIntegrate('test1.js', code, codeMap)
      })

      const { elapsed: time2 } = measurePerformance(() => {
        parser.parseAndIntegrate('test2.js', code, codeMap)
      })

      console.log(`  First parse: ${time1.toFixed(2)}ms, Second parse: ${time2.toFixed(2)}ms`)
      expect(time2).toBeLessThanOrEqual(time1)
    })
  })

  describe('Parser stress tests', () => {
    it('should parse 1000 small files in <30 seconds', () => {
      const parser = new JavaScriptParser()
      const { elapsed } = measurePerformance(() => {
        for (let i = 0; i < 1000; i++) {
          parser.parseAndIntegrate(`file${i}.js`, `export const v${i} = ${i};`, codeMap)
        }
      })
      const avgPerFile = elapsed / 1000
      console.log(`  1000 files total: ${elapsed.toFixed(2)}ms, avg: ${avgPerFile.toFixed(3)}ms/file`)
      expect(elapsed).toBeLessThan(30000)
      expect(avgPerFile).toBeLessThan(50)
    })

    it('should parse file up to 100KB in <2 seconds', () => {
      const parser = new TypeScriptParser()
      const code = generateLargeFunctionList(5000)
      const { elapsed } = measurePerformance(() => {
        parser.parseAndIntegrate('huge.ts', code, codeMap)
      })
      logBenchmark('100KB file (~5000 functions)', elapsed, 2000)
      expect(elapsed).toBeLessThan(2000)
    })
  })

  describe('Parser target goals', () => {
    it('should parse 1000 files in <10 seconds', () => {
      const parser = new TypeScriptParser()
      const { elapsed } = measurePerformance(() => {
        for (let i = 0; i < 1000; i++) {
          parser.parseAndIntegrate(`f${i}.ts`, `export const v${i} = ${i};`, codeMap)
        }
      })
      logGoal('Parse 1000 files', elapsed, 10000)
      expect(elapsed).toBeLessThan(10000)
    })

    it('should parse 1000 cached files in <500ms', () => {
      const parser = new TypeScriptParser()
      const code = codeSamples.typescript.simple

      // Warm up cache
      for (let i = 0; i < 100; i++) {
        parser.parseAndIntegrate(`file${i}.ts`, code, codeMap)
      }

      const { elapsed } = measurePerformance(() => {
        for (let i = 100; i < 1000; i++) {
          parser.parseAndIntegrate(`file${i}.ts`, code, codeMap)
        }
      })
      logGoal('Parse 1000 cached files', elapsed, 500)
      expect(elapsed).toBeLessThan(500)
    })
  })
})
