# ============================================================
# frp-proxy release packaging script (Windows PowerShell)
#
# Usage:
#   .\package.ps1
#   .\package.ps1 -SkipBuild      # skip make build
#   .\package.ps1 -SkipCross      # skip cross-compile
#
# Output layout:
#   release\frps.exe         - server binary (direct-run / test)
#   release\frpc.exe         - client binary (direct-run / test)
#   release\packages\
#     frps-<os>-<arch>-<ver>.zip   (server package)
#     frpc-<os>-<arch>-<ver>.zip   (client package)
#
# Server package (Windows) contents:
#   frps-windows-<arch>-<ver>\
#     frps.exe
#     frps.toml
#     frps-service.ps1
#     LICENSE
#
# Server package (Linux/macOS) contents:
#   frps-<os>-<arch>-<ver>\
#     frps
#     frps.toml
#     frps-service.sh
#     LICENSE
#
# Client package contents:
#   frpc-<os>-<arch>-<ver>\
#     frpc[.exe]
#     frpc.toml
#     LICENSE
# ============================================================

param(
    [switch]$SkipBuild,
    [switch]$SkipCross
)

$ErrorActionPreference = 'Stop'
Set-Location $PSScriptRoot

# 1. Build current-platform binaries into release/
if (-not $SkipBuild) {
    Write-Host "Building current-platform binaries..." -ForegroundColor Cyan
    & make build
    if ($LASTEXITCODE -ne 0) { Write-Error 'make build failed'; exit 1 }
}

$ver = (& .\release\frps.exe --version 2>&1) | Select-Object -First 1
Write-Host "Build version: $ver" -ForegroundColor Green

# 2. Cross-compile all platforms
if (-not $SkipCross) {
    Write-Host "Cross-compiling all platforms..." -ForegroundColor Cyan
    & make -f .\Makefile.cross-compiles
    if ($LASTEXITCODE -ne 0) { Write-Error 'cross-compile failed'; exit 1 }
}

# 3. Prepare output directory (keep historical packages, only add current version)
New-Item -ItemType Directory -Force release\packages | Out-Null

$os_all   = @('linux','windows','darwin','freebsd','openbsd')
$arch_all = @('386','amd64','arm','arm64','mips64','mips64le','mips','mipsle','riscv64','loong64')

Set-Location release

foreach ($os in $os_all) {
    foreach ($arch in $arch_all) {
        if ($os -eq 'windows') {
            $frps_bin = "frps_${os}_${arch}.exe"
            $frpc_bin = "frpc_${os}_${arch}.exe"
        } else {
            $frps_bin = "frps_${os}_${arch}"
            $frpc_bin = "frpc_${os}_${arch}"
        }

        if (-not (Test-Path ".\$frps_bin")) { continue }
        if (-not (Test-Path ".\$frpc_bin")) { continue }

        $frps_dir = "frps-${os}-${arch}-${ver}"
        $frpc_dir = "frpc-${os}-${arch}-${ver}"

        # --- Server package ---
        New-Item -ItemType Directory -Force ".\packages\$frps_dir" | Out-Null
        if ($os -eq 'windows') {
            Move-Item ".\$frps_bin" ".\packages\$frps_dir\frps.exe"
            Copy-Item "..\frps-service.ps1" ".\packages\$frps_dir\frps-service.ps1"
        } else {
            Move-Item ".\$frps_bin" ".\packages\$frps_dir\frps"
            Copy-Item "..\frps-service.sh" ".\packages\$frps_dir\frps-service.sh"
        }
        Copy-Item "..\conf\frps.toml" ".\packages\$frps_dir\frps.toml"
        Copy-Item "..\LICENSE" ".\packages\$frps_dir\LICENSE"

        if ($os -eq 'windows') {
            Compress-Archive -Path ".\packages\$frps_dir" -DestinationPath ".\packages\$frps_dir.zip" -Force
        } else {
            # Use tar if available, otherwise skip non-Windows archives
            if (Get-Command tar -ErrorAction SilentlyContinue) {
                tar -zcf ".\packages\$frps_dir.tar.gz" -C ".\packages" "$frps_dir"
            } else {
                Compress-Archive -Path ".\packages\$frps_dir" -DestinationPath ".\packages\$frps_dir.zip" -Force
            }
        }
        Remove-Item ".\packages\$frps_dir" -Recurse -Force

        # --- Client package ---
        New-Item -ItemType Directory -Force ".\packages\$frpc_dir" | Out-Null
        if ($os -eq 'windows') {
            Move-Item ".\$frpc_bin" ".\packages\$frpc_dir\frpc.exe"
        } else {
            Move-Item ".\$frpc_bin" ".\packages\$frpc_dir\frpc"
        }
        Copy-Item "..\conf\frpc.toml" ".\packages\$frpc_dir\frpc.toml"
        Copy-Item "..\LICENSE" ".\packages\$frpc_dir\LICENSE"

        if ($os -eq 'windows') {
            Compress-Archive -Path ".\packages\$frpc_dir" -DestinationPath ".\packages\$frpc_dir.zip" -Force
        } else {
            if (Get-Command tar -ErrorAction SilentlyContinue) {
                tar -zcf ".\packages\$frpc_dir.tar.gz" -C ".\packages" "$frpc_dir"
            } else {
                Compress-Archive -Path ".\packages\$frpc_dir" -DestinationPath ".\packages\$frpc_dir.zip" -Force
            }
        }
        Remove-Item ".\packages\$frpc_dir" -Recurse -Force

        Write-Host "Packaged: $frps_dir  +  $frpc_dir" -ForegroundColor Green
    }
}

Set-Location $PSScriptRoot
Write-Host "
Done. Packages are in release\packages\" -ForegroundColor Cyan
Get-Item release\packages\* | Select-Object Name,@{N='Size(MB)';E={[math]::Round($_.Length/1MB,1)}}
