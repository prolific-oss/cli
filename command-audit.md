# Command Audit

Audit of all CLI commands in `cmd/` against project conventions. Covers 19 resources and ~80 action commands.

## Severity Legend

| Symbol | Meaning |
|--------|---------|
| 🔴 | **Must Fix** — non-conformant; will cause inconsistent UX or broken functionality |
| 🟡 | **Should Fix** — diverges from convention; not blocking |
| 🟢 | **OK** — conforms to pattern |
| N/A | Check not applicable to this command type |

## Checklist Reference

| # | Check | Applies To |
|---|-------|-----------|
| 1 | **Output Flags** — embeds `shared.OutputOptions`, calls `shared.AddOutputFlags`, uses `shared.ResolveFormat` | list commands |
| 2 | **Dependency Injection** — `client.API` + `io.Writer` injected; no `os.Stdout`; consistent `commandName` usage | all |
| 3 | **Rendering Strategy** — switch on `shared.ResolveFormat`; supports json/csv/table/interactive | list commands |
| 4 | **Help Text** — non-empty `Short` (imperative, no period), `Long`, and `Example` with flag coverage | all |
| 5 | **Error Formatting** — `fmt.Errorf("error: %s", err)` in RunE; no bare `return err`; no `.Error()` redundancy; no `"failed:"` prefix | all |
| 6 | **Web Flag** — `--web/-W` registered and checked before API call in RunE | view commands |

---

## Findings by Resource

### aitaskbuilder

> Subcommands: `batch create`, `batch export`, `batch instructions`, `batch setup`, `batch tasks`, `batch update`, `dataset create`, `dataset upload`, `batch list` (`get-batches`), `batch view` (`get-batch`), `batch status` (`get-batch-status`), `dataset status` (`get-dataset-status`), `task responses` (`get-task-responses`)

