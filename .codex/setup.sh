#!/usr/bin/env bash
set -euo pipefail

echo "=========================================="
echo "Linux-Gamestream-Virtualdisplay"
echo "Codex development environment setup"
echo "=========================================="

# Ensure noninteractive apt
export DEBIAN_FRONTEND=noninteractive

echo "Updating package lists..."
apt-get update

echo "Installing development dependencies..."

apt-get install -y \
    git \
    curl \
    jq \
    ca-certificates \
    build-essential \
    pkg-config \
    gcc \
    make \
    golang-go \
    golangci-lint \
    shellcheck \
    python3 \
    python3-pip \
    python3-venv \
    python3-dbus \
    python3-gi \
    clang \
    llvm \
    libdrm-dev \
    libwayland-dev \
    wayland-protocols \
    mesa-common-dev \
    vulkan-tools \
    dbus \
    file \
    unzip \
    wget

echo "Cleaning apt cache..."
apt-get clean
rm -rf /var/lib/apt/lists/*

echo "Verifying toolchain..."

echo -n "Go version: "
go version

echo -n "Git version: "
git --version

echo -n "Python version: "
python3 --version

echo -n "Shellcheck version: "
shellcheck --version | head -n 1

echo ""
echo "Environment ready for development."
echo ""
echo "Note:"
echo "This Codex environment does NOT run VKMS, Gamescope, or Sunshine."
echo "It only installs the development toolchain required to build the project."
echo ""