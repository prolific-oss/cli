# Step 6 Addendum: Create, Update, and Action Commands

This addendum to Step 6 provides patterns for non-GET operations (Create, Update, Action).

For GET operations, see the main [step-06-command-file.md](step-06-command-file.md).

## Command Patterns by Operation Type

| Operation | Command Name | Key Characteristics |
|-----------|--------------|---------------------|
| **GET** | `get-<resource>` | Display resource details, optional fields |
| **POST (Create)** | `create-<resource>` | Build payload, validate, show success with ID |
| **POST (Action)** | `<action>-<resource>` | Validate action, optional re-fetch, silent flag |
| **PATCH (Update)** | `update-<resource>` | Pre-fetch for validation, show updated state |
| **DELETE** | `delete-<resource>` | Confirm deletion, show deleted ID |

## Pattern 1: Create Command (POST)

**Example from codebase:** `cmd/workspace/create.go`

### File Structure

```go
package aitaskbuilder

import (
    "errors"
    "fmt"
    "io"
    
    "github.com/prolific-oss/cli/client"
    "github.com/prolific-oss/cli/model"
    "github.com/spf13/cobra"
)

// CreateOptions holds the options for creating a resource
type CreateOptions struct {
    Args        []string
    Name        string
    Description string
    WorkspaceID string
}

// NewCreateCommand creates a new command for creating resources
func NewCreateCommand(client client.API, w io.Writer) *cobra.Command {
    var opts CreateOptions
    
    cmd := &cobra.Command{
        Use:   "create",
        Short: "Create an AI Task Builder thing",
        Long: `Create a new AI Task Builder thing

Provide a name and description to create a new thing in your workspace.`,
        Example: `
Create a thing:
$ prolific aitaskbuilder create -n "My Thing" -d "Description" -w <workspace_id>
        `,
        RunE: func(cmd *cobra.Command, args []string) error {
            opts.Args = args
            
            err := createAITaskBuilderThing(client, opts, w)
            if err != nil {
                return fmt.Errorf("error: %s", err.Error())
            }
            
            return nil
        },
    }
    
    flags := cmd.Flags()
    flags.StringVarP(&opts.Name, "name", "n", "", "Name of the thing (required)")
    flags.StringVarP(&opts.Description, "description", "d", "", "Description")
    flags.StringVarP(&opts.WorkspaceID, "workspace-id", "w", viper.GetString("workspace"), "Workspace ID (required)")
    
    _ = cmd.MarkFlagRequired("name")
    _ = cmd.MarkFlagRequired("workspace-id")
    
    return cmd
}

// createAITaskBuilderThing will create a new thing
func createAITaskBuilderThing(c client.API, opts CreateOptions, w io.Writer) error {
    // 1. Validate required fields
    if opts.Name == "" {
        return errors.New(ErrNameRequired)
    }
    if opts.WorkspaceID == "" {
        return errors.New(ErrWorkspaceIDRequired)
    }
    
    // 2. Build payload from options
    payload := client.CreateAITaskBuilderThingPayload{
        Name:        opts.Name,
        Description: opts.Description,
        WorkspaceID: opts.WorkspaceID,
    }
    
    // 3. Call API to create
    response, err := c.CreateAITaskBuilderThing(payload)
    if err != nil {
        return err
    }
    
    // 4. Output success with ID
    fmt.Fprintf(w, "Created thing: %s\n", response.AITaskBuilderThing.ID)
    
    // Optional: Display full resource details
    // thing := response.AITaskBuilderThing
    // fmt.Fprintf(w, "Name: %s\n", thing.Name)
    // fmt.Fprintf(w, "Status: %s\n", thing.Status)
    
    return nil
}
```

### Key Characteristics

**Options struct:**
- All input fields from flags
- Includes `Args []string` for positional arguments

**Validation:**
- Validate required fields at start of function
- Return descriptive errors using constants

**Payload construction:**
- Build payload struct from options
- Map flag values to payload fields
- Handle optional fields appropriately

**Success output:**
- Simple confirmation message
- Show created resource ID
- Optionally show key details (but not full resource dump)

