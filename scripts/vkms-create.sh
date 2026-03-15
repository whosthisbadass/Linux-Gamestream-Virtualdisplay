#!/usr/bin/env bash
set -euo pipefail

INSTANCE="${1:-sunshine-manual-$(date +%s)}"
BASE="/sys/kernel/config/vkms/${INSTANCE}"

modprobe vkms

mkdir -p "$BASE"
mkdir -p "$BASE/planes/plane0"
mkdir -p "$BASE/crtcs/crtc0"
mkdir -p "$BASE/encoders/encoder0"
mkdir -p "$BASE/connectors/connector0"

ln -sfn "$BASE/crtcs/crtc0" "$BASE/planes/plane0/possible_crtcs"
ln -sfn "$BASE/crtcs/crtc0" "$BASE/encoders/encoder0/possible_crtcs"
ln -sfn "$BASE/encoders/encoder0" "$BASE/connectors/connector0/possible_encoders"

echo "1" > "$BASE/planes/plane0/type"
echo "1" > "$BASE/connectors/connector0/status"
echo "1" > "$BASE/enabled"

echo "created $INSTANCE"
