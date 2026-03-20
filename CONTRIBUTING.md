# Contributing to Prolific CLI

Welcome! The Prolific CLI is an open source command-line tool for the [Prolific](https://www.prolific.com) research platform. Whether you're fixing a bug, adding a feature, improving documentation, or triaging issues, we appreciate your help. See the [README](README.md) for an overview of what the CLI does.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Ways to Contribute](#ways-to-contribute)
- [Questions](#questions)
- [Discuss Before You Code](#discuss-before-you-code)
- [Reporting Bugs](#reporting-bugs)
- [Suggesting Features](#suggesting-features)
- [Your First Contribution](#your-first-contribution)
- [Development Setup](#development-setup)
- [Pull Request Process](#pull-request-process)
- [Commit Message Format](#commit-message-format)
- [Code Style & Testing](#code-style--testing)
- [Code Review](#code-review)
- [Recognition](#recognition)
- [License](#license)

## Code of Conduct

We expect all contributors to be respectful, constructive, and inclusive. Harassment, discrimination, and bad-faith behaviour are not tolerated. We plan to formally adopt the [Contributor Covenant](https://www.contributor-covenant.org/) — until then, please treat others the way you'd want to be treated.

## Ways to Contribute

There are many ways to help:

- **Code** — fix bugs, add features, improve performance
- **Documentation** — improve guides, add examples, fix typos
- **Bug reports** — file clear, reproducible issues
- **Feature requests** — suggest new commands or enhancements
- **Triage** — help reproduce bugs, label issues, review PRs

If you're new to open source, we're happy to mentor you through your first contribution. Just mention it in your issue or PR and we'll provide extra guidance.

## Questions

Use [GitHub Issues](https://github.com/prolific-oss/cli/issues) for questions, bug reports, and feature requests.

## Discuss Before You Code

For non-trivial changes, please open an issue first to discuss your approach. This avoids wasted effort if the change doesn't align with the project's direction or if someone else is already working on it. Small fixes (typos, one-line bug fixes) can go straight to a PR.

## Reporting Bugs

[Open an issue](https://github.com/prolific-oss/cli/issues/new) to report bugs (issue templates coming soon). A good bug report includes:

- A clear description of the problem
- Steps to reproduce (with CLI commands)
- Actual vs. expected behaviour
- CLI version (`prolific --version`), OS, and Go version (if building from source)
- Error output in a code block

The more detail you provide, the faster we can help.

## Suggesting Features

[Open an issue](https://github.com/prolific-oss/cli/issues/new) to suggest features (issue templates coming soon). Describe your use case, reference existing commands or API endpoints where relevant, and wait for feedback before opening a PR.

## Your First Contribution

Look for issues labelled [`good first issue`](https://github.com/prolific-oss/cli/labels/good%20first%20issue) or [`help wanted`](https://github.com/prolific-oss/cli/labels/help%20wanted).

New to Go or open source? These resources can help:

- [How to Contribute to Open Source](https://opensource.guide/how-to-contribute/)
- [A Tour of Go](https://go.dev/tour/)
- [Effective Go](https://go.dev/doc/effective_go)

**Quick troubleshooting:**

| Problem | Solution |
|---------|----------|
| Lint failures | Run `make lint` locally before committing |
| Mock errors after changing `client/client.go` | Run `make test-gen-mock` to regenerate mocks |
| Tests fail | Run `make test` — tests use mocks, no API token needed |

For more, see [Important Gotchas](DEVELOPMENT.md#important-gotchas) in the development guide.

## Development Setup

```bash
# Clone the repo
git clone https://github.com/prolific-oss/cli.git
cd cli

# Install dependencies and build
make install
make build

# Run tests
make test
```

For full details on project structure, architecture, testing patterns, and more, see [DEVELOPMENT.md](DEVELOPMENT.md).

## Pull Request Process

1. **Branch from `main`** — create a feature branch for your changes
2. **Make your changes** — follow existing code patterns
3. **Run `make all`** — this runs clean, install, build, and test. Pre-commit hooks additionally run lint automatically on each commit
4. **Commit using conventional format** — see [Commit Message Format](#commit-message-format)
5. **Push and open a PR** — include a summary and test plan (PR template coming soon)
6. **Reference the issue** — use `Closes #NNN` in the PR description
7. **Wait for review** — 1 approval from a Prolific maintainer is required
8. **Merging** — only Prolific staff merge PRs

## Commit Message Format

Commits follow conventional commit format, enforced by pre-commit hooks and CI:

```
type(scope): description
```

**Valid types:** `feat`, `fix`, `docs`, `refactor`, `test`, `chore`, `ci`, `build`, `perf`, `style`, `revert`

**Good examples:**

```
feat(DCP-1234): add workspace delete command
fix: handle nil response from submissions endpoint
docs: update development setup instructions
test: add coverage for study list pagination
```

**Bad examples:**

```
fixed stuff                          # no type, vague description
feat: Add new feature.               # capitalised, trailing period
FEAT(DCP-1234): added workspace      # uppercase type, past tense
```

See the [DEVELOPMENT.md git workflow section](DEVELOPMENT.md#git-workflow) for more details.

## Code Style & Testing

- Run `make lint` before committing — pre-commit hooks enforce this automatically
- Run `make test` to verify your changes — tests use mocks, no API token needed
- If you modify the API interface in `client/client.go`, you **must** run `make test-gen-mock` to regenerate mocks

See [DEVELOPMENT.md](DEVELOPMENT.md) for testing patterns, code style rules, and project conventions.

## Code Review

Expect a review within approximately one week. Feedback is constructive — we're all here to build something great together.

## Recognition

Contributors are visible on the [GitHub contributors page](https://github.com/prolific-oss/cli/graphs/contributors) and through git history. Every contribution matters.

## License

This project is licensed under [Apache 2.0](LICENSE). By contributing, you agree that your contributions will be licensed under the same terms.
