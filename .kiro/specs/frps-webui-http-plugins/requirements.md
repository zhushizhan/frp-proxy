# Requirements

## Summary

为 `webui/frps` 的服务端设置页增加 `httpPlugins` 可视化编辑能力，让用户可以通过 UI 配置 frps 的 HTTP 插件列表，并在保存后自动写回配置文件并重启生效。

## Requirements

### Requirement 1: Visual HTTP plugin management

1. 用户应能在 `webui/frps` 中查看当前 `httpPlugins` 列表。
2. 用户应能新增、编辑、删除 HTTP 插件条目。
3. 每个插件条目至少应支持 `name`、`addr`、`path`、`ops`、`tlsVerify`。

### Requirement 2: Safe compatibility

1. 经典 `/static/` 入口必须保持不变。
2. 插件配置应继续写回当前 `frps` 使用的主配置文件。
3. 保存前必须经过既有服务端配置校验，非法 `ops` 或无效配置不能落盘。

### Requirement 3: UX guidance

1. 页面中应说明 HTTP 插件的用途和典型场景。
2. `ops` 应优先提供可选项，而不是要求用户手输。
3. 插件列表为空时，页面应给出清晰提示。

### Requirement 4: Verification

1. `server` 相关测试应覆盖 HTTP 插件的读写映射。
2. `webui/frps` 的 type-check 和 build 必须通过。
3. 本地运行时应能通过 `/webui/` 看到新的插件配置界面。
