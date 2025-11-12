## 如何定制自己的模板

### 模板格式选择

Nexus CLI 支持两种模板格式，选择哪种取决于你的需求:

| 格式类型 | 适用场景 | 特点 |
|---------|---------|------|
| **分段模板** | 输出 YAML/TXT 列表格式 | 每个资源类型独立渲染 |
| **整体模板** | 输出配置文件、脚本、文档 | 可访问全局变量，适合复杂格式 |

---

### 方式 1: 分段模板格式（推荐新手）

**使用场景**: 当你需要输出标准的 YAML/JSON 格式资源列表时

#### 步骤 1: 创建模板文件

创建 `my-template.yaml`:

```yaml
# 用户列表模板
users: |
  {{- range . }}
  - userId: {{ .UserID }}
    email: {{ .EmailAddress }}
  {{- end }}

# 仓库列表模板
repositories: |
  {{- range . }}
  - name: {{ .Name }}
    type: {{ .Type }}
  {{- end }}

# 角色列表模板
roles: |
  {{- range . }}
  - id: {{ .ID }}
  {{- end }}

# 权限列表模板
privileges: |
  {{- range . }}
  - name: {{ .Name }}
  {{- end }}
```

**说明**:
- 每个 section (users/repositories/roles/privileges) 都是独立的模板
- `{{- range . }}` 表示遍历当前资源列表
- `{{ .字段名 }}` 输出字段值
- 如果不需要某个 section，可以留空或删除

#### 步骤 2: 使用模板

```bash
nexus-cli create -c config.yaml \
  --output-template my-template.yaml \
  --output-file result.txt
```

#### 示例 1: 只输出用户名和邮箱

创建 `users-email.yaml`:

```yaml
users: |
  {{- range . }}
  {{ .UserID }}: {{ .EmailAddress }}
  {{- end }}

repositories: ""
roles: ""
privileges: ""
```

输出:
```
developer: john@example.com
repo-manager: manager@example.com
```

#### 示例 2: 表格格式输出

创建 `table-format.yaml`:

```yaml
users: |
  用户ID          邮箱                     角色
  -------------------------------------------------------
  {{- range . }}
  {{ .UserID | printf "%-15s" }} {{ .EmailAddress | printf "%-25s" }} {{ range .Roles }}{{ . }} {{ end }}
  {{- end }}

repositories: |
  仓库名          格式      类型        URL
  -------------------------------------------------------
  {{- range . }}
  {{ .Name | printf "%-15s" }} {{ .Format | printf "%-10s" }} {{ .Type | printf "%-10s" }} {{ .URL }}
  {{- end }}

roles: ""
privileges: ""
```

---

### 方式 2: 整体模板格式（适合高级用户）

**使用场景**: 当你需要生成配置文件、脚本或自定义格式时

#### 可用的全局变量

```yaml
$.Endpoint       # Nexus URL (如: https://nexus.example.com)
$.Host           # 主机名 (如: nexus.example.com)
$.Port           # 端口号 (如: 443)
$.Scheme         # 协议 (如: https)
$.Users          # 用户列表数组
$.Repositories   # 仓库列表数组
$.Roles          # 角色列表数组
$.Privileges     # 权限列表数组
```

#### 示例 3: 生成环境变量文件

创建 `env-vars.yaml`:

```yaml
# Nexus Environment Variables
NEXUS_URL={{ $.Endpoint }}
NEXUS_HOST={{ $.Host }}
NEXUS_PORT={{ $.Port }}

{{- range $i, $user := $.Users }}
# User {{ $i }}
USER_{{ $i }}_ID={{ $user.UserID }}
USER_{{ $i }}_EMAIL={{ $user.EmailAddress }}
USER_{{ $i }}_PASSWORD={{ $user.Password }}
{{- end }}

{{- range $i, $repo := $.Repositories }}
# Repository {{ $i }}
REPO_{{ $i }}_NAME={{ $repo.Name }}
REPO_{{ $i }}_URL={{ $repo.URL }}
{{- end }}
```

