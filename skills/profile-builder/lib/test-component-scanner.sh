#!/bin/bash

# Test suite for component-scanner.sh
# Tests Story-001 and Story-002 acceptance criteria

set -e  # Exit on error

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SCANNER="$SCRIPT_DIR/component-scanner.sh"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Helper function to run tests
run_test() {
    local test_name="$1"
    local test_func="$2"
    
    TESTS_RUN=$((TESTS_RUN + 1))
    echo -e "${BLUE}Running:${NC} $test_name"
    
    if $test_func; then
        TESTS_PASSED=$((TESTS_PASSED + 1))
        echo -e "${GREEN}✓ PASSED${NC}"
    else
        TESTS_FAILED=$((TESTS_FAILED + 1))
        echo -e "${RED}✗ FAILED${NC}"
    fi
    echo ""
}

# Story-001 Tests

test_scan_skills_profiles() {
    # Test that scan_skills scans profiles directory
    local TEST_DIR="/tmp/agent-smith-test-$$"
    mkdir -p "$TEST_DIR/.agent-smith/profiles/profile1/skills/skill-a"
    mkdir -p "$TEST_DIR/.agent-smith/profiles/profile2/skills/skill-b"
    
    local result=$(HOME="$TEST_DIR" $SCANNER scan-skills | sort)
    local expected=$'skill-a\nskill-b'
    
    rm -rf "$TEST_DIR"
    
    if [[ "$result" == "$expected" ]]; then
        return 0
    else
        echo "Expected: $expected"
        echo "Got: $result"
        return 1
    fi
}

test_scan_agents_profiles() {
    # Test that scan_agents scans profiles directory
    local TEST_DIR="/tmp/agent-smith-test-$$"
    mkdir -p "$TEST_DIR/.agent-smith/profiles/profile1/agents/category"
    touch "$TEST_DIR/.agent-smith/profiles/profile1/agents/category/agent-a.md"
    mkdir -p "$TEST_DIR/.agent-smith/profiles/profile2/agents/other"
    touch "$TEST_DIR/.agent-smith/profiles/profile2/agents/other/agent-b.md"
    
    local result=$(HOME="$TEST_DIR" $SCANNER scan-agents | sort)
    local expected=$'category/agent-a\nother/agent-b'
    
    rm -rf "$TEST_DIR"
    
    if [[ "$result" == "$expected" ]]; then
        return 0
    else
        echo "Expected: $expected"
        echo "Got: $result"
        return 1
    fi
}

test_scan_commands_profiles() {
    # Test that scan_commands scans profiles directory
    local TEST_DIR="/tmp/agent-smith-test-$$"
    mkdir -p "$TEST_DIR/.agent-smith/profiles/profile1/commands/cmd-a"
    mkdir -p "$TEST_DIR/.agent-smith/profiles/profile2/commands/cmd-b"
    
    local result=$(HOME="$TEST_DIR" $SCANNER scan-commands | sort)
    local expected=$'cmd-a\ncmd-b'
    
    rm -rf "$TEST_DIR"
    
    if [[ "$result" == "$expected" ]]; then
        return 0
    else
        echo "Expected: $expected"
        echo "Got: $result"
        return 1
    fi
}

test_deduplication() {
    # Test that scanner deduplicates components across profiles
    local TEST_DIR="/tmp/agent-smith-test-$$"
    mkdir -p "$TEST_DIR/.agent-smith/profiles/profile1/skills/duplicate-skill"
    mkdir -p "$TEST_DIR/.agent-smith/profiles/profile2/skills/duplicate-skill"
    mkdir -p "$TEST_DIR/.agent-smith/profiles/profile3/skills/duplicate-skill"
    
    local result=$(HOME="$TEST_DIR" $SCANNER scan-skills)
    local count=$(echo "$result" | wc -l | tr -d ' ')
    
    rm -rf "$TEST_DIR"
    
    if [[ "$count" == "1" ]] && [[ "$result" == "duplicate-skill" ]]; then
        return 0
    else
        echo "Expected 1 line with 'duplicate-skill', got $count lines: $result"
        return 1
    fi
}

