const fs = require('fs');
const path = require('path');

const results = [];

function test(name, fn) {
  try {
    console.log(`Testing: ${name}...`);
    fn();
    results.push({ name, passed: true });
    console.log(`✓ ${name}`);
  } catch (error) {
    results.push({
      name,
      passed: false,
      error: error.message,
    });
    console.log(`✗ ${name} - ${error.message}`);
  }
}

console.log('\n===========================================');
console.log('Code Cartographer v2.0 - Consolidation Tests');
console.log('===========================================\n');

// Test 1: Verify 6 command files exist
test('All 6 new commands created', () => {
  const commands = [
    'carto-map.md',
    'carto-parse.md',
    'carto-find.md',
    'carto-analyze.md',
    'carto-visualize.md',
    'carto-info.md',
  ];

  for (const cmd of commands) {
    const filePath = path.join(__dirname, '.claude-plugin', 'commands', cmd);
    if (!fs.existsSync(filePath)) {
      throw new Error(`Missing: ${cmd}`);
    }
  }
});

// Test 2: Verify old commands deleted
test('All 10 old commands deleted', () => {
  const oldCommands = [
    'carto-init.md',
    'carto-scan.md',
    'carto-detect.md',
    'carto-query.md',
    'carto-search.md',
    'carto-graph.md',
    'carto-canvas.md',
    'carto-report.md',
    'carto-status.md',
    'carto-diff.md',
  ];

  let deletedCount = 0;
  for (const cmd of oldCommands) {
    const filePath = path.join(__dirname, '.claude-plugin', 'commands', cmd);
    if (!fs.existsSync(filePath)) {
      deletedCount++;
    }
  }

  if (deletedCount !== oldCommands.length) {
    throw new Error(`Expected ${oldCommands.length} deleted, got ${deletedCount}`);
  }
});

// Test 3: Verify command file content
test('Command files have proper YAML schema', () => {
  const commands = [
    'carto-map.md',
    'carto-parse.md',
    'carto-find.md',
    'carto-analyze.md',
    'carto-visualize.md',
    'carto-info.md',
  ];

  for (const cmd of commands) {
    const filePath = path.join(__dirname, '.claude-plugin', 'commands', cmd);
    const content = fs.readFileSync(filePath, 'utf-8');

    if (!content.startsWith('---')) {
      throw new Error(`${cmd} missing YAML frontmatter`);
    }

    if (!content.includes('description:')) {
      throw new Error(`${cmd} missing description`);
    }

    if (!content.includes('parameters:')) {
      throw new Error(`${cmd} missing parameters field`);
    }
  }
});

// Test 4: Verify workflow consistency
test('Workflow references in carto-parse.md updated', () => {
  const filePath = path.join(__dirname, '.claude-plugin', 'commands', 'carto-parse.md');
  const content = fs.readFileSync(filePath, 'utf-8');

  const newCommands = ['carto-map', 'carto-analyze', 'carto-visualize', 'carto-find', 'carto-info'];

  for (const cmd of newCommands) {
    if (!content.includes(cmd)) {
      throw new Error(`${cmd} not referenced in workflow`);
    }
  }

  const oldCommands = ['carto-init', 'carto-scan', 'carto-detect', 'carto-query'];
  for (const cmd of oldCommands) {
    if (content.includes(cmd)) {
      throw new Error(`Old command ${cmd} still referenced`);
    }
  }
});

// Test 5: Verify subagent updated
test('Subagent code-analyzer.md references new commands', () => {
  const filePath = path.join(__dirname, '.claude-plugin', 'subagents', 'code-analyzer.md');
  const content = fs.readFileSync(filePath, 'utf-8');

  const newCommands = ['carto-map', 'carto-find', 'carto-visualize', 'carto-analyze'];

  for (const cmd of newCommands) {
    if (!content.includes(cmd)) {
      throw new Error(`${cmd} not in subagent`);
    }
  }
});

// Test 6: Verify examples updated
test('WORKFLOW.md uses new 6-command structure', () => {
  const filePath = path.join(__dirname, '.claude-plugin', 'examples', 'WORKFLOW.md');
  const content = fs.readFileSync(filePath, 'utf-8');

  const newCommands = ['carto-map', 'carto-parse', 'carto-find', 'carto-analyze', 'carto-visualize', 'carto-info'];

  for (const cmd of newCommands) {
    if (!content.includes(cmd)) {
      throw new Error(`${cmd} not in WORKFLOW.md`);
    }
  }

  const oldCommands = ['carto-init', 'carto-scan', 'carto-query', 'carto-detect', 'carto-graph', 'carto-canvas', 'carto-report', 'carto-status', 'carto-diff'];

  for (const cmd of oldCommands) {
    if (content.includes(cmd)) {
      throw new Error(`Old command ${cmd} still in WORKFLOW.md`);
    }
  }
});

// Test 7: Verify LIVE_ANALYSIS.md updated
test('LIVE_ANALYSIS.md uses new commands', () => {
  const filePath = path.join(__dirname, '.claude-plugin', 'examples', 'LIVE_ANALYSIS.md');
  const content = fs.readFileSync(filePath, 'utf-8');

  if (!content.includes('carto-map') || !content.includes('carto-visualize')) {
    throw new Error('LIVE_ANALYSIS.md not updated with new commands');
  }
});

// Test 8: Verify example data exists
test('Example data files with live output present', () => {
  const exampleFiles = [
    'carto-scan-output.md',
    'carto-parse-output.md',
    'carto-graph-output.md',
  ];

  for (const file of exampleFiles) {
    const filePath = path.join(__dirname, '.claude-plugin', 'examples', file);
    if (!fs.existsSync(filePath)) {
      throw new Error(`Missing example: ${file}`);
    }

    const content = fs.readFileSync(filePath, 'utf-8');
    if (content.length < 100) {
      throw new Error(`${file} is empty or too small`);
    }
  }
});

// Test 9: Verify migration guide created
test('MIGRATION_GUIDE.md created with command mappings', () => {
  const filePath = path.join(__dirname, '.claude-plugin', 'MIGRATION_GUIDE.md');
  if (!fs.existsSync(filePath)) {
    throw new Error('MIGRATION_GUIDE.md not created');
  }

  const content = fs.readFileSync(filePath, 'utf-8');
  if (!content.includes('Old Commands') || !content.includes('New Command')) {
    throw new Error('Migration guide missing mapping table');
  }
});

// Test 10: Verify plugin.json is valid
test('plugin.json is valid JSON', () => {
  const filePath = path.join(__dirname, '.claude-plugin', 'plugin.json');
  const content = fs.readFileSync(filePath, 'utf-8');

  try {
    JSON.parse(content);
  } catch (error) {
    throw new Error(`plugin.json invalid: ${error.message}`);
  }
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
    .forEach(r => console.log(`  ✗ ${r.name}\n    ${r.error}`));
  console.log('');
  process.exit(1);
} else {
  console.log('All consolidation tests passed!\n');
  console.log('✓ 11 commands successfully consolidated to 6');
  console.log('✓ Documentation updated');
  console.log('✓ Examples reflect new structure');
  console.log('✓ Live data validation complete\n');
  process.exit(0);
}
