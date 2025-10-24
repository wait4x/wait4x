# Copyright 2019-2025 The Wait4X Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# =============================================================================
# Configuration
# =============================================================================

# Go configuration
GO_BINARY ?= $(shell which go)
GO_ENVIRONMENTS ?=
GO_VERSION ?= $(shell $(GO_BINARY) version | cut -d' ' -f3 | sed 's/go//')

# Wait4X configuration
WAIT4X_BINARY_NAME ?= wait4x
WAIT4X_MODULE_NAME ?= wait4x.dev/v4
WAIT4X_MAIN_PATH ?= cmd/wait4x/main.go

# Test configuration
WAIT4X_COVERAGE_IGNORE_PACKAGES ?= ${WAIT4X_MODULE_NAME}/examples ${WAIT4X_MODULE_NAME}/cmd

# Build configuration
WAIT4X_BUILD_OUTPUT ?= ${CURDIR}/dist
WAIT4X_BUILD_OS ?= $(shell go env GOOS)
WAIT4X_BUILD_ARCH ?= $(shell go env GOARCH)
WAIT4X_BUILD_CGO ?= 0

# Version information
WAIT4X_COMMIT_REF_SLUG ?= $(shell [ -d ./.git ] && (git symbolic-ref -q --short HEAD || git describe --tags --always 2>/dev/null || echo "unknown"))
WAIT4X_COMMIT_HASH ?= $(shell [ -d ./.git ] && git rev-parse HEAD 2>/dev/null || echo "unknown")
WAIT4X_BUILD_TIME ?= $(shell date -u '+%FT%TZ')

# Change output to .exe for windows
ifeq (windows,$(filter windows,$(WAIT4X_BUILD_OS) $(TARGETOS)))
WAIT4X_BINARY_NAME := ${WAIT4X_BINARY_NAME}.exe
endif

# =============================================================================
# Build flags and LDFLAGS
# =============================================================================

# Base LDFLAGS for reproducible builds and metadata
WAIT4X_BUILD_LDFLAGS ?= -buildid= -w \
	-X $(WAIT4X_MODULE_NAME)/internal/cmd.BuildTime=$(WAIT4X_BUILD_TIME)

# Add version information if available
ifneq ($(WAIT4X_COMMIT_REF_SLUG),)
WAIT4X_BUILD_LDFLAGS += -X $(WAIT4X_MODULE_NAME)/internal/cmd.AppVersion=$(WAIT4X_COMMIT_REF_SLUG)
endif

ifneq ($(WAIT4X_COMMIT_HASH),)
WAIT4X_BUILD_LDFLAGS += -X $(WAIT4X_MODULE_NAME)/internal/cmd.GitCommit=$(WAIT4X_COMMIT_HASH)
endif

# Build flags for reproducible builds
WAIT4X_BUILD_FLAGS ?= -trimpath -ldflags="$(WAIT4X_BUILD_LDFLAGS)"
WAIT4X_RUN_FLAGS ?= -ldflags="$(WAIT4X_BUILD_LDFLAGS)"

# Runtime flags
WAIT4X_FLAGS ?= 

# =============================================================================
# Utility functions
# =============================================================================

# Check if a command exists
check_cmd = $(shell command -v $(1) 2> /dev/null)

# Filter coverage output to exclude ignored packages
filter_coverage = @grep $(foreach pkg,$(WAIT4X_COVERAGE_IGNORE_PACKAGES),-v -e "$(pkg)") coverage.out.tmp > coverage.out

# Common build function
define build_binary
	@mkdir -p $(WAIT4X_BUILD_OUTPUT)
	CGO_ENABLED=$(WAIT4X_BUILD_CGO) GOOS=$(WAIT4X_BUILD_OS) GOARCH=$(WAIT4X_BUILD_ARCH) \
	$(GO_ENVIRONMENTS) $(GO_BINARY) build -v $(1) \
		-o $(WAIT4X_BUILD_OUTPUT)/$(2) $(WAIT4X_MAIN_PATH)
endef

# =============================================================================
# Targets
# =============================================================================

