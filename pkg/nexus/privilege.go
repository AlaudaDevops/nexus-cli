package nexus

import (
	"encoding/json"
	"fmt"
	"strings"
)

// PrivilegeRequest 权限请求
type PrivilegeRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Type        string   `json:"type"`
	Format      string   `json:"format,omitempty"`
	Repository  string   `json:"repository,omitempty"`
	Actions     []string `json:"actions,omitempty"`
	Pattern     string   `json:"pattern,omitempty"`
	Domain      string   `json:"domain,omitempty"`
}

// PrivilegeResponse 权限响应
type PrivilegeResponse struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Type        string   `json:"type"`
	Format      string   `json:"format,omitempty"`
	Repository  string   `json:"repository,omitempty"`
	Actions     []string `json:"actions,omitempty"`
	ReadOnly    bool     `json:"readOnly"`
}

// CreatePrivilege 创建权限
func (c *Client) CreatePrivilege(req PrivilegeRequest) error {
	var path string
	switch req.Type {
	case "repository-view":
		path = "/service/rest/v1/security/privileges/repository-view"
	case "repository-admin":
		path = "/service/rest/v1/security/privileges/repository-admin"
	case "repository-content-selector":
		path = "/service/rest/v1/security/privileges/repository-content-selector"
	case "script":
		path = "/service/rest/v1/security/privileges/script"
	case "application":
		path = "/service/rest/v1/security/privileges/application"
	case "wildcard":
		path = "/service/rest/v1/security/privileges/wildcard"
	default:
		return fmt.Errorf("unsupported privilege type: %s", req.Type)
	}

	_, err := c.post(path, req)
	if err != nil {
		return fmt.Errorf("failed to create privilege %s: %w", req.Name, err)
	}
	return nil
}

// GetPrivilege 获取权限信息
func (c *Client) GetPrivilege(name string) (*PrivilegeResponse, error) {
	data, err := c.get(fmt.Sprintf("/service/rest/v1/security/privileges/%s", name))
	if err != nil {
		return nil, fmt.Errorf("failed to get privilege %s: %w", name, err)
	}

	var priv PrivilegeResponse
	if err := json.Unmarshal(data, &priv); err != nil {
		return nil, fmt.Errorf("failed to parse privilege response: %w", err)
	}

	return &priv, nil
}

// DeletePrivilege 删除权限
func (c *Client) DeletePrivilege(name string) error {
	_, err := c.delete(fmt.Sprintf("/service/rest/v1/security/privileges/%s", name))
	if err != nil {
		return fmt.Errorf("failed to delete privilege %s: %w", name, err)
	}
	return nil
}

// PrivilegeExists 检查权限是否存在
func (c *Client) PrivilegeExists(name string) (bool, error) {
	_, err := c.GetPrivilege(name)
	if err != nil {
		// 404 错误表示权限不存在
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ListPrivileges 列出所有权限
func (c *Client) ListPrivileges() ([]PrivilegeResponse, error) {
	data, err := c.get("/service/rest/v1/security/privileges")
	if err != nil {
		return nil, fmt.Errorf("failed to list privileges: %w", err)
	}

	var privs []PrivilegeResponse
	if err := json.Unmarshal(data, &privs); err != nil {
		return nil, fmt.Errorf("failed to parse privileges response: %w", err)
	}

	return privs, nil
}
