# GitHub Copilot Instructions for prolificli

## Project Overview

- **prolificli** is a CLI tool for retrieving and displaying data from the Prolific Platform.
- Written in Go, using Cobra for CLI

## Coding Guidelines

- Follow idiomatic Go practices and use `gofmt` for formatting.
- Use dependency injection for clients (see `client/client.go`).
- Prefer clear, user-friendly CLI flags and help text.
- Use semantic commit messages.

## Features & Conventions

- All user-facing strings should be clear and concise.
- Keep the codebase well-documented and update the `CHANGELOG.md` for every release.

## Testing

- Use `make test-cov` to run all tests with coverage.
- Use `make test` to run tests without coverage, which is the default.
- Use table-driven tests where appropriate.
- Where applicable use an io.Writer that captures the output and then compare it against expected output. (see `cmd/workspace/list_test.go` for examples).

## Dependencies

- Use `go get` to add dependencies.
- Keep dependencies up to date and tidy with `go mod tidy`.

## Pull Requests

- Ensure all tests pass before submitting a PR.
- Update documentation and changelog as needed.
- Keep PRs focused and well-described.

## Misc

- Use the `main.go` entrypoint for CLI execution.
- Dockerfile and Makefile are provided for builds and CI.

## Changelog

- When asked to create a release, use the last git tag and summarise the commits since.
- Increment the version in the `CHANGELOG.md` file and define a new heading.
- Follow the format in `CHANGELOG.md` for consistency, for example:
  - `## next` for the next version.
  - `## x.y.z` for previous versions, with bullet points for changes.
  - Do not use the date in the changelog, just the version number.
