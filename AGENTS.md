# Repository agent instructions

This repository uses Codex project configuration under `.codex/`.

## Required context for AI contributors

Before making changes, read:

1. `AI-COLLAB.md` — shared memory and work log for all AI agents; read this first
2. `.codex/instructions.md`
3. `.codex/config.toml`
4. `.codex/environment.toml`
5. `.codex/setup.sh`
6. `.codex/tasks/implement-vkms-display.md`

## Project guardrails

- Preserve exact Sunshine client request behavior.
- Keep VKMS as the default backend.
- Keep Gamescope as compositor.
- Keep privileged VKMS tests out of default CI.

## After completing work

Update `AI-COLLAB.md`:
- Append a dated entry to the Work Log describing what was changed and why.
- Update the Active Work section.
- Remove completed items from the Backlog.
- Leave a message for the other agent if relevant.
