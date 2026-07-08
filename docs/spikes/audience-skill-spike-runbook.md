# Audience Skill Spike Runbook

## Purpose

Run a one-day spike to test whether natural-language audience requests can be translated into valid count-preview payloads with iterative count feedback.

## Where The Skill Lives

- Skill spec: .claude/skills/audience-skill-spike/SKILL.md

## Scenario Workflow

1. Collect audience request in plain language.
2. Resolve request into concrete constraints.
3. Map constraints to live filter IDs from `./prolific filters -n`.
4. Generate count-preview template JSON.
5. Execute eligibility count preview command.
6. Capture eligible count and refinement options.
7. Optionally persist as a filter set if the audience is accepted.

## Commands

Run from repository root.

1. Validate auth and connectivity:
   - `./prolific whoami`
2. Inspect filter catalog:
   - `./prolific filters -n`
3. Preview eligible participant count from natural-language-derived template:
   - `./prolific filters count -t <template.json> -w <workspace_id>`
4. Optional persistence step:
   - `./prolific filter-sets create -t <template.json> -w <workspace_id>`

## Scenario Log Template

- Audience request:
- Parsed constraints:
- Mapped filters:
- Assumptions:
- Template path:
- Count command:
- Eligible count:
- Refinement decision:
- Outcome:

## Starter Scenario

Audience request:

- Participants with approval rate between 95 and 100, active in the last 30 days

Expected mapping:

- `approval_rate` -> `selected_range: {"lower": 95, "upper": 100}`
- `active_in_last_days` -> `selected_range: {"upper": 30}`

## Exit Criteria

1. At least 6 scenarios run.
2. At least 4 successful end-to-end loops.
3. Top 3 blockers documented.
4. Recommendation written: continue skill-first, add CLI wrapper, or request API additions.
