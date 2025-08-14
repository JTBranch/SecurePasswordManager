# Development Dependencies

This project separates CI/build dependencies from local development dependencies to avoid conflicts and keep CI builds lean.

## 🚀 Quick Setup

### **For CI/Build** (minimal dependencies)

```bash
make install-deps
```

### **For Local Development** (includes watch mode)

```bash
make install-dev-deps
```

## 📦 Dependency Categories

### **CI/Build Dependencies** (`make install-deps`)

These are the minimal dependencies required for building and CI:

- **golint** - Code linting
- **goimports** - Import organization and formatting

Used by:

- GitHub Actions CI
- Local CI pipeline (`make ci-local`)
- Build process (`make build`)

### **Local Development Dependencies** (`make install-dev-deps`)

Additional tools for local development:

- **golint** - Code linting
- **goimports** - Import organization and formatting
- **air** - Hot reload for development (`make dev-watch`)

### **System Dependencies** (`make install-system-deps`)

System libraries required for Fyne applications (Ubuntu/Debian):

- **libgl1-mesa-dev** - OpenGL development files
- **xorg-dev** - X.Org development files  
- **libx11-dev** - X11 client-side library
- **libxcursor-dev** - X cursor management library
- **libxrandr-dev** - X RandR extension library
- **libxinerama-dev** - X Xinerama extension library
- **libxi-dev** - X Input extension library
- **libxxf86vm-dev** - X Video Mode extension library
- **libglu1-mesa-dev** - OpenGL Utility library

**Note**: On macOS and Windows, these are typically pre-installed or provided by the system.

## 🔧 Tool Details

### **Air (Hot Reload)**

- **Package**: `github.com/air-verse/air@latest` ✅
- **Purpose**: Watch mode for development
- **Usage**: `make dev-watch`
- **Note**: Automatically installed when needed

**Migration Note**: The original `github.com/cosmtrek/air` package has moved to `github.com/air-verse/air`. Our setup now uses the correct new repository.

### **Linting Tools**

- **golint**: Style recommendations
- **goimports**: Import organization
- **go vet**: Built-in static analysis (no installation needed)

## 💡 Usage Examples

### **Initial Setup (New Developer)**

```bash
# Clone repository
git clone <repo>
cd go-password-manager

# Install development dependencies
make install-dev-deps

# Install pre-commit hooks
make install-hooks

# Start development
make dev-watch
```

### **CI Environment**

```bash
# Install minimal dependencies
make install-deps

# Run CI pipeline
make ci-local
```

### **Development Workflow**

```bash
# Regular development
make dev

# Watch mode (auto-reload on changes)
make dev-watch

# Format code
make fmt

# Check everything before commit
make ci-local
```

## 🚫 Removed Dependencies

These packages were removed due to version conflicts:

- `github.com/axw/gocov/gocov` - Coverage conversion (not needed for CI)
- `github.com/AlekSi/gocov-xml` - XML coverage reports (not needed for CI)
- `github.com/matm/gocov-html` - HTML coverage reports (not needed for CI)

**Note**: We use Go's built-in `go tool cover` for coverage reports instead, which is more reliable and doesn't require external dependencies.

## 🔍 Troubleshooting

### **Air Not Found**

```bash
# Install development dependencies
make install-dev-deps

# Or auto-install when running watch mode
make dev-watch
```

### **CI Dependency Conflicts**

The CI uses minimal dependencies to avoid conflicts:

```bash
# This should always work in CI
make install-deps
make ci-local
```

### **Version Conflicts**

If you encounter Go module version conflicts:

```bash
# Clean module cache
go clean -modcache

# Reinstall dependencies
make install-dev-deps
```

### **X11/OpenGL Build Errors**

If you see errors like `fatal error: X11/Xlib.h: No such file or directory`:

**Ubuntu/Debian**:
```bash
make install-system-deps
```

**Other Linux Distributions**:
```bash
# Fedora/RHEL
sudo dnf install libX11-devel libXcursor-devel libXrandr-devel libXinerama-devel libXi-devel libGL-devel

# Arch Linux  
sudo pacman -S libx11 libxcursor libxrandr libxinerama libxi mesa

# Alpine Linux
sudo apk add libx11-dev libxcursor-dev libxrandr-dev libxinerama-dev libxi-dev mesa-dev
```

**macOS/Windows**: 
These errors typically don't occur as system dependencies are pre-installed.

## 📊 Current Status

- ✅ **CI Dependencies**: Clean, minimal, conflict-free
- ✅ **System Dependencies**: X11/OpenGL libraries handled in CI
- ✅ **Air Package**: Updated to new repository (`air-verse/air`)
- ✅ **Hot Reload**: Working (`make dev-watch`)
- ✅ **GitHub Actions**: No more dependency conflicts
- ✅ **Local Development**: Full feature set available

The dependency separation ensures reliable CI builds while maintaining full development capabilities locally! 🎉
