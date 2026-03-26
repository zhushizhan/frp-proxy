# Design

## Overview

本次只做 `frps` 的 `webui` 运行接线，不扩展新的服务端配置编辑能力。

## Approach

- 复用现有 `frps` Dashboard API。
- 为 `webui/frps` 提供独立 `HTTPFileSystem()`，避免覆盖经典 `assets.FileSystem`。
- 在 `server/api_router.go` 中新增 `/webui/` 静态路由和 SPA fallback。
- 经典 `/static/` 路由保持不变。
- 本地运行配置通过独立 `frps-local.toml` 完成，不改仓库示例配置文件。
