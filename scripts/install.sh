#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN_DEST="/usr/local/bin/sunshine-virtual-display"
SERVICE_DEST="/etc/systemd/system/sunshine-virtual-display.service"

detect_distro() {
  if [[ -f /etc/os-release ]]; then
    . /etc/os-release
    echo "${ID:-unknown}"
  else
    echo "unknown"
  fi
}

install_deps() {
  local distro
  distro="$(detect_distro)"

  case "$distro" in
    ubuntu|debian)
      apt-get update
      apt-get install -y sunshine gamescope pipewire wireplumber xdg-desktop-portal mesa-utils vulkan-tools libdrm-dev jq git build-essential golang-go
      ;;
    fedora)
      dnf install -y sunshine gamescope pipewire wireplumber xdg-desktop-portal mesa-demos vulkan-tools libdrm-devel jq git @development-tools golang
      ;;
    arch)
      pacman -Sy --noconfirm sunshine gamescope pipewire wireplumber xdg-desktop-portal mesa vulkan-tools libdrm jq git base-devel go
      ;;
    *)
      echo "Unsupported distro: $distro" >&2
      exit 1
      ;;
  esac
}

build_project() {
  cd "$ROOT_DIR"
  go build -o sunshine-virtual-display ./cmd/sunshine-virtual-display
}

install_artifacts() {
  install -m 0755 "$ROOT_DIR/sunshine-virtual-display" "$BIN_DEST"
  install -m 0644 "$ROOT_DIR/systemd/sunshine-virtual-display.service" "$SERVICE_DEST"
  modprobe vkms
  systemctl daemon-reload
  systemctl enable --now sunshine-virtual-display.service
}

install_deps
build_project
install_artifacts

echo "Install complete. Use sunshine-virtual-display session-start/session-stop in Sunshine prep commands."
