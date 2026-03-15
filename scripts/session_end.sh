#!/usr/bin/env bash
set -euo pipefail

# Sunshine hook entrypoint (post-session)
"$(dirname "$0")/destroy_virtual_display.sh"
