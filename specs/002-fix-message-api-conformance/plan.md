# Implementation Plan: Fix Message Commands — Prolific API Conformance

**Branch**: `002-fix-message-api-conformance` | **Date**: 2026-02-06 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/002-fix-message-api-conformance/spec.md`

## Summary

Fix the existing Prolific CLI message commands (`list`, `send`) and
their underlying models to match the current Prolific Messages API
schema, and implement two missing commands (`bulk-send`, `send-group`)
for documented but unimplemented API endpoints. The core changes are:
model field corrections (`datetime_created` → `sent_at`, missing
`id`/`type`/`channel_id` fields), `UnreadMessage` struct elimination,
updated table output with curated columns, and two new Cobra commands
with full test coverage.

## Technical Context

**Language/Version**: Go (matching parent `prolific-oss/cli`)
**Primary Dependencies**: Cobra (`github.com/spf13/cobra`), Viper
  (`github.com/spf13/viper`), gomock
**Storage**: N/A (API client only)
**Testing**: Go standard `testing` + `gomock` (`mock_client`)
**Target Platform**: Cross-platform CLI binary
**Project Type**: Single project (Go module)
**Performance Goals**: N/A (CLI command, not a server)
**Constraints**: Changes scoped to message-related files per
  constitution; `make test` and `make lint` must pass
**Scale/Scope**: 5 API endpoints, ~10 files modified/created

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Evidence |
|-----------|--------|----------|
| I. API-First Conformance | PASS | All 5 endpoints mapped to verified API doc URLs in spec |
| II. Test-First | PASS | Plan requires tests before implementation for every change |
| III. Security & Input Validation | PASS | Flag validation specified for all commands; no sensitive data leak |
| IV. Go Style & Idioms | PASS | DI pattern maintained; JSON tags match API schema; doc comments required |
| V. Simplicity & Focus | PASS | Changes limited to `cmd/message/`, `client/`, `model/` per constitution scope |

No violations. No Complexity Tracking entries needed.

## Project Structure

### Documentation (this feature)

```text
specs/002-fix-message-api-conformance/
├── plan.md              # This file
├── research.md          # Phase 0: codebase audit findings
├── data-model.md        # Phase 1: corrected data models
├── quickstart.md        # Phase 1: verification commands
├── contracts/           # Phase 1: API endpoint contracts
│   ├── get-messages.md
│   ├── send-message.md
│   ├── bulk-message.md
│   ├── send-group.md
│   └── get-unread.md
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (files to modify/create)

```text
# Existing files to MODIFY:
model/model.go                    # Fix Message struct, remove UnreadMessage
client/client.go                  # Fix GetMessages, add BulkSend/SendGroup
client/payloads.go                # Add BulkSendPayload, GroupSendPayload
client/responses.go               # Fix ListUnreadMessagesResponse
cmd/message/message.go            # Register new subcommands
cmd/message/list.go               # Update table columns, use Message for unread
cmd/message/list_test.go          # Update tests for new model/columns
cmd/message/send_test.go          # Verify 204 handling tests

# New files to CREATE:
cmd/message/bulk_send.go          # New bulk-send command
cmd/message/bulk_send_test.go     # Tests for bulk-send
cmd/message/send_group.go         # New send-group command
cmd/message/send_group_test.go    # Tests for send-group

# Generated files (auto):
mock_client/mock_client.go        # Regenerated via make test-gen-mock
```

**Structure Decision**: This is an existing Go CLI project. All
changes follow the established directory layout. No new directories
needed — only new files within `cmd/message/`.
