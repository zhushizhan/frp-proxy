# Design

## Overview

`frps` 当前在未指定 `-c` 时会直接走命令行参数模式。新的实现会先尝试自动发现默认配置文件；如果找不到，再回退到命令行参数模式。

## Resolution Order

1. 如果显式设置了 `cfgFile`，直接使用
2. 否则，优先查找 `os.Executable()` 所在目录中的默认候选文件
3. 如果可执行文件目录未找到，则查找当前工作目录中的默认候选文件
4. 若仍未找到：
   - root run：继续走命令行参数模式
   - verify：返回清晰错误

## Affected Commands

- root run
- verify

## Packaging

Ubuntu 服务端包中的 `README.txt` 改为默认推荐使用：

```bash
./frps
```

前提是 `frps.toml` 与 `frps` 位于同一目录。
