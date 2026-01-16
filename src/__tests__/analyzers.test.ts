import { LayerAnalyzer } from '@analyzers/layerAnalyzer'
import { ChurnAnalyzer } from '@analyzers/churnAnalyzer'
import { CodeMap, CodeNode, CodeFile } from '@models/codeMap'
import { CodeNodeData, CodeFileData } from '@models/types'

describe('LayerAnalyzer', () => {
  let analyzer: LayerAnalyzer

  beforeEach(() => {
    analyzer = new LayerAnalyzer()
  })

  describe('detectLayer', () => {
    it('should detect UI layer from file path', () => {
      const layer = analyzer.detectLayer('src/ui/components/Header.tsx')
      expect(layer).toBe('ui')
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

    it('should detect utils layer', () => {
      const layer = analyzer.detectLayer('src/utils/helpers.ts')
      expect(layer).toBe('utils')
    })

    it('should detect config layer', () => {
      const layer = analyzer.detectLayer('src/config/settings.ts')
      expect(layer).toBe('config')
    })

    it('should default to other layer', () => {
      const layer = analyzer.detectLayer('src/random/file.ts')
      expect(layer).toBe('other')
    })
  })

  describe('isLayerViolation', () => {
    it('should detect violation when lower layer imports from higher layer', () => {
      const violation = analyzer.isLayerViolation('src/utils/helpers.ts', 'src/services/api.ts')
      expect(violation).toBe(true)
    })

    it('should not flag valid layer imports', () => {
      const violation = analyzer.isLayerViolation('src/services/api.ts', 'src/utils/helpers.ts')
      expect(violation).toBe(false)
    })
  })

  describe('classifyFiles', () => {
    it('should classify multiple files into layers', () => {
      const files = ['src/ui/Header.tsx', 'src/services/api.ts', 'src/utils/helpers.ts']
      const result = analyzer.classifyFiles(files)

      expect(result.get('ui')).toContain('src/ui/Header.tsx')
      expect(result.get('services')).toContain('src/services/api.ts')
      expect(result.get('utils')).toContain('src/utils/helpers.ts')
    })
  })

  describe('getLayerInfo', () => {
    it('should return layer info with color and order', () => {
      const info = analyzer.getLayerInfo('ui')
      expect(info.name).toBe('ui')
      expect(info.color).toBe('#4d9fff')
      expect(info.order).toBe(0)
    })
  })
})

describe('ChurnAnalyzer', () => {
  let analyzer: ChurnAnalyzer

  beforeEach(() => {
    analyzer = new ChurnAnalyzer()
  })

  describe('getChurnScore', () => {
    it('should calculate churn score from metrics', () => {
      const metrics = {
        commits: 10,
        additions: 500,
        deletions: 300,
        lastModified: new Date(),
      }
      const score = analyzer.getChurnScore(metrics)
      expect(score).toBe(10 + (500 + 300) / 100) // 10 + 8 = 18
    })
  })

  describe('getChurnLevel', () => {
    it('should classify critical churn', () => {
      const metrics = {
        commits: 60,
        additions: 0,
        deletions: 0,
        lastModified: new Date(),
      }
      const level = analyzer.getChurnLevel(metrics)
      expect(level).toBe('critical')
    })

    it('should classify high churn', () => {
      const metrics = {
        commits: 25,
        additions: 0,
        deletions: 0,
        lastModified: new Date(),
      }
      const level = analyzer.getChurnLevel(metrics)
      expect(level).toBe('high')
    })

    it('should classify low churn', () => {
      const metrics = {
        commits: 1,
        additions: 100,
        deletions: 50,
        lastModified: new Date(),
      }
      const level = analyzer.getChurnLevel(metrics)
      expect(level).toBe('low')
    })

    it('should classify none when no changes', () => {
      const metrics = {
        commits: 0,
        additions: 0,
        deletions: 0,
        lastModified: new Date(),
      }
      const level = analyzer.getChurnLevel(metrics)
      expect(level).toBe('none')
    })
  })

  describe('findHotspots', () => {
    it('should identify high-churn files', () => {
      const churnMap = new Map([
        [
          'file1.ts',
          {
            commits: 50,
            additions: 1000,
            deletions: 800,
            lastModified: new Date(),
          },
        ],
        [
          'file2.ts',
          {
            commits: 2,
            additions: 50,
            deletions: 30,
            lastModified: new Date(),
          },
        ],
      ])

      const hotspots = analyzer.findHotspots(churnMap, 10)
      expect(hotspots[0].file).toBe('file1.ts')
      expect(hotspots[0].score).toBeGreaterThan(hotspots[1].score)
    })
  })
})

describe('CodeMap - Blast Radius', () => {
  let codeMap: CodeMap

  beforeEach(() => {
    codeMap = new CodeMap('test-map', 'Test', '/test')

    // Add mock files
    const file1 = new CodeFile({
      id: 'file1',
      path: 'src/api.ts',
      language: 'typescript',
      size: 1000,
      lines: 100,
    })
    const file2 = new CodeFile({
      id: 'file2',
      path: 'src/auth.ts',
      language: 'typescript',
      size: 800,
      lines: 80,
    })
    const file3 = new CodeFile({
      id: 'file3',
      path: 'src/components/Header.tsx',
      language: 'typescript',
      size: 500,
      lines: 50,
    })

    codeMap.addFile(file1)
    codeMap.addFile(file2)
    codeMap.addFile(file3)
  })

  describe('getChurnScores', () => {
    it('should store and retrieve churn scores', () => {
      const scores = new Map([
        ['file1', 10],
        ['file2', 5],
      ])
      codeMap.setChurnScores(scores)

      const retrieved = codeMap.getChurnScores()
      expect(retrieved.get('file1')).toBe(10)
      expect(retrieved.get('file2')).toBe(5)
    })
  })

  describe('churn score methods', () => {
    it('should set and get individual churn scores', () => {
      codeMap.setChurnScore('file1', 15)
      expect(codeMap.getChurnScore('file1')).toBe(15)
    })

    it('should return undefined for unknown files', () => {
      expect(codeMap.getChurnScore('unknown')).toBeUndefined()
    })
  })
})
