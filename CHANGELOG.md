# CHANGELOG

## next

- Provide an easy access link to the message centre in the `message list` command.

## 0.0.27

- Provide `message list` command that allows you to pull back messages from the Prolific Platform.
- Provide `message send` command that allows you to send messages on the Prolific Platform.
- Add the researcher ID to the `whoami` command.

## 0.0.26

- Provide paging options for the `project list` command.

## 0.0.25

- Provide the ability to specify a different workspace on the `hook list` command.
- Add paging to the `hook list` command.
- Add a long description for the `hook event_types` command.

## 0.0.24

- Rename app/binary to `prolific` rather than `prolificli` for branding.

## 0.0.23

- Provide the ability to increase places on your study via the `study increase-places` command.

## 0.0.22

- Provide paging options for the following commands:
  - `submission list` command.
  - `workspace list` command.
  - This means you can now specify the following options:
  - `-l, --limit int           Limit the number of events returned (default 1)`
  - `-o, --offset int          The number of events to offset`

## 0.0.21

- Lint the Dockerfile.
- Define a `pre-commit` hook for the repo, so we can get a quicker feedback loop.
- Fix some `gosec` warnings.
- Provide paging options for the `hook events` command.
  - You can now specify the following options:
  - `-l, --limit int           Limit the number of events returned (default 1)`
  - `-o, --offset int          The number of events to offset`

## 0.0.20

- No longer build binaries for Solaris.

## 0.0.19

- Move this over to the [prolific-oss](https://github.com/prolific-oss) namespace.

## 0.0.18

- Add the ability to list participant groups via `participant list -p [project_id]`.
- Add the ability to view a participant group via `participant view [group_id]`.

## 0.0.17

- Add the ability to view hook events for a given subscription `hook events -s [subscription_id]`.

## 0.0.16

- Add the description to the `hook event-types` view.

## 0.0.15

- Ability to list studies in a different project.
  - This was nearly released from Prolific, so now you can specify a `-p` flag to `study list`.
- Ability to duplicate a study via `study duplicate [id]`.
- Provide an error message if `PROLIFIC_TOKEN` is not set.

## 0.0.14

- Partial fix for the currency display issue.
  - The study model now can handle the `presentment_currency_code` and `currency_code` fields and figure out what to display.
  - It will still default to `GBP` if required.

## 0.0.13

- Ability to list your hook secrets via `hook secrets`.
- Add the ID and WorkspaceID to the secret output.
- Ability to create a workspace via `workspace create`.
- Ability to create a project in a workspace via `project create`.

## 0.0.12

- Ability to list workspaces and projects.
  - `workspace list`
  - `project list -w [workspace-id]`

## 0.0.11

- Ability to get the event-types you can register for via `hook event-types`.
  - Will just render a list of strings.

## 0.0.10

- Set a default value for `PROLIFIC_URL`. It will default to `https://api.prolific.co`.
- Ability to get your hook subscriptions via `hook list`.

## 0.0.9

- Bumped docker image to Go 1.19 runtime.

## 0.0.8

- Ability to render a list of studies with `--csv` - essentially an export.
- Ability to render a list of submissions.
  - Includes the ability to use `--csv` for a CSV format.

## 0.0.7

- Addition of the version number in the binaries.

## 0.0.6

- The releases will now include the built binaries for different platforms/architectures.

## 0.0.5

- Ability to render a list of studies with `--non-interactive` which displays the the records in a table.
- Ability to select which fields to render for the `study list --non-interactive` command. Add a comma separated list like `--fields=ID,Status,Reward,Name` to the end of the command. Default: `ID, Name, Status`. For more information check out the [wiki](https://github.com/prolific-oss/cli/wiki/Fields-you-can-use-in-the-non-interactive-list-study-command).

## 0.0.4

- Ability to silently create a study. This is helpful if you want to script creating many studies.

## 0.0.3

- Ability to publish a study whilst creating it (if you have sufficient funds).

## 0.0.2

- Ability to create a Study via a YAML/JSON configuration file.

## 0.0.1

Initial release of the `prolific` application.

- Ability to get your user account details.
- Ability to list and filter studies.
- Ability to render details about a study, and the submissions.
