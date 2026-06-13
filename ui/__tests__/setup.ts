/**
 * Global test setup for vitest
 * Provides browser environment mocks and shared utilities
 */

import { vi } from 'vitest'

// =============================================================================
// Browser API Mocks
// =============================================================================

/**
 * Mock window.matchMedia for responsive design tests
 */
function setupMatchMedia(): void {
  Object.defineProperty(window, 'matchMedia', {
    writable: true,
    value: vi.fn().mockImplementation((query: string) => ({
      matches: false,
      media: query,
      onchange: null,
      addListener: vi.fn(),
      removeListener: vi.fn(),
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      dispatchEvent: vi.fn(),
    })),
  })
}

/**
 * Mock crypto.subtle for hash operations
 */
function setupCrypto(): void {
  if (!global.crypto) {
    global.crypto = {
      subtle: {
        digest: async (_algorithm: string, _data: ArrayBuffer) => new ArrayBuffer(32),
      },
    } as Crypto
  }
}

// =============================================================================
// IndexedDB Mock
// =============================================================================

class MockObjectStore {
  private data = new Map<string, unknown>()

  put(value: { key?: string }): { onsuccess: null } {
    if (value.key) {
      this.data.set(value.key, value)
    }
    return { onsuccess: null }
  }

  get(key: string): { result: unknown; onsuccess: null } {
    return { result: this.data.get(key), onsuccess: null }
  }

  getAll(): { result: unknown[]; onsuccess: null } {
    return { result: Array.from(this.data.values()), onsuccess: null }
  }

  delete(key: string): void {
    this.data.delete(key)
  }

  clear(): void {
    this.data.clear()
  }

  createIndex(_name: string): this {
    return this
  }

  index(_name: string): this {
    return this
  }

  openCursor(): { onsuccess: null } {
    return { onsuccess: null }
  }
}

class MockTransaction {
  objectStore(_name: string): MockObjectStore {
    return new MockObjectStore()
  }
}

class MockDatabase {
  objectStoreNames = new Set<string>()

  createObjectStore(name: string): MockObjectStore {
    this.objectStoreNames.add(name)
    return new MockObjectStore()
  }

  transaction(_storeNames: string | string[], _mode?: string): MockTransaction {
    return new MockTransaction()
  }

  close(): void {
    // No-op for mock
  }
}

class MockIndexedDB {
  open(): {
    onsuccess: null
    onerror: null
    onupgradeneeded: null
    result: MockDatabase
  } {
    return {
      onsuccess: null,
      onerror: null,
      onupgradeneeded: null,
      result: new MockDatabase(),
    }
  }
}

function setupIndexedDB(): void {
  global.indexedDB = new MockIndexedDB() as unknown as IDBFactory
}

// =============================================================================
// Initialize All Mocks
// =============================================================================

setupMatchMedia()
setupCrypto()
setupIndexedDB()

// =============================================================================
// Re-export fixtures and helpers for convenience
// =============================================================================

export * from './fixtures/index'
export * from './helpers/index'

// =============================================================================
// Legacy exports for backwards compatibility
// =============================================================================

export { createTestNode as createTestFile } from './fixtures/index'

// =============================================================================
// Test Configuration Constants
// =============================================================================

export const testConfig = {
  timeoutMs: 10000,
  largeFileSize: 100 * 1024, // 100KB
  hugeFileSize: 1000 * 1024, // 1MB
  performanceBenchmarkSize: 1000,
  performanceTargets: {
    parseMs: 5000,
    renderMs: 2000,
    queryMs: 1,
    blastRadiusMs: 100,
  },
}
