// Package cmd provides command-line interface for Nexus CLI.
package cmd

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"text/template"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/alauda/nexus-cli/pkg/config"
	"github.com/alauda/nexus-cli/pkg/nexus"
	"github.com/alauda/nexus-cli/pkg/output"
	"github.com/alauda/nexus-cli/pkg/service"
)

var (
	outputFormat   string
	outputTemplate string
	outputFile     string
	quiet          bool
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create resources from YAML configuration file",
	Long: `Create command creates or updates Nexus resources based on the YAML configuration file.
It will create users, repositories, roles, and privileges as defined in the config file.`,
	Example: `  # Create resources from file
  nexus-cli create -c config.yaml

  # With environment variables
  export NEXUS_URL=http://localhost:8081
  export NEXUS_USERNAME=admin
  export NEXUS_PASSWORD=admin123
  nexus-cli create -c config.yaml

  # Output in JSON format
  nexus-cli create -c config.yaml --output json

  # Output in YAML format
  nexus-cli create -c config.yaml --output yaml

  # Use template file to output resources
  nexus-cli create -c config.yaml --output-template templates/resource-list.yaml

  # Save output to file
  nexus-cli create -c config.yaml --output-template templates/simple.yaml --output-file result.yaml

  # Quiet mode (only show errors)
  nexus-cli create -c config.yaml --quiet`,
	RunE: runCreate,
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "Output format (text|json|yaml|template|table)")
	createCmd.Flags().StringVar(&outputTemplate, "output-template", "", "Template file to format resource output")
	createCmd.Flags().StringVar(&outputFile, "output-file", "", "File to write resource output (stdout if not specified)")
	createCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Quiet mode - only show errors")
}

func runCreate(_ *cobra.Command, _ []string) error {
	if cfgFile == "" {
		return fmt.Errorf("config file is required, use -c or --config flag")
	}

	// 创建格式化器
	formatter := output.NewFormatter(output.Format(outputFormat), os.Stdout)
	formatter.SetQuiet(quiet)

	// 设置自定义模板
	if outputFormat == "template" {
		if outputTemplate == "" {
			return fmt.Errorf("template string is required when using template format")
		}
		formatter.SetTemplate(outputTemplate)
	}

	startTime := time.Now()

	// 检查配置文件是否存在
	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s", cfgFile)
	}

	// 获取 Nexus 认证信息
	url, username, password, err := config.GetNexusCredentials()
	if err != nil {
		return fmt.Errorf("failed to get Nexus credentials: %w", err)
	}

	formatter.Info(fmt.Sprintf("Connecting to Nexus at %s...", url))

	// 创建 Nexus 客户端
	client := nexus.NewClient(url, username, password)

	// 检查连接
	if err := client.CheckConnection(); err != nil {
		return fmt.Errorf("failed to connect to Nexus: %w", err)
	}

	formatter.Success("Successfully connected to Nexus")

	// 加载配置文件
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	formatter.Info(fmt.Sprintf("Loaded configuration from %s", cfgFile))

	// 创建服务并执行
	svc := service.NewApplyService(client, cfg, formatter)
	result, err := svc.Apply()
	if err != nil {
		return fmt.Errorf("failed to create resources: %w", err)
	}

	// 计算执行时间
	duration := time.Since(startTime)

	// 输出总结
	summary := &output.Summary{
		Total:     result.Total,
		Success:   result.Success,
		Failed:    result.Failed,
		Skipped:   result.Skipped,
		Errors:    result.Errors,
		Warnings:  result.Warnings,
		StartTime: startTime.Format(time.RFC3339),
		EndTime:   time.Now().Format(time.RFC3339),
		Duration:  duration.String(),
	}

	if err := formatter.PrintSummary(summary); err != nil {
		return err
	}

	// 如果指定了输出文件或输出模板，则输出资源列表
	if outputTemplate != "" || outputFile != "" {
		// 如果没有指定模板，使用默认的 YAML 格式输出（与配置文件格式一致）
		if outputTemplate == "" {
			return outputResourcesDefault(client, cfg, outputFile)
		}
		return outputResources(client, cfg, outputTemplate, outputFile)
	}

	return nil
}

// RepositoryOutput 仓库输出结构
type RepositoryOutput struct {
	Name                        string `json:"name"`
	Format                      string `json:"format"`
	Type                        string `json:"type"`
	Online                      bool   `json:"online"`
	URL                         string `json:"url"`
	BlobStoreName               string `json:"blobStoreName"`
	WritePolicy                 string `json:"writePolicy,omitempty"`
	StrictContentTypeValidation bool   `json:"strictContentTypeValidation"`
}

