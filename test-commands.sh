#!/bin/bash

# Test suite for Code Cartographer v2.0 - 6 Optimized Commands

set -e

PROJECT_PATH="$(pwd)"
TEST_OUTPUT="test-results.txt"
PASS_COUNT=0
FAIL_COUNT=0

# Initialize output file
cat > $TEST_OUTPUT << EOF
Code Cartographer v2.0 - Command Test Suite
=============================================
Date: $(date)
Project: $PROJECT_PATH

EOF

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Run a single test command
run_test() {
  local name="$1"
  local cmd="$2"
  local output_file="/tmp/carto-test-$$.txt"
  local show_lines="${3:-0}"

  echo -e "${YELLOW}[TEST]${NC} $name"
  echo "[TEST] $name" >> $TEST_OUTPUT

  if $cmd > "$output_file" 2>&1; then
    echo -e "${GREEN}[PASS]${NC} $name"
    echo "[PASS] $name" >> $TEST_OUTPUT
    ((PASS_COUNT++))

    if [ "$show_lines" -gt 0 ]; then
      head -"$show_lines" "$output_file" >> $TEST_OUTPUT
    else
      cat "$output_file" >> $TEST_OUTPUT
    fi
  else
    echo -e "${RED}[FAIL]${NC} $name"
    echo "[FAIL] $name" >> $TEST_OUTPUT
    ((FAIL_COUNT++))
    cat "$output_file" >> $TEST_OUTPUT
  fi

  echo "" >> $TEST_OUTPUT
  rm -f "$output_file"
}

# Core command tests
run_test "carto-map (initialize)" "/carto:carto-map path='.'"
run_test "carto-parse (extract structure)" "/carto:carto-parse"

# Analysis tests
run_test "carto-analyze (full)" "/carto:carto-analyze"
run_test "carto-analyze type=frameworks" "/carto:carto-analyze type=frameworks"
run_test "carto-analyze type=health" "/carto:carto-analyze type=health"

# Search tests
run_test "carto-find (all mode)" "/carto:carto-find 'class'" 20
run_test "carto-find mode=map" "/carto:carto-find '.*' mode=map" 20
run_test "carto-find mode=source" "/carto:carto-find 'function\|class' mode=source" 20

# Visualization tests
run_test "carto-visualize --format=tree" "/carto:carto-visualize --format=tree" 40
run_test "carto-visualize --format=deps" "/carto:carto-visualize --format=deps" 40
run_test "carto-visualize --format=health" "/carto:carto-visualize --format=health"

# Info tests
run_test "carto-info (status)" "/carto:carto-info"
run_test "carto-info --diff" "/carto:carto-info --diff"

# Summary
echo ""
echo "=========================================="
echo "Test Results Summary"
echo "=========================================="
echo ""
echo -e "Passed: ${GREEN}${PASS_COUNT}${NC}"
echo -e "Failed: ${RED}${FAIL_COUNT}${NC}"
echo ""
echo "Full results saved to: $TEST_OUTPUT"
echo ""
