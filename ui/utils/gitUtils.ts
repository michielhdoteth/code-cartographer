/**
 * Cross-platform async git utilities
 * Handles git operations without blocking UI
 */

import { sanitizeFilePath, sanitizeGitCommand } from './validation.js'

// =============================================================================
// Types
// =============================================================================

export interface GitOptions {
  cwd: string
  timeout?: number
  onProgress?: (percent: number, message: string) => void
}

export interface GitCommitStats {
  hash: string
  author: string
  date: Date
  message: string
  additions: number
  deletions: number
}

export interface FileChurn {
  file: string
  commits: number
  additions: number
  deletions: number
  lastModified: Date
}

export interface RepositoryStatus {
  branch: string
  isDirty: boolean
  stagedChanges: number
  unstagedChanges: number
}

// =============================================================================
// Core Git Execution
// =============================================================================

/**
 * Execute git command asynchronously with validation and timeout
 */
export async function execGit(args: string[], options: GitOptions): Promise<string> {
  const { cwd, timeout = 30000, onProgress } = options

  const validatedPath = sanitizeFilePath(cwd)
  if (!validatedPath) {
    throw new Error('Invalid repository path')
  }

  validateGitArgs(args)

  onProgress?.(0, 'Initializing git command')

  const execFn = await getExecFunction()
  return executeWithTimeout(execFn, args, validatedPath, timeout, onProgress)
}

function validateGitArgs(args: string[]): void {
  for (const arg of args) {
    const isOption = arg.startsWith('-')
    const isValid = sanitizeGitCommand(arg) !== null
    if (!isOption && !isValid) {
      throw new Error(`Invalid git argument: ${arg}`)
    }
  }
}

async function getExecFunction(): Promise<Function> {
  try {
    const { exec } = await import('child_process')
    return exec
  } catch {
    throw new Error('Git execution not available in this environment')
  }
}

function executeWithTimeout(
  execFn: Function,
  args: string[],
  cwd: string,
  timeout: number,
  onProgress?: (percent: number, message: string) => void
): Promise<string> {
  return new Promise((resolve, reject) => {
    const timeoutHandle = setTimeout(() => {
      reject(new Error(`Git command timed out after ${timeout}ms`))
    }, timeout)

    const command = `git ${args.join(' ')}`
    execFn(
      command,
      { cwd, maxBuffer: 10 * 1024 * 1024 },
      (error: Error | null, stdout: string, stderr: string) => {
        clearTimeout(timeoutHandle)
        if (error) {
          reject(new Error(`Git error: ${stderr || error.message}`))
        } else {
          onProgress?.(100, 'Git command completed')
          resolve(stdout)
        }
      }
    )
  })
}

// =============================================================================
// Repository Operations
// =============================================================================

/**
 * Check if directory is a git repository
 */
export async function isGitRepository(dirPath: string): Promise<boolean> {
  const validatedPath = sanitizeFilePath(dirPath)
  if (!validatedPath) return false

  try {
    await execGit(['rev-parse', '--git-dir'], { cwd: validatedPath })
    return true
  } catch {
    return false
  }
}

/**
 * Get current branch name
 */
export async function getCurrentBranch(
  repoPath: string,
  options?: { onProgress?: (percent: number, message: string) => void }
): Promise<string> {
  const output = await execGit(['rev-parse', '--abbrev-ref', 'HEAD'], {
    cwd: repoPath,
    onProgress: options?.onProgress,
  })
  return output.trim()
}

/**
 * Get repository status including branch, dirty state, and change counts
 */
export async function getRepositoryStatus(
  repoPath: string,
  options?: { onProgress?: (percent: number, message: string) => void }
): Promise<RepositoryStatus> {
  const branch = await getCurrentBranch(repoPath, options)
  const statusOutput = await execGit(['status', '--porcelain'], {
    cwd: repoPath,
    onProgress: options?.onProgress,
  })

  const changes = statusOutput
    .trim()
    .split('\n')
    .filter((line) => line.trim())

  return {
    branch,
    isDirty: changes.length > 0,
    stagedChanges: changes.filter((line) => /^[AM]/.test(line)).length,
    unstagedChanges: changes.filter((line) => /^[M?]/.test(line)).length,
  }
}

