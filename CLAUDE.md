# Prolific CLI

Go CLI for the Prolific research platform. Uses Cobra (CLI), Bubbletea (TUI), Viper (config).

## Environment

- `PROLIFIC_TOKEN` - Required for CLI usage (not needed for tests - they use mocks)
- Tests run entirely with mocks, no API access needed

## Commands

```bash
make build          # Build binary
make test           # Run tests (always run after changes)
make lint           # Run linter
make test-gen-mock  # Regenerate mocks (required after interface changes)
make all            # Full workflow: clean, install, build, test
```

## Architecture

- `cmd/` - Cobra commands (one package per resource)
- `client/client.go` - API interface and HTTP client
- `model/` - Domain models
- `ui/` - Rendering (interactive/non-interactive/CSV)
- `mock_client/` - Generated mocks

## Adding Commands

Commands follow dependency injection pattern:

```go
func NewListCommand(client client.API, w io.Writer) *cobra.Command
```

Reference existing commands in `cmd/study/` or `cmd/workspace/` for patterns.

## Gotchas

1. **Mock regeneration** - After changing `client/client.go` interface, run `make test-gen-mock`
2. **Test output** - Must call `writer.Flush()` before assertions (see `cmd/workspace/list_test.go:70-87`)
3. **List commands** - Always support `-n` (non-interactive) flag for scripting
4. **Config loading** - Viper initializes before commands; client depends on this

## Testing Pattern

```go
ctrl := gomock.NewController(t)
defer ctrl.Finish()
c := mock_client.NewMockAPI(ctrl)
c.EXPECT().Method().Return(response, nil)

var b bytes.Buffer
w := bufio.NewWriter(&b)
cmd := NewCommand(c, w)
_ = cmd.RunE(cmd, nil)
w.Flush()  // Required before assertions
```

## Reference

Read `DEVELOPMENT.md` when implementing new commands or debugging test patterns. It covers:

- Full directory structure and file naming
- UI rendering helpers (`ui/ui.go`)
- Study templates (`docs/examples/`)
- CI/CD workflows
