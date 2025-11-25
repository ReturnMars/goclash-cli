package config

import (
	"goclash-cli/internal/model"
	// "log"
	"os"

	"gopkg.in/yaml.v3"
)

const ConfigFile = "config.yaml"

// LoadConfig 读取配置
func LoadConfig() (*model.Config, error) {
	data, err := os.ReadFile(ConfigFile)
	if err != nil {
		// 如果不存在，返回默认配置
		return &model.Config{Port: 7890, Nodes: []*model.ProxyNode{}}, nil
	}
	var cfg model.Config
	err = yaml.Unmarshal(data, &cfg)
	return &cfg, err
}

// SaveNodes 保存节点到配置
func SaveNodes(nodes []*model.ProxyNode) error {
	cfg, err := LoadConfig()
	if err != nil {
		cfg = &model.Config{Port: 7890}
	}
	cfg.Nodes = nodes // 覆盖旧节点
	
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(ConfigFile, data, 0644)
}

// GetFirstNode 获取第一个可用节点用于演示
func GetFirstNode() *model.ProxyNode {
	cfg, _ := LoadConfig()
	if len(cfg.Nodes) > 0 {
		return cfg.Nodes[0]
	}
	return nil
}