#!/bin/sh
#
# Pre-commit hook for Go projects
# This hook automatically formats Go code and organizes imports before committing
#

echo "🔍 Running pre-commit checks..."

# Check if this is an initial commit
if git rev-parse --verify HEAD >/dev/null 2>&1
then
    against=HEAD
else
    # Initial commit: diff against an empty tree object
    against=$(git hash-object -t tree /dev/null)
fi

# Get list of Go files that are about to be committed
go_files=$(git diff --cached --name-only --diff-filter=ACM $against | grep '\.go$')

if [ -z "$go_files" ]; then
    echo "✅ No Go files to check"
    exit 0
fi

echo "📝 Found Go files to check:"
echo "$go_files" | sed 's/^/  - /'

# Install goimports if not available
if ! command -v goimports &> /dev/null; then
    echo "📦 Installing goimports..."
    go install golang.org/x/tools/cmd/goimports@latest
fi

format_failed=false

echo "🎨 Formatting Go files..."
for file in $go_files; do
    if [ -f "$file" ]; then
        echo "  Formatting: $file"
        
        # Format the file
        gofmt -w "$file"
        
        # Organize imports
        goimports -w "$file"
        
        # Check if file was modified
        if ! git diff --quiet "$file"; then
            echo "    ✅ Formatted and staged: $file"
            git add "$file"
        fi
    fi
done

echo "🔍 Checking for any remaining formatting issues..."

# Final check for formatting issues
unformatted=$(echo "$go_files" | xargs gofmt -l 2>/dev/null)
if [ -n "$unformatted" ]; then
    echo "❌ The following files still have formatting issues:"
    echo "$unformatted" | sed 's/^/  - /'
    format_failed=true
fi

# Check for import issues
import_issues=$(echo "$go_files" | xargs goimports -l 2>/dev/null)
if [ -n "$import_issues" ]; then
    echo "❌ The following files still have import issues:"
    echo "$import_issues" | sed 's/^/  - /'
    format_failed=true
fi

if [ "$format_failed" = true ]; then
    echo ""
    echo "❌ Pre-commit formatting check failed!"
    echo "💡 Files have been automatically formatted and staged."
    echo "💡 Please review the changes and commit again."
    exit 1
fi

echo "✅ All Go files are properly formatted!"
echo "✅ Pre-commit checks passed!"
exit 0
