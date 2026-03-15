#!/usr/bin/env bash
set -euo pipefail

# create_virtual_display.sh WIDTH HEIGHT REFRESH
# Example: create_virtual_display.sh 2560 1600 120

WIDTH="${1:-1920}"
HEIGHT="${2:-1080}"
REFRESH="${3:-60}"
STATE_DIR="${XDG_RUNTIME_DIR:-/tmp}/sunshine-virtual-display"
STATE_FILE="$STATE_DIR/session.env"

mkdir -p "$STATE_DIR"

if ! lsmod | awk '{print $1}' | grep -qx vkms; then
  modprobe vkms
fi

# Optional: pick compositor command. Gamescope is recommended.
# This command is a template and may need adaptation for your environment.
GAMESCOPE_CMD=(
  gamescope
  -W "$WIDTH"
  -H "$HEIGHT"
  -r "$REFRESH"
  --prefer-output VKMS-1
  -- steam -tenfoot
)

# Launch in background if not already running.
if ! pgrep -f "gamescope.*--prefer-output VKMS-1" >/dev/null 2>&1; then
  nohup "${GAMESCOPE_CMD[@]}" >/tmp/sunshine-gamescope.log 2>&1 &
  GAMESCOPE_PID=$!
else
  GAMESCOPE_PID="$(pgrep -f "gamescope.*--prefer-output VKMS-1" | head -n1)"
fi

cat > "$STATE_FILE" <<STATE
WIDTH=$WIDTH
HEIGHT=$HEIGHT
REFRESH=$REFRESH
GAMESCOPE_PID=$GAMESCOPE_PID
STATE

echo "Created virtual display target ${WIDTH}x${HEIGHT}@${REFRESH}. PID=$GAMESCOPE_PID"
