# Design

## Overview

复用 `webui/frpc` 已经存在的轻量 i18n 方案，在 `webui/frps` 中新增一套本地消息表、语言检测与存储逻辑。前端通过 `useI18n()` 读取当前语言并渲染文案，不引入新的后端接口。

## Approach

- 新增 `webui/frps/src/i18n/`
  - `index.ts`
  - `messages-en.ts`
  - `messages-zh.ts`
- 在 `App.vue` 头部增加语言切换控件，并新增返回经典界面的链接。
- 逐步替换 `ServerOverview`、`Clients`、`ClientDetail`、`Proxies`、`ProxyDetail`、`ServerSettings` 以及相关组件中的静态文案。
- 对相对时间格式化函数改为读取当前语言，使 `x minutes ago` 等文本可切换。

## Notes

- 类型名如 `TCP`、`HTTPS` 可保留原英文缩写。
- Element Plus 自带默认控件文案本期不强制整体替换，优先覆盖项目内直接输出的业务文案。
