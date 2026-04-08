# GitHub Copilot Instructions

> For all coding instructions, architecture, patterns, and conventions, see [AGENTS.md](../AGENTS.md).

## Changelog & Releases

When asked to create a release:

- Use the last git tag and summarise the commits since.
- Increment the version in `CHANGELOG.md` and define a new heading.
- Follow the existing format in `CHANGELOG.md`:
  - `## next` for the upcoming version.
  - `## x.y.z` for previous versions, with bullet points for changes (headings are bare semver, **no** `v` prefix).
  - Do not include dates — version numbers only.
- Git tags and GitHub Release names use a `v` prefix (`v1.0.1`), matching what `create-release.yml` publishes; do not name releases with a bare version string.