# Requirements

## Summary

为 `frp-proxy` 的客户端和服务端设置页增加文件上传能力，允许用户在 UI 中直接上传证书、密钥和 CA 文件，并写入配置所引用的目标路径。

## Requirements

### Requirement 1: Upload files from settings UI

1. `webui/frps` 和 `webui/frpc` 应支持在设置页中直接选择本地文件并上传。
2. 用户应能指定目标保存路径，或使用表单当前路径作为保存目标。
3. 上传成功后，表单字段应自动回填为最终保存路径。

### Requirement 2: Supported upload scenarios

1. 服务端设置页至少支持：
   - transport TLS cert / key / trusted CA
   - dashboard TLS cert / key / trusted CA
2. 客户端设置页至少支持：
   - transport TLS cert / key / trusted CA
   - OIDC trusted CA

### Requirement 3: File persistence behavior

1. 相对路径应以当前配置文件所在目录为基准解析。
2. 若目标目录不存在，应自动创建。
3. 上传 API 不应要求用户通过 FTP/SSH 手工传文件。

### Requirement 4: Safety and compatibility

1. 经典 `/static/` 入口保持不变。
2. 不改变原有“手填路径”能力，上传只是额外增强方式。
3. 上传 API 需要复用现有认证保护。

### Requirement 5: Verification

1. 相关 Go 测试应覆盖上传路径解析和文件写入。
2. `webui/frpc` 与 `webui/frps` 的构建必须通过。
