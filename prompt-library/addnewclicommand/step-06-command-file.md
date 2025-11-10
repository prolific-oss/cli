# Step 6: Create Command Implementation File

## Purpose

Create the CLI command implementation that users will interact with. This step:
- Builds the user-facing command
- Implements flag handling and validation
- Formats output for the terminal
- Integrates with the API layer from Step 5

## AI Implementation Prompt

```
I need to create a new CLI command file for the aitaskbuilder command group.

Create `cmd/aitaskbuilder/get_<resource>.go` with:
- Options struct for command flags
- NewGet<Resource>Command constructor function
- render<Resource> function for output formatting

Command details:
- Resource name: [SPECIFY]
- Required flags: [SPECIFY, e.g., "thing-id"]
- Optional flags: [SPECIFY if any]
- API method to call: [From Step 5]
- Fields to display: [SPECIFY what to show in output]

File naming: Use snake_case with underscores (e.g., get_new_thing.go)
Command.Use: Use kebab-case with dashes (e.g., "get-new-thing")

Follow the pattern in cmd/aitaskbuilder/get_batch.go
```

## AI Implementation Guidance

### File Naming

**✅ CORRECT:**
- `get_new_thing.go`
- `get_batch_status.go`
- `list_things.go`

**❌ INCORRECT:**
- `get-new-thing.go` (dashes)
- `getnewthing.go` (no separators)
- `GetNewThing.go` (PascalCase)

### File Structure

```go
package aitaskbuilder

import (
    "errors"
    "fmt"
    "io"
    
    "github.com/prolific-oss/cli/client"
    "github.com/spf13/cobra"
    // "github.com/spf13/viper" // Only if using workspace defaults
)

// 1. OPTIONS STRUCT
type NewThingGetOptions struct {
    Args    []string
    ThingID string
}

// 2. COMMAND CONSTRUCTOR
func NewGetNewThingCommand(client client.API, w io.Writer) *cobra.Command {
    var opts NewThingGetOptions
    
    cmd := &cobra.Command{
        Use:   "get-new-thing",  // ← kebab-case
        Short: "Get an AI Task Builder thing",
        Long: `Get details about a specific AI Task Builder thing

This command allows you to retrieve details of a specific thing by providing
the thing ID.`,
        Example: `
Get a thing:
$ prolific aitaskbuilder get-new-thing -t <thing_id>
        `,
        RunE: func(cmd *cobra.Command, args []string) error {
            opts.Args = args
            
            err := renderAITaskBuilderNewThing(client, opts, w)
            if err != nil {
                return fmt.Errorf("error: %s", err.Error())  // ← Wrap here
            }
            
            return nil
        },
    }
    
    flags := cmd.Flags()
    flags.StringVarP(&opts.ThingID, "thing-id", "t", "", "Thing ID (required) - Description.")
    
    _ = cmd.MarkFlagRequired("thing-id")
    
    return cmd
}

// 3. RENDER FUNCTION
func renderAITaskBuilderNewThing(c client.API, opts NewThingGetOptions, w io.Writer) error {
    // Validation
    if opts.ThingID == "" {
        return errors.New(ErrThingIDRequired)
    }
    
    // API call
    response, err := c.GetAITaskBuilderNewThing(opts.ThingID)
    if err != nil {
        return err  // ← Return unwrapped
    }
    
    thing := response.AITaskBuilderNewThing
    
    // Output formatting
    fmt.Fprintf(w, "AI Task Builder Thing Details:\n")
    fmt.Fprintf(w, "ID: %s\n", thing.ID)
    fmt.Fprintf(w, "Name: %s\n", thing.Name)
    fmt.Fprintf(w, "Status: %s\n", thing.Status)
    fmt.Fprintf(w, "Created At: %s\n", thing.CreatedAt.Format("2006-01-02 15:04:05"))
    
    // Conditional output for optional fields
    if len(thing.Items) > 0 {
        fmt.Fprintf(w, "Items: %d\n", len(thing.Items))
        for i, item := range thing.Items {
            fmt.Fprintf(w, "  Item %d: %s\n", i+1, item.Name)
        }
    }
    
    return nil
}
```

