# ccsplit Test Suite

This project includes comprehensive tests for the ccsplit CLI tool.

## Test Files

1. **unit_test.go** - Pure unit tests that don't require git operations
   - Tests for `slugify()` function
   - Tests for `parseWorktrees()` function
   - Tests for data structures and interfaces
   - ~40% code coverage

2. **main_test.go** - Tests for TUI components and helpers
   - Tests for list models
   - Tests for session items
   - Tests for delegate rendering

3. **commands_test.go** - Integration tests for commands (requires git)
   - Tests for `createSession` command
   - Tests for `listSessions` command
   - Tests for `cleanupSession` command
   - Currently fails in environments with existing worktrees

4. **bash_wrapper_test.sh** - Tests for the bash wrapper function
   - Tests command passthrough
   - Tests session creation output parsing
   - All tests passing

## Running Tests

### Unit Tests Only (No Git Required)
```bash
go test -v -run "^Test(Slugify|ParseWorktrees|SessionItem|GetCurrentDir|RunCmd|WorktreeType)" ./...
```

### All Tests
```bash
go test -v -cover ./...
```

### Coverage Report
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Bash Wrapper Tests
```bash
./bash_wrapper_test.sh
```

## Test Coverage

Current coverage with passing tests: ~40%

The integration tests that require git operations are currently skipped in CI environments. These tests would provide additional coverage for:
- Session creation workflow
- Worktree management
- Branch operations
- Cleanup operations

## Benchmarks

The test suite includes benchmarks for performance-critical functions:
```bash
go test -bench=. ./...
```