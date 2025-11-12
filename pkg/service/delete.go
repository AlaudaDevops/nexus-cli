// Package service provides business logic for applying Nexus configurations.
package service

import (
	"fmt"
	"strings"

	"github.com/alauda/nexus-cli/pkg/config"
	"github.com/alauda/nexus-cli/pkg/nexus"
	"github.com/alauda/nexus-cli/pkg/output"
)

// DeleteService 删除服务
type DeleteService struct {
	client    *nexus.Client
	config    *config.Config
	formatter *output.Formatter
}

// DeleteResult 删除结果
type DeleteResult struct {
	Total               int
	Success             int
	Failed              int
	Skipped             int
	UsersDeleted        int
	RepositoriesDeleted int
	RolesDeleted        int
	PrivilegesDeleted   int
	Errors              []string
	Warnings            []string
}

// NewDeleteService 创建删除服务
func NewDeleteService(client *nexus.Client, cfg *config.Config, formatter *output.Formatter) *DeleteService {
	if formatter == nil {
		formatter = output.NewFormatter(output.FormatText, nil)
	}
	return &DeleteService{
		client:    client,
		config:    cfg,
		formatter: formatter,
	}
}

// Delete 删除配置中定义的资源
func (s *DeleteService) Delete() (*DeleteResult, error) {
	result := &DeleteResult{
		Errors:   []string{},
		Warnings: []string{},
	}

	s.formatter.Info("Starting to delete resources...")

	// 删除顺序与创建相反：
	// 1. 用户仓库权限（删除自动创建的角色）
	// 2. 用户
	// 3. 仓库
	// 4. 角色
	// 5. 权限

	// 1. 删除用户仓库权限相关的角色
	count, err := s.deleteUserRepositoryPermissionRoles()
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
		s.formatter.Warning(fmt.Sprintf("Failed to delete some permission roles: %v", err))
	}
	result.Success += count

	// 2. 删除用户
	count, err = s.deleteUsers()
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
		return result, fmt.Errorf("failed to delete users: %w", err)
	}
	result.UsersDeleted = count
	result.Success += count

	// 3. 删除仓库
	count, err = s.deleteRepositories()
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
		return result, fmt.Errorf("failed to delete repositories: %w", err)
	}
	result.RepositoriesDeleted = count
	result.Success += count

	// 4. 删除角色
	count, err = s.deleteRoles()
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
		return result, fmt.Errorf("failed to delete roles: %w", err)
	}
	result.RolesDeleted = count
	result.Success += count

	// 5. 删除权限
	count, err = s.deletePrivileges()
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
		return result, fmt.Errorf("failed to delete privileges: %w", err)
	}
	result.PrivilegesDeleted = count
	result.Success += count

	result.Total = result.Success + result.Failed + result.Skipped
	s.formatter.Success("Resources deleted successfully!")
	return result, nil
}

// deleteUserRepositoryPermissionRoles 删除自动创建的用户仓库权限角色
func (s *DeleteService) deleteUserRepositoryPermissionRoles() (int, error) {
	s.formatter.Info("Deleting user repository permission roles...")
	count := 0

	for _, perm := range s.config.UserRepositoryPermissions {
		roleName := fmt.Sprintf("%s-%s-role", perm.UserID, perm.Repository)

		exists, err := s.client.RoleExists(roleName)
		if err != nil {
			s.formatter.Warning(fmt.Sprintf("Failed to check role %s: %v", roleName, err))
			continue
		}

		if !exists {
			s.formatter.Info(fmt.Sprintf("Role %s does not exist, skipping...", roleName))
			continue
		}

		if err := s.client.DeleteRole(roleName); err != nil {
			s.formatter.Warning(fmt.Sprintf("Failed to delete role %s: %v", roleName, err))
			continue
		}

		s.formatter.Success(fmt.Sprintf("Deleted permission role: %s", roleName))
		count++
	}

	return count, nil
}

