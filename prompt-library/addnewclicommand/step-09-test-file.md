# Step 9: Create Test File

## Purpose

Create comprehensive tests for the new command. This step:
- Verifies command behavior
- Tests output formatting
- Validates error handling
- Ensures required field validation works
- Provides confidence in code quality

## AI Implementation Prompt

```
I need to create comprehensive tests for the new command.

Create `cmd/aitaskbuilder/get_<resource>_test.go` with 5 test functions:
1. TestNewGet<Resource>Command - Basic command structure
2. TestNewGet<Resource>CommandCallsAPI - Successful API call with output
3. TestNewGet<Resource>CommandCallsAPIWithoutOptionalFields - Optional fields test
4. TestNewGet<Resource>CommandHandlesErrors - Error handling
5. TestNewGet<Resource>CommandRequires<Field> - Required field validation

Resource name: [SPECIFY]
Required field: [SPECIFY, e.g., "thingID"]
Response fields: [SPECIFY what to include in mock]

Critical patterns:
- Package: aitaskbuilder_test (with _test suffix)
- Use gomock for mocking
- Use bufio.Writer + bytes.Buffer for output capture
- MUST call writer.Flush() before assertions
- Expected output must match exactly (including whitespace)
```

## AI Implementation Guidance

### File Naming

**✅ CORRECT:**
- `get_new_thing_test.go`
- `get_batch_status_test.go`

**❌ INCORRECT:**
- `get-new-thing_test.go` (dashes)
- `get_new_thing_tests.go` (plural)
- `test_get_new_thing.go` (wrong order)

### File Structure

```go
package aitaskbuilder_test  // ← Note _test suffix

import (
    "bufio"
    "bytes"
    "errors"
    "fmt"
    "os"
    "testing"
    "time"
    
    "github.com/golang/mock/gomock"
    "github.com/prolific-oss/cli/client"
    "github.com/prolific-oss/cli/cmd/aitaskbuilder"
    "github.com/prolific-oss/cli/mock_client"
    "github.com/prolific-oss/cli/model"
)
```

### Test 1: Basic Command Structure

**Purpose:** Verify command is created correctly

```go
func TestNewGetNewThingCommand(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()
    c := mock_client.NewMockAPI(ctrl)
    
    cmd := aitaskbuilder.NewGetNewThingCommand(c, os.Stdout)
    
    use := "get-new-thing"
    short := "Get an AI Task Builder thing"
    
    if cmd.Use != use {
        t.Fatalf("expected use: %s; got %s", use, cmd.Use)
    }
    
    if cmd.Short != short {
        t.Fatalf("expected short: %s; got %s", short, cmd.Short)
    }
}
```

### Test 2: Successful API Call

**Purpose:** Verify API call and output formatting

**CRITICAL:** Must use `bufio.Writer` and call `Flush()`

```go
func TestNewGetNewThingCommandCallsAPI(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()
    c := mock_client.NewMockAPI(ctrl)
    
    thingID := "01954894-65b3-779e-aaf6-348698e23634"
    
    // Parse time for consistent testing
    createdAt, _ := time.Parse(time.RFC3339, "2025-02-27T18:03:59.795Z")
    
    // Build response
    response := client.GetAITaskBuilderNewThingResponse{
        AITaskBuilderNewThing: model.AITaskBuilderNewThing{
            ID:        thingID,
            CreatedAt: createdAt,
            Name:      "Test Thing",
            Status:    "ACTIVE",
            Items: []model.Item{
                {ID: "item1", Name: "First Item"},
            },
        },
    }
    
    // Setup mock expectation
    c.
        EXPECT().
        GetAITaskBuilderNewThing(gomock.Eq(thingID)).
        Return(&response, nil).
        AnyTimes()
    
    // CRITICAL: Use bufio.Writer wrapped around bytes.Buffer
    var b bytes.Buffer
    writer := bufio.NewWriter(&b)
    
    // Execute command
    cmd := aitaskbuilder.NewGetNewThingCommand(c, writer)
    _ = cmd.Flags().Set("thing-id", thingID)
    _ = cmd.RunE(cmd, nil)
    
    // CRITICAL: Must call Flush() before reading buffer
    writer.Flush()
    
    // Expected output - must match EXACTLY (including whitespace)
    expected := `AI Task Builder Thing Details:
ID: 01954894-65b3-779e-aaf6-348698e23634
Name: Test Thing
Status: ACTIVE
Created At: 2025-02-27 18:03:59
Items: 1
  Item 1: First Item
