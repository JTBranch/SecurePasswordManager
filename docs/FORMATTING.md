# Code Formatting and Linting

This project enforces consistent code formatting and style using Go's built-in tools and best practices.

## 🎨 Auto-Formatting Tools

### **gofmt** - Standard Go Formatter
- Formats Go code according to Go standards
- Handles indentation, spacing, and basic formatting

### **goimports** - Import Organization
- Automatically organizes imports
- Removes unused imports
- Adds missing imports
- Groups imports (stdlib, external, local)

## 🚀 Usage

### **Manual Formatting**
```bash
# Format all code and organize imports
make fmt

# Check if code is properly formatted (without changing files)
make fmt-check

# Check import organization
make imports-check
```

### **Auto-Format on Build**
```bash
# Build automatically formats code first
make build

# Build without formatting (for CI)
make build-only
```

### **Pre-Commit Hook** ⭐
```bash
# Install pre-commit hook (recommended!)
make install-hooks

# Remove pre-commit hook
make uninstall-hooks
```

The pre-commit hook will:
- ✅ Automatically format your code before each commit
- ✅ Organize imports properly
- ✅ Stage the formatted files
- ❌ Block commits if formatting fails

## 🔍 CI/CD Integration

### **Local CI**
```bash
# Full CI with formatting checks
make ci-local
```

### **GitHub Actions**
Formatting is automatically checked on:
- Every push to `main` or `develop`
- Every pull request to `main`

Builds will **fail** if:
- Code is not properly formatted
- Imports are not organized
- Coverage is below threshold

## 📋 What Gets Checked

### ✅ **Formatting Rules**
- Consistent indentation (tabs)
- Proper spacing around operators
- Consistent brace placement
- Line length recommendations
- Comment formatting

### ✅ **Import Organization**
```go
// Correct import order:
import (
    // 1. Standard library
    "fmt"
    "os"
    
    // 2. External packages
    "fyne.io/fyne/v2/app"
    "github.com/example/package"
    
    // 3. Local packages
    "go-password-manager/internal/config"
    "go-password-manager/ui/molecules"
)
```

### ✅ **Code Quality Checks**
- `go vet` - Static analysis
- `golint` - Style recommendations (53 current issues)
- Coverage enforcement (25%+ required)

## 🛠️ Setup for New Developers

1. **Install dependencies:**
   ```bash
   make install-deps
   ```

2. **Install pre-commit hook:**
   ```bash
   make install-hooks
   ```

3. **Verify setup:**
   ```bash
   make fmt-check
   make ci-local
   ```

## 💡 Best Practices

### **Before Committing**
- Run `make fmt` or rely on pre-commit hook
- Check `make fmt-check` passes
- Ensure `make ci-local` succeeds

### **Editor Integration**
Configure your editor to:
- Run `gofmt` on save
- Run `goimports` on save
- Show `go vet` results
- Highlight linting issues

### **Popular Editor Settings**

**VS Code** (`.vscode/settings.json`):
```json
{
  "go.formatTool": "goimports",
  "editor.formatOnSave": true,
  "go.lintOnSave": "package",
  "go.vetOnSave": "package"
}
```

**GoLand/IntelliJ**: Enable "Optimize imports" and "Reformat code" in commit dialog

## 📊 Current Status

- **Formatting**: ✅ All files properly formatted
- **Imports**: ✅ All imports organized  
- **Linting**: ⚠️ 53 issues to address (non-blocking)
- **Pre-commit**: ✅ Hook installed and active

## 🎯 Future Improvements

1. **Reduce golint issues**: Address the 53 current linting warnings
2. **Add golangci-lint**: More comprehensive linting suite
3. **Custom formatting rules**: Project-specific style preferences
4. **IDE integration docs**: Setup guides for popular editors

The formatting system ensures consistent, readable code across the entire project! 🎉
