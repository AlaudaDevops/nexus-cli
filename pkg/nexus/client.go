// Package nexus provides client for Sonatype Nexus Repository Manager API.
package nexus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client Nexus API 客户端
type Client struct {
	baseURL    string
	username   string
	password   string
	httpClient *http.Client
}

// NewClient 创建新的 Nexus 客户端
func NewClient(baseURL, username, password string) *Client {
	return &Client{
		baseURL:  strings.TrimRight(baseURL, "/"),
		username: username,
		password: password,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// doRequest 执行 HTTP 请求
func (c *Client) doRequest(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	url := c.baseURL + path
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// get 执行 GET 请求
func (c *Client) get(path string) ([]byte, error) {
	return c.doRequest("GET", path, nil)
}

// post 执行 POST 请求
func (c *Client) post(path string, body interface{}) ([]byte, error) {
	return c.doRequest("POST", path, body)
}

// put 执行 PUT 请求
func (c *Client) put(path string, body interface{}) ([]byte, error) {
	return c.doRequest("PUT", path, body)
}

// delete 执行 DELETE 请求
func (c *Client) delete(path string) ([]byte, error) {
	return c.doRequest("DELETE", path, nil)
}

// CheckConnection 检查与 Nexus 的连接
func (c *Client) CheckConnection() error {
	_, err := c.get("/service/rest/v1/status")
	if err != nil {
		return fmt.Errorf("failed to connect to Nexus: %w", err)
	}
	return nil
}
