# sunshine-virtual-display: Apollo-Style Virtual Monitor for Linux + Sunshine

This document designs a Linux-native virtual display workflow that reproduces Apollo-style behavior while keeping **Sunshine** as the streaming server.

---

## 1) How Apollo Does This on Windows (Reference Behavior)

Apollo’s Windows flow relies on a virtual display driver stack (commonly SudoVDA-based) and Sunshine session orchestration:

1. **Virtual display creation (SudoVDA / virtual display driver)**
   - Apollo installs/controls a virtual display adapter driver.
   - On client connect, Apollo creates or enables a virtual monitor endpoint in Windows Display Manager.
   - The virtual monitor behaves like a real sink (EDID-like identity, display path, mode list).

2. **Automatic mode matching**
   - During Moonlight↔Sunshine negotiation, client mode preferences (width/height/FPS/HDR hints) are known.
   - Apollo applies matching display mode on the virtual monitor (e.g., 2560x1600@120).
   - Encoding then runs at the same mode, reducing scaling and latency artifacts.

3. **Per-client display identity**
   - Apollo can maintain per-client profile identity (device model/profile → preferred mode/HDR/preset).
   - Practical effect: reconnect from Steam Deck vs TV can select different modes automatically.

4. **Lifecycle control**
   - **Connect:** create/enable virtual display, set mode, route stream source.
   - **Disconnect:** teardown/disable display, optionally restore physical monitor and prior mode.

5. **Integration with Sunshine**
   - Apollo runs beside Sunshine and uses Sunshine session events (or API/polling) to trigger display state changes.
   - Sunshine remains the encoder/stream server; Apollo is display orchestration.

---

## 2) Linux Approaches: Comparative Analysis

### A. Kernel/DRM Path — VKMS

**What it is:** `vkms` is a software KMS driver exposing virtual DRM connectors/CRTC/planes without physical hardware.

**Pros**
- Native DRM object model (connectors/modesets) that Linux desktop stacks understand.
- Works headless and can exist without any physical monitor attached.
- Good for testing and virtual output plumbing.

**Constraints / caveats**
- **Dynamic connector creation:** limited compared to purpose-built virtual GPU drivers; typically fixed virtual connectors once module is loaded.
- **GPU encoder compatibility:** Sunshine generally performs best when capturing a compositor/game surface rendered by the real GPU. VKMS itself is software scanout, so direct “zero-copy GPU → encoder” paths are not guaranteed.
- **Performance:** VKMS scanout is CPU-side emulation; acceptable for control plane and light usage, but can bottleneck if treated as full rendering target.
- **Sunshine compatibility:** viable if Sunshine captures from compositor/portal/PipeWire source linked to real rendering pipeline, not relying on VKMS as high-performance renderer.

### B. Wayland Approaches

#### wlroots virtual outputs (Sway, Hyprland ecosystems)

**Pros**
- Compositor-native virtual outputs are often easier to create/destroy dynamically than kernel connectors.
- Better alignment with modern desktop capture via PipeWire portals.
- Can expose per-output mode semantics in compositor config/API.

**Caveats**
- API stability varies by compositor and version.
- Automation differs (Sway IPC vs Hyprland commands/plugins).
- HDR and high-refresh handling still compositor-dependent.

#### Wayland headless backends

**Pros**
- Great for offscreen rendering and CI-style sessions.
- Reliable in no-monitor environments.

**Caveats**
- Some capture workflows expect “real” outputs; headless surfaces may require custom PipeWire nodes or compositor support.
- Sunshine capture source selection may need explicit configuration.

#### PipeWire integration

- PipeWire is the common denominator for modern Linux screen capture.
- Best results come from compositors that provide stable DMA-BUF export paths to minimize copies.
- For streaming, PipeWire node discovery + stable node naming are critical.

### C. X11 Approaches

#### Xvfb
- Pure software framebuffer; no real DRM/KMS output, no modern GPU display pipeline.
- Useful for testing GUI automation, not ideal for low-latency game streaming.

#### Xdummy
- Better than Xvfb for some workflows but still mostly synthetic Xorg behavior.
- Often fragile under modern mixed Wayland/X11 environments.

#### xrandr virtual outputs
- Depends on Xorg driver capabilities (e.g., VirtualHeads) and is not portable across modern systems.
- Weak fit for Wayland-native desktops and PipeWire-centric capture.

**Conclusion:** X11-only virtual display methods are generally poor fits for modern Sunshine + Moonlight performance targets.

---

## 3) Explicit Option: VKMS + Gamescope + PipeWire

A strong Linux-native design:

