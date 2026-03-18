# GitHub Copilot Instructions for the Prolific CLI Project

For project overview, architecture, coding guidelines, and testing patterns, see `CLAUDE.md` in the repository root.

## Changelog

- When asked to create a release, use the last git tag and summarise the commits since.
- Increment the version in the `CHANGELOG.md` file and define a new heading.
- Follow the format in `CHANGELOG.md` for consistency, for example:
  - `## next` for the next version.
  - `## x.y.z` for previous versions, with bullet points for changes.
  - Do not use the date in the changelog, just the version number.

## Pull Requests

- Ensure all tests pass before submitting a PR.
- Update documentation and changelog as needed.
- Keep PRs focused and well-described.

## Dependencies

- Use `go get` to add dependencies.
- Keep dependencies up to date and tidy with `go mod tidy`.
