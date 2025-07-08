.PHONY: all build run install test test-unit test-integration test-docker clean help

# Variables
BINARY_NAME=ccswitch
GO_FILES=$(shell find . -name '*.go' -not -path "./vendor/*")
COVERAGE_FILE=coverage.out

# Default target
all: build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME) .

# Run the application (interactive)
run: build
	@./$(BINARY_NAME)

# Run with list command
run-list: build
	@./$(BINARY_NAME) list

# Run with cleanup command
run-cleanup: build
	@./$(BINARY_NAME) cleanup

# Run with switch command
run-switch: build
	@./$(BINARY_NAME) switch

# Install the binary to GOPATH/bin
install: build
	@echo "Installing $(BINARY_NAME)..."
	@go install
	@FIRST_GOPATH=$$(go env GOPATH | cut -d':' -f1) && \
		echo "✓ Installed binary to $$FIRST_GOPATH/bin/$(BINARY_NAME)"
	@echo ""
	@# Install shell integration
	@echo "Setting up shell integration..."
	@SHELL_CONFIG=""; \
	SHELL_NAME=""; \
	if [ -n "$$ZSH_VERSION" ] || [ "$$SHELL" = "/bin/zsh" ] || [ "$$SHELL" = "/usr/bin/zsh" ]; then \
		SHELL_CONFIG="$$HOME/.zshrc"; \
		SHELL_NAME="zsh"; \
	elif [ -n "$$BASH_VERSION" ] || [ "$$SHELL" = "/bin/bash" ] || [ "$$SHELL" = "/usr/bin/bash" ]; then \
		SHELL_CONFIG="$$HOME/.bashrc"; \
		SHELL_NAME="bash"; \
	fi; \
	if [ -n "$$SHELL_CONFIG" ]; then \
		if ! grep -q "eval \"\$$(ccswitch shell-init)\"" "$$SHELL_CONFIG" 2>/dev/null && \
		   ! grep -q "source.*ccswitch/bash.txt" "$$SHELL_CONFIG" 2>/dev/null; then \
			echo "" >> "$$SHELL_CONFIG"; \
			echo "# ccswitch shell integration" >> "$$SHELL_CONFIG"; \
			echo 'eval "$$(ccswitch shell-init)"' >> "$$SHELL_CONFIG"; \
			echo "✓ Added shell integration to $$SHELL_CONFIG"; \
			echo ""; \
			echo "To activate now, run:"; \
			echo "  source $$SHELL_CONFIG"; \
		else \
			echo "✓ Shell integration already installed in $$SHELL_CONFIG"; \
		fi; \
	else \
		echo ""; \
		echo "⚠️  Could not detect shell type. To enable shell integration, add this to your shell config:"; \
		echo ""; \
		echo "  eval \"\$$(ccswitch shell-init)\""; \
		echo ""; \
		echo "For example:"; \
		echo "  echo 'eval \"\$$(ccswitch shell-init)\"' >> ~/.bashrc"; \
		echo "  source ~/.bashrc"; \
	fi

# Uninstall ccswitch
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@FIRST_GOPATH=$$(go env GOPATH | cut -d':' -f1) && \
		rm -f "$$FIRST_GOPATH/bin/$(BINARY_NAME)" && \
		echo "✓ Removed binary from $$FIRST_GOPATH/bin/$(BINARY_NAME)"
	@echo ""
	@echo "Removing shell integration..."
	@SHELL_CONFIG=""; \
	if [ -f "$$HOME/.zshrc" ]; then \
		SHELL_CONFIG="$$HOME/.zshrc"; \
	elif [ -f "$$HOME/.bashrc" ]; then \
		SHELL_CONFIG="$$HOME/.bashrc"; \
	fi; \
	if [ -n "$$SHELL_CONFIG" ]; then \
		if grep -q "eval \"\$$(ccswitch shell-init)\"" "$$SHELL_CONFIG" 2>/dev/null || \
		   grep -q "source.*ccswitch/bash.txt" "$$SHELL_CONFIG" 2>/dev/null; then \
			cp "$$SHELL_CONFIG" "$$SHELL_CONFIG.bak"; \
			grep -v "eval \"\$$(ccswitch shell-init)\"" "$$SHELL_CONFIG.bak" | \
			grep -v "source.*ccswitch/bash.txt" | \
			grep -v "# ccswitch shell integration" > "$$SHELL_CONFIG"; \
			rm "$$SHELL_CONFIG.bak"; \
			echo "✓ Removed shell integration from $$SHELL_CONFIG"; \
		else \
			echo "✓ No shell integration found to remove"; \
		fi; \
	fi
	@echo ""
	@echo "Uninstall complete!"

# Run all tests
test: test-unit test-integration

# Run unit tests only (no git required)
test-unit:
	@echo "Running unit tests..."
	@go test -v -run "^Test(Slugify|ParseWorktrees|SessionItem|GetCurrentDir|RunCmd|WorktreeType)" ./...

# Run integration tests (requires git)
test-integration:
	@echo "Running integration tests..."
	@go test -v -run "^Test(CreateSession|ListSessions|CleanupSession|GetActiveSessions|Integration)" ./... || true

# Run tests in Docker container (for clean git environment)
test-docker:
	@echo "Building Docker test environment..."
	@docker build -t ccsplit-test -f Dockerfile.test .
	@echo "Running tests in Docker..."
	@docker run --rm ccsplit-test

# Generate coverage report
coverage:
	@echo "Generating coverage report..."
	@go test -coverprofile=$(COVERAGE_FILE) ./...
	@go tool cover -html=$(COVERAGE_FILE) -o coverage.html
	@echo "Coverage report generated at coverage.html"

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./...

# Test the bash wrapper
test-bash:
	@echo "Testing bash wrapper..."
	@bash bash_wrapper_test.sh

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -f $(COVERAGE_FILE)
	@rm -f coverage.html
	@go clean

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run || echo "Install golangci-lint: https://golangci-lint.run/usage/install/"

# Show help
help:
	@echo "Available targets:"
	@echo "  make build          - Build the binary"
	@echo "  make run            - Build and run (create new session)"
	@echo "  make run-list       - Build and run list command"
	@echo "  make run-switch     - Build and run switch command"
	@echo "  make run-cleanup    - Build and run cleanup command"
	@echo "  make install        - Install binary and shell integration"
	@echo "  make uninstall      - Remove binary and shell integration"
	@echo "  make test           - Run all tests"
	@echo "  make test-unit      - Run unit tests only"
	@echo "  make test-integration - Run integration tests"
	@echo "  make test-docker    - Run tests in Docker container"
	@echo "  make test-bash      - Test the bash wrapper"
	@echo "  make coverage       - Generate coverage report"
	@echo "  make bench          - Run benchmarks"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make fmt            - Format code"
	@echo "  make lint           - Run linter"
	@echo "  make help           - Show this help message"
