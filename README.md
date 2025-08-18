# Go Password Manager

A secure, cross-platform password manager built in Go with a modern Fyne UI. All secrets are encrypted and stored locally with full version history. Features include secret editing, version management, and a clean atomic UI design. Plans for third party secret conversion and perhaps a browser extension in the future. Will always be free, password security should be for everyone.

## AI-Assisted Development

This project serves as an **experiment in AI-assisted software development**, exploring how modern AI tools can h## Contributing

We welcome contributions! Please see [DEVELOPMENT.md](DEVELOPMENT.md) for detailed development guidelines including:

- 🔧 **Build Modes**: Development vs Production builds
- 🔍 **Debug Logging**: Automatic detection and manual overrides  
- 🧪 **Testing**: Comprehensive test suites and coverage
- 📦 **Release Process**: Automated and manual release workflows

**Quick start for contributors:**
- Fork the repo and create a feature branch
- Follow the atomic design and modular code guidelines
- Write unit and E2E tests for new features
- Ensure production builds work: `make build-prod`
- Submit a pull request with a clear description

For major changes, please open an issue first to discuss what you'd like to change.production-quality applications with robust automation, comprehensive testing, and high code quality standards. All major architectural decisions, code design patterns, and development direction have been made and supervised by the human developer, with AI assistance used as a productivity tool for implementation.

The development process leverages AI assistance for:

- **Architecture Implementation**: Translating architectural decisions into clean, modular patterns (atomic UI design, service layers)
- **Code Quality**: Automated formatting, linting, static analysis, and comprehensive test coverage
- **CI/CD Pipeline**: Automated builds, testing, coverage enforcement, and deployment workflows
- **Best Practices**: Implementing Go conventions, security patterns, and maintainable code structures

The goal is to demonstrate that AI-assisted development, under proper human guidance and supervision, can produce enterprise-grade software while maintaining code quality, security, and maintainability standards typically expected in production environments.

## 📥 Easy Installation

### 🎯 **One-Click Install** (Recommended for users)

**macOS:**
```bash
curl -fsSL https://raw.githubusercontent.com/JTBranch/SecurePasswordManager/main/install-macos.sh | bash
```

