# Step 8: Run Linter

## Purpose

Run the linter to catch code quality and style issues before writing tests. This step:
- Enforces project code style standards
- Catches common mistakes early
- Ensures code follows Go best practices

## AI Implementation Prompt

```
I need to run the linter to check code quality for the new command.

Run: make lint

If the linter reports any issues:
1. Review each issue carefully
2. Fix according to Go best practices and project conventions
3. Re-run make lint until clean
4. Don't proceed until zero errors/warnings

Common issues to fix:
- Missing doc comments on exported functions
- Error strings should not be capitalized
- Unused imports or variables
- Ineffectual assignments
```

## AI Implementation Guidance

### Command to Run

```bash
make lint
```

### Common Linting Issues and Fixes

**1. Missing comments on exported functions:**

```go
// ❌ FAIL
func NewGetNewThingCommand(client client.API, w io.Writer) *cobra.Command {

// ✅ PASS
// NewGetNewThingCommand creates a new command for getting a thing
func NewGetNewThingCommand(client client.API, w io.Writer) *cobra.Command {
```

**2. Error strings should not be capitalized:**

```go
// ❌ FAIL
return errors.New("Thing ID is required")

// ✅ PASS
return errors.New("thing ID is required")
```

**3. Unused imports:**

```go
// ❌ FAIL
import (
    "fmt"
    "io"
    "strings"  // ← Not used anywhere
)

// ✅ PASS
import (
    "fmt"
    "io"
)
```

**4. Unused variables:**

```go
// ❌ FAIL
func render(...) error {
    response, err := c.GetAPI()
    thing := response.Thing
    // 'thing' declared but never used
}

// ✅ PASS
func render(...) error {
    response, err := c.GetAPI()
    thing := response.Thing
    fmt.Fprintf(w, "ID: %s\n", thing.ID)  // Now it's used
}
```

**5. Ineffectual assignment:**

```go
// ❌ FAIL
func render(...) error {
    err := validateOptions(opts)
    response, err := c.GetAPI()  // 'err' reassigned before use
}

// ✅ PASS
func render(...) error {
    if err := validateOptions(opts); err != nil {
        return err
    }
    response, err := c.GetAPI()
}
```

**6. ST1003: Struct field name (acceptable warning):**

```go
// This warning for Options structs is OK - follows project conventions
// Can be ignored: "field ThingID should be ThingId"
```

### What NOT to Do

❌ **Don't:**
- Skip linting ("I'll fix it later")
- Ignore linting errors to proceed faster
- Remove useful code just to pass linting
- Change meaningful variable names just for linting

✅ **Do:**
- Fix all linting errors before proceeding
- Address the root cause, not just the symptom
- Follow Go best practices
- Keep code readable while meeting standards

## Human Review Criteria

### Verification Commands

```bash
# 1. Run linter
make lint

# 2. If issues found, check specific file
golangci-lint run ./cmd/aitaskbuilder/get_new_thing.go

# 3. Verify fixed (after AI fixes issues)
make lint

# 4. Double-check build still works
make build
```

### Expected Results

**Clean output:**
```
✅ Linting completed
✅ No errors
✅ No warnings (or only acceptable warnings)
```

**With issues (before fixes):**
```
❌ cmd/aitaskbuilder/get_new_thing.go:XX:YY: <description>
```

**After fixes:**
```
✅ No linting errors
✅ All issues resolved
```

### Code Review Checklist

- [ ] **Lint runs**: `make lint` executes without errors
- [ ] **Zero errors**: No errors reported
- [ ] **Comments**: All exported functions have doc comments
- [ ] **Error messages**: All lowercase, no capitalization
- [ ] **Imports**: No unused imports
- [ ] **Variables**: No unused variables
- [ ] **Build**: Still compiles after fixes

### Review Each Fix

For each linting issue fixed, verify:
- [ ] Fix addresses root cause
- [ ] Fix follows project conventions
- [ ] Fix doesn't break functionality
- [ ] Code remains readable and maintainable

## Common Issues

### Issue: Comment doesn't start with function name
**Linter**: "comment should be of the form 'FunctionName ...'"  
**Fix:**
```go
// ❌ FAIL
// Creates a new command for getting things
func NewGetNewThingCommand(...) {

// ✅ PASS
// NewGetNewThingCommand creates a new command for getting things
func NewGetNewThingCommand(...) {
```

### Issue: Capitalized error string
**Linter**: "error strings should not be capitalized"  
**Fix:**
```go
// ❌ FAIL
errors.New("Required field is missing")

// ✅ PASS
errors.New("required field is missing")
```

### Issue: Unused import
**Linter**: "imported and not used"  
**Fix:** Remove the unused import from import block

### Issue: Ineffectual assignment
**Linter**: "this value of err is never used"  
**Fix:** Check error before reassigning, or use different variable name

### Issue: Struct field naming
**Linter**: "field ThingID should be ThingId"  
**Note:** This is acceptable for Options structs - follows project convention

## Success Criteria

- [ ] `make lint` returns zero errors
- [ ] No warnings (or only acceptable ones)
- [ ] All exported functions have doc comments starting with function name
- [ ] All error messages are lowercase
- [ ] No unused imports or variables
- [ ] Code still compiles: `make build` succeeds
- [ ] Functionality unchanged by linting fixes

## Checkpoint 3: Verification

✅ **CHECKPOINT 3 REACHED**

At this point, the command should be clean and verified:

```bash
# Full verification
make lint
make clean
make build
./prolific aitaskbuilder get-new-thing --help
```

**Expected:** Lint passes, build succeeds, command works correctly.

**If this fails, do not proceed to Step 9. Fix issues first.**

**Recommended:** If you have API access, manually test the command now.

## Next Step

Once Checkpoint 3 passes, proceed to [Step 9: Create Test File](step-09-test-file.md)