`
    actual := b.String()
    if actual != expected {
        t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
    }
}
```

### Test 3: Optional Fields

**Purpose:** Verify optional fields don't appear when empty

```go
func TestNewGetNewThingCommandCallsAPIWithoutOptionalFields(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()
    c := mock_client.NewMockAPI(ctrl)
    
    thingID := "01954894-65b3-779e-aaf6-348698e23699"
    
    createdAt, _ := time.Parse(time.RFC3339, "2025-02-27T18:03:59.795Z")
    
    // Response with empty optional fields
    response := client.GetAITaskBuilderNewThingResponse{
        AITaskBuilderNewThing: model.AITaskBuilderNewThing{
            ID:        thingID,
            CreatedAt: createdAt,
            Name:      "Simple Thing",
            Status:    "INACTIVE",
            Items:     []model.Item{},  // Empty
        },
    }
    
    c.
        EXPECT().
        GetAITaskBuilderNewThing(gomock.Eq(thingID)).
        Return(&response, nil).
        AnyTimes()
    
    var b bytes.Buffer
    writer := bufio.NewWriter(&b)
    
    cmd := aitaskbuilder.NewGetNewThingCommand(c, writer)
    _ = cmd.Flags().Set("thing-id", thingID)
    _ = cmd.RunE(cmd, nil)
    
    writer.Flush()
    
    // Expected output WITHOUT optional fields
    expected := `AI Task Builder Thing Details:
ID: 01954894-65b3-779e-aaf6-348698e23699
Name: Simple Thing
Status: INACTIVE
Created At: 2025-02-27 18:03:59
`
    actual := b.String()
    if actual != expected {
        t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
    }
}
```

### Test 4: Error Handling

**Purpose:** Verify API errors are handled gracefully

```go
func TestNewGetNewThingCommandHandlesErrors(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()
    c := mock_client.NewMockAPI(ctrl)
    
    thingID := "invalid-thing-id"
    errorMessage := aitaskbuilder.ErrThingNotFound
    
    // Mock returns error
    c.
        EXPECT().
        GetAITaskBuilderNewThing(gomock.Eq(thingID)).
        Return(nil, errors.New(errorMessage)).
        AnyTimes()
    
    cmd := aitaskbuilder.NewGetNewThingCommand(c, os.Stdout)
    _ = cmd.Flags().Set("thing-id", thingID)
    err := cmd.RunE(cmd, nil)
    
    // Error should be wrapped with "error:" prefix
    expected := fmt.Sprintf("error: %s", errorMessage)
    
    if err.Error() != expected {
        t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
    }
}
```

### Test 5: Required Field Validation

**Purpose:** Verify required flags are enforced

```go
func TestNewGetNewThingCommandRequiresThingID(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()
    c := mock_client.NewMockAPI(ctrl)
    
    // Don't set required flag
    cmd := aitaskbuilder.NewGetNewThingCommand(c, os.Stdout)
    err := cmd.RunE(cmd, nil)
    
    if err == nil {
        t.Fatal("expected error when thing-id is missing")
    }
    
    // Check error message
    if !cmd.Flags().Changed("thing-id") {
        expected := aitaskbuilder.ErrThingIDRequired
        if err.Error() != "error: "+expected {
            t.Fatalf("expected error to contain '%s', got '%s'", expected, err.Error())
        }
    }
}
```

### Critical Patterns

**Output capture (MOST IMPORTANT):**
```go
// ✅ CORRECT
var b bytes.Buffer
writer := bufio.NewWriter(&b)  // ← Wrap in bufio.Writer

cmd := aitaskbuilder.NewGetNewThingCommand(c, writer)
_ = cmd.RunE(cmd, nil)

writer.Flush()  // ← CRITICAL: Must call before reading
actual := b.String()

// ❌ WRONG - Missing bufio.Writer
var b bytes.Buffer
cmd := aitaskbuilder.NewGetNewThingCommand(c, &b)

// ❌ WRONG - Not calling Flush()
writer := bufio.NewWriter(&b)
cmd := aitaskbuilder.NewGetNewThingCommand(c, writer)
_ = cmd.RunE(cmd, nil)
actual := b.String()  // Will be empty!
```

**Mock expectations:**
```go
// ✅ CORRECT
c.
    EXPECT().
    GetAITaskBuilderNewThing(gomock.Eq(thingID)).
    Return(&response, nil).
    AnyTimes()

// ❌ WRONG - Not using gomock.Eq
c.EXPECT().GetAITaskBuilderNewThing(thingID).Return(&response, nil)

// ❌ WRONG - Missing AnyTimes()
c.EXPECT().GetAITaskBuilderNewThing(gomock.Eq(thingID)).Return(&response, nil)
```

