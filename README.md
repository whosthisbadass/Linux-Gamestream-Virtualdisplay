# Linux-Gamestream-Virtualdisplay

Dynamic Linux virtual display orchestration for Sunshine using **exact client request** semantics.

## Architecture

```text
Sunshine hook -> sunshine-virtual-display session-start
  -> parse Sunshine env request (width/height/fps/hdr)
  -> apply optional override rules from /etc/sunshine-virtual-display/rules.yaml
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
- Override rules are optional and applied only when matched.
- VKMS is the default backend.
- Gamescope is the compositor.
- Lifecycle commands are idempotent and lock-protected.

## Dynamic Client Detection

Client detection reads Sunshine client environment variables and builds a dynamic request object at runtime:

- `SUNSHINE_CLIENT_WIDTH`
- `SUNSHINE_CLIENT_HEIGHT`
- `SUNSHINE_CLIENT_FPS`
- `SUNSHINE_CLIENT_NAME`
- `SUNSHINE_CLIENT_UID`
- `SUNSHINE_CLIENT_HDR` (optional)

Default flow is always dynamic:

```text
client connects
↓
Sunshine provides resolution / refresh
↓
virtual display created with those exact values
↓
gamescope launched
```

No static device profile is required for Steam Deck, phones, tablets, laptops, TVs, or ultrawide monitors.

## Optional Override Rules

Rules are loaded from:

- `/etc/sunshine-virtual-display/rules.yaml`

If no rules match, the client request is used exactly.

Example:

```yaml
rules:
  - match:
      aspect_ratio: "16:10"
      width: 1280
    override:
      refresh: 90

  - match:
      width: 3840
    override:
      hdr: true
```

Supported match fields:

- `width`
- `height`
- `aspect_ratio`
- `client_name`
- `refresh`

Supported override fields:

- `width`
- `height`
- `refresh`
- `gamescope_flags`
- `hdr`
- `disable_physical_monitors`

A template is provided in `config/rules.yaml.example`.

## CLI

```bash
sunshine-virtual-display session-start
sunshine-virtual-display session-stop
sunshine-virtual-display monitor
sunshine-virtual-display status
sunshine-virtual-display doctor
sunshine-virtual-display validate-env
sunshine-virtual-display print-request
sunshine-virtual-display detect-client
sunshine-virtual-display show-config
sunshine-virtual-display show-rules
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
- unit tests (`go test ./...`, including rule engine and dynamic client detection tests)
- Go lint (`golangci-lint`)
- shell lint

Privileged VKMS integration remains manual/self-hosted.

## Install / uninstall

```bash
sudo ./scripts/install.sh
sudo ./scripts/uninstall.sh
```

Use `SVD_SKIP_INSTALL_DEPS=1` to skip dependency installation when upgrading a host with preinstalled packages.
