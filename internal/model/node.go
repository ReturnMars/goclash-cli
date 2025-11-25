package model

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
