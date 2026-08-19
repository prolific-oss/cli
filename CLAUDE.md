# CLAUDE.md

> Claude Code-specific configuration. For all coding instructions, see [AGENTS.md](./AGENTS.md).

## Slash Commands

- `/cli-command-create` — automates the full workflow for adding a new CLI command (see `.claude/skills/cli-command-create/`)
- `/schema-drift-fix` — checks the CLI against the live API spec for drift and drafts the fix, optionally starting from a filed `schema-drift` issue (see `.claude/skills/schema-drift-fix/`)