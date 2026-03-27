package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 表示应用配置
type Config struct {
	Port            int      `yaml:"port"`
	NodeID          string   `yaml:"nodeId"`
	Secret          string   `yaml:"secret"`
	ReportEndpoints []string `yaml:"report.endpoints"`
}

// LoadConfig 从文件加载配置
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	// 设置默认值
	if cfg.Port == 0 {
		cfg.Port = 8080
	}

	return &cfg, nil
}

// GetConfigPath 获取配置文件路径
func GetConfigPath() string {
	// 首先检查环境变量
	if path := os.Getenv("NETWORK_PROBE_CONFIG"); path != "" {
		return path
	}

	// 默认路径
	return "config/api_node.yaml"
}
