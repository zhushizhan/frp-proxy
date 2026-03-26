# Requirements

## Summary

为 `frp-proxy` 增加一套固定的上游同步流程文档与自动化脚本，便于在保留自定义 UI / 国际化增强的前提下，持续从 GitHub 上游 `frp` 同步更新。

## Requirements

### Requirement 1: Repeatable upstream sync workflow

1. `frp-proxy` 应提供一个可重复执行的同步脚本。
2. 脚本应默认将 `upstream/dev` 同步到 `main`。
3. 脚本执行前应检查仓库状态，避免在脏工作区中直接合并。

### Requirement 2: Clear repository boundary documentation

1. 文档应明确说明：
   - `frp` 是纯净上游参考仓库
   - `frp-proxy` 是自定义增强仓库
2. 文档应给出推荐的日常同步命令和冲突处理方式。

### Requirement 3: Safe defaults

1. 同步脚本默认只 fetch + merge，不应强制 push。
2. 若出现冲突，脚本应中止并保留现场，供人工解决。
3. 脚本应支持可选 push 开关，用于在同步成功后推送到 `origin/main`。

### Requirement 4: Verification

1. 脚本文件与文档文件应纳入 `frp-proxy` 仓库。
2. README 至少应提供一个入口指向该同步说明。
