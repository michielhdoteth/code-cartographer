/**
 * Web Worker for parallel AST parsing
 * Runs parser.parseAndIntegrate() in a separate thread to avoid blocking main thread
 */

import { ParserRegistry } from '@parsers/base'
import { JavaScriptParser } from '@parsers/javascript'
import { TypeScriptParser } from '@parsers/typescript'
import { PythonParser } from '@parsers/python'
import { JavaParser } from '@parsers/java'
import { GoParser } from '@parsers/go'
import { RustParser } from '@parsers/rust'
import { CppParser } from '@parsers/cpp'
import { RubyParser } from '@parsers/ruby'
import { PhpParser } from '@parsers/php'
import { CParser } from '@parsers/c'
import { CSharpParser } from '@parsers/csharp'
import { SwiftParser } from '@parsers/swift'
import { KotlinParser } from '@parsers/kotlin'
import { ScalaParser } from '@parsers/scala'
import { HtmlParser } from '@parsers/html'
import { CssParser } from '@parsers/css'
import { SqlParser } from '@parsers/sql'
import { ShellParser } from '@parsers/shell'
import { YamlParser } from '@parsers/yaml'
import { JsonParser } from '@parsers/json'
import { TomlParser } from '@parsers/toml'
import { MarkdownParser } from '@parsers/markdown'
import { CodeMap } from '@models/codeMap'
import { Language } from '@models/types'

export interface ParseWorkerMessage {
  fileId: string
  fileName: string
  content: string
  language: Language
}

export interface ParseResultMessage {
  fileId: string
  success: boolean
  nodes?: any[]
  edges?: any[]
  error?: string
}

const PARSER_MAP: Record<string, () => any> = {
  javascript: () => new JavaScriptParser(),
  jsx: () => new JavaScriptParser(),
  typescript: () => new TypeScriptParser(),
  tsx: () => new TypeScriptParser(),
  python: () => new PythonParser(),
  java: () => new JavaParser(),
  go: () => new GoParser(),
  rust: () => new RustParser(),
  cpp: () => new CppParser(),
  ruby: () => new RubyParser(),
  php: () => new PhpParser(),
  c: () => new CParser(),
  csharp: () => new CSharpParser(),
  swift: () => new SwiftParser(),
  kotlin: () => new KotlinParser(),
  scala: () => new ScalaParser(),
  html: () => new HtmlParser(),
  css: () => new CssParser(),
  sql: () => new SqlParser(),
  shell: () => new ShellParser(),
  yaml: () => new YamlParser(),
  json: () => new JsonParser(),
  toml: () => new TomlParser(),
  markdown: () => new MarkdownParser(),
}

function createRegistry(): ParserRegistry {
  const registry = new ParserRegistry()
  for (const [lang, createParser] of Object.entries(PARSER_MAP)) {
    registry.register(lang as Language, createParser())
  }
  return registry
}

const registry = createRegistry()

function parseFile(message: ParseWorkerMessage): ParseResultMessage {
  const { fileId, fileName, content, language } = message

  const parser = registry.getParser(language)
  if (!parser) {
    return { fileId, success: false, error: `No parser for language: ${language}` }
  }

  const tempMap = new CodeMap(`temp_${Date.now()}`, 'TempMap', '/')
  parser.parseAndIntegrate(fileName, content, tempMap)

  return {
    fileId,
    success: true,
    nodes: tempMap.getNodes().map((n) => n.data),
    edges: tempMap.getEdges().map((e) => e.data),
  }
}

self.onmessage = (event: MessageEvent<ParseWorkerMessage>): void => {
  try {
    self.postMessage(parseFile(event.data))
  } catch (error) {
    self.postMessage({
      fileId: event.data.fileId,
      success: false,
      error: error instanceof Error ? error.message : String(error),
    })
  }
}

self.onerror = (error: ErrorEvent): void => {
  console.error('Worker error:', error.message)
  self.postMessage({ success: false, error: `Worker error: ${error.message}` })
}
