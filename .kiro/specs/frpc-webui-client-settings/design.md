# Design

## Overview

本次扩展 `frpc` 管理 API，新增 `GET/PUT /api/settings`，并在 `webui/frpc` 中新增 `ClientSettings` 页面。原有 `ClientConfigure` 页面保留为 Raw Config 文本编辑页。

## Backend

### Data Source

- 从当前 `frpc` 的主配置文件中读取 `ClientCommonConfig`
- 使用通用文档树写回方式保留 `proxies` / `visitors` / 其他未知字段，避免结构化写回破坏原配置

### Runtime Store Rebuild

当前 Store 是否启用不仅取决于配置文件，还取决于运行时是否为 `aggregator` 正确安装了 `StoreSource`。

因此，在 `ReloadFromFile` / `UpdateConfigSource` 过程中，需要：

- 根据新的 `common.Store` 配置重建或移除 `StoreSource`
- 更新 `aggregator.SetStoreSource(...)`
- 同步更新 `svr.storeSource`

## Frontend

### Routes

- 新增 `/settings` -> `ClientSettings`
- 保留 `/config` -> raw text config

### Page Structure

客户端设置页优先覆盖：

- Connection
  - user / clientID
  - serverAddr / serverPort
  - loginFailExit / dnsServer / stunServer
- Auth
  - method
  - additionalScopes
  - token / token file
  - 常见 OIDC 字段
- Admin UI & Store
  - webServer.*
  - store.path
- Logging
  - log.*
- Transport
  - protocol / poolCount / tcpMux / tls / heartbeat / proxyURL

### Store-disabled Guidance

`ProxyList` / `VisitorList` 中的 Store 未启用提示，增加跳转到客户端设置页的入口。
