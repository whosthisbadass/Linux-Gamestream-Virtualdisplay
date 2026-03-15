# Linux-Gamestream-Virtualdisplay Codex Instructions

This repository implements a dynamic virtual display system for Sunshine.

Primary architecture:

VKMS
+
Gamescope
+
Sunshine

Goals:

1. Create a temporary virtual monitor that matches the client resolution exactly.
2. Launch Gamescope at that resolution.
3. Stream the display through Sunshine.
4. Destroy the display when streaming ends.

Rules:

- Always prioritize exact client resolution
- Do not introduce resolution negotiation
- VKMS must be the default backend
- Gamescope is the compositor
- All scripts must be idempotent