**Error handling:**
- Return errors unwrapped (wrapping happens in RunE)
- Include context in error messages

## Pattern 2: Update Command (PATCH)

**Example from codebase:** `cmd/study/increase_places.go`

### File Structure

```go
package aitaskbuilder

import (
    "errors"
    "fmt"
    "io"
    
    "github.com/prolific-oss/cli/client"
    "github.com/prolific-oss/cli/model"
    "github.com/spf13/cobra"
)

// UpdateOptions holds the options for updating a resource
type UpdateOptions struct {
    Args        []string
    Name        string
    Description string
}

// NewUpdateCommand creates a new command for updating resources
func NewUpdateCommand(client client.API, w io.Writer) *cobra.Command {
    var opts UpdateOptions
    
    cmd := &cobra.Command{
        Use:   "update",
        Short: "Update an AI Task Builder thing",
        Long: `Update an existing AI Task Builder thing

Provide the thing ID and the fields you want to update.`,
        Example: `
Update a thing:
$ prolific aitaskbuilder update <thing_id> -n "New Name"
        `,
        Args: cobra.MinimumNArgs(1),  // ← Require thing ID as positional arg
        RunE: func(cmd *cobra.Command, args []string) error {
            opts.Args = args
            
            err := updateAITaskBuilderThing(client, opts, w)
            if err != nil {
                return fmt.Errorf("error: %s", err.Error())
            }
            
            return nil
        },
    }
    
    flags := cmd.Flags()
    flags.StringVarP(&opts.Name, "name", "n", "", "New name")
    flags.StringVarP(&opts.Description, "description", "d", "", "New description")
    
    return cmd
}

// updateAITaskBuilderThing will update an existing thing
func updateAITaskBuilderThing(c client.API, opts UpdateOptions, w io.Writer) error {
    thingID := opts.Args[0]
    
    // 1. Get current resource (for validation or to show changes)
    current, err := c.GetAITaskBuilderThing(thingID)
    if err != nil {
        return fmt.Errorf("unable to fetch current thing: %v", err)
    }
    
    // 2. Validate update is allowed (business logic)
    // Example: Can only update certain fields in certain states
    if current.AITaskBuilderThing.Status == "LOCKED" {
        return errors.New("cannot update thing in LOCKED status")
    }
    
    // 3. Build update payload (only changed fields)
    update := client.UpdateAITaskBuilderThingPayload{}
    
    if opts.Name != "" {
        update.Name = &opts.Name  // ← Pointer for partial update
    }
    if opts.Description != "" {
        update.Description = &opts.Description
    }
    
    // 4. Call API to update
    response, err := c.UpdateAITaskBuilderThing(thingID, update)
    if err != nil {
        return err
    }
    
    // 5. Display updated resource
    thing := response.AITaskBuilderThing
    
    fmt.Fprintf(w, "Updated thing: %s\n", thing.ID)
    fmt.Fprintf(w, "Name: %s\n", thing.Name)
    fmt.Fprintf(w, "Description: %s\n", thing.Description)
    fmt.Fprintf(w, "Status: %s\n", thing.Status)
    fmt.Fprintf(w, "Updated At: %s\n", thing.UpdatedAt.Format("2006-01-02 15:04:05"))
    
    return nil
}
```

### Key Characteristics

**Positional arguments:**
- Use `cobra.MinimumNArgs(1)` to require resource ID
- Access via `opts.Args[0]`

**Pre-fetch validation:**
- Get current resource state
- Validate update constraints (business rules)
- Show before/after if appropriate

**Partial update payload:**
- Only include fields that are being changed
- Use pointers for optional fields
- Check if flags were set before adding to payload

**Success output:**
- Show updated resource
- Highlight what changed
- Include timestamps

## Pattern 3: Action Command (POST)

**Example from codebase:** `cmd/study/transition.go`

### File Structure

