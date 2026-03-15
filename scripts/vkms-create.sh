#!/usr/bin/env bash
set -euo pipefail

WIDTH="${1:?width required}"
HEIGHT="${2:?height required}"
FPS="${3:?fps required}"
INSTANCE="${4:-sunshine-manual-$(date +%s)}"

modprobe vkms
BASE="/sys/kernel/config/vkms/${INSTANCE}"

mkdir -p "$BASE/planes/plane-1"
mkdir -p "$BASE/crtcs/crtc-1"
mkdir -p "$BASE/encoders/encoder-1"
mkdir -p "$BASE/connectors/connector-1"

ln -sfn "$BASE/crtcs/crtc-1" "$BASE/planes/plane-1/crtc"
ln -sfn "$BASE/crtcs/crtc-1" "$BASE/encoders/encoder-1/crtc"
ln -sfn "$BASE/encoders/encoder-1" "$BASE/connectors/connector-1/encoder"

[[ -f "$BASE/connectors/connector-1/mode" ]] && echo "${WIDTH}x${HEIGHT}@${FPS}" > "$BASE/connectors/connector-1/mode"
[[ -f "$BASE/connectors/connector-1/enabled" ]] && echo "1" > "$BASE/connectors/connector-1/enabled"

echo "created $INSTANCE mode=${WIDTH}x${HEIGHT}@${FPS}"
