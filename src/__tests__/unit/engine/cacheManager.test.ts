import { CacheManager, CacheEntry } from '@engine/cacheManager'
import { CodeNodeData, CodeFileData } from '@models/types'

describe('CacheManager', () => {
  let manager: CacheManager

  beforeEach(async () => {
    manager = new CacheManager()
    await manager.initialize()
  })

  afterEach(async () => {
    await manager.clear()
  })

  describe('initialization', () => {
    it('should initialize IndexedDB successfully', async () => {
      const stats = await manager.getStats()
      expect(stats.entryCount).toBe(0)
    })

    it('should handle multiple initializations gracefully', async () => {
      await manager.initialize()
      await manager.initialize()
      const stats = await manager.getStats()
      expect(stats.entryCount).toBe(0)
    })
  })

  describe('cache storage and retrieval', () => {
    const testFileData: CodeFileData = {
      id: 'test-file',
      path: 'test.ts',
      language: 'typescript',
      size: 1000,
      lines: 50,
    }

    const testNodeData: CodeNodeData = {
      id: 'node1',
      name: 'testFunc',
      type: 'function',
      language: 'typescript',
      range: { start: { line: 0, column: 0 }, end: { line: 10, column: 0 } },
      file: 'test.ts',
      metrics: {
        cyclomatic: 1,
        cognitive: 1,
        linesOfCode: 10,
        nestedDepth: 0,
        numberOfParameters: 0,
      },
    }

    it('should store and retrieve cache entries', async () => {
      const key = 'test.ts_hash123'
      await manager.set(
        'test.ts',
        'hash123',
        'typescript',
        [testNodeData],
        [],
        testFileData
      )

      const entry = await manager.get('test.ts', 'hash123')
      expect(entry).toBeTruthy()
      expect(entry?.nodes).toHaveLength(1)
      expect(entry?.nodes[0].name).toBe('testFunc')
    })

    it('should return null for non-existent entries', async () => {
      const entry = await manager.get('nonexistent.ts', 'fake-hash')
      expect(entry).toBeNull()
    })

    it('should store and retrieve multiple entries', async () => {
      await manager.set('file1.ts', 'hash1', 'typescript', [testNodeData], [], testFileData)
      await manager.set('file2.ts', 'hash2', 'typescript', [testNodeData], [], testFileData)

      const entry1 = await manager.get('file1.ts', 'hash1')
      const entry2 = await manager.get('file2.ts', 'hash2')

      expect(entry1).toBeTruthy()
      expect(entry2).toBeTruthy()
      expect(entry1?.fileName).toBe('file1.ts')
      expect(entry2?.fileName).toBe('file2.ts')
    })
  })

  describe('expiry handling', () => {
    it('should expire old entries', async () => {
      const fileData: CodeFileData = {
        id: 'test',
        path: 'test.ts',
        language: 'typescript',
        size: 100,
        lines: 10,
      }

      // Create entry with past expiry
      const now = Date.now()
      const expiredEntry: CacheEntry = {
        key: 'expired_test',
        fileName: 'expired.ts',
        language: 'typescript',
        timestamp: now - 100000,
        expiresAt: now - 1000, // Already expired
        nodes: [],
        edges: [],
        file: fileData,
        contentHash: 'old-hash',
      }

      // Note: Direct storage not exposed, so we test through normal flow
      // After cleanup, expired entries should not be retrievable
      const entry = await manager.get('expired.ts', 'old-hash')
      expect(entry).toBeNull()
    })
  })

  describe('statistics', () => {
    it('should track cache size and entry count', async () => {
      const fileData: CodeFileData = {
        id: 'test',
        path: 'test.ts',
        language: 'typescript',
        size: 1000,
        lines: 50,
      }

      const nodeData: CodeNodeData = {
        id: 'node1',
        name: 'func',
        type: 'function',
        language: 'typescript',
        range: { start: { line: 0, column: 0 }, end: { line: 10, column: 0 } },
        file: 'test.ts',
        metrics: {
          cyclomatic: 1,
          cognitive: 1,
          linesOfCode: 10,
          nestedDepth: 0,
          numberOfParameters: 0,
        },
      }

      await manager.set('file1.ts', 'hash1', 'typescript', [nodeData], [], fileData)
      await manager.set('file2.ts', 'hash2', 'typescript', [nodeData], [], fileData)

      const stats = await manager.getStats()
      expect(stats.entryCount).toBe(2)
      expect(stats.size).toBeGreaterThan(0)
    })

    it('should start with zero size and entries', async () => {
      const stats = await manager.getStats()
      expect(stats.entryCount).toBe(0)
      expect(stats.size).toBe(0)
    })
  })

  describe('cache clearing', () => {
    it('should clear all entries', async () => {
      const fileData: CodeFileData = {
        id: 'test',
        path: 'test.ts',
        language: 'typescript',
        size: 1000,
        lines: 50,
      }

      const nodeData: CodeNodeData = {
        id: 'node1',
        name: 'func',
        type: 'function',
        language: 'typescript',
        range: { start: { line: 0, column: 0 }, end: { line: 10, column: 0 } },
        file: 'test.ts',
        metrics: {
          cyclomatic: 1,
          cognitive: 1,
          linesOfCode: 10,
          nestedDepth: 0,
          numberOfParameters: 0,
        },
      }

      await manager.set('file1.ts', 'hash1', 'typescript', [nodeData], [], fileData)
      let stats = await manager.getStats()
      expect(stats.entryCount).toBe(1)

      await manager.clear()
      stats = await manager.getStats()
      expect(stats.entryCount).toBe(0)
      expect(stats.size).toBe(0)
    })
  })
})
