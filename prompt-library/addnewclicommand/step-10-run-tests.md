# Step 10: Run Full Test Suite

## Purpose

Run all tests to verify the new command works correctly and doesn't break existing functionality. This step:
- Validates test implementation
- Ensures output formatting is correct
- Verifies error handling works
- Confirms no regressions
- Provides confidence before completion

## AI Implementation Prompt

```
I need to run the full test suite to verify everything works.

Run: make test

If any tests fail:
1. Run verbose mode to see details: go test -v ./cmd/aitaskbuilder/
2. Identify which test is failing
3. Read the error message carefully (especially output diffs)
4. Common issues:
   - Whitespace mismatch in expected output
   - Missing writer.Flush()
   - Date format mismatch
   - Error message format mismatch
5. Fix the issue
6. Re-run tests
7. Repeat until all pass

Do not proceed to Step 11 until all tests pass.
```

## AI Implementation Guidance

### Commands to Run

```bash
# 1. Run all tests
make test

# 2. If failures, run verbose for details
go test -v ./cmd/aitaskbuilder/

# 3. Run specific test
go test -v ./cmd/aitaskbuilder/ -run TestNewGetNewThing

# 4. Check coverage
go test -cover ./cmd/aitaskbuilder/
```

### Interpreting Test Failures

**Output mismatch (most common):**

```
expected
'AI Task Builder Thing Details:
ID: 01954894-65b3-779e-aaf6-348698e23634
Name: Test Thing
'
got
'AI Task Builder Thing Details:
ID: 01954894-65b3-779e-aaf6-348698e23634
Name:  Test Thing
'
```

**What to look for:**
- Extra spaces: `Name:  Test` vs `Name: Test`
- Missing newlines: Compare line counts
- Date format: `2025-02-27 18:03:59` vs `2025-02-27T18:03:59Z`

**Fix:** Update expected string in test OR fix implementation to match expected

**Empty output:**

```
expected
'AI Task Builder Thing Details:
...'
got
''
```

**Cause:** Forgot `writer.Flush()`  
**Fix:** Add `writer.Flush()` before `b.String()` in test

**Mock not called:**

```
panic: Unexpected call to *mock_client.MockAPI.GetAITaskBuilderNewThing
```

**Causes:**
- Parameter mismatch: Not using `gomock.Eq()`
- Wrong parameter value
- Missing `.AnyTimes()`

**Fix:** Check mock expectation matches actual call exactly

**Error format mismatch:**

```
expected
'error: thing not found'
got
'thing not found'
```

**Cause:** RunE not wrapping error with "error:" prefix  
**Fix:** In Step 6 implementation, ensure RunE wraps errors

### Fixing Test Failures

**Process:**
1. Identify which test failed
2. Run that test in verbose mode
3. Read the entire error message
4. Identify the specific mismatch
5. Determine if issue is in test or implementation
6. Fix and re-run
7. Verify all tests still pass

**Don't:**
- Change multiple things at once
- Skip reading error messages
- Guess at fixes
- Proceed with failing tests

## Human Review Criteria

### Verification Commands

```bash
# 1. Full test suite
make test

# 2. Specific package with verbose
go test -v ./cmd/aitaskbuilder/

# 3. Coverage check
go test -cover ./cmd/aitaskbuilder/

# 4. Coverage detail
go test -coverprofile=coverage.out ./cmd/aitaskbuilder/
go tool cover -func=coverage.out | grep get_new_thing
```

### Expected Results

**1. Full test suite:**
```
✅ All tests pass
✅ No failures
✅ Includes new tests

Example:
ok      github.com/prolific-oss/cli/cmd/aitaskbuilder    0.123s
```

**2. Verbose output:**
```
✅ All 5 new tests show as PASS:

=== RUN   TestNewGetNewThingCommand
--- PASS: TestNewGetNewThingCommand (0.00s)
=== RUN   TestNewGetNewThingCommandCallsAPI
--- PASS: TestNewGetNewThingCommandCallsAPI (0.00s)
=== RUN   TestNewGetNewThingCommandCallsAPIWithoutOptionalFields
--- PASS: TestNewGetNewThingCommandCallsAPIWithoutOptionalFields (0.00s)
=== RUN   TestNewGetNewThingCommandHandlesErrors
--- PASS: TestNewGetNewThingCommandHandlesErrors (0.00s)
=== RUN   TestNewGetNewThingCommandRequiresThingID
--- PASS: TestNewGetNewThingCommandRequiresThingID (0.00s)
PASS
```

