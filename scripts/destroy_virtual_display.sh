#!/usr/bin/env bash
set -euo pipefail

STATE_DIR="${XDG_RUNTIME_DIR:-/tmp}/sunshine-virtual-display"
STATE_FILE="$STATE_DIR/session.env"

if [[ -f "$STATE_FILE" ]]; then
  # shellcheck disable=SC1090
  source "$STATE_FILE"
fi

if [[ -n "${GAMESCOPE_PID:-}" ]] && kill -0 "$GAMESCOPE_PID" 2>/dev/null; then
  kill "$GAMESCOPE_PID"
  sleep 1
fi

# Fallback kill for stale processes.
pkill -f "gamescope.*--prefer-output VKMS-1" >/dev/null 2>&1 || true

rm -f "$STATE_FILE"

echo "Destroyed virtual display session resources"
