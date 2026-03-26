# Requirements

## Summary

优化 `webui/frps` 的 `ServerSettings` 页面布局，通过“高频配置常显、高级配置折叠、高影响项警示”的方式提升可读性与可用性。

## Requirements

### Requirement 1: Separate common and advanced settings

1. 高频使用的服务端设置应优先展示，不应被大量高级字段淹没。
2. 低频或高级配置应放入可折叠区域。
3. 用户应能一键展开或收起高级配置区域。

### Requirement 2: Highlight high-impact changes

1. 可能影响客户端连通性或管理入口可访问性的配置，应提供显式提示。
2. 至少包括核心监听、认证/TLS、Dashboard 入口这三类高影响配置。

### Requirement 3: Preserve existing functionality

1. 不改变现有 `/api/settings` 保存语义。
2. 经典 `/static/` 入口保持不变。
3. 已有的校验、国际化、HTTP 插件编辑和预设能力必须继续可用。

### Requirement 4: Verification

1. `webui/frps` 的 type-check 和 build 必须通过。
2. 本地运行的新界面应能看到新的分层布局和折叠行为。
