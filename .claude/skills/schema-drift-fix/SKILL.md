---
name: schema-drift-fix
description: Check the CLI against the live Prolific API spec for drift (optionally starting from a filed schema-drift issue) and propose the client-code fix, following AGENTS.md's API-method checklist. Human reviews and commits.
argument-hint: "[issue-number]"
user-invocable: true
---

# Schema Drift Fix

## When to Use This Skill

Invoke this skill when the user:
- References a `schema-drift`-labeled issue
- Asks to "fix schema drift", "check for API drift", or similar
- Uses the slash command /schema-drift-fix

---

Check whether the CLI's `client` package still matches the live Prolific API spec, and propose the fix if it doesn't — following the same checklist a human would use when adding a new API method. This skill investigates and drafts; it does not commit, push, or open a PR. The engineer reviews and does that themselves.

## Arguments

`issue-number` (optional) — a `schema-drift` issue filed by `.github/workflows/schema-drift-check.yml`.

This skill does not depend on an issue existing. An issue is a convenient shortcut when CI already found and described drift, but the real source of truth is always a fresh `go test ./contract_test/...` run (Phase 1, step 1) — an issue's captured output can be stale by the time someone picks it up.

- If `issue-number` is given: `gh issue view <issue-number> --json title,body,state` — confirm it's an open issue with a title matching `API schema drift detected (...)`, and use its body as corroborating context.
- If omitted: skip the issue lookup entirely and just run the check directly. This is the default, everyday way to invoke this skill — proactively, not only after CI has already filed something.

## Phase 1: Investigate

1. Run `go test ./contract_test/...` locally (this downloads the live Prolific OpenAPI spec itself, same as the workflow). If it passes: report "no drift detected" and stop here — there's nothing to do, regardless of whether an issue was given. If it fails, continue.
2. If an issue number was given, cross-check its body against this fresh result — same operation(s) failing? Anything CI's snapshot called out that isn't reproducing now? Flag any mismatch rather than trusting either source silently.
3. Identify which operation(s)/test case(s) are failing from the fresh output.
4. Cross-reference `contract_test/contract_test.go`'s `operations` table entry for the affected operation(s) against the live spec to determine exactly what changed (new/renamed/removed field, changed type, new required param, etc.).

## Phase 2: Present Plan Summary

Before touching any code, present:
1. Which operation(s) drifted and how (a concrete old-shape vs. new-shape diff)
2. Which files will change
3. Whether this looks like a routine mechanical update or something structurally bigger — this automation is meant for small changes like a new/renamed field, not architectural redesigns. Flag it explicitly if it looks like the latter rather than plowing ahead.

**Ask for explicit approval before Phase 3.**

## Phase 3: Implement

Follow the "Adding a New Command" checklist in `AGENTS.md` (step 7, "If new API methods are needed") — the same checklist `schema-drift-check.yml`'s filed issue paraphrases:

1. Update the `API` interface in `client/client.go` and implement on the `Client` struct
2. Add/update request/response structs in `client/payloads.go`/`client/responses.go`
3. Run `make test-gen-mock`
4. Update the relevant entry in `contract_test/contract_test.go`'s `operations` table
5. Run `make readme-coverage`

Repeat per affected operation if more than one drifted.

## Phase 4: Verify

1. `go test ./contract_test/...` — confirm the specific drift is resolved
2. `make test`
3. `make lint`

Report results.

## Phase 5: Hand Back to the Human

This skill never commits, pushes, or opens a PR. Summarize what changed; if an issue number was given, point back to it so the engineer can reference it in their own commit message and close it themselves once merged. Every write action from here stays human-initiated.

## Reference Patterns

| Pattern | Reference File |
|---------|----------------|
| Schema drift detection workflow | `.github/workflows/schema-drift-check.yml` |
| Adding a New API Method checklist | `AGENTS.md` — "Adding a New Command", step 7 |
| Contract test operations table | `contract_test/contract_test.go` |
| Client method example | `client/client.go` |

$ARGUMENTS
