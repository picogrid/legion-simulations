# Check to see if we can use ash, in Alpine images, or default to BASH.
SHELL_PATH = /bin/ash
SHELL = $(if $(wildcard $(SHELL_PATH)),/bin/ash,/bin/bash)


# Define dependencies
GOLANG          := golang:1.24


# ==============================================================================
dev-env:
	@echo "Setting up development environment..."
	$(MAKE) dev-brew
	$(MAKE) dev-precommit
	$(MAKE) dev-gotooling
	@echo "Development environment setup complete."


dev-precommit:
	@echo "Setting up pre-commit hooks..."
	pre-commit install --hook-type pre-commit --hook-type pre-push

# Install dependencies
dev-gotooling:
	go install github.com/divan/expvarmon@latest
	go install github.com/rakyll/hey@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/go-delve/delve/cmd/dlv@latest

dev-brew:
	brew update
	brew list golangci-lint || brew install golangci-lint
	brew list pre-commit || brew install pre-commit


lint:
	golangci-lint run --timeout 5m --config=.golangci.yaml --issues-exit-code=0

lint-fix:
	golangci-lint run --fix --timeout 5m --config=.golangci.yaml --issues-exit-code=0

# Run all unit tests, including combine-specs tests
unit-test: # unit-test-only
#	@echo "\nRunning combine-specs tests..."
#	@cd cmd/tools/combine-specs && go test -v -tags skipdynamotests
	@echo "Running unit tests and outputting to unit-results.json..."
	CGO_ENABLED=0 go test -tags skipdynamotests -count=1 -json ./... > unit-results.json; test_exit_code=$$?; \
	if [ $$test_exit_code -ne 0 ]; then \
		echo "Unit tests failed with exit code $$test_exit_code"; \
		exit $$test_exit_code; \
	fi
	@echo "Unit tests completed successfully."

# New target for local debugging of unit tests
unit-test-debug:
	@echo "Running unit tests VERBOSELY..."
	CGO_ENABLED=0 go test -tags skipdynamotests -count=1 -v ./...; test_exit_code=$$?; \
	if [ $$test_exit_code -ne 0 ]; then \
		echo "Unit tests (debug) failed with exit code $$test_exit_code"; \
		exit $$test_exit_code; \
	fi
	@echo "Unit tests (debug) completed successfully."

# ==============================================================================
# Build targets

.PHONY: build
build:
	@echo "Building legion-sim CLI..."
	@go build -o bin/legion-sim ./cmd/cli
	@echo "Build complete: bin/legion-sim"

.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@go clean -cache
	@echo "Clean complete"

.PHONY: rebuild
rebuild: clean build

# ==============================================================================
# Testing shortcuts

.PHONY: test
test: unit-test

.PHONY: test-verbose
test-verbose: unit-test-debug

# ==============================================================================
# Development helpers

.PHONY: run
run: build
	@./bin/legion-sim run

.PHONY: list
list: build
	@./bin/legion-sim list

.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies updated"

.PHONY: check
check: lint test
	@echo "All checks passed!"

# ==============================================================================
# Help

.PHONY: help
help:
	@echo "Legion Simulations Makefile"
	@echo ""
	@echo "Setup targets:"
	@echo "  make dev-env        - Complete development environment setup"
	@echo "  make dev-brew       - Install Homebrew dependencies"
	@echo "  make dev-precommit  - Set up pre-commit hooks"
	@echo "  make dev-gotooling  - Install Go development tools"
	@echo ""
	@echo "Build targets:"
	@echo "  make build          - Build the legion-sim CLI"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make rebuild        - Clean and rebuild"
	@echo ""
	@echo "Test targets:"
	@echo "  make test           - Run unit tests"
	@echo "  make test-verbose   - Run tests with verbose output"
	@echo ""
	@echo "Code quality:"
	@echo "  make lint           - Run linter"
	@echo "  make lint-fix       - Run linter with auto-fix"
	@echo "  make check          - Run linter and tests"
	@echo ""
	@echo "Development:"
	@echo "  make run            - Build and run interactive CLI"
	@echo "  make list           - Build and list simulations"
	@echo "  make deps           - Update dependencies"