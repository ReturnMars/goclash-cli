# AGENTS.md - System Architecture & Agent Definition

## Overview

本项目采用模块化 Agent 架构设计，模拟类似 Clash 的流量处理流程。

## Agents (核心组件)

### 1. Interaction Agent (交互代理)

- **职责**: 处理用户 CLI 输入，解析参数，展示输出结果。
- **位置**: `cmd/`
- **能力**:
  - `start`: 唤醒 Traffic Agent。
  - `sub`: 唤醒 Subscription Agent。

### 2. Subscription Agent (订阅代理)

- **职责**: 负责从外部 URL 获取最新的代理节点信息。
- **位置**: `internal/subscription/`
- **工作流**:
  1. **Fetch**: 通过 HTTP/HTTPS 获取远程配置（需处理 User-Agent 伪装）。
  2. **Decode**: 处理 Base64 编码及字符替换。
  3. **Parse**: 识别 `vmess://`, `ss://`, `trojan://` 协议并转换为标准 `Node` 模型。
  4. **Persist**: 将清洗后的数据交给 Config Agent 持久化。

### 3. Config Agent (配置代理)

- **职责**: 管理系统状态与配置文件的读写。
- **位置**: `internal/config/`
- **功能**: 维护 `config.yaml`，提供热重载（Hot-Reload）所需的配置快照。

### 4. Traffic Agent (流量代理)

- **职责**: 核心的数据转发引擎。
- **位置**: `internal/proxy/`
- **子模块**:
  - **Inbound**: 监听本地 Socks5/HTTP 端口 (Default: 7890)。
  - **Router**: 基于规则（Domain/IP）进行分流决策。
  - **Outbound**: 建立到远端 Proxy Server 的连接。

## Data Flow (数据流向)

1. **Update Flow**: `User` -> `CLI (sub update)` -> `Subscription Agent` -> `Config Agent` -> `Disk`.
2. **Traffic Flow**: `App` -> `Inbound (7890)` -> `Router (Match Rule)` -> `Outbound (Node)` -> `Internet`.
