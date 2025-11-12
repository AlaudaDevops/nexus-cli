package output

// 预定义模板

const (
	// TemplateUserList 用户列表模板
	TemplateUserList = `{{range .}}User: {{.UserID}}
  Name: {{.FirstName}} {{.LastName}}
  Email: {{.EmailAddress}}
  Status: {{.Status}}
  Roles: {{range .Roles}}{{.}} {{end}}
{{end}}`

	// TemplateRepositoryList 仓库列表模板
	TemplateRepositoryList = `{{range .}}Repository: {{.name}}
  Format: {{.format}}
  Type: {{.type}}
  URL: {{.url}}
  Online: {{.online}}
{{end}}`

	// TemplateApplySummary 应用配置总结模板
	TemplateApplySummary = `
╔═══════════════════════════════════════╗
║       Configuration Applied           ║
╚═══════════════════════════════════════╝

Users Created:        {{.UsersCreated}}
Repositories Created: {{.RepositoriesCreated}}
Roles Created:        {{.RolesCreated}}
Privileges Created:   {{.PrivilegesCreated}}

{{if .Errors}}Errors:
{{range .Errors}}  ✗ {{.}}
{{end}}{{end}}
{{if .Warnings}}Warnings:
{{range .Warnings}}  ⚠ {{.}}
{{end}}{{end}}
Total Time: {{.Duration}}
`

	// TemplateProgress 进度模板
	TemplateProgress = `[{{.Current}}/{{.Total}}] {{.Message}}`

	// TemplateSimple 简单模板
	TemplateSimple = `{{range .}}{{.}}
{{end}}`
)

// GetTemplate 获取预定义模板
func GetTemplate(name string) string {
	templates := map[string]string{
		"user-list":       TemplateUserList,
		"repository-list": TemplateRepositoryList,
		"apply-summary":   TemplateApplySummary,
		"progress":        TemplateProgress,
		"simple":          TemplateSimple,
	}

	if tmpl, ok := templates[name]; ok {
		return tmpl
	}
	return ""
}
