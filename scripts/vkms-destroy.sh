#!/usr/bin/env bash
set -euo pipefail

INSTANCE="${1:?instance name required}"
BASE="/sys/kernel/config/vkms/${INSTANCE}"

if [[ ! -d "$BASE" ]]; then
  echo "instance $INSTANCE not found"
  exit 0
fi

echo "0" > "$BASE/enabled"

rm -f "$BASE/connectors/connector0/possible_encoders"
rm -f "$BASE/encoders/encoder0/possible_crtcs"
rm -f "$BASE/planes/plane0/possible_crtcs"

rmdir "$BASE/connectors/connector0" || true
rmdir "$BASE/encoders/encoder0" || true
rmdir "$BASE/crtcs/crtc0" || true
rmdir "$BASE/planes/plane0" || true
rmdir "$BASE/connectors" || true
rmdir "$BASE/encoders" || true
rmdir "$BASE/crtcs" || true
rmdir "$BASE/planes" || true
rmdir "$BASE" || true

echo "destroyed $INSTANCE"
