#!/usr/bin/env sh
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