使用:
```bash
nexus-cli create -c config.yaml \
  --output-template env-vars.yaml \
  --output-file .env
```

输出:
```bash
# Nexus Environment Variables
NEXUS_URL=https://nexus.example.com
NEXUS_HOST=nexus.example.com
NEXUS_PORT=443

# User 0
USER_0_ID=developer
USER_0_EMAIL=john@example.com
USER_0_PASSWORD=SecurePass123
```

---

### 常用 Go 模板语法速查

#### 1. 循环遍历

```yaml
# 简单循环
{{- range .Users }}
- {{ .UserID }}
{{- end }}

# 带索引的循环
{{- range $index, $user := .Users }}
{{ $index }}. {{ $user.UserID }}
{{- end }}

# 条件循环
{{- range .Repositories }}
  {{- if eq .Format "docker" }}
  - {{ .Name }}
  {{- end }}
{{- end }}
```

#### 2. 条件判断

```yaml
# if 判断
{{- if .WritePolicy }}
writePolicy: {{ .WritePolicy }}
{{- end }}

# if-else
{{- if .Online }}
status: online
{{- else }}
status: offline
{{- end }}

# 多条件
{{- if and (eq .Format "maven2") (eq .Type "hosted") }}
This is a Maven hosted repository
{{- end }}
```

#### 3. 比较运算符

```yaml
eq    # 等于:    {{ if eq .Type "hosted" }}
ne    # 不等于:  {{ if ne .Status "active" }}
lt    # 小于:    {{ if lt .Port 1000 }}
le    # 小于等于: {{ if le .Port 1024 }}
gt    # 大于:    {{ if gt .Port 8000 }}
ge    # 大于等于: {{ if ge .Port 8080 }}
```

#### 4. 字符串操作

```yaml
# 转大写 (需要自定义函数，标准模板不支持)
{{ .UserID }}

# 格式化
{{ printf "%-20s" .Name }}      # 左对齐，宽度20
{{ printf "%20s" .Name }}       # 右对齐，宽度20
{{ printf "%.2f" .Value }}      # 保留2位小数
```

#### 5. 访问嵌套数据

```yaml
# 访问全局变量（整体模板格式）
{{ $.Endpoint }}
{{ $.Host }}

# 访问当前元素
{{ .UserID }}
{{ .EmailAddress }}

# 访问数组
{{ range .Roles }}
  - {{ . }}
{{ end }}
```

#### 6. 长度和计数

```yaml
# 获取数组长度
{{ len .Users }}
{{ len $.Repositories }}

# 判断是否为空
{{- if .Roles }}
Has roles
{{- else }}
No roles
{{- end }}
```

#### 7. 注释

```yaml
{{/* 这是注释，不会出现在输出中 */}}

{{- /*
  多行注释
  也不会出现在输出中
*/ -}}
```

#### 8. 空格控制

```yaml
{{- .Name }}     # 去除左侧空格/换行
{{ .Name -}}     # 去除右侧空格/换行
{{- .Name -}}    # 去除两侧空格/换行
```

### 可用字段

#### 用户 (Users)
```go
.UserID           // 用户 ID
.FirstName        // 名
.LastName         // 姓
.EmailAddress     // 邮箱
.Status           // 状态 (active/disabled)
.Roles            // 角色列表 []string
.Source           // 来源 (default/ldap/...)
.Password         // 密码（仅在 UserWithPassword 中）
```

#### 仓库 (Repositories)
```go
.Name                           // 仓库名称
.Format                         // 格式 (maven2/docker/npm/pypi/go)
.Type                           // 类型 (hosted/proxy/group)
.Online                         // 是否在线 bool
.URL                            // 仓库 URL
.BlobStoreName                  // Blob 存储名称
.WritePolicy                    // 写策略 (hosted only)
.StrictContentTypeValidation    // 严格内容类型验证 bool
```

