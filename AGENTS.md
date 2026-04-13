# AGENTS.md

> Instructions for AI coding agents working on the Prolific CLI.

## Project Overview

Go CLI for the Prolific research platform — used by researchers to manage studies, participants, and data collection via the Prolific API. Module: `github.com/prolific-oss/cli`. Binary: `prolific`. Go 1.26.

**Stack**: Cobra (CLI framework), Bubbletea (TUI), Viper (configuration). See `DEVELOPMENT.md` for the full development guide.

**Key commands:**

```bash
make install        # Install deps, git hooks, and tools (run once after cloning)
make build          # Build binary
make test           # Run tests with coverage (always run after changes)
make lint           # Run linter (always run before committing)
make test-gen-mock  # Regenerate mocks (required after interface changes)
make all            # Full workflow: clean, install, build, test
```

Tests run entirely with mocks. `PROLIFIC_TOKEN` is **not** needed for tests.

## Architecture

```
cmd/                 Cobra commands (one package per API resource)
  ├── {resource}/    Each resource gets its own package
  │   ├── {resource}.go      Parent command (groups sub-commands)
  │   ├── {action}.go        Action command implementation
  │   └── {action}_test.go   Tests
  ├── shared/        Shared utilities
  └── root.go        Main CLI initialization and command registration
client/              API interface and HTTP client
  ├── client.go      API interface and Client implementation
  ├── payloads.go    Request payload structs
  └── responses.go   Response types
model/               Domain models (one package per entity)
ui/                  Rendering layer
  ├── ui.go          Common helpers (headings, money, dates, counters)
  └── {resource}/    Per-resource renderers
mock_client/         Generated mocks (do not edit manually)
version/             Build-time version info
docs/examples/       Study templates (JSON/YAML)
scripts/             Git hooks and changelog tooling
```

## Code Patterns

### Two-Level Command Structure

Resources follow a two-level hierarchy. The parent command groups sub-commands:

```go
// cmd/{resource}/{resource}.go
func New{Resource}Command(client client.API, w io.Writer) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "{resource}",
        Short: "Manage and view your {resources}",
    }
    cmd.AddCommand(
        NewListCommand("list", client, w),
        NewViewCommand(client, w),
        NewCreateCommand(client, w),
    )
    return cmd
}
```

The parent command is registered in `cmd/root.go`.

### Action Command Signature

Action commands use dependency injection. All take `client` and `w`; some also take a `commandName` string when the `Use` field is set dynamically by the caller rather than hardcoded in the function:

```go
// Command with a hardcoded Use string
func New{Action}Command(client client.API, w io.Writer) *cobra.Command

// Command where the caller supplies the Use string
func New{Action}Command(commandName string, client client.API, w io.Writer) *cobra.Command
```

Follow whichever pattern the existing commands in the same package use.

### Options Struct

Each command defines its own options struct with flag bindings:

```go
type {Action}Options struct {
    Args  []string
    // other flag-bound fields
}
```

### File Naming

- `cmd/{resource}/{resource}.go` — parent command
- `cmd/{resource}/{action}.go` — command implementation
- `cmd/{resource}/{action}_test.go` — tests

### Error Formatting in RunE

Prefix errors consistently in `RunE` using `fmt.Errorf` with `%s` and `err` directly (not `err.Error()`, which is redundant). Note this does not wrap the error for `errors.Is`/`errors.As` — that is the established pattern in this codebase:

```go
RunE: func(cmd *cobra.Command, args []string) error {
    result, err := client.SomeMethod(args[0])
    if err != nil {
        return fmt.Errorf("error: %s", err)
    }
    // ...
    return nil
}
```

### UI Rendering

List commands use a strategy pattern with multiple renderers:

- `InteractiveRenderer` — Bubbletea TUI with search/navigation
- `NonInteractiveRenderer` — plain table output
- `CsvRenderer` — machine-readable CSV

List commands **must** support the `-n` (non-interactive) flag for scripting.

### Web Flag (-W)

View commands that have a corresponding web UI page support opening in the browser:

```go
flags.BoolVarP(&opts.Web, "web", "W", false, "Open the resource in the web application")
```

Inside `RunE`, check this before making any API call:

```go
if opts.Web {
    return browser.OpenURL(resourceui.GetResourceURL(opts.Args[0]))
}
```

Uses `github.com/pkg/browser`. Add this pattern to any `view` command where a stable web URL exists.

