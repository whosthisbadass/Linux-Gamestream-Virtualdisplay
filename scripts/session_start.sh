#!/usr/bin/env bash
set -euo pipefail

# Sunshine hook entrypoint (pre-launch)
# Reads resolution from env vars if present; falls back to defaults.

WIDTH="${SUNSHINE_CLIENT_WIDTH:-${CLIENT_WIDTH:-1920}}"
HEIGHT="${SUNSHINE_CLIENT_HEIGHT:-${CLIENT_HEIGHT:-1080}}"
REFRESH="${SUNSHINE_CLIENT_FPS:-${CLIENT_FPS:-60}}"

"$(dirname "$0")/create_virtual_display.sh" "$WIDTH" "$HEIGHT" "$REFRESH"
