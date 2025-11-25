package model

// ProxyNode 定义通用的节点结构
type ProxyNode struct {
	Name     string `yaml:"name" json:"name"`
	Type     string `yaml:"type" json:"type"` // ss, vmess, trojan
	Server   string `yaml:"server" json:"server"`
	Port     int    `yaml:"port" json:"port"`
	UUID     string `yaml:"uuid,omitempty" json:"uuid,omitempty"`
	Cipher   string `yaml:"cipher,omitempty" json:"cipher,omitempty"`
	Password string `yaml:"password,omitempty" json:"password,omitempty"`
	Network  string `yaml:"network,omitempty" json:"network,omitempty"` // ws, tcp
}

// Config 定义配置文件结构
type Config struct {
	Port  int          `yaml:"port"`
	Nodes []*ProxyNode `yaml:"proxies"`
}