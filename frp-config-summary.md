# frp 配置方式总结

本文基于仓库内的官方资料整理，主要参考：

- `frp/README.md`
- `frp/README_zh.md`
- `frp/conf/frps.toml`
- `frp/conf/frpc.toml`
- `frp/conf/frps_full_example.toml`
- `frp/conf/frpc_full_example.toml`

目标是用一份更适合快速上手的方式，说明 frp 这个内网穿透工具通常怎么配置、常见配置项分别是做什么的、以及有哪些常见使用模式。

## 1. frp 的基本角色

frp 由两个程序组成：

- `frps`：服务端，通常部署在有公网 IP 的机器上，负责接收客户端连接、开放公网访问入口。
- `frpc`：客户端，通常部署在内网机器上，负责把本地服务映射到 `frps`。

可以把它理解成：

1. `frps` 负责“在公网开门”。
2. `frpc` 负责“把内网里的服务接到这扇门上”。

## 2. 配置文件格式

从当前仓库文档来看，frp 现在推荐使用：

- `TOML`
- `YAML`
- `JSON`

旧的 `INI` 配置仍可见于 `conf/legacy/`，但已经被标记为废弃，后续新功能主要面向 `TOML / YAML / JSON`。

仓库中最值得参考的文件是：

- `frp/conf/frps.toml`：服务端最小示例
- `frp/conf/frpc.toml`：客户端最小示例
- `frp/conf/frps_full_example.toml`：服务端完整示例
- `frp/conf/frpc_full_example.toml`：客户端完整示例

## 3. 最小可用配置

### 3.1 frps 最小配置

```toml
bindPort = 7000
```

含义：

- `bindPort` 是 `frps` 监听的控制端口。
- 所有 `frpc` 都会连接到这个端口。

启动命令通常是：

```bash
./frps -c ./frps.toml
```

### 3.2 frpc 最小配置

```toml
serverAddr = "你的公网服务器IP"
serverPort = 7000

[[proxies]]
name = "ssh"
type = "tcp"
localIP = "127.0.0.1"
localPort = 22
remotePort = 6000
```

含义：

- `serverAddr`：`frps` 所在公网机器地址
- `serverPort`：`frps` 的控制端口，要和 `bindPort` 对应
- `[[proxies]]`：定义一条穿透规则
- `type = "tcp"`：表示普通 TCP 端口映射
- `localIP` + `localPort`：内网机器上的真实服务地址
- `remotePort`：暴露到 `frps` 上的公网访问端口

这个配置的效果是：

- 你访问 `frps公网IP:6000`
- 实际上会转发到内网机器的 `127.0.0.1:22`

启动命令通常是：

```bash
./frpc -c ./frpc.toml
```

## 4. frps 服务端通常怎么配

`frps` 偏向“全局入口控制”。常见配置项可以分成下面几组。

### 4.1 基础监听

```toml
bindAddr = "0.0.0.0"
bindPort = 7000
```

- `bindAddr`：监听地址
- `bindPort`：客户端连接入口

如果需要其他传输协议，还可以开：

- `kcpBindPort`
- `quicBindPort`

这对应 frpc 侧的不同 `transport.protocol`。

### 4.2 HTTP/HTTPS 虚拟主机入口

```toml
vhostHTTPPort = 80
vhostHTTPSPort = 443
```

用于支持 `http`、`https` 类型代理。

如果你要把内网网站暴露出去，而不是单纯映射一个 TCP 端口，通常需要配置这两个端口中的一个或两个。

### 4.3 Dashboard 与监控

```toml
webServer.addr = "127.0.0.1"
webServer.port = 7500
webServer.user = "admin"
webServer.password = "admin"
enablePrometheus = true
```

用途：

- 提供 `frps` 的 Web 管理界面
- 查看客户端、代理状态、流量等信息
- 如果开了 `enablePrometheus = true`，还能暴露 `/metrics`

注意：

- 如果对外开放 Dashboard，建议不要使用默认用户名密码。
- 更稳妥的方式是只监听本机或内网，再通过别的安全手段访问。

### 4.4 认证

最常见的是 `token` 认证：

```toml
auth.method = "token"
auth.token = "your-secret-token"
```

客户端要使用同样的配置，否则无法连上。

frp 还支持：

- `OIDC` 认证

但大多数自建场景里，`token` 已经是最常见的方案。

### 4.5 TLS

```toml
transport.tls.force = true
```

含义：

- 强制 `frps` 只接受 TLS 连接

如果你更在意链路安全，这个选项很重要。对应地，`frpc` 侧也需要启用 TLS 并正确配证书或信任链。

