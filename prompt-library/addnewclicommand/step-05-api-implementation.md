# Step 5: Implement API Client Method

## Purpose

Implement the actual API client method that makes HTTP calls to the Prolific API. This step:
- Completes the API layer
- Provides the foundation for the CLI command
- Follows established patterns for the specific HTTP method

**The implementation pattern varies by HTTP method - see sections below.**

## Determining Your Pattern

| HTTP Method | When to Use | Pattern to Follow | Status Code |
|-------------|-------------|-------------------|-------------|
| **GET** | Retrieve resources | Simple (no body, no status check) | 200 |
| **POST (Create)** | Create new resources | Full payload, check 201 | 201 Created |
| **POST (Action)** | Perform actions | Action payload, check 200 | 200 OK |
| **POST (Clone)** | Duplicate resources | No body, check 200 | 200 OK |
| **PATCH** | Update resources | Partial payload, check 200 | 200 OK |
| **DELETE** | Delete resources | No body, check 204 | 204 No Content |

## AI Implementation Prompts

### For GET Operations
```
I need to implement a GET API client method for AITaskBuilder.

Add implementation to `client/client.go`:
- Location: AITaskBuilder section (around line 599+)
- Pattern: Simple GET (no status check)
- Method name: GetAITaskBuilder<Resource>
- API endpoint: /api/v1/data-collection/<path>
- Parameters: [SPECIFY]
- Response type: [From Step 1]

Follow the simpler AITaskBuilder pattern (without status code check).
```

### For POST Create Operations
```
I need to implement a POST API client method to create a resource.

Add implementation to `client/client.go`:
- Location: AITaskBuilder section (around line 599+)
- Pattern: POST with payload, expect 201 Created
- Method name: CreateAITaskBuilder<Resource>
- API endpoint: /api/v1/data-collection/<path>
- Payload type: [From Step 1]
- Response type: [From Step 1]

Include status code check for 201 and read error body on failure.
```

### For POST Action Operations
```
I need to implement a POST API client method to perform an action.

Add implementation to `client/client.go`:
- Location: AITaskBuilder section (around line 599+)
- Pattern: POST with action payload, expect 200 OK
- Method name: <Action>AITaskBuilder<Resource>
- API endpoint: /api/v1/data-collection/<path>/<action>
- Payload: Simple action struct (may use inline)
- Response type: [From Step 1]
```

### For PATCH Update Operations
```
I need to implement a PATCH API client method to update a resource.

Add implementation to `client/client.go`:
- Location: AITaskBuilder section (around line 599+)
- Pattern: PATCH with partial payload, expect 200 OK
- Method name: UpdateAITaskBuilder<Resource>
- API endpoint: /api/v1/data-collection/<path>/{id}
- Payload type: [From Step 1]
- Response type: [From Step 1]

Include status code check for 200.
```

## AI Implementation Guidance

### Pattern 1: GET Operation (Simple)

**Used for:** Retrieving resources

**Example from codebase:**
```go
// GetAITaskBuilderBatch will return details of an AI Task Builder batch.
func (c *Client) GetAITaskBuilderBatch(batchID string) (*GetAITaskBuilderBatchResponse, error) {
    var response GetAITaskBuilderBatchResponse
    
    url := fmt.Sprintf("/api/v1/data-collection/batches/%s", batchID)
    _, err := c.Execute(http.MethodGet, url, nil, &response)  // ← nil body, discard httpResponse
    if err != nil {
        return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
    }
    
    return &response, nil
}
```

**Key points:**
- Request body: `nil`
- Discard `httpResponse` with `_`
- No status code check (simpler pattern for AITaskBuilder)
- Return pointer to response

### Pattern 2: POST Operation - Create Resource

**Used for:** Creating new resources (workspaces, projects, batches, etc.)

**Example from codebase (workspace):**
```go
// CreateWorkspace will create a workspace
func (c *Client) CreateWorkspace(workspace model.Workspace) (*CreateWorkspacesResponse, error) {
    var response CreateWorkspacesResponse
    
    url := "/api/v1/workspaces/"
    httpResponse, err := c.Execute(http.MethodPost, url, workspace, &response)  // ← Pass payload
    if err != nil {
        return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
    }
    
    // Check for 201 Created
    if httpResponse.StatusCode != http.StatusCreated {
        body, _ := io.ReadAll(httpResponse.Body)
        return nil, fmt.Errorf("unable to create workspace: %v", string(body))
    }
    
    return &response, nil
}
```

**Key points:**
- Request body: Full resource payload
- Capture `httpResponse` (not discarded)
- Check for `http.StatusCreated` (201)
- Read response body on error for better diagnostics
- Return pointer to created resource

