# Linux-Gamestream-Virtualdisplay — Claude Code Guide

## Project Summary

Go application that orchestrates a dynamic virtual display for Sunshine game streaming on Linux. Creates a temporary virtual monitor (via VKMS) at the exact client-requested resolution, launches Gamescope compositor, streams through Sunshine, then tears everything down cleanly on session end.

Stack: **VKMS → Gamescope → Sunshine**

## Repository Layout

```
cmd/sunshine-virtual-display/   main entrypoint
internal/
  backend/         backend interface (VKMS default, portal placeholder)
  cleanup/         state persistence and lock management
  clientdetector/  parses Sunshine env vars into a request object
  config/          all env var config, optional JSON config file
  display/         display config types
  gamescope/       Gamescope launcher
  lifecycle/       session-start / session-stop controller
  rules/           optional YAML override rule engine
  vkms/            VKMS DRM connector discovery
config/            example rules.yaml and Sunshine hook config
docs/              extended documentation
scripts/           install.sh / uninstall.sh
.codex/            Codex agent configuration (not relevant for Claude Code)
```

## Build & Test

```bash
# Build
go build ./cmd/sunshine-virtual-display

# Test
go test ./...

# Format check
gofmt -l .

# Lint
golangci-lint run ./...

# Shell lint
shellcheck scripts/*.sh
```

> **Note:** This project targets Linux only. Build and tests compile on any platform (no cgo, no Linux syscalls in test paths), but the binary only runs meaningfully on a Linux host with VKMS, Gamescope, and Sunshine installed.

## Architecture Constraints (do not violate)

- **VKMS is the default backend** — do not change the default or add negotiation
- **Gamescope is the compositor** — do not swap it out
- **Exact client resolution** — no silent resolution negotiation; if a rule doesn't match, use the raw client values
- **Idempotent lifecycle** — session-start and session-stop must be safe to call multiple times
- **Privileged VKMS integration tests** must NOT be added to default CI (they require a real kernel module)

## Key Config Env Vars

| Variable | Default | Purpose |
|---|---|---|
| `SVD_BACKEND` | `vkms` | Backend driver |
| `SVD_FORCE_CONNECTOR` | — | Override DRM connector |
| `SVD_PREFER_NEWEST_VKMS_CONNECTOR` | — | Connector selection hint |
| `SVD_DRY_RUN` | — | Skip actual system calls |
| `SVD_GAMESCOPE_LOG_PATH` | — | Gamescope log file |
| `SVD_GAMESCOPE_STARTUP_TIMEOUT_SEC` | — | Gamescope startup wait |
| `SVD_GAMESCOPE_GENERATE_DRM_MODE` | — | Auto-generate DRM mode |

Sunshine client inputs read from env:
- `SUNSHINE_CLIENT_WIDTH`, `SUNSHINE_CLIENT_HEIGHT`, `SUNSHINE_CLIENT_FPS`
- `SUNSHINE_CLIENT_NAME`, `SUNSHINE_CLIENT_UID`, `SUNSHINE_CLIENT_HDR`

## Optional Override Rules

Loaded from `/etc/sunshine-virtual-display/rules.yaml` at runtime. Template: `config/rules.yaml.example`.

Match fields: `width`, `height`, `aspect_ratio`, `client_name`, `refresh`
Override fields: `width`, `height`, `refresh`, `gamescope_flags`, `hdr`, `disable_physical_monitors`

## CI

GitHub Actions workflows in `.github/workflows/`:
- `ci.yml` — formatting, unit tests, golangci-lint, shellcheck (runs on every PR)
- `manual-privileged.yml` — VKMS integration tests (manual trigger only)
