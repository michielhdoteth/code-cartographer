/**
 * Input validation and sanitization utilities
 * Prevents command injection and path traversal attacks
 */

import path from 'path'

// =============================================================================
// Types
// =============================================================================

export type SupportedLanguage =
  | 'javascript'
  | 'typescript'
  | 'jsx'
  | 'tsx'
  | 'python'
  | 'java'
  | 'go'
  | 'rust'
  | 'cpp'
  | 'ruby'
  | 'php'

// =============================================================================
// Path Validation
// =============================================================================

/**
 * Validate and sanitize file paths to prevent traversal attacks
 */
export function sanitizeFilePath(inputPath: string): string | null {
  if (!inputPath) return null

  try {
    const normalized = path.normalize(inputPath)

    if (containsPathTraversal(normalized) || containsNullBytes(normalized)) {
      return null
    }

    return normalized
  } catch {
    return null
  }
}

function containsPathTraversal(filepath: string): boolean {
  return filepath.includes('..')
}

function containsNullBytes(text: string): boolean {
  return text.includes('\x00')
}

/**
 * Validate file extension against allowed list
 */
export function validateFileExtension(fileName: string, allowedExtensions: string[] = []): boolean {
  if (allowedExtensions.length === 0) return true

  const ext = path.extname(fileName).toLowerCase()
  return allowedExtensions.includes(ext)
}

/**
 * Validate file size to prevent memory issues
 */
export function validateFileSize(sizeBytes: number, maxSizeMB: number = 50): boolean {
  const maxBytes = maxSizeMB * 1024 * 1024
  return sizeBytes > 0 && sizeBytes <= maxBytes
}

// =============================================================================
// Git Command Validation
// =============================================================================

const SAFE_GIT_COMMAND_PATTERN = /^[a-zA-Z0-9._\-/]*$/
const SHELL_INJECTION_PATTERNS = ['$(', '`', '&', '|', ';', '>']

/**
 * Sanitize git command arguments to prevent injection
 */
export function sanitizeGitCommand(command: string): string | null {
  if (!SAFE_GIT_COMMAND_PATTERN.test(command)) {
    return null
  }

  if (SHELL_INJECTION_PATTERNS.some((pattern) => command.includes(pattern))) {
    return null
  }

  return command
}

/**
 * Sanitize author name for git operations
 */
export function sanitizeAuthorName(author: string): string | null {
  const SAFE_AUTHOR_PATTERN = /^[a-zA-Z0-9\s._\-]*$/
  const MAX_AUTHOR_LENGTH = 100

  if (!SAFE_AUTHOR_PATTERN.test(author) || author.length > MAX_AUTHOR_LENGTH) {
    return null
  }

  return author.trim()
}

// =============================================================================
// Language Validation
// =============================================================================

const SUPPORTED_LANGUAGES: SupportedLanguage[] = [
  'javascript',
  'typescript',
  'jsx',
  'tsx',
  'python',
  'java',
  'go',
  'rust',
  'cpp',
  'ruby',
  'php',
]

/**
 * Validate programming language
 */
export function validateLanguage(lang: string): SupportedLanguage | null {
  const lower = lang.toLowerCase()
  return SUPPORTED_LANGUAGES.includes(lower as SupportedLanguage) ? (lower as SupportedLanguage) : null
}

// =============================================================================
// Repository Validation
// =============================================================================

/**
 * Check if path is a git repository by looking for .git directory
 */
export async function isGitRepository(dirPath: string): Promise<boolean> {
  const sanitized = sanitizeFilePath(dirPath)
  if (!sanitized) return false

  try {
    const fs = await import('fs/promises')
    const gitPath = path.join(sanitized, '.git')
    const stat = await fs.stat(gitPath)
    return stat.isDirectory()
  } catch {
    return false
  }
}

// =============================================================================
// Network Validation
// =============================================================================

/**
 * Validate port number for network operations
 */
export function validatePort(port: number): boolean {
  return Number.isInteger(port) && port > 0 && port < 65536
}

/**
 * Validate URL format
 */
export function validateUrl(url: string): boolean {
  try {
    new URL(url)
    return true
  } catch {
    return false
  }
}

// =============================================================================
// Content Validation
// =============================================================================

const SUSPICIOUS_CREDENTIAL_PATTERNS = [
  /password\s*[:=]/i,
  /api[_-]?key\s*[:=]/i,
  /secret\s*[:=]/i,
  /token\s*[:=]/i,
  /auth\s*[:=]/i,
]

/**
 * Check if string looks like it contains credentials
 */
export function containsSuspiciousPatterns(text: string): boolean {
  return SUSPICIOUS_CREDENTIAL_PATTERNS.some((pattern) => pattern.test(text))
}

/**
 * Validate JSON structure
 */
export function validateJson(text: string): boolean {
  try {
    JSON.parse(text)
    return true
  } catch {
    return false
  }
}
