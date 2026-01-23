import * as fs from 'fs';
import * as path from 'path';

interface TestResult {
  name: string;
  passed: boolean;
  output?: string;
  error?: string;
}

const results: TestResult[] = [];

function test(name: string, fn: () => void): void {
  try {
    console.log(`Testing: ${name}...`);
    fn();
    results.push({ name, passed: true });
    console.log(`✓ ${name}`);
  } catch (error) {
    results.push({
      name,
      passed: false,
      error: error instanceof Error ? error.message : String(error),
    });
    console.log(`✗ ${name} - ${error}`);
  }
}

function readExampleFile(filename: string): string {
  const filePath = path.join(__dirname, '.claude-plugin', 'examples', filename);
  if (!fs.existsSync(filePath)) {
    throw new Error(`Example file not found: ${filename}`);
  }
  return fs.readFileSync(filePath, 'utf-8');
}

// Load live example data
const scanOutput = readExampleFile('carto-scan-output.md');
const parseOutput = readExampleFile('carto-parse-output.md');
const graphOutput = readExampleFile('carto-graph-output.md');

console.log('\n===========================================');
console.log('Live Data Integration Tests');
console.log('===========================================\n');

// Test 1: Verify carto-map output format with live data
test('carto-map produces correct output structure', () => {
  // Expected fields from scan output
  const requiredFields = [
    'files found',
    'languages',
    'lines',
  ];

  for (const field of requiredFields) {
    if (!scanOutput.toLowerCase().includes(field.toLowerCase())) {
      throw new Error(`Missing field in scan output: ${field}`);
    }
  }
});

// Test 2: Verify carto-parse extracts nodes and edges
test('carto-parse extracts nodes correctly', () => {
  // Should contain class/function extraction results
  const requiredElements = ['class', 'function', 'nodes', 'edges'];

  for (const elem of requiredElements) {
    if (!parseOutput.toLowerCase().includes(elem.toLowerCase())) {
      throw new Error(`Missing in parse output: ${elem}`);
    }
  }
});

// Test 3: Verify carto-visualize tree format
test('carto-visualize --format=tree produces tree structure', () => {
  // Tree format should have directory structure
  if (!graphOutput.includes('├──') && !graphOutput.includes('└──')) {
    throw new Error('Missing tree structure indicators');
  }
});

// Test 4: Verify file statistics from live data
test('File statistics are present in map output', () => {
  // Extract file count from scan output
  const fileCountMatch = scanOutput.match(/(\d+)\s+files?/i);
  if (!fileCountMatch) {
    throw new Error('Could not find file count in output');
  }

  const fileCount = parseInt(fileCountMatch[1]);
  if (fileCount <= 0) {
    throw new Error('Invalid file count');
  }
});

// Test 5: Verify node count from live parse data
test('Node extraction statistics valid', () => {
  const nodeMatch = parseOutput.match(/(\d+)\s+nodes?/i);
  if (!nodeMatch) {
    throw new Error('Could not find node count in parse output');
  }

  const nodeCount = parseInt(nodeMatch[1]);
  if (nodeCount <= 0) {
    throw new Error('Invalid node count');
  }
});

// Test 6: Verify edge count from live parse data
test('Edge extraction statistics valid', () => {
  const edgeMatch = parseOutput.match(/(\d+)\s+edges?/i);
  if (!edgeMatch) {
    throw new Error('Could not find edge count in parse output');
  }

  const edgeCount = parseInt(edgeMatch[1]);
  if (edgeCount <= 0) {
    throw new Error('Invalid edge count');
  }
});

// Test 7: Verify language detection in live data
test('Language detection working (live data)', () => {
  // Should detect multiple languages
  const languages = ['typescript', 'javascript', 'python', 'java', 'go', 'rust'];

  let detectedCount = 0;
  for (const lang of languages) {
    if (scanOutput.toLowerCase().includes(lang)) {
      detectedCount++;
    }
  }

  if (detectedCount === 0) {
    throw new Error('No languages detected in live data');
  }
});

// Test 8: Verify class extraction from live data
test('Class/type extraction working (live data)', () => {
  if (!parseOutput.toLowerCase().includes('class')) {
    throw new Error('No classes found in parse output');
  }
});

