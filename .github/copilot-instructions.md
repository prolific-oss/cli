# GitHub Copilot Instructions

> For all coding instructions, architecture, patterns, and conventions, see [AGENTS.md](../AGENTS.md).

## Changelog & Releases

When asked to create a release:

- Use the last git tag and summarise the commits since.
- Increment the version in `CHANGELOG.md` and define a new heading.
- Follow the existing format in `CHANGELOG.md`:
  - `## next` for the upcoming version.
  - `## x.y.z` for previous versions, with bullet points for changes.
  - Do not include dates — version numbers only.