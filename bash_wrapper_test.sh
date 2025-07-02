#!/bin/bash

# Test script for the bash wrapper function

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Test counter
TESTS_RUN=0
TESTS_PASSED=0

# Mock ccsplit command for testing
mock_ccsplit() {
    case "$1" in
        "list")
            echo "Listing sessions..."
            ;;
        "cleanup")
            echo "Cleaning up session: $2"
            ;;
        *)
            # Mock session creation output
            echo "🚀 What are you working on? ✓ Created session: feature/test-feature"
            echo "  Branch: feature/test-feature"
            echo "  Path: /path/to/test-feature"
            echo ""
            echo "# Run this to enter the session:"
            echo "cd ../test-feature"
            ;;
    esac
}

# Source the wrapper function (but use mock instead of real command)
source bash.txt
# Override command lookup to use our mock
command() {
    if [ "$1" = "ccsplit" ]; then
        shift
        mock_ccsplit "$@"
    else
        /usr/bin/command "$@"
    fi
}

# Test function
run_test() {
    local test_name="$1"
    local expected="$2"
    local actual="$3"
    
    TESTS_RUN=$((TESTS_RUN + 1))
    
    if [ "$expected" = "$actual" ]; then
        echo -e "${GREEN}✓${NC} $test_name"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗${NC} $test_name"
        echo "  Expected: $expected"
        echo "  Actual: $actual"
    fi
}

# Test 1: List command should pass through directly
echo "Running bash wrapper tests..."
echo

output=$(ccsplit list 2>&1)
run_test "List command passthrough" "Listing sessions..." "$output"

# Test 2: Cleanup command should pass through directly
output=$(ccsplit cleanup test-session 2>&1)
run_test "Cleanup command passthrough" "Cleaning up session: test-session" "$output"

# Test 3: Session creation should capture and execute cd command
# This is harder to test directly since we can't actually change directories in a subshell
# We'll test that the output contains the expected text
output=$(ccsplit 2>&1)
if echo "$output" | grep -q "Created session: feature/test-feature" && \
   echo "$output" | grep -q "cd ../test-feature"; then
    run_test "Session creation output" "success" "success"
else
    run_test "Session creation output" "success" "failure"
fi

# Test 4: Empty/no arguments should work
output=$(ccsplit 2>&1)
if [ -n "$output" ]; then
    run_test "No arguments handling" "success" "success"
else
    run_test "No arguments handling" "success" "failure"
fi

# Summary
echo
echo "Tests run: $TESTS_RUN"
echo "Tests passed: $TESTS_PASSED"

if [ "$TESTS_RUN" -eq "$TESTS_PASSED" ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
fi