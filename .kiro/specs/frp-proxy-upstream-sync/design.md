# Design

## Overview

本次只增加工作流文档和 PowerShell 自动化脚本，不改动业务代码。由于当前开发环境以 Windows 为主，优先提供 `hack/sync-upstream.ps1`，并在 README 与 `doc/agents/` 中补充使用说明。

## Script Behavior

脚本默认行为：

1. 确认当前仓库位于 `frp-proxy`
2. 校验工作区是否干净
3. 切换到 `main`
4. `git fetch upstream dev`
5. `git merge --no-ff upstream/dev`
6. 可选 `git push origin main`

支持参数：

- `-TargetBranch`，默认 `main`
- `-UpstreamRemote`，默认 `upstream`
- `-UpstreamBranch`，默认 `dev`
- `-Push`，同步成功后自动推送
- `-AllowDirty`，允许跳过脏工作区检查

## Documentation

- `doc/agents/upstream_sync.md`
  - 说明仓库分工
  - 常用同步命令
  - 冲突处理建议
- `README.md` / `README_zh.md`
  - 增加简短入口

## Rationale

- 使用 merge 而非 rebase，便于保留“上游同步点”的历史边界。
- 默认不自动 push，降低误操作风险。
