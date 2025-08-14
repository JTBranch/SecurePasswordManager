# Configuration
MIN_COVERAGE ?= 25.0

# Development
dev:
	GO_PASSWORD_MANAGER_ENV=dev go run ./cmd/main.go

# Watch mode for development (requires air to be installed)
dev-watch:
	@if ! command -v air &> /dev/null; then \
		echo "❌ Air not found. Installing air for watch mode..."; \
		go install github.com/air-verse/air@latest; \
	fi
	GO_PASSWORD_MANAGER_ENV=dev air

# Build the application (with auto-formatting)
build: fmt
	go build -o main ./cmd/main.go

# Build without formatting (for CI)
build-only:
	go build -o main ./cmd/main.go

# Run in production mode
run:
	./main

# Testing
test:
	go test ./... -v

unitTest:
	go test ./internal/... ./ui/... -v

e2eTest:
	go test ./tests/e2e -v

# Coverage
coverage:
	mkdir -p tmp/output
	go test ./internal/... ./ui/... -coverprofile=tmp/output/coverage.out -covermode=atomic
	go tool cover -html=tmp/output/coverage.out -o tmp/output/coverage.html
	go tool cover -func=tmp/output/coverage.out

# Check coverage threshold
coverage-check: coverage
	@COVERAGE=$$(go tool cover -func=tmp/output/coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	if [ $$(echo "$$COVERAGE >= $(MIN_COVERAGE)" | bc -l) -eq 1 ]; then \
		echo "✅ Coverage check passed: $$COVERAGE% >= $(MIN_COVERAGE)%"; \
	else \
		echo "❌ Coverage check failed: $$COVERAGE% < $(MIN_COVERAGE)%"; \
		echo "💡 Add more unit tests to increase coverage"; \
		exit 1; \
	fi

# CI Operations
ci-local:
	./scripts/ci-local.sh

ci-test:
	mkdir -p tmp/output
	go test -v -race -coverprofile=tmp/output/coverage.out -covermode=atomic \
		./internal/... ./ui/... \
		-json > tmp/output/unit-test-results.json
	go test -v ./tests/e2e/... > tmp/output/e2e-test-output.txt 2>&1

ci-coverage:
	mkdir -p tmp/output
	go test ./internal/... ./ui/... -coverprofile=tmp/output/coverage.out -covermode=atomic
	go tool cover -html=tmp/output/coverage.out -o tmp/output/coverage.html
	go tool cover -func=tmp/output/coverage.out > tmp/output/coverage-summary.txt

ci-build:
	mkdir -p tmp/output
	go build -ldflags="-s -w" -o tmp/output/password-manager ./cmd/main.go

# Cross-platform build (used by CI)
build-cross:
	@mkdir -p tmp/output
	@if [ -z "$(GOOS)" ] || [ -z "$(GOARCH)" ]; then \
		echo "❌ GOOS and GOARCH must be set"; \
		echo "Example: make build-cross GOOS=linux GOARCH=amd64"; \
		exit 1; \
	fi
	@BINARY_NAME="password-manager"; \
	if [ "$(GOOS)" = "windows" ]; then \
		BINARY_NAME="password-manager.exe"; \
	fi; \
	echo "Building for $(GOOS)/$(GOARCH)..."; \
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="-s -w" -o tmp/output/$${BINARY_NAME}-$(GOOS)-$(GOARCH) ./cmd/main.go; \
	ls -lh tmp/output/$${BINARY_NAME}-$(GOOS)-$(GOARCH)

# Code Quality
# Format code
fmt:
	gofmt -w .
	goimports -w .

# Check if code is formatted
fmt-check:
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "❌ Code is not formatted. Run 'make fmt' to fix:"; \
		gofmt -l .; \
		exit 1; \
	else \
		echo "✅ Code is properly formatted"; \
	fi

# Check imports
imports-check:
	@if ! command -v goimports &> /dev/null; then \
		echo "Installing goimports..."; \
		go install golang.org/x/tools/cmd/goimports@latest; \
	fi
	@if [ -n "$$(goimports -l .)" ]; then \
		echo "❌ Imports are not organized. Run 'make fmt' to fix:"; \
		goimports -l .; \
		exit 1; \
	else \
		echo "✅ Imports are properly organized"; \
	fi

lint:
	go vet ./...
	golint ./... || echo "Install golint: go install golang.org/x/lint/golint@latest"

# Advanced linting with golangci-lint
lint-advanced:
	golangci-lint run

# Auto-fix linting issues where possible
lint-fix:
	golangci-lint run --fix

# Just check for comment issues (golint only)
lint-comments:
	golint ./...

# Clean build artifacts
clean:
	rm -f main tmp/main
	rm -rf tmp/output/*

# Install CI/build dependencies (no air - it's for local dev only)
install-deps:
	go install golang.org/x/lint/golint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install system dependencies for Fyne (Ubuntu/Debian only)
install-system-deps:
	@if command -v apt-get &> /dev/null; then \
		echo "📦 Installing system dependencies for Fyne..."; \
		./scripts/install-system-deps.sh; \
	else \
		echo "ℹ️ System dependency installation only supported on Ubuntu/Debian"; \
		echo "💡 On macOS: System dependencies are typically pre-installed"; \
		echo "💡 On other systems: Install OpenGL and X11 development libraries manually"; \
	fi

# Install local development dependencies (including air for watch mode)
install-dev-deps:
	go install golang.org/x/lint/golint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/air-verse/air@latest

# Install git pre-commit hook for auto-formatting
install-hooks:
	@if [ -d .git ]; then \
		echo "Installing pre-commit hook..."; \
		cp scripts/pre-commit-hook.sh .git/hooks/pre-commit; \
		chmod +x .git/hooks/pre-commit; \
		echo "✅ Pre-commit hook installed successfully!"; \
		echo "💡 Now your code will be auto-formatted on every commit"; \
	else \
		echo "❌ Not a git repository. Pre-commit hook not installed."; \
	fi

# Remove git pre-commit hook
uninstall-hooks:
	@if [ -f .git/hooks/pre-commit ]; then \
		rm .git/hooks/pre-commit; \
		echo "✅ Pre-commit hook removed"; \
	else \
		echo "ℹ️ No pre-commit hook found"; \
	fi

# Help
help:
	@echo "Available targets:"
	@echo "  dev              - Run in development mode"
	@echo "  dev-watch        - Run in watch mode (auto-reloads on changes)"
	@echo "  build            - Build the application (auto-formats code first)"
	@echo "  test             - Run all tests"
	@echo "  fmt              - Format code and organize imports"
	@echo "  fmt-check        - Check if code is properly formatted"
	@echo "  coverage         - Generate coverage report"
	@echo "  coverage-check   - Check coverage meets threshold (MIN_COVERAGE=$(MIN_COVERAGE)%)"
	@echo "  ci-local         - Run full CI pipeline locally"
	@echo "  lint             - Run basic linting (go vet + golint)"
	@echo "  lint-advanced    - Run advanced linting (golangci-lint)"
	@echo "  lint-fix         - Auto-fix linting issues where possible"
	@echo "  lint-comments    - Check only comment-related issues"
	@echo "  lint             - Run code quality checks"
	@echo "  install-deps     - Install CI/build dependencies"
	@echo "  install-system-deps - Install system dependencies for Fyne (Ubuntu/Debian)"
	@echo "  install-dev-deps - Install all development dependencies (including air)"
	@echo "  install-hooks    - Install git pre-commit hook for auto-formatting"
	@echo "  uninstall-hooks  - Remove git pre-commit hook"
	@echo "  clean            - Clean build artifacts"
	@echo ""
	@echo "Configuration:"
	@echo "  MIN_COVERAGE     - Minimum coverage threshold (default: $(MIN_COVERAGE)%)"
	@echo ""
	@echo "Examples:"
	@echo "  make fmt                              # Format all code"
	@echo "  make coverage-check MIN_COVERAGE=30.0 # Check coverage with custom threshold"
	@echo "  make install-hooks                    # Install pre-commit formatting"
	@echo "  make ci-local                         # Run full CI pipeline"

.PHONY: dev dev-watch build build-only build-cross run test unitTest e2eTest coverage coverage-check fmt fmt-check imports-check ci-local ci-test ci-coverage ci-build lint lint-advanced lint-fix lint-comments clean install-deps install-system-deps install-dev-deps install-hooks uninstall-hooks help