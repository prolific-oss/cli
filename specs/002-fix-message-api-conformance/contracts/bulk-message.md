# Contract: Send Message to Multiple Participants

**Source**: [bulk-message-participants](https://docs.prolific.com/api-reference/messages/bulk-message-participants)
**CLI Command**: `prolific message bulk-send`

## Endpoint

`POST /api/v1/messages/bulk/`

## Request Body

```json
{
  "ids": ["619e049f7648a4e1f8f3645b", "619e049f7648a4e1f8f3645c"],
  "body": "Thanks for participating in my study",
  "study_id": "6569ece7ca177d19117b1b95"
}
```

| Field | Type | Required |
|-------|------|----------|
| `ids` | array of strings | Yes |
| `body` | string | Yes |
| `study_id` | string | Yes |

## Response

**204 No Content**: Empty body. Message sent successfully.
**400**: Error response.

## CLI Flags

| Flag | Short | Maps to | Required |
|------|-------|---------|----------|
| `--ids` | `-i` | `ids` (comma-separated → array) | Yes |
| `--body` | `-b` | `body` | Yes |
| `--study` | `-s` | `study_id` | Yes |

## CLI Validation

- `--ids` MUST contain at least one value.
- Comma-separated string is split into an array before sending.

## Confirmation Output

On success, display: `Recipients: <count>`, `Study ID`, `Body`