.PHONY: help
help: ## Show this help message
	@echo " __      __        .__  __     _________  ___"
	@echo "/  \    /  \_____  |__|/  |_  /  |  \   \/  /"
	@echo "\   \/\/   /\__  \ |  \   __\/   |  |\     / "
	@echo " \        /  / __ \|  ||  | /    ^   /     \ "
	@echo "  \__/\  /  (____  /__||__| \____   /___/\  \\"
	@echo "       \/        \/              |__|     \_/"
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "Environment variables:"
	@echo "  GO_BINARY                       Go binary to use (default: $(GO_BINARY))"
	@echo "  GO_ENVIRONMENTS                 Additional environment variables for Go (default: $(if $(GO_ENVIRONMENTS),$(GO_ENVIRONMENTS),none))"
	@echo "  WAIT4X_BUILD_OUTPUT             Build output directory (default: ${WAIT4X_BUILD_OUTPUT})"
	@echo "  WAIT4X_BUILD_OS                 Target OS for cross-compilation (default: $(WAIT4X_BUILD_OS))"
	@echo "  WAIT4X_BUILD_ARCH               Target architecture for cross-compilation (default: $(WAIT4X_BUILD_ARCH))"
	@echo "  WAIT4X_FLAGS                    Additional flags for wait4x binary (default: $(if $(WAIT4X_FLAGS),$(WAIT4X_FLAGS),none))"
	@echo "  WAIT4X_COVERAGE_IGNORE_PACKAGES Space-separated list of packages to exclude from coverage (default: ${WAIT4X_COVERAGE_IGNORE_PACKAGES})"

.PHONY: version
version: ## Show version information
	@echo "Go version: $(GO_VERSION)"
	@echo "Build OS: $(WAIT4X_BUILD_OS)"
	@echo "Build ARCH: $(WAIT4X_BUILD_ARCH)"
	@echo "Commit ref: $(WAIT4X_COMMIT_REF_SLUG)"
	@echo "Commit hash: $(WAIT4X_COMMIT_HASH)"
	@echo "Build time: $(WAIT4X_BUILD_TIME)"
	@echo "Coverage ignore packages: $(WAIT4X_COVERAGE_IGNORE_PACKAGES)"

.PHONY: show-coverage-config
show-coverage-config: ## Show current coverage configuration
	@echo "Coverage ignore packages: $(WAIT4X_COVERAGE_IGNORE_PACKAGES)"
	@echo "To modify, set WAIT4X_COVERAGE_IGNORE_PACKAGES environment variable"
	@echo "Example: WAIT4X_COVERAGE_IGNORE_PACKAGES='pkg1 pkg2' make test"

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(WAIT4X_BUILD_OUTPUT)
	@go clean -cache -testcache -modcache
	@echo "Clean complete!"

.PHONY: deps
deps: ## Download and tidy dependencies
	@echo "Downloading dependencies..."
	$(GO_ENVIRONMENTS) $(GO_BINARY) mod download
	$(GO_ENVIRONMENTS) $(GO_BINARY) mod tidy
	@echo "Dependencies ready!"

.PHONY: check-deps
check-deps: ## Check for outdated dependencies
	@echo "Checking for outdated dependencies..."
	@$(GO_ENVIRONMENTS) $(GO_BINARY) list -u -m all | grep -v "go:" || echo "All dependencies are up to date"

.PHONY: test
test: ## Run unit tests with coverage (excludes integration tests)
	@echo "Running unit tests..."
	$(GO_ENVIRONMENTS) $(GO_BINARY) test -v -race -covermode=atomic -coverprofile=coverage.out.tmp ./...
	@$(filter_coverage) > coverage.out
	@rm coverage.out.tmp
	@echo "Test coverage report generated: coverage.out"

.PHONY: test-short
test-short: ## Run tests without race detection and integration tests
	@echo "Running tests (short mode)..."
	$(GO_ENVIRONMENTS) $(GO_BINARY) test -v -short -covermode=atomic -coverprofile=coverage.out.tmp ./...
	@$(filter_coverage) > coverage.out
	@rm coverage.out.tmp

.PHONY: test-integration
test-integration: ## Run integration tests (requires Docker)
	@echo "Running integration tests..."
	@echo "⚠️  Note: Integration tests require Docker to be running"
	$(GO_ENVIRONMENTS) $(GO_BINARY) test -v -race -tags=integration -covermode=atomic -coverprofile=coverage-integration.out.tmp ./...
	@$(filter_coverage) < coverage-integration.out.tmp > coverage-integration.out
	@rm coverage-integration.out.tmp
	@echo "Integration test coverage report generated: coverage-integration.out"

