# Requirements

## Summary

为 `webui/frps` 的 `ServerSettings` 页面增加前端表单校验与 HTTP 插件预设模板，减少用户在保存服务端配置时的试错成本。

## Requirements

### Requirement 1: Frontend validation for common fields

1. 用户在保存前应能获得针对关键字段的前端校验反馈。
2. 关键字段至少包括基础监听、Dashboard 监听、认证必要项和 HTTP 插件必要项。
3. 校验失败时不应发起保存请求。

### Requirement 2: HTTP plugin presets

1. 用户应能在 HTTP 插件区域通过预设快速创建常见插件模板。
2. 预设至少覆盖登录校验、代理生命周期和全量审计这类常见场景。
3. 套用预设后，用户仍可继续编辑插件字段。

### Requirement 3: Compatibility

1. 继续复用现有 `/api/settings` 保存逻辑。
2. 经典 `/static/` 入口必须保持不变。
3. 页面新增体验层能力不应改变后端配置语义。

### Requirement 4: Verification

1. `webui/frps` 的 type-check 和 build 必须通过。
2. 本地运行的新 `webui` 应能展示预设和校验后的设置页。
