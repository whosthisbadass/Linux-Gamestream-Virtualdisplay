# sunshine-virtual-display

Linux-native virtual monitor orchestration for Sunshine/Moonlight.

This project aims to replicate the *virtual monitor that matches the client device display* experience popularized by Apollo (Windows) — but as a Linux-native, modular sidecar that works with upstream Sunshine and upstream Moonlight clients.

## Why this exists

On Linux, a "virtual monitor" can mean:
- a kernel-level DRM/KMS connector (e.g., VKMS)
- a compositor-level headless output (wlroots)
- a portal-managed virtual monitor (xdg-desktop-portal ScreenCast "VIRTUAL")

Sunshine can already *advertise* client-requested resolutions/FPS and exposes the requested values to prep commands via environment variables like:
- `SUNSHINE_CLIENT_WIDTH`
- `SUNSHINE_CLIENT_HEIGHT`
- `SUNSHINE_CLIENT_FPS`
- `SUNSHINE_CLIENT_HDR`

This repo implements a consistent orchestration layer that can:
1) create an appropriate virtual display (preferred: VKMS)
2) start a dedicated compositor session at the client’s requested mode (preferred: Gamescope)
3) switch capture/primary display selection if needed
4) cleanly tear it down on session end

## Supported approaches

### Primary: VKMS + Gamescope + Sunshine capture (recommended)

- VKMS provides a software-only DRM/KMS virtual display device suitable for headless systems.
- Gamescope can “spoof a virtual screen with a desired resolution and refresh rate”.
- Sunshine captures using an existing backend (`kms` or `wlr` depending on host).

This path is designed to be “desktop-environment agnostic” because VKMS is kernel-level.

### Alternative: wlroots headless output + wlr capture

If you are on a wlroots compositor (e.g., Sway/Hyprland), you may be able to create a headless output and set its mode using compositor tools, then capture using Sunshine’s `wlr` backend.

### Experimental: xdg-desktop-portal ScreenCast “VIRTUAL” + PipeWire

For non-wlroots Wayland compositors, xdg-desktop-portal can expose a virtual monitor as a PipeWire stream.
This is currently provided as a PoC and tooling path; Sunshine does not (yet) document a PipeWire video capture backend.

## Quickstart

### Install dependencies

You need:
- Sunshine (host)
- Moonlight client (any)
- Gamescope
- PipeWire + WirePlumber (recommended on modern distros)
- libdrm tools (for debugging/modetest)
- Kernel support for VKMS (`CONFIG_DRM_VKMS`), and permission to load modules

### Build

Go:
```bash
go build ./cmd/sunshine-virtual-display
```
Rust:
```bash
cargo build --release
```

### Configure Sunshine prep commands

#### In Sunshine Web UI: Configuration → Applications → (Desktop or a dedicated "Virtual Display" app) → Command Preparations

Do:
```bash
sh -lc 'sunshine-virtual-display session-start'
```

Undo:
```bash
sh -lc 'sunshine-virtual-display session-stop'
```

#### Sunshine will set SUNSHINE_CLIENT_* variables for these commands on supported platforms.

### Architecture

Key modules:

    client-detector: input parsing (env vars, optional Sunshine log parsing)
    resolution-negotiator: converts client request into a safe, supported mode
    virtual-display-manager: creates/tears down VKMS instances and connectors
    display-switcher: selects primary output / disables physical outputs (optional)
    session-cleanup: robust teardown even on partial failure

#### Systemd service

A systemd user service is recommended so cleanup happens even if Sunshine hooks don’t run. See systemd/sunshine-virtual-display.service.
Security notes

    Sunshine kms capture requires cap_sys_admin (see Sunshine docs).
    VKMS creation and configfs operations require root privileges.
    Prefer running this as a systemd user service that uses a small, auditable root helper (or Polkit rule) rather than running the whole daemon as root.

Contributing

    Keep backends modular (VKMS, wlroots, portal)
    Add reproducible scripts under scripts/
    Include CI that at least runs unit tests and linting on every PR
    Use integration tests on a privileged Linux runner where possible
