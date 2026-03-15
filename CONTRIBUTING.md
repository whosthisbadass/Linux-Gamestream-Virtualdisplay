# Contributing

## Development setup

- Read `.codex/instructions.md` for architecture constraints.
- Optional: run `.codex/setup.sh` in a disposable dev container.
- Build: `go build ./cmd/sunshine-virtual-display`

## Validation

Run before opening a PR:

- `gofmt -l .`
- `go test ./...`
- `golangci-lint run ./...`
- `shellcheck scripts/*.sh`

## Codex usage

Codex-specific project defaults live in `.codex/`.
Use `config-dump`, `doctor`, and `validate-env` commands to troubleshoot local sessions.
