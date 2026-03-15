# Linux-Gamestream-Virtualdisplay

Dynamic Linux virtual display orchestration for Sunshine using **exact client request** semantics.

## Architecture

```text
Sunshine hook -> sunshine-virtual-display session-start
  -> parse Sunshine env request (width/height/fps/hdr)
  -> create backend instance (default: VKMS)
  -> discover DRM connector
  -> launch Gamescope pinned to connector + exact mode
  -> persist runtime state

session-stop / monitor cleanup
  -> stop Gamescope
  -> destroy backend instance
  -> remove state + lock
```

## Core guarantees

- Exact request handling by default (no silent negotiation).
- VKMS is the default backend.
- Gamescope is the compositor.
- Lifecycle commands are idempotent and lock-protected.

## CLI

```bash
sunshine-virtual-display session-start
sunshine-virtual-display session-stop
sunshine-virtual-display monitor
sunshine-virtual-display status
sunshine-virtual-display doctor
sunshine-virtual-display validate-env
sunshine-virtual-display print-request
sunshine-virtual-display cleanup-stale
sunshine-virtual-display config-dump
sunshine-virtual-display version
```

## Configuration

Environment variables are centralized in `internal/config` and may be overridden by optional JSON config file (`sunshine-virtual-display.json` or `SVD_CONFIG_FILE`).

Key vars:

- `SVD_BACKEND` (`vkms` default, `portal` experimental placeholder)
- `SVD_FORCE_CONNECTOR`
- `SVD_PREFER_NEWEST_VKMS_CONNECTOR`
- `SVD_DEBUG_CONNECTOR`
- `SVD_DRY_RUN`
- `SVD_GAMESCOPE_LOG_PATH`
- `SVD_GAMESCOPE_STARTUP_TIMEOUT_SEC`
- `SVD_GAMESCOPE_GENERATE_DRM_MODE`
- `SUNSHINE_VD_GAMESCOPE_TARGET`

## Codex integration

This repo is configured for Codex via `.codex/`:

- `.codex/config.toml`
- `.codex/instructions.md`
- `.codex/environment.toml`
- `.codex/setup.sh`
- `.codex/tasks/implement-vkms-display.md`

Also see root `AGENTS.md` and `CONTRIBUTING.md`.

## CI

CI is split into separate jobs for:

- formatting
- unit tests
- Go lint
- shell lint

Privileged VKMS integration remains manual/self-hosted.

## Install / uninstall

```bash
sudo ./scripts/install.sh
sudo ./scripts/uninstall.sh
```

Use `SVD_SKIP_INSTALL_DEPS=1` to skip dependency installation when upgrading a host with preinstalled packages.
