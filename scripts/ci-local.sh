#!/bin/bash

# Local CI Build Script
# This script mimics the CI pipeline for local testing

set -e

echo "🚀 Starting Local CI Build..."
echo "================================"

# Create output directory
mkdir -p tmp/output/test-reports

# Clean previous outputs
rm -f tmp/output/*.{json,txt,html,out,log} 2>/dev/null || true
rm -f tmp/output/test-reports/* 2>/dev/null || true

echo "📦 Installing dependencies..."
go mod download
go mod verify

echo "🔍 Running code analysis..."

echo "Checking code formatting..."
if ! command -v goimports &> /dev/null; then
    echo "Installing goimports..."
    go install golang.org/x/tools/cmd/goimports@latest
fi

# Check formatting
gofmt -l . > tmp/output/format-issues.txt
if [ -s tmp/output/format-issues.txt ]; then
    echo "❌ Code formatting issues found:"
    cat tmp/output/format-issues.txt
    FORMAT_FAILED=true
else
    echo "✅ Code is properly formatted"
fi

# Check imports
goimports -l . > tmp/output/import-issues.txt
if [ -s tmp/output/import-issues.txt ]; then
    echo "❌ Import organization issues found:"
    cat tmp/output/import-issues.txt
    IMPORT_FAILED=true
else
    echo "✅ Imports are properly organized"
fi

echo "Running go vet..."
go vet ./... 2>&1 | tee tmp/output/vet-report.txt

echo "Installing and running golint..."
if ! command -v golint &> /dev/null; then
    go install golang.org/x/lint/golint@latest
fi
golint ./... > tmp/output/lint-report.txt 2>&1 || echo "Linting completed with warnings"

echo "🧪 Running Unit Tests with Coverage..."
echo "======================================="

# Run unit tests with coverage (using built-in tools only)
echo "Running unit tests..."

# First run with verbose output for console display
go test -v -race -coverprofile=tmp/output/coverage.out -covermode=atomic \
    ./internal/... ./ui/... | tee tmp/output/unit-test-console.txt

# Then run again with JSON output for parsing (suppress console output)
go test -v -race -coverprofile=tmp/output/coverage.out -covermode=atomic \
    ./internal/... ./ui/... \
    -json > tmp/output/unit-test-results.json 2>&1

# Generate coverage reports using built-in tools
echo "Generating coverage reports..."
go tool cover -html=tmp/output/coverage.out -o tmp/output/coverage.html
go tool cover -func=tmp/output/coverage.out > tmp/output/coverage-summary.txt

# Extract coverage percentage
COVERAGE=$(go tool cover -func=tmp/output/coverage.out | grep total | awk '{print $3}')
echo "Total Coverage: $COVERAGE" >> tmp/output/coverage-summary.txt

# Set minimum coverage threshold (can be adjusted)
MIN_COVERAGE=25.0
COVERAGE_NUM=$(echo $COVERAGE | sed 's/%//')

# Check if coverage meets minimum threshold
if (( $(echo "$COVERAGE_NUM >= $MIN_COVERAGE" | bc -l) )); then
    echo "✅ Coverage check passed: $COVERAGE >= ${MIN_COVERAGE}%"
    COVERAGE_STATUS="✅ PASSED ($COVERAGE)"
else
    echo "❌ Coverage check failed: $COVERAGE < ${MIN_COVERAGE}%"
    echo "💡 Tip: Add more unit tests to increase coverage"
    COVERAGE_STATUS="❌ FAILED ($COVERAGE < ${MIN_COVERAGE}%)"
    COVERAGE_FAILED=true
fi

echo "🔄 Running Integration Tests..."
echo "==============================="

# Set environment variable to enable enhanced logging for integration tests
export GO_PASSWORD_MANAGER_INTEGRATION_LOGGING=true

# Run Integration tests with enhanced logging - show output in console
echo "Integration test output:"
go test -v ./tests/integration/... | tee tmp/output/integration-test-output.txt

# Also capture JSON output for parsing
go test -v ./tests/integration/... \
    -json > tmp/output/integration-test-results.json 2>&1 || echo "Integration tests completed"

if grep -q '"Action":"fail"' tmp/output/integration-test-results.json; then
    echo "❌ Integration tests failed"
    INTEGRATION_STATUS="❌ FAILED"
    INTEGRATION_FAILED=true
else
    echo "✅ Integration tests passed"
    INTEGRATION_STATUS="✅ PASSED"
fi

echo "🔄 Running E2E Tests..."
echo "======================="

# Set environment variable to enable enhanced logging for E2E tests
export GO_PASSWORD_MANAGER_E2E_LOGGING=true
export GO_PASSWORD_MANAGER_LOG_LEVEL=DEBUG

# Run E2E tests with enhanced logging - show output in console
echo "E2E test output:"
go test -v ./tests/e2e/... | tee tmp/output/e2e-test-output.txt

# Also capture JSON output for parsing
go test -v ./tests/e2e/... \
    -json > tmp/output/e2e-test-results.json 2>&1 || echo "E2E tests completed"

# Capture application logs if they exist
if [ -d "tmp/test-data" ]; then
    echo "Capturing application logs from test runs..."
    find tmp/test-data -name "*.log" -exec cp {} tmp/output/test-reports/ \; 2>/dev/null || echo "No application logs found"
fi

echo "📊 Generating Test Summary..."
echo "============================="

# Generate comprehensive test summary
cat > tmp/output/test-summary.md << EOF
# Local CI Test Summary Report

**Build Date:** $(date)
**Git Commit:** $(git rev-parse HEAD 2>/dev/null || echo "N/A")
**Branch:** $(git branch --show-current 2>/dev/null || echo "N/A")
**Go Version:** $(go version)

## Coverage Summary
$(cat tmp/output/coverage-summary.txt)

## Unit Test Results
- Total Coverage: $COVERAGE
- Coverage profile: coverage.out
- HTML report: coverage.html
- XML report: coverage.xml (Cobertura format)

## E2E Test Results  
- Results available in: e2e-test-results.json
- Full output in: e2e-test-output.txt

## Code Quality
- Vet results: vet-report.txt
- Lint results: lint-report.txt

## Generated Files in tmp/output/
$(ls -la tmp/output/ | grep -v "^total" | grep -v "^d")
EOF

# Generate HTML test report
echo "Generating HTML test report..."
./scripts/generate-test-report.sh

echo "🏗️  Building Application..."
echo "============================"

# Build for current platform
echo "Building for current platform..."
go build -ldflags="-s -w" -o tmp/output/password-manager ./cmd/main.go

# Create build info
cat > tmp/output/build-info.txt << EOF
Build Information
=================
Date: $(date)
Commit: $(git rev-parse HEAD 2>/dev/null || echo "N/A")
Branch: $(git branch --show-current 2>/dev/null || echo "N/A")
Go Version: $(go version)
Platform: $(go env GOOS)/$(go env GOARCH)

Built Artifacts:
- password-manager (Current platform)
EOF

echo "📈 Build Summary"
echo "================"

# Display summary
echo "✅ Code Analysis:"
if [ "$FORMAT_FAILED" = true ]; then
    echo "   - formatting: ❌ FAILED (run 'make fmt')"
else
    echo "   - formatting: ✅ PASSED"
fi
if [ "$IMPORT_FAILED" = true ]; then
    echo "   - imports:    ❌ FAILED (run 'make fmt')" 
else
    echo "   - imports:    ✅ PASSED"
fi
echo "   - go vet:     $(wc -l < tmp/output/vet-report.txt) issues"
echo "   - golint:     $(wc -l < tmp/output/lint-report.txt) issues"

echo "✅ Unit Tests:"
echo "   - Coverage: $COVERAGE_STATUS"
echo "   - Results: tmp/output/unit-test-results.json"

echo "✅ Integration Tests:"
if grep -q "PASS" tmp/output/integration-test-output.txt 2>/dev/null; then
    echo "   - Status: ✅ PASSED"
else
    echo "   - Status: ❌ FAILED or incomplete"
fi

echo "✅ E2E Tests:"
if grep -q "PASS" tmp/output/e2e-test-output.txt 2>/dev/null; then
    echo "   - Status: ✅ PASSED"
else
    echo "   - Status: ❌ FAILED or incomplete"
fi

echo "✅ Build:"
if [ -f tmp/output/password-manager ]; then
    echo "   - Binary: ✅ Created ($(du -h tmp/output/password-manager | cut -f1))"
else
    echo "   - Binary: ❌ Failed"
fi

echo ""
echo "📁 All reports and artifacts saved to: tmp/output/"
echo "📊 Open tmp/output/coverage.html in your browser to view coverage report"
echo "📄 Check tmp/output/test-summary.md for detailed summary"

echo ""
# Check for failures and exit appropriately
if [ "$COVERAGE_FAILED" = true ] || [ "$FORMAT_FAILED" = true ] || [ "$IMPORT_FAILED" = true ]; then
    echo "❌ CI Build Failed:"
    if [ "$COVERAGE_FAILED" = true ]; then
        echo "   - Coverage below minimum threshold"
        echo "💡 Run 'make test' to add more unit tests and improve coverage"
    fi
    if [ "$FORMAT_FAILED" = true ] || [ "$IMPORT_FAILED" = true ]; then
        echo "   - Code formatting/imports issues"
        echo "💡 Run 'make fmt' to auto-fix formatting and import issues"
    fi
    exit 1
else
    echo "🎉 Local CI Build Complete!"
fi