.PHONY: test-all
test-all: ## Run all tests (unit + integration)
	@echo "Running all tests (unit + integration)..."
	@echo "⚠️  Note: Integration tests require Docker to be running"
	$(GO_ENVIRONMENTS) $(GO_BINARY) test -v -race -tags=integration -covermode=atomic -coverprofile=coverage-all.out.tmp ./...
	@$(filter_coverage) < coverage-all.out.tmp > coverage-all.out
	@rm coverage-all.out.tmp
	@echo "Complete test coverage report generated: coverage-all.out"

.PHONY: test-coverage
test-coverage: ## Run unit tests and show coverage report
	@echo "Running unit tests with coverage..."
	$(GO_ENVIRONMENTS) $(GO_BINARY) test -v -race -covermode=atomic -coverprofile=coverage.out.tmp ./...
	@$(filter_coverage) > coverage.out
	@rm coverage.out.tmp
	@echo "Coverage report:"
	$(GO_ENVIRONMENTS) $(GO_BINARY) tool cover -func=coverage.out
	@echo ""
	@echo "HTML coverage report:"
	$(GO_ENVIRONMENTS) $(GO_BINARY) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report saved to coverage.html"

.PHONY: test-coverage-all
test-coverage-all: ## Run all tests (unit + integration) and show coverage report
	@echo "Running all tests with coverage..."
	@echo "⚠️  Note: Integration tests require Docker to be running"
	$(GO_ENVIRONMENTS) $(GO_BINARY) test -v -race -tags=integration -covermode=atomic -coverprofile=coverage-all.out.tmp ./...
	@$(filter_coverage) < coverage-all.out.tmp > coverage-all.out
	@rm coverage-all.out.tmp
	@echo "Coverage report:"
	$(GO_ENVIRONMENTS) $(GO_BINARY) tool cover -func=coverage-all.out
	@echo ""
	@echo "HTML coverage report:"
	$(GO_ENVIRONMENTS) $(GO_BINARY) tool cover -html=coverage-all.out -o coverage-all.html
	@echo "Coverage report saved to coverage-all.html"

.PHONY: lint
lint: ## Run all linting checks
	@echo "Running linting checks..."
	@$(MAKE) check-gofmt
	@$(MAKE) check-revive
	@$(MAKE) check-govet
	@echo "All linting checks passed!"

.PHONY: check-gofmt
check-gofmt: ## Check Go code formatting
	@echo "Checking Go code formatting..."
	@if [ -n "$(shell gofmt -s -l .)" ]; then \
		echo "❌ Go code is not formatted. Run 'make fmt' to fix."; \
		exit 1; \
	fi
	@echo "✅ Go code formatting is correct"

.PHONY: fmt
fmt: ## Format Go code
	@echo "Formatting Go code..."
	$(GO_ENVIRONMENTS) $(GO_BINARY) fmt ./...
	gofmt -s -w .
	@echo "✅ Go code formatted"

.PHONY: check-revive
check-revive: ## Run revive linter
	@echo "Running revive linter..."
	@if [ -z "$(call check_cmd,revive)" ]; then \
		echo "❌ revive not found. Install with: go install github.com/mgechev/revive@latest (or run 'nix develop' to install)"; \
		exit 1; \
	fi
	revive -config .revive.toml -formatter friendly ./...

.PHONY: check-govet
check-govet: ## Run go vet
	@echo "Running go vet..."
	$(GO_ENVIRONMENTS) $(GO_BINARY) vet ./...

.PHONY: security
security: ## Run security checks
	@echo "Running security checks..."
	@if [ -n "$(call check_cmd,gosec)" ]; then \
		gosec ./...; \
	else \
		echo "⚠️  gosec not found. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest (or run 'nix develop' to install)"; \
	fi

.PHONY: build
build: ## Build Wait4X binary
	@echo "Building Wait4X..."
	$(call build_binary,$(WAIT4X_BUILD_FLAGS),$(WAIT4X_BINARY_NAME))
	@echo "✅ Binary built: $(WAIT4X_BUILD_OUTPUT)/$(WAIT4X_BINARY_NAME)"

