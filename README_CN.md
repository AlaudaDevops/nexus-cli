# Nexus CLI 使用指南

## 简介

Nexus CLI 是一个命令行工具，用于自动化管理 Nexus Repository Manager。它允许你通过 YAML 配置文件批量创建和管理：

- 用户账户
- 各种类型的仓库（Maven、Docker、NPM 等）
- 角色和权限
- 用户与仓库的权限映射

## 安装步骤

### 方式一：从源码构建

```bash
# 1. 确保安装了 Go 1.21 或更高版本
go version

# 2. 克隆项目
git clone https://github.com/alauda/nexus-cli.git
cd nexus-cli

# 3. 下载依赖
go mod download

# 4. 构建
make build

# 5. 验证安装
./nexus-cli version
```

### 方式二：直接安装

```bash
go install github.com/alauda/nexus-cli@latest
```

## 配置 Nexus 认证

Nexus CLI 通过环境变量获取认证信息，不在配置文件中存储管理员密码。

```bash
# Linux/macOS
export NEXUS_URL=http://your-nexus-server:8081
export NEXUS_USERNAME=admin
export NEXUS_PASSWORD=your-admin-password
```

建议将这些环境变量添加到你的 shell 配置文件中（如 `.bashrc`, `.zshrc`）。

## 创建配置文件

### 基础示例

创建一个名为 `my-config.yaml` 的文件：

```yaml
# 创建一个开发者用户
users:
  - id: "dev-user"
    firstName: "Dev"
    lastName: "User"
    emailAddress: "dev@company.com"
    password: "DevPassword123"
    status: "active"
    roles:
      - "nx-deploy"

# 创建一个 Maven 仓库
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

# 给用户分配仓库权限
userRepositoryPermissions:
  - userId: "dev-user"
    repository: "company-maven"
    privileges:
      - "READ"
      - "BROWSE"
      - "ADD"
```

### 完整示例

参考 `config/example.yaml` 获取更多配置选项。

## 应用配置

```bash
# 应用配置文件
nexus-cli apply -c my-config.yaml
```

输出示例：

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

## 常见使用场景

### 场景 1: 为新项目创建仓库和用户

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

### 场景 2: 配置代理仓库

```yaml
repositories:
  # Maven 中央仓库代理
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

  # NPM 仓库代理
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

  # Docker Hub 代理
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

### 场景 3: 创建角色和权限体系

```yaml
# 先定义权限
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

# 创建角色
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

# 分配给用户
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