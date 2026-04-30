# Nexus CLI User Guide

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

## Introduction

Nexus CLI is a command-line tool for automating the management of Nexus Repository Manager. It allows you to batch create and manage the following through YAML configuration files:

- User accounts
- Various types of repositories (Maven, Docker, NPM, etc.)
- Roles and permissions
- User-to-repository permission mappings

## Installation

### Option 1: Build from Source

```bash
# 1. Ensure Go 1.21 or higher is installed
go version

# 2. Clone the project
git clone https://github.com/alauda/nexus-cli.git
cd nexus-cli

# 3. Download dependencies
go mod download

# 4. Build
make build

# 5. Verify installation
./nexus-cli version
```

### Option 2: Direct Installation

```bash
go install github.com/alauda/nexus-cli@latest
```

## Configure Nexus Authentication

Nexus CLI obtains authentication information through environment variables and does not store admin passwords in configuration files.

```bash
# Linux/macOS
export NEXUS_URL=http://your-nexus-server:8081
export NEXUS_USERNAME=admin
export NEXUS_PASSWORD=your-admin-password
```

It's recommended to add these environment variables to your shell configuration file (such as `.bashrc`, `.zshrc`).

## Create Configuration File

### Basic Example

Create a file named `my-config.yaml`:

```yaml
# Create a developer user
users:
  - id: "dev-user"
    firstName: "Dev"
    lastName: "User"
    emailAddress: "dev@company.com"
    password: "DevPassword123"
    status: "active"
    roles:
      - "nx-deploy"

# Create a Maven repository
repositories:
  - name: "company-maven"
    format: "maven2"
    type: "hosted"
    online: true
    storage:
      blobStoreName: "default"
      strictContentTypeValidation: true
      writePolicy: "ALLOW_ONCE"
    maven:
      versionPolicy: "RELEASE"
      layoutPolicy: "STRICT"

# Assign repository permissions to user
userRepositoryPermissions:
  - userId: "dev-user"
    repository: "company-maven"
    privileges:
      - "READ"
      - "BROWSE"
      - "ADD"
```

### Complete Example

Refer to `config/example.yaml` for more configuration options.

### NameMode (Optional)

`nameMode` controls how resource identifiers are generated before resources are created:

- `name`: keep configured values unchanged (default behavior)
- `suffix`: append an auto-generated unique suffix (format: `YYYYMMDDHHMMSSmmm-rand4`)

You can set `nameMode` globally and override it per resource:

```yaml
nameMode: "suffix"

users:
  - id: "repo-admin"
    emailAddress: "repo-admin@example.com"
    roles:
      - "repository-manager"

repositories:
  - nameMode: "name" # override global suffix mode for this repository
    name: "maven-releases"
    format: "maven2"
    type: "hosted"
    online: true
    storage:
      blobStoreName: "default"
      strictContentTypeValidation: true
```

To reuse the same resolved names in later cleanup workflows, persist the resolved config:

```bash
nexus-cli create -c my-config.yaml --resolved-config resolved-config.yaml
nexus-cli delete -c resolved-config.yaml
```

## Apply Configuration

```bash
# Apply configuration file
nexus-cli apply -c my-config.yaml
```

Example output:

```
2024/01/01 10:00:00 Connecting to Nexus at http://localhost:8081...
2024/01/01 10:00:01 Successfully connected to Nexus
2024/01/01 10:00:01 Loaded configuration from my-config.yaml
2024/01/01 10:00:01 Starting to apply configuration...
2024/01/01 10:00:01 Applying privileges...
2024/01/01 10:00:02 Applying roles...
2024/01/01 10:00:02 Applying repositories...
2024/01/01 10:00:03 Created repository: company-maven (format: maven2, type: hosted)
2024/01/01 10:00:03 Applying users...
2024/01/01 10:00:04 Created user: dev-user
2024/01/01 10:00:04 Applying user repository permissions...
2024/01/01 10:00:05 Configuration applied successfully!
```