.PHONY: build-debug
build-debug: ## Build Wait4X binary with debug information
	@echo "Building Wait4X (debug mode)..."
	$(call build_binary,-gcflags="all=-N -l",$(WAIT4X_BINARY_NAME)-debug)
	@echo "✅ Debug binary built: $(WAIT4X_BUILD_OUTPUT)/$(WAIT4X_BINARY_NAME)-debug"

.PHONY: build-cross
build-cross: ## Build Wait4X for multiple platforms
	@echo "Building Wait4X for multiple platforms..."
	@mkdir -p $(WAIT4X_BUILD_OUTPUT)
	@for os in linux darwin windows; do \
		for arch in amd64 arm64; do \
			if [ "$$os" = "windows" ]; then \
				binary_name="$(WAIT4X_BINARY_NAME).exe"; \
			else \
				binary_name="$(WAIT4X_BINARY_NAME)"; \
			fi; \
			echo "Building for $$os/$$arch..."; \
			CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch \
			$(GO_ENVIRONMENTS) $(GO_BINARY) build -v $(WAIT4X_BUILD_FLAGS) \
				-o $(WAIT4X_BUILD_OUTPUT)/$(WAIT4X_BINARY_NAME)-$$os-$$arch/$$binary_name $(WAIT4X_MAIN_PATH); \
		done; \
	done
	@echo "✅ Cross-platform builds complete!"

.PHONY: install
install: build ## Install Wait4X to system
	@echo "Installing Wait4X..."
	$(GO_ENVIRONMENTS) $(GO_BINARY) install $(WAIT4X_BUILD_FLAGS) $(WAIT4X_MAIN_PATH)
	@echo "✅ Wait4X installed to $(shell go env GOPATH)/bin/"

.PHONY: run
run: ## Run Wait4X
	@echo "Running Wait4X..."
	$(GO_ENVIRONMENTS) $(GO_BINARY) run $(WAIT4X_RUN_FLAGS) $(WAIT4X_MAIN_PATH) $(WAIT4X_FLAGS)

.PHONY: run-examples
run-examples: ## Run example configurations
	@echo "Running examples..."
	@if [ -d "examples" ]; then \
		for example in examples/*/; do \
			if [ -f "$$example/main.go" ]; then \
				echo "Running example: $$(basename $$example)"; \
				cd $$example && go run main.go; \
				cd - > /dev/null; \
			fi; \
		done; \
	else \
		echo "No examples directory found"; \
	fi

.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t wait4x:latest .
	@echo "✅ Docker image built: wait4x:latest"

.PHONY: docker-run
docker-run: ## Run Wait4X in Docker
	@echo "Running Wait4X in Docker..."
	docker run --rm wait4x:latest $(WAIT4X_FLAGS)

.PHONY: release
release: clean deps lint test build-cross ## Prepare release (clean, lint, test, build)
	@echo "✅ Release preparation complete!"

.PHONY: ci
ci: deps lint test ## Run CI pipeline (unit tests only)
	@echo "✅ CI pipeline completed successfully!"

.PHONY: ci-integration
ci-integration: deps lint test-all ## Run full CI pipeline (unit + integration tests)
	@echo "✅ Full CI pipeline completed successfully!"

# =============================================================================
# Development helpers
# =============================================================================

.PHONY: dev-setup
dev-setup: ## Setup development environment
	@echo "Setting up development environment..."
	@echo "Using Nix for development environment setup..."
	@echo "Run 'nix develop' to enter the development shell with all tools pre-installed"
	@echo "Or run 'nix build' to build the project"
	@echo "✅ Development environment ready! (Managed by Nix)"

# =============================================================================
# Documentation
# =============================================================================

.PHONY: docs
docs: ## Generate documentation
	@echo "Generating documentation..."
	@if [ -z "$(call check_cmd,godoc)" ]; then \
		echo "⚠️  godoc not found. Install with: go install golang.org/x/tools/cmd/godoc@latest (or run 'nix develop' to install)"; \
		exit 1; \
	else \
		echo "Running 'godoc -http=:6060' to start the documentation server"; \
		echo "Open your browser to http://localhost:6060/pkg/$(WAIT4X_MODULE_NAME)/"; \
		echo "Press Ctrl+C to stop the server"; \
		godoc -http=:6060; \
	fi

# =============================================================================
# Default target
# =============================================================================

.DEFAULT_GOAL := help
