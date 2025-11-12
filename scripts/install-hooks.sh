#!/bin/bash
# Install git hooks for nexus-cli

set -e

HOOK_DIR=".git/hooks"
HOOK_FILE="$HOOK_DIR/pre-commit"

echo "Installing pre-commit hook..."

cat > "$HOOK_FILE" << 'EOF'
#!/bin/bash
# Pre-commit hook for nexus-cli
# Runs linting and formatting checks before commit

set -e

echo "Running pre-commit checks..."

# Add GOPATH/bin to PATH
export PATH=$PATH:$(go env GOPATH)/bin

# Get list of staged Go files
STAGED_GO_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$' || true)

if [ -z "$STAGED_GO_FILES" ]; then
    echo "No Go files to check"
    exit 0
fi

echo "Checking $(echo "$STAGED_GO_FILES" | wc -l) Go files..."

# Format code
echo "Running gofmt..."
for file in $STAGED_GO_FILES; do
    gofmt -w -s "$file"
    git add "$file"
done

# Run goimports if available
if command -v goimports &> /dev/null; then
    echo "Running goimports..."
    for file in $STAGED_GO_FILES; do
        goimports -w "$file"
        git add "$file"
    done
else
    echo "Warning: goimports not found, skipping..."
fi

# Run go vet
echo "Running go vet..."
if ! go vet ./...; then
    echo "❌ go vet failed"
    exit 1
fi

# Run golangci-lint if available
if command -v golangci-lint &> /dev/null; then
    echo "Running golangci-lint..."
    if ! golangci-lint run --timeout=5m; then
        echo "❌ golangci-lint failed"
        echo "💡 Run 'golangci-lint run --fix' to auto-fix some issues"
        exit 1
    fi
else
    echo "Warning: golangci-lint not found, skipping..."
    echo "Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
fi

# Run tests
echo "Running tests..."
if ! go test ./... -short; then
    echo "❌ Tests failed"
    exit 1
fi

echo "✅ All checks passed!"
exit 0
EOF

chmod +x "$HOOK_FILE"

echo "✅ Pre-commit hook installed successfully!"
echo ""
echo "The hook will run automatically before each commit."
echo "To bypass the hook, use: git commit --no-verify"
echo ""
echo "To uninstall, run: rm .git/hooks/pre-commit"
