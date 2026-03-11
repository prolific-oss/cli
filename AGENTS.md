# AGENTS.md

> Instructions for AI coding agents working on the Prolific CLI.

## Project Overview

Go CLI for the Prolific research platform — used by researchers to manage studies, participants, and data collection via the Prolific API. Module: `github.com/prolific-oss/cli`. Binary: `prolific`. Go 1.26.

**Stack**: Cobra (CLI framework), Bubbletea (TUI), Viper (configuration). See `DEVELOPMENT.md` for the full development guide.

**Key commands:**

```bash
make build          # Build binary
make test           # Run tests (always run after changes)
make lint           # Run linter (always run before committing)
make test-gen-mock  # Regenerate mocks (required after interface changes)
make all            # Full workflow: clean, install, build, test
```

Tests run entirely with mocks. `PROLIFIC_TOKEN` is **not** needed for tests.

## Architecture

```
cmd/                 Cobra commands (one package per API resource, e.g. study/, workspace/, project/)
  ├── {resource}/    Each resource gets its own package with {action}.go files
  ├── shared/        Shared command utilities
  └── root.go        Main CLI initialization and command registration
client/              API interface and HTTP client
  ├── client.go      API interface (50+ methods)
  ├── payloads.go    Request payload structs
  └── responses.go   Response types
model/               Domain models (one package per entity)
ui/                  Rendering layer
  ├── ui.go          Common helpers (headings, money, dates, counters)
  └── {resource}/    Per-resource renderers
mock_client/         Generated mocks (do not edit manually)
version/             Build-time version info
docs/examples/       Study templates (JSON/YAML)
```

## Code Patterns

### Command Signature

All commands use dependency injection:

```go
func New{Action}Command(client client.API, w io.Writer) *cobra.Command
```

### Options Struct

```go
type {Action}Options struct {
    // Flag bindings
}
```

### File Naming

- `cmd/{resource}/{action}.go` — command implementation
- `cmd/{resource}/{action}_test.go` — tests

### UI Rendering

List commands use a strategy pattern with multiple renderers:

- `InteractiveRenderer` — Bubbletea TUI with search/navigation
- `NonInteractiveRenderer` — plain table output
- `CsvRenderer` — machine-readable CSV

List commands **must** support the `-n` (non-interactive) flag for scripting.

## Testing

All tests use gomock with `mock_client.NewMockAPI`. Prefer table-driven tests.

### Canonical Test Pattern

```go
func TestCommand(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()
    c := mock_client.NewMockAPI(ctrl)

    c.EXPECT().GetWorkspaces(
        client.DefaultRecordLimit,
        client.DefaultRecordOffset,
    ).Return(&response, nil).AnyTimes()

    var b bytes.Buffer
    writer := bufio.NewWriter(&b)

    cmd := workspace.NewListCommand("workspaces", c, writer)
    _ = cmd.RunE(cmd, nil)

    writer.Flush() // CRITICAL: must flush before assertions

    actual := b.String()
    // assertions against actual
}
```

### After Changing the API Interface

If you modify `client/client.go` (add/change/remove methods), you **must** run:

```bash
make test-gen-mock
```

This regenerates `mock_client/mock_client.go`. Never edit that file manually.

## Code Style & Linting

- **Formatting**: `gofmt` + `goimports` (enforced)
- **Linters**: see `.golangci.yml` for the full list (includes gosec, govet, staticcheck, and others)
- **gosec note**: custom G101 pattern excludes "cred" — this is a domain term (CredentialPool), not a credential leak

## Commit Messages

Conventional commits enforced by git hook and CI:

```
<type>(<scope>): <description>
```

- **Types**: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`, `ci`, `build`, `perf`, `style`, `revert`
- **Scope**: optional (e.g., ticket number)
- **Description**: imperative mood, present tense, no period, under 72 chars

Example: `feat(DCP-2060): remove AITB LaunchDarkly feature flag`

The pre-commit hook runs `make lint` and `make test` automatically.

## Adding a New Command

1. Create a package under `cmd/{resource}/`
2. Implement `New{Action}Command(client client.API, w io.Writer) *cobra.Command`
3. Define an options struct with cobra flag bindings
4. Write tests in `{action}_test.go` using the gomock pattern above
5. Register the command in `cmd/root.go`
6. If new API methods are needed:
   - Add to the `API` interface in `client/client.go`
   - Implement on the `Client` struct
   - Add request/response structs to `payloads.go`/`responses.go`
   - Run `make test-gen-mock`

Reference `cmd/study/` or `cmd/workspace/` for complete examples.

> **Tip:** Claude Code users can run `/cli-command-create` to automate this entire workflow.

## Boundaries

**Always do:**
- Run `make test` and `make lint` after making changes
- Run `make test-gen-mock` after modifying `client/client.go`
- Support the `-n` (non-interactive) flag on list commands
- Use dependency injection (`client.API`, `io.Writer`) in command constructors
- Call `writer.Flush()` before test assertions

**Never do:**
- Hardcode API tokens — use `PROLIFIC_TOKEN` environment variable
- Commit `.env` files or config files containing tokens
- Manually edit generated files in `mock_client/`
- Skip `make lint` or `make test` before committing

## Configuration

| Variable | Required | Default |
|---|---|---|
| `PROLIFIC_TOKEN` | Yes (runtime only, not tests) | — |
| `PROLIFIC_URL` | No | `https://api.prolific.com` |
| `PROLIFIC_DEBUG` | No | — |

Config file: `$HOME/.config/prolific-oss/prolific.yaml`

## Common Gotchas

1. **Viper init order** — config loads before commands are built; the client depends on this
2. **Interactive mode** — Bubbletea requires a terminal; always provide a `-n` flag alternative

## Related Documentation

- `DEVELOPMENT.md` — comprehensive development guide
- `CLAUDE.md` — Claude Code-specific instructions
- `.github/copilot-instructions.md` — GitHub Copilot instructions
- `docs/examples/` — study template examples (JSON/YAML)
