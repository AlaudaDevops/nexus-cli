#!/bin/bash
# Comprehensive lint fix script

set -e

echo "=== Fixing all linting issues ==="

# Add package comment to main.go
if ! grep -q "^//" main.go; then
cat > main.go.tmp << 'EOF'
// Package main is the entry point for Nexus CLI application.
EOF
cat main.go >> main.go.tmp
mv main.go.tmp main.go
fi

# Add package comments to all packages
for pkg_dir in cmd pkg/config pkg/nexus pkg/service; do
    first_file=$(find "$pkg_dir" -name "*.go" -not -name "*_test.go" | head -1)
    if [ -f "$first_file" ] && ! grep -q "^// Package" "$first_file"; then
        pkg_name=$(basename "$pkg_dir")
        case "$pkg_name" in
            cmd) comment="// Package cmd provides command-line interface for Nexus CLI.";;
            config) comment="// Package config provides configuration types and loading functionality.";;
            nexus) comment="// Package nexus provides client for Sonatype Nexus Repository Manager API.";;
            service) comment="// Package service provides business logic for applying Nexus configurations.";;
            *) comment="// Package $pkg_name provides functionality for Nexus CLI.";;
        esac
        echo "$comment" | cat - "$first_file" > "$first_file.tmp" && mv "$first_file.tmp" "$first_file"
    fi
done

# Add variable comments
sed -i '' 's/^	Version   = "dev"$/	\/ Version is the current version of Nexus CLI.\n	Version   = "dev"/' cmd/version.go 2>/dev/null || true

# Fix unused parameters by renaming to _
sed -i '' 's/func runCreate(cmd \*cobra.Command, args \[\]string)/func runCreate(_ \*cobra.Command, args \[\]string)/' cmd/create.go
sed -i '' 's/func runDelete(cmd \*cobra.Command, args \[\]string)/func runDelete(_ \*cobra.Command, args \[\]string)/' cmd/delete.go
sed -i '' 's/Run: func(cmd \*cobra.Command, args \[\]string)/Run: func(_ \*cobra.Command, args \[\]string)/' cmd/version.go
sed -i '' 's/func showDeletePlan(cfg \*config.Config, formatter \*output.Formatter)/func showDeletePlan(cfg \*config.Config, _ \*output.Formatter)/' cmd/delete.go

echo "=== Running gofmt and goimports ==="
gofmt -w -s .
export PATH=$PATH:$(go env GOPATH)/bin
goimports -w .

echo "=== Done! Manual fixes still needed for complex issues ==="
echo "Please review the changes and fix remaining errcheck and gocyclo issues manually"
