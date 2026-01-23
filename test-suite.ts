import * as fs from 'fs';
import * as path from 'path';

interface TestResult {
  name: string;
  passed: boolean;
  output?: string;
  error?: string;
  duration: number;
}

const results: TestResult[] = [];

async function runTest(
  name: string,
  testFn: () => Promise<void>
): Promise<void> {
  const start = Date.now();
  try {
    console.log(`Running: ${name}...`);
    await testFn();
    const duration = Date.now() - start;
    results.push({ name, passed: true, duration });
    console.log(`✓ ${name} (${duration}ms)`);
  } catch (error) {
    const duration = Date.now() - start;
    results.push({
      name,
      passed: false,
      error: error instanceof Error ? error.message : String(error),
      duration,
    });
    console.log(`✗ ${name} - ${error}`);
  }
}

// Test 1: Verify all 6 command files exist
async function testCommandFilesExist(): Promise<void> {
  const commands = [
    'carto-map.md',
    'carto-parse.md',
    'carto-find.md',
    'carto-analyze.md',
    'carto-visualize.md',
    'carto-info.md',
  ];

  for (const cmd of commands) {
    const filePath = path.join(
      __dirname,
      '.claude-plugin',
      'commands',
      cmd
    );
    if (!fs.existsSync(filePath)) {
      throw new Error(`Command file missing: ${cmd}`);
    }
  }
}

// Test 2: Verify no old commands exist
async function testOldCommandsDeleted(): Promise<void> {
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

  for (const cmd of oldCommands) {
    const filePath = path.join(
      __dirname,
      '.claude-plugin',
      'commands',
      cmd
    );
    if (fs.existsSync(filePath)) {
      throw new Error(`Old command still exists: ${cmd}`);
    }
  }
}

// Test 3: Verify YAML schema for all commands
async function testCommandYamlSchema(): Promise<void> {
  const commands = [
    'carto-map.md',
    'carto-parse.md',
    'carto-find.md',
    'carto-analyze.md',
    'carto-visualize.md',
    'carto-info.md',
  ];

  for (const cmd of commands) {
    const filePath = path.join(
      __dirname,
      '.claude-plugin',
      'commands',
      cmd
    );
    const content = fs.readFileSync(filePath, 'utf-8');

    // Check for YAML frontmatter
    if (!content.startsWith('---')) {
      throw new Error(`${cmd} missing YAML frontmatter`);
    }

    // Check for description
    if (!content.includes('description:')) {
      throw new Error(`${cmd} missing description field`);
    }

    // Check for parameters field (can be empty)
    if (!content.includes('parameters:')) {
      throw new Error(`${cmd} missing parameters field`);
    }
  }
}

// Test 4: Verify documentation sections
async function testCommandDocumentation(): Promise<void> {
  const commands = [
    'carto-map.md',
    'carto-parse.md',
    'carto-find.md',
    'carto-analyze.md',
    'carto-visualize.md',
    'carto-info.md',
  ];

  const requiredSections = ['Usage', 'Parameters', 'Output', 'Workflow'];

  for (const cmd of commands) {
    const filePath = path.join(
      __dirname,
      '.claude-plugin',
      'commands',
      cmd
    );
    const content = fs.readFileSync(filePath, 'utf-8').toLowerCase();

    for (const section of requiredSections) {
      if (!content.includes(`## ${section.toLowerCase()}`)) {
        throw new Error(
          `${cmd} missing ${section} section`
        );
      }
    }
  }
}

// Test 5: Verify migration guide exists
async function testMigrationGuideExists(): Promise<void> {
  const migrationPath = path.join(
    __dirname,
    '.claude-plugin',
    'MIGRATION_GUIDE.md'
  );

  if (!fs.existsSync(migrationPath)) {
    throw new Error('MIGRATION_GUIDE.md not found');
  }

  const content = fs.readFileSync(migrationPath, 'utf-8');

  // Check for migration table
  if (!content.includes('Old Commands') || !content.includes('New Command')) {
    throw new Error('Migration guide missing command mapping table');
  }

  // Check for before/after examples
  if (!content.match(/Before.*After/is)) {
    throw new Error('Migration guide missing before/after examples');
  }
}

// Test 6: Verify subagent updated
async function testSubagentUpdated(): Promise<void> {
  const subagentPath = path.join(
    __dirname,
    '.claude-plugin',
    'subagents',
    'code-analyzer.md'
  );

  const content = fs.readFileSync(subagentPath, 'utf-8');

  // Check for new command references
  const newCommands = ['carto-map', 'carto-find', 'carto-visualize', 'carto-analyze'];

  for (const cmd of newCommands) {
    if (!content.includes(cmd)) {
      throw new Error(`Subagent missing reference to ${cmd}`);
    }
  }

  // Check that old command references are removed
  const oldCommands = ['carto-init', 'carto-scan', 'carto-query', 'carto-detect'];

  for (const cmd of oldCommands) {
    if (content.includes(cmd)) {
      throw new Error(`Subagent still contains old command ${cmd}`);
    }
  }
}