test_empty_profiles() {
    # Test that scanner handles empty profiles gracefully
    local TEST_DIR="/tmp/agent-smith-test-$$"
    mkdir -p "$TEST_DIR/.agent-smith/profiles/empty-profile/skills"
    mkdir -p "$TEST_DIR/.agent-smith/profiles/empty-profile/agents"
    mkdir -p "$TEST_DIR/.agent-smith/profiles/empty-profile/commands"
    
    local result=$(HOME="$TEST_DIR" $SCANNER scan-skills 2>&1)
    local exit_code=$?
    
    rm -rf "$TEST_DIR"
    
    if [[ $exit_code -eq 0 ]] && [[ -z "$result" ]]; then
        return 0
    else
        echo "Expected empty output with exit code 0, got exit code $exit_code with output: $result"
        return 1
    fi
}

test_nonexistent_directory() {
    # Test that scanner handles non-existent profiles directory gracefully
    local TEST_DIR="/tmp/agent-smith-test-nonexistent-$$"
    
    local result=$(HOME="$TEST_DIR" $SCANNER scan-skills 2>&1)
    local exit_code=$?
    
    if [[ $exit_code -eq 0 ]] && [[ -z "$result" ]]; then
        return 0
    else
        echo "Expected empty output with exit code 0, got exit code $exit_code with output: $result"
        return 1
    fi
}

# Story-002 Tests

test_find_profiles_with_skill() {
    # Test finding profiles containing a skill
    local TEST_DIR="/tmp/agent-smith-test-$$"
    mkdir -p "$TEST_DIR/.agent-smith/profiles/profile-alpha/skills/shared-skill"
    mkdir -p "$TEST_DIR/.agent-smith/profiles/profile-beta/skills/shared-skill"
    mkdir -p "$TEST_DIR/.agent-smith/profiles/profile-gamma/skills/unique-skill"
    
    local result=$(HOME="$TEST_DIR" $SCANNER find-profiles-with-skill "shared-skill" | sort)
    local expected=$'profile-alpha\nprofile-beta'
    
    rm -rf "$TEST_DIR"
    
    if [[ "$result" == "$expected" ]]; then
        return 0
    else
        echo "Expected: $expected"
        echo "Got: $result"
        return 1
    fi
}

test_find_profiles_with_agent() {
    # Test finding profiles containing an agent
    local TEST_DIR="/tmp/agent-smith-test-$$"
    mkdir -p "$TEST_DIR/.agent-smith/profiles/profile1/agents/category"
    touch "$TEST_DIR/.agent-smith/profiles/profile1/agents/category/agent-x.md"
    mkdir -p "$TEST_DIR/.agent-smith/profiles/profile2/agents/category"
    touch "$TEST_DIR/.agent-smith/profiles/profile2/agents/category/agent-x.md"
    
    local result=$(HOME="$TEST_DIR" $SCANNER find-profiles-with-agent "category/agent-x" | sort)
    local expected=$'profile1\nprofile2'
    
    rm -rf "$TEST_DIR"
    
    if [[ "$result" == "$expected" ]]; then
        return 0
    else
        echo "Expected: $expected"
        echo "Got: $result"
        return 1
    fi
}

test_find_profiles_with_command() {
    # Test finding profiles containing a command
    local TEST_DIR="/tmp/agent-smith-test-$$"
    mkdir -p "$TEST_DIR/.agent-smith/profiles/profile-x/commands/test-cmd"
    mkdir -p "$TEST_DIR/.agent-smith/profiles/profile-y/commands/test-cmd"
    mkdir -p "$TEST_DIR/.agent-smith/profiles/profile-z/commands/other-cmd"
    
    local result=$(HOME="$TEST_DIR" $SCANNER find-profiles-with-command "test-cmd" | sort)
    local expected=$'profile-x\nprofile-y'
    
    rm -rf "$TEST_DIR"
    
    if [[ "$result" == "$expected" ]]; then
        return 0
    else
        echo "Expected: $expected"
        echo "Got: $result"
        return 1
    fi
}

