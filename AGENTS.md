# AGENTS.md

## Development Commands

### Build
- `make build` - Build both frps and frpc binaries
- `make frps` - Build server binary only
- `make frpc` - Build client binary only
- `make all` - Build everything with formatting

### Testing
- `make test` - Run unit tests
- `make e2e` - Run end-to-end tests
- `make e2e-trace` - Run e2e tests with trace logging
- `make alltest` - Run all tests including vet, unit tests, and e2e

### Code Quality
- `make fmt` - Run go fmt
- `make fmt-more` - Run gofumpt for more strict formatting
- `make gci` - Run gci import organizer
- `make vet` - Run go vet
- `golangci-lint run` - Run comprehensive linting (configured in .golangci.yml)

### Assets
- `make web` - Build web dashboards (frps and frpc)

### Cleanup
- `make clean` - Remove built binaries and temporary files

## Testing

- E2E tests using Ginkgo/Gomega framework
- Mock servers in `/test/e2e/mock/`
- Run: `make e2e` or `make alltest`

## Agent Runbooks

Operational procedures for agents are in `doc/agents/`:
- `doc/agents/release.md` - Release process

## Release & Packaging Rules (MUST follow)

### Version Bumping
- **Any new feature or bug fix MUST increment the version** in `pkg/util/version/version.go`.
  - New features added on top of a release → bump patch: `v0.X.Y` → `v0.X.Y+1`
  - Breaking changes or major feature sets → bump minor: `v0.X.0` → `v0.X+1.0`
- **Never overwrite or delete an existing GitHub Release.** Each release is permanent history.
- Commit the version bump as a separate commit: `chore: bump version to vX.Y.Z`

### Artifact Naming Convention
- Server-side (frps) artifacts use the **`frps-`** prefix: `frps-<os>-<arch>-vX.Y.Z.tar.gz`
- Client-side (frpc) artifacts use the **`frpc-`** prefix: `frpc-<os>-<arch>-vX.Y.Z.zip`
- Do NOT mix prefixes — never name a server artifact with `frpc-` or vice versa.
- Example:
  - `frps-ubuntu-amd64-v0.68.1.tar.gz` ✅
  - `frpc-windows-amd64-v0.68.1.zip` ✅
  - `frp-proxy-ubuntu-amd64.tar.gz` ❌ (ambiguous)
  - `frps-proxy-ubuntu-amd64.tar.gz` ❌ (wrong prefix)

### Release Notes
- `Release.md` must be written in **bilingual format (English + Chinese)**.
- Each feature/fix entry must have both an English description and a Chinese description.
- New entries go at the **top** of the Features section.
