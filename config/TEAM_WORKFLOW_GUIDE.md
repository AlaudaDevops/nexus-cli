# 团队工作流程使用指南

本指南说明如何使用三层权限模型来管理 Nexus 仓库：
1. **Nexus Admin** → 创建团队管理员
2. **团队管理员** → 创建团队仓库管理员
3. **团队仓库管理员** → 创建和管理团队仓库

## 权限层级

```
┌─────────────────────────────────────────────────┐
│           Nexus Admin (超级管理员)                │
│  - 完整的系统管理权限                              │
│  - 创建团队管理员                                 │
└────────────────┬────────────────────────────────┘
                 │
                 ├─→ 执行: team-admin.yaml
                 │
                 ▼
┌─────────────────────────────────────────────────┐
│        Team Admin (团队管理员)                    │
│  - 创建和管理用户                                 │
│  - 创建和管理角色                                 │
│  - 不能直接管理仓库                               │
└────────────────┬────────────────────────────────┘
                 │
                 ├─→ 执行: team-repo-manager.yaml
                 │
                 ▼
┌─────────────────────────────────────────────────┐
│    Team Repository Manager (团队仓库管理员)       │
│  - 创建、删除、修改仓库                           │
│  - 管理仓库配置                                   │
│  - 为团队成员分配仓库权限                          │
└────────────────┬────────────────────────────────┘
                 │
                 ├─→ 执行: team-repositories.yaml
                 │
                 ▼
┌─────────────────────────────────────────────────┐
│        Team Developer (团队开发者)                │
│  - 使用分配的仓库                                 │
│  - 上传/下载制品                                  │
│  - 不能管理仓库                                   │
└─────────────────────────────────────────────────┘
```

## 完整操作流程

### 步骤 1: Nexus Admin 创建团队管理员

**由谁执行**: Nexus 超级管理员

**目的**: 为每个团队创建一个团队管理员账号

**操作步骤**:

```bash
# 1. 使用 Nexus Admin 凭证
export NEXUS_URL=http://localhost:8081
export NEXUS_USERNAME=admin
export NEXUS_PASSWORD=admin123

# 2. 应用配置，创建团队管理员
nexus-cli apply -c config/team-admin.yaml
```

**创建内容**:
- ✅ `team-admin` 角色（具有用户和角色管理权限）
- ✅ `team1-admin` 用户（Team1 的管理员，密码: Team1Admin123）
- ✅ `team2-admin` 用户（Team2 的管理员，密码: Team2Admin123）

**完成后**:
- 通知 Team1 管理员：用户名 `team1-admin`，密码 `Team1Admin123`
- 通知 Team2 管理员：用户名 `team2-admin`，密码 `Team2Admin123`

---

### 步骤 2: 团队管理员创建团队仓库管理员

**由谁执行**: 团队管理员（如 team1-admin）

**目的**: 创建团队的仓库管理员和开发者账号

**操作步骤**:

```bash
# 1. 使用团队管理员凭证
export NEXUS_URL=http://localhost:8081
export NEXUS_USERNAME=team1-admin
export NEXUS_PASSWORD=Team1Admin123

# 2. 应用配置，创建团队仓库管理员
nexus-cli apply -c config/team-repo-manager.yaml
```

**创建内容**:
- ✅ `team1-repo-admin` 角色（具有仓库管理权限）
- ✅ `team1-repo-manager` 用户（Team1 仓库管理员，密码: Team1Repo123）
- ✅ `team1-developer` 角色（开发者角色）
- ✅ `team1-dev1` 用户（Team1 开发者，密码: Team1Dev123）

**完成后**:
- 通知仓库管理员：用户名 `team1-repo-manager`，密码 `Team1Repo123`
- 通知开发者：用户名 `team1-dev1`，密码 `Team1Dev123`

---

### 步骤 3: 团队仓库管理员创建和管理仓库

**由谁执行**: 团队仓库管理员（如 team1-repo-manager）

**目的**: 创建团队的仓库并为团队成员分配权限

**操作步骤**:

