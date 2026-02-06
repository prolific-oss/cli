# Contract: Retrieve Unread Messages

**Source**: [get-unread-messages.md](https://docs.prolific.com/api-reference/messages/get-unread-messages.md)
**CLI Command**: `prolific message list -U`

## Endpoint

`GET /api/v1/messages/unread/`

## Query Parameters

None.

## Response (200)

Same `Message` schema as the retrieve messages endpoint:

```json
{
  "results": [
    {
      "id": "string",
      "sender_id": "string",
      "body": "string",
      "sent_at": "2026-01-15T10:30:00Z",
      "type": "message",
      "channel_id": "string",
      "data": {
        "study_id": "string",
        "category": "feedback"
      }
    }
  ]
}
```

## Error (400)

Error response (schema not detailed in API docs).

## CLI Behavior

- Accessed via `prolific message list -U` (the `--unread` flag).
- Mutually exclusive with `--user` and `--created_after`.
- Retrieves only received, unread messages.
- Does NOT mark messages as read upon retrieval.

## Display Columns

Same curated columns as regular message list:
`ID`, `Sender ID`, `Study ID`, `Category`, `Sent At`, `Body`
