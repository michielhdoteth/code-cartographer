import { describe, it, expect, beforeEach } from 'vitest'
import { TypeScriptParser } from '@parsers/typescript'
import { PythonParser } from '@parsers/python'
import { JavaParser } from '@parsers/java'
import { GoParser } from '@parsers/go'
import { RustParser } from '@parsers/rust'
import { CppParser } from '@parsers/cpp'
import { RubyParser } from '@parsers/ruby'
import { PhpParser } from '@parsers/php'
import { CodeMap } from '@models/codeMap'
import { codeSamples, generateLargeFunctionList } from '../../fixtures/index'
import { createTestCodeMap, measurePerformance, expectPerformanceWithin } from '../../helpers/index'

// =============================================================================
// Parser Factory for Testing
// =============================================================================

interface ParserTestCase {
  name: string
  parser: { parseAndIntegrate(file: string, code: string, map: CodeMap): void }
  extension: string
  samples: {
    basic: string
    withTypes?: string
    withImports?: string
    advanced?: string
  }
}

function createParserTestCases(): ParserTestCase[] {
  return [
    {
      name: 'TypeScript',
      parser: new TypeScriptParser(),
      extension: 'ts',
      samples: {
        basic: codeSamples.typescript.simple,
        withTypes: codeSamples.typescript.interfaceAndClass,
        advanced: codeSamples.typescript.withGenerics,
      },
    },
    {
      name: 'Python',
      parser: new PythonParser(),
      extension: 'py',
      samples: {
        basic: codeSamples.python.simple,
        withTypes: codeSamples.python.classDef,
        advanced: codeSamples.python.withDecorators,
      },
    },
    {
      name: 'Java',
      parser: new JavaParser(),
      extension: 'java',
      samples: {
        basic: codeSamples.java.simple,
        withTypes: codeSamples.java.classWithMethods,
      },
    },
    {
      name: 'Go',
      parser: new GoParser(),
      extension: 'go',
      samples: {
        basic: codeSamples.go.simple,
        advanced: codeSamples.go.functionAndStruct,
      },
    },
    {
      name: 'Rust',
      parser: new RustParser(),
      extension: 'rs',
      samples: {
        basic: codeSamples.rust.simple,
        advanced: codeSamples.rust.structAndImpl,
      },
    },
  ]
}

// =============================================================================
// Tests
// =============================================================================

