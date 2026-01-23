/**
 * Streaming parser for large files (>100KB)
 * Breaks files into chunks to prevent memory spikes
 */

import { CodeMap } from '@models/codeMap'
import { ParserRegistry } from '@parsers/base'
import { Language } from '@models/types'

export interface StreamingParseOptions {
  chunkSize?: number
  maxChunks?: number
}

const DEFAULT_CHUNK_SIZE = 50 * 1024
const DEFAULT_MAX_CHUNKS = 50
const STREAM_THRESHOLD = 100 * 1024

export class StreamingParser {
  private registry: ParserRegistry
  private chunkSize: number
  private maxChunks: number

  constructor(registry: ParserRegistry, options: StreamingParseOptions = {}) {
    this.registry = registry
    this.chunkSize = options.chunkSize ?? DEFAULT_CHUNK_SIZE
    this.maxChunks = options.maxChunks ?? DEFAULT_MAX_CHUNKS
  }

  static shouldStream(content: string): boolean {
    return content.length > STREAM_THRESHOLD
  }

  static estimateChunks(contentSize: number, chunkSize = DEFAULT_CHUNK_SIZE): number {
    return Math.ceil(contentSize / chunkSize)
  }

  async parseStreaming(
    fileName: string,
    content: string,
    language: Language,
    codeMap: CodeMap,
    onProgress?: (chunk: number, total: number) => void
  ): Promise<{ nodeCount: number; edgeCount: number }> {
    const parser = this.registry.getParser(language)
    if (!parser) {
      throw new Error(`No parser for language: ${language}`)
    }

    const totalChunks = Math.ceil(content.length / this.chunkSize)
    const chunksToProcess = Math.min(totalChunks, this.maxChunks)

    if (totalChunks > this.maxChunks) {
      console.warn(`File ${fileName} truncated to ${this.maxChunks} chunks`)
    }

    let nodeCount = 0
    let edgeCount = 0

    for (let i = 0; i < chunksToProcess; i++) {
      const start = i * this.chunkSize
      const chunk = content.substring(start, Math.min(start + this.chunkSize, content.length))

      try {
        const tempMap = new CodeMap(`chunk_${i}`, `Chunk ${i}`, '/')
        parser.parseAndIntegrate(`${fileName}_chunk_${i}`, chunk, tempMap)

        for (const node of tempMap.getNodes()) {
          codeMap.addNode(node)
          nodeCount++
        }

        for (const edge of tempMap.getEdges()) {
          codeMap.addEdge(edge)
          edgeCount++
        }

        onProgress?.(i + 1, chunksToProcess)
      } catch (error) {
        console.error(`Failed to parse chunk ${i} of ${fileName}:`, error)
      }
    }

    return { nodeCount, edgeCount }
  }
}