#### `prolific aitaskbuilder batch create` (`batch_create.go`)
- 🟢 Dependency Injection, Help Text
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` is redundant with `%s` (line 42)
- 🟡 **Error Formatting**: bare `return err` in helper (line 99)

#### `prolific aitaskbuilder batch export` (`batch_export.go`)
- 🟢 Dependency Injection, Help Text
- 🟡 **Error Formatting**: bare `return err` (line 171); internal helpers use `"failed to ..."` with `%w` — acceptable for deep error context but inconsistent with top-level pattern

#### `prolific aitaskbuilder batch instructions` (`batch_instructions.go`)
- 🟢 Dependency Injection, Help Text
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 57); bare `return err` in multiple internal branches (lines 125, 163, 167, 171, 225)

#### `prolific aitaskbuilder batch setup` (`batch_setup.go`)
- 🟢 Dependency Injection, Help Text
- 🟡 **Error Formatting**: bare `return err` (line 71)

#### `prolific aitaskbuilder batch tasks` (`batch_tasks.go`)
- 🟢 Dependency Injection, Help Text
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 36); bare `return err` (line 58)

#### `prolific aitaskbuilder batch update` (`batch_update.go`)
- 🟢 Dependency Injection, Help Text
- 🟡 **Error Formatting**: bare `return err` (lines 106, 133)

#### `prolific aitaskbuilder dataset create` (`create_dataset.go`)
- 🟢 Dependency Injection, Help Text
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 42); bare `return err` (line 77)

#### `prolific aitaskbuilder dataset upload` (`upload_dataset.go`)
- 🟢 Dependency Injection, Help Text
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 44); bare `return err` (lines 108, 114, 119, 125); internal helpers use `"failed to ..."` with `%w`

#### `prolific aitaskbuilder batch list` (`get_batches.go`)
- 🟡 **Output Flags**: No `shared.OutputOptions` — uses a simple `fmt.Println` loop. Not a full list command but inconsistent with other list patterns
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 60); bare `return err` (line 25)

#### `prolific aitaskbuilder batch view` (`get_batch.go`)
- 🟢 Dependency Injection, Help Text
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 35); bare `return err` (line 58)

#### `prolific aitaskbuilder batch status` (`get_batch_status.go`)
- 🟢 Dependency Injection, Help Text
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 37); bare `return err` (line 59)

#### `prolific aitaskbuilder dataset status` (`get_dataset_status.go`)
- 🟢 Dependency Injection, Help Text
- 🟡 **Error Formatting**: bare `return err` (line 66)

#### `prolific aitaskbuilder task responses` (`get_task_responses.go`)
- 🟢 Dependency Injection, Help Text
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 38); bare `return err` (line 60)

---

### bonus

#### `prolific bonus create` (`create.go`)
- 🔴 **Output Flags**: Uses custom `NonInteractive bool` and `Csv bool` fields instead of embedding `shared.OutputOptions`; does not call `shared.AddOutputFlags`
- 🔴 **Rendering Strategy**: Branches on `opts.Csv` / `opts.NonInteractive` booleans instead of `shared.ResolveFormat`; missing `json` and `table` cases
- 🟡 **Error Formatting**: bare `return err` in multiple branches (lines 96, 106, 165, 176)

#### `prolific bonus pay` (`pay.go`)
- 🟢 Dependency Injection, Help Text
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 34); bare `return err` (lines 50, 60)

---

### campaign

#### `prolific campaign list` (`list.go`)
- 🔴 **Output Flags**: `ListOptions` does not embed `shared.OutputOptions`; `shared.AddOutputFlags` not called
- 🔴 **Rendering Strategy**: Hardcoded `tabwriter` output (lines 68–88); no `shared.ResolveFormat` switch; `--json`, `--csv`, `--table` flags unavailable
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 52); bare `return err` (line 74)
- 🟡 **Help Text**: Examples missing `--table`, `--csv`, `--json` variants

---

### collection

#### `prolific collection list` (`list.go`) — canonical reference
- 🟢 Output Flags, Dependency Injection, Rendering Strategy, Help Text
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 83)

#### `prolific collection create` (`create_collection.go`)
- 🟢 Dependency Injection, Help Text
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 114); bare `return err` in helpers (lines 177, 182)

#### `prolific collection get` (`get.go`)
- 🟢 Dependency Injection, Help Text
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 50)

#### `prolific collection export` (`export.go`)
- 🟢 Dependency Injection, Help Text, Error Formatting
- Note: Internal download helpers use `"failed to ..."` with `%w` (lines 171–206) — acceptable for internal error chaining; not surfaced directly to users via RunE

#### `prolific collection update` (`update.go`)
- 🟢 Dependency Injection, Help Text
- 🟡 **Error Formatting**: bare `return err` in helper branches (lines 70, 75)

#### `prolific collection publish` (`publish.go`)
- 🟢 Dependency Injection, Help Text
- 🔴 **Error Formatting**: Six errors using wrong `"failed to ..."` prefix with `.Error()` redundancy — must use `"error: %s"` pattern:
  - line 113: `"failed to get collection: %s"`
  - line 123: `"failed to read template file: %s"`
  - line 127: `"failed to parse template file: %s"`
  - line 187: `"failed to create study: %s"`
  - line 202: `"failed to publish study: %s"`
  - line 208: `"failed to get study details: %s"`

#### `prolific collection preview` (`preview.go`)
- 🟡 **Help Text**: Missing `Short` field (line 41)
- 🔴 **Error Formatting**: `fmt.Errorf("failed to get collection: %s", err.Error())` — wrong prefix and `.Error()` redundant (line 69)

---

### credentials

#### `prolific credentials list` (`list.go`)
- 🔴 **Output Flags**: `ListOptions` does not embed `shared.OutputOptions`; `shared.AddOutputFlags` not called
- 🔴 **Rendering Strategy**: Hardcoded `fmt.Fprintf` table output (lines 36–54); no format resolution; `--json`, `--csv`, `--table` flags unavailable
- 🟡 **Dependency Injection**: Constructor missing `commandName` parameter (line 17) — inconsistent with package pattern

#### `prolific credentials create` (`create.go`)
- 🟢 Dependency Injection, Help Text
- 🟡 **Error Formatting**: bare `return err` in helper branches (lines 49, 54)

#### `prolific credentials update` (`update.go`)
- 🟢 Dependency Injection, Help Text
- 🟡 **Error Formatting**: bare `return err` in helper branches (lines 50, 55)

---

### filters

#### `prolific filters list` (`list.go`)
- 🟡 **Output Flags**: Uses raw `nonInteractive bool` flag instead of `shared.OutputOptions`; does not call `shared.AddOutputFlags` — missing `--json`, `--csv`, `--table` flags
- 🟡 **Dependency Injection**: Constructor missing `commandName` parameter (line 13)
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 46); bare `return err` (lines 62, 75)
- 🟡 **Help Text**: Examples show `-n` but not `--table`, `--csv`, `--json`

---

### filtersets

#### `prolific filtersets list` (`list.go`)
- 🔴 **Output Flags**: Does not embed `shared.OutputOptions`; `shared.AddOutputFlags` not called
- 🔴 **Rendering Strategy**: Hardcoded `tabwriter` output (lines 75–86); no `shared.ResolveFormat` switch
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 44); bare `return err` (line 67)

#### `prolific filtersets view` (`view.go`)
- 🟢 Dependency Injection, Help Text, Web Flag (registered at line 52, checked before API call)
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 43); bare `return err` (line 65)

#### `prolific filtersets create` (`create.go`)
- 🟢 Dependency Injection, Help Text
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 54); bare `return err` (lines 74, 99)

---

### hook

#### `prolific hook list` (`list.go`)
- 🔴 **Output Flags**: Does not embed `shared.OutputOptions`; `shared.AddOutputFlags` not called
- 🔴 **Rendering Strategy**: Hardcoded `tabwriter` output (lines 88–96); no `shared.ResolveFormat` switch
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 53); bare `return err` (line 80)

#### `prolific hook create` (`create.go`)
- 🟢 Help Text
- 🟡 **Dependency Injection**: Constructor missing `commandName` parameter (line 20) — inconsistent with sibling commands
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (lines 45, 50)

#### `prolific hook delete` (`delete.go`)
- 🟢 Help Text
- 🟡 **Dependency Injection**: Constructor missing `commandName` parameter (line 12)
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 28)

#### `prolific hook update` (`update.go`)
- 🟢 Help Text
- 🟡 **Dependency Injection**: Constructor missing `commandName` parameter (line 21)
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 65)

#### `prolific hook events` (`event_list.go`)
- 🔴 **Output Flags**: Does not embed `shared.OutputOptions`; `shared.AddOutputFlags` not called
- 🔴 **Rendering Strategy**: Hardcoded `tabwriter` output (lines 78–88); no `shared.ResolveFormat` switch
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 52); bare `return err` (line 75)

#### `prolific hook event-types` (`event_type.go`)
- 🔴 **Output Flags**: Does not embed `shared.OutputOptions`; `shared.AddOutputFlags` not called
- 🔴 **Rendering Strategy**: Hardcoded `tabwriter` output (lines 43–50); no `shared.ResolveFormat` switch
- 🟡 **Help Text**: Missing `Long` description (lines 18–22)
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 26); bare `return err` (line 40)

#### `prolific hook secrets` (`secret.go`)
- 🟡 **Dependency Injection**: `NewListSecretCommand` and `NewCreateSecretCommand` missing `commandName` parameter (lines 23, 55)
- 🟡 **Help Text**: `Short` text is minimal for list; `Long` missing for the list sub-command
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (lines 33, 86); bare `return err` (lines 74, 132)

---

### invitation

#### `prolific invitation create` (`create.go`)
- 🟢 Dependency Injection, Help Text
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 48); bare `return err` propagated from helper (line 89)

---

### message

#### `prolific message list` (`list.go`)
- 🔴 **Output Flags**: Does not embed `shared.OutputOptions`; `shared.AddOutputFlags` not called
- 🔴 **Rendering Strategy**: Hardcoded `tabwriter` output (lines 70–126); no `shared.ResolveFormat` switch
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 53); bare `return err` (lines 92, 98)

#### `prolific message send` (`send.go`)
- 🟢 Dependency Injection, Help Text
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 44); bare `return err` (line 75)

#### `prolific message bulk-send` (`bulk_send.go`)
- 🟢 Dependency Injection, Help Text
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 38); bare `return err` (line 69)

#### `prolific message send-group` (`send_group.go`)
- 🟢 Dependency Injection, Help Text
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 38); bare `return err` (line 69)

---

### participantgroup

#### `prolific participantgroup list` (`list.go`)
- 🔴 **Output Flags**: Does not embed `shared.OutputOptions`; `shared.AddOutputFlags` not called
- 🔴 **Rendering Strategy**: Hardcoded `tabwriter` output (lines 60–86); no `shared.ResolveFormat` switch
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 44); bare `return err` (line 67)

#### `prolific participantgroup view` (`view.go`)
- 🔴 **Dependency Injection / Struct**: Uses `ListOptions` struct instead of a dedicated `ViewOptions` — apparent copy-paste error (line 21)
- 🟡 **Web Flag**: No `--web/-W` flag implemented
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 41); bare `return err` (line 59)

#### `prolific participantgroup create` (`create.go`)
- 🟢 Dependency Injection, Help Text
- 🟡 **Error Formatting**: bare `return err` (line 85)

#### `prolific participantgroup remove` (`remove.go`)
- 🟢 Dependency Injection, Help Text, Error Formatting
- 🟡 **Error Formatting**: bare `return err` in helper (lines 73, 84)

---

### project

#### `prolific project list` (`list.go`)
- 🔴 **Output Flags**: `ListOptions` does not embed `shared.OutputOptions`; `shared.AddOutputFlags` not called
- 🔴 **Rendering Strategy**: Hardcoded `tabwriter` output (lines 81–87); no `shared.ResolveFormat` switch
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 50); bare `return err` (line 73)
- 🟡 **Help Text**: Examples missing `--table`, `--csv`, `--json` variants

#### `prolific project create` (`create.go`)
- 🟢 Dependency Injection
- 🟡 **Help Text**: Example missing `--template` usage
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (lines 42, 74); bare `return err` (line 85)

#### `prolific project view` (`project/view.go`) — canonical web-flag reference
- 🟢 Dependency Injection, Web Flag (registered correctly, checked before API call)
- 🟡 **Help Text**: Example does not show `--web/-W` usage
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 47); bare `return err` (line 69)

---

### researcher

#### `prolific researcher create-participant` (`create_participant.go`)
- 🟢 Help Text
- 🔴 **Error Formatting**: bare `return err` directly in RunE (line 34) — error from `client.CreateTestParticipant()` is not wrapped with `"error: %s"` prefix

---

### study

#### `prolific study list` (`list.go`) — canonical reference
- 🟢 Output Flags, Dependency Injection, Rendering Strategy, Help Text, Error Formatting

#### `prolific study view` (`view.go`)
- 🟢 Dependency Injection, Help Text, Web Flag
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 45)

#### `prolific study create` (`create.go`)
- 🟢 Dependency Injection, Help Text
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 113); bare `return err` in helpers (lines 133, 144, 150, 155)

#### `prolific study update` (`update.go`)
- 🟢 Dependency Injection, Help Text
- 🔴 **Error Formatting**: bare `return err` directly in RunE (line 103) — no prefix
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (lines 82, 87, 93)

#### `prolific study credentials-report` (`credentials_report.go`)
- 🟢 Help Text
- 🔴 **Error Formatting**: bare `return err` directly in RunE (line 34) — no prefix

#### `prolific study demographic-export` (`demographic_export.go`)
- 🟢 Help Text
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 28)

#### `prolific study duplicate` (`duplicate.go`)
- 🟢 Help Text
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 29)

#### `prolific study increase-places` (`increase_places.go`)
- 🟡 **Help Text**: `Short` field is empty (line 24)
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 41); bare `return err` (line 50)

#### `prolific study set-credential-pool` (`set_credential_pool.go`)
- 🟡 **Help Text**: `Short` field is empty (line 24)
- 🟡 **Error Formatting**: bare `return err` (line 44)

#### `prolific study submission-counts` (`submission_counts.go`)
- 🔴 **Output Flags**: Uses `NonInteractive bool` instead of embedding `shared.OutputOptions`; `shared.AddOutputFlags` not called (line 21, 87)
- 🔴 **Rendering Strategy**: Hardcoded `tabwriter` table output; no `shared.ResolveFormat` switch; `--csv`, `--table`, `--json` flags unavailable through standard interface
- 🟡 **Help Text**: Examples show `-n` and `--json` but not `--table`, `--csv`
- 🟡 **Error Formatting**: bare `return err` (lines 48, 54)

#### `prolific study test` (`test_study.go`)
- 🟢 Help Text
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 33)

#### `prolific study transition` (`transition.go`)
- 🟢 Dependency Injection, Help Text
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 37); bare `return err` (lines 58, 64)

---

### submission

#### `prolific submission list` (`list.go`) — fully conformant
- 🟢 Output Flags, Dependency Injection, Rendering Strategy, Help Text, Error Formatting

#### `prolific submission bulk-approve` (`bulk_approve.go`)
- 🟢 Dependency Injection, Help Text
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 49); bare `return err` in helpers (lines 77, 111)

#### `prolific submission request-return` (`request_return.go`)
- 🟢 Dependency Injection, Help Text
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 43); bare `return err` (line 60)

#### `prolific submission transition` (`transition.go`)
- 🟢 Dependency Injection, Help Text
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 77); bare `return err` (line 136)

---

### survey

#### `prolific survey list` (`list.go`)
- 🟢 Output Flags, Dependency Injection, Rendering Strategy, Help Text
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (lines 53, 58, 66, 71, 76)

#### `prolific survey view` (`view.go`)
- 🟢 Dependency Injection, Help Text
- 🟡 **Web Flag**: Not implemented — consider adding if a stable web URL exists for surveys
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 39); bare `return err` (line 57)

#### `prolific survey create` (`create.go`)
- 🟢 Dependency Injection, Help Text
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 49); bare `return err` (lines 68, 84, 91)

#### `prolific survey delete` (`delete.go`)
- 🟢 Dependency Injection, Help Text
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 38); bare `return err` (line 51)

#### `prolific survey response list` (`response_list.go`)
- 🟢 Output Flags, Dependency Injection, Rendering Strategy, Help Text
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (lines 55, 63, 68, 73, 78)

#### `prolific survey response view` (`response_view.go`)
- 🟢 Dependency Injection, Help Text
- 🟡 **Web Flag**: Not implemented
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 39); bare `return err` (line 56)

#### `prolific survey response create` (`response_create.go`)
- 🟢 Dependency Injection, Help Text
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 44); bare `return err` (lines 62, 75)

#### `prolific survey response delete` (`response_delete.go`)
- 🟢 Dependency Injection, Help Text
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 38); bare `return err` (line 51)

#### `prolific survey response delete-all` (`response_delete_all.go`)
- 🟢 Dependency Injection, Help Text
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 38); bare `return err` (line 53)

#### `prolific survey response summary` (`response_summary.go`)
- 🟢 Dependency Injection, Help Text
- 🟡 **Output Flags**: Registers only a standalone `--json` flag instead of using `shared.OutputOptions` and `shared.AddOutputFlags` — missing `--csv`, `--table` consistency
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 46); bare `return err` (line 63)

---

### template

#### `prolific template list` (`list.go`)
- 🟡 **Output Flags**: No `shared.OutputOptions` embed; `shared.AddOutputFlags` not called (lines 89–112)
- 🟡 **Rendering Strategy**: Hardcoded `tabwriter` output (line 100); no format switch
- 🟡 **Help Text**: Example only shows bare `prolific template list`; missing format flag variants

#### `prolific template view` (`view.go`)
- 🟢 Dependency Injection, Help Text, Error Formatting

---

### user

#### `prolific user whoami` (`me.go`)
- 🟡 **Help Text**: `Short` field missing (line 17)
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 21); bare `return err` (line 35)

---

### workspace

#### `prolific workspace list` (`list.go`)
- 🔴 **Output Flags**: `WorkspaceListOptions` does not embed `shared.OutputOptions`; `shared.AddOutputFlags` not called
- 🔴 **Rendering Strategy**: Hardcoded `tabwriter` output (lines 69–75); no `shared.ResolveFormat` switch; `--json`, `--csv`, `--table` flags unavailable
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 48); bare `return err` (line 66)
- 🟡 **Help Text**: Examples missing `--table`, `--csv`, `--json` variants

#### `prolific workspace create` (`create.go`)
- 🟢 Dependency Injection, Help Text
- 🟡 **Error Formatting**: `fmt.Errorf("error: %s", err.Error())` — `.Error()` redundant (line 41); bare `return err` (line 66)

#### `prolific workspace balance` (`balance.go`)
- 🟢 Dependency Injection, Help Text, Error Formatting

---

## Summary

### 🔴 Must Fix — 16 issues across 13 commands

| Command | File | Issue |
|---------|------|-------|
| `bonus create` | `bonus/create.go` | Custom `NonInteractive`/`Csv` flags instead of `shared.OutputOptions`; missing `shared.ResolveFormat` |
| `campaign list` | `campaign/list.go` | Missing `shared.OutputOptions`; hardcoded `tabwriter` |
| `collection publish` | `collection/publish.go` | Wrong `"failed to ..."` prefix on 6 errors (lines 113, 123, 127, 187, 202, 208) |
| `collection preview` | `collection/preview.go` | Wrong `"failed to get collection: %s"` prefix (line 69) |
| `credentials list` | `credentials/list.go` | Missing `shared.OutputOptions`; hardcoded `fmt.Fprintf` table |
| `filtersets list` | `filtersets/list.go` | Missing `shared.OutputOptions`; hardcoded `tabwriter` |
| `hook list` | `hook/list.go` | Missing `shared.OutputOptions`; hardcoded `tabwriter` |
| `hook events` | `hook/event_list.go` | Missing `shared.OutputOptions`; hardcoded `tabwriter` |
| `hook event-types` | `hook/event_type.go` | Missing `shared.OutputOptions`; hardcoded `tabwriter` |
| `message list` | `message/list.go` | Missing `shared.OutputOptions`; hardcoded `tabwriter` |
| `participantgroup list` | `participantgroup/list.go` | Missing `shared.OutputOptions`; hardcoded `tabwriter` |
| `participantgroup view` | `participantgroup/view.go` | Uses `ListOptions` struct instead of `ViewOptions` (copy-paste error) |
| `project list` | `project/list.go` | Missing `shared.OutputOptions`; hardcoded `tabwriter` |
| `researcher create-participant` | `researcher/create_participant.go` | Bare `return err` in RunE (line 34) |
| `study credentials-report` | `study/credentials_report.go` | Bare `return err` in RunE (line 34) |
| `study submission-counts` | `study/submission_counts.go` | `NonInteractive bool` instead of `shared.OutputOptions`; hardcoded `tabwriter` |
| `study update` | `study/update.go` | Bare `return err` in RunE (line 103) |
| `workspace list` | `workspace/list.go` | Missing `shared.OutputOptions`; hardcoded `tabwriter` |

### 🟡 Should Fix — Cross-cutting patterns

**`.Error()` redundancy** (present in nearly every file):  
`fmt.Errorf("error: %s", err.Error())` — the `.Error()` call is redundant with `%s`; use `fmt.Errorf("error: %s", err)`.  
Affects: all resources except `submission/list.go`, `workspace/balance.go`, `participantgroup/remove.go`.

**Bare `return err` in helper functions** (widespread):  
Many helper functions called from RunE return bare `err` without wrapping. While some propagation is acceptable, consider consistent wrapping at the helper level for clearer error messages.

**Missing `commandName` parameter** in constructors (hook, credentials, filters packages):  
`hook/create.go`, `hook/delete.go`, `hook/update.go`, `hook/secret.go`, `credentials/list.go`, `filters/list.go` use `NewXxxCommand(client, w)` without the leading `commandName string` argument — inconsistent with the rest of the codebase.

**Missing `Short` help text**:  
`study/increase_places.go`, `study/set_credential_pool.go`, `collection/preview.go`, `user/me.go`

**Missing output flag examples in help**:  
`workspace/list.go`, `project/list.go`, `campaign/list.go`, `filters/list.go`, `template/list.go` — Examples do not show `--table`, `--csv`, `--json` variants.

**`--web` flag not implemented on view commands**:  
`survey/view.go`, `survey/response_view.go`, `participantgroup/view.go` — consider adding if stable web URLs exist.

### 🟢 Fully Conformant Commands

- `study list` (canonical list reference)
- `submission list` (canonical list reference)
- `collection export`
- `collection update`
- `workspace balance`
- `template view`
- `participantgroup remove`
- `project view` (canonical view-with-web reference)
- All parent/grouping commands (`study.go`, `collection.go`, `workspace.go`, etc.)
