# Tasks

- [x] Task 1: 建立 `webui` 前端目录与构建接线
  - 复制 `frp/web` 为 `frp/webui`
  - 调整 `webui/frpc` 的构建基准路径到 `/webui/`
  - 为 `frpc` 增加 `webui` 静态资源嵌入与路由挂载
  - 验证新旧 UI 可以并行访问

- [x] Task 2: 为 `webui/frpc` 建立引导式信息架构
  - 增加类型说明 catalog
  - 更新 `App.vue`、导航、页面标题和经典界面切换入口
  - 将代理页和访客页改造为“便捷操作首页”

- [x] Task 3: 改造代理创建与编辑体验
  - 增加类型选择步骤
  - 重排代理编辑页，突出高频字段
  - 为各类型添加说明和注意事项
  - 保持 Store API 请求体兼容

- [x] Task 4: 改造访客创建与编辑体验
  - 增加类型选择步骤
  - 重排访客编辑页，突出高频字段
  - 为私有访问相关类型补充配对说明
  - 保持 Store API 请求体兼容

- [x] Task 5: 兼容 Store-disabled 和 Config 文本编辑体验
  - 在 `webui` 中清晰呈现 Store 未启用时的说明
  - 保留 Config 页原有读写和 reload 能力
  - 明确提示“结构化条目由 Store 管理”

- [x] Task 6: 完成验证
  - 运行 Go 构建验证 `frpc` 可编译
  - 运行 `webui/frpc` 的 type-check 和 build
  - 验证新旧 UI 路由互不冲突
