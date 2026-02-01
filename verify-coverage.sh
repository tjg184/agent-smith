#!/bin/bash

# Coverage Verification Script
# Compares test coverage before and after cleanup to ensure no coverage is lost

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "==================================="
echo "Test Coverage Verification"
echo "==================================="
echo ""

# Function to extract coverage percentage from a line
extract_coverage() {
    echo "$1" | grep -oE '[0-9]+\.[0-9]+%' | grep -oE '[0-9]+\.[0-9]+' || echo "0.0"
}

# Function to compare coverage
compare_coverage() {
    local package=$1
    local before=$2
    local after=$3
    
    # Convert to numeric for comparison
    local diff=$(echo "$after - $before" | bc -l)
    
    if (( $(echo "$diff > 0.5" | bc -l) )); then
        echo -e "${GREEN}✓ $package: ${before}% → ${after}% (+${diff}%)${NC}"
        return 0
    elif (( $(echo "$diff < -2.0" | bc -l) )); then
        echo -e "${RED}✗ $package: ${before}% → ${after}% (${diff}%)${NC}"
        return 1
    else
        echo -e "${YELLOW}• $package: ${before}% → ${after}% (${diff}%)${NC}"
        return 0
    fi
}

# Check if baseline files exist
if [ ! -f "coverage-before-unit.txt" ]; then
    echo -e "${RED}Error: coverage-before-unit.txt not found${NC}"
    echo "Run: go test -cover ./... | tee coverage-before-unit.txt"
    exit 1
fi

if [ ! -f "coverage-before-integration.txt" ]; then
    echo -e "${RED}Error: coverage-before-integration.txt not found${NC}"
    echo "Run: go test -tags=integration -cover ./... | tee coverage-before-integration.txt"
    exit 1
fi

# Generate current coverage reports
echo "Running current coverage tests..."
echo ""

echo "Unit tests..."
go test -cover ./... 2>&1 | tee coverage-after-unit.txt

echo ""
echo "Integration tests..."
go test -tags=integration -cover ./... 2>&1 | tee coverage-after-integration.txt

echo ""
echo "==================================="
echo "Coverage Comparison - Unit Tests"
echo "==================================="
echo ""

# Parse and compare unit test coverage
failures=0
declare -A packages=(
    ["internal/detector"]="internal/detector"
    ["internal/fileutil"]="internal/fileutil"
    ["internal/git"]="internal/git"
    ["internal/linker"]="internal/linker"
    ["internal/testutil"]="internal/testutil"
    ["internal/updater"]="internal/updater"
    ["pkg/config"]="pkg/config"
    ["pkg/logger"]="pkg/logger"
    ["pkg/paths"]="pkg/paths"
    ["pkg/profiles"]="pkg/profiles"
)

for pkg_name in "${!packages[@]}"; do
    pkg_path="${packages[$pkg_name]}"
    
    # Extract coverage from before file
    before_line=$(grep "$pkg_path" coverage-before-unit.txt | grep "coverage:" || echo "")
    before_cov=$(extract_coverage "$before_line")
    
    # Extract coverage from after file
    after_line=$(grep "$pkg_path" coverage-after-unit.txt | grep "coverage:" || echo "")
    after_cov=$(extract_coverage "$after_line")
    
    if [ "$before_cov" != "0.0" ] || [ "$after_cov" != "0.0" ]; then
        if ! compare_coverage "$pkg_name" "$before_cov" "$after_cov"; then
            failures=$((failures + 1))
        fi
    fi
done

echo ""
echo "==================================="
echo "Coverage Comparison - Integration Tests"
echo "==================================="
echo ""

# Check integration test coverage for downloader (should increase after refactor)
before_line=$(grep "internal/downloader" coverage-before-integration.txt | grep "coverage:" || echo "")
before_cov=$(extract_coverage "$before_line")

after_line=$(grep "internal/downloader" coverage-after-integration.txt | grep "coverage:" || echo "")
after_cov=$(extract_coverage "$after_line")

if [ "$before_cov" != "0.0" ] || [ "$after_cov" != "0.0" ]; then
    if ! compare_coverage "internal/downloader" "$before_cov" "$after_cov"; then
        failures=$((failures + 1))
    fi
fi

echo ""
echo "==================================="
echo "Summary"
echo "==================================="
echo ""

if [ $failures -eq 0 ]; then
    echo -e "${GREEN}✓ All coverage checks passed!${NC}"
    echo "Coverage maintained or improved across all packages."
    exit 0
else
    echo -e "${RED}✗ $failures package(s) showed significant coverage decrease${NC}"
    echo "Coverage decreased by more than 2% in some packages."
    echo "Review the changes and add tests to maintain coverage."
    exit 1
fi