// =============================================================================
// Log and History Operations
// =============================================================================

/**
 * Get git log with commit statistics
 */
export async function getGitLog(
  repoPath: string,
  options?: {
    author?: string
    limit?: number
    onProgress?: (percent: number, message: string) => void
  }
): Promise<GitCommitStats[]> {
  const { author, limit = 100, onProgress } = options || {}

  const args = ['log', `--max-count=${limit}`, '--pretty=format:%H|%an|%ai|%s', '--name-status']
  if (author) {
    args.push(`--author=${author.replace(/['"]/g, '')}`)
  }

  const output = await execGit(args, { cwd: repoPath, onProgress })
  return parseGitLog(output)
}

function parseGitLog(output: string): GitCommitStats[] {
  return output
    .split('\n')
    .filter((line) => line.includes('|'))
    .map((line) => {
      const [hash, author, dateStr, message] = line.split('|')
      return {
        hash,
        author,
        date: new Date(dateStr),
        message,
        additions: 0,
        deletions: 0,
      }
    })
}

/**
 * Get git authors with commit counts
 */
export async function getGitAuthors(
  repoPath: string,
  options?: { onProgress?: (percent: number, message: string) => void }
): Promise<Map<string, number>> {
  const output = await execGit(['log', '--pretty=format:%an'], {
    cwd: repoPath,
    onProgress: options?.onProgress,
  })

  const authorMap = new Map<string, number>()
  for (const author of output.split('\n').filter((a) => a.trim())) {
    authorMap.set(author, (authorMap.get(author) || 0) + 1)
  }

  return authorMap
}

// =============================================================================
// File Churn Analysis
// =============================================================================

/**
 * Get file churn statistics showing which files change most frequently
 */
export async function getFileChurn(
  repoPath: string,
  options?: { onProgress?: (percent: number, message: string) => void }
): Promise<Map<string, FileChurn>> {
  const output = await execGit(['log', '--pretty=format:%ai', '--name-status', '--diff-filter=ACMRTU'], {
    cwd: repoPath,
    onProgress: options?.onProgress,
  })

  const churnMap = parseChurnOutput(output)
  await enrichChurnWithStats(repoPath, churnMap)

  return churnMap
}

function parseChurnOutput(output: string): Map<string, FileChurn> {
  const churnMap = new Map<string, FileChurn>()
  const lines = output.split('\n')
  let currentDate: Date | null = null

  for (const line of lines) {
    const trimmed = line.trim()
    if (!trimmed) continue

    if (/^\d{4}-\d{2}-\d{2}/.test(trimmed)) {
      currentDate = new Date(trimmed.split(' ')[0])
      continue
    }

    if (currentDate) {
      const [, file] = trimmed.split('\t')
      if (file) {
        updateChurnEntry(churnMap, file, currentDate)
      }
    }
  }

  return churnMap
}

function updateChurnEntry(churnMap: Map<string, FileChurn>, file: string, date: Date): void {
  if (!churnMap.has(file)) {
    churnMap.set(file, {
      file,
      commits: 0,
      additions: 0,
      deletions: 0,
      lastModified: date,
    })
  }

  const entry = churnMap.get(file)!
  entry.commits++
  entry.lastModified = date
}

async function enrichChurnWithStats(repoPath: string, churnMap: Map<string, FileChurn>): Promise<void> {
  for (const [file, churn] of churnMap.entries()) {
    try {
      const diffOutput = await execGit(['log', '-p', '--', file], { cwd: repoPath })
      const { additions, deletions } = countDiffChanges(diffOutput)
      churn.additions = additions
      churn.deletions = deletions
    } catch {
      // Skip detailed stats on error
    }
  }
}

function countDiffChanges(diffOutput: string): { additions: number; deletions: number } {
  let additions = 0
  let deletions = 0

  for (const line of diffOutput.split('\n')) {
    if (line.startsWith('+') && !line.startsWith('+++')) additions++
    if (line.startsWith('-') && !line.startsWith('---')) deletions++
  }

  return { additions, deletions }
}

// =============================================================================
// Abort Controller
// =============================================================================

export interface GitAbortController {
  cancel(): void
}

export function createGitAbortController(): GitAbortController {
  return {
    cancel() {
      // Implementation depends on the actual execution environment
    },
  }
}
