#!/usr/bin/env bash
set -euo pipefail

WIDTH="${1:?width required}"
HEIGHT="${2:?height required}"
FPS="${3:?fps required}"
TARGET="${SUNSHINE_VD_GAMESCOPE_TARGET:-sleep infinity}"

exec gamescope -W "$WIDTH" -H "$HEIGHT" -r "$FPS" -- sh -lc "$TARGET"