// Test 7: Verify example files updated
async function testExamplesUpdated(): Promise<void> {
  const exampleFiles = [
    'WORKFLOW.md',
    'LIVE_ANALYSIS.md',
    'QUICK_START_V2.md',
  ];

  for (const file of exampleFiles) {
    const filePath = path.join(
      __dirname,
      '.claude-plugin',
      'examples',
      file
    );

    if (!fs.existsSync(filePath)) {
      throw new Error(`Example file missing: ${file}`);
    }

    const content = fs.readFileSync(filePath, 'utf-8');

    // Check for new command usage
    const newCommands = ['carto-map', 'carto-parse', 'carto-find', 'carto-visualize', 'carto-analyze', 'carto-info'];
    let foundCommands = 0;

    for (const cmd of newCommands) {
      if (content.includes(cmd)) {
        foundCommands++;
      }
    }

    if (foundCommands === 0) {
      throw new Error(`${file} doesn't reference any new commands`);
    }
  }
}

// Test 8: Verify plugin.json is valid
async function testPluginJsonValid(): Promise<void> {
  const pluginPath = path.join(
    __dirname,
    '.claude-plugin',
    'plugin.json'
  );

  const content = fs.readFileSync(pluginPath, 'utf-8');

  try {
    const json = JSON.parse(content);

    if (!json.commands) {
      throw new Error('plugin.json missing commands field');
    }

    if (!json.subagents) {
      throw new Error('plugin.json missing subagents field');
    }
  } catch (error) {
    throw new Error(`plugin.json is invalid JSON: ${error}`);
  }
}

// Test 9: Verify consolidation is complete (6 commands only)
async function testCommandConsolidation(): Promise<void> {
  const commandsDir = path.join(
    __dirname,
    '.claude-plugin',
    'commands'
  );

  const files = fs.readdirSync(commandsDir).filter(f => f.endsWith('.md'));

  if (files.length !== 6) {
    throw new Error(
      `Expected 6 command files, found ${files.length}: ${files.join(', ')}`
    );
  }

  const expectedCommands = [
    'carto-map',
    'carto-parse',
    'carto-find',
    'carto-analyze',
    'carto-visualize',
    'carto-info',
  ];

  const foundCommands = files.map(f => f.replace('.md', ''));

  for (const cmd of expectedCommands) {
    if (!foundCommands.includes(cmd)) {
      throw new Error(`Expected command not found: ${cmd}`);
    }
  }
}

// Test 10: Verify workflow consistency
async function testWorkflowConsistency(): Promise<void> {
  const commands = {
    'carto-map.md': 'Step 1',
    'carto-parse.md': 'Step 2',
    'carto-analyze.md': 'Step 3',
    'carto-visualize.md': 'Step 4',
    'carto-find.md': 'Step 5',
    'carto-info.md': 'Step 6',
  };

  // Check that each command references the workflow
  for (const [cmd] of Object.entries(commands)) {
    const filePath = path.join(
      __dirname,
      '.claude-plugin',
      'commands',
      cmd
    );
    const content = fs.readFileSync(filePath, 'utf-8');

    if (!content.includes('Workflow') && !content.includes('workflow')) {
      throw new Error(`${cmd} missing workflow reference`);
    }
  }
}

// Main test runner
async function runAllTests(): Promise<void> {
  console.log('\n===========================================');
  console.log('Code Cartographer v2.0 - Test Suite');
  console.log('===========================================\n');

  await runTest('Command files exist (6 commands)', testCommandFilesExist);
  await runTest('Old commands deleted (10 removed)', testOldCommandsDeleted);
  await runTest('YAML schema validation', testCommandYamlSchema);
  await runTest('Documentation sections present', testCommandDocumentation);
  await runTest('Migration guide exists', testMigrationGuideExists);
  await runTest('Subagent updated', testSubagentUpdated);
  await runTest('Examples updated', testExamplesUpdated);
  await runTest('plugin.json valid', testPluginJsonValid);
  await runTest('Command consolidation (11→6)', testCommandConsolidation);
  await runTest('Workflow consistency', testWorkflowConsistency);

  console.log('\n===========================================');
  console.log('Test Results Summary');
  console.log('===========================================\n');

  const passed = results.filter(r => r.passed).length;
  const failed = results.filter(r => !r.passed).length;

  console.log(`Total: ${results.length} tests`);
  console.log(`Passed: ${passed}`);
  console.log(`Failed: ${failed}`);
  console.log(`Total Duration: ${results.reduce((sum, r) => sum + r.duration, 0)}ms\n`);

  if (failed > 0) {
    console.log('Failed Tests:');
    results
      .filter(r => !r.passed)
      .forEach(r => console.log(`  - ${r.name}: ${r.error}`));
    console.log('');
    process.exit(1);
  } else {
    console.log('All tests passed!');
  }
}

runAllTests().catch(error => {
  console.error('Test suite error:', error);
  process.exit(1);
});
