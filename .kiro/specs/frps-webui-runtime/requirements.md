# Requirements

## Summary

为 `frps` 增加与经典 Dashboard 并行的 `webui` 服务端入口，并在本地完成打包与启动验证。

## Requirements

### Requirement 1: Parallel server webui entry

1. `frps` 保留经典 `/static/` Dashboard 不变。
2. `frps` 新增 `/webui/` 入口用于服务 `webui/frps`。
3. `/webui/` 支持 SPA 刷新与深链访问。

### Requirement 2: Separate embedded assets

1. `webui/frps` 不与经典 `web/frps` 共用全局 `assets.Register`。
2. 服务端运行时应能独立获取 `webui/frps` 的嵌入式文件系统。

### Requirement 3: Local package and run validation

1. `webui/frps` 可独立构建。
2. `frps.exe` 可带上 `webui/frps` 一起编译。
3. 本地启动后，经典页面与 `/webui/` 都可访问。
