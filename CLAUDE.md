# Prolific CLI

Go CLI for the Prolific research platform — used by researchers to manage studies, participants, and data collection via the Prolific API. Module: `github.com/prolific-oss/cli`. Binary: `prolific`.

**Stack**: Cobra (CLI framework), Bubbletea (TUI), Viper (configuration).

## Environment

| Variable | Required | Default |
|---|---|---|
| `PROLIFIC_TOKEN` | Yes (runtime only, not tests) | — |
| `PROLIFIC_URL` | No | `https://api.prolific.com` |
| `PROLIFIC_DEBUG` | No | — |

Config file: `$HOME/.config/prolific-oss/prolific.yaml`

Tests run entirely with mocks — `PROLIFIC_TOKEN` is **not** needed for tests.

## Commands

```bash
make build          # Build binary
make test           # Run tests (always run after changes)
make lint           # Run linter (always run before committing)
make test-gen-mock  # Regenerate mocks (required after interface changes)
make all            # Full workflow: clean, install, build, test
```

## Architecture

- `cmd/` - Cobra commands (one package per resource)
- `client/client.go` - API interface and HTTP client
- `model/` - Domain models
- `ui/` - Rendering (interactive/non-interactive/CSV)
- `mock_client/` - Generated mocks

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

## Gotchas

1. **Mock regeneration** — After changing `client/client.go` interface, run `make test-gen-mock`
2. **Test output** — Must call `writer.Flush()` before assertions
3. **List commands** — Always support `-n` (non-interactive) flag for scripting
4. **Config loading** — Viper initializes before commands; client depends on this
5. **Interactive mode** — Bubbletea requires a terminal; always provide a `-n` flag alternative

## Reference

- `DEVELOPMENT.md` — comprehensive development guide with full directory structure, UI rendering helpers, study templates, and CI/CD workflows
- `.github/copilot-instructions.md` — GitHub Copilot-specific instructions
- `docs/examples/` — study template examples (JSON/YAML)
