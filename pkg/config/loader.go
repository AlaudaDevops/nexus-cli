package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Load 从文件加载配置
func Load(filepath string) (*Config, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// GetNexusCredentials 从环境变量获取 Nexus 认证信息
func GetNexusCredentials() (url, username, password string, err error) {
	url = os.Getenv("NEXUS_URL")
	username = os.Getenv("NEXUS_USERNAME")
	password = os.Getenv("NEXUS_PASSWORD")

	if url == "" {
		return "", "", "", fmt.Errorf("NEXUS_URL environment variable is not set")
	}
	if username == "" {
		return "", "", "", fmt.Errorf("NEXUS_USERNAME environment variable is not set")
	}
	if password == "" {
		return "", "", "", fmt.Errorf("NEXUS_PASSWORD environment variable is not set")
	}

	return url, username, password, nil
}
