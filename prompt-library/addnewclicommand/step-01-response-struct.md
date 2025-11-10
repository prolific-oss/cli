# Step 1: Define Response/Payload Structs

## Purpose

Create the data structures for API communication. This is done first because:
- Pure data structures with no dependencies
- Doesn't break any existing code
- Defines the contract for API requests and responses
- Will be referenced by API implementation in Step 5

**Note:** The structures you create depend on the HTTP method:
- **GET operations**: Response struct only
- **POST/PATCH operations**: May need both payload and response structs
- **DELETE operations**: Usually response struct only (or none)

## Determining What You Need

| HTTP Method | Need Payload? | Need Response? | Example |
|-------------|---------------|----------------|---------|
| GET | ❌ No | ✅ Yes | GetAITaskBuilderBatch |
| POST (Create) | ✅ Yes | ✅ Yes | CreateWorkspace |
| POST (Action) | ✅ Maybe* | ✅ Yes | TransitionStudy |
| PATCH (Update) | ✅ Yes | ✅ Yes | UpdateStudy |
| DELETE | ❌ No | ⚠️ Maybe** | DeleteResource |

\* Simple actions may use inline structs  
\** DELETE often returns 204 No Content

## AI Implementation Prompt

### For GET Operations
```
I need to add a response struct for a new AITaskBuilder GET endpoint.

Create a response struct in `client/responses.go`:
- Name: Get<Resource>Response (e.g., GetAITaskBuilderNewThingResponse)
- Location: Add near other AITaskBuilder responses (around line 190+)
- Include doc comment starting with the struct name
- Use embedded model types from the model package
- Include proper JSON tags

Resource name: [SPECIFY THE RESOURCE]
API response field name: [SPECIFY THE JSON FIELD]
```

### For POST/PATCH Operations
```
I need to add request payload and response structs for a new AITaskBuilder POST endpoint.

1. Create payload struct in `client/payloads.go`:
   - Name: <Action><Resource>Payload (e.g., CreateAITaskBuilderThingPayload)
   - Fields: [SPECIFY FIELDS]
   - JSON tags for each field

2. Create response struct in `client/responses.go`:
   - Name: <Action><Resource>Response (e.g., CreateAITaskBuilderThingResponse)
   - Response type: [SPECIFY]

Operation: [CREATE/UPDATE/ACTION]
Resource name: [SPECIFY]
Request fields: [SPECIFY]
Response field name: [SPECIFY]
```

## AI Implementation Guidance

### Response Structs (All Operations)

**Location:** `client/responses.go` (around line 190+)

**Single resource response:**
```go
// GetAITaskBuilderNewThingResponse is the response for getting a new thing
type GetAITaskBuilderNewThingResponse struct {
    AITaskBuilderNewThing model.AITaskBuilderNewThing `json:"thing"`
}
```

**Collection response:**
```go
// GetAITaskBuilderNewThingsResponse is the response for listing things
type GetAITaskBuilderNewThingsResponse struct {
    Results []model.AITaskBuilderNewThing `json:"results"`
}
```

**Create response (returns created resource):**
```go
// CreateAITaskBuilderThingResponse is the response for creating a thing
type CreateAITaskBuilderThingResponse struct {
    AITaskBuilderThing model.AITaskBuilderThing `json:"thing"`
}
```

**Action response (may return action result):**
```go
// TransitionStudyResponse is the response for transitioning a study
type TransitionStudyResponse struct {
    Study model.Study `json:"study"`
}
```

### Payload Structs (POST/PATCH Only)

**Location:** `client/payloads.go`

**Create payload (full resource):**
```go
// CreateAITaskBuilderThingPayload represents the request for creating a thing
type CreateAITaskBuilderThingPayload struct {
    Name        string `json:"name"`
    Description string `json:"description"`
    WorkspaceID string `json:"workspace_id"`
}
```

**Update payload (partial resource):**
```go
// UpdateAITaskBuilderThingPayload represents the request for updating a thing
type UpdateAITaskBuilderThingPayload struct {
    Name        *string `json:"name,omitempty"`         // ← Pointer for optional
    Description *string `json:"description,omitempty"`  // ← Pointer for optional
}
```

**Action payload (simple):**
```go
// TransitionStudyPayload represents the request for transitioning a study
type TransitionStudyPayload struct {
    Action string `json:"action"`
}
```

**Or use inline struct in API method (see Step 5):**
```go
// Simple payloads can be defined inline in the API method
transition := struct {
    Action string `json:"action"`
}{
    Action: action,
}
```

### Naming Conventions

**Response structs:**
- GET: `Get<Resource>Response`, `Get<Resource>sResponse`
- POST: `Create<Resource>Response`, `<Action><Resource>Response`
- PATCH: `Update<Resource>Response`
- DELETE: Usually none (204 No Content)

