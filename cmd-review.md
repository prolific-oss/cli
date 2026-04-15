# CLI Command Review Report

> Reviewed ~50 command files across 20 resources against the AGENTS.md checklist.
> Reference implementations: `study/list.go`, `collection/list.go`, `project/view.go`.

---

## campaign

### list
- 🔴 No `shared.OutputOptions`/`AddOutputFlags`/`ResolveFormat` — hardcoded `tabwriter` only, no interactive TUI, JSON, CSV, or table flags
- 🟡 Redundant `.Error()` in `fmt.Errorf("error: %s", err.Error())`
- 🟢 DI correct; help text present

---

## collection

### list *(reference)*
- 🟡 Redundant `.Error()` in `fmt.Errorf`
- 🟢 Full output strategy correct

### get *(view command)*
- 🟡 Redundant `.Error()` in `fmt.Errorf`
- 🟡 No `--web`/`-W` flag (if a stable web URL exists)
- 🟢 DI correct; help text present

### create
- 🟡 Redundant `.Error()` in `fmt.Errorf`
- 🟢 DI correct; help text present

### update
- 🔴 Bare `return err` ×2 in `RunE` (errors from `validateTemplate` and `UpdateCollection` unwrapped)
- 🟢 DI correct; help text present

### publish
- 🔴 Inconsistent error prefixes in `publishCollection` helper — `"failed to get collection: %s"`, `"failed to read template file: %s"`, etc. — all must use `"error: %s"`
- 🟡 Redundant `.Error()` throughout helper
- 🟢 DI correct; help text present

### export
- 🔴 Mixed error prefixes in `exportCollection` helper — mix of `"error requesting export: %s"`, `"failed to create download request: %w"`, etc.
- 🟡 Redundant `.Error()` in several calls
- 🟢 DI correct; help text present

### preview
- 🟢 All OK

---

## credentials

### list
- 🔴 No `shared.OutputOptions`/`AddOutputFlags`/`ResolveFormat` — plain `fmt.Fprintf` only
- 🔴 Bare `return err` in `RunE` (`ListCredentialPools` error unwrapped)
- 🟢 DI correct; help text present

### create
- 🔴 Bare `return err` ×2 in `RunE`
- 🟢 DI correct; help text present

### update
- 🔴 Bare `return err` ×2 in `RunE`
- 🟢 DI correct; help text present

---

## filters

### list
- 🔴 Uses old `nonInteractive bool` flag pattern — missing `--table`/`-t`, `--csv`/`-c`, `--json`/`-j`; no `AddOutputFlags`/`ResolveFormat`
- 🟡 Redundant `.Error()` in `fmt.Errorf`
- 🟢 DI correct; help text present

---

## filtersets

### list
- 🔴 No `shared.OutputOptions`/`AddOutputFlags`/`ResolveFormat` — hardcoded `tabwriter` only
- 🟡 Redundant `.Error()` in `fmt.Errorf`
- 🟢 DI correct; help text present

### view
- 🔴 `opts.Web` is checked **inside `renderProject` after `client.GetFilterSet()` is called** — must be checked before any API call in `RunE`
- 🟡 Redundant `.Error()` in `fmt.Errorf`
- 🟡 `Example` block missing a `--web`/`-W` usage example
- 🟢 `--web`/`-W` flag registered; DI correct; help text present

### create
- 🟡 Redundant `.Error()` in `fmt.Errorf`
- 🟢 DI correct; help text present

---

## hook

### list
- 🔴 No `shared.OutputOptions`/`AddOutputFlags`/`ResolveFormat` — hardcoded `tabwriter` only
- 🟡 Redundant `.Error()` in `fmt.Errorf`
- 🟢 DI correct; help text present

### event_list
- 🔴 No `shared.OutputOptions`/`AddOutputFlags`/`ResolveFormat` — hardcoded `tabwriter` only
- 🟡 Redundant `.Error()` in `fmt.Errorf`
- 🟢 DI correct; help text present

### create / delete / update
- 🟡 Redundant `.Error()` in `fmt.Errorf` (all three)
- 🟢 DI correct; help text present

### secret (`NewListSecretCommand`)
- 🔴 Missing `Long` and `Example` fields — only `Short` is set
- 🟡 Redundant `.Error()` in `NewCreateSecretCommand`
- 🟢 DI correct

---

## invitation

### create
- 🟡 Redundant `.Error()` in `fmt.Errorf`
- 🟢 DI correct; help text present

---

## message

### list
- 🔴 No `shared.OutputOptions`/`AddOutputFlags`/`ResolveFormat` — hardcoded `tabwriter` only
- 🟡 Redundant `.Error()` in `fmt.Errorf`
- 🟢 DI correct; help text present

### send / bulk_send / send_group
- 🟡 Redundant `.Error()` in `fmt.Errorf` (all three)
- 🟢 DI correct; help text present

---

## participantgroup

### list
- 🔴 No `shared.OutputOptions`/`AddOutputFlags`/`ResolveFormat` — hardcoded `tabwriter` only
- 🟡 Redundant `.Error()` in `fmt.Errorf`
- 🟢 DI correct; help text present

