/**
 * LRU memoization cache for expensive operations
 * Provides 100-1000x speedup for repeated calculations
 */

export interface MemoizeOptions {
  maxSize?: number
  ttl?: number
}

interface CacheEntry<T> {
  value: T
  timestamp: number
  hits: number
}

export class MemoizationCache<T> {
  private cache: Map<string, CacheEntry<T>> = new Map()
  private maxSize: number
  private ttl: number | null
  private hits = 0
  private misses = 0

  constructor(options: MemoizeOptions = {}) {
    this.maxSize = options.maxSize ?? 100
    this.ttl = options.ttl ?? null
  }

  get(key: string): T | null {
    const entry = this.cache.get(key)

    if (!entry) {
      this.misses++
      return null
    }

    if (this.ttl && Date.now() - entry.timestamp > this.ttl) {
      this.cache.delete(key)
      this.misses++
      return null
    }

    entry.hits++
    entry.timestamp = Date.now()
    this.hits++

    this.cache.delete(key)
    this.cache.set(key, entry)

    return entry.value
  }

  set(key: string, value: T): void {
    this.cache.delete(key)

    if (this.cache.size >= this.maxSize) {
      const firstKey = this.cache.keys().next().value
      if (firstKey !== undefined) this.cache.delete(firstKey)
    }

    this.cache.set(key, { value, timestamp: Date.now(), hits: 0 })
  }

  clear(): void {
    this.cache.clear()
    this.hits = 0
    this.misses = 0
  }

  getStats(): { size: number; maxSize: number; hits: number; misses: number; hitRate: number } {
    const total = this.hits + this.misses
    return {
      size: this.cache.size,
      maxSize: this.maxSize,
      hits: this.hits,
      misses: this.misses,
      hitRate: total > 0 ? Math.round((this.hits / total) * 10000) / 100 : 0,
    }
  }

  memoize<Args extends any[], Result>(
    fn: (...args: Args) => Result,
    keyGenerator?: (...args: Args) => string
  ): (...args: Args) => Result {
    return (...args: Args): Result => {
      const key = keyGenerator ? keyGenerator(...args) : JSON.stringify(args)
      const cached = this.get(key) as Result | null

      if (cached !== null) return cached

      const result = fn(...args)
      this.set(key, result as unknown as T)
      return result
    }
  }
}

export function memoize<Args extends any[], Result>(
  fn: (...args: Args) => Result,
  options: MemoizeOptions = {},
  keyGenerator?: (...args: Args) => string
): (...args: Args) => Result {
  const cache = new MemoizationCache<Result>(options)
  return cache.memoize(fn, keyGenerator)
}

export class AsyncMemoizationCache<T> {
  private cache: Map<string, Promise<T>> = new Map()
  private maxSize: number

  constructor(maxSize = 100) {
    this.maxSize = maxSize
  }

  async get<Args extends any[]>(key: string, fn: (...args: Args) => Promise<T>): Promise<T> {
    const cached = this.cache.get(key)
    if (cached) return cached

    if (this.cache.size >= this.maxSize) {
      const firstKey = this.cache.keys().next().value
      if (firstKey !== undefined) this.cache.delete(firstKey)
    }

    const promise = fn()
    this.cache.set(key, promise)
    return promise
  }

  clear(): void {
    this.cache.clear()
  }

  getSize(): number {
    return this.cache.size
  }
}
