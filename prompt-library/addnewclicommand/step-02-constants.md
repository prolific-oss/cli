# Step 2: Add Constants

## Purpose

Define error message constants that will be used by the command implementation. This is done early because:
- Pure constants with no dependencies
- Doesn't break any existing code
- Will be referenced by command validation in Step 6
- Centralizes error messages for consistency

## AI Implementation Prompt

```
I need to add error message constants for a new AITaskBuilder command.

Add constants to `cmd/aitaskbuilder/constants.go` following these requirements:
- Add within the existing const block
- Use Err prefix for all error constants
- Use PascalCase for constant names
- Use lowercase, descriptive messages (no punctuation)
- Follow the pattern of existing constants

Constants needed:
- Required field validation error (e.g., "thing ID is required")
- Not found error (e.g., "thing not found")

Resource name: [SPECIFY THE RESOURCE]
```

## AI Implementation Guidance

### Naming Convention

**Pattern:** `Err` + `Resource` + `ValidationType`

```go
const (
    ErrThingIDRequired   = "thing ID is required"
    ErrThingNotFound     = "thing not found"
)
```

### Where to Add

Add to the existing const block in `cmd/aitaskbuilder/constants.go`:

```go
const (
    // Error messages
    ErrBatchIDRequired   = "batch ID is required"
    ErrBatchNotFound     = "batch not found"
    ErrDatasetIDRequired = "dataset ID is required"
    ErrThingIDRequired   = "thing ID is required"     // ← Add here
    ErrThingNotFound     = "thing not found"          // ← Add here
)
```

### Message Format Rules

**✅ CORRECT:**
```go
ErrThingIDRequired   = "thing ID is required"
ErrThingNotFound     = "thing not found"
ErrWorkspaceIDRequired = "workspace ID is required"
```

**❌ INCORRECT:**
```go
ErrThingIDRequired   = "Thing ID is required"    // Capitalized
ErrThingIDRequired   = "thing ID is required."   // Has punctuation
ErrThingIDRequired   = "Error: thing not found"  // Has prefix
ErrThingIDRequired   = "THING_NOT_FOUND"         // Wrong style
```

### What NOT to Do

❌ **Don't:**
- Create a new const block (add to existing)
- Use lowercase variable names (won't be exported)
- Capitalize error messages
- Add punctuation to messages
- Use redundant "Err" suffix (e.g., `ErrThingNotFoundErr`)

✅ **Do:**
- Add to existing const block
- Use PascalCase names with Err prefix
- Use lowercase messages
- Keep messages concise and clear
- Alphabetize within related constants (optional but nice)

## Human Review Criteria

### Code Review Checklist

- [ ] **Location**: Added in `cmd/aitaskbuilder/constants.go` within existing const block
- [ ] **Naming**: Starts with `Err`, uses PascalCase, descriptive
- [ ] **Value format**: Lowercase message, no punctuation, user-friendly
- [ ] **Consistency**: Matches style of existing constants
- [ ] **Completeness**: Has both required field and not found errors (if applicable)

### Specific Checks

**Constant naming:**
```go
// ✅ CORRECT
ErrThingIDRequired   = "thing ID is required"
ErrThingNotFound     = "thing not found"

// ❌ INCORRECT
ThingIDRequired      = "thing ID is required"    // Missing "Err" prefix
errThingIDRequired   = "thing ID is required"    // Not exported (lowercase)
ErrThingNotFoundErr  = "thing not found"         // Redundant "Err" suffix
ErrThing             = "invalid thing"           // Too vague
```

**Message format:**
```go
// ✅ CORRECT
"thing ID is required"
"thing not found"
"workspace ID is required"

// ❌ INCORRECT
"Thing ID is required"           // Capitalized
"thing ID is required."          // Has punctuation
"Error: thing not found"         // Has prefix
"The thing ID field is required" // Too verbose
```

## Verification Commands

```bash
# 1. Verify file compiles
go build ./cmd/aitaskbuilder/...

# 2. Check constants exist
grep "ErrThing" cmd/aitaskbuilder/constants.go

# 3. Verify they're in the const block
grep -A 10 "const (" cmd/aitaskbuilder/constants.go
```

## Expected Results

**1. Build output:**
```
✅ Build succeeds with no errors
```

**2. Grep output:**
```
✅ Shows the new constants:
    ErrThingIDRequired   = "thing ID is required"
    ErrThingNotFound     = "thing not found"
```

**3. Const block:**
```
✅ Constants appear within existing const block:
const (
    // Error messages
    ErrBatchIDRequired   = "batch ID is required"
    ErrBatchNotFound     = "batch not found"
    ErrDatasetIDRequired = "dataset ID is required"
    ErrThingIDRequired   = "thing ID is required"
    ErrThingNotFound     = "thing not found"
)
```

## Common Issues

### Issue: Not exported
**Cause**: Constant name starts with lowercase  
**Fix**: Change to PascalCase starting with uppercase

### Issue: Message is capitalized
**Cause**: Message starts with capital letter  
**Fix**: Change to lowercase (error messages should be lowercase)

### Issue: Created new const block
**Cause**: Added `const ()` instead of adding to existing  
**Fix**: Remove new block, add to existing const block

### Issue: Wrong format
**Cause**: Message has punctuation or wrong style  
**Fix**: Remove punctuation, follow existing message patterns

## Success Criteria

- [ ] File compiles: `go build ./cmd/aitaskbuilder/...`
- [ ] Constants are within existing const block
- [ ] Names follow `Err` + `PascalCase` pattern
- [ ] Messages are lowercase without punctuation
- [ ] At least 2 constants added (required field, not found)
- [ ] Style matches existing constants exactly

## Next Step

Proceed to [Step 3: Add API Method Signature](step-03-api-interface.md)
