# Data Model: Fix Message Commands — Prolific API Conformance

## Entity: Message (MODIFIED)

**File**: `model/model.go`
**Source**: [get-messages.md](https://docs.prolific.com/api-reference/messages/get-messages.md)

```go
// Message represents a message on the Prolific platform.
type Message struct {
    ID        string       `json:"id"`
    SenderID  string       `json:"sender_id"`
    Body      string       `json:"body"`
    SentAt    time.Time    `json:"sent_at"`
    Type      string       `json:"type,omitempty"`
    ChannelID string       `json:"channel_id"`
    Data      *MessageData `json:"data,omitempty"`
}

// MessageData contains metadata associated with a message.
type MessageData struct {
    StudyID  string `json:"study_id,omitempty"`
    Category string `json:"category,omitempty"`
}
```

**Changes from current**:
- ADD `ID` (`json:"id"`)
- RENAME `DatetimeCreated` → `SentAt` (`json:"sent_at"`)
- ADD `Type` (`json:"type,omitempty"`)
- ADD `ChannelID` (`json:"channel_id"`)
- RESTRUCTURE `Data` from `map[string]any` → `*MessageData`
- REMOVE top-level `StudyID` (now accessed via `Data.StudyID`)

## Entity: UnreadMessage (DELETED)

**File**: `model/model.go`

The `UnreadMessage` struct is removed. Both `/messages/` and
`/messages/unread/` return the same `Message` schema.

## Entity: MessageData (NEW)

**File**: `model/model.go`

See `MessageData` struct above. The `category` field uses `string`
type (not a Go enum) because the API enum values (`payment-timing`,
`payment-issues`, `technical-issues`, `feedback`, `rejections`,
`other`) contain hyphens which are not valid Go identifiers, and
the CLI only needs to display the value.

## Entity: SendMessagePayload (UNCHANGED)

**File**: `client/payloads.go`
**Source**: [send-message.md](https://docs.prolific.com/api-reference/messages/send-message.md)

```go
type SendMessagePayload struct {
    RecipientID string `json:"recipient_id"`
    StudyID     string `json:"study_id"`
    Body        string `json:"body"`
}
```

No changes needed — already matches the API.

## Entity: BulkSendMessagePayload (NEW)

**File**: `client/payloads.go`
**Source**: [bulk-message-participants](https://docs.prolific.com/api-reference/messages/bulk-message-participants)

```go
// BulkSendMessagePayload represents the JSON payload for
// sending a message to multiple participants.
type BulkSendMessagePayload struct {
    IDs     []string `json:"ids"`
    Body    string   `json:"body"`
    StudyID string   `json:"study_id"`
}
```

## Entity: SendGroupMessagePayload (NEW)

**File**: `client/payloads.go`
**Source**: [send-message-to-participant-group](https://docs.prolific.com/api-reference/messages/send-message-to-participant-group)

```go
// SendGroupMessagePayload represents the JSON payload for
// sending a message to a participant group.
type SendGroupMessagePayload struct {
    ParticipantGroupID string `json:"participant_group_id"`
    Body               string `json:"body"`
    StudyID            string `json:"study_id,omitempty"`
}
```

Note: `StudyID` uses `omitempty` because it is optional for this
endpoint (unlike the other send endpoints where it is required).

## Response Types (MODIFIED)

**File**: `client/responses.go`

### ListMessagesResponse (UNCHANGED)
```go
type ListMessagesResponse struct {
    Results []model.Message `json:"results"`
    *JSONAPILinks
    *JSONAPIMeta
}
```

### ListUnreadMessagesResponse (MODIFIED)
```go
type ListUnreadMessagesResponse struct {
    Results []model.Message `json:"results"`  // Was: []model.UnreadMessage
    *JSONAPILinks
    *JSONAPIMeta
}
```

## Client Interface Methods

**File**: `client/client.go`

### Existing (UNCHANGED signature, MODIFIED implementation):
```go
GetMessages(userID *string, createdAfter *string) (*ListMessagesResponse, error)
SendMessage(body, recipientID, studyID string) error
GetUnreadMessages() (*ListUnreadMessagesResponse, error)
```

`GetMessages` implementation change: remove the `study_id` extraction
loop (lines 660-667) since the new typed `MessageData` struct handles
this via JSON unmarshaling.

### New methods:
```go
BulkSendMessage(ids []string, body, studyID string) error
SendGroupMessage(participantGroupID, body string, studyID *string) error
```
