# Prolific CLI Development Guide

## Project Overview

**Prolific CLI** is a command-line interface for the [Prolific](https://www.prolific.com) research platform. It's written in Go using the Cobra framework for CLI commands and Bubbletea for interactive TUI components.

- **Language**: Go 1.26+
- **Primary Frameworks**: Cobra (CLI), Bubbletea (TUI), Viper (config)
- **API Client**: Custom HTTP client in `client/client.go`
- **License**: Apache 2.0

> **New contributor?** Start with [CONTRIBUTING.md](CONTRIBUTING.md) for an overview of how to contribute, then return here for detailed development guidance.

## Prerequisites & Setup

### Installing Go

This project requires **Go 1.26 or later**.

**macOS:**

```bash
brew install go
```

**Other platforms:**
Download from <https://go.dev/dl/>

### Setting up your PATH

After installing Go, ensure `$GOPATH/bin` is in your PATH so tools like `mockgen` and `gosec` can be found:

```bash
# Add to ~/.zshrc (zsh) or ~/.bash_profile (bash)
export PATH="$PATH:$(go env GOPATH)/bin"

# Reload your shell
source ~/.zshrc  # or source ~/.bash_profile
```

Verify setup:

```bash
go version           # Should show Go 1.26+
go env GOPATH        # Should show your Go workspace (typically ~/go)
echo $PATH | grep go # Should include your GOPATH/bin
```

### First Time Setup

Once Go is installed and PATH is configured:

```bash
# Install dependencies and setup git hooks
make install

# Verify mockgen is available
mockgen -version
```

## Essential Commands

### Development Workflow

```bash
# Install dependencies and setup git hooks
make install

# Build the binary
make build

# Build static binary with version info
make static

# Run tests (default - with coverage output)
make test

# Run tests with HTML coverage report
make test-cov

# Run linter
make lint

# Lint Dockerfile
make lint-dockerfile

# Full workflow (clean, install, build, test)
make all

# Generate mocks for testing
make test-gen-mock

# Run everything with static build
make static-all
```

### Docker

```bash
# Build Docker image
make docker-build

# Push Docker image
make docker-push

# Security scan
make docker-scout
```

### Running the CLI

```bash
# After building
./prolific --help

# List studies
./prolific study list

# View a specific study
./prolific study view <study-id>

# Create a study from template
./prolific study create -t docs/examples/standard-sample.yaml

# Get user account details
./prolific whoami
```

## Configuration

### Environment Variables

**Required:**

- `PROLIFIC_TOKEN` - API token from Prolific (get from <https://app.prolific.com/researcher/tokens/>)

**Optional:**

- `PROLIFIC_URL` - Override API URL (defaults to `https://api.prolific.com`)
- `PROLIFIC_DEBUG` - Enable debug output for API requests

### Config File

Location: `$HOME/.config/prolific-oss/prolific.yaml`

Available settings:

```yaml
workspace: xxxxxxxxxx  # Default workspace ID for commands
```

## Code Organization

### Directory Structure

```
.
├── cmd/                      # Cobra command implementations
│   ├── root.go              # Root command and app initialization
│   ├── aitaskbuilder/       # AI task builder commands
│   ├── campaign/            # Campaign management
│   ├── filters/             # Filter management
│   ├── filtersets/          # Filter set management
│   ├── hook/                # Webhook management
│   ├── message/             # Messaging commands
│   ├── participantgroup/    # Participant group management
│   ├── project/             # Project management
│   ├── requirements/        # Eligibility requirements
│   ├── study/               # Study management (create, list, view, etc.)
│   ├── submission/          # Submission management
│   ├── user/                # User account commands
│   └── workspace/           # Workspace management
├── client/                  # HTTP API client
│   ├── client.go           # Client implementation and API interface
│   ├── payloads.go         # Request payload structs
│   └── responses.go        # Response structs
├── config/                 # Configuration helpers
├── model/                  # Domain models
├── ui/                     # UI components and rendering
│   ├── ui.go              # Common UI helpers
│   ├── study/             # Study-specific UI (interactive lists, views)
│   ├── submission/        # Submission UI
│   ├── requirement/       # Requirement UI
│   └── filter/            # Filter UI
├── mock_client/           # Generated mocks for testing
├── version/               # Version information
├── docs/examples/         # Study template examples (JSON/YAML)
└── main.go               # Entry point
```

### Key Architectural Patterns

#### Command Structure

All commands follow a consistent pattern:

```go
// Each command package exports a New*Command function
func NewListCommand(commandName string, client client.API, w io.Writer) *cobra.Command {
    var opts ListOptions

    cmd := &cobra.Command{
        Use:   commandName,
        Short: "Description",
        Long:  `Detailed description`,
        Example: `Example usage`,
        RunE: func(cmd *cobra.Command, args []string) error {
            // Implementation
            return nil
        },
    }

    // Add flags
    cmd.Flags().StringVarP(&opts.Field, "flag", "f", "default", "help")

    return cmd
}
```

#### Dependency Injection

- The `client.API` interface is used throughout for API calls
- Dependency injection pattern used in `cmd/root.go:52-83`
- Commands receive: `client.API`, `io.Writer` for testability
- See `client/client.go:26-70` for the full API interface

#### Client Implementation

- Base client in `client/client.go`
- Uses Viper for configuration
- Token-based authentication
- Standardized error handling with `JSONAPIError` struct
- All requests go through `Execute()` method for consistency

## Testing Conventions

### Test Structure

- Use table-driven tests where appropriate
- Each command package has corresponding `*_test.go` files
- Mock the `client.API` interface using `golang/mock`

### Example Test Pattern

```go
func TestNewListCommand(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()
    c := mock_client.NewMockAPI(ctrl)

    // Setup expectations
    c.EXPECT().
        GetStudies(status, projectID).
        Return(&response, nil).
        AnyTimes()

    // Capture output
    var b bytes.Buffer
    writer := bufio.NewWriter(&b)

    // Execute command
    cmd := study.NewListCommand("studies", c, writer)
    _ = cmd.RunE(cmd, nil)

    writer.Flush()

    // Assert output
    expected := `...`
    if b.String() != expected {
        t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, b.String())
    }
}
```

See `cmd/workspace/list_test.go:37-88` for a complete example.

### Generating Mocks

```bash
make test-gen-mock
```

This regenerates `mock_client/mock_client.go` from `client/client.go`.

## Code Style and Conventions

### Go Standards

- Use `gofmt` for formatting (enforced by linter)
- Follow idiomatic Go practices
- Use semantic commit messages
- Prefer clear, user-friendly CLI flags and help text

### Linting

Enabled linters (`.golangci.yml`):

- dogsled, dupl, errcheck, exhaustive
- gochecknoinits, goconst, gocyclo
- goprintffuncname, gosec, govet
- ineffassign, misspell, nakedret
- noctx, nolintlint, staticcheck
- unconvert, unparam, unused, whitespace

Formatters: `gofmt`, `goimports`

### Naming Conventions

- **Commands**: Verb-based (list, create, view, transition)
- **Options structs**: `*Options` suffix (e.g., `ListOptions`, `CreateOptions`)
- **Response structs**: `*Response` suffix (e.g., `ListStudiesResponse`)
- **Constants**: PascalCase with descriptive prefixes (e.g., `StatusActive`, `TransitionStudyPublish`)
- **Interfaces**: Descriptive names (e.g., `API`, `ListStrategy`)

### File Naming

- Command implementation: `<action>.go` (e.g., `list.go`, `create.go`, `view.go`)
- Tests: `<action>_test.go`
- Command group root: `<resource>.go` (e.g., `study.go`, `workspace.go`)

### UI Rendering

Common helpers in `ui/ui.go`:

- `RenderSectionMarker()` - Visual separator
- `RenderHeading(heading)` - Bold headings
- `RenderMoney(amount, currencyCode)` - Currency formatting
- `RenderRecordCounter(count, total)` - "Showing X records of Y"
- `RenderApplicationLink(entity, slug)` - Links to web app

Date format: `AppDateTimeFormat = "02-01-2006 15:04"`

### Interactive vs Non-Interactive Commands

Many list commands support multiple rendering strategies:

- **Interactive**: Bubbletea TUI with search/navigation
- **Non-interactive**: Table output to stdout
- **CSV**: Machine-readable format

Example from `cmd/study/list.go:89-99`:

```go
renderer := study.ListRenderer{}
if opts.Csv {
    renderer.SetStrategy(&study.CsvRenderer{})
} else if opts.NonInteractive {
    renderer.SetStrategy(&study.NonInteractiveRenderer{})
} else {
    renderer.SetStrategy(&study.InteractiveRenderer{})
}
```

## Model Layer

Key models in `model/`:

- `Study` - Research study with eligibility requirements, filters
- `Workspace` - Organizational workspace
- `Project` - Project within a workspace
- `Requirement` - Eligibility requirement
- `FilterSet` - Reusable filter sets
- `Submission` - Study submission from participants
- `Campaign` - Bring your own participants campaigns
- `Hook` - Webhook subscriptions

Study statuses (see `model/study.go:8-38`):

- `unpublished`, `active`, `scheduled`, `awaiting review`, `completed`

Study transitions:

- `PUBLISH`, `START`, `PAUSE`, `STOP`

## API Client Defaults

From `client/client.go`:

- `DefaultRecordOffset = 0`
- `DefaultRecordLimit = 200`

All API methods follow pattern: `Get*`, `Create*`, `Update*`, `Transition*`

## Study Creation Templates

Template examples in `docs/examples/`:

- `standard-sample.yaml` / `.json` - Basic study
- `minimal-study.json` - Minimal required fields
- `study-with-ethnicity-screener.json` - With eligibility requirements
- `study-with-filter-set.json` - Using filter sets
- `study-with-participant-group.json` - Using participant groups
- `study-in-project.json` - Assigned to a project
- Many more examples for various configurations

Templates support both JSON and YAML formats.

## Git Workflow

### Git Hooks

Installed via `make install`. Hook scripts live in `scripts/hooks/`.

#### Pre-commit (`scripts/hooks/pre-commit`)

Automatically runs before each commit:

1. `make lint` - Lints all Go code
2. `make test` - Runs test suite
3. `make lint-dockerfile` - Lints Dockerfile (if changed)

#### Commit-msg (`scripts/hooks/commit-msg`)

Enforces [Conventional Commits](https://www.conventionalcommits.org/) format on commit messages:

```
<type>(<scope>): <description>
```

- **Types:** `feat`, `fix`, `docs`, `refactor`, `test`, `chore`, `ci`, `build`, `perf`, `style`, `revert`
- **Scope** is optional (e.g. `fix: thing` and `fix(DCP-123): thing` are both valid)
- Merge commits and fixup/squash prefixes are automatically skipped

### Branch and Release Strategy

- Main branch: `main`
- Clean status expected (no uncommitted changes)
- Releases tagged with version numbers (e.g., `v0.0.60`)

#### Release flow

1. Run `make changelog VERSION=0.0.60` to generate grouped release notes
2. Create a PR with the updated `CHANGELOG.md` and include `[run-release]` in the PR title (e.g. `chore: release v0.0.60 [run-release]`)
3. Get the PR reviewed and merge to `main` — CI automatically creates the git tag, GitHub Release, and uploads binaries

Two CI gates guard the PR:
- **`changelog-gate.yml`** — fails if `CHANGELOG.md` is not modified when `[run-release]` is present
- **`release-tag-gate.yml`** — fails on common `[run-release]` misspellings or wrong casing

## Changelog Conventions

Changelog entries are generated from conventional commits by [git-cliff](https://git-cliff.org/). The configuration lives in `cliff.toml`, and a Go tool in `scripts/changelog/` groups entries by subcommand area.

### Versioning

This project uses [Semantic Versioning](https://semver.org/) (`MAJOR.MINOR.PATCH`). While the project is pre-1.0, all releases use `0.0.x`.

**Release naming:** Git tags and GitHub Releases must use a leading `v` (e.g. `v1.0.1`), not a bare version string (`1.0.1`). The automated release workflow creates both the tag and the release title as `vx.y.z`. `CHANGELOG.md` headings stay as bare semver (`## 1.0.1`) — that is intentional. When you run `make changelog`, pass `VERSION` **without** the `v` (e.g. `VERSION=1.0.1`); tooling adds the prefix for tags and releases.

| Change | Version bump | Example |
|--------|-------------|---------|
| Breaking change to an existing command (flag removed, output format changed) | `MINOR` (`0.0.x` → `0.1.0`) | Removing a flag, changing JSON output shape |
| New command or flag | `PATCH` | Adding `study delete` |
| Bug fix, docs, refactor, CI | `PATCH` | Fixing a nil-pointer crash |

**Pre-1.0 note:** `MAJOR` stays at `0` until the API and command surface are considered stable. `MINOR` bumps signal breaking changes for the duration of `0.x`.

The version is passed to `make changelog VERSION=x.y.z` (numeric only, no `v` prefix) — there is no automated bump calculation; the release author decides based on the table above.

### Manual release notes

To include hand-written notes in the next release, add them under the `## next` section in `CHANGELOG.md`:

```markdown
## next

- My manual release note here
```

At release time `make changelog` merges any `## next` content with the generated notes and resets the section.

### What gets included

Only user-facing commit types appear in the changelog:
- `feat` → **Features**, `fix` → **Bug Fixes**, `docs` → **Documentation**, `perf` → **Performance**, `refactor` → **Refactoring**, `revert` → **Reverts**, `test` → **Testing**
- `chore`, `ci`, `build`, `style` are **skipped** (internal housekeeping)

### Format

- `## x.y.z` for released versions (no dates)
- Grouped by subcommand area with bold scope prefix

## CI/CD

GitHub Actions workflows in `.github/workflows/`:

### `go.yml` (runs on every push)

1. Setup Go 1.26.x
2. `make install` - Get dependencies
3. Lint Go code with golangci-lint
4. `make lint-dockerfile`
5. `make build`
6. `make test`

### `docker.yml`

Builds and pushes Docker images.

### `changelog-gate.yml`

Runs on pull request events. Fails if the PR title contains `[run-release]` but `CHANGELOG.md` is not modified.

### `release-tag-gate.yml`

Runs on pull request events. Validates that `[run-release]` is spelled and cased correctly in the PR title. Catches common typos and wrong separators so that the release is not silently skipped.

### `create-release.yml`

Runs when a PR is merged to `main` with `[run-release]` in the title. Uses `go run ./scripts/changelog extract-version` to read the version from the top-most `## x.y.z` section in `CHANGELOG.md`, then:

1. Creates and pushes a `vx.y.z` annotated tag
2. Creates a GitHub Release named `vx.y.z` (tag and release title both use the `v` prefix) with the matching changelog entry as notes — publishing it immediately, which triggers `release.yml` via the `release: published` event

### `release.yml`

Builds binaries when a GitHub Release is published. Triggered by the `release: published` event (including releases created by `create-release.yml`). Builds for darwin, linux, windows, and freebsd, then uploads assets to the release.

## Common Patterns

### Adding a New Command

1. Create package under `cmd/<resource>/`
2. Implement command function(s) following pattern:

   ```go
   func NewXCommand(client client.API, w io.Writer) *cobra.Command
   ```

3. Add tests in `<command>_test.go`
4. Register in `cmd/root.go:65-80`
5. Update `CHANGELOG.md`

### Adding a New API Method

1. Add method signature to `client.API` interface (`client/client.go:26-70`)
2. Implement method on `Client` struct
3. Add request/response structs to `payloads.go`/`responses.go`
4. Update mock: `make test-gen-mock`
5. Write tests using the mock

### Error Handling

API errors are structured as `JSONAPIError`:

```go
type JSONAPIError struct {
    Error struct {
        Detail string `json:"detail"`
    } `json:"error"`
}
```

Client automatically handles 400+ status codes and returns formatted errors.

## Important Gotchas

### Configuration Loading

- Viper config is loaded in `cmd/root.go:initConfig()` **before** commands are built
- The `New()` client constructor depends on Viper being initialized
- Config file is optional - app works with just environment variables

### Testing Output

- Commands accept `io.Writer` for output
- Tests should use `bytes.Buffer` wrapped in `bufio.Writer`
- **Must call `writer.Flush()`** before asserting output
- See `cmd/workspace/list_test.go:70-87` for the pattern

### Mock Regeneration

- Mocks are generated from the `client.API` interface
- After changing interface, **must** run `make test-gen-mock`
- Generated file: `mock_client/mock_client.go`
- Do not edit generated mocks manually

### Interactive Mode

- Interactive commands use Bubbletea's `tea.Program`
- Requires terminal support
- Always provide non-interactive flag option (`-n`)
- Can't test interactive mode easily - test the underlying logic

### API Token

- All API calls require `PROLIFIC_TOKEN` environment variable
- Client will return error if token not set: `"PROLIFIC_TOKEN not set"`
- Get token from: <https://app.prolific.com/researcher/tokens/>

## Dependencies

Key dependencies from `go.mod`:

**CLI/UI:**

- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration management
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/lipgloss` - Terminal styling
- `github.com/charmbracelet/bubbles` - TUI components

**Testing:**

- `github.com/golang/mock` - Mocking framework

**Utilities:**

- `github.com/mitchellh/go-homedir` - Home directory detection
- `github.com/pkg/browser` - Open URLs in browser
- `golang.org/x/text` - Text processing (currency, i18n)

Update dependencies with `go get` and run `go mod tidy`.

## Version Information

Version is injected at build time via `-ldflags`:

```bash
-X github.com/prolific-oss/cli/version.Version=$(GIT_RELEASE)
```

Stored in `version/version.go` and displayed in root command.

## Browser Integration

Some commands support `-W` flag to open resources in browser:

- `project view [id] -W`
- `study view [id] -W`
- `filter-sets view [id] -W`

Uses `github.com/pkg/browser` package.

## Security Notes

- Never commit `PROLIFIC_TOKEN` to git
- Use `.gitignore` to exclude config files with tokens
- `gosec` linter runs in CI for security checks
- Docker images scanned with `docker scout`

## Useful Links

- [Project Repository](https://github.com/prolific-oss/cli)
- [Project Wiki](https://github.com/prolific-oss/cli/wiki)
- [Prolific API](https://docs.prolific.com/docs/api-docs/public/)
- [Get API Token](https://app.prolific.com/researcher/tokens/)

## Quick Reference

| Task | Command |
|------|---------|
| Install deps & setup | `make install` |
| Build | `make build` |
| Test | `make test` |
| Test with coverage HTML | `make test-cov` |
| Lint | `make lint` |
| Full workflow | `make all` |
| Generate mocks | `make test-gen-mock` |
| Run CLI | `./prolific --help` |
| List studies | `./prolific study list` |
| Create study | `./prolific study create -t <template>` |

## Notes for AI Agents

- **Always run tests after changes**: `make test`
- **Regenerate mocks** after interface changes: `make test-gen-mock`
- **Follow existing patterns** - look at similar commands for reference
- **Test output capture**: Remember to `Flush()` the writer in tests
- **Check linter** before committing: `make lint`
- **Update CHANGELOG.md** when adding features or fixing bugs
- **Use dependency injection** - pass `client.API` and `io.Writer` to commands
- **Consider interactive vs non-interactive** modes for list commands
- **Look at examples** in `docs/examples/` for study template structure
