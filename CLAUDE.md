# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Summary

Go application that orchestrates a dynamic virtual display for Sunshine game streaming on Linux. Creates a temporary virtual monitor (via VKMS) at the exact client-requested resolution, launches Gamescope compositor, streams through Sunshine, then tears everything down cleanly on session end.

Stack: **VKMS → Gamescope → Sunshine**

## Build & Test

```bash
# Build (output binary: sunshine-virtual-display)
go build -o sunshine-virtual-display ./cmd/sunshine-virtual-display

# All unit tests
go test ./...

# Single test
go test ./internal/<package> -run TestFunctionName -v

# Format check
gofmt -l .

# Lint
golangci-lint run ./...

# Shell lint
shellcheck scripts/*.sh
```

> **Note:** This project targets Linux only. Build and tests compile on any platform (no cgo, no Linux syscalls in test paths), but the binary only runs meaningfully on a Linux host with VKMS, Gamescope, and Sunshine installed.

The privileged VKMS integration test (`TestCreateDestroyPrivileged` in `internal/vkms`) requires a real kernel module and `SVD_PRIVILEGED_TESTS=1` — never add it to default CI.

## Architecture

**Data flow:** Sunshine hook → `session-start` command → `clientdetector` parses `SUNSHINE_CLIENT_*` env vars → `rules` engine optionally overrides display config → `vkms` creates a configfs instance in `/sys/kernel/config/vkms` → `vkms/discovery` finds the new DRM connector in `/sys/class/drm` → `gamescope` launcher starts the compositor → `cleanup` persists session state to `${XDG_RUNTIME_DIR:-/tmp}/sunshine-virtual-display/session-state.json` → on disconnect, `session-stop` reads state, kills Gamescope, and destroys the VKMS instance.

**Key interfaces:**
- `backend.Backend` — Create/Destroy VKMS instances (VKMS default; Portal is a non-functional placeholder)
- `lifecycle.Controller` — owns session-start, session-stop, and the monitor loop
- `rules.Apply()` — takes a `ClientRequest`, returns a `DisplayConfig` with any overrides applied

**Session state & locking:** A JSON state file and a lock file (both under `XDG_RUNTIME_DIR`) prevent concurrent sessions. Lock uses `O_EXCL` for atomicity.

## Repository Layout

```
cmd/sunshine-virtual-display/   main entrypoint (13 CLI commands)
internal/
  backend/         backend interface (VKMS default, portal placeholder)
  cleanup/         state persistence and lock management
  clientdetector/  parses Sunshine env vars into a request object
  config/          all env var config, optional JSON config file
  display/         display config types
  gamescope/       Gamescope launcher
  lifecycle/       session-start / session-stop controller
  rules/           optional YAML override rule engine
  vkms/            VKMS DRM connector discovery and manager
config/            example rules.yaml and Sunshine hook config
docs/              extended documentation
scripts/           install.sh / uninstall.sh and session wrappers
.github/workflows/ ci.yml (default) and manual-privileged.yml
```

## Architecture Constraints (do not violate)

- **VKMS is the default backend** — do not change the default or add negotiation
- **Gamescope is the compositor** — do not swap it out
- **Exact client resolution** — no silent resolution negotiation; if a rule doesn't match, use the raw client values
- **Idempotent lifecycle** — session-start and session-stop must be safe to call multiple times
- **Privileged VKMS integration tests** must NOT be added to default CI

## Key Config Env Vars

| Variable | Default | Purpose |
|---|---|---|
| `SVD_BACKEND` | `vkms` | Backend driver |
| `SVD_FORCE_CONNECTOR` | — | Override DRM connector |
| `SVD_PREFER_NEWEST_VKMS_CONNECTOR` | — | Connector selection hint |
| `SVD_DRY_RUN` | — | Skip actual system calls |
| `SVD_GAMESCOPE_LOG_PATH` | — | Gamescope log file |
| `SVD_GAMESCOPE_STARTUP_TIMEOUT_SEC` | `10` | Gamescope startup wait |
| `SVD_GAMESCOPE_GENERATE_DRM_MODE` | `cvt` | DRM mode generation method |
| `SVD_MONITOR_INTERVAL_SEC` | `5` | Monitor daemon poll interval |
| `SVD_MONITOR_MAX_RUNTIME_SEC` | `0` (infinite) | Monitor daemon max lifetime |
| `SVD_CONFIG_FILE` | `sunshine-virtual-display.json` (CWD) | Optional JSON config file path |

Sunshine client inputs read from env:
- `SUNSHINE_CLIENT_WIDTH`, `SUNSHINE_CLIENT_HEIGHT`, `SUNSHINE_CLIENT_FPS`
- `SUNSHINE_CLIENT_NAME`, `SUNSHINE_CLIENT_HDR`
- `SUNSHINE_VD_GAMESCOPE_TARGET` — command passed to Gamescope as the target (via `sh -lc`)

## Optional Override Rules

Loaded from `/etc/sunshine-virtual-display/rules.yaml` at runtime. Template: `config/rules.yaml.example`.

Match fields: `width`, `height`, `aspect_ratio`, `client_name`, `refresh`
Override fields: `width`, `height`, `refresh`, `gamescope_flags`, `hdr`, `disable_physical_monitors`

## CI

GitHub Actions workflows in `.github/workflows/`:
- `ci.yml` — formatting, unit tests, golangci-lint, shellcheck (runs on every PR)
- `manual-privileged.yml` — VKMS integration tests (manual trigger only, requires `SVD_PRIVILEGED_TESTS=1`)
