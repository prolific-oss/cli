# Step 4: Regenerate Mocks

## Purpose

Regenerate mock implementations of the API interface to fix the broken build from Step 3. This step:
- Generates mock methods for testing
- Restores the build to working state
- Allows existing tests to continue passing

## AI Implementation Prompt

```
The build is broken because I added a new method to the API interface. 
I need to regenerate the mocks to fix it.

Run the mock generation command:
make test-gen-mock

This will:
1. Read the API interface from client/client.go
2. Generate mock implementations in mock_client/mock_client.go
3. Fix the broken build

After running, verify the build is fixed by running: make test
```

## AI Implementation Guidance

### Command to Run

Simply execute:
```bash
make test-gen-mock
```

### What This Does

The command:
1. Uses `mockgen` to read `client/client.go`
2. Generates mock implementations for all interface methods
3. Writes to `mock_client/mock_client.go`
4. Includes your new method automatically

### Expected File Changes

The `mock_client/mock_client.go` file will be updated with:

```go
// GetAITaskBuilderNewThing mocks base method
func (m *MockAPI) GetAITaskBuilderNewThing(thingID string) (*client.GetAITaskBuilderNewThingResponse, error) {
    m.ctrl.T.Helper()
    ret := m.ctrl.Call(m, "GetAITaskBuilderNewThing", thingID)
    ret0, _ := ret[0].(*client.GetAITaskBuilderNewThingResponse)
    ret1, _ := ret[1].(error)
    return ret0, ret1
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockAPIMockRecorder) GetAITaskBuilderNewThing(thingID interface{}) *gomock.Call {
    m.mock.ctrl.T.Helper()
    return m.mock.ctrl.RecordCallWithMethodType(m.mock, "GetAITaskBuilderNewThing", reflect.TypeOf((*MockAPI)(nil).GetAITaskBuilderNewThing), thingID)
}
```

### What NOT to Do

❌ **Don't:**
- Manually edit `mock_client/mock_client.go` (it's auto-generated)
- Skip this step (build will remain broken)
- Try to fix the build another way
- Proceed to Step 5 if this fails

✅ **Do:**
- Run the make command exactly as shown
- Verify the command completes without errors
- Check that mock file is updated
- Verify build is now fixed

## Human Review Criteria

### Verification Steps

Run these commands in order:

```bash
# 1. Regenerate mocks
make test-gen-mock

# 2. Check that mock file changed
git diff mock_client/mock_client.go

# 3. Verify build is fixed
go build ./client/...

# 4. Run all tests to ensure nothing broke
make test
```

### Expected Results

**1. Mock generation:**
```
✅ Command completes successfully
✅ Output shows generation process
✅ No errors reported
```

**2. Git diff:**
```
✅ Shows additions to mock_client/mock_client.go
✅ New mock method for GetAITaskBuilderNewThing
✅ New EXPECT recorder method
✅ Approximately 15-20 new lines
```

**3. Build output:**
```
✅ Build succeeds (previously failed)
✅ No "does not implement API" error
✅ No compilation errors
```

**4. Test output:**
```
✅ All existing tests pass
✅ No test failures
✅ No new errors introduced
```

### Code Review Checklist

- [ ] **Mock generation ran**: `make test-gen-mock` completed successfully
- [ ] **Mock file updated**: `git diff` shows changes to `mock_client/mock_client.go`
- [ ] **Mock method added**: New `GetAITaskBuilderNewThing` method exists
- [ ] **EXPECT method added**: New `EXPECT` recorder method exists
- [ ] **Signature matches**: Mock signature matches API interface exactly
- [ ] **Build fixed**: `go build ./client/...` succeeds
- [ ] **Tests pass**: `make test` succeeds with all tests passing

### What to Look For in Git Diff

```diff
+// GetAITaskBuilderNewThing mocks base method
+func (m *MockAPI) GetAITaskBuilderNewThing(thingID string) (*client.GetAITaskBuilderNewThingResponse, error) {
+    m.ctrl.T.Helper()
+    ret := m.ctrl.Call(m, "GetAITaskBuilderNewThing", thingID)
+    ret0, _ := ret[0].(*client.GetAITaskBuilderNewThingResponse)
+    ret1, _ := ret[1].(error)
+    return ret0, ret1
+}
+
+// EXPECT returns an object that allows the caller to indicate expected use
+func (m *MockAPIMockRecorder) GetAITaskBuilderNewThing(thingID interface{}) *gomock.Call {
+    m.mock.ctrl.T.Helper()
+    return m.mock.ctrl.RecordCallWithMethodType(m.mock, "GetAITaskBuilderNewThing", reflect.TypeOf((*MockAPI)(nil).GetAITaskBuilderNewThing), thingID)
+}
```

## Common Issues

### Issue: Command not found
**Symptom**: `make: test-gen-mock: command not found`  
**Cause**: mockgen not installed  
**Fix**: Run `make install` first to install dependencies

### Issue: Mock file not updated
**Symptom**: Git diff shows no changes  
**Cause**: Command failed silently or interface wasn't changed  
**Fix**: Check for error messages, verify Step 3 was completed

### Issue: Build still fails
**Symptom**: Build error persists after regeneration  
**Cause**: Interface signature has errors  
**Fix**: Review Step 3, check for typos in method signature

### Issue: Tests fail after regeneration
**Symptom**: Previously passing tests now fail  
**Cause**: Unexpected interface change  
**Fix**: Review interface changes, ensure only new method was added

## Success Criteria

- [ ] `make test-gen-mock` runs without errors
- [ ] `mock_client/mock_client.go` shows changes in git
- [ ] New mock method exists with correct signature
- [ ] `go build ./client/...` succeeds (was failing before)
- [ ] `make test` passes with all tests passing
- [ ] Git diff shows only expected mock changes

## Checkpoint 1: Verification

✅ **CHECKPOINT 1 REACHED**

At this point, verify that the foundation is complete:

```bash
# Run full test suite
make test
```

**Expected:** All existing tests pass, build succeeds, no errors.

**If this fails, do not proceed to Step 5. Fix issues first.**

## Next Step

Once Checkpoint 1 passes, proceed to [Step 5: Implement API Client Method](step-05-api-implementation.md)
