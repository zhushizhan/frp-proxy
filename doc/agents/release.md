# Release Process

## 0. Rules (MUST read before every release)

### Version Bumping
- **Any new feature or bug fix MUST increment the version** in `pkg/util/version/version.go`.
  - New features added on top of a release → bump patch: `v0.X.Y` → `v0.X.Y+1`
  - Breaking changes or major feature sets → bump minor: `v0.X.0` → `v0.X+1.0`
- **Never overwrite or delete an existing GitHub Release.** Each release is permanent history.
- Commit the version bump as a **separate commit**: `chore: bump version to vX.Y.Z`

### Artifact Naming Convention
- Server-side (frps) artifacts use the **`frps-`** prefix: `frps-<os>-<arch>-vX.Y.Z.tar.gz`
- Client-side (frpc) artifacts use the **`frpc-`** prefix: `frpc-<os>-<arch>-vX.Y.Z.zip`
- Do NOT mix prefixes — never name a server artifact with `frpc-` or vice versa.
- Correct examples:
  - `frps-ubuntu-amd64-v0.68.1.tar.gz` ✅
  - `frpc-windows-amd64-v0.68.1.zip` ✅
- Wrong examples:
  - `frp-proxy-ubuntu-amd64.tar.gz` ❌ (ambiguous, missing frps/frpc prefix)
  - `frps-proxy-ubuntu-amd64.tar.gz` ❌ (wrong prefix for packaging)
  - `frpc-proxy-windows-amd64.zip` ❌ (wrong prefix for packaging)

### Release Notes Format
- `Release.md` must be written in **bilingual format (English + Chinese)**.
- Each feature/fix entry must have both an English description and a Chinese description.
- New entries go at the **top** of the Features section.
- Section headers: `## Features / 新功能`, `## Improvements / 优化`, `## Fixes / 修复`

---

## 1. Update Release Notes

Edit `Release.md` in the project root with the changes for this version (bilingual):

```markdown
## Features / 新功能

* **[component] English description** — detail.
  
  **[component] 中文说明** — 详情。

## Improvements / 优化

* English improvement description.
  
  中文优化说明。
```

## 2. Bump Version

Update the version string in `pkg/util/version/version.go`:

```go
var version = "0.X.Y"
```

Commit as a **separate commit**:

```bash
git add pkg/util/version/version.go
git commit -m "chore: bump version to v0.X.Y"
```

Also commit feature changes separately before the version bump:

```bash
git add <changed files>
git commit -m "feat: <short description>"
```

## 3. Build Frontend Assets

```bash
make web
```

This builds both frps and frpc web dashboards into `web/frps/dist` and `web/frpc/dist`.

## 4. Cross-Compile Binaries

For targeted release (Ubuntu server + Windows client):

```powershell
# Linux/amd64 server
$env:CGO_ENABLED='0'; $env:GOOS='linux'; $env:GOARCH='amd64'
go build -trimpath -ldflags '-s -w' -tags frps -o .\release\frps_linux_amd64 .\cmd\frps

# Windows/amd64 client
$env:GOOS='windows'; $env:GOARCH='amd64'
go build -trimpath -ldflags '-s -w' -tags frpc -o .\release\frpc_windows_amd64.exe .\cmd\frpc
```

## 5. Package Artifacts

Naming format: `frps-<os>-<arch>-vX.Y.Z.tar.gz` / `frpc-<os>-<arch>-vX.Y.Z.zip`

```powershell
$v = '0.X.Y'
$rel = '.\release'

# Server: frps- prefix, tar.gz
New-Item -ItemType Directory -Force -Path "$rel\frps-ubuntu-amd64"
Copy-Item "$rel\frps_linux_amd64" "$rel\frps-ubuntu-amd64\frps"
Copy-Item .\conf\frps.toml "$rel\frps-ubuntu-amd64\frps.toml"
Copy-Item .\LICENSE "$rel\frps-ubuntu-amd64\LICENSE"
tar -czf "$rel\artifacts\frps-ubuntu-amd64-v$v.tar.gz" -C "$rel" "frps-ubuntu-amd64"
Remove-Item -Recurse -Force "$rel\frps-ubuntu-amd64"

# Client: frpc- prefix, zip
New-Item -ItemType Directory -Force -Path "$rel\frpc-windows-amd64"
Copy-Item "$rel\frpc_windows_amd64.exe" "$rel\frpc-windows-amd64\frpc.exe"
Copy-Item .\conf\frpc.toml "$rel\frpc-windows-amd64\frpc.toml"
Copy-Item .\LICENSE "$rel\frpc-windows-amd64\LICENSE"
Compress-Archive -Force -Path "$rel\frpc-windows-amd64" -DestinationPath "$rel\artifacts\frpc-windows-amd64-v$v.zip"
Remove-Item -Recurse -Force "$rel\frpc-windows-amd64"
```

## 6. Tag and Push

```bash
git tag v0.X.Y <commit-hash>
git push github v0.X.Y
git push github main
```

## 7. Create GitHub Release

```powershell
$v = '0.X.Y'
gh release create "v$v" \
  ".\release\artifacts\frps-ubuntu-amd64-v$v.tar.gz" \
  ".\release\artifacts\frpc-windows-amd64-v$v.zip" \
  --repo zhushizhan/frp-proxy \
  --title "frp-proxy v$v" \
  --notes-file .\Release.md \
  --latest
```

**IMPORTANT**: Never use `release delete` or overwrite an existing release tag. Each release is permanent.

## Key Files

| File | Purpose |
|------|--------|
| `pkg/util/version/version.go` | Version string — bump for every release |
| `Release.md` | Bilingual release notes (EN + ZH) |
| `doc/agents/release.md` | This runbook |
| `AGENTS.md` | Agent rules including naming constraints |

## Versioning Summary

| Change type | Version bump | Example |
|-------------|-------------|--------|
| New feature / bug fix | Patch +1 | `v0.68.0` → `v0.68.1` |
| Major feature set / breaking change | Minor +1 | `v0.68.x` → `v0.69.0` |