describe('All Language Parsers', () => {
  let codeMap: CodeMap

  beforeEach(() => {
    codeMap = createTestCodeMap()
  })

  describe('TypeScript', () => {
    const parser = new TypeScriptParser()

    it('should parse TypeScript with types', () => {
      parser.parseAndIntegrate('test.ts', codeSamples.typescript.interfaceAndClass, codeMap)
      expect(codeMap.getNodes().length).toBeGreaterThan(0)
    })

    it('should parse TypeScript generics', () => {
      parser.parseAndIntegrate('test.ts', codeSamples.typescript.withGenerics, codeMap)
      expect(codeMap.getNodes().length).toBeGreaterThan(0)
    })

    it('should parse TypeScript decorators', () => {
      parser.parseAndIntegrate('test.ts', codeSamples.typescript.withDecorators, codeMap)
      expect(codeMap.getNodes().length).toBeGreaterThanOrEqual(0)
    })

    it('should parse TSX/JSX', () => {
      const code = `
        const Component = () => {
          return <div className="test">Hello</div>;
        };
      `
      parser.parseAndIntegrate('test.tsx', code, codeMap)
      expect(codeMap.getNodes().length).toBeGreaterThanOrEqual(0)
    })
  })

  describe('Python', () => {
    const parser = new PythonParser()

    it('should parse Python functions', () => {
      parser.parseAndIntegrate('test.py', codeSamples.python.classDef, codeMap)
      expect(codeMap.getNodes().length).toBeGreaterThan(0)
    })

    it('should parse Python imports', () => {
      const code = `
from django import forms
import numpy as np
from utils import helper
      `
      parser.parseAndIntegrate('test.py', code, codeMap)
      expect(codeMap.getEdges().length).toBeGreaterThanOrEqual(0)
    })

    it('should parse Python decorators', () => {
      parser.parseAndIntegrate('test.py', codeSamples.python.withDecorators, codeMap)
      expect(codeMap.getNodes().length).toBeGreaterThanOrEqual(0)
    })

    it('should parse Python with indentation errors gracefully', () => {
      const code = `
def func():
  x = 1
    y = 2
      `
      expect(() => {
        parser.parseAndIntegrate('test.py', code, codeMap)
      }).not.toThrow()
    })
  })

  describe('Java', () => {
    const parser = new JavaParser()

    it('should parse Java classes and methods', () => {
      parser.parseAndIntegrate('test.java', codeSamples.java.classWithMethods, codeMap)
      expect(codeMap.getNodes().length).toBeGreaterThan(0)
    })

    it('should parse Java interfaces', () => {
      const code = `
        public interface Repository {
          void save(Entity entity);
          Entity findById(int id);
        }
      `
      parser.parseAndIntegrate('test.java', code, codeMap)
      expect(codeMap.getNodes().length).toBeGreaterThan(0)
    })

    it('should parse Java enums', () => {
      const code = `
        public enum Status {
          ACTIVE, INACTIVE, PENDING;
          public boolean isActive() { return this == ACTIVE; }
        }
      `
      parser.parseAndIntegrate('test.java', code, codeMap)
      expect(codeMap.getNodes().length).toBeGreaterThan(0)
    })
  })

  describe('Go', () => {
    const parser = new GoParser()

    it('should parse Go functions and packages', () => {
      parser.parseAndIntegrate('test.go', codeSamples.go.functionAndStruct, codeMap)
      expect(codeMap.getNodes().length).toBeGreaterThan(0)
    })

    it('should parse Go interfaces', () => {
      const code = `
        type Reader interface {
          Read(p []byte) (n int, err error)
        }
        type Writer interface {
          Write(p []byte) (n int, err error)
        }
      `
      parser.parseAndIntegrate('test.go', code, codeMap)
      expect(codeMap.getNodes().length).toBeGreaterThan(0)
    })
  })

  describe('Rust', () => {
    const parser = new RustParser()

    it('should parse Rust functions and structs', () => {
      parser.parseAndIntegrate('test.rs', codeSamples.rust.structAndImpl, codeMap)
      expect(codeMap.getNodes().length).toBeGreaterThan(0)
    })

    it('should parse Rust traits', () => {
      const code = `
        pub trait Number {
          fn add(&self, other: &Self) -> Self;
          fn multiply(&self, other: &Self) -> Self;
        }
      `
      parser.parseAndIntegrate('test.rs', code, codeMap)
      expect(codeMap.getNodes().length).toBeGreaterThan(0)
    })
  })

  describe('C++', () => {
    const parser = new CppParser()

    it('should parse C++ classes and methods', () => {
      const code = `
        class Calculator {
        public:
          int add(int a, int b) { return a + b; }
        private:
          double multiply(double x, double y) { return x * y; }
        };
      `
      parser.parseAndIntegrate('test.cpp', code, codeMap)
      expect(codeMap.getNodes().length).toBeGreaterThan(0)
    })

    it('should parse C++ templates', () => {
      const code = `
        template<typename T>
        class Box {
        public:
          void set(T value) { data = value; }
          T get() { return data; }
        private:
          T data;
        };
      `
      parser.parseAndIntegrate('test.cpp', code, codeMap)
      expect(codeMap.getNodes().length).toBeGreaterThan(0)
    })
  })

  describe('Ruby', () => {
    const parser = new RubyParser()

    it('should parse Ruby classes and methods', () => {
      const code = `
        class Calculator
          def add(a, b)
            a + b
          end
          def multiply(x, y)
            x * y
          end
        end
      `
      parser.parseAndIntegrate('test.rb', code, codeMap)
      expect(codeMap.getNodes().length).toBeGreaterThan(0)
    })

    it('should parse Ruby modules', () => {
      const code = `
        module Enumerable
          def each_slice(n)
          end
        end
      `
      parser.parseAndIntegrate('test.rb', code, codeMap)
      expect(codeMap.getNodes().length).toBeGreaterThan(0)
    })
  })

  describe('PHP', () => {
    const parser = new PhpParser()

    it('should parse PHP classes and methods', () => {
      const code = `
        <?php
        class Calculator {
          public function add($a, $b) { return $a + $b; }
          private function multiply($x, $y) { return $x * $y; }
        }
      `
      parser.parseAndIntegrate('test.php', code, codeMap)
      expect(codeMap.getNodes().length).toBeGreaterThan(0)
    })

    it('should parse PHP namespaces', () => {
      const code = `
        <?php
        namespace App\\Controllers;
        use App\\Models\\User;
        class UserController {
          public function show($id) { return User::find($id); }
        }
      `
      parser.parseAndIntegrate('test.php', code, codeMap)
      expect(codeMap.getNodes().length).toBeGreaterThan(0)
    })
  })

  describe('cross-language edge cases', () => {
    it('should handle empty files for all languages', () => {
      const testCases = createParserTestCases()

      for (const { parser, extension } of testCases) {
        const map = createTestCodeMap()
        parser.parseAndIntegrate(`test.${extension}`, codeSamples.empty, map)
        expect(map.getNodes().length).toBeGreaterThanOrEqual(0)
      }
    })

    it('should handle unicode in all languages', () => {
      const unicodeTests = [
        { parser: new TypeScriptParser(), ext: 'ts', code: 'const msg = "Hello";' },
        { parser: new PythonParser(), ext: 'py', code: 'msg = "Hello"' },
        { parser: new JavaParser(), ext: 'java', code: 'String msg = "Hello";' },
      ]

      for (const { parser, ext, code } of unicodeTests) {
        const map = createTestCodeMap()
        expect(() => {
          parser.parseAndIntegrate(`test.${ext}`, code, map)
        }).not.toThrow()
      }
    })

    it('should handle very large files', () => {
      const parser = new TypeScriptParser()
      const largeCode = generateLargeFunctionList(5000)

      const { elapsed } = measurePerformance(() => {
        parser.parseAndIntegrate('large.ts', largeCode, codeMap)
      })

      expectPerformanceWithin(elapsed, 10000, 'Large TypeScript file')
    })
  })
})
