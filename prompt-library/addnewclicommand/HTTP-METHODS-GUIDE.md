# HTTP Method Patterns Guide

This guide documents the different patterns for GET, POST, PATCH, PUT, and DELETE operations in the Prolific CLI.

## Overview

The Prolific CLI implements different patterns based on the HTTP method:

| Method | Usage | Request Body | Status Code | Pattern Example |
|--------|-------|--------------|-------------|-----------------|
| **GET** | Retrieve resources | `nil` | 200 OK | GetAITaskBuilderBatch |
| **POST** | Create/Action resources | Payload struct | 201 Created or 200 OK | CreateWorkspace, TransitionStudy |
| **PATCH** | Update resources | Update struct | 200 OK | UpdateStudy |
| **PUT** | Replace resources | Full resource | 200 OK | *(Not used in codebase)* |
| **DELETE** | Delete resources | `nil` | 204 No Content | *(Not used in codebase)* |

## GET Operations (Read)

### API Client Pattern

```go
// GetAITaskBuilderBatch will return details of an AI Task Builder batch.
func (c *Client) GetAITaskBuilderBatch(batchID string) (*GetAITaskBuilderBatchResponse, error) {
    var response GetAITaskBuilderBatchResponse
    
    url := fmt.Sprintf("/api/v1/data-collection/batches/%s", batchID)
    _, err := c.Execute(http.MethodGet, url, nil, &response)  // ← nil body
    if err != nil {
        return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
    }
    
    return &response, nil
}
```

**Key characteristics:**
- Request body: `nil` (3rd parameter to Execute)
- Response: Pointer to response struct
- Status check: Usually omitted for GET in AITaskBuilder methods
- Error handling: Simple error wrapping

### Command Pattern

```go
func renderAITaskBuilderBatch(c client.API, opts BatchGetOptions, w io.Writer) error {
    if opts.BatchID == "" {
        return errors.New(ErrBatchIDRequired)
    }
    
    response, err := c.GetAITaskBuilderBatch(opts.BatchID)
    if err != nil {
        return err
    }
    
    batch := response.AITaskBuilderBatch
    
    fmt.Fprintf(w, "Details:\n")
    fmt.Fprintf(w, "ID: %s\n", batch.ID)
    // ... more output
    
    return nil
}
```

## POST Operations (Create/Action)

### API Client Pattern - Create Resource

**Example: CreateWorkspace (client.go:404-414)**

```go
// CreateWorkspace will create a workspace
func (c *Client) CreateWorkspace(workspace model.Workspace) (*CreateWorkspacesResponse, error) {
    var response CreateWorkspacesResponse
    
    url := "/api/v1/workspaces/"
    httpResponse, err := c.Execute(http.MethodPost, url, workspace, &response)  // ← Pass payload
    if err != nil {
        return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
    }
    
    if httpResponse.StatusCode != http.StatusCreated {  // ← Check for 201
        body, _ := io.ReadAll(httpResponse.Body)
        return nil, fmt.Errorf("unable to create workspace: %v", string(body))
    }
    
    return &response, nil
}
```

**Key characteristics:**
- Request body: Model struct or payload struct
- Expected status: `http.StatusCreated` (201)
- Error handling: Read and return response body on error
- Response: Created resource with ID

### API Client Pattern - Action Resource

**Example: TransitionStudy (client.go:279-295)**

