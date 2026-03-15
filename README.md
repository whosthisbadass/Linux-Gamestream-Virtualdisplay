# Linux-Gamestream-Virtualdisplay

Apollo-style dynamic virtual display orchestration for Sunshine on Linux.

## Core behavior

On `session-start`, this project reads Sunshine's client request and uses it **exactly as requested**:

- `SUNSHINE_CLIENT_WIDTH`
- `SUNSHINE_CLIENT_HEIGHT`
- `SUNSHINE_CLIENT_FPS`
- `SUNSHINE_CLIENT_HDR`

If the client asks for `2560x1600@120`, the stack attempts to create exactly `2560x1600@120` with no negotiation logic in this project.
Fallback is only possible if kernel/compositor rejects the mode.

## Stack

- VKMS (virtual monitor)
- Gamescope (compositor at exact mode)
- Sunshine (streaming server)
- Optional experimental path: PipeWire Portal Virtual Display

## Repository layout

```text
Linux-Gamestream-Virtualdisplay
├─ cmd/
│  └─ sunshine-virtual-display/
├─ internal/
│  ├─ clientdetector/
│  ├─ vkms/
│  ├─ gamescope/
│  ├─ lifecycle/
│  └─ cleanup/
├─ scripts/
│  ├─ install.sh
│  ├─ vkms-create.sh
│  ├─ vkms-destroy.sh
│  └─ run-gamescope.sh
├─ systemd/
│  └─ sunshine-virtual-display.service
├─ docs/
└─ README.md
```

## Build

```bash
go build -o sunshine-virtual-display ./cmd/sunshine-virtual-display
```

## CLI

```bash
sunshine-virtual-display session-start
sunshine-virtual-display session-stop
```

### session-start

1. Parse Sunshine env vars into `ClientDisplayRequest`.
2. Create dynamic VKMS instance under `/sys/kernel/config/vkms/<instance>`.
3. Build and link VKMS pipeline objects.
4. Apply exact mode string `WIDTHxHEIGHT@FPS` and enable connector.
5. Launch Gamescope with:

```bash
gamescope -W <width> -H <height> -r <fps>
```

6. Save runtime state for teardown.

### session-stop

1. Stop Gamescope.
2. Destroy VKMS instance.
3. Remove runtime state.

## Install

```bash
sudo ./scripts/install.sh
```

The installer:
- detects distro (Debian/Ubuntu, Fedora, Arch)
- installs dependencies
- builds and installs binary
- installs/enables systemd service
- runs `modprobe vkms`

## Sunshine hook example

Use prep commands:

- Do: `sh -lc 'sunshine-virtual-display session-start'`
- Undo: `sh -lc 'sunshine-virtual-display session-stop'`

## Validation scripts

- `scripts/test-session-start.sh`
- `scripts/test-session-stop.sh`

These implement the required start/stop validation workflow.
