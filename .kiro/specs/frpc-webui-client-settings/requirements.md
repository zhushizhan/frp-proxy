# Requirements

## Summary

为 `webui/frpc` 增加与服务端类似的客户端设置表单，让用户可以通过 UI 配置常用客户端参数、启用 Store API，并在保存后自动重载 frpc。

## Requirements

### Requirement 1: Visual client settings

1. `webui/frpc` 应提供结构化客户端设置页，而不只是原始文本编辑。
2. 设置页至少应支持常用客户端参数：
   - `serverAddr`
   - `serverPort`
   - `auth`
   - `webServer`
   - `store`
   - `log`
   - 常见 transport 配置
3. 用户应能直接通过表单启用 Store API，而不必手写 `[store]`.

### Requirement 2: Preserve raw config editing

1. 原有文本配置页仍需保留，作为高级/兜底配置方式。
2. 设置页保存不应阻止用户继续使用文本页。

### Requirement 3: File-backed save and runtime reload

1. 设置页保存时应写回当前 `frpc` 启动所使用的配置文件。
2. 保存前必须经过配置校验。
3. 保存成功后应自动重载 frpc，并且新的 Store 配置应在运行时真正生效。

### Requirement 4: UX

1. 当 Store API 未启用时，页面应清楚引导用户去客户端设置页开启它。
2. 客户端设置页应像服务端设置页一样，优先展示高频配置。

### Requirement 5: Verification

1. 相关 Go 测试应覆盖设置读写和 Store 重建逻辑。
2. `webui/frpc` 的 type-check 和 build 必须通过。
