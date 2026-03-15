#!/usr/bin/env bash
set -euo pipefail

./sunshine-virtual-display session-stop

STATE_FILE="${XDG_RUNTIME_DIR:-/tmp}/sunshine-virtual-display/session-state.json"
if [[ -f "$STATE_FILE" ]]; then
  echo "state file still exists" >&2
  exit 1
fi

echo "session-stop validation passed"
