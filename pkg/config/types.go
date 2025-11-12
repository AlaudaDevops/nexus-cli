package config

// Config 主配置结构
type Config struct {
	Users                     []User                     `yaml:"users"`
	Repositories              []Repository               `yaml:"repositories"`
	Privileges                []Privilege                `yaml:"privileges"`
	Roles                     []Role                     `yaml:"roles"`
	UserRepositoryPermissions []UserRepositoryPermission `yaml:"userRepositoryPermissions"`
}

// User 用户配置
type User struct {
	ID           string   `yaml:"id"`
	FirstName    string   `yaml:"firstName"`
	LastName     string   `yaml:"lastName"`
	EmailAddress string   `yaml:"emailAddress"`
	Password     string   `yaml:"password"`
	Status       string   `yaml:"status"`
	Roles        []string `yaml:"roles"`
}

// Repository 仓库配置
type Repository struct {
	Name    string         `yaml:"name"`
	Format  string         `yaml:"format"`
	Type    string         `yaml:"type"`
	Online  bool           `yaml:"online"`
	Storage StorageConfig  `yaml:"storage"`
	Proxy   *ProxyConfig   `yaml:"proxy,omitempty"`
	Maven   *MavenConfig   `yaml:"maven,omitempty"`
	Docker  *DockerConfig  `yaml:"docker,omitempty"`
	Apt     *AptConfig     `yaml:"apt,omitempty"`
	Cleanup *CleanupConfig `yaml:"cleanup,omitempty"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	BlobStoreName               string `yaml:"blobStoreName"`
	StrictContentTypeValidation bool   `yaml:"strictContentTypeValidation"`
	WritePolicy                 string `yaml:"writePolicy,omitempty"`
}

// ProxyConfig 代理配置
type ProxyConfig struct {
	RemoteURL      string      `yaml:"remoteUrl"`
	ContentMaxAge  int         `yaml:"contentMaxAge"`
	MetadataMaxAge int         `yaml:"metadataMaxAge"`
	Authentication *AuthConfig `yaml:"authentication,omitempty"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	Type       string `yaml:"type"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
	NtlmHost   string `yaml:"ntlmHost,omitempty"`
	NtlmDomain string `yaml:"ntlmDomain,omitempty"`
}

// MavenConfig Maven 仓库配置
type MavenConfig struct {
	VersionPolicy string `yaml:"versionPolicy"`
	LayoutPolicy  string `yaml:"layoutPolicy"`
}

// DockerConfig Docker 仓库配置
type DockerConfig struct {
	HTTPPort       int    `yaml:"httpPort,omitempty"`
	HTTPSPort      int    `yaml:"httpsPort,omitempty"`
	ForceBasicAuth bool   `yaml:"forceBasicAuth"`
	V1Enabled      bool   `yaml:"v1Enabled"`
	SubdomainAddr  string `yaml:"subdomainAddr,omitempty"`
}

// AptConfig Apt 仓库配置
type AptConfig struct {
	Distribution string `yaml:"distribution"`
	Flat         bool   `yaml:"flat,omitempty"`
}

// CleanupConfig 清理策略配置
type CleanupConfig struct {
	PolicyNames []string `yaml:"policyNames"`
}

// Privilege 权限配置
type Privilege struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Type        string   `yaml:"type"`
	Format      string   `yaml:"format"`
	Repository  string   `yaml:"repository"`
	Actions     []string `yaml:"actions"`
}

// Role 角色配置
type Role struct {
	ID          string   `yaml:"id"`
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Privileges  []string `yaml:"privileges"`
	Roles       []string `yaml:"roles,omitempty"`
}

// UserRepositoryPermission 用户仓库权限映射
type UserRepositoryPermission struct {
	UserID     string   `yaml:"userId"`
	Repository string   `yaml:"repository"`
	Privileges []string `yaml:"privileges"`
}