### view
- 🔴 `var opts ListOptions` declared instead of `var opts ViewOptions` — `ViewOptions` struct exists but is never used; `ListOptions` (with irrelevant pagination fields) is used instead
- 🟡 Redundant `.Error()` in `fmt.Errorf`
- 🟡 No `--web`/`-W` flag considered
- 🟢 DI correct; help text present

### create
- 🟡 Redundant `.Error()` in `fmt.Errorf`
- 🟢 DI correct; help text present

### remove
- 🟢 All OK

---

## project

### list
- 🔴 No `shared.OutputOptions`/`AddOutputFlags`/`ResolveFormat` — hardcoded `tabwriter` only
- 🟡 Redundant `.Error()` in `fmt.Errorf`
- 🟢 DI correct; help text present

### view *(reference)*
- 🟡 Redundant `.Error()` in `fmt.Errorf`
- 🟡 `Example` block missing a `--web`/`-W` usage example (even though the flag is implemented)
- 🟢 `--web` flag registered and checked before API call; DI correct

### create
- 🟡 Inconsistent error prefix `"unable to get current user: %s"` — must use `"error: %s"`
- 🟡 Redundant `.Error()` in `fmt.Errorf`
- 🟢 DI correct; help text present

---

## researcher

### create_participant
- 🔴 Bare `return err` in `RunE` — `CreateTestParticipant` error unwrapped
- 🟢 DI correct; help text present

---

## study

### list *(reference)*
- 🟡 Bare `return err` at the `GetStudies` API call (~line 105) — renderer errors below are correctly wrapped but this one is not
- 🟢 Full output strategy correct; DI correct

### view
- 🟡 Redundant `.Error()` in `fmt.Errorf`
- 🟡 `Long` is too terse (`"View study details"`) — should be a descriptive paragraph
- 🟡 `Example` missing a `--web`/`-W` usage example
- 🟢 `--web` flag registered and checked before API call; DI correct

### create
- 🔴 `Short: "Creation of studies"` — not imperative mood; should be `"Create a study"`
- 🟡 Redundant `.Error()` in `fmt.Errorf`
- 🟢 DI correct; help text present

### update
- 🔴 Bare `return err` in `updateStudy` helper for the final `UpdateStudy` API call
- 🟡 Redundant `.Error()` throughout helper
- 🟢 DI correct; help text present

### transition
- 🔴 `Example` is completely empty (`Example: \`\``)
- 🔴 Bare `return err` ×2 in `transitionStudy` helper — API errors unwrapped
- 🟡 Redundant `.Error()` in `RunE`
- 🟢 DI correct; `Short`/`Long` present

### duplicate
- 🟡 Redundant `.Error()` in `fmt.Errorf`
- 🟢 DI correct; help text present

### increase_places
- 🔴 Bare `return err` in `RunE` for `UpdateStudy` call
- 🟡 Redundant `.Error()` in `fmt.Errorf`
- 🟢 DI correct; help text present

### demographic_export
- 🟡 Redundant `.Error()` in `fmt.Errorf`
- 🟢 DI correct; help text present

### credentials_report
- 🔴 Bare `return err` in `RunE` — `GetStudyCredentialsUsageReportCSV` error unwrapped
- 🟢 DI correct; help text present

### submission_counts
- 🔴 Own `NonInteractive bool` and `JSON bool` flags instead of `shared.OutputOptions` — no CSV mode, no `AddOutputFlags`, no `ResolveFormat`
- 🔴 Bare `return err` ×2 in `RunE`
- 🟢 DI correct; help text present

### test_study
- 🟡 Redundant `.Error()` in `fmt.Errorf`
- 🟢 DI correct; help text present

### set_credential_pool
- 🔴 Bare `return err` in `RunE` — `UpdateStudy` error unwrapped
- 🟢 DI correct; help text present

---

## submission

### list
- 🟡 Bare `return err` at `GetSubmissions` API call before format switch (~line 91) — renderer errors below are correctly wrapped
- 🟢 Full output strategy correct; DI correct; help text present

### bulk_approve
- 🔴 Bare `return err` ×2 in `bulkApprove` helper (file-read and API errors)
- 🟡 Redundant `.Error()` in `fmt.Errorf`
- 🟢 DI correct; help text present

### request_return
- 🔴 Bare `return err` in `requestReturn` helper
- 🟡 Redundant `.Error()` in `fmt.Errorf`
- 🟢 DI correct; help text present

### transition
- 🔴 Bare `return err` in `transitionSubmission` helper
- 🟡 Redundant `.Error()` in `fmt.Errorf`
- 🟢 DI correct; help text present

---

## survey

### list
- 🔴 Missing `case "csv":` in the `ResolveFormat` switch — `--csv`/`-c` is registered but silently falls through to interactive renderer
- 🟡 `Example` uses `-n` (hidden alias) instead of canonical `--table`/`-t`
- 🟡 Redundant `.Error()` ×2
- 🟢 `OutputOptions`/`AddOutputFlags`/`ResolveFormat` present; DI correct

