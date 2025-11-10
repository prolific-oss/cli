# Step 11: Update Documentation (Optional)

## Purpose

Update the CHANGELOG to document the new command. This step:
- Records what was added for release notes
- Follows project documentation conventions
- **Is optional** - only done when explicitly requested or preparing a release

## AI Implementation Prompt

```
I need to update the CHANGELOG to document the new command.

Update `CHANGELOG.md`:
- Add entry under the "## next" section
- Format: "- Add `aitaskbuilder <command>` command to <description>."
- Use backticks around command name
- Start with action verb (Add, Update, Fix, Remove)
- Keep it concise (one line)
- No dates (follows project convention)

Command added: [SPECIFY]
Brief description: [SPECIFY what it does]

Example:
- Add `aitaskbuilder get-new-thing` command to retrieve thing details.
```

## AI Implementation Guidance

### When to Do This Step

✅ **Do this when:**
- User explicitly requests documentation update
- Preparing for a release
- Contributing via PR that requires changelog
- Project workflow requires it

❌ **Skip this when:**
- Just testing/experimenting
- Internal development
- Not explicitly requested

### Where to Add

Add under the `## next` section in `CHANGELOG.md`:

```markdown
## next

- Add `aitaskbuilder get-new-thing` command to retrieve thing details.
```

### Format Rules

**Pattern:** `- <Verb> <what> <purpose>.`

**✅ CORRECT:**
```markdown
- Add `aitaskbuilder get-new-thing` command to retrieve thing details.
- Add `aitaskbuilder list-things` command to list all things.
- Update `aitaskbuilder getbatch` to include new metadata field.
- Fix error handling in `aitaskbuilder getbatchstatus` command.
```

**❌ INCORRECT:**
```markdown
- aitaskbuilder get-new-thing command          # Missing backticks, no verb
- Added get-new-thing command                  # Past tense, not specific enough
- Add get-new-thing                            # Not enough context
- Add `aitaskbuilder get-new-thing` command.   # Missing description
- [2025-01-01] Add command                     # Has date (against convention)
- Add aitaskbuilder get-new-thing              # Missing backticks
```

### Verb Choices

**Use present tense, imperative mood:**

- **Add** - For new features/commands
- **Update** - For enhancements to existing features
- **Fix** - For bug fixes
- **Remove** - For deprecated features
- **Change** - For behavior modifications

**Don't use past tense:**
- ❌ Added, Updated, Fixed, Removed, Changed

### What NOT to Do

❌ **Don't:**
- Add dates to entries (project convention: no dates)
- Use past tense verbs
- Be too verbose (keep it one line)
- Forget backticks around command names
- Add to wrong section (must be under `## next`)

✅ **Do:**
- Keep it concise and clear
- Use backticks for command names
- Use present tense
- Describe what the command does
- Follow existing entry style

## Human Review Criteria

### Verification Commands

```bash
# 1. Check CHANGELOG exists
ls -la CHANGELOG.md

# 2. Verify ## next section exists
grep "## next" CHANGELOG.md

# 3. View the changes
git diff CHANGELOG.md

# 4. Check formatting
grep -A 5 "## next" CHANGELOG.md
```

### Expected Results

**1. File exists:**
```
✅ -rw-r--r-- ... CHANGELOG.md
```

**2. Section exists:**
```
✅ ## next
```

**3. Git diff:**
```diff
 ## next
 
+- Add `aitaskbuilder get-new-thing` command to retrieve thing details.
```

**4. Formatted correctly:**
```
✅ ## next

- Add `aitaskbuilder get-new-thing` command to retrieve thing details.
```

### Code Review Checklist

- [ ] **Correct file**: Modified `CHANGELOG.md`
- [ ] **Correct section**: Entry under `## next`
- [ ] **Format**: Bullet point with `-`
- [ ] **Verb**: Present tense (Add, Update, Fix, etc.)
- [ ] **Backticks**: Command name in backticks
- [ ] **Concise**: One line, clear description
- [ ] **No date**: Does not include date
- [ ] **Style**: Matches existing entries

