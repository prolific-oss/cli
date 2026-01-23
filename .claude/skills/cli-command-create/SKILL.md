---
name: cli-command-create
description: Create a new CLI command with planning and implementation. Use when adding a new command to the Prolific CLI.
disable-model-invocation: true
argument-hint: <ticket> <resource> <command> [command-type]
---

# CLI Command Creator

Create a new CLI command for the Prolific CLI from start to finish.

## Arguments

Parse from `$ARGUMENTS`:
- `ticket` - Jira ticket number (e.g., DCP-2154)
- `resource` - Resource name (e.g., collection, study, workspace)
- `command` - Command name (e.g., list, get, create, update)
- `command-type` - Optional: LIST | VIEW | CREATE | UPDATE | ACTION (will infer if not provided)

---

## Phase 1: Gather Requirements

### 1.1 Understand the API Contract

Ask the user to provide ONE of:
- **Bruno file path** - e.g., `path/to/request.bru`
- **Inline API contract** - endpoint, request/response examples, error codes

### 1.2 Determine Command Type

If not provided, infer from command name:
- `list` → LIST (paginated results, multiple renderers)
- `get`, `view` → VIEW (single resource, one renderer)
- `create` → CREATE (accepts template file)
- `update` → UPDATE (modifies existing resource)
- Other → ACTION (state transitions, operations)

### 1.3 Identify Required Flags

Based on command type, confirm which flags are needed:

| Flag | LIST | VIEW | CREATE | UPDATE | ACTION |
|------|------|------|--------|--------|--------|
| `--workspace` / `-w` | Often | Often | Often | Often | Rare |
| `--non-interactive` / `-n` | Yes | No | No | No | No |
| `--csv` / `-c` | Yes | No | No | No | No |
| `--json` | Optional | Optional | No | No | No |
| `--limit` / `--offset` | Yes | No | No | No | No |
| `--web` / `-W` | No | Yes | No | No | No |
| `--template` / `-t` | No | No | Yes | Yes | No |

Ask user to confirm or add custom flags.

### 1.4 Present Plan Summary

Before proceeding, present:
1. Files to create/modify
2. Implementation approach
3. Test strategy

**Ask for explicit approval before Phase 2.**

---

## Phase 2: Implement

### 2.1 Model Layer

Create/update `model/{resource}.go`:
- Struct with JSON tags matching API response
- For LIST commands: implement `FilterValue()`, `Title()`, `Description()` for bubbletea

### 2.2 Client Layer

1. Add response type to `client/responses.go`
2. Add method signature to `API` interface in `client/client.go`
3. Implement method on `Client` struct

### 2.3 Command Layer

1. Create parent command `cmd/{resource}/{resource}.go` (if new resource)
2. Create `cmd/{resource}/{command}.go`:
   - Options struct
   - `New{Command}Command(client client.API, w io.Writer) *cobra.Command`
   - Flag definitions
   - RunE implementation with dependency injection

### 2.4 UI Layer

**LIST commands:**
- Create `ui/{resource}/list.go` with:
  - `ListStrategy` interface
  - `InteractiveRenderer` (bubbletea)
  - `NonInteractiveRenderer` (table)
  - `CsvRenderer`
  - Optional: `JSONRenderer`

**VIEW commands:**
- Create `ui/{resource}/view.go` with single `Render{Resource}()` function

### 2.5 Wire Up

Add command to `cmd/root.go`:
```go
rootCmd.AddCommand({resource}.New{Resource}Command(c, os.Stdout))
```

### 2.6 Generate Mocks

Run: `make test-gen-mock`

### 2.7 Tests

Create `cmd/{resource}/{command}_test.go`:
- Use `gomock` with `mock_client.NewMockAPI(ctrl)`
- Test success cases
- Test error handling
- Remember: `defer ctrl.Finish()` and `writer.Flush()`

For LIST commands only: Create `ui/{resource}/list_test.go`

---

## Phase 3: Verify

1. Run `make build` - ensure it compiles
2. Run `make test` - ensure all tests pass
3. Run `make lint` - ensure no lint errors
4. Manual smoke test (if user has `PROLIFIC_TOKEN` set)

Report results and any issues found.

---

## Reference Patterns

| Pattern | Reference File |
|---------|----------------|
| LIST command | `cmd/collection/list.go` |
| LIST renderers | `ui/collection/list.go` |
| VIEW command | `cmd/project/view.go` |
| CREATE command | `cmd/project/create.go` |
| UPDATE command | `cmd/credentials/update.go` |
| ACTION command | `cmd/study/transition.go` |
| Parent command | `cmd/collection/collection.go` |
| Model | `model/collection.go` |
| Client method | `client/client.go:GetCollections` |
| Test pattern | `cmd/workspace/list_test.go` |

$ARGUMENTS
