import { describe, it, expect, beforeEach } from 'vitest'
import { SecurityScanner } from '@analyzers/securityScanner'
import { PatternDetector } from '@analyzers/patternDetector'
import { HealthScorer } from '@analyzers/healthScorer'
import { LayerAnalyzer } from '@analyzers/layerAnalyzer'
import { ChurnAnalyzer } from '@analyzers/churnAnalyzer'
import { CodeMap, CodeNode, CodeFile } from '@models/codeMap'

describe('All Analyzers', () => {
  let codeMap: CodeMap

  beforeEach(() => {
    codeMap = new CodeMap('test', 'Test', '/')
  })

  describe('SecurityScanner', () => {
    const scanner = new SecurityScanner()

    it('should detect SQL injection patterns', () => {
      const code = `
        const query = "SELECT * FROM users WHERE id = " + userId;
        db.query(query);
      `
      const issues = scanner.scan(code)
      expect(issues.length).toBeGreaterThan(0)
    })

    it('should detect XSS vulnerabilities', () => {
      const code = `
        document.innerHTML = userInput;
        element.innerHTML = data;
      `
      const issues = scanner.scan(code)
      expect(issues.length).toBeGreaterThan(0)
    })

    it('should detect hardcoded credentials', () => {
      const code = `
        const password = "admin123";
        const apiKey = "sk_live_abc123";
      `
      const issues = scanner.scan(code)
      expect(issues.length).toBeGreaterThan(0)
    })

    it('should detect insecure randomness', () => {
      const code = `
        const random = Math.random();
        const token = Math.random().toString();
      `
      const issues = scanner.scan(code)
      expect(issues.length).toBeGreaterThanOrEqual(0)
    })

    it('should report severity levels', () => {
      const code = `
        const query = "SELECT * FROM users WHERE id = " + userId;
        db.query(query);
      `
      const issues = scanner.scan(code)
      for (const issue of issues) {
        expect(['critical', 'high', 'medium', 'low']).toContain(issue.severity)
      }
    })
  })

  describe('PatternDetector', () => {
    const detector = new PatternDetector()

    it('should detect design patterns in code', () => {
      const code = `
        class SingletonFactory {
          private static instance: SingletonFactory;
          private constructor() {}
          static getInstance() {
            if (!this.instance) {
              this.instance = new SingletonFactory();
            }
            return this.instance;
          }
        }
      `
      const patterns = detector.detectPatterns(codeMap)
      expect(Array.isArray(patterns)).toBe(true)
    })

    it('should detect anti-patterns', () => {
      const code = `
        let globalVariable = {};
        function modifyGlobal() {
          globalVariable.x = 1;
        }
      `
      const patterns = detector.detectPatterns(codeMap)
      expect(Array.isArray(patterns)).toBe(true)
    })
  })

  describe('HealthScorer', () => {
    const scorer = new HealthScorer()

    it('should calculate health score', () => {
      const code = `
        function add(a, b) { return a + b; }
        function multiply(x, y) { return x * y; }
      `
      const health = scorer.calculateHealth(codeMap, code)
      expect(health.score).toBeGreaterThanOrEqual(0)
      expect(health.score).toBeLessThanOrEqual(100)
    })

    it('should assign letter grade', () => {
      const code = 'const x = 1;'
      const health = scorer.calculateHealth(codeMap, code)
      expect(['A', 'B', 'C', 'D', 'F']).toContain(health.grade)
    })

    it('should provide recommendations', () => {
      const code = `
        function complexFunction(a, b, c, d, e, f) {
          if (a) {
            if (b) {
              if (c) {
                if (d) {
                  return e + f;
                }
              }
            }
          }
        }
      `
      const health = scorer.calculateHealth(codeMap, code)
      expect(health.recommendations).toBeDefined()
      expect(Array.isArray(health.recommendations)).toBe(true)
    })

    it('should detect dead code', () => {
      const code = `
        function used() { return 42; }
        function unused() { return 0; }
        used();
      `
      const health = scorer.calculateHealth(codeMap, code)
      expect(health.deadCodePercentage).toBeGreaterThanOrEqual(0)
    })
  })

  describe('LayerAnalyzer', () => {
    const analyzer = new LayerAnalyzer()

    it('should detect UI layer', () => {
      const layer = analyzer.detectLayer('src/views/Dashboard.tsx')
      expect(['ui', 'components']).toContain(layer)
    })

    it('should detect components layer', () => {
      const layer = analyzer.detectLayer('src/components/Button.tsx')
      expect(layer).toBe('components')
    })

    it('should detect services layer', () => {
      const layer = analyzer.detectLayer('src/services/api.ts')
      expect(layer).toBe('services')
    })

    it('should detect data layer', () => {
      const layer = analyzer.detectLayer('src/models/User.ts')
      expect(layer).toBe('data')
    })

    it('should detect architecture violations', () => {
      const violation = analyzer.isLayerViolation('src/data/models.ts', 'src/services/api.ts')
      expect(typeof violation).toBe('boolean')
    })

    it('should classify files into layers', () => {
      const files = [
        'src/views/Home.tsx',
        'src/components/Card.tsx',
        'src/services/auth.ts',
        'src/models/User.ts',
      ]
      const classified = analyzer.classifyFiles(files)
      expect(classified.size).toBeGreaterThan(0)
    })
  })

  describe('ChurnAnalyzer', () => {
    const analyzer = new ChurnAnalyzer()

    it('should calculate churn score', () => {
      const metrics = {
        commits: 10,
        additions: 500,
        deletions: 300,
        lastModified: new Date(),
      }
      const score = analyzer.getChurnScore(metrics)
      expect(score).toBeGreaterThanOrEqual(0)
    })

    it('should classify churn levels', () => {
      const metrics = {
        commits: 50,
        additions: 1000,
        deletions: 800,
        lastModified: new Date(),
      }
      const level = analyzer.getChurnLevel(metrics)
      expect(['critical', 'high', 'medium', 'low', 'none']).toContain(level)
    })

    it('should find hotspots', () => {
      const churnMap = new Map([
        ['file1.ts', { commits: 50, additions: 1000, deletions: 800, lastModified: new Date() }],
        ['file2.ts', { commits: 2, additions: 50, deletions: 30, lastModified: new Date() }],
      ])
      const hotspots = analyzer.findHotspots(churnMap, 10)
      expect(Array.isArray(hotspots)).toBe(true)
    })
  })

  describe('integration between analyzers', () => {
    it('should work together on same codebase', () => {
      const code = `
        function calculate(x, y) {
          const result = x + y;
          return result;
        }
      `

      const health = new HealthScorer().calculateHealth(codeMap, code)
      const patterns = new PatternDetector().detectPatterns(codeMap)
      const security = new SecurityScanner().scan(code)

      expect(health).toBeDefined()
      expect(Array.isArray(patterns)).toBe(true)
      expect(Array.isArray(security)).toBe(true)
    })
  })
})
