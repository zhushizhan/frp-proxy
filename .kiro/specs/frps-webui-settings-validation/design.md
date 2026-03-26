# Design

## Overview

本次只增强 `webui/frps` 的交互体验，不新增新的后端接口。`ServerSettings.vue` 继续作为单页表单，通过 Element Plus 的表单校验能力和少量本地预设数据来提升使用效率。

## Validation Scope

- 必填文本：
  - `bindAddr`
  - `proxyBindAddr`
  - `dashboardAddr`
  - `authToken`（当 token inline 模式）
  - `authTokenSourceFile`（当 token file 模式）
  - `oidcIssuer`（当 OIDC 模式）
- 端口范围：
  - `bindPort`
  - `dashboardPort`
- HTTP 插件：
  - `name`
  - `addr`
  - `path`
  - `ops` 至少一个
  - 插件名称不重复
  - `path` 需以 `/` 开头

## Presets

新增前端本地预设 catalog，例如：

- Login Guard
  - `ops = ["Login"]`
- Proxy Lifecycle
  - `ops = ["NewProxy", "CloseProxy"]`
- Access Gate
  - `ops = ["Login", "Ping", "NewWorkConn", "NewUserConn"]`
- Full Audit
  - 所有支持的 `ops`

预设只负责填充默认值，不新增后端字段。

## UX

- 保存按钮触发 `form.validate()`，失败时给出统一 warning。
- HTTP 插件区域在“Add Plugin”旁增加“Add Preset”入口。
- 预设卡片或按钮同时展示名称与一句用途说明。
