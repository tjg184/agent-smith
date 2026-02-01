#!/bin/bash

# verify-test-separation.sh
# Verifies that unit tests and integration tests are properly separated
# Story-002: Unit tests should remain separate from integration tests

set -e

echo "========================================"
echo "Verifying Test Separation (Story-002)"
echo "========================================"
echo ""

# Clean test cache for accurate timing
echo "1. Cleaning test cache..."
go clean -testcache
echo "   ✓ Cache cleared"
echo ""

# Test 1: Verify unit tests run without integration tests
echo "2. Running unit tests (should be fast)..."
UNIT_START=$(date +%s)
go test ./... > /dev/null 2>&1
UNIT_END=$(date +%s)
UNIT_TIME=$((UNIT_END - UNIT_START))
echo "   ✓ Unit tests completed in ${UNIT_TIME}s"
echo ""

# Test 2: Verify integration tests are in separate directory with build tags
echo "3. Verifying integration test structure..."
INTEGRATION_FILES=$(find ./tests/integration -name "*_integration_test.go" 2>/dev/null | wc -l | tr -d ' ')
if [ "$INTEGRATION_FILES" -eq 0 ]; then
    echo "   ✗ ERROR: No integration test files found"
    exit 1
fi
echo "   ✓ Found ${INTEGRATION_FILES} integration test files in tests/integration/"

# Verify all integration tests have build tags
MISSING_TAGS=0
for file in ./tests/integration/*_integration_test.go; do
    if ! head -1 "$file" | grep -q "//go:build integration"; then
        echo "   ✗ ERROR: $file missing build tag"
        MISSING_TAGS=$((MISSING_TAGS + 1))
    fi
done

if [ "$MISSING_TAGS" -gt 0 ]; then
    echo "   ✗ ERROR: ${MISSING_TAGS} files missing build tags"
    exit 1
fi
echo "   ✓ All integration tests have proper build tags"
echo ""

# Test 3: Verify integration tests run with tag
echo "4. Running integration tests (should be slower)..."
INTEGRATION_START=$(date +%s)
go test -tags=integration ./tests/integration/... > /dev/null 2>&1
INTEGRATION_END=$(date +%s)
INTEGRATION_TIME=$((INTEGRATION_END - INTEGRATION_START))
echo "   ✓ Integration tests completed in ${INTEGRATION_TIME}s"
echo ""

# Test 4: Verify integration tests are significantly slower
echo "5. Verifying performance difference..."
if [ "$INTEGRATION_TIME" -le "$UNIT_TIME" ]; then
    echo "   ⚠ WARNING: Integration tests (${INTEGRATION_TIME}s) should be slower than unit tests (${UNIT_TIME}s)"
else
    RATIO=$((INTEGRATION_TIME / UNIT_TIME))
    echo "   ✓ Integration tests are ${RATIO}x slower than unit tests"
fi
echo ""

# Test 5: Verify Makefile targets work correctly
echo "6. Verifying Makefile targets..."
if ! make test > /dev/null 2>&1; then
    echo "   ✗ ERROR: 'make test' failed"
    exit 1
fi
echo "   ✓ 'make test' works (unit tests only)"

if ! make test-integration > /dev/null 2>&1; then
    echo "   ✗ ERROR: 'make test-integration' failed"
    exit 1
fi
echo "   ✓ 'make test-integration' works"

if ! make test-all > /dev/null 2>&1; then
    echo "   ✗ ERROR: 'make test-all' failed"
    exit 1
fi
echo "   ✓ 'make test-all' works"
echo ""

# Summary
echo "========================================"
echo "✅ Test Separation Verification PASSED"
echo "========================================"
echo ""
echo "Summary:"
echo "  - Unit tests: ${UNIT_TIME}s"
echo "  - Integration tests: ${INTEGRATION_TIME}s"
echo "  - Unit tests remain separate from integration tests ✓"
echo "  - Developers can run fast unit tests during development ✓"
echo ""
