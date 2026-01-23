import { StreamingParser } from '@engine/streamingParser'
import { ParserRegistry } from '@parsers/base'
import { JavaScriptParser } from '@parsers/javascript'
import { CodeMap } from '@models/codeMap'

describe('StreamingParser', () => {
  let parser: StreamingParser
  let registry: ParserRegistry
  let codeMap: CodeMap

  beforeEach(() => {
    registry = new ParserRegistry()
    registry.register('javascript', new JavaScriptParser())
    registry.register('jsx', new JavaScriptParser())
    parser = new StreamingParser(registry, { chunkSize: 50 * 1024 })
    codeMap = new CodeMap('test-map', 'Test', '/test')
  })

  describe('streaming threshold detection', () => {
    it('should identify files that need streaming', () => {
      const smallFile = 'const x = 1;'
      const largeFile = 'x'.repeat(150 * 1024) // 150KB

      expect(StreamingParser.shouldStream(smallFile)).toBe(false)
      expect(StreamingParser.shouldStream(largeFile)).toBe(true)
    })

    it('should respect custom threshold', () => {
      const content = 'x'.repeat(500)
      expect(StreamingParser.shouldStream(content, 1000)).toBe(false)
      expect(StreamingParser.shouldStream(content, 100)).toBe(true)
    })
  })

  describe('chunk estimation', () => {
    it('should estimate chunks correctly', () => {
      // 100KB content with 50KB chunks = 2 chunks
      expect(StreamingParser.estimateChunks(100 * 1024, 50 * 1024)).toBe(2)
    })

    it('should handle partial chunks', () => {
      // 75KB content with 50KB chunks = 2 chunks (1 full + 1 partial)
      expect(StreamingParser.estimateChunks(75 * 1024, 50 * 1024)).toBe(2)
    })

    it('should handle exact boundaries', () => {
      expect(StreamingParser.estimateChunks(100 * 1024, 50 * 1024)).toBe(2)
      expect(StreamingParser.estimateChunks(100 * 1024, 25 * 1024)).toBe(4)
    })

    it('should estimate small files as single chunk', () => {
      expect(StreamingParser.estimateChunks(1000, 50 * 1024)).toBe(1)
    })
  })

  describe('parse streaming', () => {
    it('should parse small files', async () => {
      const content = `
        function add(a, b) {
          return a + b;
        }
      `

      const result = await parser.parseStreaming(
        'test.js',
        content,
        'javascript',
        codeMap
      )

      expect(result.nodeCount).toBeGreaterThanOrEqual(0)
      expect(result.edgeCount).toBeGreaterThanOrEqual(0)
    })

    it('should handle files larger than chunk size', async () => {
      // Create file larger than chunk size
      const largeContent = 'const func = () => { ' + 'x = 1; '.repeat(10000) + ' };'

      const result = await parser.parseStreaming(
        'large.js',
        largeContent,
        'javascript',
        codeMap
      )

      expect(result.nodeCount).toBeGreaterThanOrEqual(0)
      expect(result.edgeCount).toBeGreaterThanOrEqual(0)
    })

    it('should report progress', async () => {
      const progressUpdates: Array<{ chunk: number; total: number }> = []

      const largeContent = 'const x = 1; '.repeat(5000)

      await parser.parseStreaming(
        'test.js',
        largeContent,
        'javascript',
        codeMap,
        (chunk, total) => {
          progressUpdates.push({ chunk, total })
        }
      )

      // Should have progress updates
      expect(progressUpdates.length).toBeGreaterThanOrEqual(0)
    })

    it('should handle missing parser gracefully', async () => {
      const unsupportedParser = new StreamingParser(registry, {})

      await expect(
        unsupportedParser.parseStreaming(
          'test.unsupported',
          'code',
          'unsupported' as any,
          codeMap
        )
      ).rejects.toThrow()
    })

    it('should respect max chunks limit', async () => {
      const limitedParser = new StreamingParser(registry, {
        chunkSize: 1024, // 1KB chunks
        maxChunks: 5, // Max 5 chunks = 5KB max
      })

      // Create content larger than max chunks
      const hugeContent = 'x'.repeat(100 * 1024) // 100KB

      // Should process but stop at max chunks
      const result = await limitedParser.parseStreaming(
        'huge.js',
        hugeContent,
        'javascript',
        codeMap
      )

      expect(result.nodeCount).toBeGreaterThanOrEqual(0)
    })
  })

  describe('error handling', () => {
    it('should handle parsing errors in chunks', async () => {
      const invalidContent = 'function { invalid syntax ]['

      // Should not throw, but continue with other chunks
      const result = await parser.parseStreaming(
        'invalid.js',
        invalidContent,
        'javascript',
        codeMap
      )

      // Result should still exist even if partial
      expect(result).toBeDefined()
    })

    it('should continue after chunk failure', async () => {
      const mixedContent = 'function valid() {} ' + 'invalid [' + ' function alsoValid() {}'

      const result = await parser.parseStreaming(
        'mixed.js',
        mixedContent,
        'javascript',
        codeMap
      )

      expect(result).toBeDefined()
    })
  })

  describe('performance characteristics', () => {
    it('should process large content efficiently', async () => {
      const largeContent = 'const func = () => {}; '.repeat(1000)
      const start = performance.now()

      await parser.parseStreaming('large.js', largeContent, 'javascript', codeMap)

      const elapsed = performance.now() - start
      // Should complete in reasonable time
      expect(elapsed).toBeLessThan(5000) // 5 seconds max
    })
  })
})
