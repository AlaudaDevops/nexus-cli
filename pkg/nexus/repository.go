package nexus

import (
	"encoding/json"
	"fmt"
	"strings"
)

// RepositoryRequest 通用仓库请求结构
type RepositoryRequest struct {
	Name          string                  `json:"name"`
	Online        bool                    `json:"online"`
	Storage       map[string]interface{}  `json:"storage"`
	Cleanup       *CleanupPolicy          `json:"cleanup,omitempty"`
	Proxy         *ProxySettings          `json:"proxy,omitempty"`
	NegativeCache *NegativeCacheSettings  `json:"negativeCache,omitempty"`
	HTTPClient    *HTTPClientSettings     `json:"httpClient,omitempty"`
	Maven         *MavenSettings          `json:"maven,omitempty"`
	Docker        *DockerSettings         `json:"docker,omitempty"`
	Apt           *AptSettings            `json:"apt,omitempty"`
}

// CleanupPolicy 清理策略
type CleanupPolicy struct {
	PolicyNames []string `json:"policyNames"`
}

// ProxySettings 代理设置
type ProxySettings struct {
	RemoteURL      string          `json:"remoteUrl"`
	ContentMaxAge  int             `json:"contentMaxAge"`
	MetadataMaxAge int             `json:"metadataMaxAge"`
	Authentication *Authentication `json:"authentication,omitempty"`
}