### Critical Patterns

**1. Command.Use field:**
```go
// ✅ CORRECT - kebab-case for consistency
Use: "get-new-thing",

// ⚠️ INCONSISTENT - but exists in codebase
Use: "getnewthing",  // No separator

// ❌ WRONG
Use: "get_new_thing",  // Underscores
Use: "GetNewThing",    // PascalCase
```

**2. Error handling in RunE:**
```go
// ✅ CORRECT - Wrap errors in RunE
RunE: func(cmd *cobra.Command, args []string) error {
    opts.Args = args
    
    err := renderFunction(client, opts, w)
    if err != nil {
        return fmt.Errorf("error: %s", err.Error())  // Wrap HERE
    }
    
    return nil
}

// ❌ WRONG - Not wrapping
RunE: func(cmd *cobra.Command, args []string) error {
    return renderFunction(client, opts, w)  // Not wrapped
}
```

**3. Error handling in render function:**
```go
// ✅ CORRECT - Return unwrapped errors
func renderFunction(...) error {
    if opts.ThingID == "" {
        return errors.New(ErrThingIDRequired)  // Unwrapped
    }
    
    response, err := c.GetAPI()
    if err != nil {
        return err  // Unwrapped
    }
    
    return nil
}

// ❌ WRONG - Wrapping in render function
func renderFunction(...) error {
    if err != nil {
        return fmt.Errorf("error: %s", err)  // Don't wrap here
    }
}
```

**4. Output formatting:**
```go
// ✅ CORRECT - Using fmt.Fprintf with writer
fmt.Fprintf(w, "Details:\n")
fmt.Fprintf(w, "ID: %s\n", thing.ID)
fmt.Fprintf(w, "Created At: %s\n", thing.CreatedAt.Format("2006-01-02 15:04:05"))

// ❌ WRONG
fmt.Printf("ID: %s\n", thing.ID)           // Not using writer
fmt.Println("ID:", thing.ID)                // Not using writer
print("ID:", thing.ID)                      // Wrong function
```

**5. Flag definitions:**
```go
// ✅ CORRECT
flags := cmd.Flags()
flags.StringVarP(&opts.ThingID, "thing-id", "t", "", "Thing ID (required) - Description")
_ = cmd.MarkFlagRequired("thing-id")

// ✅ CORRECT - With workspace default
flags.StringVarP(&opts.WorkspaceID, "workspace-id", "w", viper.GetString("workspace"), "Workspace ID")

// ❌ WRONG
flags.StringVarP(&opts.ThingID, "thingId", "t", "", "Thing ID")  // camelCase
flags.StringVarP(&opts.ThingID, "thing_id", "t", "", "Thing ID")  // underscores
cmd.MarkFlagRequired("thing-id")  // Not discarding return
```

**6. Date formatting:**
```go
// ✅ CORRECT - Consistent format
thing.CreatedAt.Format("2006-01-02 15:04:05")

// ❌ WRONG
thing.CreatedAt.String()  // Inconsistent format
thing.CreatedAt.Format("2006-01-02T15:04:05Z")  // ISO format (inconsistent)
```

### What NOT to Do

❌ **Don't:**
- Name file with dashes (`get-new-thing.go`)
- Use `Command.Use: "get_new_thing"` (underscores)
- Wrap errors in render function
- Use `fmt.Printf` instead of `fmt.Fprintf(w, ...)`
- Forget to mark required flags
- Use inconsistent date formats

✅ **Do:**
- Name file with underscores (`get_new_thing.go`)
- Use `Command.Use: "get-new-thing"` (kebab-case)
- Wrap errors only in RunE
- Use `fmt.Fprintf(w, ...)` for all output
- Mark required flags and discard return
- Use standard date format

## Human Review Criteria

### File Naming Check

```bash
# Verify file exists with correct name
ls -la cmd/aitaskbuilder/get_new_thing.go
```

**Expected:**
```
✅ cmd/aitaskbuilder/get_new_thing.go (snake_case with underscores)
```

### Code Review Checklist

