# Design Notes

## Lifecycle

### session-start

- Input source is Sunshine environment variables.
- No negotiation layer exists in code.
- VKMS instance is created dynamically in configfs using kernel-documented object links.
- Gamescope starts with exact width/height/refresh on the discovered VKMS connector (`-O`).
- State is persisted to `${XDG_RUNTIME_DIR:-/tmp}/sunshine-virtual-display/session-state.json`.

### session-stop

- Gamescope PID from state is terminated using SIGTERM with SIGKILL fallback.
- VKMS instance is removed from `/sys/kernel/config/vkms`.
- State file is deleted.

## Error handling

- Missing required env vars fails fast.
- If Gamescope start fails, VKMS instance is cleaned up.
- If state write fails, both VKMS and Gamescope are cleaned up.

## Experimental path

PipeWire portal virtual display can be added as an alternative backend later, but VKMS+Gamescope is default.