### 4.6 端口访问控制

```toml
allowPorts = [
  { start = 2000, end = 3000 },
  { single = 3001 }
]
```

作用：

- 限制 `frpc` 只能申请指定范围的 `remotePort`
- 防止客户端随意占用服务端任意公网端口

对于多人共用一台 frps 的场景，这个配置很实用。

### 4.7 子域名支持

```toml
subDomainHost = "frps.com"
```

配好之后，客户端的 `http/https` 代理可以写：

```toml
subdomain = "blog"
```

这样就能通过：

```text
blog.frps.com
```

访问对应服务。

## 5. frpc 客户端通常怎么配

`frpc` 偏向“声明我要暴露哪些服务”。配置一般由两部分组成：

1. 全局连接配置
2. 多条 `[[proxies]]` 或 `[[visitors]]`

### 5.1 全局连接配置

典型写法：

```toml
serverAddr = "x.x.x.x"
serverPort = 7000
loginFailExit = true

auth.method = "token"
auth.token = "your-secret-token"

transport.protocol = "tcp"
transport.tls.enable = true
```

说明：

- `serverAddr` / `serverPort`：服务端地址和端口
- `loginFailExit`：首次登录失败时是否直接退出
- `auth.*`：认证方式
- `transport.protocol`：连接 `frps` 的协议，可选 `tcp`、`kcp`、`quic`、`websocket`、`wss`
- `transport.tls.enable = true`：当前版本默认就比较偏向启用 TLS

### 5.2 客户端管理接口

```toml
webServer.addr = "127.0.0.1"
webServer.port = 7400
webServer.user = "admin"
webServer.password = "admin"
```

这个接口主要用于：

- `frpc reload`
- `frpc status`
- HTTP API 管理

如果不开这个管理接口，很多动态管理能力就用不了。

## 6. `[[proxies]]` 的常见配置方式

每个 `[[proxies]]` 都代表一条穿透规则。常用类型如下。

### 6.1 TCP 代理

```toml
[[proxies]]
name = "ssh"
type = "tcp"
localIP = "127.0.0.1"
localPort = 22
remotePort = 6000
```

适用场景：

- SSH
- MySQL / PostgreSQL
- Redis
- 远程桌面
- 任意 TCP 服务

这是 frp 最基础、最常用的配置方式。

### 6.2 UDP 代理

```toml
[[proxies]]
name = "dns"
type = "udp"
localIP = "8.8.8.8"
localPort = 53
remotePort = 6001
```

适用场景：

- DNS
- 游戏服务
- 其他基于 UDP 的应用

### 6.3 HTTP 代理

```toml
[[proxies]]
name = "web"
type = "http"
localIP = "127.0.0.1"
localPort = 80
customDomains = ["web.example.com"]
```

特点：

- 通过域名路由
- 支持 `customDomains`
- 支持 `subdomain`
- 支持 `locations`
- 支持请求头改写、Basic Auth、健康检查

如果多人共享同一个 `frps`，HTTP 类型通常比单纯暴露端口更整洁。

### 6.4 HTTPS 代理

```toml
[[proxies]]
name = "secure-web"
type = "https"
localIP = "127.0.0.1"
localPort = 443
customDomains = ["secure.example.com"]
```

适合直接暴露 HTTPS 网站。

### 6.5 STCP 代理

```toml
[[proxies]]
name = "secret_tcp"
type = "stcp"
secretKey = "abcdefg"
localIP = "127.0.0.1"
localPort = 22
allowUsers = ["*"]
```

特点：

- 不直接占用公网 `remotePort`
- 不能像普通 TCP 那样直接通过 `frps:端口` 访问
- 需要另一端配置 `visitor`

适合：

- 想隐藏服务，不愿意直接暴露公网端口
- 仅允许已知客户端通过密钥访问

### 6.6 XTCP 代理

```toml
[[proxies]]
name = "p2p_tcp"
type = "xtcp"
secretKey = "abcdefg"
localIP = "127.0.0.1"
localPort = 22
```

特点：

- 更偏向 P2P 打洞
- 适合对直连能力有要求的场景
- 实际效果会受 NAT 环境影响

### 6.7 TCPMUX 代理

```toml
[[proxies]]
name = "tcpmuxhttpconnect"
type = "tcpmux"
multiplexer = "httpconnect"
localIP = "127.0.0.1"
localPort = 10701
customDomains = ["tunnel1"]
```

特点：

- 在一个入口上复用多个 TCP 服务
- 常见于 HTTP CONNECT 复用模式

## 7. `[[visitors]]` 是什么

`visitor` 主要用于配合：

- `stcp`
- `xtcp`
- `sudp`

