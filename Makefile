# Configuration
MIN_COVERAGE ?= 25.0

# Development
dev:
	GO_PASSWORD_MANAGER_ENV=dev go run ./cmd/main.go

dev-watch:
	@if ! command -v air &> /dev/null; then \
		echo "Installing air for watch mode..."; \
		go install github.com/air-verse/air@latest; \
	fi
	GO_PASSWORD_MANAGER_ENV=dev air

# Build
build: fmt
	go build -o main ./cmd/main.go

build-release:
	mkdir -p tmp/output
	go build -ldflags="-s -w" -o tmp/output/password-manager ./cmd/main.go

run:
	./main

# Testing
test:
	go test ./... -v -shuffle=on

test-unit:
	go test ./internal/... ./ui/... -v -shuffle=on

test-integration:
	mkdir -p tmp/output/test-reports
	go test ./tests/integration -v -shuffle=on -json > tmp/output/integration-test-results.json 2>&1 || true
	go test ./tests/integration -v -shuffle=on > tmp/output/integration-test-output.txt 2>&1 || true

test-e2e:
	mkdir -p tmp/output/test-reports
	go test ./tests/e2e/tests -v -shuffle=on -json > tmp/output/e2e-test-results.json 2>&1 || true
	go test ./tests/e2e/tests -v -shuffle=on > tmp/output/e2e-test-output.txt 2>&1 || true

test-all: test-unit test-integration test-e2e
	@echo "Generating test report..."
	@./scripts/generate-test-report.sh || echo "Failed to generate test report"

# Coverage
coverage:
	mkdir -p tmp/output
	go test ./internal/... ./ui/... -coverprofile=tmp/output/coverage.out -covermode=atomic -shuffle=on
	go tool cover -html=tmp/output/coverage.out -o tmp/output/coverage.html
	go tool cover -func=tmp/output/coverage.out

coverage-check: coverage
	@COVERAGE=$$(go tool cover -func=tmp/output/coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	if [ $$(echo "$$COVERAGE >= $(MIN_COVERAGE)" | bc -l) -eq 1 ]; then \
		echo "✅ Coverage check passed: $$COVERAGE% >= $(MIN_COVERAGE)%"; \
	else \
		echo "❌ Coverage check failed: $$COVERAGE% < $(MIN_COVERAGE)%"; \
		echo "💡 Add more unit tests to increase coverage"; \
		exit 1; \
	fi

# Code Quality
fmt:
	gofmt -w .
	goimports -w .

fmt-check:
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "❌ Code is not formatted. Run 'make fmt' to fix:"; \
		gofmt -l .; \
		exit 1; \
	else \
		echo "✅ Code is properly formatted"; \
	fi

imports-check:
	@if [ -n "$$(goimports -l .)" ]; then \
		echo "❌ Imports are not organized. Run 'make fmt' to fix:"; \
		goimports -l .; \
		exit 1; \
	else \
		echo "✅ Imports are properly organized"; \
	fi

lint:
	go vet ./...
	@if command -v golint &> /dev/null; then \
		golint ./...; \
	else \
		echo "Note: Install golint for additional checks: go install golang.org/x/lint/golint@latest"; \
	fi

lint-advanced:
	@if command -v golangci-lint &> /dev/null; then \
		golangci-lint run; \
	else \
		echo "❌ golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

# CI/CD
ci-local:
	./scripts/ci-local.sh

# Setup
install-deps:
	go install golang.org/x/lint/golint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

install-dev-deps: install-deps
	go install github.com/air-verse/air@latest

install-system-deps:
	@if command -v apt-get &> /dev/null; then \
		echo "Installing system dependencies for Fyne..."; \
		./scripts/install-system-deps.sh; \
	else \
		echo "System dependency installation only supported on Ubuntu/Debian"; \
	fi

install-hooks:
	@if [ -d .git ]; then \
		cp scripts/pre-commit-hook.sh .git/hooks/pre-commit; \
		chmod +x .git/hooks/pre-commit; \
		echo "✅ Pre-commit hook installed"; \
	else \
		echo "❌ Not a git repository"; \
	fi

# Utilities
clean:
	rm -f main tmp/main
	rm -rf tmp/output/*

help:
	@echo "Development:"
	@echo "  dev              - Run in development mode"
	@echo "  dev-watch        - Run with auto-reload"
	@echo "  build            - Build application"
	@echo "  build-release    - Build optimized release binary"
	@echo ""
	@echo "Testing:"
	@echo "  test             - Run all tests"
	@echo "  test-unit        - Run unit tests only"
	@echo "  test-integration - Run integration tests only"
	@echo "  test-e2e         - Run E2E tests only"
	@echo "  test-all         - Run all tests with report"
	@echo "  coverage         - Generate coverage report"
	@echo "  coverage-check   - Check coverage threshold ($(MIN_COVERAGE)%)"
	@echo ""
	@echo "Code Quality:"
	@echo "  fmt              - Format code"
	@echo "  fmt-check        - Check code formatting"
	@echo "  imports-check    - Check import organization"
	@echo "  lint             - Run basic linting"
	@echo "  lint-advanced    - Run advanced linting"
	@echo ""
	@echo "CI/Setup:"
	@echo "  ci-local         - Run full CI pipeline"
	@echo "  install-deps     - Install build dependencies"
	@echo "  install-dev-deps - Install dev dependencies"
	@echo "  install-hooks    - Install git hooks"
	@echo "  clean            - Clean build artifacts"

.PHONY: dev dev-watch build build-release run test test-unit test-integration test-e2e test-all coverage coverage-check fmt fmt-check imports-check lint lint-advanced ci-local install-deps install-dev-deps install-system-deps install-hooks clean help
