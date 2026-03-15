# Repository agent instructions

This repository uses Codex project configuration under `.codex/`.

## Required context for AI contributors

Before making changes, read:

1. `.codex/instructions.md`
2. `.codex/config.toml`
3. `.codex/environment.toml`
4. `.codex/setup.sh`
5. `.codex/tasks/implement-vkms-display.md`

## Project guardrails

- Preserve exact Sunshine client request behavior.
- Keep VKMS as the default backend.
- Keep Gamescope as compositor.
- Keep privileged VKMS tests out of default CI.