```bash
# 1. 使用团队仓库管理员凭证
export NEXUS_URL=http://localhost:8081
export NEXUS_USERNAME=team1-repo-manager
export NEXUS_PASSWORD=Team1Repo123

# 2. 应用配置，创建团队仓库
nexus-cli apply -c config/team-repositories.yaml
```

**创建内容**:
- ✅ `team1-maven-releases` - Maven 正式版本仓库
- ✅ `team1-maven-snapshots` - Maven 快照版本仓库
- ✅ `team1-pypi-hosted` - Python 私有包仓库
- ✅ `team1-go-proxy` - Go 模块代理仓库
- ✅ 为 `team1-dev1` 分配各仓库的访问权限

**完成后**:
- 团队成员可以开始使用仓库

---

## 配置文件说明

### 1. team-admin.yaml

**使用者**: Nexus Admin
**目的**: 创建团队管理员

```yaml
roles:
  - id: "team-admin"
    privileges:
      - "nx-users-all"      # 用户管理
      - "nx-roles-all"      # 角色管理
      - "nx-privileges-read" # 查看权限

users:
  - id: "team1-admin"
    roles:
      - "team-admin"
```

### 2. team-repo-manager.yaml

**使用者**: 团队管理员（如 team1-admin）
**目的**: 创建团队仓库管理员和开发者

```yaml
roles:
  - id: "team1-repo-admin"
    privileges:
      - "nx-repository-admin-*-*-*"  # 仓库管理
      - "nx-repositories-create"     # 创建仓库
      - "nx-repositories-delete"     # 删除仓库

users:
  - id: "team1-repo-manager"
    roles:
      - "team1-repo-admin"
```

### 3. team-repositories.yaml

**使用者**: 团队仓库管理员（如 team1-repo-manager）
**目的**: 创建和管理团队仓库

```yaml
repositories:
  - name: "team1-maven-releases"
    format: "maven2"
    type: "hosted"

userRepositoryPermissions:
  - userId: "team1-dev1"
    repository: "team1-maven-releases"
    privileges:
      - "READ"
      - "BROWSE"
      - "ADD"
```

---

## 实际使用示例

### 场景：公司有两个团队（Team1 和 Team2）

#### 初始化（Nexus Admin 执行一次）

```bash
# Admin 创建两个团队的管理员
export NEXUS_USERNAME=admin
export NEXUS_PASSWORD=admin123
nexus-cli apply -c config/team-admin.yaml
```

结果：
- team1-admin 创建成功
- team2-admin 创建成功

#### Team1 设置（team1-admin 执行）

```bash
# Team1 管理员创建仓库管理员和开发者
export NEXUS_USERNAME=team1-admin
export NEXUS_PASSWORD=Team1Admin123

# 可以修改配置文件，添加更多 Team1 的用户
nexus-cli apply -c config/team-repo-manager.yaml
```

#### Team1 创建仓库（team1-repo-manager 执行）

```bash
# Team1 仓库管理员创建仓库
export NEXUS_USERNAME=team1-repo-manager
export NEXUS_PASSWORD=Team1Repo123

# 可以修改配置文件，添加更多仓库
nexus-cli apply -c config/team-repositories.yaml
```

#### Team2 独立操作

Team2 使用相同流程，但使用自己的凭证和配置文件（需要相应修改 yaml 文件中的 team2）。

---

## 权限矩阵

| 操作 | Nexus Admin | Team Admin | Repo Manager | Developer |
|------|-------------|------------|--------------|-----------|
| 创建用户 | ✅ | ✅ | ❌ | ❌ |
| 删除用户 | ✅ | ✅ | ❌ | ❌ |
| 创建角色 | ✅ | ✅ | ❌ | ❌ |
| 创建仓库 | ✅ | ❌ | ✅ | ❌ |
| 删除仓库 | ✅ | ❌ | ✅ | ❌ |
| 修改仓库 | ✅ | ❌ | ✅ | ❌ |
| 上传制品 | ✅ | ❌ | ✅ | ✅ |
| 下载制品 | ✅ | ❌ | ✅ | ✅ |

---

## 自定义配置

### 为 Team2 创建配置

复制并修改配置文件：

