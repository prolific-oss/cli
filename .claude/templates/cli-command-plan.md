# {TICKET-NUMBER}: CLI {resource} {command} Command

## Overview

{Brief description of command purpose and context}

## Command Type

<!-- Select one: LIST | VIEW | CREATE | UPDATE | ACTION -->
**Type**: {COMMAND_TYPE}

## API Contract

<!-- Provide ONE of the following: Bruno file OR inline contract -->

### Option A: Bruno File
<!-- Provide file path OR paste .bru contents directly -->
```
{path/to/request.bru OR paste .bru file contents here}
```

### Option B: Inline Contract

**Example:**
```
GET /api/v1/data-collection/collections?workspace_id={id}&limit={n}&offset={n}

Response 200:
{
  "results": [
    {
      "id": "uuid",
      "name": "string",
      "created_at": "ISO8601",
      "created_by": "string",
      "item_count": number
    }
  ],
  "meta": { "count": number }
}

Response 404:
{ "error": { "detail": "Workspace not found" } }
```

### Error Responses

- 400: {when/why}
- 404: {when/why}
- 500: {when/why}

## Flags to Support

<!-- Check flags needed for this command -->
- [ ] `--workspace` / `-w` - Workspace ID (required)
- [ ] `--non-interactive` / `-n` - Table output to terminal
- [ ] `--csv` / `-c` - CSV format output
- [ ] `--json` - JSON format output
- [ ] `--fields` / `-f` - Comma-separated field list
- [ ] `--limit` - Pagination limit (default 200)
- [ ] `--offset` - Pagination offset (default 0)
- [ ] `--web` / `-W` - Open in browser (VIEW commands)
- [ ] {custom flag}

## Renderers Required

<!-- LIST commands: select which renderers to implement -->
<!-- VIEW commands: skip this section - use single Render function in ui/{resource}/view.go -->

- [ ] InteractiveRenderer (bubbletea)
- [ ] NonInteractiveRenderer (table)
- [ ] CsvRenderer
- [ ] JSONRenderer

## Implementation Layers

| Layer | File | Purpose |
|-------|------|---------|
| Model | `model/{resource}.go` | Go struct with JSON tags |
| Response | `client/responses.go` | API response wrapper |
| Client Interface | `client/client.go` | Add method to API interface |
| Client Impl | `client/client.go` | Implement API method |
| Parent Command | `cmd/{resource}/{resource}.go` | Group subcommands |
| Command | `cmd/{resource}/{command}.go` | Cobra command + flags |
| UI Renderers | `ui/{resource}/{command}.go` | Output formatters |
| Mock | `mock_client/mock_client.go` | Test mock |
| Tests | `cmd/{resource}/{command}_test.go` | Command tests |
| UI Tests | `ui/{resource}/list_test.go` | Renderer tests (LIST commands only) |

## Pre-implementation Tasks

{Ask user if they wish to set up a feature branch, if yes carry out the following}

- [ ] Stash any local changes: `git stash`
- [ ] Checkout main and pull latest: `git checkout main && git pull`
- [ ] Create feature branch: `git checkout -b {TICKET-NUMBER}-{resource}-{command}-command`

## Implementation Tasks

### 1. Model
- [ ] Create `model/{resource}.go`
  - Struct with json tags
  - `FilterValue()`, `Title()`, `Description()` for bubbletea (LIST only)

### 2. Client
- [ ] Add response type to `client/responses.go`
- [ ] Add method to `API` interface in `client/client.go`
- [ ] Implement method on `Client` struct

### 3. Commands
- [ ] Create parent command `cmd/{resource}/{resource}.go` (if new resource)
- [ ] Create `cmd/{resource}/{command}.go`
  - Options struct
  - `New{Command}Command()` function
  - Flag definitions
  - RunE implementation

### 4. UI (command type specific)

**LIST commands:**
- [ ] Create `ui/{resource}/list.go`
  - `ListUsedOptions` struct
  - `ListStrategy` interface
  - Renderer implementations (Interactive, NonInteractive, CSV, JSON)

**VIEW commands:**
- [ ] Create `ui/{resource}/view.go`
  - Single `Render{Resource}()` function

### 5. Mocks & Tests
- [ ] Regenerate mock: `mockgen -source=client/client.go -destination=mock_client/mock_client.go -package=mock_client`
- [ ] Create `cmd/{resource}/{command}_test.go`

**LIST commands only:**
- [ ] Create `ui/{resource}/list_test.go`

**Note:** VIEW commands do NOT require separate UI tests. The command tests in `cmd/{resource}/{command}_test.go` exercise the `Render{Resource}()` function indirectly.

### 6. Wire Up
- [ ] Add command to `cmd/root.go`:
  ```go
  rootCmd.AddCommand({resource}.New{Resource}Command(c, os.Stdout))
  ```

## Reference Files

| Pattern | File |
|---------|------|
| LIST command | `cmd/collection/list.go` |
| LIST renderers | `ui/collection/list.go` |
| VIEW command | `cmd/project/view.go` |
| CREATE command | `cmd/project/create.go` |
| UPDATE command | `cmd/credentials/update.go` |
| ACTION command | `cmd/study/transition.go` |
| Parent command | `cmd/collection/collection.go` |
| Model | `model/collection.go` |
| Client method | `client/client.go:GetCollections` |
