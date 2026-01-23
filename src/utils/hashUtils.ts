/**
 * Fast hashing utilities for cache key generation
 * Uses SHA-256 async with sync fallback
 */

export async function hashContent(content: string): Promise<string> {
  try {
    const data = new TextEncoder().encode(content)
    const hashBuffer = await crypto.subtle.digest('SHA-256', data)
    return bufferToHex(hashBuffer)
  } catch {
    return hashContentSync(content)
  }
}

export function hashContentSync(content: string): string {
  let hash = 0
  for (let i = 0; i < content.length; i++) {
    hash = ((hash << 5) - hash + content.charCodeAt(i)) | 0
  }
  return `hash_${Math.abs(hash).toString(36)}_${content.length}`
}

function bufferToHex(buffer: ArrayBuffer): string {
  return Array.from(new Uint8Array(buffer))
    .map((b) => b.toString(16).padStart(2, '0'))
    .join('')
}

export async function generateCacheKey(fileName: string, content: string): Promise<string> {
  return `${fileName}_${await hashContent(content)}`
}

export function generateCacheKeySync(fileName: string, content: string): string {
  return `${fileName}_${hashContentSync(content)}`
}

export function hasContentLikelyChanged(oldContent: string, newContent: string): boolean {
  if (oldContent.length !== newContent.length) return true
  if (oldContent === newContent) return false

  const sampleSize = 100
  return (
    oldContent.slice(0, sampleSize) !== newContent.slice(0, sampleSize) ||
    oldContent.slice(-sampleSize) !== newContent.slice(-sampleSize)
  )
}