**Template:**
```go
func (c *Client) CreateAITaskBuilder<Resource>(payload <PayloadType>) (*<ResponseType>, error) {
    var response <ResponseType>
    
    url := "/api/v1/data-collection/<resources>/"
    httpResponse, err := c.Execute(http.MethodPost, url, payload, &response)
    if err != nil {
        return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
    }
    
    if httpResponse.StatusCode != http.StatusCreated {
        body, _ := io.ReadAll(httpResponse.Body)
        return nil, fmt.Errorf("unable to create <resource>: %v", string(body))
    }
    
    return &response, nil
}
```

### Pattern 3: POST Operation - Action

**Used for:** Performing actions on resources (transition, publish, archive, etc.)

**Example from codebase (transition study):**
```go
// TransitionStudy will transition a study to a new status
func (c *Client) TransitionStudy(ID, action string) (*TransitionStudyResponse, error) {
    var response TransitionStudyResponse
    
    // Simple inline payload struct
    transition := struct {
        Action string `json:"action"`
    }{
        Action: action,
    }
    
    url := fmt.Sprintf("/api/v1/studies/%s/transition/", ID)
    _, err := c.Execute(http.MethodPost, url, transition, &response)  // ← Pass action
    if err != nil {
        return nil, fmt.Errorf("unable to transition study to %s: %v", action, err)
    }
    
    return &response, nil
}
```

**Key points:**
- Request body: Simple struct (often inline)
- May discard `httpResponse` (expects 200, less critical)
- Include action context in error message
- Return result of action

**Template:**
```go
func (c *Client) <Action>AITaskBuilder<Resource>(resourceID string, params <ParamType>) (*<ResponseType>, error) {
    var response <ResponseType>
    
    payload := <PayloadType>{
        // Set fields from params
    }
    
    url := fmt.Sprintf("/api/v1/data-collection/<resources>/%s/<action>/", resourceID)
    _, err := c.Execute(http.MethodPost, url, payload, &response)
    if err != nil {
        return nil, fmt.Errorf("unable to <action> <resource>: %v", err)
    }
    
    return &response, nil
}
```

### Pattern 4: POST Operation - Clone/Duplicate (No Body)

**Used for:** Duplicating resources without requiring input

**Example from codebase (duplicate study):**
```go
// DuplicateStudy will duplicate an existing study
func (c *Client) DuplicateStudy(ID string) (*model.Study, error) {
    var response model.Study
    
    url := fmt.Sprintf("/api/v1/studies/%s/clone/", ID)
    httpResponse, err := c.Execute(http.MethodPost, url, nil, &response)  // ← nil body
    if err != nil {
        return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
    }
    
    if httpResponse.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(httpResponse.Body)
        return nil, fmt.Errorf("unable to duplicate study: %v", string(body))
    }
    
    return &response, nil
}
```

**Key points:**
- Request body: `nil` (no input needed)
- Check for `http.StatusOK` (200)
- Read response body on error
- Return new duplicated resource

### Pattern 5: PATCH Operation - Update

**Used for:** Updating existing resources with partial data

**Example from codebase (update study):**
```go
// UpdateStudy will update an existing study
func (c *Client) UpdateStudy(ID string, study model.UpdateStudy) (*model.Study, error) {
    var response model.Study
    
    url := fmt.Sprintf("/api/v1/studies/%s/", ID)
    httpResponse, err := c.Execute(http.MethodPatch, url, study, &response)  // ← Partial payload
    if err != nil {
        return nil, fmt.Errorf("unable to update study: %v", err)
    }
    
    if httpResponse.StatusCode != http.StatusOK {
        return nil, errors.New(`unable to update study`)
    }
    
    return &response, nil
}
```

**Key points:**
- Request body: Partial update struct (only changed fields)
- Check for `http.StatusOK` (200)
- Return updated full resource

**Template:**
```go
func (c *Client) UpdateAITaskBuilder<Resource>(resourceID string, update <UpdatePayloadType>) (*<ResponseType>, error) {
    var response <ResponseType>
    
    url := fmt.Sprintf("/api/v1/data-collection/<resources>/%s/", resourceID)
    httpResponse, err := c.Execute(http.MethodPatch, url, update, &response)
    if err != nil {
        return nil, fmt.Errorf("unable to update <resource>: %v", err)
    }
    
    if httpResponse.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(httpResponse.Body)
        return nil, fmt.Errorf("unable to update <resource>: %v", string(body))
    }
    
    return &response, nil
}
```

### Pattern 6: DELETE Operation

**Used for:** Deleting resources

**Not currently in codebase, but typical pattern:**
```go
// DeleteAITaskBuilderBatch will delete an AI Task Builder batch.
func (c *Client) DeleteAITaskBuilderBatch(batchID string) error {
    url := fmt.Sprintf("/api/v1/data-collection/batches/%s", batchID)
    httpResponse, err := c.Execute(http.MethodDelete, url, nil, nil)  // ← nil body and response
    if err != nil {
        return fmt.Errorf("unable to fulfil request %s: %s", url, err)
    }
    
    if httpResponse.StatusCode != http.StatusNoContent {
        body, _ := io.ReadAll(httpResponse.Body)
        return fmt.Errorf("unable to delete batch: %v", string(body))
    }
    
    return nil
}
```

