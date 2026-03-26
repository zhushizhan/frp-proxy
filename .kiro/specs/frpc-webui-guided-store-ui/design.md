# Design

## Overview

本功能只新增一套前端入口和交互层，不改变 `frpc` 的后端配置能力边界。核心策略如下：

1. `webui` 与现有 `web` 并行存在。
2. `webui` 继续复用现有 `Store API`、`Config API`、`Reload API`。
3. 后端唯一新增能力是为 `webui` 提供独立静态资源挂载路径 `/webui/`。
4. 交互层以“类型选择 -> 精简表单 -> 调 Store API”取代当前偏通用编辑器的入口方式。

## Runtime integration

### Separate embedded asset package

- 现有经典界面继续使用 `frp/web/frpc` 的 embed 注册和 `/static/` 路由。
- 新界面不复用全局 `assets.Register` 机制，避免覆盖全局 `assets.FileSystem`。
- 新增 `frp/webui/frpc` embed 包，对外暴露 `fs.FS` 或 `http.FileSystem` 供 `client/api_router.go` 直接挂载。

### Route strategy

- 经典界面保留：
  - `/`
  - `/static/...`
- 新界面新增：
  - `/webui/`
  - `/webui/assets/...`
  - `/webui/favicon.ico`
- `/webui/` 及其子路由走同一套受保护子路由，继续复用现有 auth middleware。
- 对 `/webui/` 使用 SPA fallback：若请求的静态文件不存在，则返回 `index.html`。

## Frontend application structure

## Directory strategy

- 直接复制 `frp/web` 为 `frp/webui`，确保共享目录结构一致。
- 本期只对 `frp/webui/frpc` 进行功能开发。
- `frp/webui/frps` 保持可构建骨架，但不接运行时。

## Data flow

### Config page

- 继续复用：
  - `GET /api/config`
  - `PUT /api/config`
  - `GET /api/reload`
- Config 页行为与经典界面一致，仅文案增加“结构化条目由 Store 管理”的提示。

### Proxy and visitor flows

- 列表、创建、更新、删除继续复用：
  - `GET/POST/PUT/DELETE /api/store/proxies`
  - `GET/POST/PUT/DELETE /api/store/visitors`
- 运行状态继续复用已有状态接口和 detail/config 接口。
- 新 UI 的所有说明文案、推荐字段分组和类型入口都属于前端本地视图层数据，不进入后端 schema。

## UX design

### Guided landing pages

- `Proxies` 页改为便捷操作首页：
  - 顶部说明卡
  - 类型选择卡片网格
  - 现有 Store 条目列表
  - 经典界面切换入口
- `Visitors` 页同样改为：
  - 顶部说明卡
  - 类型选择卡片网格
  - 现有 Store 条目列表

### Two-step create flow

- 第一步：选择类型卡片。
- 第二步：进入已锁定类型的编辑页。
- 编辑页沿用现有表单组件，但重排为：
  - 基础信息
  - 高频核心字段
  - 类型特定说明
  - 高级选项折叠区

### Guidance content

- 说明文案来源于 `frp-config-summary.md` 的知识整理，但实现时固化到 `webui/frpc` 本地 guide catalog。
- 每个类型至少提供：
  - 标题
  - 一句话用途
  - 适合的典型场景
  - 注意事项
  - 推荐填写顺序提示

## Component reuse strategy

- 复用现有：
  - `types`
  - `converters`
  - `stores`
  - `api`
  - 表单 section 组件
- 新增前端本地 guide/preset 文件，为各代理/访客类型提供展示元数据。
- 不改变现有 Store API payload 结构。

## Risks and mitigations

### Asset collision risk

- 风险：新旧 UI 都使用嵌入式静态资源，若继续走全局 `assets.Register`，会互相覆盖。
- 规避：`webui` 走独立 embed 包与独立 `http.FileSystem`，只在 `/webui/` 下服务。

### Store-disabled usability risk

- 风险：用户进入新 UI 后无法理解为什么不能新增代理/访客。
- 规避：统一在页面顶部展示 Store 启用提示和配置示例，并禁用结构化新建入口。

### Scope creep risk

- 风险：把插件代理、主配置写回、managed API 等额外能力混入本期。
- 规避：本期仅优化 UI，不新增新后端业务能力，不改变存储模型。