// UserWithPassword 带密码的用户信息（用于模板输出）
type UserWithPassword struct {
	*nexus.UserResponse
	Password string
}

// OutputConfig 整体模板输出配置（类似 gitlab-cli 格式）
type OutputConfig struct {
	Endpoint     string
	Host         string
	Port         int
	Scheme       string
	Users        []*UserWithPassword
	Repositories []RepositoryOutput
	Roles        []*nexus.RoleResponse
	Privileges   []*nexus.PrivilegeResponse
}

// outputResourcesDefault 使用默认 YAML 格式输出资源（与配置文件格式一致）
func outputResourcesDefault(client *nexus.Client, cfg *config.Config, outputFile string) error {
	// 创建默认输出结构（与 config.Config 格式一致）
	defaultOutput := struct {
		Users        []*UserWithPassword               `yaml:"users,omitempty"`
		Repositories []RepositoryOutput                `yaml:"repositories,omitempty"`
		Roles        []*nexus.RoleResponse             `yaml:"roles,omitempty"`
		Privileges   []*nexus.PrivilegeResponse        `yaml:"privileges,omitempty"`
		Permissions  []config.UserRepositoryPermission `yaml:"userRepositoryPermissions,omitempty"`
	}{}

	// 创建用户ID到密码的映射
	userPasswordMap := make(map[string]string)
	for _, u := range cfg.Users {
		userPasswordMap[u.ID] = u.Password
	}

	// 获取用户列表
	if len(cfg.Users) > 0 {
		for _, u := range cfg.Users {
			user, err := client.GetUser(u.ID)
			if err != nil {
				continue // 跳过获取失败的用户
			}
			userWithPassword := &UserWithPassword{
				UserResponse: user,
				Password:     userPasswordMap[u.ID],
			}
			defaultOutput.Users = append(defaultOutput.Users, userWithPassword)
		}
	}

	// 获取仓库列表
	if len(cfg.Repositories) > 0 {
		for _, r := range cfg.Repositories {
			repo, err := client.GetRepository(r.Name)
			if err != nil {
				continue
			}
			repoOut := RepositoryOutput{
				Name:   getString(repo, "name"),
				Format: getString(repo, "format"),
				Type:   getString(repo, "type"),
				Online: getBool(repo, "online"),
				URL:    getString(repo, "url"),
			}
			if storage, ok := repo["storage"].(map[string]interface{}); ok {
				repoOut.BlobStoreName = getString(storage, "blobStoreName")
				repoOut.StrictContentTypeValidation = getBool(storage, "strictContentTypeValidation")
				repoOut.WritePolicy = getString(storage, "writePolicy")
			}
			defaultOutput.Repositories = append(defaultOutput.Repositories, repoOut)
		}
	}

	// 获取角色列表
	if len(cfg.Roles) > 0 {
		for _, r := range cfg.Roles {
			role, err := client.GetRole(r.ID)
			if err != nil {
				continue
			}
			defaultOutput.Roles = append(defaultOutput.Roles, role)
		}
	}

	// 获取权限列表
	if len(cfg.Privileges) > 0 {
		for _, p := range cfg.Privileges {
			priv, err := client.GetPrivilege(p.Name)
			if err != nil {
				continue
			}
			defaultOutput.Privileges = append(defaultOutput.Privileges, priv)
		}
	}

	// 添加用户仓库权限映射（如果配置中有）
	if len(cfg.UserRepositoryPermissions) > 0 {
		defaultOutput.Permissions = cfg.UserRepositoryPermissions
	}

	// 序列化为 YAML
	yamlData, err := yaml.Marshal(defaultOutput)
	if err != nil {
		return fmt.Errorf("failed to marshal resources to YAML: %w", err)
	}

	// 输出结果
	var writer io.Writer
	if outputFile != "" {
		f, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer f.Close()
		writer = f
	} else {
		writer = os.Stdout
		fmt.Fprintln(writer, "\n===== Resource Output =====")
	}

	if _, err := writer.Write(yamlData); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	if outputFile != "" {
		fmt.Printf("✓ Resources written to %s\n", outputFile)
	}

	return nil
}