**Key points:**
- Request body: `nil`
- Response: `nil` (no body expected)
- Check for `http.StatusNoContent` (204)
- Return error only (no resource returned)

## Method Signature Patterns

### GET Methods
```go
Get<Resource>(id string) (*<Resource>Response, error)
Get<Resources>(workspaceID string) (*List<Resources>Response, error)
```

### POST Methods (Create)
```go
Create<Resource>(payload <CreatePayload>) (*<Resource>Response, error)
```

### POST Methods (Action)
```go
<Action><Resource>(id string, params ...) (*<Action>Response, error)
Transition<Resource>(id string, action string) (*TransitionResponse, error)
Duplicate<Resource>(id string) (*<Resource>, error)
```

### PATCH Methods
```go
Update<Resource>(id string, update <UpdatePayload>) (*<Resource>, error)
```

### DELETE Methods
```go
Delete<Resource>(id string) error
```

## What NOT to Do

❌ **Don't:**
- Use wrong HTTP method constant
- Forget to check status code for write operations
- Use string "POST" instead of `http.MethodPost`
- Forget to read error body on failures (for better errors)
- Return wrong type (e.g., value instead of pointer)
- Use wrong status code (201 vs 200)

✅ **Do:**
- Use correct HTTP method constant
- Check appropriate status code for operation type
- Read and include error body for diagnostics
- Return pointer to response
- Match signature from interface (Step 3)
- Include descriptive error messages with context

## Human Review Criteria

### Code Review Checklist

**All Operations:**
- [ ] **Location**: Added in `client/client.go` AITaskBuilder section (line 599+)
- [ ] **Doc comment**: Present and describes what method does
- [ ] **Signature**: Matches API interface from Step 3 exactly
- [ ] **Response variable**: Declared as value type (not pointer)
- [ ] **URL**: Uses `fmt.Sprintf` with correct endpoint
- [ ] **Return**: Returns `&response` (pointer) on success

**GET Operations:**
- [ ] **HTTP method**: Uses `http.MethodGet`
- [ ] **Body parameter**: Passes `nil`
- [ ] **Response capture**: Discards httpResponse with `_`
- [ ] **Status check**: Omitted (simpler pattern)

**POST Create Operations:**
- [ ] **HTTP method**: Uses `http.MethodPost`
- [ ] **Body parameter**: Passes payload struct
- [ ] **Response capture**: Captures `httpResponse`
- [ ] **Status check**: Checks for `http.StatusCreated` (201)
- [ ] **Error handling**: Reads response body on error

**POST Action Operations:**
- [ ] **HTTP method**: Uses `http.MethodPost`
- [ ] **Body parameter**: Passes action struct (or inline)
- [ ] **Status check**: May check for `http.StatusOK` (200) or omit
- [ ] **Error context**: Includes action name in error message

**PATCH Operations:**
- [ ] **HTTP method**: Uses `http.MethodPatch`
- [ ] **Body parameter**: Passes update struct
- [ ] **Response capture**: Captures `httpResponse`
- [ ] **Status check**: Checks for `http.StatusOK` (200)

**DELETE Operations:**
- [ ] **HTTP method**: Uses `http.MethodDelete`
- [ ] **Body parameter**: Passes `nil`
- [ ] **Response parameter**: Passes `nil`
- [ ] **Status check**: Checks for `http.StatusNoContent` (204)
- [ ] **Return type**: Returns `error` only

## Verification Commands

```bash
# 1. Verify syntax
go build ./client/...

# 2. Run tests (should still pass)
make test

# 3. Check implementation location
grep -n "func (c \*Client) <MethodName>" client/client.go

# 4. Verify it matches interface
grep "<MethodName>" client/client.go
```

## Common Issues by Operation Type

### GET Operations
- Using `http.MethodPost` instead of `http.MethodGet`
- Passing body instead of `nil`
- Unnecessary status code check

### POST Create Operations
- Checking for 200 instead of 201
- Not reading error body on failure
- Not capturing httpResponse

### POST Action Operations
- Complex payload when simple inline struct would work
- Missing action context in error messages

### PATCH Operations
- Using PUT instead of PATCH
- Passing full resource instead of partial update

### All Write Operations
- Not checking status code at all
- Using string literals for HTTP methods
- Not returning pointer to response

## Success Criteria

- [ ] Method implemented in `client/client.go` around line 599+
- [ ] Doc comment present
- [ ] Signature matches interface exactly
- [ ] Uses correct HTTP method constant
- [ ] Appropriate status code check for operation type
- [ ] Error handling includes context
- [ ] Returns correct type (pointer for resources, error for deletes)
- [ ] `go build ./client/...` succeeds
- [ ] `make test` passes

## Reference

See [HTTP-METHODS-GUIDE.md](HTTP-METHODS-GUIDE.md) for comprehensive examples and patterns for each HTTP method.

## Next Step

Proceed to [Step 6: Create Command Implementation File](step-06-command-file.md)
