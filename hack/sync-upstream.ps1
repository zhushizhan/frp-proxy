param(
    [string]$TargetBranch = "main",
    [string]$UpstreamRemote = "upstream",
    [string]$UpstreamBranch = "dev",
    [switch]$Push,
    [switch]$AllowDirty
)

$ErrorActionPreference = "Stop"

function Invoke-Git {
    param(
        [Parameter(Mandatory = $true)]
        [string[]]$Args
    )

    & git @Args
    if ($LASTEXITCODE -ne 0) {
        throw "git $($Args -join ' ') failed with exit code $LASTEXITCODE"
    }
}

function Get-GitOutput {
    param(
        [Parameter(Mandatory = $true)]
        [string[]]$Args
    )

    $output = & git @Args
    if ($LASTEXITCODE -ne 0) {
        throw "git $($Args -join ' ') failed with exit code $LASTEXITCODE"
    }
    return $output
}

$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
Push-Location $repoRoot

try {
    $topLevel = (Get-GitOutput -Args @("rev-parse", "--show-toplevel")).Trim()
    if ((Resolve-Path $topLevel).Path -ne $repoRoot.Path) {
        throw "Please run this script inside the frp-proxy repository."
    }

    if (-not $AllowDirty) {
        $status = Get-GitOutput -Args @("status", "--short")
        if ($status) {
            throw "Working tree is not clean. Commit or stash changes first, or rerun with -AllowDirty."
        }
    }

    $remoteNames = Get-GitOutput -Args @("remote")
    if ($remoteNames -notcontains $UpstreamRemote) {
        throw "Missing remote '$UpstreamRemote'."
    }

    $originNames = $remoteNames | Where-Object { $_ -eq "origin" }
    if (-not $originNames) {
        throw "Missing remote 'origin'."
    }

    $currentBranch = (Get-GitOutput -Args @("branch", "--show-current")).Trim()
    if ($currentBranch -ne $TargetBranch) {
        Invoke-Git -Args @("switch", $TargetBranch)
    }

    Write-Host "Fetching $UpstreamRemote/$UpstreamBranch ..."
    Invoke-Git -Args @("fetch", $UpstreamRemote, $UpstreamBranch)

    $mergeRef = "$UpstreamRemote/$UpstreamBranch"
    Write-Host "Merging $mergeRef into $TargetBranch ..."
    & git merge --no-ff --log $mergeRef
    if ($LASTEXITCODE -ne 0) {
        $conflicts = & git diff --name-only --diff-filter=U
        if ($conflicts) {
            Write-Host ""
            Write-Host "Merge conflicts detected:" -ForegroundColor Yellow
            $conflicts | ForEach-Object { Write-Host "  $_" }
            Write-Host ""
            Write-Host "Resolve conflicts, then run:" -ForegroundColor Yellow
            Write-Host "  git add <files>"
            Write-Host "  git commit"
        }
        throw "Merge failed."
    }

    if ($Push) {
        Write-Host "Pushing $TargetBranch to origin ..."
        Invoke-Git -Args @("push", "origin", $TargetBranch)
    }

    Write-Host ""
    Write-Host "Upstream sync completed." -ForegroundColor Green
    Write-Host "Current branch: $TargetBranch"
    if (-not $Push) {
        Write-Host "If everything looks good, push with:"
        Write-Host "  git push origin $TargetBranch"
    }
}
finally {
    Pop-Location
}
