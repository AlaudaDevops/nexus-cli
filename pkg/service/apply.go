// Package service provides business logic for applying Nexus configurations.
package service

import (
	"fmt"
	"strings"

	"github.com/alauda/nexus-cli/pkg/config"
	"github.com/alauda/nexus-cli/pkg/nexus"
	"github.com/alauda/nexus-cli/pkg/output"
)

// ApplyService 应用服务
type ApplyService struct {
	client    *nexus.Client
	config    *config.Config
	formatter *output.Formatter
}

// ApplyResult 应用结果
type ApplyResult struct {
	Total               int
	Success             int
	Failed              int
	Skipped             int
	UsersCreated        int
	RepositoriesCreated int
	RolesCreated        int
	PrivilegesCreated   int
	Errors              []string
	Warnings            []string
}

// NewApplyService 创建应用服务
func NewApplyService(client *nexus.Client, cfg *config.Config, formatter *output.Formatter) *ApplyService {
	if formatter == nil {
		formatter = output.NewFormatter(output.FormatText, nil)
	}
	return &ApplyService{
		client:    client,
		config:    cfg,
		formatter: formatter,
	}
}

// Apply 应用配置
func (s *ApplyService) Apply() (*ApplyResult, error) {
	result := &ApplyResult{
		Errors:   []string{},
		Warnings: []string{},
	}

	s.formatter.Info("Starting to apply configuration...")

	// 1. 创建权限
	count, err := s.applyPrivileges()
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
		return result, fmt.Errorf("failed to apply privileges: %w", err)
	}
	result.PrivilegesCreated = count
	result.Success += count

	// 2. 创建角色
	count, err = s.applyRoles()
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
		return result, fmt.Errorf("failed to apply roles: %w", err)
	}
	result.RolesCreated = count
	result.Success += count

	// 3. 创建仓库
	count, err = s.applyRepositories()
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
		return result, fmt.Errorf("failed to apply repositories: %w", err)
	}
	result.RepositoriesCreated = count
	result.Success += count

	// 4. 创建用户
	count, err = s.applyUsers()
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
		return result, fmt.Errorf("failed to apply users: %w", err)
	}
	result.UsersCreated = count
	result.Success += count

	// 5. 配置用户仓库权限
	count, err = s.applyUserRepositoryPermissions()
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
		return result, fmt.Errorf("failed to apply user repository permissions: %w", err)
	}
	result.Success += count

	result.Total = result.Success + result.Failed + result.Skipped
	s.formatter.Success("Configuration applied successfully!")
	return result, nil
}

// applyPrivileges 应用权限配置
func (s *ApplyService) applyPrivileges() (int, error) {
	s.formatter.Info("Applying privileges...")
	count := 0
	for _, priv := range s.config.Privileges {
		exists, err := s.client.PrivilegeExists(priv.Name)
		if err != nil {
			return count, fmt.Errorf("failed to check privilege %s: %w", priv.Name, err)
		}

		req := nexus.PrivilegeRequest{
			Name:        priv.Name,
			Description: priv.Description,
			Type:        priv.Type,
			Format:      priv.Format,
			Repository:  priv.Repository,
			Actions:     priv.Actions,
		}

		if exists {
			s.formatter.Info(fmt.Sprintf("Privilege %s already exists, skipping...", priv.Name))
			continue
		}

		if err := s.client.CreatePrivilege(req); err != nil {
			return count, fmt.Errorf("failed to create privilege %s: %w", priv.Name, err)
		}
		s.formatter.Success(fmt.Sprintf("Created privilege: %s", priv.Name))
		count++
	}
	return count, nil
}

// applyRoles 应用角色配置
func (s *ApplyService) applyRoles() (int, error) {
	s.formatter.Info("Applying roles...")
	count := 0
	for _, role := range s.config.Roles {
		exists, err := s.client.RoleExists(role.ID)
		if err != nil {
			return count, fmt.Errorf("failed to check role %s: %w", role.ID, err)
		}

		req := nexus.RoleRequest{
			ID:          role.ID,
			Name:        role.Name,
			Description: role.Description,
			Privileges:  role.Privileges,
			Roles:       role.Roles,
		}

		if exists {
			if err := s.client.UpdateRole(role.ID, req); err != nil {
				return count, fmt.Errorf("failed to update role %s: %w", role.ID, err)
			}
			s.formatter.Success(fmt.Sprintf("Updated role: %s", role.ID))
		} else {
			if err := s.client.CreateRole(req); err != nil {
				return count, fmt.Errorf("failed to create role %s: %w", role.ID, err)
			}
			s.formatter.Success(fmt.Sprintf("Created role: %s", role.ID))
		}
		count++
	}
	return count, nil
}