#### 角色 (Roles)
```go
.ID            // 角色 ID
.Name          // 角色名称
.Description   // 描述
.Privileges    // 权限列表 []string
.Roles         // 继承的角色 []string
.Source        // 来源
.ReadOnly      // 是否只读 bool
```

#### 权限 (Privileges)
```go
.Name          // 权限名称
.Description   // 描述
.Type          // 类型
.Format        // 格式
.Repository    // 仓库
.Actions       // 操作列表 []string
.ReadOnly      // 是否只读 bool
```

### Go 模板语法

```yaml
# 循环
{{- range .Users }}
  - {{ .UserID }}
{{- end }}

# 条件判断
{{- if .WritePolicy }}
writePolicy: {{ .WritePolicy }}
{{- end }}

# 访问全局变量（整体模板格式）
endpoint: {{ $.Endpoint }}

# 索引和计数
{{- range $index, $user := .Users }}
  {{- if $index }},{{ end }}
  "{{ $user.UserID }}"
{{- end }}
```

---

## 实战演练: 从零开始创建自定义模板

### 场景: 为团队创建简洁的资源报告

**需求**: 输出一个简单的文本报告，列出所有用户和仓库信息，格式清晰易读。

#### 第 1 步: 确定输出格式

我们想要这样的输出:
```
=== Nexus Resources Report ===
Endpoint: https://nexus.example.com

Users (2):
  1. developer (john@example.com) - Roles: developer
  2. repo-manager (manager@example.com) - Roles: repo-admin

Repositories (3):
  1. maven-releases (maven2/hosted) - https://nexus.example.com/repository/maven-releases/
  2. docker-hosted (docker/hosted) - https://nexus.example.com/repository/docker-hosted/
  3. npm-group (npm/group) - https://nexus.example.com/repository/npm-group/
```

#### 第 2 步: 选择模板格式

因为我们需要访问全局变量 `$.Endpoint`，所以选择**整体模板格式**。

#### 第 3 步: 编写模板

创建 `team-report.yaml`:

```yaml
=== Nexus Resources Report ===
Endpoint: {{ $.Endpoint }}

Users ({{ len $.Users }}):
{{- range $i, $user := $.Users }}
  {{ add $i 1 }}. {{ $user.UserID }} ({{ $user.EmailAddress }}) - Roles: {{ range $user.Roles }}{{ . }} {{ end }}
{{- end }}

Repositories ({{ len $.Repositories }}):
{{- range $i, $repo := $.Repositories }}
  {{ add $i 1 }}. {{ $repo.Name }} ({{ $repo.Format }}/{{ $repo.Type }}) - {{ $repo.URL }}
{{- end }}
```

**注意**: `{{ add $i 1 }}` 是因为索引从 0 开始，我们想显示 1, 2, 3...

#### 第 4 步: 测试模板

```bash
nexus-cli create -c config.yaml \
  --output-template team-report.yaml \
  --output-file report.txt
```

#### 第 5 步: 优化模板（处理空数据）

如果可能没有仓库，添加条件判断:

```yaml
=== Nexus Resources Report ===
Endpoint: {{ $.Endpoint }}

Users ({{ len $.Users }}):
{{- if $.Users }}
{{- range $i, $user := $.Users }}
  {{ add $i 1 }}. {{ $user.UserID }} ({{ $user.EmailAddress }}) - Roles: {{ range $user.Roles }}{{ . }} {{ end }}
{{- end }}
{{- else }}
  No users configured.
{{- end }}

Repositories ({{ len $.Repositories }}):
{{- if $.Repositories }}
{{- range $i, $repo := $.Repositories }}
  {{ add $i 1 }}. {{ $repo.Name }} ({{ $repo.Format }}/{{ $repo.Type }}) - {{ $repo.URL }}
{{- end }}
{{- else }}
  No repositories configured.
{{- end }}
```