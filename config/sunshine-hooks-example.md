# Sunshine Hook Integration (Example)

Configure Sunshine to call these scripts:

- Pre-launch: `/usr/local/bin/session_start.sh`
- Post-session: `/usr/local/bin/session_end.sh`

If your Sunshine build supports environment variables for client mode, export:

- `SUNSHINE_CLIENT_WIDTH`
- `SUNSHINE_CLIENT_HEIGHT`
- `SUNSHINE_CLIENT_FPS`

Otherwise, implement API polling or a wrapper that infers requested mode and passes
`CLIENT_WIDTH`, `CLIENT_HEIGHT`, `CLIENT_FPS` into `session_start.sh`.
