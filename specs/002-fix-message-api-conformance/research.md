# Research: Fix Message Commands — Prolific API Conformance

## 1. Message Model Field Mapping

**Decision**: Replace `Message` struct fields to match API schema exactly.

**Rationale**: The API documentation at [get-messages.md](https://docs.prolific.com/api-reference/messages/get-messages.md)
defines the canonical field names. The current CLI model uses non-standard
names that don't match the JSON response, causing either silent data loss
or brittle post-processing workarounds.

**Current → Target mapping**:

| Current CLI Field | JSON Tag | API Field | JSON Tag | Action |
|-------------------|----------|-----------|----------|--------|
| `DatetimeCreated` | `datetime_created` | `SentAt` | `sent_at` | RENAME |
| `Body` | `body` | `Body` | `body` | KEEP |
| `SenderID` | `sender_id` | `SenderID` | `sender_id` | KEEP |
| `StudyID` | `study_id,omitempty` | — | — | REMOVE (now in Data) |
| `Data` | `data,omitempty` | `Data` | `data,omitempty` | RESTRUCTURE (typed struct) |
| — | — | `ID` | `id` | ADD |
| — | — | `Type` | `type` | ADD |
| — | — | `ChannelID` | `channel_id` | ADD |

**Alternatives considered**:
- Keep `datetime_created` and add an alias: Rejected because JSON
  unmarshaling would silently drop `sent_at` responses.
- Add fields without removing old ones: Rejected because it creates
  ambiguity about which field is canonical.

## 2. UnreadMessage Struct Elimination

**Decision**: Remove `UnreadMessage` struct entirely. Use `Message`
for both endpoints.

**Rationale**: The [get-unread-messages.md](https://docs.prolific.com/api-reference/messages/get-unread-messages.md)
endpoint returns the identical `Message` schema as the regular messages
endpoint. Maintaining a separate struct with different field names
(`sender` vs `sender_id`) is incorrect and causes inconsistent output.

**Impact analysis**:
- `model/model.go`: Delete `UnreadMessage` struct (lines 78-83)
- `client/responses.go`: Change `ListUnreadMessagesResponse.Results`
  from `[]model.UnreadMessage` to `[]model.Message`
- `cmd/message/list.go`: Unify the unread rendering path to use the
  same `Message` fields and table columns as the regular path
- `cmd/message/list_test.go`: Update mock expectations and assertions

**Alternatives considered**:
- Type alias (`type UnreadMessage = Message`): Rejected because it
  adds indirection with no benefit.
- Keep both structs with identical fields: Rejected because it creates
  maintenance burden for no gain.

## 3. GetMessages Post-Processing Removal

**Decision**: Remove the `study_id` extraction loop in `GetMessages`
(`client/client.go:660-667`).

**Rationale**: The current code iterates over results, extracts
`study_id` from the `Data` map, promotes it to a top-level field, then
nullifies `Data`. This was a workaround for the model not having a
typed `Data` struct. With the new `MessageData` sub-struct, JSON
unmarshaling will handle this natively — `data.study_id` maps directly
to `Data.StudyID`.

**Code to remove**:
```go
for index, message := range response.Results {
    if value, ok := message.Data["study_id"].(string); ok {
        response.Results[index].StudyID = value
    }
    response.Results[index].Data = nil
}
```

**Alternatives considered**:
- Keep the loop for backward compatibility: Rejected because the new
  typed struct makes it unnecessary and it actively destroys `category`
  data.

## 4. MessageData Sub-Struct Design

**Decision**: Create a typed `MessageData` struct instead of using
`map[string]any`.

**Rationale**: The API defines `data` as an object with known fields
(`study_id` string, `category` enum). A typed struct provides compile-
time safety, proper JSON unmarshaling, and makes `category` accessible
to the CLI output.

**Category enum values** (from API docs): `payment-timing`,
`payment-issues`, `technical-issues`, `feedback`, `rejections`, `other`.

**Alternatives considered**:
- Keep `map[string]any` and access fields by key: Rejected because
  it's error-prone and doesn't surface `category` properly.
- Use a separate `Category` type with constants: Considered but
  unnecessary for display purposes — string is sufficient.

## 5. New Client Methods

**Decision**: Add `BulkSendMessage` and `SendGroupMessage` to the
`client.API` interface.

**Rationale**: Two documented API endpoints have no CLI implementation:
- `POST /api/v1/messages/bulk/` ([bulk-message-participants](https://docs.prolific.com/api-reference/messages/bulk-message-participants))
- `POST /api/v1/messages/participant-group/` ([send-message-to-participant-group](https://docs.prolific.com/api-reference/messages/send-message-to-participant-group))

Both return 204 No Content, so the client methods return `error` only
(no response struct), matching the existing `SendMessage` pattern.

**Method signatures**:
```go
BulkSendMessage(ids []string, body, studyID string) error
SendGroupMessage(participantGroupID, body string, studyID *string) error
```

Note: `SendGroupMessage` takes `studyID` as `*string` because the
API makes it optional for this endpoint only.

**Alternatives considered**:
- Single generic `SendMessage` with variadic options: Rejected for
  simplicity — the existing CLI uses distinct methods per endpoint.

## 6. Table Output Column Selection

**Decision**: Display curated columns: `ID`, `Sender ID`, `Study ID`,
`Category`, `Sent At`, `Body`.

**Rationale**: Per clarification session, the full API schema has 8+
fields. Displaying all of them exceeds typical terminal width. The
selected 6 columns provide the most useful information. `type` (always
"message") and `channel_id` (internal thread reference) are omitted
from display but present in the model.

This is a breaking change from the current 3-4 column layout. Per
clarification, backward compatibility is not required since the old
output was incorrect.

## 7. 204 No Content Handling

**Decision**: All three send commands (`send`, `bulk-send`,
`send-group`) display user-provided input as confirmation after a
successful 204 response.

**Rationale**: The existing `send` command already does this
(`send.go:66-74`). The same pattern will be used for `bulk-send`
(showing count of IDs + body) and `send-group` (showing group ID +
body).

The client's `Execute` method passes `nil` as the response struct
pointer for send operations, so it correctly skips JSON decoding.
No changes needed to the Execute method.
