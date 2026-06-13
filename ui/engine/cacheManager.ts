/**
 * IndexedDB-based cache manager for parsed code results
 * Provides O(1) retrieval with LRU eviction policy
 */

import { CodeNodeData, CodeFileData } from '@models/types'

export interface CacheEntry {
  key: string
  fileName: string
  language: string
  timestamp: number
  expiresAt: number
  nodes: CodeNodeData[]
  edges: CodeEdgeData[]
  file: CodeFileData
  contentHash: string
}

export interface CodeEdgeData {
  id: string
  source: string
  target: string
  type: string
}

const DB_NAME = 'CodeCartographerCache'
const STORE_NAME = 'ParsedFiles'
const EXPIRY_MS = 7 * 24 * 60 * 60 * 1000 // 7 days
const LRU_CHECK_INTERVAL = 10

export class CacheManager {
  private db: IDBDatabase | null = null
  private initialized = false
  private operationCount = 0

  async initialize(): Promise<void> {
    if (this.initialized) return

    return new Promise((resolve) => {
      const request = indexedDB.open(DB_NAME, 1)

      request.onerror = () => {
        console.error('Failed to open IndexedDB')
        this.initialized = true
        resolve()
      }

      request.onsuccess = (event) => {
        this.db = (event.target as IDBOpenDBRequest).result
        this.initialized = true
        this.cleanExpiredEntries()
        resolve()
      }

      request.onupgradeneeded = (event) => {
        const db = (event.target as IDBOpenDBRequest).result
        if (!db.objectStoreNames.contains(STORE_NAME)) {
          const store = db.createObjectStore(STORE_NAME, { keyPath: 'key' })
          store.createIndex('timestamp', 'timestamp', { unique: false })
          store.createIndex('expiresAt', 'expiresAt', { unique: false })
        }
      }
    })
  }

  private async ensureInitialized(): Promise<void> {
    if (!this.initialized) await this.initialize()
  }

  async get(fileName: string, contentHash: string): Promise<CacheEntry | null> {
    await this.ensureInitialized()
    if (!this.db) return null

    const key = `${fileName}_${contentHash}`

    return new Promise((resolve) => {
      try {
        const store = this.db!.transaction([STORE_NAME], 'readonly').objectStore(STORE_NAME)
        const request = store.get(key)

        request.onsuccess = () => {
          const entry = request.result as CacheEntry | undefined
          if (entry && entry.expiresAt > Date.now()) {
            this.updateTimestamp(key)
            resolve(entry)
          } else {
            resolve(null)
          }
        }

        request.onerror = () => resolve(null)
      } catch {
        resolve(null)
      }
    })
  }

  async set(
    fileName: string,
    contentHash: string,
    language: string,
    nodes: CodeNodeData[],
    edges: CodeEdgeData[],
    file: CodeFileData
  ): Promise<void> {
    await this.ensureInitialized()
    if (!this.db) return

    const key = `${fileName}_${contentHash}`
    const now = Date.now()

    const entry: CacheEntry = {
      key,
      fileName,
      language,
      timestamp: now,
      expiresAt: now + EXPIRY_MS,
      nodes,
      edges,
      file,
      contentHash,
    }

    try {
      this.db.transaction([STORE_NAME], 'readwrite').objectStore(STORE_NAME).put(entry)
      this.operationCount++

      if (this.operationCount % LRU_CHECK_INTERVAL === 0) {
        this.evictOldestIfNeeded()
      }
    } catch (error) {
      console.error('Cache storage failed:', error)
    }
  }

  private cleanExpiredEntries(): void {
    if (!this.db) return

    try {
      const index = this.db.transaction([STORE_NAME], 'readwrite')
        .objectStore(STORE_NAME)
        .index('expiresAt')

      index.openCursor(IDBKeyRange.upperBound(Date.now())).onsuccess = (event) => {
        const cursor = (event.target as IDBRequest).result
        if (cursor) {
          cursor.delete()
          cursor.continue()
        }
      }
    } catch {
      // Silently fail - cache cleanup is non-critical
    }
  }

  private updateTimestamp(key: string): void {
    if (!this.db) return

    try {
      const store = this.db.transaction([STORE_NAME], 'readwrite').objectStore(STORE_NAME)
      const request = store.get(key)

      request.onsuccess = () => {
        const entry = request.result as CacheEntry
        if (entry) {
          entry.timestamp = Date.now()
          store.put(entry)
        }
      }
    } catch {
      // Silently fail - LRU update is non-critical
    }
  }

  private evictOldestIfNeeded(): void {
    if (!this.db) return

    try {
      const index = this.db.transaction([STORE_NAME], 'readwrite')
        .objectStore(STORE_NAME)
        .index('timestamp')

      index.openCursor().onsuccess = (event) => {
        const cursor = (event.target as IDBRequest).result
        if (cursor) cursor.delete()
      }
    } catch {
      // Silently fail - eviction is non-critical
    }
  }

  async clear(): Promise<void> {
    await this.ensureInitialized()
    if (!this.db) return

    try {
      this.db.transaction([STORE_NAME], 'readwrite').objectStore(STORE_NAME).clear()
      this.operationCount = 0
    } catch (error) {
      console.error('Cache clear failed:', error)
    }
  }

  async getStats(): Promise<{ sizeMB: number; entryCount: number }> {
    await this.ensureInitialized()
    if (!this.db) return { sizeMB: 0, entryCount: 0 }

    return new Promise((resolve) => {
      try {
        const request = this.db!.transaction([STORE_NAME], 'readonly')
          .objectStore(STORE_NAME)
          .getAll()

        request.onsuccess = () => {
          const entries = request.result as CacheEntry[]
          const totalBytes = entries.reduce((sum, e) => sum + JSON.stringify(e).length, 0)
          resolve({
            sizeMB: Math.round(totalBytes / 1024 / 1024),
            entryCount: entries.length,
          })
        }

        request.onerror = () => resolve({ sizeMB: 0, entryCount: 0 })
      } catch {
        resolve({ sizeMB: 0, entryCount: 0 })
      }
    })
  }
}

// Export singleton instance
export const cacheManager = new CacheManager()
