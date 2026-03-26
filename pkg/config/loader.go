// Package config provides configuration types and loading functionality.
package config

import (
	"fmt"
	"os"
	"path/filepath"

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

// Save writes configuration to file in YAML format.
func Save(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	return nil
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