// applyRepositories 应用仓库配置
func (s *ApplyService) applyRepositories() (int, error) {
	s.formatter.Info("Applying repositories...")
	count := 0
	for _, repo := range s.config.Repositories {
		exists, err := s.client.RepositoryExists(repo.Name)
		if err != nil {
			return count, fmt.Errorf("failed to check repository %s: %w", repo.Name, err)
		}

		if exists {
			s.formatter.Info(fmt.Sprintf("Repository %s already exists, skipping...", repo.Name))
			continue
		}

		if err := s.createRepository(repo); err != nil {
			return count, fmt.Errorf("failed to create repository %s: %w", repo.Name, err)
		}
		s.formatter.Success(fmt.Sprintf("Created repository: %s (format: %s, type: %s)", repo.Name, repo.Format, repo.Type))
		count++
	}
	return count, nil
}

// createRepository 创建仓库
func (s *ApplyService) createRepository(repo config.Repository) error {
	req := nexus.RepositoryRequest{
		Name:   repo.Name,
		Online: repo.Online,
		Storage: map[string]interface{}{
			"blobStoreName":               repo.Storage.BlobStoreName,
			"strictContentTypeValidation": repo.Storage.StrictContentTypeValidation,
		},
	}

	// 添加写策略（仅 hosted 类型）
	if repo.Type == "hosted" && repo.Storage.WritePolicy != "" {
		req.Storage["writePolicy"] = repo.Storage.WritePolicy
	}

	// 添加代理配置
	if repo.Proxy != nil {
		req.Proxy = &nexus.ProxySettings{
			RemoteURL:      repo.Proxy.RemoteURL,
			ContentMaxAge:  repo.Proxy.ContentMaxAge,
			MetadataMaxAge: repo.Proxy.MetadataMaxAge,
		}
		if repo.Proxy.Authentication != nil {
			req.Proxy.Authentication = &nexus.Authentication{
				Type:       repo.Proxy.Authentication.Type,
				Username:   repo.Proxy.Authentication.Username,
				Password:   repo.Proxy.Authentication.Password,
				NtlmHost:   repo.Proxy.Authentication.NtlmHost,
				NtlmDomain: repo.Proxy.Authentication.NtlmDomain,
			}
		}

		// 添加必需的代理设置（negativeCache 和 httpClient）
		req.NegativeCache = &nexus.NegativeCacheSettings{
			Enabled:    true,
			TimeToLive: 1440,
		}
		req.HTTPClient = &nexus.HTTPClientSettings{
			Blocked:   false,
			AutoBlock: true,
			Connection: &nexus.HTTPClientConnection{
				RetryAttempts: 3,
				Timeout:       60,
			},
		}
	}

	// 添加 Maven 配置
	if repo.Maven != nil {
		req.Maven = &nexus.MavenSettings{
			VersionPolicy: repo.Maven.VersionPolicy,
			LayoutPolicy:  repo.Maven.LayoutPolicy,
		}
	}

	// 添加 Docker 配置
	if repo.Docker != nil {
		req.Docker = &nexus.DockerSettings{
			ForceBasicAuth: repo.Docker.ForceBasicAuth,
			V1Enabled:      repo.Docker.V1Enabled,
		}
		if repo.Docker.HTTPPort > 0 {
			req.Docker.HTTPPort = &repo.Docker.HTTPPort
		}
		if repo.Docker.HTTPSPort > 0 {
			req.Docker.HTTPSPort = &repo.Docker.HTTPSPort
		}
		if repo.Docker.SubdomainAddr != "" {
			req.Docker.SubdomainAddr = repo.Docker.SubdomainAddr
		}
	}

	// 添加清理策略
	if repo.Cleanup != nil {
		req.Cleanup = &nexus.CleanupPolicy{
			PolicyNames: repo.Cleanup.PolicyNames,
		}
	}

	// 根据格式和类型调用相应的创建方法
	switch repo.Format {
	case "maven2":
		switch repo.Type {
		case "hosted":
			return s.client.CreateMavenHostedRepository(req)
		case "proxy":
			return s.client.CreateMavenProxyRepository(req)
		case "group":
			return s.client.CreateMavenGroupRepository(req)
		}
	case "docker":
		switch repo.Type {
		case "hosted":
			return s.client.CreateDockerHostedRepository(req)
		case "proxy":
			return s.client.CreateDockerProxyRepository(req)
		case "group":
			return s.client.CreateDockerGroupRepository(req)
		}
	case "npm":
		switch repo.Type {
		case "hosted":
			return s.client.CreateNpmHostedRepository(req)
		case "proxy":
			return s.client.CreateNpmProxyRepository(req)
		case "group":
			return s.client.CreateNpmGroupRepository(req)
		}
	case "pypi":
		switch repo.Type {
		case "hosted":
			return s.client.CreatePypiHostedRepository(req)
		case "proxy":
			return s.client.CreatePypiProxyRepository(req)
		case "group":
			return s.client.CreatePypiGroupRepository(req)
		}
	case "go":
		switch repo.Type {
		case "proxy":
			return s.client.CreateGoProxyRepository(req)
		case "group":
			return s.client.CreateGoGroupRepository(req)
		default:
			return fmt.Errorf("go format only supports proxy and group types")
		}
	default:
		return fmt.Errorf("unsupported repository format: %s", repo.Format)
	}

	return fmt.Errorf("unsupported repository type: %s for format: %s", repo.Type, repo.Format)
}

