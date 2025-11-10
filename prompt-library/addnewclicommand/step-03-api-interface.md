# Step 3: Add API Method Signature to Interface

## Purpose

Add the method signature to the API interface. This is done now to:
- Define the contract between CLI and API client
- Allow mock generation in the next step
- Note: **This WILL break the build** - that's expected and fixed in Step 4

## AI Implementation Prompt

```
I need to add a new method signature to the client API interface.

Add the method signature to the `API` interface in `client/client.go`:
- Location: Within the API interface (around line 26-70)
- Position: Alphabetically with other GetAITaskBuilder* methods
- Pattern: GetAITaskBuilder<Resource>(parameters) (*Response, error)
- Return type: Must match the response struct from Step 1

Method details:
- Resource name: [SPECIFY]
- Parameters: [SPECIFY, e.g., "thingID string"]
- Response type: [SPECIFY, e.g., "GetAITaskBuilderNewThingResponse"]

Note: This will temporarily break the build. That's expected and will be fixed in Step 4.
```

## AI Implementation Guidance

### Method Signature Pattern

**Standard GET method:**
```go
GetAITaskBuilderNewThing(thingID string) (*GetAITaskBuilderNewThingResponse, error)
```

**List/collection method:**
```go
GetAITaskBuilderNewThings(workspaceID string) (*GetAITaskBuilderNewThingsResponse, error)
```

**With multiple parameters:**
```go
GetAITaskBuilderDataset(datasetID string, includeMetadata bool) (*GetAITaskBuilderDatasetResponse, error)
```

### Where to Add

Add within the `API` interface, grouped with other AITaskBuilder methods:

```go
type API interface {
    // ... other methods ...
    
    GetAITaskBuilderBatch(batchID string) (*GetAITaskBuilderBatchResponse, error)
    GetAITaskBuilderBatchStatus(batchID string) (*GetAITaskBuilderBatchStatusResponse, error)
    GetAITaskBuilderBatches(workspaceID string) (*GetAITaskBuilderBatchesResponse, error)
    GetAITaskBuilderNewThing(thingID string) (*GetAITaskBuilderNewThingResponse, error)  // ← Add here
    GetAITaskBuilderResponses(batchID string) (*GetAITaskBuilderResponsesResponse, error)
    
    // ... other methods ...
}
```

### Naming Rules

**Method name:**
- Prefix: `GetAITaskBuilder` (or `CreateAITaskBuilder`, `UpdateAITaskBuilder`)
- Resource: PascalCase resource name
- Pattern: `<Verb><Namespace><Resource>`

**Parameters:**
- Use descriptive names: `thingID`, `workspaceID`, not `id`
- Common types: `string` for IDs, `bool` for flags, `int` for counts
- Multiple words: camelCase (e.g., `workspaceID`, `includeMetadata`)

**Return type:**
- Always return a pointer: `*ResponseType`
- Always include error: `error`
- Must match response struct from Step 1

### What NOT to Do

❌ **Don't:**
- Add implementation (just the signature)
- Use vague parameter names (`id` instead of `thingID`)
- Forget the pointer `*` in return type
- Forget the `error` return value
- Add to wrong location (outside the API interface)

✅ **Do:**
- Add only the method signature
- Use descriptive parameter names
- Return pointer and error
- Group with other AITaskBuilder methods
- Match response type from Step 1 exactly

## Human Review Criteria

### Code Review Checklist

- [ ] **Location**: Added in `client/client.go` within `API` interface (line 26-70)
- [ ] **Grouping**: Placed with other `GetAITaskBuilder*` methods
- [ ] **Naming**: Follows pattern `GetAITaskBuilder<Resource>`
- [ ] **Parameters**: Use descriptive names, correct types
- [ ] **Return type**: Pointer to response struct, plus error
- [ ] **Response type**: Matches struct name from Step 1

### Specific Checks

**Method signature:**
```go
// ✅ CORRECT
GetAITaskBuilderNewThing(thingID string) (*GetAITaskBuilderNewThingResponse, error)
GetAITaskBuilderNewThings(workspaceID string) (*GetAITaskBuilderNewThingsResponse, error)

// ❌ INCORRECT
GetAITaskBuilderNewThing(id string) (*GetAITaskBuilderNewThingResponse, error)           // vague param
GetAITaskBuilderNewThing(thingID string) (GetAITaskBuilderNewThingResponse, error)       // not pointer
GetAITaskBuilderNewThing(thingID string) (*GetAITaskBuilderNewThingResponse)             // no error
getAITaskBuilderNewThing(thingID string) (*GetAITaskBuilderNewThingResponse, error)      // not exported
GetNewThing(thingID string) (*GetAITaskBuilderNewThingResponse, error)                   // missing prefix
```

**Parameter naming:**
```go
// ✅ CORRECT - Descriptive names
GetAITaskBuilderNewThing(thingID string)
GetAITaskBuilderBatches(workspaceID string)
GetAITaskBuilderDataset(datasetID string, includeMetadata bool)

// ❌ INCORRECT - Vague names
GetAITaskBuilderNewThing(id string)
GetAITaskBuilderBatches(workspace string)
GetAITaskBuilderDataset(dataset string, flag bool)
```

## Verification Commands

```bash
# 1. This WILL fail - that's expected
go build ./client/...

# 2. Verify method was added to interface
grep "GetAITaskBuilderNewThing" client/client.go

# 3. Check it's in the API interface
grep -B 5 -A 5 "GetAITaskBuilderNewThing" client/client.go | grep "type API"
```

## Expected Results

**1. Build output:**
```
❌ Build FAILS with error like:
   cannot use Client{...} (type Client) as type API in return argument:
   Client does not implement API (missing GetAITaskBuilderNewThing method)

✅ This is CORRECT and EXPECTED - will be fixed in Step 4
```

**2. Grep output:**
```
✅ Shows the method in the interface:
   GetAITaskBuilderNewThing(thingID string) (*GetAITaskBuilderNewThingResponse, error)
```

**3. Context grep:**
```
✅ Shows method is within the API interface:
   type API interface {
       ...
       GetAITaskBuilderNewThing(thingID string) (*GetAITaskBuilderNewThingResponse, error)
       ...
   }
```

## Common Issues

### Issue: Build succeeds when it should fail
**Cause**: Method not actually added to interface, or added in wrong place  
**Fix**: Verify method is within `type API interface { ... }` block

### Issue: Wrong method signature
**Cause**: Doesn't match response type from Step 1  
**Fix**: Ensure return type exactly matches response struct name

### Issue: Not in API interface
**Cause**: Added elsewhere in file, not in interface  
**Fix**: Place within `type API interface { }` block

### Issue: Vague parameter names
**Cause**: Used `id` instead of descriptive name  
**Fix**: Use descriptive names like `thingID`, `workspaceID`

## Success Criteria

- [ ] Method signature added to `API` interface
- [ ] Method name follows `GetAITaskBuilder<Resource>` pattern
- [ ] Parameters use descriptive names
- [ ] Return type is pointer to response struct from Step 1
- [ ] Return type includes error
- [ ] Build fails with "does not implement API" error ✅ (this is correct!)

## Important Note

⚠️ **The build is supposed to fail after this step.** This is expected and correct. The build will be fixed in Step 4 when mocks are regenerated.

**Do not try to fix the build manually - proceed directly to Step 4.**

## Next Step

Proceed immediately to [Step 4: Regenerate Mocks](step-04-regenerate-mocks.md) to fix the broken build.