这类不直接暴露传统公网端口的模式。

典型示例：

```toml
[[visitors]]
name = "secret_tcp_visitor"
type = "stcp"
serverName = "secret_tcp"
secretKey = "abcdefg"
bindAddr = "127.0.0.1"
bindPort = 9000
```

它的意思是：

- 本地先监听 `127.0.0.1:9000`
- 你访问这个本地端口时，流量会通过 frp 连接到远端定义好的 `secret_tcp`

可以把它理解成“访问端的本地接入器”。

## 8. 一些常见增强项

### 8.1 单个代理禁用

可以在某个代理里写：

```toml
enabled = false
```

这样该代理不会启动。

这个方式比全局 `start = [...]` 更细粒度，也更适合新配置。

### 8.2 配置拆分

主配置文件里可以写：

```toml
includes = ["./confd/*.toml"]
```

适合把多个代理拆到不同文件里管理。

常见用法：

- 主配置只保留全局项
- 每个业务或每台机器单独一个 `proxy` 配置文件

### 8.3 环境变量模板

frp 支持在配置里引用环境变量，例如：

```toml
serverAddr = "{{ .Envs.FRP_SERVER_ADDR }}"
serverPort = 7000

[[proxies]]
name = "ssh"
type = "tcp"
localIP = "127.0.0.1"
localPort = 22
remotePort = {{ .Envs.FRP_SSH_REMOTE_PORT }}
```

适合：

- 容器环境
- CI/CD
- 多环境部署

### 8.4 带宽、压缩、加密、健康检查

在单个代理下经常会配这些：

```toml
transport.bandwidthLimit = "1MB"
transport.bandwidthLimitMode = "client"
transport.useEncryption = false
transport.useCompression = false

healthCheck.type = "tcp"
healthCheck.timeoutSeconds = 3
healthCheck.maxFailed = 3
healthCheck.intervalSeconds = 10
```

用途分别是：

- 限速
- 限速发生在客户端还是服务端
- 代理流量是否额外加密
- 是否压缩
- 本地服务是否健康，不健康时自动摘除

## 9. 校验、热重载和状态查看

frpc 的配置管理能力比较实用，常用命令有：

```bash
frpc verify -c ./frpc.toml
frpc reload -c ./frpc.toml
frpc status -c ./frpc.toml
```

作用：

- `verify`：先检查配置文件是否有错误
- `reload`：热重载配置，不必完全重启
- `status`：查看所有代理状态

要想使用 `reload` 和 `status`，一般需要在 `frpc.toml` 里启用 `webServer` 管理接口。

## 10. 实际使用时最常见的几种模式

### 模式 A：暴露 SSH

最简单最常见：

- `frps` 开 `bindPort`
- `frpc` 定义一个 `type = "tcp"` 的代理
- `remotePort` 映射到本地 `22`

### 模式 B：暴露网站

推荐用：

- `frps` 配 `vhostHTTPPort` / `vhostHTTPSPort`
- `frpc` 用 `type = "http"` 或 `type = "https"`
- 配 `customDomains` 或 `subdomain`

### 模式 C：隐藏式内网访问

推荐用：

- 服务端机器的 `frpc` 配 `stcp`
- 访问端机器的 `frpc` 配 `visitor`

这样不会直接暴露公网端口，安全性更高。

### 模式 D：多业务统一管理

推荐用：

- 主配置写全局项
- 业务代理拆分到 `confd/*.toml`
- 配合环境变量和 `verify/reload`

更适合长期维护。

## 11. 一个更实用的上手思路

如果是第一次使用 frp，建议按这个顺序理解：

1. 先只跑通一个最简单的 `tcp` 代理。
2. 再学习 `http/https` 类型代理。
3. 然后再看认证、TLS、Dashboard。
4. 最后再研究 `stcp`、`xtcp`、插件、VirtualNet 等高级能力。

因为 frp 的核心思路其实并不复杂：

- `frps` 决定入口和规则边界
- `frpc` 决定把哪些本地服务接出去
- `[[proxies]]` 是最核心的配置单元

## 12. 一句话总结

frp 的配置本质上就是：

- 服务端 `frps` 配公网入口、认证、安全和路由能力
- 客户端 `frpc` 配本地服务到公网的映射规则
- 常规场景优先使用 `tcp / http / https`
- 有隐藏访问需求时再使用 `stcp / xtcp + visitor`

如果后续要继续细化，这份总结最适合再往下扩成：

- SSH 穿透配置模板
- Windows 远程桌面配置模板
- NAS/面板站点配置模板
- 多站点共用一台 frps 的域名配置模板
