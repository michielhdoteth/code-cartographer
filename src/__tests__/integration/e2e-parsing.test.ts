import { CodeMap } from '@models/codeMap'
import { TypeScriptParser } from '@parsers/typescript'
import { PythonParser } from '@parsers/python'
import { JavaScriptParser } from '@parsers/javascript'

describe('E2E: Parsing Workflow', () => {
  let codeMap: CodeMap

  beforeEach(() => {
    codeMap = new CodeMap('test-project', 'Test Project', '/test')
  })

  describe('single language parsing', () => {
    it('should parse TypeScript file and integrate', () => {
      const parser = new TypeScriptParser()
      const code = `
        export interface User {
          id: number;
          name: string;
          email: string;
        }

        export class UserService {
          async getUser(id: number): Promise<User> {
            return { id, name: 'John', email: 'john@example.com' };
          }

          async updateUser(id: number, data: Partial<User>): Promise<User> {
            return { ...data, id } as User;
          }
        }
      `
      parser.parseAndIntegrate('user.ts', code, codeMap)
      expect(codeMap.getNodes().length).toBeGreaterThan(0)
      expect(codeMap.getEdges().length).toBeGreaterThanOrEqual(0)
    })

    it('should parse Python file and integrate', () => {
      const parser = new PythonParser()
      const code = `
        class Calculator:
          def add(self, a, b):
              return a + b

          def multiply(self, x, y):
              return x * y

          @staticmethod
          def subtract(a, b):
              return a - b

        def process_data(data):
          calc = Calculator()
          return calc.add(1, 2)
      `
      parser.parseAndIntegrate('calculator.py', code, codeMap)
      expect(codeMap.getNodes().length).toBeGreaterThan(0)
    })

    it('should parse JavaScript file and integrate', () => {
      const parser = new JavaScriptParser()
      const code = `
        export function createUser(userData) {
          const user = {
            id: Math.random(),
            ...userData,
            createdAt: new Date()
          };
          return user;
        }

        export class UserManager {
          constructor() {
            this.users = [];
          }

          addUser(userData) {
            const user = createUser(userData);
            this.users.push(user);
            return user;
          }

          getUser(id) {
            return this.users.find(u => u.id === id);
          }
        }
      `
      parser.parseAndIntegrate('users.js', code, codeMap)
      expect(codeMap.getNodes().length).toBeGreaterThan(0)
    })
  })

  describe('multi-language parsing', () => {
    it('should parse multiple files in sequence', () => {
      const tsParser = new TypeScriptParser()
      const pyParser = new PythonParser()

      const tsCode = 'export const version = "1.0.0";'
      const pyCode = 'VERSION = "1.0.0"'

      tsParser.parseAndIntegrate('version.ts', tsCode, codeMap)
      pyParser.parseAndIntegrate('version.py', pyCode, codeMap)

      expect(codeMap.getNodes().length).toBeGreaterThan(0)
    })

    it('should handle mixed language projects', () => {
      const files = [
        { parser: new TypeScriptParser(), name: 'api.ts', code: 'export interface API { getData(): Promise<any> }' },
        { parser: new PythonParser(), name: 'utils.py', code: 'def process(data): return data' },
        { parser: new JavaScriptParser(), name: 'app.js', code: 'export const app = {}' },
      ]

      for (const file of files) {
        file.parser.parseAndIntegrate(file.name, file.code, codeMap)
      }

      expect(codeMap.getNodes().length).toBeGreaterThan(0)
    })
  })

  describe('error recovery', () => {
    it('should handle malformed code gracefully', () => {
      const parser = new TypeScriptParser()
      const malformedCode = 'function { invalid } ] syntax ['

      expect(() => {
        parser.parseAndIntegrate('bad.ts', malformedCode, codeMap)
      }).not.toThrow()
    })

    it('should continue after parsing errors', () => {
      const tsParser = new TypeScriptParser()

      // Parse good code
      tsParser.parseAndIntegrate('good.ts', 'export const x = 1;', codeMap)
      const goodNodeCount = codeMap.getNodes().length

      // Parse bad code
      tsParser.parseAndIntegrate('bad.ts', 'function { invalid [', codeMap)

      // Should still have the good nodes
      expect(codeMap.getNodes().length).toBeGreaterThanOrEqual(goodNodeCount)
    })
  })

  describe('large file parsing', () => {
    it('should parse file with 100 functions', () => {
      const parser = new JavaScriptParser()
      const code = Array.from({ length: 100 }, (_, i) => `export function func${i}() { return ${i}; }`).join('\n')

      parser.parseAndIntegrate('large.js', code, codeMap)
      expect(codeMap.getNodes().length).toBeGreaterThan(0)
    })

    it('should parse deeply nested structures', () => {
      const parser = new JavaScriptParser()
      const code = `
        function level1() {
          function level2() {
            function level3() {
              function level4() {
                return 42;
              }
              return level4;
            }
            return level3;
          }
          return level2;
        }
      `
      parser.parseAndIntegrate('nested.js', code, codeMap)
      expect(codeMap.getNodes().length).toBeGreaterThan(0)
    })
  })

  describe('metrics collection', () => {
    it('should collect metrics during parsing', () => {
      const parser = new TypeScriptParser()
      const code = `
        export function calculate(a, b, c) {
          if (a > 0) {
            if (b > 0) {
              if (c > 0) {
                return a + b + c;
              }
            }
          }
          return 0;
        }
      `
      parser.parseAndIntegrate('metrics.ts', code, codeMap)

      const nodes = codeMap.getNodes()
      for (const node of nodes) {
        expect(node.metrics).toBeDefined()
        expect(node.metrics.cyclomatic).toBeGreaterThanOrEqual(1)
        expect(node.metrics.linesOfCode).toBeGreaterThan(0)
      }
    })
  })

  describe('file tracking', () => {
    it('should track nodes by file', () => {
      const parser = new TypeScriptParser()
      parser.parseAndIntegrate('user.ts', 'export class User {}', codeMap)
      parser.parseAndIntegrate('admin.ts', 'export class Admin {}', codeMap)

      const userNodes = codeMap.getNodesByFile('user.ts')
      const adminNodes = codeMap.getNodesByFile('admin.ts')

      expect(userNodes.length).toBeGreaterThan(0)
      expect(adminNodes.length).toBeGreaterThan(0)
    })
  })

  describe('performance', () => {
    it('should parse 100 files in reasonable time', () => {
      const parser = new JavaScriptParser()
      const start = performance.now()

      for (let i = 0; i < 100; i++) {
        const code = `
          export function func${i}() { return ${i}; }
          export const var${i} = ${i};
        `
        parser.parseAndIntegrate(`file${i}.js`, code, codeMap)
      }

      const elapsed = performance.now() - start
      expect(elapsed).toBeLessThan(10000)
      expect(codeMap.getNodes().length).toBeGreaterThan(0)
    })

    it('should handle parsing with cache hits', () => {
      const parser = new JavaScriptParser()
      const code = 'export const x = 1;'

      const start1 = performance.now()
      parser.parseAndIntegrate('test1.js', code, codeMap)
      const elapsed1 = performance.now() - start1

      const start2 = performance.now()
      parser.parseAndIntegrate('test2.js', code, codeMap)
      const elapsed2 = performance.now() - start2

      expect(elapsed2).toBeLessThanOrEqual(elapsed1)
    })
  })

  describe('edge cases', () => {
    it('should handle empty files', () => {
      const parser = new TypeScriptParser()
      expect(() => {
        parser.parseAndIntegrate('empty.ts', '', codeMap)
      }).not.toThrow()
    })

    it('should handle files with only comments', () => {
      const parser = new JavaScriptParser()
      const code = '// This is a comment\n/* Multi-line comment */'
      expect(() => {
        parser.parseAndIntegrate('comments.js', code, codeMap)
      }).not.toThrow()
    })

    it('should handle unicode content', () => {
      const parser = new JavaScriptParser()
      const code = 'const emoji = "你好 مرحبا 😀";'
      expect(() => {
        parser.parseAndIntegrate('unicode.js', code, codeMap)
      }).not.toThrow()
    })
  })
})
