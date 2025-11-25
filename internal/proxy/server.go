package proxy

import (
	"encoding/binary"
	"fmt"
	"goclash-cli/internal/config"
	"goclash-cli/internal/model"
	"io"
	"log"
	"net"
)

func StartServer() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Config load failed:", err)
	}

	// 监听端口
	addr := fmt.Sprintf("127.0.0.1:%d", cfg.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("[Traffic Agent] Socks5 Server running on: %s\n", addr)
	
	// 获取当前选中的节点（目前先演示获取第一个，实际还需要实现路由逻辑）
	node := config.GetFirstNode()
	if node != nil {
		fmt.Printf("[Router] Default Node: %s (%s)\n", node.Name, node.Type)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		// 并发处理每个连接
		go handleSocks5(conn, node)
	}
}

func handleSocks5(conn net.Conn, node *model.ProxyNode) {
	defer conn.Close()

	// --- 1. 协议版本协商阶段 ---
	buf := make([]byte, 256)
	// 读取: VER, NMETHODS, METHODS
	if _, err := io.ReadFull(conn, buf[:2]); err != nil {
		return
	}
	ver, nMethods := buf[0], buf[1]
	if ver != 5 {
		return // 只支持 Socks5
	}
	if _, err := io.ReadFull(conn, buf[:nMethods]); err != nil {
		return
	}
	// 回复: VER=5, METHOD=00 (无认证)
	conn.Write([]byte{0x05, 0x00})

	// --- 2. 请求阶段 ---
	// 读取: VER, CMD, RSV, ATYP, DST.ADDR, DST.PORT
	if _, err := io.ReadFull(conn, buf[:4]); err != nil {
		return
	}
	cmd := buf[1]
	if cmd != 1 { // 只支持 CONNECT 命令
		return
	}
	addrType := buf[3]
	
	var targetAddr string
	switch addrType {
	case 1: // IPv4
		if _, err := io.ReadFull(conn, buf[:4]); err != nil { return }
		targetAddr = fmt.Sprintf("%d.%d.%d.%d", buf[0], buf[1], buf[2], buf[3])
	case 3: // Domain
		if _, err := io.ReadFull(conn, buf[:1]); err != nil { return }
		addrLen := int(buf[0])
		if _, err := io.ReadFull(conn, buf[:addrLen]); err != nil { return }
		targetAddr = string(buf[:addrLen])
	case 4: // IPv6 (暂略)
		return
	}

	// 读取端口 (2 bytes)
	if _, err := io.ReadFull(conn, buf[:2]); err != nil { return }
	port := binary.BigEndian.Uint16(buf[:2])
	target := fmt.Sprintf("%s:%d", targetAddr, port)

	// --- 3. 响应阶段 ---
	// 告诉客户端：连接建立成功 (这里暂时欺骗客户端说成功了)
	// VER=5, REP=0(Success), RSV=0, ATYP=1, BND.ADDR=0, BND.PORT=0
	conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0})

	// --- 4. 转发阶段 (核心) ---
	log.Printf("[REQ] %s", target)

	// TODO: 这里是接入 VMess/Shadowsocks 协议的地方
	// 目前为了演示跑通流程，我们先做 "Direct" (直连)
	// 也就是让你的 CLI 工具直接去连接目标网站
	
	destConn, err := net.Dial("tcp", target)
	if err != nil {
		log.Printf("[ERR] Dial %s failed: %v", target, err)
		return
	}
	defer destConn.Close()

	// 双向拷贝数据
	go io.Copy(destConn, conn) // 本地 -> 目标
	io.Copy(conn, destConn)    // 目标 -> 本地
}