# Requirements

## Summary

为 `frps` 增加默认配置自动发现能力，支持在 Ubuntu 上直接执行 `./frps` 时自动从可执行文件同目录加载配置文件。

## Requirements

### Requirement 1: Auto-discover config when `-c` is omitted

1. 当用户未传入 `-c/--config` 时，`frps` 应自动查找默认配置文件。
2. 默认查找应优先在可执行文件同目录进行。
3. 若同目录未找到，可继续兼容查找当前工作目录中的默认配置文件。

### Requirement 2: Supported default config names

1. 自动发现至少支持：
   - `frps.toml`
   - `frps.yaml`
   - `frps.yml`
   - `frps.json`
   - `frps.ini`

### Requirement 3: Preserve command-line mode

1. 若用户显式传入 `-c`，应继续按指定路径加载配置。
2. 若用户未传 `-c` 且未找到默认配置文件，应继续保持原有“命令行参数配置”模式。
3. `verify` 子命令在未显式传 `-c` 时，也应复用自动发现逻辑。

### Requirement 4: Verification

1. 增加单元测试覆盖默认配置查找逻辑。
2. `cmd/frps` 相关测试应通过。
