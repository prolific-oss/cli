---
name: audience-skill-spike
description: Convert natural-language audience requests into count-preview filter payloads using existing Prolific CLI primitives, then refine via eligible-count feedback.
argument-hint: "[audience request]"
user-invocable: true
---

# Audience Skill Spike

## When to Use This Skill

Invoke this skill when the user asks to:

- Build an audience from plain language
- Turn audience intent into filters
- Iterate audience constraints based on eligible count
- Preview audience size before deciding whether to persist any artifact

---

Build and validate a thin external Claude workflow over the existing CLI.
Do not add model runtime into the CLI binary during this spike.

## Inputs

Primary input may be provided as `$ARGUMENTS` or asked interactively:

- Audience request in natural language

If `$ARGUMENTS` is empty, start with exactly:

What audience are you looking for?
Share it in plain language.

Do not present starter-scenario menus or option pickers.

Required environment assumptions:

- CLI executable available as `./prolific` in repository root
- `PROLIFIC_TOKEN` configured for target environment
- `workspace_id` available via `-w` or config file

## Output Contract

For each run, return:

1. Parsed constraints from user request
2. Mapping from constraints to concrete filter IDs and values/ranges
3. Assumptions and ambiguities
4. Generated count-template JSON
5. Eligible participant count result
6. Refinement options (broaden, narrow, swap, remove)
7. If user confirms they are happy with this audience, ask:
   1. Would you like to create a filter set from these filters?
   2. Would you like to create a draft study with these filters?

## Conversation Contract

- Keep responses short and conversational.
- Start by understanding the audience in human terms, then map to filters.
- End each turn with one clear next question when a decision is needed.
- After count is returned and user is happy, present the two post-acceptance questions as a numbered list in order.
- Avoid menu-style option pickers unless the user asks for options explicitly.

## Audience Language Contract

- Describe audiences as people and criteria, not implementation details.
- Do not expose raw filter IDs in user-facing prose unless the user asks.
- Keep JSON and command syntax in dedicated code blocks only.

## Phase 1: Resolve Intent

1. Parse the audience request into atomic constraints.
2. Label each constraint as one of:
   - select
   - range
   - participant group
   - unsupported/unknown
3. If confidence is low for any constraint, ask a clarifying question before generating payload.

## Ambiguity Policy

When a request has multiple plausible readings:

1. Choose the most sensible default interpretation and proceed.
2. State that interpretation plainly.
3. Offer the main alternative as the closing question for that turn.

Never invent missing facts that were not implied by the user request.

## Phase 2: Resolve Against Live Catalog

1. Fetch live filters via:
   - `./prolific filters -n`
2. Use only filter IDs and choice IDs that exist in this environment.
3. If no direct match exists, report the gap and propose nearest alternatives.
4. For capability-style questions ("what can we target for X?"), inspect catalog first before answering.

## Phase 3: Generate and Execute

1. Build a count-template JSON with:
   - `workspace_id` (recommended)
   - `filters`
2. Preview eligibility count without saving via:
   - `./prolific filters count -t <template> -w <workspace_id>`
3. Capture the eligible participant count from command output.
4. Return both the count result and the concrete payload used.
5. If user asks for approximate global crowd size, you may send an empty `filters` array and let the API decide behavior.

## Phase 4: Refine Loop

1. If count is too small, propose broadening edits.
2. If count is too broad, propose narrowing edits.
3. Re-run with one selected edit.
4. Stop when the user accepts audience definition or after three refinement turns.

## Phase 5: Post-Acceptance Actions

After the user confirms they are happy with the audience:

1. Ask as a numbered list:
   1. Would you like to create a filter set from these filters?
   2. Would you like to create a draft study with these filters?
2. If yes, create filter set via:
   - `./prolific filter-sets create -t <template> -w <workspace_id>`
3. If user says yes to item 2 in the list, proceed with draft-study follow-up.
4. If yes, gather only missing required study inputs and then proceed with draft study creation flow.
5. Do not ask the filters attachment choice again if it was already confirmed in the previous step.

## Count Integrity Rules

- Never invent or estimate participant counts.
- Every number must come from a fresh `filters count` result for the current criteria.
- If criteria change, re-run count before stating a number.
- If no count can be fetched, say so plainly and continue without fabricating numbers.

## Run Rules

- Keep the flow explicit and deterministic.
- Do not silently mutate constraints.
- Always show assumptions before execution.
- Prefer one clarifying question over guessing when uncertain.

## Deliverables Per Scenario

1. Final audience JSON artifact
2. Command used to preview eligible count
3. Eligible count result
4. One-paragraph summary of interpretation quality

## Count Template Shape

Use this shape for count preview requests:

```json
{
  "workspace_id": "<workspace_id>",
  "filters": [
    {
      "filter_id": "<select-or-range-filter-id>",
      "selected_values": ["<choice-id>"]
    }
  ]
}
```

Global count probe shape (when supported):

```json
{
  "workspace_id": "<workspace_id>",
  "filters": []
}
```

## Minimum Spike Success Criteria

1. End-to-end loop succeeds on at least 4 of 6 scenarios.
2. Generated payloads require no manual schema fixes.
3. All mappings use real filter IDs from the active environment.
4. Findings include top blockers with concrete examples.

$ARGUMENTS