```bash
# 1. 复制 Team1 的配置
cp config/team-repo-manager.yaml config/team2-repo-manager.yaml
cp config/team-repositories.yaml config/team2-repositories.yaml

# 2. 批量替换 team1 为 team2
sed -i 's/team1/team2/g' config/team2-repo-manager.yaml
sed -i 's/team1/team2/g' config/team2-repositories.yaml

# 3. Team2 管理员执行
export NEXUS_USERNAME=team2-admin
export NEXUS_PASSWORD=Team2Admin123
nexus-cli apply -c config/team2-repo-manager.yaml

# 4. Team2 仓库管理员执行
export NEXUS_USERNAME=team2-repo-manager
export NEXUS_PASSWORD=Team2Repo123
nexus-cli apply -c config/team2-repositories.yaml
```

### 添加更多开发者

编辑 `team-repo-manager.yaml`，添加新用户：

```yaml
users:
  - id: "team1-dev2"
    firstName: "Team1"
    lastName: "Developer2"
    emailAddress: "team1-dev2@example.com"
    password: "Team1Dev456"
    status: "active"
    roles:
      - "team1-developer"
```

然后重新应用：

```bash
export NEXUS_USERNAME=team1-admin
export NEXUS_PASSWORD=Team1Admin123
nexus-cli apply -c config/team-repo-manager.yaml
```

### 添加更多仓库

编辑 `team-repositories.yaml`，添加新仓库：

```yaml
repositories:
  - name: "team1-npm-hosted"
    format: "npm"
    type: "hosted"
    online: true
    storage:
      blobStoreName: "default"
      strictContentTypeValidation: true
      writePolicy: "ALLOW_ONCE"
```

然后重新应用：

```bash
export NEXUS_USERNAME=team1-repo-manager
export NEXUS_PASSWORD=Team1Repo123
nexus-cli apply -c config/team-repositories.yaml
```

---

## 安全建议

1. **密码管理**
   - 首次创建后立即修改密码
   - 使用强密码（至少 12 位，包含大小写、数字、特殊字符）
   - 定期轮换密码

2. **配置文件管理**
   - 不要将配置文件提交到 Git
   - 使用 `.gitignore` 忽略 `*.yaml`（保留示例文件）
   - 使用密钥管理系统存储敏感信息

3. **最小权限原则**
   - 只分配必要的权限
   - 定期审查用户权限
   - 及时删除离职员工账号

4. **审计**
   - 定期检查 Nexus 审计日志
   - 监控异常操作
   - 保留操作记录

---

## 故障排查

### 问题 1: 团队管理员无法创建仓库管理员

**现象**:
```
Error: failed to create role: request failed with status 403
```

**原因**: 团队管理员只有用户和角色管理权限，不能直接创建仓库相关角色

**解决**:
- 确认使用的是 `team-repo-manager.yaml`（不是 `team-repositories.yaml`）
- 仓库相关角色由仓库管理员创建，不是团队管理员

### 问题 2: 仓库管理员创建仓库失败

**现象**:
```
Error: failed to create repository: request failed with status 403
```

**原因**: 用户没有仓库创建权限

**解决**:
1. 确认使用正确的凭证（`team1-repo-manager`）
2. 确认角色已分配 `nx-repositories-create` 权限
3. 检查 `team-repo-manager.yaml` 是否正确应用

### 问题 3: 开发者无法访问仓库

**现象**: 开发者无法上传或下载制品

**解决**:
1. 检查 `team-repositories.yaml` 中的 `userRepositoryPermissions` 配置
2. 确认已为用户分配相应权限
3. 重新应用配置文件

---

## 下一步

1. **应用配置**: 按照上述三个步骤执行
2. **测试权限**: 使用各角色账号登录 Nexus Web UI 验证权限
3. **配置客户端**: 团队成员配置 Maven/Python/Go 客户端使用仓库
4. **文档培训**: 为团队成员提供使用培训

---

## 相关文档

- [README.md](../README.md) - 项目主文档
- [QUICKSTART.md](../QUICKSTART.md) - 快速开始
- [REPOSITORY_MANAGER_GUIDE.md](REPOSITORY_MANAGER_GUIDE.md) - 仓库管理指南
