#!/usr/bin/env bash
set -euo pipefail

BIN_DEST="/usr/local/bin/sunshine-virtual-display"
SERVICE_DEST="/etc/systemd/system/sunshine-virtual-display.service"

if [[ "${EUID}" -ne 0 ]]; then
  echo "error: uninstall.sh must run as root" >&2
  exit 1
fi

systemctl disable --now sunshine-virtual-display.service 2>/dev/null || true
rm -f "$SERVICE_DEST" "$BIN_DEST"
systemctl daemon-reload

echo "Uninstall complete"
