import { SecurityIssue, Severity } from '@models/types'
import { CodeMap } from '@models/codeMap'

export class SecurityScanner {
  private patterns: Map<string, { regex: RegExp; severity: Severity; message: string; suggestion?: string }> = new Map()

  constructor() {
    this.initializePatterns()
  }

  private initializePatterns(): void {
    // Hardcoded secrets
    this.patterns.set('hardcoded_secret', {
      regex:
        /(?:password|apikey|api_key|secret|token|auth|credential)\s*[:=]\s*['"]?([a-zA-Z0-9_\-]+)['"]?/gi,
      severity: 'critical',
      message: 'Hardcoded secret or credentials detected',
      suggestion: 'Move secrets to environment variables or a secure vault',
    })

    // AWS access keys
    this.patterns.set('aws_key', {
      regex: /AKIA[0-9A-Z]{16}/g,
      severity: 'critical',
      message: 'AWS Access Key ID detected',
      suggestion: 'Rotate the key immediately and use IAM roles instead',
    })

    // Private keys
    this.patterns.set('private_key', {
      regex: /-----BEGIN\s+(?:RSA\s+)?PRIVATE\s+KEY-----/gi,
      severity: 'critical',
      message: 'Private key detected in code',
      suggestion: 'Remove the key and add to .gitignore',
    })

    // SQL injection vulnerabilities
    this.patterns.set('sql_injection', {
      regex: /(?:SELECT|INSERT|UPDATE|DELETE).*?['"]\s*\+\s*\w+|query\s*\(\s*["'].*?\$|execute\s*\(\s*["'].*?\$|sql\s*=\s*["'].*?[+]/gi,
      severity: 'high',
      message: 'Potential SQL injection vulnerability',
      suggestion: 'Use parameterized queries or prepared statements',
    })

    // eval() usage
    this.patterns.set('eval', {
      regex: /\b(?:eval|Function|exec|execute)\s*\(\s*[^)]*(?:variable|user|input|request|param|data)[^)]*\)/gi,
      severity: 'high',
      message: 'Dangerous eval() or equivalent detected',
      suggestion: 'Avoid eval(). Use JSON.parse() or safer alternatives',
    })

    // innerHTML/dangerouslySetInnerHTML
    this.patterns.set('xss', {
      regex: /innerHTML\s*=|dangerouslySetInnerHTML|\.html\s*\(|@Html\.|html\s*\(|\.setHtml|\.append\s*\(\s*["']\s*<|\.prepend\s*\(\s*["']\s*</gi,
      severity: 'high',
      message: 'Potential XSS vulnerability (innerHTML/dangerouslySetInnerHTML)',
      suggestion: 'Use textContent instead of innerHTML, or sanitize user input',
    })

    // Command execution
    this.patterns.set('command_execution', {
      regex: /(?:system|exec|shell_exec|subprocess|Runtime\.getRuntime|ProcessBuilder|popen|os\.system|os\.popen|shell=True)\s*\(\s*[^)]*(?:input|user|param|data|request)[^)]*\)/gi,
      severity: 'high',
      message: 'Potential command execution vulnerability',
      suggestion: 'Validate and sanitize all user inputs before passing to system commands',
    })

    // Debug statements
    this.patterns.set('debug_statement', {
      regex: /console\.log\s*\(\s*(?:password|secret|token|key|auth|credential)/gi,
      severity: 'high',
      message: 'Sensitive information logged to console',
      suggestion: 'Remove debug statements from production code',
    })

    // TODO/FIXME with security implications
    this.patterns.set('security_todo', {
      regex: /\/\/\s*(?:TODO|FIXME|HACK|XXX|BUG).*?(?:security|auth|permission|access|secret|password|token|validate|sanitize)/gi,
      severity: 'medium',
      message: 'Unresolved security TODO/FIXME comment',
      suggestion: 'Address the security concern mentioned in the TODO',
    })

    // Weak cryptography
    this.patterns.set('weak_crypto', {
      regex: /MD5|SHA1|DES(?!C)|RC4|MD4|RIPEMD|CRC32|CRC16/gi,
      severity: 'high',
      message: 'Weak or deprecated cryptographic algorithm detected',
      suggestion: 'Use SHA-256 or stronger algorithms like SHA-3',
    })
  }

  scan(code: string): SecurityIssue[] {
    const issues: SecurityIssue[] = []
    const lines = code.split('\n')

    for (let patternName of this.patterns.keys()) {
      const { regex, severity, message, suggestion } = this.patterns.get(patternName)!
      let match
      const globalRegex = new RegExp(regex.source, regex.flags + (regex.flags.includes('g') ? '' : 'g'))

      while ((match = globalRegex.exec(code)) !== null) {
        const lineNum = code.substring(0, match.index).split('\n').length - 1
        const lineStart = code.lastIndexOf('\n', match.index) + 1
        const colNum = match.index - lineStart

        issues.push({
          id: `${patternName}:${lineNum}:${colNum}`,
          type: this.getIssueType(patternName),
          severity,
          message,
          location: {
            start: { line: lineNum, column: colNum },
            end: { line: lineNum, column: colNum + match[0].length },
          },
          suggestion,
          code: match[0].substring(0, 50),
        })
      }
    }

    return issues
  }

  scanCodeMap(codeMap: CodeMap): SecurityIssue[] {
    const issues: SecurityIssue[] = []

    for (const file of codeMap.getFiles()) {
      // Note: In a full implementation, we'd read the actual file content
      // For now, we just collect potential issues
    }

    return issues
  }

  private getIssueType(
    patternName: string
  ): 'secret' | 'sql_injection' | 'xss' | 'eval' | 'command_execution' | 'other' {
    const mapping: Record<string, SecurityIssue['type']> = {
      hardcoded_secret: 'secret',
      aws_key: 'secret',
      private_key: 'secret',
      sql_injection: 'sql_injection',
      eval: 'eval',
      xss: 'xss',
      command_execution: 'command_execution',
    }
    return mapping[patternName] || 'other'
  }
}