// applyUsers 应用用户配置
func (s *ApplyService) applyUsers() (int, error) {
	s.formatter.Info("Applying users...")
	count := 0
	for _, user := range s.config.Users {
		exists, err := s.client.UserExists(user.ID)
		if err != nil {
			return count, fmt.Errorf("failed to check user %s: %w", user.ID, err)
		}

		req := nexus.UserRequest{
			UserID:       user.ID,
			FirstName:    user.FirstName,
			LastName:     user.LastName,
			EmailAddress: user.EmailAddress,
			Status:       user.Status,
			Roles:        user.Roles,
		}

		if exists {
			// 获取现有用户信息以保留 Source 字段
			existingUser, err := s.client.GetUser(user.ID)
			if err != nil {
				return count, fmt.Errorf("failed to get existing user %s: %w", user.ID, err)
			}
			req.Source = existingUser.Source

			if err := s.client.UpdateUser(user.ID, req); err != nil {
				return count, fmt.Errorf("failed to update user %s: %w", user.ID, err)
			}
			s.formatter.Success(fmt.Sprintf("Updated user: %s", user.ID))

			// 更新密码（如果提供）
			if user.Password != "" {
				if err := s.client.ChangePassword(user.ID, user.Password); err != nil {
					s.formatter.Warning(fmt.Sprintf("Failed to change password for user %s: %v", user.ID, err))
				}
			}
		} else {
			req.Password = user.Password
			if err := s.client.CreateUser(req); err != nil {
				return count, fmt.Errorf("failed to create user %s: %w", user.ID, err)
			}
			s.formatter.Success(fmt.Sprintf("Created user: %s", user.ID))
		}
		count++
	}
	return count, nil
}

// applyUserRepositoryPermissions 应用用户仓库权限
func (s *ApplyService) applyUserRepositoryPermissions() (int, error) {
	s.formatter.Info("Applying user repository permissions...")
	count := 0

	// 为每个用户仓库权限映射创建专门的角色
	for _, perm := range s.config.UserRepositoryPermissions {
		roleName := fmt.Sprintf("%s-%s-role", perm.UserID, perm.Repository)

		// 获取仓库信息以确定 format
		repo, err := s.client.GetRepository(perm.Repository)
		if err != nil {
			return count, fmt.Errorf("failed to get repository %s: %w", perm.Repository, err)
		}
		repoFormat := ""
		if formatVal, ok := repo["format"]; ok {
			if formatStr, ok := formatVal.(string); ok {
				repoFormat = formatStr
			}
		}
		if repoFormat == "" {
			return count, fmt.Errorf("failed to determine format for repository %s", perm.Repository)
		}

		// 构建权限列表
		var privileges []string
		for _, action := range perm.Privileges {
			// 使用 Nexus 内置的权限命名格式: nx-repository-view-{format}-{name}-{action}
			privName := fmt.Sprintf("nx-repository-view-%s-%s-%s", repoFormat, perm.Repository, strings.ToLower(action))
			privileges = append(privileges, privName)
		}

		// 创建或更新角色
		exists, err := s.client.RoleExists(roleName)
		if err != nil {
			return count, fmt.Errorf("failed to check role %s: %w", roleName, err)
		}

		roleReq := nexus.RoleRequest{
			ID:          roleName,
			Name:        fmt.Sprintf("%s access to %s", perm.UserID, perm.Repository),
			Description: fmt.Sprintf("Auto-generated role for %s to access %s", perm.UserID, perm.Repository),
			Privileges:  privileges,
		}

		if exists {
			if err := s.client.UpdateRole(roleName, roleReq); err != nil {
				return count, fmt.Errorf("failed to update role %s: %w", roleName, err)
			}
			s.formatter.Success(fmt.Sprintf("Updated permission role: %s", roleName))
		} else {
			if err := s.client.CreateRole(roleReq); err != nil {
				return count, fmt.Errorf("failed to create role %s: %w", roleName, err)
			}
			s.formatter.Success(fmt.Sprintf("Created permission role: %s", roleName))
		}

		// 更新用户，添加此角色
		user, err := s.client.GetUser(perm.UserID)
		if err != nil {
			return count, fmt.Errorf("failed to get user %s: %w", perm.UserID, err)
		}

		// 检查角色是否已存在
		roleExists := false
		for _, r := range user.Roles {
			if r == roleName {
				roleExists = true
				break
			}
		}

		if !roleExists {
			user.Roles = append(user.Roles, roleName)
			userReq := nexus.UserRequest{
				UserID:       user.UserID,
				FirstName:    user.FirstName,
				LastName:     user.LastName,
				EmailAddress: user.EmailAddress,
				Status:       user.Status,
				Source:       user.Source,
				Roles:        user.Roles,
			}
			if err := s.client.UpdateUser(perm.UserID, userReq); err != nil {
				return count, fmt.Errorf("failed to update user %s with role %s: %w", perm.UserID, roleName, err)
			}
			s.formatter.Success(fmt.Sprintf("Assigned role %s to user %s", roleName, perm.UserID))
		}
		count++
	}

	return count, nil
}