### Compare with Existing Entries

Check existing CHANGELOG format:

```bash
# View recent entries
head -20 CHANGELOG.md
```

**Ensure new entry matches the style:**

```markdown
## next

- Add Apache 2 License.
- Add `aitaskbuilder` command to the root of the application.
- Bump the project to Go 1.25.

## 0.0.56

- Remove the slack notification from releases.
```

## Common Issues

### Issue: Wrong section
**Symptom**: Entry added to versioned section instead of `## next`  
**Fix:** Move entry under `## next` section

### Issue: No backticks
**Symptom**: Command name not formatted: `aitaskbuilder get-new-thing`  
**Fix:** Add backticks: `` `aitaskbuilder get-new-thing` ``

### Issue: Past tense
**Symptom**: "Added command" or "Updated feature"  
**Fix:** Use present tense: "Add command", "Update feature"

### Issue: Too verbose
**Symptom**: Multiple sentences or long description  
**Fix:** Condense to one concise line

### Issue: Missing description
**Symptom**: "Add `aitaskbuilder get-new-thing` command."  
**Fix:** Add purpose: "Add `aitaskbuilder get-new-thing` command to retrieve thing details."

### Issue: Includes date
**Symptom**: "[2025-01-01] Add command"  
**Fix:** Remove date (project convention)

## Success Criteria

- [ ] CHANGELOG.md is modified
- [ ] Entry is under `## next` section
- [ ] Format: `- Add `command` to <description>.`
- [ ] Uses present tense verb
- [ ] Has backticks around command name
- [ ] Concise one-line description
- [ ] No date included
- [ ] Matches style of existing entries
- [ ] Git diff shows only the new entry

## Checkpoint 5: Final Verification

✅ **CHECKPOINT 5 REACHED**

The implementation is complete. Run final verification:

```bash
# 1. Clean build from scratch
make clean
make all

# 2. Verify all tests pass
make test

# 3. Verify linter passes
make lint

# 4. Verify command works
./prolific aitaskbuilder --help | grep get-new-thing
./prolific aitaskbuilder get-new-thing --help

# 5. Check all changes
git status
git diff --stat
```

### Expected Results

**All commands succeed:**
- ✅ `make clean` - Cleans successfully
- ✅ `make all` - Builds successfully
- ✅ `make test` - All tests pass
- ✅ `make lint` - No errors
- ✅ Command appears in help
- ✅ Command help displays correctly

**Git shows expected changes:**
```
✅ Modified: 
    client/responses.go
    cmd/aitaskbuilder/constants.go
    client/client.go
    mock_client/mock_client.go
    cmd/aitaskbuilder/aitaskbuilder.go
    CHANGELOG.md (optional)

✅ New files:
    cmd/aitaskbuilder/get_new_thing.go
    cmd/aitaskbuilder/get_new_thing_test.go
```

### Final Checklist

**Code Quality:**
- [ ] All files follow project conventions
- [ ] No debug code or comments
- [ ] All functions documented
- [ ] Error handling is consistent

**Testing:**
- [ ] All tests pass
- [ ] Coverage is complete
- [ ] No flaky tests
- [ ] All scenarios covered

**Documentation:**
- [ ] Command has help text
- [ ] Examples included
- [ ] CHANGELOG updated (if applicable)

**Integration:**
- [ ] Command registered
- [ ] Appears in help
- [ ] Works correctly
- [ ] Error messages clear

**Git Hygiene:**
- [ ] Only expected files changed
- [ ] No untracked files (except new ones)
- [ ] Changes are focused

### Success Criteria

All must pass:
- [ ] `make clean && make all` succeeds
- [ ] `make test` passes (0 failures)
- [ ] `make lint` passes (0 errors)
- [ ] Command appears and works
- [ ] Git changes are clean and expected
- [ ] All checkpoints were passed

## Ready for Review/Commit

The implementation is complete and ready for:
- Code review
- Pull request
- Commit to repository
- Release (if applicable)

**Congratulations!** You have successfully added a new CLI command following all project conventions and best practices.
