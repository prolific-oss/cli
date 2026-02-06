# Quickstart: Verifying Message API Conformance

## Prerequisites

```bash
# Build the CLI
make build

# Run all tests (should pass before AND after changes)
make test

# Run linter
make lint
```

## Verification Steps

### 1. Model Verification

After modifying `model/model.go`, verify the `Message` struct has
all required fields:

```bash
# Check Message struct fields
grep -A 10 'type Message struct' model/model.go

# Verify UnreadMessage is removed
grep 'type UnreadMessage struct' model/model.go
# Expected: no output (struct deleted)

# Check MessageData struct exists
grep -A 5 'type MessageData struct' model/model.go
```

### 2. Test Verification

```bash
# Run all tests
make test

# Run only message tests
go test ./cmd/message/... -v

# Run with race detection
go test ./cmd/message/... -race -v
```

### 3. Mock Regeneration

After changing `client/client.go` interface:

```bash
make test-gen-mock
make test  # Verify regenerated mocks work
```

### 4. Command Help Verification

```bash
# Verify all message subcommands are registered
./prolific message --help

# Verify list command help
./prolific message list --help

# Verify send command help
./prolific message send --help

# Verify new bulk-send command
./prolific message bulk-send --help

# Verify new send-group command
./prolific message send-group --help
```

### 5. Live API Verification (requires PROLIFIC_TOKEN)

```bash
export PROLIFIC_TOKEN="your-token"

# List messages by user
./prolific message list -u <user_id>
# Expected columns: ID, Sender ID, Study ID, Category, Sent At, Body

# List unread messages
./prolific message list -U
# Expected: same column layout as above

# Send a message
./prolific message send -r <recipient_id> -s <study_id> -b "Test"
# Expected: confirmation table

# Bulk send (if you have multiple participant IDs)
./prolific message bulk-send -i <id1>,<id2> -s <study_id> -b "Test"
# Expected: confirmation with recipient count

# Send to group (if you have a participant group)
./prolific message send-group -g <group_id> -b "Test"
# Expected: confirmation table
```

### 6. Edge Case Verification

```bash
# No filter provided (should error)
./prolific message list
# Expected: error about required filter

# Unread with other flags (should error)
./prolific message list -U -u <id>
# Expected: error about mutually exclusive flags

# Bulk send with no IDs (should error)
./prolific message bulk-send -s <study_id> -b "Test"
# Expected: error about required --ids flag

# Send group without study (should work)
./prolific message send-group -g <group_id> -b "Test"
# Expected: success (study_id is optional)
```

### 7. Full Validation

```bash
make all  # clean, install, build, test
make lint
```
