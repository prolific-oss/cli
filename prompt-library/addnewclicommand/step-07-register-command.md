# Step 7: Register Command

## Purpose

Register the new command with the parent aitaskbuilder command so it appears in the CLI. This step:
- Makes the command accessible to users
- Integrates with the command hierarchy
- Enables `prolific aitaskbuilder get-new-thing` usage

## AI Implementation Prompt

```
I need to register the new command with the parent aitaskbuilder command.

Update `cmd/aitaskbuilder/aitaskbuilder.go`:
- Location: Within the cmd.AddCommand() block
- Add: NewGet<Resource>Command(client, w),
- Parameters: client, w (no commandName string for aitaskbuilder)
- Include trailing comma

Command function name: [From Step 6, e.g., NewGetNewThingCommand]
```

## AI Implementation Guidance

### Where to Add

Add within the `cmd.AddCommand()` call in `NewAITaskBuilderCommand`:

```go
func NewAITaskBuilderCommand(client client.API, w io.Writer) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "aitaskbuilder",
        Short: "AI Task Builder tools and utilities",
        Long:  "Manage AI task building workflows, datasets, and batch operations for the Prolific platform",
    }
    
    cmd.AddCommand(
        NewGetBatchCommand(client, w),
        NewGetBatchStatusCommand(client, w),
        NewGetBatchesCommand(client, w),
        NewGetResponsesCommand(client, w),
        NewGetDatasetStatusCommand(client, w),
        NewGetNewThingCommand(client, w),  // ← Add here
    )
    
    return cmd
}
```

### Pattern to Follow

**✅ CORRECT:**
```go
cmd.AddCommand(
    NewGetBatchCommand(client, w),
    NewGetNewThingCommand(client, w),  // ← Your new command
)
```

**Key points:**
- Function name matches Step 6 exactly
- Parameters are `client, w` (in that order)
- Has trailing comma
- No `commandName` string parameter (different from study/workspace commands)

### What NOT to Do

❌ **Don't:**
```go
// Wrong - Missing trailing comma
cmd.AddCommand(
    NewGetNewThingCommand(client, w)
)

// Wrong - Adding commandName parameter
cmd.AddCommand(
    NewGetNewThingCommand("get-new-thing", client, w),
)

// Wrong - Missing w parameter
cmd.AddCommand(
    NewGetNewThingCommand(client),
)

// Wrong - Wrong function name
cmd.AddCommand(
    getNewThingCommand(client, w),  // Lowercase
)

// Wrong - Added outside AddCommand
cmd.AddCommand(...)
NewGetNewThingCommand(client, w)  // This won't work
```

✅ **Do:**
```go
cmd.AddCommand(
    NewGetNewThingCommand(client, w),  // Correct
)
```

## Human Review Criteria

### Code Review Checklist

- [ ] **File**: Modified `cmd/aitaskbuilder/aitaskbuilder.go`
- [ ] **Location**: Within `cmd.AddCommand()` block
- [ ] **Function name**: Matches command constructor from Step 6
- [ ] **Parameters**: `client, w` only (no commandName)
- [ ] **Trailing comma**: Present
- [ ] **Ordering**: Alphabetical or logical (optional but nice)

### Verification Commands

```bash
# 1. Verify registration
grep -A 10 "cmd.AddCommand" cmd/aitaskbuilder/aitaskbuilder.go

# 2. Build CLI
make build

# 3. Check command appears in help
./prolific aitaskbuilder --help

# 4. Check command help works
./prolific aitaskbuilder get-new-thing --help
```

### Expected Results

**1. Grep output:**
```
✅ Shows your command in AddCommand list:
cmd.AddCommand(
    NewGetBatchCommand(client, w),
    ...
    NewGetNewThingCommand(client, w),
    ...
)
```

**2. Build output:**
```
✅ Build succeeds
✅ Binary created: ./prolific
```

**3. aitaskbuilder help:**
```
✅ Command appears in list:

Available Commands:
  ...
  get-new-thing    Get an AI Task Builder thing
  ...

Use "prolific aitaskbuilder [command] --help" for more information about a command.
```

**4. Command help:**
```
✅ Displays command help:

Get an AI Task Builder thing

Usage:
  prolific aitaskbuilder get-new-thing [flags]

Flags:
  -h, --help             help for get-new-thing
  -t, --thing-id string  Thing ID (required) - The ID of the thing to retrieve.
```

### Specific Checks

**Git diff should show:**
```diff
 cmd.AddCommand(
     NewGetBatchCommand(client, w),
     NewGetBatchStatusCommand(client, w),
     NewGetBatchesCommand(client, w),
     NewGetResponsesCommand(client, w),
     NewGetDatasetStatusCommand(client, w),
+    NewGetNewThingCommand(client, w),
 )
```

## Common Issues

### Issue: Command doesn't appear in help
**Symptom**: `./prolific aitaskbuilder --help` doesn't show new command  
**Cause**: Not added to AddCommand, or syntax error  
**Fix**: Verify it's inside `cmd.AddCommand()` block with correct syntax

### Issue: Build fails
**Symptom**: Compilation error after adding registration  
**Cause**: Function name typo, missing comma, or wrong parameters  
**Fix**: Check function name matches Step 6, has trailing comma

### Issue: Command help doesn't work
**Symptom**: `./prolific aitaskbuilder get-new-thing --help` fails  
**Cause**: Command not actually registered, or constructor has errors  
**Fix**: Rebuild and check for errors in Step 6 implementation

### Issue: Wrong parameters
**Symptom**: Build error about argument mismatch  
**Cause**: Added commandName parameter or wrong order  
**Fix**: Use `(client, w)` only - no commandName for aitaskbuilder

## Success Criteria

- [ ] Modified `cmd/aitaskbuilder/aitaskbuilder.go`
- [ ] Added within `cmd.AddCommand()` block
- [ ] Function name matches Step 6 constructor
- [ ] Parameters are `client, w` only
- [ ] Has trailing comma
- [ ] `make build` succeeds
- [ ] Command appears in `./prolific aitaskbuilder --help`
- [ ] `./prolific aitaskbuilder get-new-thing --help` displays correctly

## Manual Testing (Optional)

```bash
# Test flag validation
./prolific aitaskbuilder get-new-thing

# Expected: Error about missing required flag "thing-id"
```

## Checkpoint 2: Verification

✅ **CHECKPOINT 2 REACHED**

At this point, the command should be fully accessible:

```bash
# Verify build
make clean
make build

# Verify command appears
./prolific aitaskbuilder --help | grep get-new-thing

# Verify command help
./prolific aitaskbuilder get-new-thing --help

# Test required flag validation
./prolific aitaskbuilder get-new-thing
```

**Expected:** Command builds, appears in help, help displays correctly, flag validation works.

**If this fails, do not proceed to Step 8. Fix issues first.**

## Next Step

Once Checkpoint 2 passes, proceed to [Step 8: Run Linter](step-08-run-linter.md)
