# Bangumi Terminal UI

Bangumi CLI and TUI in Golang

## Commands

- `auth`
  Auth commands
- `completion`
  Generate the autocompletion script for the specified shell
- `help`
  Help about any command
- `list`
  List collection
- `sub`
  Subject/Collection actions
- `version`
  Print the version number of bgm-cli
- `ui`
  Start terminal UI

## Development

Lint

```bash
gofumpt -w .
golangci-lint run
```

CLI

`go run . help`

UI

`go run . ui`