// deleteUsers 删除用户
func (s *DeleteService) deleteUsers() (int, error) {
	s.formatter.Info("Deleting users...")
	count := 0

	for _, user := range s.config.Users {
		exists, err := s.client.UserExists(user.ID)
		if err != nil {
			return count, fmt.Errorf("failed to check user %s: %w", user.ID, err)
		}

		if !exists {
			s.formatter.Info(fmt.Sprintf("User %s does not exist, skipping...", user.ID))
			continue
		}

		if err := s.client.DeleteUser(user.ID); err != nil {
			return count, fmt.Errorf("failed to delete user %s: %w", user.ID, err)
		}

		s.formatter.Success(fmt.Sprintf("Deleted user: %s", user.ID))
		count++
	}

	return count, nil
}

// deleteRepositories 删除仓库
func (s *DeleteService) deleteRepositories() (int, error) {
	s.formatter.Info("Deleting repositories...")
	count := 0

	for _, repo := range s.config.Repositories {
		exists, err := s.client.RepositoryExists(repo.Name)
		if err != nil {
			return count, fmt.Errorf("failed to check repository %s: %w", repo.Name, err)
		}

		if !exists {
			s.formatter.Info(fmt.Sprintf("Repository %s does not exist, skipping...", repo.Name))
			continue
		}

		if err := s.client.DeleteRepository(repo.Name); err != nil {
			return count, fmt.Errorf("failed to delete repository %s: %w", repo.Name, err)
		}

		s.formatter.Success(fmt.Sprintf("Deleted repository: %s", repo.Name))
		count++
	}

	return count, nil
}

// deleteRoles 删除角色
func (s *DeleteService) deleteRoles() (int, error) {
	s.formatter.Info("Deleting roles...")
	count := 0

	for _, role := range s.config.Roles {
		exists, err := s.client.RoleExists(role.ID)
		if err != nil {
			return count, fmt.Errorf("failed to check role %s: %w", role.ID, err)
		}

		if !exists {
			s.formatter.Info(fmt.Sprintf("Role %s does not exist, skipping...", role.ID))
			continue
		}

		// 检查角色是否是只读的（内置角色）
		roleInfo, err := s.client.GetRole(role.ID)
		if err != nil {
			return count, fmt.Errorf("failed to get role %s: %w", role.ID, err)
		}

		if roleInfo.ReadOnly {
			s.formatter.Warning(fmt.Sprintf("Role %s is read-only (built-in), skipping...", role.ID))
			continue
		}

		if err := s.client.DeleteRole(role.ID); err != nil {
			// 如果是因为角色被使用而无法删除，给出提示
			if strings.Contains(err.Error(), "in use") || strings.Contains(err.Error(), "assigned") {
				s.formatter.Warning(fmt.Sprintf("Cannot delete role %s: still in use by users", role.ID))
				continue
			}
			return count, fmt.Errorf("failed to delete role %s: %w", role.ID, err)
		}

		s.formatter.Success(fmt.Sprintf("Deleted role: %s", role.ID))
		count++
	}

	return count, nil
}

// deletePrivileges 删除权限
func (s *DeleteService) deletePrivileges() (int, error) {
	s.formatter.Info("Deleting privileges...")
	count := 0

	for _, priv := range s.config.Privileges {
		exists, err := s.client.PrivilegeExists(priv.Name)
		if err != nil {
			return count, fmt.Errorf("failed to check privilege %s: %w", priv.Name, err)
		}

		if !exists {
			s.formatter.Info(fmt.Sprintf("Privilege %s does not exist, skipping...", priv.Name))
			continue
		}

		// 检查权限是否是只读的（内置权限）
		privInfo, err := s.client.GetPrivilege(priv.Name)
		if err != nil {
			return count, fmt.Errorf("failed to get privilege %s: %w", priv.Name, err)
		}

		if privInfo.ReadOnly {
			s.formatter.Warning(fmt.Sprintf("Privilege %s is read-only (built-in), skipping...", priv.Name))
			continue
		}

		if err := s.client.DeletePrivilege(priv.Name); err != nil {
			return count, fmt.Errorf("failed to delete privilege %s: %w", priv.Name, err)
		}

		s.formatter.Success(fmt.Sprintf("Deleted privilege: %s", priv.Name))
		count++
	}

	return count, nil
}
