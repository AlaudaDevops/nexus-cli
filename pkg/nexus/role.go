package nexus

import (
	"encoding/json"
	"fmt"
	"strings"
)

// RoleRequest 角色请求
type RoleRequest struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Privileges  []string `json:"privileges,omitempty"`
	Roles       []string `json:"roles,omitempty"`
}

// RoleResponse 角色响应
type RoleResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Privileges  []string `json:"privileges,omitempty"`
	Roles       []string `json:"roles,omitempty"`
	Source      string   `json:"source"`
	ReadOnly    bool     `json:"readOnly"`
}

// CreateRole 创建角色
func (c *Client) CreateRole(req RoleRequest) error {
	_, err := c.post("/service/rest/v1/security/roles", req)
	if err != nil {
		return fmt.Errorf("failed to create role %s: %w", req.ID, err)
	}
	return nil
}

// GetRole 获取角色信息
func (c *Client) GetRole(roleID string) (*RoleResponse, error) {
	data, err := c.get(fmt.Sprintf("/service/rest/v1/security/roles/%s", roleID))
	if err != nil {
		return nil, fmt.Errorf("failed to get role %s: %w", roleID, err)
	}

	var role RoleResponse
	if err := json.Unmarshal(data, &role); err != nil {
		return nil, fmt.Errorf("failed to parse role response: %w", err)
	}

	return &role, nil
}

// UpdateRole 更新角色
func (c *Client) UpdateRole(roleID string, req RoleRequest) error {
	_, err := c.put(fmt.Sprintf("/service/rest/v1/security/roles/%s", roleID), req)
	if err != nil {
		return fmt.Errorf("failed to update role %s: %w", roleID, err)
	}
	return nil
}

// DeleteRole 删除角色
func (c *Client) DeleteRole(roleID string) error {
	_, err := c.delete(fmt.Sprintf("/service/rest/v1/security/roles/%s", roleID))
	if err != nil {
		return fmt.Errorf("failed to delete role %s: %w", roleID, err)
	}
	return nil
}

// RoleExists 检查角色是否存在
func (c *Client) RoleExists(roleID string) (bool, error) {
	_, err := c.GetRole(roleID)
	if err != nil {
		// 404 错误表示角色不存在
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ListRoles 列出所有角色
func (c *Client) ListRoles() ([]RoleResponse, error) {
	data, err := c.get("/service/rest/v1/security/roles")
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	var roles []RoleResponse
	if err := json.Unmarshal(data, &roles); err != nil {
		return nil, fmt.Errorf("failed to parse roles response: %w", err)
	}

	return roles, nil
}