**Windows:**
Download and run: [install-windows.bat](https://raw.githubusercontent.com/JTBranch/SecurePasswordManager/main/install-windows.bat)

**Linux:**
```bash
curl -fsSL https://raw.githubusercontent.com/JTBranch/SecurePasswordManager/main/install-linux.sh | bash
```

### 📦 **Manual Download**

Visit [Releases](https://github.com/JTBranch/SecurePasswordManager/releases/latest) and download:
- **macOS Apple Silicon**: `go-password-manager-macos-arm64`
- **macOS Intel**: `go-password-manager-macos-amd64` 
- **Windows**: `password-manager-windows-amd64.exe`
- **Linux**: `password-manager-linux-amd64`

> 💡 **First-time users**: See [INSTALL.md](INSTALL.md) for detailed setup instructions

## Quick Start

```bash
# Development
make dev          # Run in development mode (with debug logs)
make build        # Build binary for development
make build-prod   # Build binary for production (no debug logs)
make test         # Run tests

# Releases (requires GitHub CLI)
make version      # Show current version
make release-patch    # v1.0.0 -> v1.0.1
make release-minor    # v1.0.0 -> v1.1.0
make release-major    # v1.0.0 -> v2.0.0
```

## 🔧 Build Modes

The application supports two distinct build modes to optimize for different environments:

### 🛠️ **Development Mode**
- **Debug Logging**: Full debug output for troubleshooting
- **Local Storage**: Secrets stored in project directory  
- **Hot Reload**: Live code updates (with `make dev-watch`)
- **Version Detection**: Automatically detected when running from source

```bash
make dev          # Run with debug logging
make build        # Build development binary
go run ./cmd      # Direct execution (debug mode)
```

### 🚀 **Production Mode**  
- **Clean Output**: No debug logs for end users
- **Secure Storage**: OS-appropriate config directories
- **Optimized**: Reduced binary size and logging overhead
- **Version Injection**: Build-time version embedding

```bash
make build-prod   # Build production binary (no debug logs)
```

### 🔀 **Build Mode Detection**

The application automatically detects its running mode:

| Build Type | Debug Logs | Detection Method |
|------------|------------|------------------|
| `make dev` / `go run` | ✅ Enabled | Version = "dev" |
| `make build` | ✅ Enabled | Version = "dev" |
| `make build-prod` | ❌ Disabled | Version = "1.0.0" |
| CI/Release builds | ❌ Disabled | Version = tag/CI environment |
| fyne-cross builds | ❌ Disabled | No version = production default |

**Manual Override**: Set `DEV_MODE=true` environment variable to force debug logging in any build.

## Testing

This project uses a suite of automated tests to ensure code quality and stability.

### Running Tests

You can run all tests using the following command:

```bash
make test
```

This will execute unit, integration, and end-to-end (E2E) tests.

### Writing Integration Tests

To ensure consistency and reduce boilerplate, all new integration tests **must** use the `IntegrationTestSuite` helper. This suite handles the setup and teardown of the test environment, including creating temporary data directories and initializing services.

**Example Usage:**

Here is an example of how to structure a new integration test:

```go
package integration

import (
	"testing"

	"go-password-manager/pkg/reporting"
	"go-password-manager/tests/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestYourNewFeature(t *testing.T) {
	reporting.WithReporting(t, "TestYourNewFeature", func(reporter *reporting.TestWrapper) {
		// 1. Initialize the test suite
		suite := helpers.NewIntegrationTestSuite(reporter)
		suite.SetupTestEnvironment()
		defer suite.Cleanup()

		// 2. Access services from the suite
		// Example: Load secrets using the suite's service
		secrets, err := suite.SecretsService.LoadLatestSecrets()
		require.NoError(t, err)
		assert.Empty(t, secrets.Secrets, "Should start with no secrets")

		// 3. Write your test logic here...
	})
}
```

**Key Principles:**

- **Use `helpers.NewIntegrationTestSuite(reporter)`**: This creates a new test suite.
- **Call `suite.SetupTestEnvironment()`**: This prepares the test environment (e.g., temporary directories).
- **Use `defer suite.Cleanup()`**: This ensures the test environment is cleaned up after the test runs.
- **Access services via `suite.*`**: Use `suite.SecretsService`, `suite.ConfigService`, etc., to interact with the application's components.

## Automated Release System

This project features a fully automated release pipeline that creates multi-platform binaries and GitHub releases:

### 🤖 Automatic Releases (Recommended)

The system automatically detects version bumps from your commit messages using [Conventional Commits](https://www.conventionalcommits.org/):

```bash
# Patch release (1.0.0 → 1.0.1)
git commit -m "fix: resolve authentication issue"

# Minor release (1.0.0 → 1.1.0)
git commit -m "feat: add secret export functionality"

# Major release (1.0.0 → 2.0.0)
git commit -m "feat!: redesign storage format"
# or
git commit -m "feat: add new feature

BREAKING CHANGE: changes the API interface"
```

**How it works:**

1. Push commits with conventional prefixes (`feat:`, `fix:`, etc.)
2. CI pipeline runs comprehensive tests (unit, integration, E2E)
3. On CI success, release workflow automatically triggers
4. Version is determined from commit messages
5. Multi-platform binaries are built and released

### 🛠️ Manual Releases

For immediate releases or custom versioning:

```bash
make release-patch      # 1.0.0 → 1.0.1 (bug fixes)
make release-minor      # 1.0.0 → 1.1.0 (new features)
make release-major      # 1.0.0 → 2.0.0 (breaking changes)
make release-prerelease # 1.0.0 → 1.0.1-rc.1 (pre-release)
```

### 📦 Multi-Platform Builds

Every release automatically generates optimized binaries for:

- **Linux x64** (`GOOS=linux GOARCH=amd64`)
- **macOS ARM64** (`GOOS=darwin GOARCH=arm64`) - Apple Silicon  
- **macOS Intel** (`GOOS=darwin GOARCH=amd64`) - Intel processors
- **Windows x64** (`GOOS=windows GOARCH=amd64`)

**Production Quality:**
- ✅ **No debug logs** - Clean user experience
- ✅ **Version injection** - Proper version detection (macOS builds)
- ✅ **Optimized storage** - OS-appropriate config directories
- ✅ **Cross-compilation** - fyne-cross for Windows/Linux, native builds for macOS

### 🔄 Release Workflow

1. **Code Changes** → Push to `main` branch
2. **CI Pipeline** → Runs all tests with coverage reporting
3. **Version Detection** → Analyzes commit messages for version bump
4. **Release Trigger** → Only activates after successful CI completion
5. **Multi-Platform Build** → Compiles binaries for all target platforms
6. **GitHub Release** → Creates release with binaries and auto-generated notes

### 📋 Monitoring Releases

Track release progress with GitHub CLI:

```bash
# View CI pipeline status
gh run list --workflow="ci.yml"

# View automatic releases
gh run list --workflow="release.yml"

# View manual releases
gh run list --workflow="manual-release.yml"

# View latest releases
gh release list
```

### ⚙️ Release Configuration

The release system is configured via GitHub Actions workflows:

- **`.github/workflows/ci.yml`** - Comprehensive testing pipeline
- **`.github/workflows/release.yml`** - Automated release on CI success
- **`.github/workflows/manual-release.yml`** - Manual release triggers

**Prerequisites for releases:**

- GitHub CLI installed (`gh` command)
- Repository permissions: Settings → Actions → General → "Read and write permissions"
- Valid GitHub token with workflow access

- **Secure Local Storage**: AES encrypted secrets stored locally
- **Version History**: Full version tracking with ability to view previous secret values
- **Secret Management**: Create, edit, view, and delete secrets with ease
- **Atomic UI Design**: Clean component structure (pages, molecules, atoms)
- **Development Tools**: Hot reload, comprehensive testing, easy build system
- **Cross-Platform**: Runs on macOS, Linux, and Windows
- **Environment-Aware**: Separate development and production storage locations

## Getting Started

### Prerequisites

- Go 1.20+
- [Fyne](https://fyne.io/) (UI library)

### Installation

Clone the repo:

```sh
git clone https://github.com/JTBranch/SecurePasswordManager.git
cd SecurePasswordManager
```

Install dependencies:

```sh
go mod tidy
```

### Quick Start

The easiest way to get started is using the Makefile:

```sh
# Run in development mode (with debug logging)
make dev

# Run with hot reload (install air first with `make install-deps`)
make dev-watch

# Build development binary (includes debug logging)
make build

# Build production binary (clean output, no debug logs)
make build-prod
```

### Available Commands

| Command                   | Description                                    |
| ------------------------- | ---------------------------------------------- |
| **Development**           |                                                |
| `make dev`                | Run in development mode with debug logging     |
| `make build`              | Build development binary with debug logging    |
| `make build-prod`         | Build production binary (no debug logs)       |
| **Testing**               |                                                |
| `make test-unit`          | Run unit tests with race detection             |
| `make test-integration`   | Run integration tests with coverage            |
| `make test-e2e`           | Run end-to-end tests with detailed logging     |
| `make test-all`           | Run all tests with comprehensive reporting     |
| `make ci-reports`         | Run complete CI pipeline locally               |
| **Code Quality**          |                                                |
| `make fmt`                | Format code with go fmt                        |
| `make lint`               | Run linting (golangci-lint or go vet)          |
| **Release Management**    |                                                |
| `make version`            | Show current git version                       |
| `make release-patch`      | Trigger patch release (1.0.0 → 1.0.1)          |
| `make release-minor`      | Trigger minor release (1.0.0 → 1.1.0)          |
| `make release-major`      | Trigger major release (1.0.0 → 2.0.0)          |
| `make release-prerelease` | Trigger prerelease (1.0.0 → 1.0.1-rc.1)        |
| **Utilities**             |                                                |
| `make clean`              | Remove build artifacts and test reports        |
| `make help`               | Show detailed help with all available commands |

### Environment Variables

The application supports different environments and logging modes:

#### **Build-Time Configuration**

- **Development Builds** (`make dev`, `make build`)
  - Debug logging: **Enabled**
  - Version detection: `"dev"`
  - Storage: Project root directory

- **Production Builds** (`make build-prod`, CI releases)
  - Debug logging: **Disabled** 
  - Version injection: Semantic version (e.g., `"1.0.0"`)
  - Storage: OS user config directory

#### **Runtime Environment Variables**

- **`DEV_MODE=true`** - Force enable debug logging in any build
- **`GO_PASSWORD_MANAGER_ENV=prod`** - Force production storage location
- **`CI=true`** - Detected automatically in CI environments (forces production mode)

#### **Storage Locations**

**Development Mode:**
- `secrets.json` stored in project root
- Debug logging enabled
- Suitable for development and testing

**Production Mode:**
- `secrets.json` stored in OS user config directory:
  - **macOS:** `~/Library/Application Support/GoPasswordManager/secrets.json`
  - **Linux:** `~/.config/GoPasswordManager/secrets.json`
  - **Windows:** `%APPDATA%\GoPasswordManager\secrets.json`
- Clean output (no debug logs)
- Secure file permissions

### YAML Configuration

The application now supports structured configuration using YAML files for better organization and environment management. This system works alongside the existing environment variables for backward compatibility.

#### **Configuration Files**

Configuration files are located in the `configs/` directory:

```
configs/
├── default.yaml      # Base configuration (all environments)
├── development.yaml  # Development overrides  
├── production.yaml   # Production overrides
└── test.yaml        # Test environment overrides
```

#### **Environment Selection**

Set the environment using the `GO_PASSWORD_MANAGER_ENV` variable:

```bash
export GO_PASSWORD_MANAGER_ENV=production  # Use production config
export GO_PASSWORD_MANAGER_ENV=development # Use development config (default)
export GO_PASSWORD_MANAGER_ENV=test       # Use test config
```

#### **Configuration Priority**

Configuration is loaded in this order (later values override earlier ones):

1. **YAML Default**: `configs/default.yaml`
2. **YAML Environment**: `configs/{environment}.yaml` 
3. **Environment Variables**: Traditional env vars (highest priority)

#### **Example Configuration**

```yaml
# configs/production.yaml
application:
  environment: "production"
  
logging:
  debug: false
  level: "info"
  format: "json"
  
ui:
  window:
    width: 1400
    height: 800
```

For detailed configuration options and migration guide, see [YAML_CONFIG.md](docs/YAML_CONFIG.md).

### Data Structure

Secrets are stored with full version history in a nested JSON structure:

```json
{
  "secrets": [
    {
      "secretName": "example",
      "type": "key_value",
      "currentVersion": 2,
      "versions": [
        {
          "secretValueEnc": "encrypted_value_v1",
          "version": 1,
          "updatedAt": "2025-08-14T10:29:00+01:00"
        },
        {
          "secretValueEnc": "encrypted_value_v2",
          "version": 2,
          "updatedAt": "2025-08-14T10:30:00+01:00"
        }
      ]
    }
  ]
}
```

This structure allows you to:

- Track all changes to secrets over time
- View previous versions of any secret
- Maintain audit trail of when secrets were modified

### Testing

The project includes comprehensive testing with detailed reporting:

```sh
# Run individual test suites
make test-unit          # Unit tests with race detection
make test-integration   # Integration tests with coverage
make test-e2e          # End-to-end tests with detailed logging

# Run all tests with HTML reports
make test-all          # Comprehensive test suite with coverage reports

# Run complete CI pipeline locally
make ci-reports        # Full CI pipeline with all reports generated
```

**Test Reports:**
All tests generate detailed reports in the `tmp/output/` directory:

- `coverage.html` - Interactive HTML coverage report
- `coverage-summary.txt` - Coverage percentage summary
- `test-results.json` - Detailed test execution results

These reports are automatically generated in CI builds and available for download as artifacts.

## CI/CD Pipeline

The project features a comprehensive GitHub Actions pipeline with:

### 🧪 **Comprehensive Testing**

- **Unit Tests**: Fast, isolated component testing with race detection
- **Integration Tests**: Cross-component functionality testing with coverage
- **E2E Tests**: Full application workflow testing with GUI simulation
- **Coverage Reporting**: Integrated with [Codecov](https://codecov.io) for coverage tracking

### 📊 **Quality Assurance**

- **Code Coverage**: Minimum coverage thresholds with detailed HTML reports
- **Race Detection**: Concurrent access testing with `-race` flag
- **Linting**: Code quality checks with `golangci-lint`
- **Formatting**: Automatic code formatting validation

### 🚀 **Automated Deployment**

- **Multi-Platform Builds**: Linux, macOS (ARM64 + Intel), Windows binaries
- **Production Quality**: Clean output with no debug logs in releases
- **Version Injection**: Automatic version embedding for macOS builds
- **Smart Detection**: fyne-cross builds default to production mode
- **Semantic Versioning**: Automatic version bumping via conventional commits
- **Release Automation**: CI-dependent releases with comprehensive artifacts
- **GitHub Releases**: Automatic release creation with binaries and notes

### 📁 **Artifact Management**

- **Test Reports**: HTML coverage reports and JSON test results
- **Build Artifacts**: Multi-platform binaries with 30-day retention
- **Release Assets**: Optimized production binaries attached to releases

## Application Features

### Secret Management

- **Create Secrets**: Add new encrypted secrets with automatic versioning
- **Edit Secrets**: Modify existing secrets while preserving version history
- **View Secrets**: Reveal secret values with click-to-show functionality
- **Delete Secrets**: Remove secrets entirely (including all versions)

### Version History

- **Track Changes**: Every edit creates a new version with timestamp
- **View History**: Browse previous versions of any secret
- **Restore Values**: Copy previous versions for restoration
- **Audit Trail**: Complete history of when secrets were modified

### User Interface

- **Clean Design**: Atomic UI components for consistent experience
- **Responsive Layout**: Adapts to different window sizes
- **Intuitive Navigation**: Easy-to-use buttons and forms
- **Security First**: Values hidden by default with reveal functionality

## Project Structure

### Secret Management

- **Create Secrets**: Add new encrypted secrets with automatic versioning
- **Edit Secrets**: Modify existing secrets while preserving version history
- **View Secrets**: Reveal secret values with click-to-show functionality
- **Delete Secrets**: Remove secrets entirely (including all versions)

### Version History

- **Track Changes**: Every edit creates a new version with timestamp
- **View History**: Browse previous versions of any secret
- **Restore Values**: Copy previous versions for restoration
- **Audit Trail**: Complete history of when secrets were modified

### User Interface

- **Clean Design**: Atomic UI components for consistent experience
- **Responsive Layout**: Adapts to different window sizes
- **Intuitive Navigation**: Easy-to-use buttons and forms
- **Security First**: Values hidden by default with reveal functionality

## Project Structure

```
go-password-manager/
├── cmd/                    # Application entrypoint
│   └── main.go            # Main application file
├── internal/              # Core application logic
│   ├── config/            # Configuration management
│   ├── crypto/            # Encryption/decryption
│   ├── domain/            # Data models and types
│   ├── env/               # Environment detection & build modes
│   ├── logger/            # Conditional logging system
│   ├── service/           # Business logic services
│   ├── storage/           # File storage handling
│   └── versioning/        # Version management
├── ui/                    # User interface components
│   ├── atoms/             # Smallest UI components
│   ├── molecules/         # Composite UI components
│   ├── pages/             # Full page layouts
│   └── e2e/               # UI end-to-end tests
├── tests/                 # Test suites
│   └── e2e/               # End-to-end workflow tests
├── scripts/               # Build automation
│   ├── build-macos.sh     # Native macOS builds with version injection
│   ├── build-windows.sh   # Cross-compiled Windows builds  
│   └── build-linux.sh     # Cross-compiled Linux builds
├── Makefile              # Build and development commands
├── .air.toml             # Hot reload configuration
└── README.md             # This file
```

### Architecture Principles

- **Atomic Design**: UI components organized in atoms → molecules → pages hierarchy
- **Clean Architecture**: Clear separation between domain, service, and UI layers
- **Security First**: All secrets encrypted at rest with AES encryption
- **Version Control**: Complete audit trail of all secret modifications
- **Environment Aware**: Separate development and production configurations
- **Build Mode Detection**: Intelligent debug logging based on build context
- **Production Ready**: Clean output in releases, verbose logging in development

## Contributing

We welcome contributions! Please:

- Fork the repo and create a feature branch
- Follow the atomic design and modular code guidelines
- Write unit and E2E tests for new features
- Submit a pull request with a clear description

For major changes, please open an issue first to discuss what you’d like to change.

## License

MIT
