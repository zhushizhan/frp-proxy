# Design

## Overview

本次只优化前端布局，不新增后端字段或接口。`ServerSettings.vue` 保持单页表单模型不变，但把设置区域拆成两层：

- 常用区域：核心监听、虚拟主机、Dashboard
- 高级区域：认证、传输、运行参数、日志、HTTP 插件、SSH、其他

## UX

- 常用区域继续使用卡片平铺展示。
- 高级区域改为 `el-collapse`，默认收起。
- 在高级区域顶部提供“Expand All / Collapse All”按钮。
- 为高影响卡片添加醒目的 `el-alert` 提示。

## Validation Interaction

- 当保存时校验失败，如果错误落在高级区域，应自动展开全部高级分组，确保用户能看到错误位置。

## Internationalization

- 新增以下文案组：
  - 布局标题/副标题
  - 展开/收起动作
  - Common / Advanced / High Impact 标记
  - 高影响提示文案
