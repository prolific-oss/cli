# Feature Specification: Fix Message Commands — Prolific API Conformance

**Feature Branch**: `002-fix-message-api-conformance`
**Created**: 2026-02-06
**Status**: Draft
**Input**: Fix existing message CLI commands and add missing commands to
conform with the Prolific Messages API documentation.
**Supersedes**: `001-check-message-api-conformance` (contained unverified
endpoint assumptions; this spec uses verified documentation URLs)

## API Reference (Verified Endpoints)

The following 5 endpoints are the complete Prolific Messages API surface.
Each URL links to the authoritative documentation:

1. **Retrieve messages**: [get-messages.md](https://docs.prolific.com/api-reference/messages/get-messages.md)
   `GET /api/v1/messages/`
2. **Send a message**: [send-message.md](https://docs.prolific.com/api-reference/messages/send-message.md)
   `POST /api/v1/messages/`
3. **Send to multiple participants**: [bulk-message-participants](https://docs.prolific.com/api-reference/messages/bulk-message-participants)
   `POST /api/v1/messages/bulk/`
4. **Send to participant group**: [send-message-to-participant-group](https://docs.prolific.com/api-reference/messages/send-message-to-participant-group)
   `POST /api/v1/messages/participant-group/`
5. **Retrieve unread messages**: [get-unread-messages.md](https://docs.prolific.com/api-reference/messages/get-unread-messages.md)
   `GET /api/v1/messages/unread/`

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Retrieve messages with correct fields (Priority: P1)

A researcher uses `prolific message list` to view message history.
The CLI MUST return all fields defined in the API response schema so
that output matches what the API actually provides. The current CLI
model is missing fields and uses incorrect field names.

**Current gaps** (verified against [get-messages.md](https://docs.prolific.com/api-reference/messages/get-messages.md)):
- CLI `Message` struct has `datetime_created` — API uses `sent_at`
- CLI `Message` struct is missing `id`, `type`, `channel_id` fields
- CLI extracts `study_id` from `data` map at runtime — API returns
  it as a nested `data.study_id` field
- API also provides `data.category` (enum) which CLI ignores

**Why this priority**: Most commonly used message command. Incorrect
model causes silent data loss.

**Independent Test**: Run `prolific message list -u <user_id>` and
verify output contains correct field names and all API-specified data.

**Acceptance Scenarios**:

1. **Given** the API returns a message with fields `id`, `sender_id`,
   `body`, `sent_at`, `type`, `channel_id`, and `data` (containing
   `study_id` and `category`),
   **When** the user runs `prolific message list -u <user_id>`,
   **Then** the CLI displays curated columns: `id`, `sender_id`,
   `study_id`, `category`, `sent_at`, `body`. Fields `type` and
   `channel_id` are present in the domain model but omitted from
   default table display.
2. **Given** no `user_id` or `created_after` parameter is provided,
   **When** the user runs `prolific message list`,
   **Then** the CLI returns an error stating at least one filter is
   required.

---

### User Story 2 - Retrieve unread messages with correct fields (Priority: P1)

A researcher uses `prolific message list -U` to see unread messages.
The API's unread endpoint returns the same `Message` schema as the
regular retrieve endpoint, but the CLI uses a separate
`UnreadMessage` struct with fewer fields and different names.

**Current gaps** (verified against [get-unread-messages.md](https://docs.prolific.com/api-reference/messages/get-unread-messages.md)):
- CLI `UnreadMessage` struct uses `sender` — API uses `sender_id`
- CLI `UnreadMessage` struct uses `datetime_created` — API uses
  `sent_at`
- CLI `UnreadMessage` struct is missing `id`, `type`, `channel_id`,
  and `data` fields
- API returns same `Message` object for both endpoints

**Why this priority**: Unread messages are critical for researchers
responding to participant inquiries. Wrong field mapping produces
confusing output.

**Independent Test**: Run `prolific message list -U` and verify
output fields match the API's unread message response schema and
are consistent with the regular message list output.

**Acceptance Scenarios**:

1. **Given** the API returns unread messages with the full `Message`
   schema,
   **When** the user runs `prolific message list -U`,
   **Then** the CLI displays the same curated columns as regular
   message list (consistent output).
2. **Given** the user combines `--unread` with `--user` or
   `--created_after`,
   **When** the command processes the flags,
   **Then** a clear error states that `--unread` cannot combine
   with other filter flags.

---

### User Story 3 - Send a message with correct payload (Priority: P1)

A researcher uses `prolific message send` to contact a participant.
The payload MUST match the API's request body and the CLI MUST
handle the 204 No Content response correctly.

**Current state** (verified against [send-message.md](https://docs.prolific.com/api-reference/messages/send-message.md)):
- CLI payload fields (`recipient_id`, `body`, `study_id`) match API
- CLI needs to correctly handle 204 No Content (no response body)

**Why this priority**: Core messaging workflow. Incorrect handling
could cause silent failures.

**Independent Test**: Run `prolific message send -r <id> -s <id>
-b "message"` and verify the request payload and 204 response
handling.

**Acceptance Scenarios**:

1. **Given** valid `--recipient`, `--study`, and `--body` flags,
   **When** the user runs `prolific message send`,
   **Then** the CLI sends a POST to `/api/v1/messages/` with
   `recipient_id`, `body`, `study_id` and displays a confirmation.
2. **Given** a required flag is omitted,
   **When** the command executes,
   **Then** a clear error indicates which flag is missing.
3. **Given** the API returns 204 No Content,
   **When** the CLI receives this response,
   **Then** it does not attempt to parse a response body.

---

### User Story 4 - Bulk send messages to multiple participants (Priority: P2)

A researcher needs to send the same message to multiple participants.
This endpoint exists in the API but has no corresponding CLI command.

**API specification** (verified against [bulk-message-participants](https://docs.prolific.com/api-reference/messages/bulk-message-participants)):
- `POST /api/v1/messages/bulk/`
- Request body: `ids` (array of strings, required), `body` (string,
  required), `study_id` (string, required)
- Response: 204 No Content (empty body)

**Why this priority**: Documented API capability needed for batch
communications with study participants.

**Independent Test**: Run `prolific message bulk-send --ids <id1,id2>
-s <study_id> -b "message"` and verify the request and response.

**Acceptance Scenarios**:

1. **Given** a list of participant IDs, study ID, and message body,
   **When** the user runs the bulk send command,
   **Then** the CLI sends a POST to `/api/v1/messages/bulk/` with
   `ids` (array), `body`, and `study_id` and displays a confirmation.
2. **Given** the user provides an empty IDs list,
   **When** the command executes,
   **Then** the CLI returns an error before calling the API.

---

### User Story 5 - Send message to a participant group (Priority: P2)

A researcher needs to message an entire participant group. This
endpoint exists in the API but has no corresponding CLI command.

**API specification** (verified against [send-message-to-participant-group](https://docs.prolific.com/api-reference/messages/send-message-to-participant-group)):
- `POST /api/v1/messages/participant-group/`
- Request body: `participant_group_id` (string, required), `body`
  (string, required), `study_id` (string, optional)
- Response: 204 No Content (empty body)

**Why this priority**: Documented API capability for communicating
with participant cohorts.

**Independent Test**: Run `prolific message send-group
--group <group_id> -b "message"` and verify the request and response.

**Acceptance Scenarios**:

1. **Given** a participant group ID and message body (optionally a
   study ID),
   **When** the user runs the group send command,
   **Then** the CLI sends a POST to
   `/api/v1/messages/participant-group/` with the correct fields.
2. **Given** the required `--group` flag is omitted,
   **When** the command executes,
   **Then** a clear error indicates the missing flag.

---

### Edge Cases

- Empty results array from message list: CLI MUST display headers
  but no rows, not an error.
- `created_after` exceeds 30-day limit: CLI SHOULD relay the API's
  error clearly.
- `bulk-send` with empty IDs list: CLI MUST validate before calling
  the API.
- Message body with special characters or extreme length: CLI MUST
  pass through without truncation (API handles sanitization).
- `send-group` without `--study`: CLI MUST allow it since `study_id`
  is optional for this endpoint (unlike other send endpoints).

## Clarifications

### Session 2026-02-06 (carried from 001 spec)

- Q: Should adding missing API fields to `message list` output
  preserve backward compatibility? → A: Breaking change acceptable.
  The old output was incomplete/incorrect.
- Q: Which columns should `message list` display by default? →
  A: Curated default: `id`, `sender_id`, `study_id`, `category`,
  `sent_at`, `body`. Omit `type` and `channel_id`.
- Q: Should `UnreadMessage` be eliminated or kept? → A: Eliminate
  `UnreadMessage` entirely. Reuse `Message` for both endpoints.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The `Message` model MUST include all fields from the
  API response schema per [get-messages.md](https://docs.prolific.com/api-reference/messages/get-messages.md):
  `id` (string), `sender_id` (string), `body` (string),
  `sent_at` (date-time), `type` (string), `channel_id` (string),
  and `data` (object with `study_id` string and `category` enum).
- **FR-002**: The `UnreadMessage` struct MUST be eliminated. Both
  `/messages/` and `/messages/unread/` return the same `Message`
  schema per [get-unread-messages.md](https://docs.prolific.com/api-reference/messages/get-unread-messages.md).
  The `ListUnreadMessagesResponse` MUST use `[]model.Message`.
- **FR-003**: The `message list` command MUST display these columns
  by default: `id`, `sender_id`, `study_id`, `category`, `sent_at`,
  `body`. Fields `type` and `channel_id` MUST be in the domain model
  but omitted from default table output. This is a breaking change.
- **FR-004**: The `message list` command MUST enforce that at least
  one of `--user` or `--created_after` is provided (matching the
  API's conditional requirement).
- **FR-005**: The `message send` command MUST handle 204 No Content
  responses without attempting to parse a response body, per
  [send-message.md](https://docs.prolific.com/api-reference/messages/send-message.md).
- **FR-006**: A `message bulk-send` command MUST be implemented for
  `POST /api/v1/messages/bulk/` per [bulk-message-participants](https://docs.prolific.com/api-reference/messages/bulk-message-participants).
  Required fields: `ids` (array of strings), `body` (string),
  `study_id` (string). Response: 204 No Content.
- **FR-007**: A `message send-group` command MUST be implemented for
  `POST /api/v1/messages/participant-group/` per
  [send-message-to-participant-group](https://docs.prolific.com/api-reference/messages/send-message-to-participant-group).
  Required fields: `participant_group_id` (string), `body` (string).
  Optional: `study_id` (string). Response: 204 No Content.
- **FR-008**: All new commands MUST follow the existing dependency
  injection pattern (`client.API` + `io.Writer`).
- **FR-009**: All new commands MUST have unit tests using the
  `gomock` mock client pattern.
- **FR-010**: The `client.API` interface MUST be updated to include
  methods for the bulk send and group send endpoints.

### Key Entities

- **Message**: Per [get-messages.md](https://docs.prolific.com/api-reference/messages/get-messages.md) —
  `id` (string), `sender_id` (string), `body` (string),
  `sent_at` (date-time), `type` (string), `channel_id` (string),
  `data` (object: `study_id` string, `category` enum).
- **SendMessagePayload**: Per [send-message.md](https://docs.prolific.com/api-reference/messages/send-message.md) —
  `recipient_id` (string), `body` (string), `study_id` (string).
  All required.
- **BulkSendPayload**: Per [bulk-message-participants](https://docs.prolific.com/api-reference/messages/bulk-message-participants) —
  `ids` (array of strings), `body` (string), `study_id` (string).
  All required.
- **GroupSendPayload**: Per [send-message-to-participant-group](https://docs.prolific.com/api-reference/messages/send-message-to-participant-group) —
  `participant_group_id` (string, required), `body` (string,
  required), `study_id` (string, optional).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Every field in the Prolific Messages API response
  schema is represented in the CLI's domain model with matching
  JSON field names as documented in the API reference.
- **SC-002**: All 5 documented message API endpoints have
  corresponding CLI commands and client methods.
- **SC-003**: All existing message command tests pass after model
  and client updates.
- **SC-004**: All new commands (bulk-send, send-group) have unit
  tests covering happy path, error responses, and flag validation.
- **SC-005**: `make test` and `make lint` pass with zero failures
  and zero suppressed warnings.
- **SC-006**: Running each message command with `--help` displays
  accurate usage matching the API documentation.

### Assumptions

- The Prolific API documentation at the 5 URLs listed in the API
  Reference section above is authoritative and current.
- The API `sent_at` field corresponds to what the current CLI maps
  as `datetime_created`. This MUST be corrected.
- The `UnreadMessage` struct's `sender` field corresponds to the
  API's `sender_id` field. `UnreadMessage` will be removed.
- The API 204 No Content response for all send operations means the
  client MUST NOT attempt to decode a JSON response body.
- The `data.category` enum values are: `payment-timing`,
  `payment-issues`, `technical-issues`, `feedback`, `rejections`,
  `other`.
