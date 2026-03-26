# Design

## Overview

本次新增一套轻量文件上传 API，分别挂在 `frpc` 和 `frps` 的管理 API 下。前端设置页保留原有路径输入框，在其旁边增加“选择文件并上传”的操作入口。

## Backend

### API

- `POST /api/files/upload`

请求使用 `multipart/form-data`：

- `targetPath`: 目标路径
- `file`: 上传文件

返回 JSON：

- `savedPath`: 最终写入路径

### Path Resolution

- 若 `targetPath` 是相对路径，则基于当前配置文件目录解析
- 若为空，则使用默认相对路径：
  - `./certs/<原文件名>`

### Shared Helper

新增共享工具函数处理：

- 目标路径解析
- 父目录创建
- 文件写入

## Frontend

### UX

- 保留路径输入框
- 每个可上传字段旁边增加：
  - 文件选择按钮
  - 上传按钮
- 若路径为空，上传时自动按默认规则生成 `./certs/<filename>`
- 上传成功后提示并更新字段值

### Scope

- `ServerSettings.vue`
- `ClientSettings.vue`

必要时可抽一个小型 `FileUploadButton` / helper 函数，但本期不强制做复杂组件抽象。
