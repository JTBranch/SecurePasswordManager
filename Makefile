# Go Password Manager - Simple Makefile

# Development
dev:
	go run ./cmd

build:
	go build -o bin/password-manager ./cmd

# Testing
test:
	go test ./...

test-all:
	@echo "🧪 Running comprehensive test suite..."
	@mkdir -p tmp/output
	go test -v -race -coverprofile=tmp/output/coverage.out -covermode=atomic ./...

test-reports:
	@echo "📊 Generating comprehensive test reports..."
	@mkdir -p tmp/output
	@echo "Running tests with JSON output..."
	go test -v -json -race -coverprofile=tmp/output/coverage.out -covermode=atomic ./... | tee tmp/output/test-results.json
	@echo "Generating HTML coverage report..."
	go tool cover -html=tmp/output/coverage.out -o tmp/output/coverage.html
	@echo "Generating coverage summary..."
	go tool cover -func=tmp/output/coverage.out > tmp/output/coverage-summary.txt
	@echo "Test reports generated in tmp/output/"

test-unit:
	@echo "🔬 Running unit tests..."
	@mkdir -p tmp/output
	go test -v -race -coverprofile=tmp/output/unit-coverage.out -covermode=atomic ./internal/...

test-integration:
	@echo "🔗 Running integration tests..."
	@mkdir -p tmp/output
	go test -v -race -coverprofile=tmp/output/integration-coverage.out -covermode=atomic ./tests/integration/...

test-e2e:
	@echo "🎭 Running E2E tests..."
	@mkdir -p tmp/output
	go test -v -race -timeout=5m ./tests/e2e/...

# CI Pipeline
ci-local:
	@echo "🚀 Running local CI pipeline..."
	@echo "1. Formatting code..."
	@make fmt
	@echo "2. Running linter..."
	-@make lint
	@echo "3. Running comprehensive tests..."
	@make test-all
	@echo "4. Building application..."
	@make build
	@echo "✅ CI pipeline completed!"

ci-strict:
	@echo "🚀 Running strict CI pipeline..."
	@echo "1. Formatting code..."
	@make fmt
	@echo "2. Running linter (strict)..."
	@make lint
	@echo "3. Running comprehensive tests..."
	@make test-all
	@echo "4. Building application..."
	@make build
	@echo "✅ Strict CI pipeline completed successfully!"

ci-reports:
	@echo "🚀 Running CI with comprehensive reports..."
	@echo "1. Formatting code..."
	@make fmt
	@echo "2. Running linter..."
	-@make lint
	@echo "3. Running all tests with reports..."
	@make test-reports
	@echo "4. Running unit tests..."
	@make test-unit
	@echo "5. Running integration tests..."
	@make test-integration
	@echo "6. Running E2E tests..."
	@make test-e2e
	@echo "7. Building application..."
	@make build
	@echo "✅ CI pipeline with reports completed!"

# Code Quality
fmt:
	go fmt ./...

lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, running basic checks..."; \
		go vet ./...; \
	fi

# Release (uses git tags + GitHub Actions)
version:
	@git describe --tags --always 2>/dev/null || echo "v0.1.0"

release-patch:
	gh workflow run release.yml -f release_type=patch

release-minor:
	gh workflow run release.yml -f release_type=minor

release-major:
	gh workflow run release.yml -f release_type=major

release-pre:
	gh workflow run release.yml -f release_type=prerelease

# Utilities
clean:
	rm -rf bin/ tmp/

help:
	@echo "Commands:"
	@echo "  dev          - Run in development"
	@echo "  build        - Build binary"
	@echo "  test         - Run tests (quick)"
	@echo "  test-all     - Run comprehensive tests with coverage"
	@echo "  test-reports - Run tests with detailed reports"
	@echo "  test-unit    - Run unit tests only"
	@echo "  test-integration - Run integration tests only"
	@echo "  test-e2e     - Run E2E tests only"
	@echo "  ci-local     - Run CI pipeline (warnings allowed)"
	@echo "  ci-strict    - Run CI pipeline (strict mode)"
	@echo "  ci-reports   - Run CI with comprehensive test reports"
	@echo "  fmt          - Format code"
	@echo "  lint         - Run linter"
	@echo "  version      - Show current version"
	@echo "  release-*    - Trigger releases (patch/minor/major/pre)"
	@echo "  clean        - Clean build artifacts"

.PHONY: dev build test test-all test-reports test-unit test-integration test-e2e ci-local ci-strict ci-reports fmt lint version release-patch release-minor release-major release-pre clean help
