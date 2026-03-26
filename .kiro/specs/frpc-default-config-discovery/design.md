# Design

## Overview

当前 `frpc` 的 `--config` 默认值是 `./frpc.ini`，这对双击启动不友好，也不兼容当前默认推荐的 TOML 配置。新的实现改为：

- `cfgFile` 默认值为空
- 在真正需要配置路径时调用统一的自动发现函数

## Resolution Order

1. 如果用户显式设置了 `cfgFile`，直接使用
2. 否则，优先查找 `os.Executable()` 所在目录中的默认候选文件
3. 如果可执行文件目录未找到，则查找当前工作目录中的默认候选文件
4. 若仍未找到，则返回清晰错误

## Affected Commands

- root run
- verify
- admin 子命令（reload/status/stop）

## Testing

增加针对候选路径解析和优先级的测试，避免回归。
