# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Nexus CLI is a Go-based command-line tool for automating the management of Sonatype Nexus Repository Manager 3.x. It enables declarative configuration of users, repositories, roles, and permissions through YAML files.

## Build and Development Commands

```bash
# Build the binary
make build

# Run tests
make test

# Format code
make fmt

# Lint code
make vet

# Build for all platforms
make build-all

# Clean build artifacts
make clean
```

## Architecture

### Layered Design

The codebase follows a strict layered architecture with clear separation of concerns:

1. **cmd/** - CLI interface layer using Cobra
   - `root.go` - Root command with global flags
   - `apply.go` - Main command for applying YAML configurations
   - `version.go` - Version information command

2. **pkg/nexus/** - HTTP API client layer
   - `client.go` - Base HTTP client with Basic Auth
   - `user.go`, `repository.go`, `role.go`, `privilege.go` - Resource-specific API methods
   - Each file handles CRUD operations for one resource type

3. **pkg/service/** - Business logic layer
   - `apply.go` - Orchestrates the application of configurations
   - Implements idempotent operations (safe to re-run)
   - Handles the order of operations (privileges → roles → repositories → users → permissions)

4. **pkg/config/** - Configuration management layer
   - `types.go` - All YAML configuration structures
   - `loader.go` - YAML parsing and environment variable handling

5. **pkg/output/** - Output formatting layer
   - `formatter.go` - Multiple output formats (text, JSON, YAML, template)
   - `templates.go` - Predefined output templates

### Authentication Model

**Critical**: Nexus admin credentials are NEVER stored in YAML files. They must be provided via environment variables:

```bash
export NEXUS_URL=http://localhost:8081
export NEXUS_USERNAME=admin
export NEXUS_PASSWORD=admin123
```

YAML files only contain the resources to be created (users, repos, etc.), not authentication credentials.

### Three-Tier Permission Workflow

The tool supports a hierarchical permission model:

1. **Nexus Admin** (uses system admin creds) → Creates Team Admins
   - Config: `config/team-admin.yaml`
   - Creates roles with user/role management permissions

2. **Team Admin** (uses team admin creds) → Creates Repo Managers
   - Config: `config/team-repo-manager.yaml`
   - Creates roles with repository management permissions

3. **Repo Manager** (uses repo manager creds) → Creates Repositories
   - Config: `config/team-repositories.yaml`
   - Creates actual repositories and assigns permissions

Each tier uses different credentials via environment variables.

## Supported Repository Formats

- **Maven** (maven2): hosted, proxy, group
- **Docker**: hosted, proxy, group
- **NPM**: hosted, proxy, group
- **Python** (pypi): hosted, proxy, group
- **Go**: proxy, group only (no hosted type supported by Nexus)

## Key Design Patterns

### Idempotency

All apply operations are idempotent:
- Before creating, check if resource exists
- If exists, either skip or update (depending on resource type)
- Users and roles are updated; repositories are skipped
- Safe to run the same config multiple times

### Resource Application Order

The service layer applies resources in a specific order to satisfy dependencies:

1. Privileges (independent)
2. Roles (depend on privileges)
3. Repositories (independent)
4. Users (depend on roles)
5. User-Repository Permissions (depend on users and repositories)

This order is critical and should not be changed.

### Error Handling

- Use `fmt.Errorf` with `%w` to wrap errors and preserve error chains
- Return detailed context in error messages (e.g., "failed to create user %s: %w")
- The apply command stops on first error to prevent partial configurations

## Configuration Files

### Structure

All config files follow this structure:

```yaml
users: []           # User definitions
repositories: []    # Repository definitions
privileges: []      # Custom privileges
roles: []           # Custom roles
userRepositoryPermissions: []  # User-to-repo permission mappings
```

### Config File Locations

- `config/example.yaml` - Complete reference example
- `config/team-*.yaml` - Three-tier workflow examples
- `config/repository-manager.yaml` - Repository admin setup

### Output Templates

Templates (in `templates/`) define how created resources are reported:

```yaml
users: |
  {{- range . }}
  - userId: {{ .UserID }}
    email: {{ .EmailAddress }}
  {{- end }}
```

Templates are pure Go templates operating on resource lists.

## Testing

Tests use table-driven patterns:

```go
tests := []struct {
    name    string
    input   SomeType
    want    ExpectedType
    wantErr bool
}{
    // test cases
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // test logic
    })
}
```

When adding new Nexus API methods, add corresponding tests in `*_test.go` files.

## Adding New Repository Formats

To add support for a new repository format (e.g., Helm):

1. Add API methods in `pkg/nexus/repository.go`:
   ```go
   func (c *Client) CreateHelmHostedRepository(req RepositoryRequest) error
   func (c *Client) CreateHelmProxyRepository(req RepositoryRequest) error
   ```

2. Add case in `pkg/service/apply.go` `createRepository()`:
   ```go
   case "helm":
       switch repo.Type {
       case "hosted":
           return s.client.CreateHelmHostedRepository(req)
       // ...
       }
   ```

3. Update config types in `pkg/config/types.go` if format needs specific config

4. Update documentation in README.md

## Common Gotchas

1. **Go repository limitation**: Nexus Go repositories only support proxy/group, not hosted
2. **Write policies**: Only apply to hosted repositories, not proxy/group
3. **Role updates**: When updating roles, ALL privileges must be specified (not incremental)
4. **User password changes**: Require a separate API call to `ChangePassword()`
5. **Repository URLs**: Generated by Nexus, not configurable in request

## CI/CD Integration

GitHub Actions workflows in `.github/workflows/`:
- `ci.yml` - Runs on PR: lint, test, build
- `release.yml` - Runs on tag: multi-platform build and GitHub release

Version info is injected at build time via ldflags in Makefile.

## Output Template Feature

The CLI supports customizable output templates for displaying created resources after `apply` command execution.

### Implementation Details

**Location**: `cmd/apply.go`

**New Flags**:
- `--output-template <file>`: Path to template file (YAML format with Go text/template syntax)
- `--output-file <file>`: Optional file to write output (defaults to stdout)

**Template Structure** (see `templates/` directory):
```yaml
users: |
  {{- range . }}
  - userId: {{ .UserID }}
    ...
  {{- end }}

repositories: |
  {{- range . }}
  - name: {{ .Name }}
    ...
  {{- end }}

roles: |
  {{- range . }}
  - id: {{ .ID }}
    ...
  {{- end }}

privileges: |
  {{- range . }}
  - name: {{ .Name }}
    ...
  {{- end }}
```

**Data Structures**:
- Users: `[]*nexus.UserResponse` - from API
- Repositories: `[]RepositoryOutput` - custom struct (API returns `map[string]interface{}`)
- Roles: `[]*nexus.RoleResponse` - from API
- Privileges: `[]*nexus.PrivilegeResponse` - from API

**How It Works**:
1. After successful apply, if `--output-template` is specified
2. Load template file and parse YAML to get 4 section templates
3. Fetch each resource from Nexus API (only resources in config file)
4. Render each template section with corresponding data
5. Output to file or stdout

**Key Functions**:
- `outputResources()`: Main orchestrator
- `renderTemplate()`: Renders single template with data
- `getString()`, `getBool()`: Extract values from repository map

### Bug Fixes Applied (Latest Session)

**Problem**: Multiple compilation errors in `pkg/service/apply.go`

**Errors Fixed**:
1. Function signature mismatch: All `apply*()` functions changed from returning `error` to `(int, error)` for counting
2. Missing formatter: Replaced all `log.Printf()` calls with `s.formatter.Info/Success/Warning()`
3. Counting logic: Added proper resource counting in each apply function
4. Return values: Updated `Apply()` to collect counts properly in `ApplyResult`

**Files Modified**:
- `pkg/service/apply.go`: Complete rewrite of all apply functions (433 lines)
- `cmd/apply.go`: Added output template feature (324 lines)

**Template Files** (already existed):
- `templates/resource-list.yaml`: Standard format
- `templates/simple.yaml`: Names only
- `templates/detailed.yaml`: All fields
- `templates/README.md`: Comprehensive usage guide

### Usage Examples

```bash
# Apply and show resources with standard template
nexus-cli apply -c config.yaml --output-template templates/resource-list.yaml

# Apply and save resources to file
nexus-cli apply -c config.yaml \
  --output-template templates/simple.yaml \
  --output-file created-resources.yaml

# Apply only (no resource output)
nexus-cli apply -c config.yaml
```