## Common Use Cases

### Scenario 1: Create Repositories and Users for a New Project

```yaml
users:
  - id: "project-a-dev"
    firstName: "Project A"
    lastName: "Developer"
    emailAddress: "project-a@company.com"
    password: "SecurePass123"
    status: "active"
    roles: []

repositories:
  - name: "project-a-maven"
    format: "maven2"
    type: "hosted"
    online: true
    storage:
      blobStoreName: "default"
      strictContentTypeValidation: true
      writePolicy: "ALLOW"
    maven:
      versionPolicy: "MIXED"
      layoutPolicy: "STRICT"

  - name: "project-a-docker"
    format: "docker"
    type: "hosted"
    online: true
    storage:
      blobStoreName: "default"
      strictContentTypeValidation: true
      writePolicy: "ALLOW"
    docker:
      httpPort: 8082
      forceBasicAuth: true
      v1Enabled: false

userRepositoryPermissions:
  - userId: "project-a-dev"
    repository: "project-a-maven"
    privileges:
      - "READ"
      - "BROWSE"
      - "ADD"
      - "EDIT"

  - userId: "project-a-dev"
    repository: "project-a-docker"
    privileges:
      - "READ"
      - "BROWSE"
      - "ADD"
```

### Scenario 2: Configure Proxy Repositories

```yaml
repositories:
  # Maven Central proxy
  - name: "maven-central-proxy"
    format: "maven2"
    type: "proxy"
    online: true
    storage:
      blobStoreName: "default"
      strictContentTypeValidation: true
    proxy:
      remoteUrl: "https://repo1.maven.org/maven2/"
      contentMaxAge: 1440
      metadataMaxAge: 1440
    maven:
      versionPolicy: "RELEASE"
      layoutPolicy: "STRICT"

  # NPM registry proxy
  - name: "npmjs-proxy"
    format: "npm"
    type: "proxy"
    online: true
    storage:
      blobStoreName: "default"
      strictContentTypeValidation: true
    proxy:
      remoteUrl: "https://registry.npmjs.org"
      contentMaxAge: 1440
      metadataMaxAge: 1440

  # Docker Hub proxy
  - name: "dockerhub-proxy"
    format: "docker"
    type: "proxy"
    online: true
    storage:
      blobStoreName: "default"
      strictContentTypeValidation: true
    proxy:
      remoteUrl: "https://registry-1.docker.io"
      contentMaxAge: 1440
      metadataMaxAge: 1440
    docker:
      httpPort: 8083
      forceBasicAuth: true
      v1Enabled: false
```

### Scenario 3: Create Roles and Permission System

```yaml
# Define privileges first
privileges:
  - name: "maven-read-only"
    description: "Read-only access to Maven repositories"
    type: "repository-content-selector"
    format: "maven2"
    repository: "*"
    actions:
      - "READ"
      - "BROWSE"

  - name: "maven-deploy"
    description: "Deploy access to Maven repositories"
    type: "repository-content-selector"
    format: "maven2"
    repository: "*"
    actions:
      - "READ"
      - "BROWSE"
      - "ADD"
      - "EDIT"

# Create roles
roles:
  - id: "developer-readonly"
    name: "Developer (Read-Only)"
    description: "Developers with read-only access"
    privileges:
      - "maven-read-only"

  - id: "developer-deploy"
    name: "Developer (Deploy)"
    description: "Developers with deploy permissions"
    privileges:
      - "maven-deploy"

# Assign to users
users:
  - id: "readonly-dev"
    firstName: "ReadOnly"
    lastName: "Developer"
    emailAddress: "readonly@company.com"
    password: "Pass123"
    status: "active"
    roles:
      - "developer-readonly"

  - id: "deploy-dev"
    firstName: "Deploy"
    lastName: "Developer"
    emailAddress: "deploy@company.com"
    password: "Pass123"
    status: "active"
    roles:
      - "developer-deploy"
```

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
