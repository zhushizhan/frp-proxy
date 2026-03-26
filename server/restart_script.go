package server

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
)

const frpsServiceScriptPS1 = `param(
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
`

const frpsServiceScriptSH = `#!/usr/bin/env sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
ACTION="${1:-start}"
CURRENT_PID="${2:-}"
WAIT_SECONDS="${FRPS_STOP_WAIT_SECONDS:-30}"

EXE_PATH="${FRPS_EXE_PATH:-$SCRIPT_DIR/frps}"
CONFIG_PATH="${FRPS_CONFIG_PATH:-$SCRIPT_DIR/frps.toml}"
WORK_DIR="${FRPS_WORK_DIR:-$SCRIPT_DIR}"

start_frps() {
  cd "$WORK_DIR"
  nohup "$EXE_PATH" -c "$CONFIG_PATH" >/dev/null 2>&1 &
}

stop_and_wait() {
  pid="$1"
  sleep 1
  kill "$pid" 2>/dev/null || true

  elapsed=0
  while kill -0 "$pid" 2>/dev/null; do
    elapsed=$((elapsed + 1))
    if [ "$elapsed" -ge "$WAIT_SECONDS" ]; then
      kill -9 "$pid" 2>/dev/null || true
    fi
    sleep 1
  done
}

case "$ACTION" in
  start)
    start_frps
    ;;
  restart)
    if [ -n "$CURRENT_PID" ]; then
      stop_and_wait "$CURRENT_PID"
    fi
    start_frps
    ;;
  *)
    echo "Unsupported action: $ACTION" >&2
    exit 1
    ;;
esac
`

func ensureFRPSServiceScript(executablePath string) (string, error) {
	exeDir := filepath.Dir(executablePath)
	scriptName := "frps-service.sh"
	content := frpsServiceScriptSH
	mode := os.FileMode(0o755)

	if runtime.GOOS == "windows" {
		scriptName = "frps-service.ps1"
		content = frpsServiceScriptPS1
		mode = 0o644
	}

	scriptPath := filepath.Join(exeDir, scriptName)
	existing, err := os.ReadFile(scriptPath)
	if err == nil && string(existing) == content {
		if runtime.GOOS != "windows" {
			_ = os.Chmod(scriptPath, 0o755)
		}
		return scriptPath, nil
	}
	if err := os.WriteFile(scriptPath, []byte(content), mode); err != nil {
		return "", err
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(scriptPath, 0o755); err != nil {
			return "", err
		}
	}
	return scriptPath, nil
}

func buildFRPSServiceCommand(scriptPath string, currentPID int, executablePath string, configFilePath string, workDir string) (*exec.Cmd, error) {
	if scriptPath == "" {
		return nil, fmt.Errorf("restart script path is required")
	}
	if configFilePath == "" {
		return nil, fmt.Errorf("config file path is required")
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell.exe", "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", scriptPath, "restart", strconv.Itoa(currentPID))
	} else {
		cmd = exec.Command(scriptPath, "restart", strconv.Itoa(currentPID))
	}

	cmd.Dir = filepath.Dir(scriptPath)
	cmd.Env = append(os.Environ(),
		"FRPS_EXE_PATH="+executablePath,
		"FRPS_CONFIG_PATH="+configFilePath,
		"FRPS_WORK_DIR="+workDir,
		"FRPS_STOP_WAIT_SECONDS=30",
	)
	return cmd, nil
}
