# Sunshine Hook Integration (Example)

Configure Sunshine to call the compiled helper directly:

- Pre-launch: `sunshine-virtual-display session-start`
- Post-session: `sunshine-virtual-display session-stop`

If your Sunshine build supports environment variables for client mode, export:

- `SUNSHINE_CLIENT_WIDTH`
- `SUNSHINE_CLIENT_HEIGHT`
- `SUNSHINE_CLIENT_FPS`

Otherwise, implement API polling or a wrapper that infers requested mode and exports
`SUNSHINE_CLIENT_WIDTH`, `SUNSHINE_CLIENT_HEIGHT`, `SUNSHINE_CLIENT_FPS`, and
`SUNSHINE_CLIENT_HDR` before invoking `sunshine-virtual-display session-start`.