**File structure:**
- [ ] Package: `package aitaskbuilder`
- [ ] Imports: Organized correctly
- [ ] Options struct: At top of file
- [ ] Command constructor: Present with correct signature
- [ ] Render function: Present with correct signature

**Options struct:**
- [ ] Name follows pattern: `<Resource><Action>Options`
- [ ] Has `Args []string` field
- [ ] Has fields for all flags
- [ ] All fields exported (PascalCase)

**Command constructor:**
- [ ] Signature: `func NewGet<Resource>Command(client client.API, w io.Writer) *cobra.Command`
- [ ] Declares `var opts` at start
- [ ] `Use`: kebab-case (e.g., `"get-new-thing"`)
- [ ] `Short`: Present and descriptive
- [ ] `Long`: Present with detailed description
- [ ] `Example`: Shows usage with `$` prompt
- [ ] `RunE`: Calls render function and wraps errors
- [ ] Flags: Defined and required flags marked

**Render function:**
- [ ] Signature: `func render<Resource>(c client.API, opts <OptionsType>, w io.Writer) error`
- [ ] Validates required fields
- [ ] Calls API method from Step 5
- [ ] Returns unwrapped errors
- [ ] Uses `fmt.Fprintf(w, ...)` for output
- [ ] Formats dates consistently
- [ ] Handles optional fields conditionally

### Specific Checks

Run these verification commands:

```bash
# 1. File exists with correct name
ls -la cmd/aitaskbuilder/get_new_thing.go

# 2. Check Use field
grep "Use:" cmd/aitaskbuilder/get_new_thing.go

# 3. Check command function exists
grep "func NewGet.*Command" cmd/aitaskbuilder/get_new_thing.go

# 4. Check render function exists
grep "func render.*AITaskBuilder" cmd/aitaskbuilder/get_new_thing.go

# 5. Build the command
go build ./cmd/aitaskbuilder/...
```

### Expected Results

**1. File exists:**
```
✅ -rw-r--r-- ... cmd/aitaskbuilder/get_new_thing.go
```

**2. Use field:**
```
✅ Use: "get-new-thing",
```

**3. Command function:**
```
✅ func NewGetNewThingCommand(client client.API, w io.Writer) *cobra.Command {
```

**4. Render function:**
```
✅ func renderAITaskBuilderNewThing(c client.API, opts NewThingGetOptions, w io.Writer) error {
```

**5. Build:**
```
✅ Compiles successfully
```

## Common Issues

### Issue: File name has dashes
**Symptom**: File is `get-new-thing.go`  
**Fix**: Rename to `get_new_thing.go`

### Issue: Command.Use has underscores
**Symptom**: `Use: "get_new_thing"`  
**Fix**: Change to `Use: "get-new-thing"`

### Issue: Errors wrapped in render function
**Symptom**: `return fmt.Errorf("error: %s", err)` in render  
**Fix**: Return unwrapped: `return err`

### Issue: Using fmt.Printf
**Symptom**: `fmt.Printf("ID: %s\n", thing.ID)`  
**Fix**: Use `fmt.Fprintf(w, "ID: %s\n", thing.ID)`

### Issue: Required flag not marked
**Symptom**: Flag validation doesn't work  
**Fix**: Add `_ = cmd.MarkFlagRequired("flag-name")`

### Issue: Inconsistent date format
**Symptom**: Different format than other commands  
**Fix**: Use `"2006-01-02 15:04:05"`

## Success Criteria

- [ ] File named correctly with underscores
- [ ] Package declaration: `package aitaskbuilder`
- [ ] Options struct defined
- [ ] Command constructor exists with correct signature
- [ ] `Use` field is kebab-case
- [ ] RunE wraps errors with `fmt.Errorf("error: %s", ...)`
- [ ] Render function returns unwrapped errors
- [ ] All output uses `fmt.Fprintf(w, ...)`
- [ ] Required flags marked
- [ ] Date format is consistent
- [ ] `go build ./cmd/aitaskbuilder/...` succeeds

## Next Step

Proceed to [Step 7: Register Command](step-07-register-command.md)
