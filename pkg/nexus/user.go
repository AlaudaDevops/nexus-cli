package nexus

import (
	"encoding/json"
	"fmt"
	"strings"
)

// UserRequest 用户创建/更新请求
type UserRequest struct {
	UserID       string   `json:"userId"`
	FirstName    string   `json:"firstName"`
	LastName     string   `json:"lastName"`
	EmailAddress string   `json:"emailAddress"`
	Password     string   `json:"password,omitempty"`
	Status       string   `json:"status"`
	Source       string   `json:"source,omitempty"`
	Roles        []string `json:"roles"`
}

// UserResponse 用户响应
type UserResponse struct {
	UserID       string   `json:"userId"`
	FirstName    string   `json:"firstName"`
	LastName     string   `json:"lastName"`
	EmailAddress string   `json:"emailAddress"`
	Status       string   `json:"status"`
	Roles        []string `json:"roles"`
	Source       string   `json:"source"`
}

// CreateUser 创建用户
func (c *Client) CreateUser(req UserRequest) error {
	_, err := c.post("/service/rest/v1/security/users", req)
	if err != nil {
		return fmt.Errorf("failed to create user %s: %w", req.UserID, err)
	}
	return nil
}

// GetUser 获取用户信息
func (c *Client) GetUser(userID string) (*UserResponse, error) {
	data, err := c.get(fmt.Sprintf("/service/rest/v1/security/users?userId=%s", userID))
	if err != nil {
		return nil, fmt.Errorf("failed to get user %s: %w", userID, err)
	}

	var users []UserResponse
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, fmt.Errorf("failed to parse user response: %w", err)
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("user %s not found", userID)
	}

	return &users[0], nil
}

// UpdateUser 更新用户
func (c *Client) UpdateUser(userID string, req UserRequest) error {
	_, err := c.put(fmt.Sprintf("/service/rest/v1/security/users/%s", userID), req)
	if err != nil {
		return fmt.Errorf("failed to update user %s: %w", userID, err)
	}
	return nil
}

// DeleteUser 删除用户
func (c *Client) DeleteUser(userID string) error {
	_, err := c.delete(fmt.Sprintf("/service/rest/v1/security/users/%s", userID))
	if err != nil {
		return fmt.Errorf("failed to delete user %s: %w", userID, err)
	}
	return nil
}

// UserExists 检查用户是否存在
func (c *Client) UserExists(userID string) (bool, error) {
	_, err := c.GetUser(userID)
	if err != nil {
		// 404 错误表示用户不存在
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ChangePassword 修改用户密码
func (c *Client) ChangePassword(userID, newPassword string) error {
	data := map[string]string{
		"password": newPassword,
	}
	_, err := c.put(fmt.Sprintf("/service/rest/v1/security/users/%s/change-password", userID), data)
	if err != nil {
		return fmt.Errorf("failed to change password for user %s: %w", userID, err)
	}
	return nil
}