// outputResources 输出资源列表
func outputResources(client *nexus.Client, cfg *config.Config, templateFile, outputFile string) error {
	// 加载模板文件
	templateData, err := os.ReadFile(templateFile)
	if err != nil {
		return fmt.Errorf("failed to read template file: %w", err)
	}

	// 尝试解析为分段模板格式（旧格式）
	var templates struct {
		Users        string `yaml:"users"`
		Repositories string `yaml:"repositories"`
		Roles        string `yaml:"roles"`
		Privileges   string `yaml:"privileges"`
	}

	// 如果能解析为 YAML 且包含已知的段，则使用旧格式
	isLegacyFormat := false
	if err := yaml.Unmarshal(templateData, &templates); err == nil {
		if templates.Users != "" || templates.Repositories != "" || templates.Roles != "" || templates.Privileges != "" {
			isLegacyFormat = true
		}
	}

	// 如果是旧格式，使用原来的逻辑
	if isLegacyFormat {
		legacyTemplates := struct {
			Users        string
			Repositories string
			Roles        string
			Privileges   string
		}{
			Users:        templates.Users,
			Repositories: templates.Repositories,
			Roles:        templates.Roles,
			Privileges:   templates.Privileges,
		}
		return outputResourcesLegacy(client, cfg, legacyTemplates, outputFile)
	}

	// 否则使用新的整体模板格式
	return outputResourcesToolchain(client, cfg, string(templateData), outputFile)
}

// outputResourcesLegacy 使用旧的分段模板格式输出资源
func outputResourcesLegacy(client *nexus.Client, cfg *config.Config, templates struct {
	Users        string
	Repositories string
	Roles        string
	Privileges   string
}, outputFile string) error {
	// 获取所有资源
	resources := make(map[string]string)

	// 获取用户列表
	if templates.Users != "" && len(cfg.Users) > 0 {
		var users []*nexus.UserResponse
		for _, u := range cfg.Users {
			user, err := client.GetUser(u.ID)
			if err != nil {
				continue // 跳过获取失败的用户
			}
			users = append(users, user)
		}
		if rendered, err := renderTemplate("users", templates.Users, users); err == nil {
			resources["users"] = rendered
		}
	}

	// 获取仓库列表
	if templates.Repositories != "" && len(cfg.Repositories) > 0 {
		var repos []RepositoryOutput
		for _, r := range cfg.Repositories {
			repo, err := client.GetRepository(r.Name)
			if err != nil {
				continue
			}
			// 转换 map 到结构体
			repoOut := RepositoryOutput{
				Name:   getString(repo, "name"),
				Format: getString(repo, "format"),
				Type:   getString(repo, "type"),
				Online: getBool(repo, "online"),
				URL:    getString(repo, "url"),
			}
			if storage, ok := repo["storage"].(map[string]interface{}); ok {
				repoOut.BlobStoreName = getString(storage, "blobStoreName")
				repoOut.StrictContentTypeValidation = getBool(storage, "strictContentTypeValidation")
				repoOut.WritePolicy = getString(storage, "writePolicy")
			}
			repos = append(repos, repoOut)
		}
		if rendered, err := renderTemplate("repositories", templates.Repositories, repos); err == nil {
			resources["repositories"] = rendered
		}
	}

	// 获取角色列表
	if templates.Roles != "" && len(cfg.Roles) > 0 {
		var roles []*nexus.RoleResponse
		for _, r := range cfg.Roles {
			role, err := client.GetRole(r.ID)
			if err != nil {
				continue
			}
			roles = append(roles, role)
		}
		if rendered, err := renderTemplate("roles", templates.Roles, roles); err == nil {
			resources["roles"] = rendered
		}
	}

	// 获取权限列表
	if templates.Privileges != "" && len(cfg.Privileges) > 0 {
		var privileges []*nexus.PrivilegeResponse
		for _, p := range cfg.Privileges {
			priv, err := client.GetPrivilege(p.Name)
			if err != nil {
				continue
			}
			privileges = append(privileges, priv)
		}
		if rendered, err := renderTemplate("privileges", templates.Privileges, privileges); err == nil {
			resources["privileges"] = rendered
		}
	}

	// 输出结果
	var writer io.Writer
	if outputFile != "" {
		f, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer f.Close()
		writer = f
	} else {
		writer = os.Stdout
		fmt.Fprintln(writer, "\n===== Resource Output =====")
	}

	// 按顺序输出各部分，模板本身包含 section 名称
	for _, section := range []string{"users", "repositories", "roles", "privileges"} {
		if content, ok := resources[section]; ok && content != "" {
			fmt.Fprintf(writer, "%s:%s", section, content)
			if content[len(content)-1] != '\n' {
				fmt.Fprintln(writer)
			}
		}
	}

	if outputFile != "" {
		fmt.Printf("✓ Resources written to %s\n", outputFile)
	}

	return nil
}