**3. Coverage:**
```
✅ Package coverage shown:
ok      github.com/prolific-oss/cli/cmd/aitaskbuilder    0.123s  coverage: 87.5% of statements
```

**4. Coverage detail:**
```
✅ New functions show 100% coverage:
github.com/prolific-oss/cli/cmd/aitaskbuilder/get_new_thing.go:XX:   NewGetNewThingCommand         100.0%
github.com/prolific-oss/cli/cmd/aitaskbuilder/get_new_thing.go:YY:   renderAITaskBuilderNewThing   100.0%
```

### Code Review Checklist

- [ ] **All tests pass**: `make test` succeeds
- [ ] **5 new tests**: All test functions execute and pass
- [ ] **Coverage**: New code has high coverage (ideally 100%)
- [ ] **No regressions**: All existing tests still pass
- [ ] **Consistent results**: Running multiple times gives same result
- [ ] **Fast execution**: Tests run in < 1 second

### What to Check in Results

**Test output quality:**
- All tests have clear names
- No skipped tests
- No flaky behavior (run 2-3 times to verify)
- Coverage is complete

**Error messages (if failures):**
- Read entire error message
- Look for exact mismatch location
- Compare expected vs actual carefully
- Check for invisible characters (spaces, newlines)

## Common Issues

### Issue: Output mismatch - whitespace
**Symptom**: Test shows difference in whitespace  
**Debugging:**
```bash
# Run test and save output
go test -v ./cmd/aitaskbuilder/ -run TestNewGetNewThingCommandCallsAPI 2>&1 | tee test-output.txt

# Examine exact differences
cat -A test-output.txt  # Shows $ for newlines, spaces visible
```
**Fix:** Match whitespace exactly in expected output

### Issue: Empty output in test
**Symptom**: Got empty string, expected full output  
**Cause:** Missing `writer.Flush()`  
**Fix:** Add `writer.Flush()` before `b.String()` in test

### Issue: Date format mismatch
**Symptom**: Different date format in output  
**Fix:** Ensure test uses same format as implementation: `"2006-01-02 15:04:05"`

### Issue: Error not wrapped
**Symptom**: Error message doesn't have "error:" prefix  
**Cause:** RunE not wrapping error  
**Fix:** In Step 6, ensure RunE wraps: `fmt.Errorf("error: %s", err.Error())`

### Issue: Mock expectation not met
**Symptom**: Panic about unexpected call  
**Fix:** Ensure mock uses `gomock.Eq()` and `AnyTimes()`

### Issue: Test timeout
**Symptom**: Test hangs or times out  
**Cause:** Possible infinite loop or blocking operation  
**Fix:** Review implementation for blocking code

## Success Criteria

- [ ] `make test` passes with zero failures
- [ ] All 5 new tests pass individually
- [ ] Coverage for new code is 100% or near-100%
- [ ] All existing tests still pass (no regressions)
- [ ] Tests complete quickly (< 1 second)
- [ ] Results are consistent across multiple runs
- [ ] No flaky tests

## Debugging Failed Tests

**Step-by-step debugging process:**

1. **Run verbose:**
   ```bash
   go test -v ./cmd/aitaskbuilder/ -run TestNewGetNewThing
   ```

2. **Identify failure:**
   - Which test failed?
   - What's the error message?
   - Expected vs actual output?

3. **Isolate issue:**
   - Is it whitespace?
   - Is it missing output?
   - Is it wrong format?

4. **Fix:**
   - Update test expected output, OR
   - Fix implementation
   - Re-run test

5. **Verify:**
   - Run all tests again
   - Ensure fix didn't break others

## Checkpoint 4: Verification

✅ **CHECKPOINT 4 REACHED**

At this point, full test coverage should be verified:

```bash
# Clean build and test
make clean
make all

# Verify specific tests
go test -v ./cmd/aitaskbuilder/

# Check coverage
go test -cover ./cmd/aitaskbuilder/

# Verify command still works
./prolific aitaskbuilder get-new-thing --help
```

**Expected:** All tests pass, coverage is high, build succeeds, command works.

**If this fails, do not proceed to Step 11. Fix issues first.**

## Next Step

Once Checkpoint 4 passes, proceed to [Step 11: Update Documentation](step-11-documentation.md) (optional)
