# Upstream Sync

This repository is organized around two sibling directories in the local workspace:

- `frp`: clean upstream reference cloned from GitHub
- `frp-proxy`: your customized edition with UI / i18n / guided configuration changes

Inside `frp-proxy`, the remotes are expected to be:

- `origin` -> your feature repository
- `upstream` -> `https://github.com/fatedier/frp.git`

## Recommended Workflow

Run the sync script from inside `frp-proxy`:

```powershell
.\hack\sync-upstream.ps1
```

This will:

1. verify the working tree is clean
2. switch to `main`
3. fetch `upstream/dev`
4. merge `upstream/dev` into `main`

If you also want to push immediately after a successful merge:

```powershell
.\hack\sync-upstream.ps1 -Push
```

## Optional Arguments

```powershell
.\hack\sync-upstream.ps1 -TargetBranch main -UpstreamRemote upstream -UpstreamBranch dev
```

Arguments:

- `-TargetBranch`: local branch to receive upstream changes, default `main`
- `-UpstreamRemote`: upstream remote name, default `upstream`
- `-UpstreamBranch`: upstream branch to merge, default `dev`
- `-Push`: push to `origin/<TargetBranch>` after a successful merge
- `-AllowDirty`: skip the clean working tree check

## Conflict Handling

If the script reports merge conflicts:

1. inspect the conflicted files
2. resolve the conflicts manually
3. run:

```powershell
git add <resolved-files>
git commit
```

After that, if needed:

```powershell
git push origin main
```

## Notes

- The script intentionally uses `merge`, not `rebase`, so upstream sync points remain visible in history.
- The default upstream branch is `dev`, matching the current upstream frp development branch.
