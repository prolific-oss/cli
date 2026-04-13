# GitHub Copilot Instructions

> For all coding instructions, architecture, patterns, and conventions, see [AGENTS.md](../AGENTS.md).

## Changelog & Releases

When asked to create a release:

- Run `make changelog VERSION=x.y.z` (with `git-cliff` installed) so grouped notes are generated; do not hand-edit the full release section unless adding content under `## next` before running the command.
- Open a PR with the updated `CHANGELOG.md` and add the **`release`** label. Merging to `main` triggers `create-release.yml`, which only creates the tag, GitHub Release, and binaries when that merged PR has the `release` label.
- Follow the existing format in `CHANGELOG.md`:
  - `## next` for hand-written notes to merge into the next run of `make changelog`.
  - `## x.y.z` for released versions, with bullet points for changes (headings are bare semver, **no** `v` prefix).
  - Do not include dates — version numbers only.
- Git tags and GitHub Release names use a `v` prefix (`v1.0.1`), matching what `create-release.yml` publishes; do not name releases with a bare version string.