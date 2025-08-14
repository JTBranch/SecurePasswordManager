# Code Coverage Enforcement

This project enforces a minimum code coverage threshold to maintain code quality.

## Current Settings

- **Minimum Coverage**: 25.0%
- **Current Coverage**: ~26.5%
- **Coverage Status**: ✅ Passing

## How It Works

### Local Development

- Run `make coverage-check` to verify coverage meets the threshold
- Run `make ci-local` for full CI pipeline with coverage enforcement
- Coverage failure will exit with code 1 and block the build

### GitHub Actions

- Coverage is automatically checked on every push/PR
- Builds will fail if coverage drops below the minimum threshold
- Coverage reports are generated and uploaded as artifacts

## Adjusting the Threshold

You can customize the minimum coverage threshold:

```bash
# Set a higher threshold
make coverage-check MIN_COVERAGE=30.0

# Update the default in Makefile
# Edit the line: MIN_COVERAGE ?= 25.0
```

## Improving Coverage

To increase coverage, focus on these areas with 0% coverage:

- `internal/domain/secret.go` - Domain model methods
- `internal/logger/logger.go` - Logging utilities
- `internal/storage/storage.go` - Storage layer
- `internal/versioning/versioning.go` - Version management
- `ui/` modules - User interface components

## Coverage Reports

Coverage reports are generated in multiple formats:

- **HTML**: `tmp/output/coverage.html` - Interactive browser report
- **Text**: `tmp/output/coverage-summary.txt` - Function-level breakdown
- **Profile**: `tmp/output/coverage.out` - Go coverage profile

## Best Practices

1. **Incremental Improvement**: Gradually increase the threshold as you add tests
2. **Focus on Critical Code**: Prioritize testing business logic and error handling
3. **Quality over Quantity**: Write meaningful tests, not just coverage filler
4. **CI Integration**: Let the CI system enforce the threshold automatically

## Recommendations

- Start with 25% and aim for 80%+ over time
- Test critical paths: encryption, secret management, data persistence
- Mock external dependencies for reliable unit tests
- Use E2E tests for user workflows

Current coverage by module:

- Config: ~67.9%
- Crypto: ~64.3%
- Service: ~38.9%
- UI Molecules: ~17.9%
- Other modules: 0%
