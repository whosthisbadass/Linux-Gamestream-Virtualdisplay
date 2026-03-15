#!/usr/bin/env bash
set -euo pipefail

INSTANCE="${1:?instance name required}"
BASE="/sys/kernel/config/vkms/${INSTANCE}"

if [[ -d "$BASE" ]]; then
  rm -rf "$BASE"
  echo "destroyed $INSTANCE"
else
  echo "instance $INSTANCE not found"
fi
