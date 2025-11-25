package subscription

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"goclash-cli/internal/model"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3" // 必须引入 yaml 包
)

// ClashConfig 用于解析 YAML 格式订阅
type ClashConfig struct {
	Proxies []map[string]interface{} `yaml:"proxies"`
}

// FetchAndParse 下载并解析订阅
func FetchAndParse(subUrl string) ([]*model.ProxyNode, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", subUrl, nil)

	// --- 策略变更 ---
	// 使用 Clash Meta 的 UA，这通常会被机场白名单放行，绕过 Cloudflare
	req.Header.Set("User-Agent", "Clash.Meta/v1.16.0")
	// ----------------

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	rawStr := string(body)

	// DEBUG: 查看这次拿到了什么
	preview := rawStr
	if len(preview) > 100 {
		preview = preview[:100] + "..."
	}
	fmt.Printf("[DEBUG] Response: %s\n", preview)

	// 1. 尝试解析为 Clash YAML
	// 因为我们伪装成了 Clash，大概率会拿到 YAML
	var clashCfg ClashConfig
	if err := yaml.Unmarshal(body, &clashCfg); err == nil && len(clashCfg.Proxies) > 0 {
		fmt.Println("[DEBUG] Detected YAML format")
		return parseClashProxies(clashCfg.Proxies), nil
	}

	// 2. 如果 YAML 解析失败，尝试 Base64 解码 (兜底)
	fmt.Println("[DEBUG] YAML parse failed, trying Base64...")
	decodedBytes, err := base64Decode(rawStr)
	var decodedStr string
	if err != nil {
		decodedStr = rawStr // 可能是明文
	} else {
		decodedStr = string(decodedBytes)
	}

	return parseBase64Lines(decodedStr)
}

// 解析 Clash YAML 结构
func parseClashProxies(proxies []map[string]interface{}) []*model.ProxyNode {
	var nodes []*model.ProxyNode
	for _, p := range proxies {
		// 简单的字段映射
		name, _ := p["name"].(string)
		server, _ := p["server"].(string)
		port := 0
		
		// 处理端口类型
		switch v := p["port"].(type) {
		case int:
			port = v
		case float64:
			port = int(v)
		case string:
			port, _ = strconv.Atoi(v)
		}

		typ, _ := p["type"].(string)
		uuid, _ := p["uuid"].(string)
		password, _ := p["password"].(string)
		cipher, _ := p["cipher"].(string)

		// 过滤掉不支持的类型 (比如 URLTest 等)
		if server == "" || port == 0 {
			continue
		}

		nodes = append(nodes, &model.ProxyNode{
			Name:     name,
			Type:     typ, // ss, vmess, trojan
			Server:   server,
			Port:     port,
			UUID:     uuid,
			Password: password,
			Cipher:   cipher,
		})
	}
	return nodes
}

// 解析 Base64 文本行
func parseBase64Lines(content string) ([]*model.ProxyNode, error) {
	lines := strings.Split(content, "\n")
	var nodes []*model.ProxyNode

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "vmess://") {
			if node, err := parseVmess(line); err == nil {
				nodes = append(nodes, node)
			}
		} else if strings.HasPrefix(line, "ss://") {
			if node, err := parseSS(line); err == nil {
				nodes = append(nodes, node)
			}
		}
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes found")
	}
	return nodes, nil
}

// 辅助函数保持不变...
func base64Decode(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "-", "+")
	s = strings.ReplaceAll(s, "_", "/")
	pad := len(s) % 4
	if pad > 0 {
		s += strings.Repeat("=", 4-pad)
	}
	return base64.StdEncoding.DecodeString(s)
}

func parseVmess(link string) (*model.ProxyNode, error) {
	b64 := strings.TrimPrefix(link, "vmess://")
	jsonBytes, err := base64Decode(b64)
	if err != nil {
		return nil, err
	}
	var v map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &v); err != nil {
		return nil, err
	}
	
	node := &model.ProxyNode{
		Name:   fmt.Sprintf("%v", v["ps"]),
		Type:   "vmess",
		Server: fmt.Sprintf("%v", v["add"]),
		UUID:   fmt.Sprintf("%v", v["id"]),
	}
	switch p := v["port"].(type) {
	case string:
		node.Port, _ = strconv.Atoi(p)
	case float64:
		node.Port = int(p)
	}
	return node, nil
}

func parseSS(link string) (*model.ProxyNode, error) {
	link = strings.TrimPrefix(link, "ss://")
	parts := strings.SplitN(link, "#", 2)
	base64Part := parts[0]
	name := "Shadowsocks"
	if len(parts) > 1 {
		name, _ = url.QueryUnescape(parts[1])
	}

	decoded, err := base64Decode(base64Part)
	if err != nil {
		return nil, err
	}
	userInfo := string(decoded)
	if !strings.Contains(userInfo, "@") {
		return nil, fmt.Errorf("invalid ss format")
	}
	serverParts := strings.SplitN(userInfo, "@", 2)
	authParts := strings.SplitN(serverParts[0], ":", 2)
	addrParts := strings.SplitN(serverParts[1], ":", 2)
	if len(authParts) != 2 || len(addrParts) != 2 {
		return nil, fmt.Errorf("invalid ss parts")
	}
	port, _ := strconv.Atoi(addrParts[1])

	return &model.ProxyNode{
		Name:     name,
		Type:     "ss",
		Server:   addrParts[0],
		Port:     port,
		Cipher:   authParts[0],
		Password: authParts[1],
	}, nil
}