### Shared Utilities

`cmd/shared/errors.go` provides `IsFeatureNotEnabledError(err error) bool` — use this when a command calls an API endpoint that may be behind a feature flag, to give a clear user-facing message instead of a raw API error.

## Testing

All tests use gomock with `mock_client.NewMockAPI`. Prefer table-driven tests. `github.com/stretchr/testify` is available for assertions (`require`, `assert`).

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

- **Formatting**: `gofmt` + `goimports` (enforced). Run `make format` to apply both.
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
- `chore`, `ci`, `build`, `style` commits are excluded from the changelog

Example: `feat(DCP-2060): remove AITB LaunchDarkly feature flag`

The pre-commit hook (installed by `make install`) runs `make lint` and `make test` automatically.

## Adding a New Command

1. Create a package under `cmd/{resource}/`
2. If a parent grouping command doesn't exist for the resource:
  - Create `cmd/{resource}/{resource}.go`
  - Implement `New{Resource}Command(client client.API, w io.Writer) *cobra.Command`
3. Implement the action as `New{Action}Command(client client.API, w io.Writer) *cobra.Command` at `cmd/{resource}/{action}.go`
4. Define an options struct with cobra flag bindings per action
5. Write tests in `{action}_test.go` using the gomock pattern above
6. Register the parent command in `cmd/root.go`
7. If new API methods are needed:
   - Add to the `API` interface in `client/client.go`
   - Implement on the `Client` struct
   - Add request/response structs to `payloads.go`/`responses.go`
   - Run `make test-gen-mock`

Reference `cmd/study/` or `cmd/workspace/` for complete examples.

> **Tip:** Claude Code users can run `/cli-command-create` to automate this entire workflow.

## Releasing

```bash
make changelog VERSION=x.y.z   # Generate grouped CHANGELOG.md entry for the release
```

This uses `git-cliff` (must be installed separately: `brew install git-cliff`) plus a custom Go script in `scripts/changelog/` that groups entries by subcommand area. After running:

1. Review and commit the updated `CHANGELOG.md`
2. Open a PR, add the `release` label, and merge to `main`
3. The `create-release.yml` workflow runs on the push to `main`. It creates the GitHub Release with tag and title `vx.y.z` (e.g. `v1.0.1`, never a bare `1.0.1`) and uploads binaries **only** when the merged PR had the `release` label; ordinary pushes to `main` do not release

To include hand-written notes in the next release, add them under `## next` in `CHANGELOG.md` before running `make changelog` — they will be merged in automatically.

## Boundaries

**Always do:**
- Run `make test` and `make lint` after making changes
- Run `make test-gen-mock` after modifying `client/client.go`
- Support the `-n` (non-interactive) flag on list commands
- Use dependency injection (`client.API`, `io.Writer`) in command constructors
- Call `writer.Flush()` before test assertions
- Prefix errors in `RunE` with `fmt.Errorf("error: %s", err)`

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

The config file is optional — the CLI works with environment variables alone.

## API Client Defaults

Defined in `client/client.go`:

- `DefaultRecordOffset = 0`
- `DefaultRecordLimit = 200`

Use these constants (not magic numbers) when calling paginated API methods in commands and tests.

## Common Gotchas

1. **Viper init order** — config loads in `cmd/root.go:initConfig()` before commands are built; the `client.New()` constructor depends on Viper being initialized
2. **Interactive mode** — Bubbletea requires a real terminal; always provide a `-n` flag alternative and test only the non-interactive path
3. **Writer flush** — commands write through a `bufio.Writer`; forgetting `writer.Flush()` before asserting in tests produces empty output
4. **Mock regeneration** — changing any method signature in the `client.API` interface without running `make test-gen-mock` causes compile errors across the test suite
5. **Git hooks** — hooks are installed by `make install`, not automatically on clone; contributors who skip `make install` won't have lint/test run pre-commit

## Related Documentation

- `CONTRIBUTING.md` — Contributor guidelines and PR process
- `DEVELOPMENT.md` — comprehensive development guide
- `CLAUDE.md` — Claude Code-specific configuration (slash commands); points to this file for all coding instructions
- `.github/copilot-instructions.md` — GitHub Copilot-specific configuration (changelog/release instructions); points to this file for all coding instructions
- `docs/examples/` — study template examples (JSON/YAML)
