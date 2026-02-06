# Tasks: Fix Message Commands — Prolific API Conformance

**Input**: Design documents from `/specs/002-fix-message-api-conformance/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Included — the spec explicitly requires tests (FR-009, SC-004, Constitution Principle II).

**Organization**: Tasks grouped by user story. Foundational phase blocks all stories because model and interface changes must compile before commands can be updated.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story (US1–US5) or cross-cutting
- Exact file paths included in descriptions

---

## Phase 1: Foundational (Blocking Prerequisites)

**Purpose**: Model corrections, interface changes, and mock regeneration that MUST complete before ANY command-level work. These changes are atomic — the project won't compile mid-phase.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete and `make test-gen-mock` has run.

- [x] T001 Update `Message` struct in `model/model.go`: rename `DatetimeCreated` → `SentAt` (`json:"sent_at"`), add `ID` (`json:"id"`), `Type` (`json:"type,omitempty"`), `ChannelID` (`json:"channel_id"`), restructure `Data` from `map[string]any` to `*MessageData`, remove top-level `StudyID` field
- [x] T002 Add `MessageData` struct in `model/model.go` with `StudyID` (`json:"study_id,omitempty"`) and `Category` (`json:"category,omitempty"`)
- [x] T003 Delete `UnreadMessage` struct from `model/model.go` (lines 78-83)
- [x] T004 Update `ListUnreadMessagesResponse` in `client/responses.go`: change `Results` type from `[]model.UnreadMessage` to `[]model.Message`
- [x] T005 Add `BulkSendMessagePayload` struct in `client/payloads.go` per `data-model.md`
- [x] T006 [P] Add `SendGroupMessagePayload` struct in `client/payloads.go` per `data-model.md`
- [x] T007 Remove study_id extraction loop in `GetMessages` implementation (`client/client.go:660-667`) — the typed `MessageData` struct handles this via JSON unmarshaling (see research.md Decision 3)
- [x] T008 Add `BulkSendMessage(ids []string, body, studyID string) error` to `client.API` interface and implement in `client/client.go` — POST to `/api/v1/messages/bulk/` with `BulkSendMessagePayload`, nil response (204)
- [x] T009 Add `SendGroupMessage(participantGroupID, body string, studyID *string) error` to `client.API` interface and implement in `client/client.go` — POST to `/api/v1/messages/participant-group/` with `SendGroupMessagePayload`, nil response (204)
- [x] T010 Run `make test-gen-mock` to regenerate `mock_client/mock_client.go` with updated interface

**Checkpoint**: Model, payloads, responses, and client interface are all updated. Mocks are regenerated. The project compiles but existing message tests will fail (expected — they reference old field names).

---

## Phase 2: User Story 1 — Retrieve Messages with Correct Fields (Priority: P1) 🎯

**Goal**: `prolific message list -u <user_id>` displays curated columns matching the API schema.
**Spec**: US1 (FR-001, FR-003, FR-004)
**Contract**: `contracts/get-messages.md`

**Independent Test**: `go test ./cmd/message/... -run TestListMessages -v`

### Tests for User Story 1

- [x] T011 [US1] Update `TestListMessages` in `cmd/message/list_test.go`: mock response uses new `Message` struct with `ID`, `SenderID`, `Body`, `SentAt`, `Data` (containing `StudyID` and `Category`); assert output contains 6 curated columns: `ID`, `Sender ID`, `Study ID`, `Category`, `Sent At`, `Body`
- [x] T012 [US1] Update `TestListMessagesRaw` in `cmd/message/list_test.go` to use new `Message` struct fields

### Implementation for User Story 1

- [x] T013 [US1] Update `renderMessages` function in `cmd/message/list.go`: replace current column headers and row formatting with 6 curated columns (`ID`, `Sender ID`, `Study ID`, `Category`, `Sent At`, `Body`); access `StudyID` and `Category` via `msg.Data.StudyID` and `msg.Data.Category` (with nil-safe check on `Data`)
- [x] T014 [US1] Run tests: `go test ./cmd/message/... -run TestListMessages -v` — verify pass

**Checkpoint**: `prolific message list -u <user_id>` shows correct columns from API schema.

---

## Phase 3: User Story 2 — Retrieve Unread Messages with Correct Fields (Priority: P1)

**Goal**: `prolific message list -U` uses the unified `Message` model and displays same columns as regular list.
**Spec**: US2 (FR-002, FR-003)
**Contract**: `contracts/get-unread.md`

**Independent Test**: `go test ./cmd/message/... -run TestListUnreadMessages -v`

### Tests for User Story 2

- [x] T015 [US2] Update `TestListUnreadMessages` in `cmd/message/list_test.go`: change mock to return `ListUnreadMessagesResponse` with `[]model.Message` (not `[]model.UnreadMessage`); assert same 6 curated column output as regular list
- [x] T016 [US2] Update `TestListUnreadMessagesRaw` in `cmd/message/list_test.go` to use `Message` struct fields

### Implementation for User Story 2

- [x] T017 [US2] Update unread branch in `cmd/message/list.go`: remove the separate `renderUnreadMessages` code path (if any), reuse the `renderMessages` function for unread results since the response type is now `[]model.Message`
- [x] T018 [US2] Ensure mutual exclusivity validation: `--unread` flag cannot combine with `--user` or `--created_after` — verify error message in existing code or add if missing
- [x] T019 [US2] Run tests: `go test ./cmd/message/... -run TestListUnread -v` — verify pass

**Checkpoint**: `prolific message list -U` produces identical column layout to regular list.

---

## Phase 4: User Story 3 — Send Message with Correct Payload (Priority: P1)

**Goal**: `prolific message send` handles 204 No Content correctly and sends correct payload.
**Spec**: US3 (FR-005)
**Contract**: `contracts/send-message.md`

**Independent Test**: `go test ./cmd/message/... -run TestSendMessage -v`

### Tests for User Story 3

- [x] T020 [US3] Review `TestSendMessage` in `cmd/message/send_test.go`: verify mock expectation uses `SendMessage(body, recipientID, studyID)` returning `nil` error (204 pattern); assert confirmation output shows `Recipient ID`, `Study ID`, `Body`
- [x] T021 [US3] Review `TestSendMessageError` in `cmd/message/send_test.go`: verify error path still works with updated mock interface

### Implementation for User Story 3

- [x] T022 [US3] Verify `cmd/message/send.go` confirmation output table matches contract columns (`Recipient ID`, `Study ID`, `Body`); fix if needed
- [x] T023 [US3] Run tests: `go test ./cmd/message/... -run TestSend -v` — verify pass

**Checkpoint**: `prolific message send` correctly sends payload and handles 204.

---

## Phase 5: User Story 4 — Bulk Send to Multiple Participants (Priority: P2)

**Goal**: New `prolific message bulk-send` command for `POST /api/v1/messages/bulk/`.
**Spec**: US4 (FR-006, FR-008, FR-009, FR-010)
**Contract**: `contracts/bulk-message.md`

**Independent Test**: `go test ./cmd/message/... -run TestBulkSend -v`

### Tests for User Story 4

- [x] T024 [US4] Create `cmd/message/bulk_send_test.go`: test happy path — mock `BulkSendMessage([]string{"id1","id2"}, "body", "study_id")` returns nil; assert confirmation output shows `Recipients: 2`, `Study ID`, `Body`
- [x] T025 [P] [US4] Add error test in `cmd/message/bulk_send_test.go`: mock returns error, assert error propagated
- [x] T026 [P] [US4] Add validation test in `cmd/message/bulk_send_test.go`: empty `--ids` flag produces error before API call

### Implementation for User Story 4

- [x] T027 [US4] Create `cmd/message/bulk_send.go`: `NewBulkSendCommand(client client.API, w io.Writer) *cobra.Command` with flags `--ids/-i` (string, comma-separated), `--body/-b` (string), `--study/-s` (string); all required. Split `--ids` on comma to `[]string`, validate at least one ID, call `client.BulkSendMessage(ids, body, studyID)`, display confirmation table
- [x] T028 [US4] Register `bulk-send` subcommand in `cmd/message/message.go`
- [x] T029 [US4] Run tests: `go test ./cmd/message/... -run TestBulkSend -v` — verify pass

**Checkpoint**: `prolific message bulk-send --ids id1,id2 -s study -b "msg"` works end-to-end with mocks.

---

## Phase 6: User Story 5 — Send Message to Participant Group (Priority: P2)

**Goal**: New `prolific message send-group` command for `POST /api/v1/messages/participant-group/`.
**Spec**: US5 (FR-007, FR-008, FR-009, FR-010)
**Contract**: `contracts/send-group.md`

**Independent Test**: `go test ./cmd/message/... -run TestSendGroup -v`

### Tests for User Story 5

- [x] T030 [US5] Create `cmd/message/send_group_test.go`: test happy path with study ID — mock `SendGroupMessage("group_id", "body", &studyID)` returns nil; assert confirmation shows `Group ID`, `Study ID`, `Body`
- [x] T031 [P] [US5] Add happy path test without study ID: mock `SendGroupMessage("group_id", "body", nil)` returns nil; assert `Study ID` shows "N/A"
- [x] T032 [P] [US5] Add error test: mock returns error, assert error propagated
- [x] T033 [P] [US5] Add validation test: missing `--group` flag produces error

### Implementation for User Story 5

- [x] T034 [US5] Create `cmd/message/send_group.go`: `NewSendGroupCommand(client client.API, w io.Writer) *cobra.Command` with flags `--group/-g` (string, required), `--body/-b` (string, required), `--study/-s` (string, optional). If `--study` empty, pass `nil` for studyID. Call `client.SendGroupMessage(groupID, body, studyID)`, display confirmation table with `Group ID`, `Study ID` (or "N/A"), `Body`
- [x] T035 [US5] Register `send-group` subcommand in `cmd/message/message.go`
- [x] T036 [US5] Run tests: `go test ./cmd/message/... -run TestSendGroup -v` — verify pass

**Checkpoint**: `prolific message send-group --group gid -b "msg"` works end-to-end with mocks.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Full validation, linting, and final checks across all stories.

- [x] T037 Run `make test` — all tests pass (SC-005)
- [x] T038 Run `make lint` — zero warnings (SC-005)
- [x] T039 Verify `--help` output for all 5 message subcommands matches API documentation (SC-006): `message list --help`, `message send --help`, `message bulk-send --help`, `message send-group --help`
- [x] T040 Run quickstart.md verification steps (Section 1: model verification, Section 2: test verification, Section 3: mock regeneration check)
- [x] T041 Verify constitution compliance: all changes limited to scoped files (`cmd/message/`, `client/`, `model/`)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Foundational (Phase 1)**: No dependencies — start immediately. BLOCKS all user stories.
- **US1 (Phase 2)**: Depends on Phase 1 completion.
- **US2 (Phase 3)**: Depends on Phase 1 completion. Can run in parallel with US1 (different code paths in `list.go`), but recommend sequential since US1's `renderMessages` changes inform US2's unread path.
- **US3 (Phase 4)**: Depends on Phase 1 completion. Can run in parallel with US1/US2 (different file: `send.go`).
- **US4 (Phase 5)**: Depends on Phase 1 completion. Can run in parallel with other stories (new file: `bulk_send.go`).
- **US5 (Phase 6)**: Depends on Phase 1 completion. Can run in parallel with other stories (new file: `send_group.go`).
- **Polish (Phase 7)**: Depends on ALL user stories being complete.

### Within Each User Story

- Tests MUST be written/updated first (TDD per Constitution Principle II)
- Implementation follows tests
- Story-level test run confirms the story works

### Parallel Opportunities

After Phase 1 completes:
- US3 (Phase 4), US4 (Phase 5), and US5 (Phase 6) operate on different files and can run fully in parallel
- US1 (Phase 2) and US2 (Phase 3) share `list.go` — recommend sequential (US1 → US2)

### Recommended Execution Order (Single Developer)

```
Phase 1 (Foundational) → Phase 2 (US1) → Phase 3 (US2) → Phase 4 (US3) → Phase 5 (US4) → Phase 6 (US5) → Phase 7 (Polish)
```

This order follows priority (P1 before P2) and minimizes merge conflicts in shared files.

---

## Task Summary

| Phase | Tasks | Stories | Priority |
|-------|-------|---------|----------|
| 1. Foundational | T001–T010 | All | Blocking |
| 2. US1 Retrieve Messages | T011–T014 | US1 | P1 |
| 3. US2 Unread Messages | T015–T019 | US2 | P1 |
| 4. US3 Send Message | T020–T023 | US3 | P1 |
| 5. US4 Bulk Send | T024–T029 | US4 | P2 |
| 6. US5 Send Group | T030–T036 | US5 | P2 |
| 7. Polish | T037–T041 | Cross-cutting | — |
| **Total** | **41 tasks** | **5 stories** | |
