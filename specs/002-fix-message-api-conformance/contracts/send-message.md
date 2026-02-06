# Contract: Send a Message

**Source**: [send-message.md](https://docs.prolific.com/api-reference/messages/send-message.md)
**CLI Command**: `prolific message send`

## Endpoint

`POST /api/v1/messages/`

## Request Body

```json
{
  "recipient_id": "619e049f7648a4e1f8f3645b",
  "body": "Thanks for participating in my study",
  "study_id": "719e049f7648a4e1f8f3645a"
}
```

| Field | Type | Required |
|-------|------|----------|
| `recipient_id` | string | Yes |
| `body` | string | Yes |
| `study_id` | string | Yes |

## Response

**204 No Content**: Empty body. Message sent successfully.
**400**: Error response.

## CLI Flags

| Flag | Short | Maps to | Required |
|------|-------|---------|----------|
| `--recipient` | `-r` | `recipient_id` | Yes |
| `--body` | `-b` | `body` | Yes |
| `--study` | `-s` | `study_id` | Yes |

## Confirmation Output

On success, display a table with the user-provided values:
`Recipient ID`, `Study ID`, `Body`
