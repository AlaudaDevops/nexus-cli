#!/bin/bash
# Fix linting issues automatically

set -e

echo "=== Running gofmt ==="
gofmt -w .

echo "=== Running goimports ==="
export PATH=$PATH:$(go env GOPATH)/bin
goimports -w .

echo "=== Adding package comments ==="

# Add package comment to cmd package
if ! grep -q "^// Package cmd" cmd/root.go 2>/dev/null; then
    echo '// Package cmd provides command-line interface for Nexus CLI.' | cat - cmd/root.go > cmd/root.go.tmp && mv cmd/root.go.tmp cmd/root.go
fi

# Add package comment to config package
if ! grep -q "^// Package config" pkg/config/types.go 2>/dev/null; then
    echo '// Package config provides configuration types and loading functionality.' | cat - pkg/config/types.go > pkg/config/types.go.tmp && mv pkg/config/types.go.tmp pkg/config/types.go
fi

# Add package comment to nexus package
if ! grep -q "^// Package nexus" pkg/nexus/client.go 2>/dev/null; then
    echo '// Package nexus provides client for Sonatype Nexus Repository Manager API.' | cat - pkg/nexus/client.go > pkg/nexus/client.go.tmp && mv pkg/nexus/client.go.tmp pkg/nexus/client.go
fi

# Add package comment to service package
if ! grep -q "^// Package service" pkg/service/apply.go 2>/dev/null; then
    echo '// Package service provides business logic for applying Nexus configurations.' | cat - pkg/service/apply.go > pkg/service/apply.go.tmp && mv pkg/service/apply.go.tmp pkg/service/apply.go
fi

echo "=== Fixing errcheck issues in create.go ==="
# This will be done manually as it requires careful review

echo "=== Done! Please review changes before committing ==="
