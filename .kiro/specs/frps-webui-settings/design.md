# Design

## Overview

本次在现有 `frps webui` 运行接线基础上，继续增加一页“Server Settings”可视化配置页。该页面直接复用本地文件配置作为真实数据源，通过新的 `GET/PUT /api/settings` 与后端交互，保存后由服务端进程自行拉起新实例并退出旧实例。

## Scope

- 保留经典 `/static/` 入口和现有 `webui/frps` 概览、客户端、代理页面。
- 扩展 `ServerSettings` 数据模型，覆盖常用全局服务端配置。
- 保持“文本配置仍然存在”的能力，不尝试把所有复杂嵌套配置都完整可视化。

## Data Model

`ServerSettings` 分层承载以下配置组：

- 基础监听：`bindAddr`、`bindPort`、`proxyBindAddr`、`kcpBindPort`、`quicBindPort`、`tcpmuxHTTPConnectPort`
- Vhost：`vhostHTTPPort`、`vhostHTTPSPort`、`vhostHTTPTimeout`、`subdomainHost`、`tcpmuxPassthrough`
- 认证与访问：`authToken`、`tlsForce`、`allowPorts`
- 传输层：`tcpMux`、`tcpMuxKeepaliveInterval`、`maxPoolCount`、`heartbeatTimeout`、`tcpKeepAlive`、`quic.*`
- Dashboard：地址、端口、账号、密码、pprof、assetsDir、dashboard TLS 路径
- 日志：输出目标、级别、保留天数、是否关闭控制台颜色
- 其他：`udpPacketSize`、`natholeAnalysisDataReserveHours`、`custom404Page`
- SSH gateway：端口与密钥路径

本期不把 `httpPlugins`、`auth.tokenSource.exec` 等复杂列表或动态值源做成完整 UI 表单；这些项目继续由文本配置页维护。

## Backend Design

- `server/http/model/settings.go` 扩展响应/请求字段。
- `server/config_manager.go`
  - 从 `v1.ServerConfig` 映射更多字段到 `ServerSettings`
  - 将表单值写回配置结构后再统一校验
  - 对未纳入 UI 的复杂项，尽量通过“只更新已建模字段”的方式保留原配置
- 延续现有 `scheduleRestart()` 机制，无需新增独立 reload API。

## Frontend Design

- 在 `webui/frps/src/views/ServerSettings.vue` 中扩展多分组卡片表单。
- 每个分组增加简短说明，重点提示：
  - `vhostHTTPSPort > 0` 才能支持客户端 `https` 类型代理
  - `0` 常用于禁用监听端口
  - TLS 证书路径留空时表示不启用对应能力
- 复杂但高价值的高级字段放到后段卡片，不把页面顶部挤成难用的长表单。

## Validation Strategy

- 新增 `server/config_manager_test.go`，验证高级字段映射和写回。
- 运行 `webui/frps` 的 `type-check` 和 `build`。
- 重新构建 `frps.exe` 并本地访问 `/webui/#/settings` 验证。
