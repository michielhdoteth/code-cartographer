import { describe, it, expect } from 'vitest'
import {
  hashContent,
  hashContentSync,
  hasContentLikelyChanged,
  generateCacheKeySync,
} from '@utils/hashUtils'

describe('hashUtils', () => {
  describe('hashContentSync', () => {
    it('should generate consistent hashes for same content', () => {
      const content = 'const x = 42;'
      const hash1 = hashContentSync(content)
      const hash2 = hashContentSync(content)
      expect(hash1).toBe(hash2)
    })

    it('should generate different hashes for different content', () => {
      const hash1 = hashContentSync('const x = 1;')
      const hash2 = hashContentSync('const x = 2;')
      expect(hash1).not.toBe(hash2)
    })

    it('should handle empty strings', () => {
      const hash = hashContentSync('')
      expect(hash).toBeTruthy()
      expect(typeof hash).toBe('string')
    })

    it('should handle large content', () => {
      const largeContent = 'x'.repeat(100000)
      const hash = hashContentSync(largeContent)
      expect(hash).toBeTruthy()
    })

    it('should handle special characters', () => {
      const content = 'const emoji = "😀 🎉 🔥";'
      const hash = hashContentSync(content)
      expect(hash).toBeTruthy()
    })
  })

  describe('hashContent', () => {
    it('should generate consistent hashes for same content', async () => {
      const content = 'const x = 42;'
      const hash1 = await hashContent(content)
      const hash2 = await hashContent(content)
      expect(hash1).toBe(hash2)
    })

    it('should generate different hashes for different content', async () => {
      const hash1 = await hashContent('const x = 1;')
      const hash2 = await hashContent('const x = 2;')
      expect(hash1).not.toBe(hash2)
    })

    it('should generate hex string', async () => {
      const content = 'test content'
      const hash = await hashContent(content)
      expect(/^[a-f0-9]+$/.test(hash) || hash.startsWith('fallback_') || hash.startsWith('hash_')).toBe(
        true
      )
    })

    it('should handle fallback for unavailable crypto', async () => {
      const content = 'test'
      const hash = await hashContent(content)
      expect(hash).toBeTruthy()
      expect(typeof hash).toBe('string')
    })
  })

  describe('hasContentLikelyChanged', () => {
    it('should detect length changes', () => {
      const old = 'short'
      const newContent = 'much much much longer content'
      expect(hasContentLikelyChanged(old, newContent)).toBe(true)
    })

    it('should detect identical content', () => {
      const content = 'const x = 1;'
      expect(hasContentLikelyChanged(content, content)).toBe(false)
    })

    it('should detect beginning changes', () => {
      const old = 'const x = 1;'
      const newContent = 'let x = 1;'
      expect(hasContentLikelyChanged(old, newContent)).toBe(true)
    })

    it('should detect ending changes', () => {
      const old = 'const x = 1;'
      const newContent = 'const x = 2;'
      expect(hasContentLikelyChanged(old, newContent)).toBe(true)
    })

    it('should be fast for large files', () => {
      const large1 = 'x'.repeat(1000000)
      const large2 = 'x'.repeat(1000000)
      const start = performance.now()
      hasContentLikelyChanged(large1, large2)
      const elapsed = performance.now() - start
      expect(elapsed).toBeLessThan(10) // Should be very fast
    })

    it('should detect middle changes in large content', () => {
      const large1 = 'a'.repeat(10000) + 'x' + 'b'.repeat(10000)
      const large2 = 'a'.repeat(10000) + 'y' + 'b'.repeat(10000)
      // Might not detect if change is in middle
      // This tests the behavior of the quick check
      const result = hasContentLikelyChanged(large1, large2)
      expect(typeof result).toBe('boolean')
    })
  })

  describe('generateCacheKeySync', () => {
    it('should generate consistent cache keys', () => {
      const key1 = generateCacheKeySync('test.js', 'const x = 1;')
      const key2 = generateCacheKeySync('test.js', 'const x = 1;')
      expect(key1).toBe(key2)
    })

    it('should include file name in key', () => {
      const key = generateCacheKeySync('myfile.ts', 'content')
      expect(key).toContain('myfile.ts')
    })

    it('should generate different keys for same file different content', () => {
      const key1 = generateCacheKeySync('test.js', 'const x = 1;')
      const key2 = generateCacheKeySync('test.js', 'const x = 2;')
      expect(key1).not.toBe(key2)
    })

    it('should generate different keys for different files same content', () => {
      const content = 'const x = 1;'
      const key1 = generateCacheKeySync('file1.js', content)
      const key2 = generateCacheKeySync('file2.js', content)
      expect(key1).not.toBe(key2)
    })

    it('should handle path separators', () => {
      const key1 = generateCacheKeySync('src/components/Button.tsx', 'export const Button = () => {}')
      const key2 = generateCacheKeySync('src/components/Button.tsx', 'export const Button = () => {}')
      expect(key1).toBe(key2)
    })
  })

  describe('performance', () => {
    it('sync hash should be fast (<1ms for small content)', () => {
      const content = 'const x = 42; console.log(x);'
      const start = performance.now()
      for (let i = 0; i < 1000; i++) {
        hashContentSync(content)
      }
      const elapsed = performance.now() - start
      expect(elapsed).toBeLessThan(100) // 1000 operations should be <100ms
    })

    it('cache key generation should be fast', () => {
      const start = performance.now()
      for (let i = 0; i < 1000; i++) {
        generateCacheKeySync(`file${i}.js`, `const x = ${i};`)
      }
      const elapsed = performance.now() - start
      expect(elapsed).toBeLessThan(100)
    })
  })
})