**Payload structs:**
- POST: `Create<Resource>Payload`, `<Action><Resource>Payload`
- PATCH: `Update<Resource>Payload`

**Examples:**
```go
// GET
type GetAITaskBuilderBatchResponse struct { ... }
type GetAITaskBuilderBatchesResponse struct { ... }

// POST (Create)
type CreateWorkspacePayload struct { ... }
type CreateWorkspacesResponse struct { ... }

// POST (Action)
type TransitionStudyPayload struct { ... }
type TransitionStudyResponse struct { ... }

// PATCH
type UpdateStudyPayload struct { ... }
type UpdateStudyResponse struct { ... }
```

### What NOT to Do

❌ **Don't:**
- Forget doc comments
- Mix payloads and responses in same file
- Use wrong naming pattern
- Forget JSON tags
- Use pointers in response structs (use value types)
- Forget `omitempty` for optional fields in update payloads

✅ **Do:**
- Separate payloads (payloads.go) from responses (responses.go)
- Match naming pattern for operation type
- Use value types for embedded models in responses
- Use pointers for optional fields in update payloads
- Include descriptive doc comments

## Human Review Criteria

### Code Review Checklist

**Response Struct:**
- [ ] **Location**: Added in `client/responses.go` near line 190+
- [ ] **Naming**: Follows pattern for operation type
- [ ] **Comment**: Has doc comment starting with struct name
- [ ] **Fields**: Uses embedded model types from `model/` package
- [ ] **JSON tags**: Present and match expected API response
- [ ] **Field types**: Value types (not pointers) for embedded models

**Payload Struct (if needed):**
- [ ] **Location**: Added in `client/payloads.go`
- [ ] **Naming**: Follows pattern `<Action><Resource>Payload`
- [ ] **Comment**: Has doc comment
- [ ] **Fields**: Match API requirements
- [ ] **JSON tags**: Present and correct
- [ ] **Optional fields**: Use pointers with `omitempty` for PATCH

### Verification Commands

```bash
# 1. Verify files compile
go build ./client/...

# 2. Check response struct exists
grep -A 3 "type Get.*Response\|type Create.*Response\|type Update.*Response" client/responses.go | tail -20

# 3. Check payload struct exists (if needed)
grep -A 5 "type Create.*Payload\|type Update.*Payload" client/payloads.go | tail -15

# 4. Verify location
grep -n "AITaskBuilder.*Response\|AITaskBuilder.*Payload" client/responses.go client/payloads.go
```

### Expected Results

**1. Build output:**
```
✅ Build succeeds with no errors
```

**2. Response struct:**
```
✅ Shows the new response struct definition
```

**3. Payload struct (if applicable):**
```
✅ Shows the new payload struct definition
```

**4. Location:**
```
✅ Structs appear in correct files and sections
```

## Common Issues

### Issue: Build fails with undefined type
**Cause**: Model type doesn't exist  
**Fix**: Verify the model type exists in `model/` package or create it first

### Issue: Payload struct in wrong file
**Cause**: Added to responses.go instead of payloads.go  
**Fix**: Move to correct file

### Issue: Update payload doesn't use pointers
**Cause**: Forgot to use pointer types for optional fields  
**Fix**: Change to `*string` with `omitempty` tag

### Issue: Using pointer type in response
**Cause**: Declared as `*model.Type` instead of `model.Type`  
**Fix**: Remove the `*` - use value type for embedded models in responses

## Success Criteria by Operation Type

### GET Operation
- [ ] Response struct in `responses.go`
- [ ] Follows `Get<Resource>Response` naming
- [ ] No payload struct needed

### POST Create Operation
- [ ] Payload struct in `payloads.go` with full resource fields
- [ ] Response struct in `responses.go` with created resource
- [ ] Naming: `Create<Resource>Payload` and `Create<Resource>Response`

### POST Action Operation
- [ ] Payload struct in `payloads.go` (or will use inline struct)
- [ ] Response struct in `responses.go` with result
- [ ] Naming: `<Action><Resource>Payload` and `<Action><Resource>Response`

### PATCH Update Operation
- [ ] Payload struct in `payloads.go` with optional fields as pointers
- [ ] Response struct in `responses.go` with updated resource
- [ ] Naming: `Update<Resource>Payload` and `Update<Resource>Response`
- [ ] `omitempty` tags on optional fields

### DELETE Operation
- [ ] Usually no structs needed (204 No Content)
- [ ] Or simple confirmation response struct if API returns one

## Examples by Operation

See [HTTP-METHODS-GUIDE.md](HTTP-METHODS-GUIDE.md) for complete examples of each operation type.

## Next Step

Proceed to [Step 2: Add Constants](step-02-constants.md)