test_list_profiles() {
    # Test listing all profiles
    local TEST_DIR="/tmp/agent-smith-test-$$"
    mkdir -p "$TEST_DIR/.agent-smith/profiles/profile-alpha"
    mkdir -p "$TEST_DIR/.agent-smith/profiles/profile-beta"
    mkdir -p "$TEST_DIR/.agent-smith/profiles/profile-gamma"
    
    local result=$(HOME="$TEST_DIR" $SCANNER list-profiles)
    local expected=$'profile-alpha\nprofile-beta\nprofile-gamma'
    
    rm -rf "$TEST_DIR"
    
    if [[ "$result" == "$expected" ]]; then
        return 0
    else
        echo "Expected: $expected"
        echo "Got: $result"
        return 1
    fi
}

test_find_nonexistent_component() {
    # Test finding profiles for non-existent component
    local TEST_DIR="/tmp/agent-smith-test-$$"
    mkdir -p "$TEST_DIR/.agent-smith/profiles/profile1/skills/skill-a"
    
    local result=$(HOME="$TEST_DIR" $SCANNER find-profiles-with-skill "nonexistent-skill")
    
    rm -rf "$TEST_DIR"
    
    if [[ -z "$result" ]]; then
        return 0
    else
        echo "Expected empty output, got: $result"
        return 1
    fi
}

# Integration Tests

test_real_profiles() {
    # Test against actual ~/.agent-smith/profiles structure
    if [[ ! -d "$HOME/.agent-smith/profiles" ]]; then
        echo "Skipping: No real profiles directory found"
        return 0
    fi
    
    # Just verify the commands don't error
    $SCANNER scan-skills > /dev/null || return 1
    $SCANNER scan-agents > /dev/null || return 1
    $SCANNER scan-commands > /dev/null || return 1
    $SCANNER list-profiles > /dev/null || return 1
    
    return 0
}

# Main test execution

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}Component Scanner Test Suite${NC}"
echo -e "${YELLOW}========================================${NC}"
echo ""

echo -e "${BLUE}Testing Story-001: Component Discovery${NC}"
echo ""

run_test "scan_skills scans profiles directory" test_scan_skills_profiles
run_test "scan_agents scans profiles directory" test_scan_agents_profiles
run_test "scan_commands scans profiles directory" test_scan_commands_profiles
run_test "Scanner deduplicates components" test_deduplication
run_test "Scanner handles empty profiles" test_empty_profiles
run_test "Scanner handles non-existent directory" test_nonexistent_directory

echo -e "${BLUE}Testing Story-002: Profile Tracking${NC}"
echo ""

run_test "find_profiles_with_skill returns correct profiles" test_find_profiles_with_skill
run_test "find_profiles_with_agent returns correct profiles" test_find_profiles_with_agent
run_test "find_profiles_with_command returns correct profiles" test_find_profiles_with_command
run_test "list_profiles lists all profiles" test_list_profiles
run_test "Finding non-existent component returns empty" test_find_nonexistent_component

echo -e "${BLUE}Integration Tests${NC}"
echo ""

run_test "Scanner works with real profiles" test_real_profiles

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}Test Summary${NC}"
echo -e "${YELLOW}========================================${NC}"
echo ""
echo -e "Tests Run:    ${BLUE}$TESTS_RUN${NC}"
echo -e "Tests Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests Failed: ${RED}$TESTS_FAILED${NC}"
echo ""

if [[ $TESTS_FAILED -eq 0 ]]; then
    echo -e "${GREEN}✓ All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}✗ Some tests failed${NC}"
    exit 1
fi
