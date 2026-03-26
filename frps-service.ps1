param(
  [string]$Action = "start",
  [int]$CurrentPid = 0
)

$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$exePath = if ($env:FRPS_EXE_PATH) { $env:FRPS_EXE_PATH } else { Join-Path $scriptDir "frps.exe" }
$configPath = if ($env:FRPS_CONFIG_PATH) { $env:FRPS_CONFIG_PATH } else { Join-Path $scriptDir "frps.toml" }
$workDir = if ($env:FRPS_WORK_DIR) { $env:FRPS_WORK_DIR } else { $scriptDir }
$waitSeconds = if ($env:FRPS_STOP_WAIT_SECONDS) { [int]$env:FRPS_STOP_WAIT_SECONDS } else { 30 }

function Start-Frps {
  Start-Process -FilePath $exePath -ArgumentList @("-c", $configPath) -WorkingDirectory $workDir | Out-Null
}

function Stop-And-Wait([int]$PidToStop) {
  Start-Sleep -Seconds 1
  try {
    Stop-Process -Id $PidToStop -ErrorAction Stop
  } catch {
  }

  for ($elapsed = 0; $elapsed -lt $waitSeconds; $elapsed++) {
    if (-not (Get-Process -Id $PidToStop -ErrorAction SilentlyContinue)) {
      return
    }
    Start-Sleep -Seconds 1
  }

  if (Get-Process -Id $PidToStop -ErrorAction SilentlyContinue) {
    Stop-Process -Id $PidToStop -Force -ErrorAction SilentlyContinue
  }
}

switch ($Action.ToLowerInvariant()) {
  "start" {
    Start-Frps
  }
  "restart" {
    if ($CurrentPid -gt 0) {
      Stop-And-Wait -PidToStop $CurrentPid
    }
    Start-Frps
  }
  default {
    throw "Unsupported action: $Action"
  }
}