**Expected output:**
```go
// ✅ CORRECT - Exact match with trailing newline
expected := `AI Task Builder Thing Details:
ID: 01954894-65b3-779e-aaf6-348698e23634
Name: Test Thing
`  // ← Note trailing newline

// ❌ WRONG - Missing trailing newline
expected := `AI Task Builder Thing Details:
ID: 01954894-65b3-779e-aaf6-348698e23634
Name: Test Thing`

// ❌ WRONG - Extra space
expected := `ID:  Test Thing`  // Two spaces after colon
```

### What NOT to Do

❌ **Don't:**
- Forget `writer.Flush()` before assertions
- Use `bytes.Buffer` directly without `bufio.Writer`
- Approximate expected output strings
- Forget trailing newlines in expected strings
- Miss whitespace differences

✅ **Do:**
- Always use `bufio.Writer` wrapped around `bytes.Buffer`
- Always call `writer.Flush()` before reading buffer
- Match expected output character-for-character
- Include all whitespace exactly
- Use `gomock.Eq()` for parameter matching

## Human Review Criteria

### File and Package Check

```bash
# Verify file exists
ls -la cmd/aitaskbuilder/get_new_thing_test.go

# Check package declaration
head -1 cmd/aitaskbuilder/get_new_thing_test.go

# Count test functions
grep -c "^func Test" cmd/aitaskbuilder/get_new_thing_test.go
```

### Expected Results

**File exists:**
```
✅ -rw-r--r-- ... cmd/aitaskbuilder/get_new_thing_test.go
```

**Package:**
```
✅ package aitaskbuilder_test
```

**Test count:**
```
✅ 5
```

### Code Review Checklist

- [ ] **Package**: `package aitaskbuilder_test` (with _test suffix)
- [ ] **Test count**: 5 test functions present
- [ ] **Imports**: All necessary imports included
- [ ] **Test 1**: Checks Use and Short fields
- [ ] **Test 2**: Uses bufio.Writer, calls Flush(), exact output match
- [ ] **Test 3**: Tests optional fields case
- [ ] **Test 4**: Tests error handling with wrapped error
- [ ] **Test 5**: Tests required field validation
- [ ] **Mock usage**: Uses gomock.Eq() and AnyTimes()
- [ ] **Compilable**: `go test -c ./cmd/aitaskbuilder/` succeeds

### Verification Commands

```bash
# 1. Verify tests compile (don't run yet)
go test -c ./cmd/aitaskbuilder/

# 2. List test function names
grep "^func Test" cmd/aitaskbuilder/get_new_thing_test.go

# 3. Check for writer.Flush() usage
grep "Flush()" cmd/aitaskbuilder/get_new_thing_test.go

# 4. Verify bufio.Writer usage
grep "bufio.NewWriter" cmd/aitaskbuilder/get_new_thing_test.go
```

## Common Issues

### Issue: Package name wrong
**Symptom**: `package aitaskbuilder` instead of `aitaskbuilder_test`  
**Fix**: Add `_test` suffix to package declaration

### Issue: Missing writer.Flush()
**Symptom**: Tests fail with empty output  
**Fix**: Add `writer.Flush()` before `b.String()`

### Issue: Not using bufio.Writer
**Symptom**: Output doesn't capture correctly  
**Fix**: Wrap bytes.Buffer in bufio.NewWriter()

### Issue: Expected output mismatch
**Symptom**: Test fails with whitespace differences  
**Fix**: Copy exact output from implementation, include all whitespace

### Issue: Missing trailing newline
**Symptom**: Test fails with difference at end  
**Fix**: Add newline at end of expected string

### Issue: Not using gomock.Eq()
**Symptom**: Mock not called, panic about unexpected call  
**Fix**: Wrap parameters in `gomock.Eq()`

## Success Criteria

- [ ] File named correctly: `get_<resource>_test.go`
- [ ] Package: `package aitaskbuilder_test`
- [ ] 5 test functions present
- [ ] All tests use correct patterns
- [ ] Output capture uses `bufio.Writer`
- [ ] All tests call `writer.Flush()`
- [ ] Expected output matches exactly
- [ ] Mock expectations use `gomock.Eq()`
- [ ] Tests compile: `go test -c ./cmd/aitaskbuilder/`

## Next Step

Proceed to [Step 10: Run Full Test Suite](step-10-run-tests.md)