// outputResourcesToolchain 使用整体模板格式输出资源（类似 gitlab-cli）
func outputResourcesToolchain(client *nexus.Client, cfg *config.Config, templateContent, outputFile string) error {
	// 解析 NEXUS_URL 环境变量以获取 endpoint, host, port, scheme
	nexusURL := os.Getenv("NEXUS_URL")
	if nexusURL == "" {
		return fmt.Errorf("NEXUS_URL environment variable is not set")
	}

	parsedURL, err := url.Parse(nexusURL)
	if err != nil {
		return fmt.Errorf("failed to parse NEXUS_URL: %w", err)
	}

	// 提取 host 和 port
	host := parsedURL.Hostname()
	port := 443 // 默认端口
	if parsedURL.Port() != "" {
		port, _ = strconv.Atoi(parsedURL.Port())
	} else if parsedURL.Scheme == "http" {
		port = 80
	}

	// 准备输出配置
	outputCfg := OutputConfig{
		Endpoint:     nexusURL,
		Host:         host,
		Port:         port,
		Scheme:       parsedURL.Scheme,
		Users:        []*UserWithPassword{},
		Repositories: []RepositoryOutput{},
		Roles:        []*nexus.RoleResponse{},
		Privileges:   []*nexus.PrivilegeResponse{},
	}

	// 创建用户ID到密码的映射
	userPasswordMap := make(map[string]string)
	for _, u := range cfg.Users {
		userPasswordMap[u.ID] = u.Password
	}

	// 获取用户列表
	if len(cfg.Users) > 0 {
		for _, u := range cfg.Users {
			user, err := client.GetUser(u.ID)
			if err != nil {
				continue // 跳过获取失败的用户
			}
			// 创建带密码的用户对象
			userWithPassword := &UserWithPassword{
				UserResponse: user,
				Password:     userPasswordMap[u.ID],
			}
			outputCfg.Users = append(outputCfg.Users, userWithPassword)
		}
	}

	// 获取仓库列表
	if len(cfg.Repositories) > 0 {
		for _, r := range cfg.Repositories {
			repo, err := client.GetRepository(r.Name)
			if err != nil {
				continue
			}
			// 转换 map 到结构体
			repoOut := RepositoryOutput{
				Name:   getString(repo, "name"),
				Format: getString(repo, "format"),
				Type:   getString(repo, "type"),
				Online: getBool(repo, "online"),
				URL:    getString(repo, "url"),
			}
			if storage, ok := repo["storage"].(map[string]interface{}); ok {
				repoOut.BlobStoreName = getString(storage, "blobStoreName")
				repoOut.StrictContentTypeValidation = getBool(storage, "strictContentTypeValidation")
				repoOut.WritePolicy = getString(storage, "writePolicy")
			}
			outputCfg.Repositories = append(outputCfg.Repositories, repoOut)
		}
	}

	// 获取角色列表
	if len(cfg.Roles) > 0 {
		for _, r := range cfg.Roles {
			role, err := client.GetRole(r.ID)
			if err != nil {
				continue
			}
			outputCfg.Roles = append(outputCfg.Roles, role)
		}
	}

	// 获取权限列表
	if len(cfg.Privileges) > 0 {
		for _, p := range cfg.Privileges {
			priv, err := client.GetPrivilege(p.Name)
			if err != nil {
				continue
			}
			outputCfg.Privileges = append(outputCfg.Privileges, priv)
		}
	}

	// 渲染模板
	tmpl, err := template.New("toolchain").Parse(templateContent)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// 输出结果
	var writer io.Writer
	if outputFile != "" {
		f, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer f.Close()
		writer = f
	} else {
		writer = os.Stdout
		fmt.Fprintln(writer, "\n===== Resource Output =====")
	}

	if err := tmpl.Execute(writer, outputCfg); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	if outputFile != "" {
		fmt.Printf("✓ Resources written to %s\n", outputFile)
	}

	return nil
}

// renderTemplate 渲染单个模板
func renderTemplate(name, tmplStr string, data interface{}) (string, error) {
	tmpl, err := template.New(name).Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse %s template: %w", name, err)
	}

	var buf []byte
	w := &writer{buf: &buf}
	if err := tmpl.Execute(w, data); err != nil {
		return "", fmt.Errorf("failed to execute %s template: %w", name, err)
	}

	return string(buf), nil
}

// writer 实现 io.Writer 接口，用于收集模板输出
type writer struct {
	buf *[]byte
}

func (w *writer) Write(p []byte) (n int, err error) {
	*w.buf = append(*w.buf, p...)
	return len(p), nil
}

// getString 从 map 中获取字符串值
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// getBool 从 map 中获取布尔值
func getBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}