```go
// TransitionStudy will transition a study to a new status
func (c *Client) TransitionStudy(ID, action string) (*TransitionStudyResponse, error) {
    var response TransitionStudyResponse
    
    // Create inline payload struct
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

**Key characteristics:**
- Request body: Simple struct (often inline)
- Expected status: `http.StatusOK` (200) for actions
- Error handling: Include action context in error message

### API Client Pattern - No Request Body

**Example: DuplicateStudy (client.go:187-202)**

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

### Command Pattern - Create

**Example: workspace/create.go:54-72**

```go
func createWorkspace(client client.API, opts CreateOptions, w io.Writer) error {
    // Validation
    if opts.Title == "" {
        return errors.New("title is required")
    }
    
    // Build payload
    workspace := model.Workspace{
        Title: opts.Title,
    }
    
    // Create resource
    record, err := client.CreateWorkspace(workspace)
    if err != nil {
        return err
    }
    
    // Output success with ID
    fmt.Fprintf(w, "Created workspace: %s\n", record.ID)
    
    return nil
}
```

**Key characteristics:**
- Validation before API call
- Build payload from options
- Success message shows created ID
- Simple output (not full resource details)

### Command Pattern - Action

**Example: study/transition.go:51-71**

```go
func transitionStudy(client client.API, opts TransitionOptions, w io.Writer) error {
    if opts.Action == "" {
        return fmt.Errorf("you must provide an action to transition the study to")
    }
    
    // Perform action
    _, err := client.TransitionStudy(opts.Args[0], opts.Action)
    if err != nil {
        return err
    }
    
    // Optionally re-fetch to show updated state
    if !opts.Silent {
        study, err := client.GetStudy(opts.Args[0])
        if err != nil {
            return err
        }
        
        fmt.Fprintln(w, studyui.RenderStudy(*study))
    }
    
    return nil
}
```

**Key characteristics:**
- Action validation
- Optional re-fetch to show updated resource
- Silent flag to suppress output
- Uses UI renderer for full resource display

## PATCH Operations (Update)

### API Client Pattern

**Example: UpdateStudy (client.go:311-325)**

```go
// UpdateStudy will update an existing study
func (c *Client) UpdateStudy(ID string, study model.UpdateStudy) (*model.Study, error) {
    var response model.Study
    
    url := fmt.Sprintf("/api/v1/studies/%s/", ID)
    httpResponse, err := c.Execute(http.MethodPatch, url, study, &response)  // ← Partial update
    if err != nil {
        return nil, fmt.Errorf("unable to update study: %v", err)
    }
    
    if httpResponse.StatusCode != http.StatusOK {
        return nil, errors.New(`unable to update study`)
    }
    
    return &response, nil
}
```

**Key characteristics:**
- Request body: Partial update struct (not full resource)
- Expected status: `http.StatusOK` (200)
- Response: Updated full resource

### Command Pattern

**Example: study/increase_places.go:36-56**

```go
func RunE: func(cmd *cobra.Command, args []string) error {
    opts.Args = args
    
    // Get current resource
    study, err := client.GetStudy(args[0])
    if err != nil {
        return fmt.Errorf("error: %s", err.Error())
    }
    
    // Validate update is allowed
    if study.TotalAvailablePlaces > opts.Places {
        return fmt.Errorf("study currently has %v places, and you cannot decrease", study.TotalAvailablePlaces)
    }
    
    // Perform update
    updatedStudy, err := client.UpdateStudy(study.ID, model.UpdateStudy{TotalAvailablePlaces: opts.Places})
    if err != nil {
        return err
    }
    
    // Show updated resource
    fmt.Fprintln(w, studyui.RenderStudy(*updatedStudy))
    return nil
}
```

**Key characteristics:**
- Often fetch current resource first for validation
- Validate update constraints (business logic)
- Build partial update payload
- Display updated resource

## PUT Operations (Replace)

**Status:** Not used in current codebase

**Typical pattern if implemented:**
```go
func (c *Client) ReplaceResource(ID string, resource model.Resource) (*model.Resource, error) {
    var response model.Resource
    
    url := fmt.Sprintf("/api/v1/resources/%s", ID)
    httpResponse, err := c.Execute(http.MethodPut, url, resource, &response)
    if err != nil {
        return nil, fmt.Errorf("unable to replace resource: %v", err)
    }
    
    if httpResponse.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(httpResponse.Body)
        return nil, fmt.Errorf("unable to replace resource: %v", string(body))
    }
    
    return &response, nil
}
```

## DELETE Operations

**Status:** Not used in current codebase

**Typical pattern if implemented:**
```go
func (c *Client) DeleteResource(ID string) error {
    url := fmt.Sprintf("/api/v1/resources/%s", ID)
    httpResponse, err := c.Execute(http.MethodDelete, url, nil, nil)
    if err != nil {
        return fmt.Errorf("unable to delete resource: %v", err)
    }
    
    if httpResponse.StatusCode != http.StatusNoContent {
        body, _ := io.ReadAll(httpResponse.Body)
        return fmt.Errorf("unable to delete resource: %v", string(body))
    }
    
    return nil
}
```

**Command pattern:**
```go
func deleteResource(client client.API, opts DeleteOptions, w io.Writer) error {
    if opts.ResourceID == "" {
        return errors.New(ErrResourceIDRequired)
    }
    
    err := client.DeleteResource(opts.ResourceID)
    if err != nil {
        return err
    }
    
    fmt.Fprintf(w, "Deleted resource: %s\n", opts.ResourceID)
    
    return nil
}
```

## Payload Structs

### Location
Request payload structs are defined in `client/payloads.go`

**Example:**
```go
// SendMessagePayload represents the JSON payload for sending a message
type SendMessagePayload struct {
    RecipientID string `json:"recipient_id"`
    StudyID     string `json:"study_id"`
    Body        string `json:"body"`
}
```

### Patterns
- **Full resource**: Use model struct (e.g., `model.Workspace`)
- **Partial update**: Create specific update struct (e.g., `model.UpdateStudy`)
- **Simple action**: Inline struct or simple payload struct
- **No body**: Pass `nil`

## Summary Table

| Operation | Method | Body | Status | Response | Error Body Read |
|-----------|--------|------|--------|----------|-----------------|
| Get/List | GET | `nil` | 200 | Resource(s) | Optional |
| Create | POST | Full model | 201 | Created resource | Yes |
| Action | POST | Action params | 200 | Result | Optional |
| Clone | POST | `nil` | 200 | New resource | Yes |
| Update | PATCH | Partial model | 200 | Updated resource | Optional |
| Replace | PUT | Full model | 200 | Replaced resource | Yes |
| Delete | DELETE | `nil` | 204 | None | Yes |

## Command Output Patterns

**GET operations:**
- Display formatted resource details
- Show all relevant fields
- Handle optional fields conditionally

**POST operations (Create):**
- Simple success message with ID
- Or: Re-fetch and display full resource

**POST operations (Action):**
- Optional re-fetch to show updated state
- Support silent flag for scripting

**PATCH operations (Update):**
- Display updated resource
- Show confirmation of what changed

**DELETE operations:**
- Simple confirmation message
- Show ID of deleted resource
