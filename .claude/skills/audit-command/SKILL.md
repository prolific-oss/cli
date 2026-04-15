---
name: audit-command
description: Audit an existing CLI command for correctness against established patterns.
argument-hint: "[resource] [command]"
user-invocable: true
---

# CLI Command Reviewer

## When to Use This Skill

Invoke this skill when the user:
- Asks to "review a command", "check a command", or "audit a command"
- Wants to verify a command conforms to project conventions
- Uses the slash command /review-command

---

## Arguments

Arguments may be provided via `$ARGUMENTS` or gathered interactively.

**Expected arguments:**
- `resource` - Resource name (e.g., collection, study, workspace)
- `command` - Command name (e.g., list, get, create, update). Omit to review all commands under the resource.

**If arguments are missing, use the ask_user tool to gather them.**

---

## Review Checklist

Work through each section below. For each item, check the relevant source files in
`cmd/{resource}/` and `ui/{resource}/`. Report findings at the end grouped by severity:
- 🔴 **Must fix** — non-conformant, will cause inconsistent UX or compile errors
- 🟡 **Should fix** — diverges from convention, not blocking
- 🟢 **OK** — conforms to the pattern

---

### 1. Output Flag (list commands only)

**What to check:** Every list command must register the output flags via `shared.AddOutputFlags`.

**Expected pattern** in `cmd/{resource}/list.go` (or equivalent list file):

```go
// Options struct must embed OutputOptions
type ListOptions struct {
    Output shared.OutputOptions
    // ... other fields
}

// At the bottom of the constructor, after all other flags:
shared.AddOutputFlags(cmd, &opts.Output)
```

**How to verify:**
1. Read the command file
2. Confirm `shared.OutputOptions` is present in the options struct
3. Confirm `shared.AddOutputFlags(cmd, &opts.Output)` is called in the constructor
4. Confirm `shared.ResolveFormat(opts.Output)` drives the rendering switch (see §3)

**Why it matters:** Without this, users cannot use `--json`, `--csv`, `--table`, or `-n` flags,
which breaks scripting and pipeline use cases.

---

### 2. Dependency Injection

**What to check:** Command constructors must use dependency injection — never construct
a real client or writer internally.

**Expected constructor signatures:**

```go
// Command with a hardcoded Use string
func New{Action}Command(client client.API, w io.Writer) *cobra.Command

// Command where the caller supplies the Use string
func New{Action}Command(commandName string, client client.API, w io.Writer) *cobra.Command
```

Check which pattern the other commands in the same package use and be consistent.

**What to flag:**
- Constructors that accept concrete `*client.Client` instead of `client.API`
- Any use of `os.Stdout` inside a command file (should always use the injected `w`)
- Constructors that create their own HTTP client or read `PROLIFIC_TOKEN` directly

---

### 3. Rendering Strategy (list commands only)

**What to check:** List commands must use `shared.ResolveFormat` to select the renderer,
and must support all four output modes.

**Expected switch in `RunE`:**

```go
format := shared.ResolveFormat(opts.Output)
switch format {
case "json":
    r := ui.JSONRenderer[model.{Resource}]{}
    if err := r.Render(results, w); err != nil {
        return fmt.Errorf("error: %s", err)
    }
case "csv":
    r := ui.CsvRenderer[model.{Resource}]{}
    if err := r.Render(results, fields, w); err != nil {
        return fmt.Errorf("error: %s", err)
    }
case "table":
    r := ui.TableRenderer[model.{Resource}]{}
    if err := r.Render(results, fields, w); err != nil {
        return fmt.Errorf("error: %s", err)
    }
default:
    r := &InteractiveRenderer{}
    if err := r.Render(client, results, w); err != nil {
        return fmt.Errorf("error: %s", err)
    }
}
```

**What to flag:**
- Hardcoded `tabwriter` output instead of going through a renderer
- Missing `json` or `csv` cases in the switch
- Using a boolean flag like `nonInteractive bool` instead of `shared.OutputOptions`
- `InteractiveRenderer` called unconditionally (no format check)

Reference: `cmd/study/list.go`, `cmd/collection/list.go`

---

### 4. Help Text & Examples

**What to check:** Every command must have `Short`, `Long`, and `Example` populated.

| Field     | Requirement |
|-----------|-------------|
| `Short`   | One sentence, imperative, no trailing period |
| `Long`    | Paragraph describing what the command does and when to use it |
| `Example` | At least one `$ prolific ...` invocation per major flag combination |

**For list commands**, examples must cover:
- Interactive (no flags)
- Table output: `--table` / `-t`
- CSV output: `--csv` / `-c`
- JSON output: `--json` / `-j`
- Any resource-specific flags (e.g., `--workspace`, `--status`)

**For view commands**, examples must cover:
- Basic usage with an ID argument
- `--web` / `-W` flag (if implemented)

**For create/update commands**, examples must cover:
- Template file usage: `--template` / `-t`
- Any publish/activate flag

**What to flag:**
- Empty `Long` or `Example` fields
- Examples that use placeholder text like `<id>` without explaining what it is
- Outdated flag names in examples (e.g., `--non-interactive` instead of `--table`)

---

### 5. Error Formatting

**What to check:** Errors returned from `RunE` must be prefixed consistently.

**Expected pattern:**

```go
if err != nil {
    return fmt.Errorf("error: %s", err)
}
```

**What to flag:**
- `return err` without a prefix (loses context at the cobra level)
- `fmt.Errorf("error: %s", err.Error())` — `.Error()` call is redundant with `%s`
- Inconsistent prefix strings (e.g., `"failed: "`, `"could not: "`)

---

### 6. Web Flag (view commands only)

**What to check:** If a stable web URL exists for the resource, view commands should
support `--web` / `-W` to open it in the browser.

**Expected pattern:**

```go
flags.BoolVarP(&opts.Web, "web", "W", false, "Open the resource in the web application")
```

And inside `RunE`, checked **before** any API call:

```go
if opts.Web {
    return browser.OpenURL({resourceui}.Get{Resource}URL(opts.Args[0]))
}
```

Uses `github.com/pkg/browser`. If a URL helper doesn't exist yet, note it as a missing
piece but do not block the review on it.

---

## Reporting

After completing all checks, produce a summary:

```
## Review: prolific {resource} {command}

### 🔴 Must Fix
- [Check Type] [issue description] (`cmd/{resource}/{file}.go:{line}`)

### 🟡 Should Fix
- [Check Type] [issue description]

### 🟢 OK
- [Check Type], [Check Type], ...
```

**Example**
```
- 🔴 [Error Formatting]: Inconsistent error prefixes in `publishCollection` helper — `"failed to get collection: %s"`, `"failed to read template file: %s"`, etc. — all must use `"error: %s"`
```

If there are no issues, say so clearly: "All checks passed — no issues found."

After reporting, ask the user whether they want you to fix any of the flagged issues.

---

## Reference Files

| Pattern | Reference File |
|---------|----------------|
| Output flags + rendering | `cmd/study/list.go` |
| Output flags + rendering | `cmd/collection/list.go` |
| View with --web flag | `cmd/project/view.go` |
| Shared output helpers | `cmd/shared/list_format_flags.go` |
| Test pattern | `cmd/workspace/list_test.go` |

$ARGUMENTS
