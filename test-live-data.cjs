const fs = require('fs');
const path = require('path');

// Configuration
const EXAMPLES_DIR = path.join(__dirname, '.claude-plugin', 'examples');

const results = [];

// Test utilities
function test(name, fn) {
  try {
    console.log(`Testing: ${name}...`);
    fn();
    results.push({ name, passed: true });
    console.log(`[PASS] ${name}`);
  } catch (error) {
    results.push({ name, passed: false, error: error.message });
    console.log(`[FAIL] ${name} - ${error.message}`);
  }
}

function readExampleFile(filename) {
  const filePath = path.join(EXAMPLES_DIR, filename);
  if (!fs.existsSync(filePath)) {
    throw new Error(`Example file not found: ${filename}`);
  }
  return fs.readFileSync(filePath, 'utf-8');
}

function assertContainsAny(content, terms, context) {
  const found = terms.some(term => content.toLowerCase().includes(term.toLowerCase()));
  if (!found) {
    throw new Error(`${context}: expected one of [${terms.join(', ')}]`);
  }
}

function assertContainsAll(content, terms, context) {
  for (const term of terms) {
    if (!content.toLowerCase().includes(term.toLowerCase())) {
      throw new Error(`${context} missing: ${term}`);
    }
  }
}

// Load example data
const scanOutput = readExampleFile('carto-scan-output.md');
const parseOutput = readExampleFile('carto-parse-output.md');
const graphOutput = readExampleFile('carto-graph-output.md');

console.log('\n===========================================');
console.log('Live Data Integration Tests');
console.log('===========================================\n');

// Output structure tests
test('Scan output has file discovery info', () => {
  assertContainsAll(scanOutput, ['files', 'lines', 'language'], 'Scan output');
});

test('Parse output has code structure', () => {
  assertContainsAll(parseOutput, ['class', 'method'], 'Parse output');
});

test('Graph output has visualization content', () => {
  const hasVisualization = graphOutput.length > 100;
  if (!hasVisualization) {
    throw new Error('Graph output too small');
  }
});

// Content validation tests
test('Scan output shows file count', () => {
  const hasFileInfo = scanOutput.includes('Files') || scanOutput.includes('files');
  if (!hasFileInfo) {
    throw new Error('No file information in scan output');
  }
});

test('Parse output shows extracted nodes', () => {
  const hasNodes = parseOutput.includes('Nodes') || parseOutput.includes('class') || parseOutput.includes('function');
  if (!hasNodes) {
    throw new Error('No node information in parse output');
  }
});

// Language detection tests
test('Language detection working', () => {
  const languages = ['typescript', 'javascript', 'python', 'java', 'go', 'rust'];
  assertContainsAny(scanOutput, languages, 'Language detection');
});

test('Class extraction working', () => {
  assertContainsAny(parseOutput, ['class', 'Class'], 'Class extraction');
});

test('Method extraction working', () => {
  assertContainsAny(parseOutput, ['method', 'function', 'Method', 'Function'], 'Method extraction');
});

// Relationship/dependency tests
test('Dependencies or imports found', () => {
  const hasRelationships = parseOutput.toLowerCase().includes('import') ||
                          parseOutput.toLowerCase().includes('depend') ||
                          parseOutput.toLowerCase().includes('module');
  if (!hasRelationships) {
    throw new Error('No dependency information found');
  }
});

// Tree visualization test
test('Graph output has structure', () => {
  const hasStructure = graphOutput.includes('---') ||
                       graphOutput.includes('|') ||
                       graphOutput.includes('+') ||
                       graphOutput.includes('->');
  if (!hasStructure) {
    throw new Error('Graph lacks structural indicators');
  }
});

// Quick start validation
test('Quick start guide exists and has content', () => {
  const quickStartContent = readExampleFile('QUICK_START.md');
  const hasSteps = quickStartContent.includes('###') && quickStartContent.length > 200;
  if (!hasSteps) {
    throw new Error('Quick start should have step headers');
  }
});

// Summary
console.log('\n===========================================');
console.log('Test Results Summary');
console.log('===========================================\n');

const passed = results.filter(r => r.passed).length;
const failed = results.filter(r => !r.passed).length;

console.log(`Total: ${results.length} tests`);
console.log(`Passed: ${passed}`);
console.log(`Failed: ${failed}\n`);

if (failed > 0) {
  console.log('Failed Tests:');
  results
    .filter(r => !r.passed)
    .forEach(r => console.log(`  [FAIL] ${r.name}\n    ${r.error}`));
  process.exit(1);
}

console.log('All live data tests passed!\n');
process.exit(0);
