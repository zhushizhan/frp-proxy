# Tasks

- [x] Task 1: 扩展 `ServerSettings` 后端模型
  - 增加更多常用全局服务端字段
  - 保持现有 `/api/settings` 读写接口不变

- [x] Task 2: 扩展文件配置管理器
  - 从配置文件读取高级字段回填到设置模型
  - 将设置模型写回到 TOML/YAML/JSON 配置文件
  - 避免误覆盖未纳入 UI 的复杂配置

- [x] Task 3: 增强 `webui/frps` 设置页
  - 增加更多分组表单
  - 为 HTTPS vhost、日志、TLS、SSH gateway 等能力补充说明
  - 保持经典入口不变

- [x] Task 4: 增加验证
  - 添加配置管理器测试
  - 运行前端 type-check 和 build
  - 运行 Go 构建并验证本地 `frps webui`
