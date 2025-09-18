# Go Password Manager - Simple Makefile

# Development
dev:
	go run ./cmd

run-dev:
	air 

build:
	go build -o bin/password-manager ./cmd

build-prod:
	go build -ldflags "-X main.version=1.0.0" -o bin/password-manager ./cmd

# Testing
test:
	go test ./...

test-all:
	@echo "🧪 Running comprehensive test suite..."
	rm -rf **/tmp/output
	rm -rf **/keys
	@mkdir -p tmp/output
	go test -v -race -coverprofile=tmp/output/coverage.out -covermode=atomic -coverpkg=./cmd/...,./internal/...,./ui/... ./...

test-coverage:
	@echo "📊 Running tests and showing coverage percentage..."
	@mkdir -p tmp/output
	@go test -race -coverprofile=tmp/output/coverage.out -covermode=atomic -coverpkg=./cmd/...,./internal/...,./ui/... ./... > /dev/null 2>&1
	@go tool cover -func=tmp/output/coverage.out | grep total | awk '{print "Total Coverage: " $$3}'

test-reports:
	@echo "📊 Generating comprehensive test reports..."
	@mkdir -p tmp/output
	@echo "Running tests with JSON output..."
	go test -v -json -race -coverprofile=tmp/output/coverage.out -covermode=atomic -coverpkg=./cmd/...,./internal/...,./ui/... ./... | tee tmp/output/test-results.json
	@echo "Generating HTML coverage report..."
	go tool cover -html=tmp/output/coverage.out -o tmp/output/coverage.html
	@echo "Generating coverage summary..."
	go tool cover -func=tmp/output/coverage.out > tmp/output/coverage-summary.txt
	@echo "Test reports generated in tmp/output/"

test-unit:
	@echo "🔬 Running unit tests..."
	rm -rf **/tmp/output
	rm -rf **/keys
	@mkdir -p tmp/output
	go test -v -race -coverprofile=tmp/output/unit-coverage.out -covermode=atomic -coverpkg=./internal/... ./internal/...

test-integration:
	@echo "🔗 Running integration tests..."
	rm -rf **/tmp/output
	rm -rf **/keys
	@mkdir -p tmp/output
	go test -v -race -coverprofile=tmp/output/integration-coverage.out -covermode=atomic -coverpkg=./internal/...,./ui/... ./tests/integration/...

test-e2e:
	@echo "🎭 Running E2E tests..."
	rm -rf **/tmp/output
	rm -rf **/keys
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

# Release Management
version:
	@git describe --tags --always 2>/dev/null || echo "v0.1.0"

release-patch:
	gh workflow run manual-release.yml -f version_bump=patch

release-minor:
	gh workflow run manual-release.yml -f version_bump=minor

release-major:
	gh workflow run manual-release.yml -f version_bump=major

release-prerelease:
	gh workflow run manual-release.yml -f version_bump=prerelease

# Utilities
clean:
	rm -rf bin/ tmp/

help:
	@echo "Development:"
	@echo "  dev              - Run in development mode"
	@echo "  build            - Build application"
	@echo ""
	@echo "Testing:"
	@echo "  test-unit        - Run unit tests with race detection"
	@echo "  test-integration - Run integration tests with coverage"
	@echo "  test-e2e         - Run E2E tests with detailed logging"
	@echo "  test-all         - Run all tests with comprehensive reporting"
	@echo "  ci-reports       - Run complete CI pipeline locally"
	@echo ""
	@echo "Code Quality:"
	@echo "  fmt              - Format code"
	@echo "  lint             - Run linting" 
	@echo "  vet              - Run go vet"
	@echo ""
	@echo "Release Management:"
	@echo "  release-patch    - Trigger patch release (1.0.0 -> 1.0.1)"
	@echo "  release-minor    - Trigger minor release (1.0.0 -> 1.1.0)"
	@echo "  release-major    - Trigger major release (1.0.0 -> 2.0.0)"
	@echo "  release-prerelease - Trigger prerelease (1.0.0 -> 1.0.1-rc.1)"
	@echo ""
	@echo "Utilities:"
	@echo "  clean            - Clean build artifacts"
	@echo "  dev-setup        - Install developer toolchain (protoc, plugins)"

dev-setup:
	@echo "🔧 Checking protoc..."
	@if ! command -v protoc >/dev/null 2>&1; then \
		PROTOC_VERSION=28.2; \
		UNAME_S=$$(uname -s); \
		UNAME_M=$$(uname -m); \
		case "$$UNAME_S" in Darwin) OS=osx ;; Linux) OS=linux ;; *) echo "Unsupported OS $$UNAME_S"; exit 1 ;; esac; \
		case "$$UNAME_M" in x86_64) ARCH=x86_64 ;; arm64|aarch64) ARCH=aarch_64 ;; *) echo "Unsupported arch $$UNAME_M"; exit 1 ;; esac; \
		if command -v brew >/dev/null 2>&1; then echo "Installing protobuf via brew"; brew install protobuf || true; fi; \
		if ! command -v protoc >/dev/null 2>&1; then \
			echo "Attempting direct download of protoc $$PROTOC_VERSION for $$OS-$$ARCH"; \
			URL="https://github.com/protocolbuffers/protobuf/releases/download/v$$PROTOC_VERSION/protoc-$$PROTOC_VERSION-$$OS-$$ARCH.zip"; \
			TMP=$$(mktemp -d); \
			curl -LsS $$URL -o $$TMP/protoc.zip || { echo "Download failed: $$URL"; exit 1; }; \
			unzip -q $$TMP/protoc.zip -d $$TMP/out; \
			mkdir -p .tools/protoc; cp -R $$TMP/out/* .tools/protoc/; \
			chmod +x .tools/protoc/bin/protoc; \
			echo "Installed local protoc to .tools/protoc/bin"; \
		fi; \
		if ! command -v protoc >/dev/null 2>&1; then PATH=$$(pwd)/.tools/protoc/bin:$$PATH protoc --version || true; fi; \
	else echo "protoc already installed"; fi
	@if ! command -v protoc >/dev/null 2>&1; then \
		if [ -x .tools/protoc/bin/protoc ]; then \
			echo "Using locally downloaded protoc"; \
		else \
			echo "❌ protoc still missing. Install manually: https://github.com/protocolbuffers/protobuf/releases"; exit 1; \
		fi; \
	fi
	@echo "🔧 Installing protoc-gen-go plugin..."
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@echo "🚀 Generating protobuf code..."
	@PATH=$$(pwd)/.tools/protoc/bin:$$PATH go generate ./internal/transport/proto
	@echo "✅ Dev setup complete"

.PHONY: dev build test test-all test-reports test-unit test-integration test-e2e ci-local ci-strict ci-reports fmt lint version release-patch release-minor release-major release-pre clean help dev-setup
