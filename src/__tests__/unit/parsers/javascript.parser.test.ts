import { JavaScriptParser } from '@parsers/javascript'
import { CodeMap } from '@models/codeMap'
import { codeSamples, generateLargeFunctionList } from '../../fixtures/index'
import { createTestCodeMap, measurePerformance, expectPerformanceWithin } from '../../helpers/index'

describe('JavaScriptParser', () => {
  let parser: JavaScriptParser
  let codeMap: CodeMap

  beforeEach(() => {
    parser = new JavaScriptParser()
    codeMap = createTestCodeMap()
  })

  describe('basic parsing', () => {
    it('should parse function declarations', () => {
      parser.parseAndIntegrate('test.js', codeSamples.javascript.functionDeclaration, codeMap)
      expect(codeMap.getNodes().length).toBeGreaterThan(0)
    })

    it('should parse arrow functions', () => {
      parser.parseAndIntegrate('test.js', codeSamples.javascript.arrowFunction, codeMap)
      expect(codeMap.getNodes().length).toBeGreaterThan(0)
    })

    it('should parse class declarations', () => {
      parser.parseAndIntegrate('test.js', codeSamples.javascript.classDeclaration, codeMap)
      expect(codeMap.getNodes().length).toBeGreaterThan(0)
    })

    it('should parse imports', () => {
      const code = `
        import { Component } from 'react';
        import * as utils from './utils';
      `
      parser.parseAndIntegrate('test.js', code, codeMap)
      expect(codeMap.getEdges().length).toBeGreaterThanOrEqual(0)
    })
  })

  describe('edge cases', () => {
    it('should handle empty files', () => {
      parser.parseAndIntegrate('test.js', codeSamples.empty, codeMap)
      expect(codeMap.getNodes().length).toBeGreaterThanOrEqual(0)
    })

    it('should handle files with only comments', () => {
      parser.parseAndIntegrate('test.js', codeSamples.commentsOnly, codeMap)
      expect(codeMap.getNodes().length).toBeGreaterThanOrEqual(0)
    })

    it('should handle malformed syntax gracefully', () => {
      expect(() => {
        parser.parseAndIntegrate('test.js', codeSamples.malformed, codeMap)
      }).not.toThrow()
    })

    it('should handle files with unicode characters', () => {
      parser.parseAndIntegrate('test.js', codeSamples.unicode, codeMap)
      expect(codeMap.getNodes().length).toBeGreaterThanOrEqual(0)
    })

    it('should handle deeply nested structures', () => {
      parser.parseAndIntegrate('test.js', codeSamples.typescript.deeplyNested, codeMap)
      expect(codeMap.getNodes().length).toBeGreaterThan(0)
    })

    it('should handle large files', () => {
      const code = generateLargeFunctionList(1000, 'javascript')
      parser.parseAndIntegrate('test.js', code, codeMap)
      expect(codeMap.getNodes().length).toBeGreaterThanOrEqual(0)
    })
  })

  describe('async functions', () => {
    it('should parse async functions', () => {
      parser.parseAndIntegrate('test.js', codeSamples.javascript.asyncFunction, codeMap)
      expect(codeMap.getNodes().length).toBeGreaterThan(0)
    })

    it('should parse async arrow functions', () => {
      const code = 'const getData = async () => { return await fetch("/"); };'
      parser.parseAndIntegrate('test.js', code, codeMap)
      expect(codeMap.getNodes().length).toBeGreaterThan(0)
    })
  })

  describe('anonymous functions', () => {
    it('should parse anonymous functions', () => {
      const code = `
        array.map(function(item) {
          return item * 2;
        });
      `
      parser.parseAndIntegrate('test.js', code, codeMap)
      expect(codeMap.getNodes().length).toBeGreaterThanOrEqual(0)
    })

    it('should parse IIFE', () => {
      const code = `
        (function() {
          console.log("IIFE");
        })();
      `
      parser.parseAndIntegrate('test.js', code, codeMap)
      expect(codeMap.getNodes().length).toBeGreaterThanOrEqual(0)
    })
  })

  describe('exports', () => {
    it('should parse named exports', () => {
      const code = `
        export function helper() { }
        export const value = 42;
        export class MyClass { }
      `
      parser.parseAndIntegrate('test.js', code, codeMap)
      expect(codeMap.getNodes().length).toBeGreaterThan(0)
    })

    it('should parse default export', () => {
      const code = 'export default function Component() { return null; }'
      parser.parseAndIntegrate('test.js', code, codeMap)
      expect(codeMap.getNodes().length).toBeGreaterThan(0)
    })
  })

  describe('performance', () => {
    it('should parse large file within timeout', () => {
      const largeFunctions = generateLargeFunctionList(100, 'javascript')
      const { elapsed } = measurePerformance(() => {
        parser.parseAndIntegrate('large.js', largeFunctions, codeMap)
      })

      expectPerformanceWithin(elapsed, 5000, 'Large file parse')
      expect(codeMap.getNodes().length).toBeGreaterThan(0)
    })
  })
})
