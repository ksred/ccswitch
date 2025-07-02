# Contributing to ccswitch

Thank you for your interest in contributing to ccswitch! This document provides guidelines and instructions for contributing.

## Getting Started

### Prerequisites
- Go 1.21 or higher
- Git 2.20 or higher
- Make (optional but recommended)

### Setting Up Development Environment

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/ksred/ccswitch.git
   cd ccswitch
   ```

3. Install dependencies:
   ```bash
   go mod download
   ```

4. Build the project:
   ```bash
   make build
   # or
   go build -o ccswitch main.go
   ```

## Development Workflow

### Running Tests

We have comprehensive test coverage. Please ensure all tests pass before submitting a PR:

```bash
# Run all tests
make test

# Run unit tests only (no git required)
make test-unit

# Run integration tests (requires git)
make test-integration

# Run tests in Docker (clean environment)
make test-docker

# Generate coverage report
make coverage
```

### Code Style

- Follow standard Go conventions and idioms
- Run `go fmt` before committing:
  ```bash
  make fmt
  ```
- Ensure code passes linting:
  ```bash
  make lint
  ```

### Making Changes

1. Create a new branch for your feature/fix:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes and ensure:
   - All tests pass
   - New features have appropriate tests
   - Code is properly formatted
   - Commit messages are clear and descriptive

3. Test your changes manually:
   ```bash
   make install
   ccswitch "test your changes"
   ```

## Submitting Changes

### Pull Request Process

1. Update documentation if needed
2. Add tests for new functionality
3. Ensure all tests pass
4. Update the README.md if adding new features
5. Submit a pull request with:
   - Clear title and description
   - Reference to any related issues
   - Description of testing performed

### Pull Request Guidelines

- PRs should focus on a single feature or fix
- Keep changes small and focused
- Include tests for new functionality
- Update documentation as needed
- Ensure backwards compatibility

## Reporting Issues

### Bug Reports

When reporting bugs, please include:
- ccswitch version (`ccswitch --version`)
- Go version (`go version`)
- Git version (`git --version`)
- Operating system
- Steps to reproduce
- Expected vs actual behavior
- Any error messages

### Feature Requests

Feature requests are welcome! Please:
- Check existing issues first
- Clearly describe the feature
- Explain use cases and benefits
- Consider submitting a PR!

## Code Organization

Currently, all code is in `main.go`. We're planning to modularize (see IMPROVEMENTS.md), but for now:
- Keep functions focused and well-named
- Add comments for complex logic
- Maintain existing patterns and conventions

## Testing Guidelines

### Unit Tests
- Test pure functions without external dependencies
- Use table-driven tests where appropriate
- Mock git operations when needed

### Integration Tests
- Test actual git operations
- Use temporary repositories
- Clean up after tests

### Shell Wrapper Tests
- Test the bash wrapper functionality
- Ensure shell integration works correctly

## Community

- Be respectful and constructive
- Help others when you can
- Follow our Code of Conduct

## Questions?

If you have questions about contributing:
1. Check existing issues and PRs
2. Read the documentation
3. Open a discussion issue

Thank you for contributing to ccswitch!