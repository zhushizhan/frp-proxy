# Requirements

## Summary

为 `webui/frps` 增加中英文国际化能力，保持经典 `/static/` 入口不变，使服务端新界面与客户端新界面在语言体验上保持一致。

## Requirements

### Requirement 1: Bilingual server webui

1. `webui/frps` 应支持至少英文和简体中文两种语言。
2. 用户应能在界面中主动切换语言。
3. 未命中的翻译键应回退到英文，避免界面出现空白。

### Requirement 2: Preserve runtime behavior

1. 经典 Dashboard 入口 `/static/` 必须保持不变。
2. 新界面 `/webui/` 的现有功能和路由不应因国际化而回归。
3. 国际化只作用于 `webui/frps`，不要求同步修改经典服务端前端。

### Requirement 3: Coverage

1. 顶部导航、页面标题、空状态、错误提示、过滤器和详情页核心文案应支持国际化。
2. 新增的服务端设置页文案和字段说明应支持国际化。
3. 常见状态文案和时间相对描述应支持国际化。

### Requirement 4: Verification

1. `webui/frps` 的 type-check 和 build 必须通过。
2. 本地运行后的 `/webui/` 应能正确显示最新翻译 bundle。
