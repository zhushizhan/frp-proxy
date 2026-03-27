# frp-proxy release packaging script for Windows
# Usage: .\package-release.ps1 [-Version v0.68.2]
# Requires: go, npm/node, tar (Windows 10+)

param(
    [string]$Version = ""
)

Set-Location $PSScriptRoot

# ========== Step 1: Get version ==========
if (-not $Version) {
    $Version = (Get-Content .\pkg\util\version\version.go | Select-String 'version = "(.+)"').Matches[0].Groups[1].Value
    $Version = "v$Version"
}
Write-Host "Building version: $Version" -ForegroundColor Cyan

# ========== Step 2: Build web assets ==========
Write-Host "`n[1/5] Building web assets..." -ForegroundColor Cyan
Set-Location .\web\frps;   npm run build --if-present | Select-Object -Last 3; Set-Location $PSScriptRoot
Set-Location .\web\frpc;   npm run build --if-present | Select-Object -Last 3; Set-Location $PSScriptRoot
Set-Location .\webui\frps; npm run build --if-present | Select-Object -Last 3; Set-Location $PSScriptRoot
Set-Location .\webui\frpc; npm run build --if-present | Select-Object -Last 3; Set-Location $PSScriptRoot

# ========== Step 3: Cross-compile binaries ==========
Write-Host "`n[2/5] Cross-compiling binaries..." -ForegroundColor Cyan
$env:CGO_ENABLED = '0'

$env:GOOS = 'linux';   $env:GOARCH = 'amd64'
go build -trimpath -ldflags "-s -w" -tags "frps" -o release\frps_linux_amd64   .\cmd\frps; Write-Host "  frps linux/amd64 done"
go build -trimpath -ldflags "-s -w" -tags "frpc" -o release\frpc_linux_amd64   .\cmd\frpc; Write-Host "  frpc linux/amd64 done"

$env:GOOS = 'windows'; $env:GOARCH = 'amd64'
go build -trimpath -ldflags "-s -w" -tags "frps" -o release\frps_windows_amd64.exe .\cmd\frps; Write-Host "  frps windows/amd64 done"
go build -trimpath -ldflags "-s -w" -tags "frpc" -o release\frpc_windows_amd64.exe .\cmd\frpc; Write-Host "  frpc windows/amd64 done"

# Reset env
Remove-Item Env:\GOOS; Remove-Item Env:\GOARCH; Remove-Item Env:\CGO_ENABLED

# ========== Step 4: Package ==========
Write-Host "`n[3/5] Packaging..." -ForegroundColor Cyan

$pkgDir = ".\release\packages"
New-Item -ItemType Directory -Force -Path $pkgDir | Out-Null

$conf = ".\conf"
$lic  = ".\LICENSE"
$rel  = ".\release"

# --- frps ubuntu: frps + frps.toml + frps-service.sh + LICENSE ---
$d = "frps-ubuntu-amd64-$Version"
New-Item -ItemType Directory -Force -Path "$rel\$d" | Out-Null
Copy-Item "$rel\frps_linux_amd64"  "$rel\$d\frps"
Copy-Item "$conf\frps.toml"        "$rel\$d\"
Copy-Item ".\frps-service.sh"      "$rel\$d\"
Copy-Item $lic                     "$rel\$d\"
tar -czf "$pkgDir\$d.tar.gz" -C $rel $d
Remove-Item -Recurse -Force "$rel\$d"
Write-Host "  packed $d.tar.gz"

# --- frpc ubuntu: frpc + frpc.toml + LICENSE ---
$d = "frpc-ubuntu-amd64-$Version"
New-Item -ItemType Directory -Force -Path "$rel\$d" | Out-Null
Copy-Item "$rel\frpc_linux_amd64"  "$rel\$d\frpc"
Copy-Item "$conf\frpc.toml"        "$rel\$d\"
Copy-Item $lic                     "$rel\$d\"
tar -czf "$pkgDir\$d.tar.gz" -C $rel $d
Remove-Item -Recurse -Force "$rel\$d"
Write-Host "  packed $d.tar.gz"

# --- frps windows: frps.exe + frps.toml + frps-service.ps1 + LICENSE ---
$d = "frps-windows-amd64-$Version"
New-Item -ItemType Directory -Force -Path "$rel\$d" | Out-Null
Copy-Item "$rel\frps_windows_amd64.exe" "$rel\$d\frps.exe"
Copy-Item "$conf\frps.toml"             "$rel\$d\"
Copy-Item ".\frps-service.ps1"          "$rel\$d\"
Copy-Item $lic                          "$rel\$d\"
Compress-Archive -Path "$rel\$d" -DestinationPath "$pkgDir\$d.zip" -Force
Remove-Item -Recurse -Force "$rel\$d"
Write-Host "  packed $d.zip"

# --- frpc windows: frpc.exe + frpc.toml + LICENSE ---
$d = "frpc-windows-amd64-$Version"
New-Item -ItemType Directory -Force -Path "$rel\$d" | Out-Null
Copy-Item "$rel\frpc_windows_amd64.exe" "$rel\$d\frpc.exe"
Copy-Item "$conf\frpc.toml"             "$rel\$d\"
Copy-Item $lic                          "$rel\$d\"
Compress-Archive -Path "$rel\$d" -DestinationPath "$pkgDir\$d.zip" -Force
Remove-Item -Recurse -Force "$rel\$d"
Write-Host "  packed $d.zip"

# ========== Step 5: Summary ==========
Write-Host "`n[4/5] Packages ready:" -ForegroundColor Green
Get-ChildItem $pkgDir | Select-Object Name, @{N='Size';E={"$([math]::Round($_.Length/1MB,1)) MB"}}
Write-Host "`nDone! Upload the files above to GitHub Release $Version" -ForegroundColor Green
