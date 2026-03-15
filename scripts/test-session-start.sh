#!/usr/bin/env bash
set -euo pipefail

SUNSHINE_CLIENT_WIDTH=2560 \
SUNSHINE_CLIENT_HEIGHT=1440 \
SUNSHINE_CLIENT_FPS=120 \
SUNSHINE_CLIENT_HDR=0 \
./sunshine-virtual-display session-start

STATE_FILE="${XDG_RUNTIME_DIR:-/tmp}/sunshine-virtual-display/session-state.json"
[[ -f "$STATE_FILE" ]]

jq . "$STATE_FILE"

PID="$(jq -r '.gamescope_pid' "$STATE_FILE")"
kill -0 "$PID"

echo "session-start validation passed"
