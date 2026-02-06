# Contract: Retrieve Messages

**Source**: [get-messages.md](https://docs.prolific.com/api-reference/messages/get-messages.md)
**CLI Command**: `prolific message list`

## Endpoint

`GET /api/v1/messages/`

## Query Parameters

| Parameter | Type | Required | Notes |
|-----------|------|----------|-------|
| `user_id` | string | Conditional | Required if `created_after` not provided |
| `created_after` | string (ISO8601) | Conditional | Required if `user_id` not provided. Max 30 days. |

## Response (200)

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

Returned when neither `user_id` nor `created_after` is provided,
or when `created_after` exceeds the 30-day limit.

## CLI Flags

| Flag | Short | Maps to | Required |
|------|-------|---------|----------|
| `--user` | `-u` | `user_id` query param | Conditional |
| `--created_after` | `-c` | `created_after` query param | Conditional |
| `--unread` | `-U` | Switches to unread endpoint | Mutually exclusive |

## Display Columns

`ID`, `Sender ID`, `Study ID`, `Category`, `Sent At`, `Body`