```go
package aitaskbuilder

import (
    "errors"
    "fmt"
    "io"
    
    "github.com/prolific-oss/cli/client"
    "github.com/spf13/cobra"
)

// ActionOptions holds the options for performing an action
type ActionOptions struct {
    Args   []string
    Action string
    Silent bool
}

// NewActionCommand creates a new command for performing actions
func NewActionCommand(client client.API, w io.Writer) *cobra.Command {
    var opts ActionOptions
    
    cmd := &cobra.Command{
        Use:   "action",
        Short: "Perform an action on an AI Task Builder thing",
        Long: `Perform an action on an existing thing

Available actions: PUBLISH, ARCHIVE, ACTIVATE`,
        Example: `
Publish a thing:
$ prolific aitaskbuilder action <thing_id> -a PUBLISH

Perform action silently (no output):
$ prolific aitaskbuilder action <thing_id> -a PUBLISH -s
        `,
        Args: cobra.MinimumNArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            opts.Args = args
            
            err := performAction(client, opts, w)
            if err != nil {
                return fmt.Errorf("error: %s", err.Error())
            }
            
            return nil
        },
    }
    
    flags := cmd.Flags()
    flags.StringVarP(&opts.Action, "action", "a", "", "Action to perform (PUBLISH, ARCHIVE, ACTIVATE)")
    flags.BoolVarP(&opts.Silent, "silent", "s", false, "Don't display result")
    
    _ = cmd.MarkFlagRequired("action")
    
    return cmd
}

// performAction will perform an action on a thing
func performAction(c client.API, opts ActionOptions, w io.Writer) error {
    thingID := opts.Args[0]
    
    // 1. Validate action
    validActions := []string{"PUBLISH", "ARCHIVE", "ACTIVATE"}
    if !contains(validActions, opts.Action) {
        return fmt.Errorf("invalid action: %s. Must be one of: %v", opts.Action, validActions)
    }
    
    // 2. Perform action
    _, err := c.PerformActionOnThing(thingID, opts.Action)
    if err != nil {
        return err
    }
    
    // 3. Optionally re-fetch to show updated state
    if !opts.Silent {
        thing, err := c.GetAITaskBuilderThing(thingID)
        if err != nil {
            return err
        }
        
        fmt.Fprintf(w, "Action complete. New state:\n")
        fmt.Fprintf(w, "ID: %s\n", thing.AITaskBuilderThing.ID)
        fmt.Fprintf(w, "Status: %s\n", thing.AITaskBuilderThing.Status)
        fmt.Fprintf(w, "Updated At: %s\n", thing.AITaskBuilderThing.UpdatedAt.Format("2006-01-02 15:04:05"))
    }
    
    return nil
}

func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}
```

### Key Characteristics

**Action validation:**
- Define allowed actions as constants or slice
- Validate action is in allowed list
- Return descriptive error if invalid

**Silent flag:**
- Support `-s` or `--silent` for scripting
- Skip output when silent
- Always perform action, only suppress display

**Re-fetch pattern:**
- Action API may not return full updated resource
- Re-fetch resource to show current state
- Display relevant changed fields

**Success output:**
- Show confirmation of action
- Display new state of resource
- Include timestamp

## Pattern 4: Delete Command (DELETE)

**Not in codebase, but typical pattern:**

### File Structure

```go
package aitaskbuilder

import (
    "errors"
    "fmt"
    "io"
    
    "github.com/prolific-oss/cli/client"
    "github.com/spf13/cobra"
)

// DeleteOptions holds the options for deleting a resource
type DeleteOptions struct {
    Args    []string
    Force   bool
}

// NewDeleteCommand creates a new command for deleting resources
func NewDeleteCommand(client client.API, w io.Writer) *cobra.Command {
    var opts DeleteOptions
    
    cmd := &cobra.Command{
        Use:   "delete",
        Short: "Delete an AI Task Builder thing",
        Long: `Delete an existing AI Task Builder thing

This action cannot be undone.`,
        Example: `
Delete a thing:
$ prolific aitaskbuilder delete <thing_id>