### view
- 🟡 No `--web`/`-W` flag considered
- 🟡 Redundant `.Error()` in `fmt.Errorf`
- 🟢 DI correct; help text present

### create / delete / response_view / response_create / response_delete / response_delete_all
- 🟡 Redundant `.Error()` in `fmt.Errorf` (all)
- 🟢 DI correct; help text present

### response_list
- 🟡 `Example` uses `-n` instead of canonical `--table`/`-t`
- 🟡 Redundant `.Error()` in `fmt.Errorf`
- 🟢 All four output modes present; DI correct; help text present

### response_summary
- 🟡 Defines own `--json`/`-j` bool flag instead of using `shared.OutputOptions` — will diverge further if more formats are added
- 🟡 Redundant `.Error()` in `fmt.Errorf`
- 🟢 DI correct; help text present

---

## template

### list / view
- ℹ️ Templates are embedded in the binary — no API client needed; output flags not applicable
- 🟢 All OK

---

## workspace

### list
- 🔴 No `shared.OutputOptions`/`AddOutputFlags`/`ResolveFormat` — hardcoded `tabwriter` only
- 🟡 Redundant `.Error()` in `fmt.Errorf`
- 🟢 DI correct; help text present

### balance
- 🟡 `errors.New("error: please provide a workspace ID")` bakes the `"error: "` prefix into the string — should be `errors.New("please provide a workspace ID")` or `fmt.Errorf("error: ...")`
- 🟢 DI correct; help text present

### create
- 🟡 Redundant `.Error()` in `fmt.Errorf`
- 🟢 DI correct; help text present

---

## bonus

### create
- 🔴 Own `NonInteractive bool`/`Csv bool` flags instead of `shared.OutputOptions` — bypasses shared rendering strategy entirely
- 🟡 Redundant `.Error()` in `fmt.Errorf`
- 🟢 DI correct; help text present

### pay
- 🟡 Redundant `.Error()` in `fmt.Errorf`
- 🟢 DI correct; help text present

---

## Summary

### 🔴 Must Fix (31 issues across 20 files)

| File | Issue |
|---|---|
| `campaign/list.go` | No output strategy (`OutputOptions`/`AddOutputFlags`/`ResolveFormat`) |
| `credentials/list.go` | No output strategy; bare `return err` |
| `credentials/create.go` | Bare `return err` ×2 |
| `credentials/update.go` | Bare `return err` ×2 |
| `filters/list.go` | Old `nonInteractive bool` pattern; missing `-t`/`-c`/`-j` flags |
| `filtersets/list.go` | No output strategy |
| `filtersets/view.go` | `opts.Web` checked after API call, not before |
| `hook/list.go` | No output strategy |
| `hook/event_list.go` | No output strategy |
| `hook/secret.go` | `NewListSecretCommand` missing `Long` and `Example` |
| `message/list.go` | No output strategy |
| `participantgroup/list.go` | No output strategy |
| `participantgroup/view.go` | Uses `ListOptions` struct instead of `ViewOptions` |
| `project/list.go` | No output strategy |
| `researcher/create_participant.go` | Bare `return err` |
| `study/create.go` | `Short` not imperative (`"Creation of studies"`) |
| `study/transition.go` | Empty `Example`; bare `return err` ×2 in helper |
| `study/update.go` | Bare `return err` in helper |
| `study/increase_places.go` | Bare `return err` in `RunE` |
| `study/credentials_report.go` | Bare `return err` |
| `study/submission_counts.go` | Own bool flags instead of `shared.OutputOptions`; bare `return err` ×2 |
| `study/set_credential_pool.go` | Bare `return err` |
| `collection/update.go` | Bare `return err` ×2 |
| `collection/publish.go` | Inconsistent error prefixes (`"failed to ..."`) |
| `collection/export.go` | Mixed/inconsistent error prefixes |
| `submission/bulk_approve.go` | Bare `return err` ×2 |
| `submission/request_return.go` | Bare `return err` |
| `submission/transition.go` | Bare `return err` |
| `survey/list.go` | Missing `case "csv":` in switch despite `--csv` being registered |
| `workspace/list.go` | No output strategy |
| `bonus/create.go` | Own `NonInteractive bool`/`Csv bool` instead of `shared.OutputOptions` |

### 🟡 Should Fix (widespread)

| Pattern | Files Affected |
|---|---|
| `fmt.Errorf("error: %s", err.Error())` — drop `.Error()` | Nearly every file |
| `Example` uses `-n` alias instead of canonical `--table`/`-t` | `survey/list.go`, `survey/response_list.go` |
| Missing `--web`/`-W` example in `Example` block | `study/view.go`, `project/view.go` |
| No `--web`/`-W` flag | `participantgroup/view.go`, `survey/view.go`, `collection/get.go` |
| `errors.New("error: ...")` bakes prefix into string | `workspace/balance.go` |
| Own `--json` flag instead of `shared.OutputOptions` | `survey/response_summary.go` |
| Inconsistent prefix `"unable to get current user:"` | `project/create.go` |
| Bare `return err` before format switch (API call) | `study/list.go`, `submission/list.go` |
| `Long` is too terse | `study/view.go` |
