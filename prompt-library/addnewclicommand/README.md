# Add New CLI Command - Prompt Library

This directory contains a step-by-step prompt library for adding a new command to the AITaskBuilder command group in the Prolific CLI.

## Purpose

These prompts are optimized for AI coding agents to implement CLI commands incrementally, with human review checkpoints at each stage.

## Supports All HTTP Methods

This library provides patterns for **all** HTTP methods:
- **GET** - Retrieve resources (list, view)
- **POST** - Create resources, perform actions, duplicate
- **PATCH** - Update resources (partial updates)
- **PUT** - Replace resources (not used in current codebase)
- **DELETE** - Delete resources (not used in current codebase)

See [HTTP-METHODS-GUIDE.md](HTTP-METHODS-GUIDE.md) for comprehensive examples and patterns for each method.

## Implementation Flow

The implementation is divided into **5 checkpoints** where the code should build and run:

- ✅ **Checkpoint 1** (after Step 4): Code builds, existing tests pass
- ✅ **Checkpoint 2** (after Step 7): New command builds and can be run manually  
- ✅ **Checkpoint 3** (after Step 8): New command is linted and verified
- ✅ **Checkpoint 4** (after Step 10): Full test coverage verified
- ✅ **Checkpoint 5** (after Step 11): Complete and documented

## Essential Reference Documents

Before starting, review these guides:
- [**HTTP Methods Guide**](HTTP-METHODS-GUIDE.md) - Comprehensive patterns for GET, POST, PATCH, DELETE
- [**Step 6 Addendum**](step-06-create-update-commands.md) - Create, Update, and Action command patterns

## Steps

### Checkpoint 1: Foundation (Steps 1-4)

1. [**Step 1: Define Response/Payload Structs**](step-01-response-struct.md) - Data structures for requests and responses
2. [**Step 2: Add Constants**](step-02-constants.md) - Error messages and constants
3. [**Step 3: Add API Method Signature**](step-03-api-interface.md) - Interface definition (breaks build)
4. [**Step 4: Regenerate Mocks**](step-04-regenerate-mocks.md) - Fix build

**Verify:** Run `make test` - all existing tests pass

### Checkpoint 2: Feature Complete (Steps 5-7)

5. [**Step 5: Implement API Client Method**](step-05-api-implementation.md) - API layer (supports all HTTP methods)
6. [**Step 6: Create Command Implementation**](step-06-command-file.md) - CLI layer (GET operations)
   - See also: [**Step 6 Addendum**](step-06-create-update-commands.md) - Create/Update/Action patterns
7. [**Step 7: Register Command**](step-07-register-command.md) - Integration

**Verify:** Run `make build` and `./prolific aitaskbuilder <command> --help`

### Checkpoint 3: Quality Assured (Step 8)

8. [**Step 8: Run Linter**](step-08-run-linter.md) - Code quality check

**Verify:** Run `make lint` - zero errors

### Checkpoint 4: Fully Tested (Steps 9-10)

9. [**Step 9: Create Test File**](step-09-test-file.md) - Comprehensive tests
10. [**Step 10: Run Full Test Suite**](step-10-run-tests.md) - Verify coverage

**Verify:** Run `make test` - all tests pass including new ones

### Checkpoint 5: Complete (Step 11)

11. [**Step 11: Update Documentation**](step-11-documentation.md) - CHANGELOG (optional)

**Verify:** Run `make clean && make all` - complete build successful

## Usage

### For AI Coding Agents

1. Start with Step 1
2. Follow the "AI Implementation Guidance" in each step
3. Output code according to the patterns shown
4. Announce when checkpoint is reached
5. Wait for human verification before proceeding

### For Human Reviewers

1. Review the "Human Review Criteria" section in each step
2. Run the verification commands listed
3. Check success criteria before approving
4. Provide feedback if criteria not met
5. Approve proceeding to next step only when checkpoint passes

## File Structure

Each step file contains:

```markdown
# Step N: Title

## Purpose
Why this step is done at this point in the sequence

## AI Implementation Prompt
Ready-to-use prompt for the AI agent

## AI Implementation Guidance
Detailed instructions and patterns for the AI

## Human Review Criteria
What the human reviewer should check

## Verification Commands
Commands to run to verify the step

## Expected Results
What success looks like

## Common Issues
Typical problems and how to identify them

## Success Criteria
Checklist of requirements to proceed
```

## Quick Reference

| Step | File | Breaks Build? | Testable? |
|------|------|---------------|-----------|
| 1 | Response struct | ❌ No | ❌ No |
| 2 | Constants | ❌ No | ❌ No |
| 3 | API interface | ✅ Yes | ❌ No |
| 4 | Regenerate mocks | ✅ Fixes | ✅ Yes |
| 5 | API implementation | ❌ No | ✅ Yes |
| 6 | Command file | ❌ No | ❌ No |
| 7 | Register command | ❌ No | ✅ Manual |
| 8 | Lint | ❌ No | ✅ Yes |
| 9 | Test file | ❌ No | ❌ No |
| 10 | Run tests | ❌ No | ✅ Yes |
| 11 | Documentation | ❌ No | ✅ Manual |

## Example Usage

**AI Agent:**
```
I will implement Step 1: Define Response Struct

[Reads step-01-response-struct.md]
[Implements according to guidance]
[Outputs code]

Step 1 complete. Ready for human review.
```

**Human Reviewer:**
```
[Reads "Human Review Criteria" section]
[Runs verification commands]
[Checks success criteria]

✅ Approved. Proceed to Step 2.
```

## Contributing

When adding new steps or modifying existing ones:

1. Keep each step focused and atomic
2. Include clear success criteria
3. Provide exact code patterns
4. List all verification commands
5. Document common failure modes

## Notes

- Steps are ordered for incremental review
- Each checkpoint represents a working state
- Never skip verification steps
- If any step fails, fix before proceeding
