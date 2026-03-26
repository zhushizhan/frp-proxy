# Requirements

## Summary

为 `frps` 的 `webui` 增加可视化服务端设置页，让用户可以通过表单启用 HTTPS vhost 和其他常用全局服务端配置，并在保存后自动写回配置文件并重启 `frps`。经典 `/static/` 入口保持不变。

## Requirements

### Requirement 1: Parallel server settings UI

1. `frps` 继续保留经典 Dashboard 入口 `/static/`。
2. `frps` 的新界面继续通过 `/webui/` 提供访问。
3. 服务端可视化配置能力只增加到 `webui/frps`，不影响经典界面现有行为。

### Requirement 2: Visual editing for common server settings

1. 用户应能通过表单配置 HTTPS vhost 支持，包括 `vhostHTTPSPort` 及相关说明。
2. 用户应能通过表单配置其他常用全局服务端设置，例如监听端口、Dashboard、认证、传输层、日志、SSH gateway 等。
3. 与当前表单范围不匹配或结构过于复杂的配置项，应继续保留在文本配置方式中，不应伪装成已完整支持。

### Requirement 3: File-backed save and restart

1. 设置页保存时必须继续写回当前 `frps` 启动所使用的配置文件。
2. 写回前必须进行配置校验，非法配置不能落盘。
3. 保存成功后应沿用现有自动重启机制，使新配置生效。

### Requirement 4: Safe editing behavior

1. UI 不应因为保存服务端设置而破坏经典 Dashboard 入口。
2. 对于仍未在 UI 中建模的复杂配置，应尽量避免在保存时被意外清空或错误覆盖。
3. 页面中应明确提示哪些配置是开关型能力，填写 `0` 或空值时代表禁用或回退默认行为。

### Requirement 5: Validation and local verification

1. `webui/frps` 的 type-check 和 build 必须通过。
2. `frps` Go 构建必须通过。
3. 本地运行时应能访问最新 `webui` bundle，并能通过设置页读写服务端配置。