// Authentication 认证信息
type Authentication struct {
	Type       string `json:"type"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	NtlmHost   string `json:"ntlmHost,omitempty"`
	NtlmDomain string `json:"ntlmDomain,omitempty"`
}

// NegativeCacheSettings 负缓存设置
type NegativeCacheSettings struct {
	Enabled    bool `json:"enabled"`
	TimeToLive int  `json:"timeToLive"`
}

// HTTPClientSettings HTTP客户端设置
type HTTPClientSettings struct {
	Blocked    bool                   `json:"blocked"`
	AutoBlock  bool                   `json:"autoBlock"`
	Connection *HTTPClientConnection  `json:"connection,omitempty"`
}

// HTTPClientConnection HTTP客户端连接设置
type HTTPClientConnection struct {
	RetryAttempts int `json:"retries,omitempty"`
	Timeout       int `json:"timeout,omitempty"`
}

// MavenSettings Maven 设置
type MavenSettings struct {
	VersionPolicy string `json:"versionPolicy"`
	LayoutPolicy  string `json:"layoutPolicy"`
}

// DockerSettings Docker 设置
type DockerSettings struct {
	HTTPPort       *int   `json:"httpPort,omitempty"`
	HTTPSPort      *int   `json:"httpsPort,omitempty"`
	ForceBasicAuth bool   `json:"forceBasicAuth"`
	V1Enabled      bool   `json:"v1Enabled"`
	SubdomainAddr  string `json:"subdomain,omitempty"`
}

// AptSettings Apt 设置
type AptSettings struct {
	Distribution string `json:"distribution"`
	Flat         bool   `json:"flat,omitempty"`
}

// CreateMavenHostedRepository 创建 Maven hosted 仓库
func (c *Client) CreateMavenHostedRepository(req RepositoryRequest) error {
	_, err := c.post("/service/rest/v1/repositories/maven/hosted", req)
	if err != nil {
		return fmt.Errorf("failed to create maven hosted repository %s: %w", req.Name, err)
	}
	return nil
}

// CreateMavenProxyRepository 创建 Maven proxy 仓库
func (c *Client) CreateMavenProxyRepository(req RepositoryRequest) error {
	_, err := c.post("/service/rest/v1/repositories/maven/proxy", req)
	if err != nil {
		return fmt.Errorf("failed to create maven proxy repository %s: %w", req.Name, err)
	}
	return nil
}

// CreateMavenGroupRepository 创建 Maven group 仓库
func (c *Client) CreateMavenGroupRepository(req RepositoryRequest) error {
	_, err := c.post("/service/rest/v1/repositories/maven/group", req)
	if err != nil {
		return fmt.Errorf("failed to create maven group repository %s: %w", req.Name, err)
	}
	return nil
}

// CreateDockerHostedRepository 创建 Docker hosted 仓库
func (c *Client) CreateDockerHostedRepository(req RepositoryRequest) error {
	_, err := c.post("/service/rest/v1/repositories/docker/hosted", req)
	if err != nil {
		return fmt.Errorf("failed to create docker hosted repository %s: %w", req.Name, err)
	}
	return nil
}

// CreateDockerProxyRepository 创建 Docker proxy 仓库
func (c *Client) CreateDockerProxyRepository(req RepositoryRequest) error {
	_, err := c.post("/service/rest/v1/repositories/docker/proxy", req)
	if err != nil {
		return fmt.Errorf("failed to create docker proxy repository %s: %w", req.Name, err)
	}
	return nil
}

// CreateDockerGroupRepository 创建 Docker group 仓库
func (c *Client) CreateDockerGroupRepository(req RepositoryRequest) error {
	_, err := c.post("/service/rest/v1/repositories/docker/group", req)
	if err != nil {
		return fmt.Errorf("failed to create docker group repository %s: %w", req.Name, err)
	}
	return nil
}

// CreateNpmHostedRepository 创建 NPM hosted 仓库
func (c *Client) CreateNpmHostedRepository(req RepositoryRequest) error {
	_, err := c.post("/service/rest/v1/repositories/npm/hosted", req)
	if err != nil {
		return fmt.Errorf("failed to create npm hosted repository %s: %w", req.Name, err)
	}
	return nil
}

// CreateNpmProxyRepository 创建 NPM proxy 仓库
func (c *Client) CreateNpmProxyRepository(req RepositoryRequest) error {
	_, err := c.post("/service/rest/v1/repositories/npm/proxy", req)
	if err != nil {
		return fmt.Errorf("failed to create npm proxy repository %s: %w", req.Name, err)
	}
	return nil
}

// CreateNpmGroupRepository 创建 NPM group 仓库
func (c *Client) CreateNpmGroupRepository(req RepositoryRequest) error {
	_, err := c.post("/service/rest/v1/repositories/npm/group", req)
	if err != nil {
		return fmt.Errorf("failed to create npm group repository %s: %w", req.Name, err)
	}
	return nil
}

// GetRepository 获取仓库信息
func (c *Client) GetRepository(name string) (map[string]interface{}, error) {
	data, err := c.get(fmt.Sprintf("/service/rest/v1/repositories/%s", name))
	if err != nil {
		return nil, fmt.Errorf("failed to get repository %s: %w", name, err)
	}

	var repo map[string]interface{}
	if err := json.Unmarshal(data, &repo); err != nil {
		return nil, fmt.Errorf("failed to parse repository response: %w", err)
	}

	return repo, nil
}

// DeleteRepository 删除仓库
func (c *Client) DeleteRepository(name string) error {
	_, err := c.delete(fmt.Sprintf("/service/rest/v1/repositories/%s", name))
	if err != nil {
		return fmt.Errorf("failed to delete repository %s: %w", name, err)
	}
	return nil
}

// RepositoryExists 检查仓库是否存在
func (c *Client) RepositoryExists(name string) (bool, error) {
	_, err := c.GetRepository(name)
	if err != nil {
		// 404 错误表示仓库不存在
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ListRepositories 列出所有仓库
func (c *Client) ListRepositories() ([]map[string]interface{}, error) {
	data, err := c.get("/service/rest/v1/repositories")
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}

	var repos []map[string]interface{}
	if err := json.Unmarshal(data, &repos); err != nil {
		return nil, fmt.Errorf("failed to parse repositories response: %w", err)
	}

	return repos, nil
}

// CreatePypiHostedRepository 创建 PyPI hosted 仓库
func (c *Client) CreatePypiHostedRepository(req RepositoryRequest) error {
	_, err := c.post("/service/rest/v1/repositories/pypi/hosted", req)
	if err != nil {
		return fmt.Errorf("failed to create pypi hosted repository %s: %w", req.Name, err)
	}
	return nil
}

// CreatePypiProxyRepository 创建 PyPI proxy 仓库
func (c *Client) CreatePypiProxyRepository(req RepositoryRequest) error {
	_, err := c.post("/service/rest/v1/repositories/pypi/proxy", req)
	if err != nil {
		return fmt.Errorf("failed to create pypi proxy repository %s: %w", req.Name, err)
	}
	return nil
}

// CreatePypiGroupRepository 创建 PyPI group 仓库
func (c *Client) CreatePypiGroupRepository(req RepositoryRequest) error {
	_, err := c.post("/service/rest/v1/repositories/pypi/group", req)
	if err != nil {
		return fmt.Errorf("failed to create pypi group repository %s: %w", req.Name, err)
	}
	return nil
}

// CreateGoProxyRepository 创建 Go proxy 仓库
func (c *Client) CreateGoProxyRepository(req RepositoryRequest) error {
	_, err := c.post("/service/rest/v1/repositories/go/proxy", req)
	if err != nil {
		return fmt.Errorf("failed to create go proxy repository %s: %w", req.Name, err)
	}
	return nil
}

// CreateGoGroupRepository 创建 Go group 仓库
func (c *Client) CreateGoGroupRepository(req RepositoryRequest) error {
	_, err := c.post("/service/rest/v1/repositories/go/group", req)
	if err != nil {
		return fmt.Errorf("failed to create go group repository %s: %w", req.Name, err)
	}
	return nil
}
