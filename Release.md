## Features / 新功能

* **[frpc WebUI] VirtualNet Configuration Wizard** — Added a dedicated "Virtual Network" page to the frpc dashboard. Users can select the node role (server/proxy or client/visitor), fill in the virtual IP (CIDR), secret key and peer address, then generate and copy the TOML config snippet or apply it directly to the Store. Supports bilingual UI (zh/en). *(Alpha: Linux/macOS only, requires root)*

  **[frpc WebUI] 组网配置向导** — frpc 管理面板新增「组网」页面。用户可选择节点角色（服务端/代理 或 客户端/访客），填写虚拟 IP（CIDR）、密钥和对端地址，一键生成 TOML 配置片段或直接应用到 Store。支持中英双语界面。*（Alpha 功能，仅支持 Linux/macOS，需 root 权限）*

* **[frpc WebUI] Hide empty config sections** — The "Config File" sections in both the Proxy list and Visitor list pages are now automatically hidden when no config-file-based proxies/visitors exist, giving a cleaner UI when only Store-managed or API-managed tunnels are in use.

  **[frpc WebUI] 自动隐藏空的配置文件区块** — 代理列表和访客列表页面中的「配置文件」区块，当不存在来自配置文件的代理/访客时自动隐藏，使仅使用 Store 或 API 管理隧道时界面更简洁。

* **[frpc WebUI] Pairing Wizard** — Added a guided two-side pairing wizard for STCP/XTCP/SUDP private proxy connections. The host side fills in a service form and generates a **share code** (URL-safe Base64) or JSON config with one click. The access side pastes the share code or JSON to automatically pre-fill a visitor configuration, chooses a local bind port, saves, and immediately sees the **local access address** with copy and open-link buttons. Entry buttons "Start Pairing" and "Join Pairing" are added to the Visitors page.

  **[frpc WebUI] 互访向导** — 新增面向 STCP/XTCP/SUDP 私有代理的双端互访向导。发起方填写服务表单后一键生成**分享码**（URL-safe Base64）或 JSON 配置；接入方粘贴分享码或 JSON 即可自动填充访客配置，选择本地绑定端口后保存，界面立即展示**本地访问地址**，支持一键复制与跳转链接。访客页面新增「发起互访」和「接入互访」入口按钮。

* **[frps WebUI] Kick Client** — Added a "Kick Client" button to the client detail page in the frps dashboard. Clicking it force-disconnects the selected online client immediately. The button is only shown when the client is online and refreshes the status automatically after the operation. Backend route: `DELETE /api/clients/{key}`.

  **[frps WebUI] 踢出客户端** — frps 管理面板的客户端详情页新增「踢出客户端」按钮，点击后立即强制断开该在线客户端的连接，操作后自动刷新状态。该按钮仅在客户端在线时显示。后端路由：`DELETE /api/clients/{key}`。

* **[frpc WebUI] Copy Connect Address** — Added a connect-address banner to the proxy detail page in the frpc dashboard. When `remote_addr` is available the address is displayed in a monospace badge alongside a one-click copy button that gives visual confirmation ("Copied") for 2 seconds.

  **[frpc WebUI] 一键复制连接地址** — frpc 管理面板的代理详情页新增连接地址展示条。当 `remote_addr` 有值时，以等宽字体显示地址，并提供一键复制按钮，复制成功后按钮显示「已复制」2 秒。

* **[frps] Healthz JSON response** — `GET /healthz` now returns `{"status":"ok","version":"<version>"}` with `Content-Type: application/json` instead of an empty 200 body, making it easier to verify the running version from monitoring tools and load-balancer health checks.

  **[frps] Healthz 返回 JSON** — `GET /healthz` 现在返回 `{"status":"ok","version":"<版本号>"}` 并附带 `Content-Type: application/json`，而非空响应体，便于监控工具和负载均衡健康检查确认当前运行版本。

## Improvements / 优化

* Kept proxy/visitor names as raw config names during completion; moved user-prefix handling to explicit wire-level naming logic.

  保持代理/访客名称为原始配置名称，将用户前缀处理移至显式的线路级命名逻辑。

* Added `noweb` build tag to allow compiling without frontend assets. `make build` now auto-detects missing `web/*/dist` directories and skips embedding, so a fresh clone can build without running `make web` first. The dashboard gracefully returns 404 when assets are not embedded.

  新增 `noweb` 编译标签，支持不嵌入前端资源进行编译。`make build` 自动检测 `web/*/dist` 目录是否存在并跳过嵌入，全新克隆无需先执行 `make web`。未嵌入资源时管理面板优雅返回 404。

* Improved config parsing errors: for `.toml` files, syntax errors now return immediately with parser position details (line/column when available) instead of falling through to YAML/JSON parsing, and TOML type mismatches report field-level errors without misleading line numbers.

  改进配置解析错误提示：`.toml` 文件的语法错误现在立即返回带解析位置（行/列）的详细信息，不再降级到 YAML/JSON 解析；TOML 类型不匹配时报告字段级错误，不再显示误导性行号。

* OIDC auth now caches the access token and refreshes it before expiry, avoiding a new token request on every heartbeat. Falls back to per-request fetch when the provider omits `expires_in`.

  OIDC 认证现在缓存访问令牌并在过期前刷新，避免每次心跳都请求新令牌。当提供方省略 `expires_in` 时回退到每次请求获取。

* Updated frpc settings saves to follow the classic `Save & Reload` behavior while keeping generated config files minimal instead of writing large blocks of defaults.

  frpc 设置保存更新为遵循经典的「保存并重载」行为，同时保持生成的配置文件精简，不写入大量默认值块。

* Updated frps settings saves to use a companion restart script that stops the old process first, waits for it to exit, and then starts the new process on the same port.

  frps 设置保存更新为使用配套的重启脚本：先停止旧进程并等待其退出，再在相同端口启动新进程。

## Previous Features / 历史功能

* Added a built-in `store` capability for frpc, including persisted store source (`[store] path = "..."`), Store CRUD admin APIs (`/api/store/proxies*`, `/api/store/visitors*`) with runtime reload, and Store management pages in the frpc web dashboard.

  为 frpc 添加内置 `store` 功能，包括持久化存储源（`[store] path = "..."`）、Store CRUD 管理 API（`/api/store/proxies*`、`/api/store/visitors*`，支持运行时重载），以及 frpc 管理面板中的 Store 管理页面。

* Added a dedicated `webui` experience for frpc/frps with guided forms, internationalization, and structured settings pages on top of the classic dashboard.

  为 frpc/frps 添加专属 `webui` 体验，包含引导式表单、国际化支持以及基于经典面板之上的结构化设置页面。

* Added client-side settings APIs and UI for frpc, including direct file upload support for certificate, key, CA, and token-source files.

  为 frpc 添加客户端设置 API 和 UI，支持直接上传证书、密钥、CA 及令牌源文件。

* Added server-side settings APIs and UI for frps, including direct file upload support for TLS and SSH-related files.

  为 frps 添加服务端设置 API 和 UI，支持直接上传 TLS 及 SSH 相关文件。
