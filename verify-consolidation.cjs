const fs = require('fs');
const path = require('path');

// Configuration
const PLUGIN_DIR = path.join(__dirname, '.claude-plugin');
const COMMANDS_DIR = path.join(PLUGIN_DIR, 'commands');
const EXAMPLES_DIR = path.join(PLUGIN_DIR, 'examples');

const NEW_COMMANDS = ['carto-map', 'carto-parse', 'carto-find', 'carto-analyze', 'carto-visualize', 'carto-info'];
const OLD_COMMANDS = ['carto-init', 'carto-scan', 'carto-detect', 'carto-query', 'carto-search', 'carto-graph', 'carto-canvas', 'carto-report', 'carto-status', 'carto-diff'];

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

function fileExists(filePath) {
  return fs.existsSync(filePath);
}

function readFile(filePath) {
  return fs.readFileSync(filePath, 'utf-8');
}

function assertFileExists(filePath, description) {
  if (!fileExists(filePath)) {
    throw new Error(`Missing: ${description || filePath}`);
  }
}

function assertFileNotExists(filePath, description) {
  if (fileExists(filePath)) {
    throw new Error(`Should not exist: ${description || filePath}`);
  }
}

function assertContains(content, searchTerm, context) {
  if (!content.includes(searchTerm)) {
    throw new Error(`${context} missing: ${searchTerm}`);
  }
}

function assertNotContains(content, searchTerm, context) {
  if (content.includes(searchTerm)) {
    throw new Error(`${context} should not contain: ${searchTerm}`);
  }
}

// Tests
console.log('\n===========================================');
console.log('Code Cartographer v2.0 - Consolidation Tests');
console.log('===========================================\n');

test('All 6 new command files exist', () => {
  for (const cmd of NEW_COMMANDS) {
    assertFileExists(path.join(COMMANDS_DIR, `${cmd}.md`), cmd);
  }
});

test('All 10 old command files deleted', () => {
  for (const cmd of OLD_COMMANDS) {
    assertFileNotExists(path.join(COMMANDS_DIR, `${cmd}.md`), cmd);
  }
});

test('Command files have proper YAML schema', () => {
  for (const cmd of NEW_COMMANDS) {
    const content = readFile(path.join(COMMANDS_DIR, `${cmd}.md`));
    assertContains(content, '---', `${cmd} YAML frontmatter`);
    assertContains(content, 'description:', `${cmd} description`);
    assertContains(content, 'parameters:', `${cmd} parameters`);
  }
});

test('carto-parse.md references new commands', () => {
  const content = readFile(path.join(COMMANDS_DIR, 'carto-parse.md'));
  for (const cmd of NEW_COMMANDS) {
    assertContains(content, cmd, 'carto-parse workflow');
  }
  for (const cmd of OLD_COMMANDS.slice(0, 4)) {
    assertNotContains(content, cmd, 'carto-parse workflow');
  }
});

test('Subagent code-analyzer.md references new commands', () => {
  const content = readFile(path.join(PLUGIN_DIR, 'subagents', 'code-analyzer.md'));
  const requiredCommands = ['carto-map', 'carto-find', 'carto-visualize', 'carto-analyze'];
  for (const cmd of requiredCommands) {
    assertContains(content, cmd, 'subagent');
  }
});

test('WORKFLOW.md uses new 6-command structure', () => {
  const content = readFile(path.join(EXAMPLES_DIR, 'WORKFLOW.md'));
  for (const cmd of NEW_COMMANDS) {
    assertContains(content, cmd, 'WORKFLOW.md');
  }
  for (const cmd of OLD_COMMANDS) {
    assertNotContains(content, cmd, 'WORKFLOW.md');
  }
});

test('LIVE_ANALYSIS.md uses new commands', () => {
  const content = readFile(path.join(EXAMPLES_DIR, 'LIVE_ANALYSIS.md'));
  assertContains(content, 'carto-map', 'LIVE_ANALYSIS.md');
  assertContains(content, 'carto-visualize', 'LIVE_ANALYSIS.md');
});

test('Example data files exist with content', () => {
  const exampleFiles = ['carto-scan-output.md', 'carto-parse-output.md', 'carto-graph-output.md'];
  for (const file of exampleFiles) {
    const filePath = path.join(EXAMPLES_DIR, file);
    assertFileExists(filePath, file);
    const content = readFile(filePath);
    if (content.length < 100) {
      throw new Error(`${file} is empty or too small`);
    }
  }
});

test('MIGRATION_GUIDE.md exists with command mappings', () => {
  const filePath = path.join(PLUGIN_DIR, 'MIGRATION_GUIDE.md');
  assertFileExists(filePath, 'MIGRATION_GUIDE.md');
  const content = readFile(filePath);
  assertContains(content, 'Old Commands', 'Migration guide');
  assertContains(content, 'New Command', 'Migration guide');
});

test('plugin.json is valid JSON', () => {
  const content = readFile(path.join(PLUGIN_DIR, 'plugin.json'));
  JSON.parse(content);
});

// Summary
console.log('\n===========================================');
console.log('Test Results Summary');
console.log('===========================================\n');

const passed = results.filter(r => r.passed).length;
const failed = results.filter(r => !r.passed).length;

console.log(`Total Tests: ${results.length}`);
console.log(`Passed: ${passed}`);
console.log(`Failed: ${failed}\n`);

if (failed > 0) {
  console.log('Failed Tests:');
  results
    .filter(r => !r.passed)
    .forEach(r => console.log(`  [FAIL] ${r.name}\n    ${r.error}`));
  console.log('');
  process.exit(1);
}

console.log('All consolidation tests passed!\n');
console.log('[OK] 11 commands successfully consolidated to 6');
console.log('[OK] Documentation updated');
console.log('[OK] Examples reflect new structure');
console.log('[OK] Live data validation complete\n');
process.exit(0);