- **VKMS:** provides a guaranteed virtual KMS output in headless/no-monitor scenarios.
- **Gamescope:** lightweight compositor to host the game/session at requested mode.
- **PipeWire:** exposes capture stream node to Sunshine.

### Data flow

```text
Moonlight connects
   ↓
Sunshine session metadata available (mode/FPS/client)
   ↓
Virtual Display Manager loads vkms (if needed)
   ↓
Gamescope launches at WxH@R on virtual output
   ↓
PipeWire exports capture node for gamescope output
   ↓
Sunshine captures that node and encodes to client
```

### Why this can replicate Apollo behavior

- Create on-demand virtual target without physical monitor.
- Set mode per session dynamically.
- Keep Sunshine as the server while externalizing display orchestration.
- Works for handheld/TV/laptop clients with differing mode profiles.

### Advantages
- Fully Linux-native stack.
- Works in headless hosts (homelab/server).
- Wayland-compatible capture pipeline.
- Good fit for Steam Deck-like clients and varying aspect ratios.

---

## 4) System Architecture: `sunshine-virtual-display`

### Components

1. **client-detector**
   - Detects Sunshine session start/stop.
   - Sources: hook env vars, local API polling, log/event parsing.

2. **resolution-negotiator**
   - Parses requested client mode (width/height/refresh/HDR).
   - Applies policy (limits, supported mode fallback, profile overrides).

3. **virtual-display-manager**
   - Ensures vkms/compositor resources exist.
   - Creates/enables virtual output and applies target mode.

4. **display-switcher**
   - Optional: disable DPMS or logically deactivate physical outputs.
   - Sets virtual output as primary capture target.

5. **session-cleanup**
   - Triggered on disconnect/timeout.
   - Restores previous display state and destroys virtual resources.

### Interaction diagram

```text
Moonlight client connects
      ↓
Sunshine launches hook (or API event observed)
      ↓
client-detector
      ↓
resolution-negotiator (parse 2560x1600@120)
      ↓
virtual-display-manager (vkms + gamescope + pipewire target)
      ↓
display-switcher (optional physical monitor off)
      ↓
Sunshine captures/encodes virtual display
      ↓
Client disconnects
      ↓
session-cleanup (destroy virtual display + restore state)
```

---

## 5) Resolution Matching Strategy

Primary preference order:

1. **Sunshine-provided launch environment variables** (if available)
2. **Sunshine local API session data**
3. **Moonlight negotiation parameters via wrapper/hook**
4. **Fallback profile defaults** (per client or global)

### Negotiation logic

```text
requested = client_mode || profile_default || global_default
validated = clamp_to_host_caps(requested)
create_virtual_display(width, height, refresh, hdr)
```

Example:

- `client_resolution=2560x1600`
- `client_refresh=120`

Result:

- create virtual target as `2560x1600@120`
- launch gamescope at exactly that mode
- instruct Sunshine to capture that output/node

---

## 6) Proof-of-Concept Scripts

See scripts in this repo:

- `scripts/create_virtual_display.sh`
- `scripts/destroy_virtual_display.sh`
- `scripts/session_start.sh`
- `scripts/session_end.sh`

They demonstrate:
- vkms load
- mode parsing
- gamescope launch template
- cleanup lifecycle

---

## 7) Sunshine Integration

Recommended integrations:

1. **Pre-launch hook**
   - `session_start.sh` called before app/game launch.
   - Parses client mode and starts virtual stack.

2. **Post-session hook**
   - `session_end.sh` tears down resources.

3. **Event hook / API polling fallback**
   - If hook env vars are unavailable, daemon polls Sunshine API for active sessions.

---

## 8) Optional Enhancements

- **Per-client profiles** (Steam Deck, TV, phone) with preferred mode/HDR/bitrate.
- **HDR policy engine** with fallback to SDR if compositor/encoder path can’t guarantee HDR metadata.
- **Physical monitor power policy** via `wlr-randr`/compositor IPC/DPMS.
- **Multi-client** isolated virtual outputs (resource-heavy; requires scheduler).
- **GPU selection** for multi-GPU hosts (iGPU for desktop, dGPU for encode/game).

---

## 9) Recommended Implementation (Practical)

For today’s Linux ecosystem, the most maintainable path is:

1. Wayland compositor workflow using **Gamescope + PipeWire** for capture.
2. Use **VKMS** as headless safety net when no physical output exists.
3. Wire into Sunshine using pre/post hooks first, API daemon second.
4. Keep state files for idempotent cleanup and crash recovery.

This gives an Apollo-like lifecycle while staying fully Linux-native and Sunshine-compatible.