Delete without confirmation:
$ prolific aitaskbuilder delete <thing_id> -f
        `,
        Args: cobra.MinimumNArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            opts.Args = args
            
            err := deleteAITaskBuilderThing(client, opts, w)
            if err != nil {
                return fmt.Errorf("error: %s", err.Error())
            }
            
            return nil
        },
    }
    
    flags := cmd.Flags()
    flags.BoolVarP(&opts.Force, "force", "f", false, "Skip confirmation")
    
    return cmd
}

// deleteAITaskBuilderThing will delete a thing
func deleteAITaskBuilderThing(c client.API, opts DeleteOptions, w io.Writer) error {
    thingID := opts.Args[0]
    
    // 1. Optional: Get resource to show what will be deleted
    if !opts.Force {
        thing, err := c.GetAITaskBuilderThing(thingID)
        if err != nil {
            return err
        }
        
        fmt.Fprintf(w, "This will delete:\n")
        fmt.Fprintf(w, "  ID: %s\n", thing.AITaskBuilderThing.ID)
        fmt.Fprintf(w, "  Name: %s\n", thing.AITaskBuilderThing.Name)
        fmt.Fprintf(w, "\nAre you sure? Use --force to skip this check.\n")
        
        // For a real implementation, you'd prompt for confirmation here
        // For now, we'll just require --force
        return errors.New("confirmation required: use --force to delete")
    }
    
    // 2. Delete resource
    err := c.DeleteAITaskBuilderThing(thingID)
    if err != nil {
        return err
    }
    
    // 3. Confirm deletion
    fmt.Fprintf(w, "Deleted thing: %s\n", thingID)
    
    return nil
}
```

### Key Characteristics

**Force flag:**
- Require `--force` or prompt for confirmation
- Show what will be deleted before confirming
- Safety mechanism for destructive operations

**Confirmation:**
- Display resource details before deletion
- Explicit confirmation required
- Clear warning that action is destructive

**Success output:**
- Simple confirmation message
- Show deleted resource ID
- Don't try to display resource (it's gone)

## Command Naming Conventions

### By Operation Type

```
GET:      get-<resource>, get-<resources>
POST:     create-<resource>, <action>-<resource>
PATCH:    update-<resource>
DELETE:   delete-<resource>
```

### Examples

```go
Use: "get-thing"           // GET single
Use: "get-things"          // GET list
Use: "create-thing"        // POST create
Use: "publish-thing"       // POST action
Use: "archive-thing"       // POST action
Use: "update-thing"        // PATCH update
Use: "delete-thing"        // DELETE
```

## Common Patterns Summary

| Pattern | Validation | Pre-fetch | Payload | Output | Re-fetch |
|---------|------------|-----------|---------|--------|----------|
| **GET** | Required fields | ❌ No | ❌ No | Full details | ❌ No |
| **Create** | Required fields | ❌ No | ✅ Yes | ID + success | ❌ No |
| **Update** | Business rules | ✅ Often | ✅ Partial | Updated resource | ❌ No |
| **Action** | Action valid | ❌ No | ✅ Simple | New state | ✅ Often |
| **Delete** | Confirmation | ✅ Optional | ❌ No | Deleted ID | ❌ No |

## Testing Considerations

Tests for non-GET operations should include:

1. **Basic structure test** (same as GET)
2. **Successful operation test** with payload validation
3. **Validation test** (required fields, business rules)
4. **Error handling test**
5. **Optional flags test** (silent, force, etc.)

Example for create:
```go
func TestNewCreateCommandCallsAPI(t *testing.T) {
    // Setup mock
    payload := client.CreateAITaskBuilderThingPayload{
        Name:        "Test Thing",
        Description: "Test Description",
        WorkspaceID: "workspace-123",
    }
    
    response := client.CreateAITaskBuilderThingResponse{
        AITaskBuilderThing: model.AITaskBuilderThing{
            ID:   "thing-456",
            Name: "Test Thing",
        },
    }
    
    c.EXPECT().
        CreateAITaskBuilderThing(gomock.Eq(payload)).
        Return(&response, nil).
        AnyTimes()
    
    // Execute and verify...
}
```

## Reference

See [HTTP-METHODS-GUIDE.md](HTTP-METHODS-GUIDE.md) for complete examples from the codebase.
