# Design

## Overview

本次扩展 `ServerSettings` 数据模型，使其承载一个结构化的 `httpPlugins` 列表，并在 `ServerSettings.vue` 中新增“HTTP Plugins”分组卡片。保存时直接写回 `cfg.HTTPPlugins`，再交由既有 `validation.ValidateServerConfig` 做统一校验。

## Data Model

- 后端
  - `ServerSettings.HTTPPlugins []HTTPPluginSettings`
  - `HTTPPluginSettings` 映射 `v1.HTTPPluginOptions`
- 前端
  - `ServerSettings.httpPlugins: HTTPPluginSettings[]`
  - 受支持的 `ops` 使用固定选项列表：
    - `Login`
    - `NewProxy`
    - `CloseProxy`
    - `Ping`
    - `NewWorkConn`
    - `NewUserConn`

## UI

- 在设置页靠后位置增加插件管理卡片，避免干扰基础监听配置。
- 每个插件使用一张内嵌卡片：
  - 名称
  - 地址
  - Path
  - Ops 多选
  - TLS 校验开关
  - 删除按钮
- 提供“Add Plugin”按钮创建新条目。

## Validation

- 后端继续依赖现有 `SupportedHTTPPluginOps` 校验。
- 前端不做复杂 schema 校验，重点提供结构化输入和受支持 ops 选择，降低输错概率。