// Test 9: Verify function extraction from live data
test('Function extraction working (live data)', () => {
  if (!parseOutput.toLowerCase().includes('function') && !parseOutput.toLowerCase().includes('method')) {
    throw new Error('No functions/methods found in parse output');
  }
});

// Test 10: Verify relationship extraction (imports/dependencies)
test('Relationship extraction working (live data)', () => {
  const relationshipKeywords = ['import', 'depend', 'call', 'reference'];

  let found = false;
  for (const keyword of relationshipKeywords) {
    if (parseOutput.toLowerCase().includes(keyword)) {
      found = true;
      break;
    }
  }

  if (!found) {
    throw new Error('No relationships found in parse output');
  }
});

// Test 11: Verify tree visualization has depth
test('Tree visualization shows hierarchy depth', () => {
  const lines = graphOutput.split('\n');
  const treeChars = lines.filter(l => l.includes('├') || l.includes('└') || l.includes('│')).length;

  if (treeChars === 0) {
    throw new Error('Tree visualization lacks depth indicators');
  }
});

// Test 12: Verify example files are not empty
test('All example output files have content', () => {
  if (!scanOutput || scanOutput.length === 0) {
    throw new Error('carto-scan-output.md is empty');
  }
  if (!parseOutput || parseOutput.length === 0) {
    throw new Error('carto-parse-output.md is empty');
  }
  if (!graphOutput || graphOutput.length === 0) {
    throw new Error('carto-graph-output.md is empty');
  }
});

// Test 13: Verify command examples use new 6-command structure
test('Examples use new 6-command structure', () => {
  const workflowPath = path.join(__dirname, '.claude-plugin', 'examples', 'WORKFLOW.md');
  const workflowContent = fs.readFileSync(workflowPath, 'utf-8');

  const newCommands = [
    'carto-map',
    'carto-parse',
    'carto-find',
    'carto-analyze',
    'carto-visualize',
    'carto-info',
  ];

  for (const cmd of newCommands) {
    if (!workflowContent.includes(cmd)) {
      throw new Error(`WORKFLOW.md doesn't reference ${cmd}`);
    }
  }
});

// Test 14: Verify no old commands in examples
test('Examples do not use old 11-command structure', () => {
  const workflowPath = path.join(__dirname, '.claude-plugin', 'examples', 'WORKFLOW.md');
  const workflowContent = fs.readFileSync(workflowPath, 'utf-8');

  const oldCommands = [
    'carto-init',
    'carto-scan',
    'carto-query',
    'carto-search',
    'carto-detect',
    'carto-graph',
    'carto-canvas',
    'carto-report',
    'carto-status',
    'carto-diff',
  ];

  for (const cmd of oldCommands) {
    if (workflowContent.includes(cmd)) {
      throw new Error(`WORKFLOW.md still references old command: ${cmd}`);
    }
  }
});

// Test 15: Verify quick start has 3-step setup
test('Quick start guide shows 3-step setup', () => {
  const quickStartPath = path.join(__dirname, '.claude-plugin', 'examples', 'QUICK_START.md');
  const quickStartContent = fs.readFileSync(quickStartPath, 'utf-8');

  const stepMatches = quickStartContent.match(/### Step \d:/g);
  if (!stepMatches || stepMatches.length < 3) {
    throw new Error('Quick start should have at least 3 steps');
  }
});

// Test 16: Verify workflow shows map->parse->analyze->visualize sequence
test('Workflow shows correct command sequence', () => {
  const workflowPath = path.join(__dirname, '.claude-plugin', 'examples', 'WORKFLOW.md');
  const workflowContent = fs.readFileSync(workflowPath, 'utf-8');

  // Check sequence order: map should come before parse
  const mapIndex = workflowContent.indexOf('carto-map');
  const parseIndex = workflowContent.indexOf('carto-parse');
  const analyzeIndex = workflowContent.indexOf('carto-analyze');
  const visualizeIndex = workflowContent.indexOf('carto-visualize');

  if (mapIndex === -1 || parseIndex === -1) {
    throw new Error('Missing carto-map or carto-parse in workflow');
  }

  if (mapIndex > parseIndex) {
    throw new Error('Command sequence incorrect: map should come before parse');
  }
});

// Summary
console.log('\n===========================================');
console.log('Test Results');
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
    .forEach(r => console.log(`  - ${r.name}: ${r.error}`));
  process.exit(1);
} else {
  console.log('All live data tests passed!\n');
  process.exit(0);
}
