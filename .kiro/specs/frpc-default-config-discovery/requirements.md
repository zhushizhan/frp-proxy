# Requirements

## Summary

为 `frpc` 增加默认配置自动发现能力，支持在 Windows 上双击 `frpc.exe` 时自动从可执行文件同目录加载配置文件。

## Requirements

### Requirement 1: Auto-discover config when `-c` is omitted

1. 当用户未传入 `-c/--config` 时，`frpc` 应自动查找默认配置文件。
2. 默认查找应优先在可执行文件同目录进行。
3. 若可执行文件同目录未找到，可继续兼容查找当前工作目录中的默认配置文件。

### Requirement 2: Supported default config names

1. 自动发现至少支持：
   - `frpc.toml`
   - `frpc.yaml`
   - `frpc.yml`
   - `frpc.json`
   - `frpc.ini`

### Requirement 3: Preserve explicit config behavior

1. 若用户显式传入 `-c`，应继续按用户指定路径加载，不应覆盖。
2. 现有 `verify`、`reload`、`status`、`stop` 等命令在未显式传 `-c` 时，也应复用自动发现逻辑。

### Requirement 4: Verification

1. 增加单元测试覆盖默认配置查找逻辑。
2. `cmd/frpc` 相关测试应通过。
