# Contract: Send Message to Participant Group

**Source**: [send-message-to-participant-group](https://docs.prolific.com/api-reference/messages/send-message-to-participant-group)
**CLI Command**: `prolific message send-group`

## Endpoint

`POST /api/v1/messages/participant-group/`

## Request Body

```json
{
  "participant_group_id": "619e049f7648a4e1f8f3645b",
  "body": "Thanks for participating in my study",
  "study_id": "6569ece7ca177d19117b1b95"
}
```

| Field | Type | Required |
|-------|------|----------|
| `participant_group_id` | string | Yes |
| `body` | string | Yes |
| `study_id` | string | No |

## Response

**204 No Content**: Empty body. Message sent successfully.
**400**: Error response.

## CLI Flags

| Flag | Short | Maps to | Required |
|------|-------|---------|----------|
| `--group` | `-g` | `participant_group_id` | Yes |
| `--body` | `-b` | `body` | Yes |
| `--study` | `-s` | `study_id` | No |

## CLI Validation

- `--group` and `--body` are required.
- `--study` is optional (unlike other send endpoints).

## Confirmation Output

On success, display: `Group ID`, `Study ID` (or "N/A"), `Body`
