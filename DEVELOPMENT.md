# 🛠️ Development Guide

This guide covers development workflows, build modes, and debugging for contributors.

## 🚀 Quick Start

```bash
# Clone the repository
git clone https://github.com/JTBranch/SecurePasswordManager.git
cd SecurePasswordManager

# Install dependencies
go mod tidy

# Run in development mode
make dev
```

## 🔧 Build Modes

The application supports intelligent build mode detection to optimize for different environments:

### 🛠️ Development Mode

**Characteristics:**

- ✅ Debug logging enabled
- ✅ Verbose output for troubleshooting
- ✅ Local storage (project directory)
- ✅ Hot reload support

**Commands:**

```bash
make dev          # Run with debug logging
make build        # Build development binary
go run ./cmd      # Direct execution (auto-detects dev mode)
```

**Detection Logic:**

- Version = `"dev"` or `"development"`
- Running with `go run`
- No version injection

### 🚀 Production Mode

**Characteristics:**

- ❌ Debug logging disabled
- ✅ Clean output for end users
- ✅ OS-appropriate storage directories
- ✅ Optimized for distribution

**Commands:**

```bash
make build-prod   # Build production binary
```

**Detection Logic:**

- Version = semantic version (e.g., `"1.0.0"`)
- CI environment detected (`CI=true`)
- Cross-compiled binaries (fyne-cross)

## 🔍 Debug Logging

### Automatic Detection

The logging system automatically determines the appropriate log level:

| Build Type        | Debug Logs  | Version      | Storage Location |
| ----------------- | ----------- | ------------ | ---------------- |
| `make dev`        | ✅ Enabled  | `"dev"`      | Project root     |
| `make build`      | ✅ Enabled  | `"dev"`      | Project root     |
| `make build-prod` | ❌ Disabled | `"1.0.0"`    | OS config dir    |
| CI builds         | ❌ Disabled | Git tag      | OS config dir    |
| fyne-cross        | ❌ Disabled | `""` (empty) | OS config dir    |

### Manual Override

Force debug logging in any build:

```bash
DEV_MODE=true ./bin/password-manager
```

### Log Output Examples

**Development Mode:**

```
2025-08-18T12:15:46.350+0100    DEBUG   ui/app.go:38    Loaded window size from config, width: 1600, height: 900
2025-08-18T12:15:46.350+0100    DEBUG   service/secrets_service.go:83   Loading all secrets from file: | secrets.json
```

**Production Mode:**

```
(no debug output - clean user experience)
```

## 🏗️ Build Scripts

### macOS Builds (`scripts/build-macos.sh`)

- **Method**: Native Go builds with CGO
- **Architectures**: ARM64 (Apple Silicon) + AMD64 (Intel)
- **Version Injection**: ✅ Supported via ldflags
- **Debug Logging**: ❌ Disabled in releases

### Windows Builds (`scripts/build-windows.sh`)

- **Method**: fyne-cross (Docker-based)
- **Architectures**: AMD64
- **Version Injection**: ❌ Not supported (fyne-cross limitation)
- **Debug Logging**: ❌ Disabled (defaults to production)

### Linux Builds (`scripts/build-linux.sh`)

- **Method**: fyne-cross (Docker-based)
- **Architectures**: AMD64
- **Version Injection**: ❌ Not supported (fyne-cross limitation)
- **Debug Logging**: ❌ Disabled (defaults to production)

## 🧪 Testing

```bash
# Unit tests
make test-unit

# Integration tests
make test-integration

# End-to-end tests
make test-e2e

# All tests with coverage
make test-all
```

## 🔄 Hot Reload

For rapid development iteration:

```bash
# Install air (hot reload tool)
go install github.com/cosmtrek/air@latest

# Run with hot reload
make dev-watch
```

## 📦 Release Process

### Automated Releases

```bash
# Commit with conventional commit format
git commit -m "feat: add new feature"
git push

# CI automatically detects version bump and releases
```

### Manual Releases

```bash
make release-patch     # 1.0.0 → 1.0.1
make release-minor     # 1.0.0 → 1.1.0
make release-major     # 1.0.0 → 2.0.0
```

## 🔧 Environment Variables

| Variable                  | Purpose                  | Values            |
| ------------------------- | ------------------------ | ----------------- |
| `DEV_MODE`                | Force debug logging      | `true` / `false`  |
| `GO_PASSWORD_MANAGER_ENV` | Override environment     | `dev` / `prod`    |
| `CI`                      | CI environment detection | `true` (auto-set) |
| `GITHUB_ACTIONS`          | GitHub Actions detection | `true` (auto-set) |

## 📁 Project Structure

```
internal/
├── env/           # Environment & build mode detection
├── logger/        # Conditional logging system
├── config/        # Configuration management
├── service/       # Business logic
└── ...

scripts/
├── build-macos.sh     # Native builds with version injection
├── build-windows.sh   # Cross-compiled builds
└── build-linux.sh     # Cross-compiled builds
```

## 🐛 Debugging

### Local Development Issues

1. **Debug logs not appearing in production build:**

   ```bash
   # Check version detection
   ./bin/password-manager -version

   # Force debug mode
   DEV_MODE=true ./bin/password-manager
   ```

2. **Build failing on macOS:**

   ```bash
   # Ensure Xcode command line tools
   xcode-select --install
   ```

3. **fyne-cross issues:**

   ```bash
   # Update fyne-cross
   go install github.com/fyne-io/fyne-cross@latest

   # Clean Docker containers
   docker system prune -f
   ```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make changes and test: `make test-all`
4. Ensure clean production builds: `make build-prod`
5. Commit with conventional format: `git commit -m "feat: add amazing feature"`
6. Push and create a Pull Request

---

**Happy coding!** 🎉 For questions, check the [GitHub Issues](https://github.com/JTBranch/SecurePasswordManager/issues).
