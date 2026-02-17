# CHANGELOG

## next

- Bump Go version to 1.26.

## 0.0.58

### AI Task Builder

- **Add `file_upload` instruction type support:**
  - Create file upload instructions with configurable file types, size limits, and count constraints
  - Validate file extensions (must start with dot), positive file sizes, and min/max file counts
  - Render file upload responses with file metadata (name, size, content type, file key)
  - Full validation ensures `max_file_count >= min_file_count` and both must be >= 1
- **Add `free_text_with_unit` instruction type:**
  - Supersedes deprecated `multiple_choice_with_unit` instruction type
  - Support unit selection with customizable unit options (e.g., kg/lbs, °C/°F, USD/EUR)
  - Configure unit position (prefix/suffix) and optional default unit
  - Properly typed `UnitPosition` enum with `prefix` and `suffix` constants
- **Add `completion_codes` support:**
  - Support completion codes in batch and collection configurations
  - Fix `omitempty` handling for completion_code field
- **Improve response rendering:**
  - Correctly handle all response types using unified Answer array structure
  - Display explanations for `multiple_choice_with_free_text` responses
  - Support empty/nil responses gracefully with fallback display
  - Update response model to match API schemas exactly (Zod schema compliance)
- **Add collection commands:**
  - `aitaskbuilder collection create` - Create collections from JSON/YAML templates
  - `aitaskbuilder collection update` - Update existing collections
  - `aitaskbuilder collection list` - List all collections in a workspace
  - `aitaskbuilder collection get` - View collection details
  - `aitaskbuilder collection preview` - Open collection preview in browser
  - `aitaskbuilder collection publish` - Publish collections with study configuration
- **Improve collection support:**
  - Add `task_details` validation for collections
  - Support content block types (rich_text, image) in collections
  - Correctly map collection schema with `collection_items` and `page_items`
  - Add comprehensive collection examples (JSON/YAML)
- **Breaking changes:**
  - Removed `multiple_choice_with_unit` instruction type (superseded by `free_text_with_unit`)
  - Removed duplicate `PageItemType` constants (use `InstructionType` from collection.go)
  - Updated `CollectionPageItem` to reference `InstructionType` enum

### Participant Management

- Fix participant group list command to use `workspace_id` query parameter

### Bonus Payments

- **Add `bonus` commands:**
  - `bonus create` - Create bonus payments from CSV file or command-line arguments
  - `bonus pay` - Pay pending bonus payments
  - Validate bonus amounts (reject NaN and Inf values)
  - Support CSV file format with participant IDs and amounts
  - Document minor currency units in response fields
  - Add comprehensive CSV format examples to help output

### Messaging

- **Add group messaging commands:**
  - `message bulk-send` - Send messages to multiple participants
  - `message send-group` - Send messages to participant groups
- Fix message API conformance (sender vs sender_id, datetime_created fields)
- Add required flag validation to all message send commands

### Developer Experience

- Add `CLAUDE.md` with project guidelines for AI-assisted development
- Rename `CRUSH.md` to `DEVELOPMENT.md` for clarity
- Convert CLI command template to Claude Code skill
- Add CLI command plan template

### Bug Fixes

- Fix `goconst` lint for shared test error string
- Correct Go workflow badge on README
- Improve feature access error detection and messaging
- Fix collection item_count display in detail view

## 0.0.57

- Add `study credentials-report` command to download CSV report of credential usage for a study:
  - Returns participant IDs, submission IDs, usernames, and credential status (USED/UNUSED)
  - Available only for studies with credentials configured
  - Usage: `prolific study credentials-report <study-id> > report.csv`
- Add `credentials` command to manage credential pools:
  - `credentials create` - Create new credential pools with comma-separated credentials or from a file
  - `credentials update` - Update existing credential pools with new credentials
  - `credentials list` - List all credential pools in a workspace
- Add `study set-credential-pool` command to set or update the credential pool on a draft study:
  - Allows attaching a credential pool to a study created without one
  - Allows changing the credential pool on an existing draft study
  - Usage: `prolific study set-credential-pool <study-id> -c <credential-pool-id>`
- Restructure `aitaskbuilder` dataset commands under `dataset` entity:
  - `aitaskbuilder dataset create` - Create new datasets (previously `aitaskbuilder create-dataset`)
  - `aitaskbuilder dataset check` - Check dataset status (previously `aitaskbuilder getdatasetstatus`)
  - `aitaskbuilder dataset upload` - Upload CSV files to datasets
- Restructure `aitaskbuilder` batch commands under `batch` entity:
  - `aitaskbuilder batch create` - Create new batches with task details
  - `aitaskbuilder batch instructions` - Add instructions to batches
  - `aitaskbuilder batch setup` - Configure batches with dataset and task groups
  - `aitaskbuilder batch view` - View batch details (previously `aitaskbuilder getbatch`)
  - `aitaskbuilder batch list` - List batches in a workspace (previously `aitaskbuilder getbatches`)
  - `aitaskbuilder batch check` - Check batch status (previously `aitaskbuilder getbatchstatus`)
  - `aitaskbuilder batch responses` - List batch task responses (previously `aitaskbuilder getresponses`)
  - `aitaskbuilder batch tasks` - Retrieve all task IDs for a batch
- Fix `aitaskbuilder dataset create` command:
  - Corrected API endpoint from `/api/v1/data-collection/workspaces/{id}/datasets/` to `/api/v1/data-collection/datasets`
  - Added `workspace_id` field to request payload
  - Updated response handling to match API structure (fields at top level)
  - Enhanced output to display all dataset details (ID, name, status, created_at, etc.)
- Fix `aitaskbuilder batch instructions` command:
  - Corrected response handling to expect array of instruction objects
  - Enhanced output to display created instruction IDs and metadata
- Fix `aitaskbuilder batch setup` command:
  - Fixed handling of empty response body (202 Accepted)
  - Improved error message parsing for AI Task Builder endpoints
  - Added support for flat error format `{message, detail}`
- Add JSON tags to `CreateBatchParams` struct for proper API serialization
- Use type-safe enums for AI Task Builder types:
  - `AITaskBuilderBatchStatusEnum` for batch statuses
  - `DatasetStatus` for dataset statuses
  - `InstructionType` for instruction types
- Use `ErrWorkspaceIDRequired` constant for consistent error handling
- Make it clear if a study is underpaying in the `study view` command.
- Provide an `--underpaying` flag for the `study list` command to filter down to only underpaying studies.
- Add support for `go install github.com/prolific-oss/cli/cmd/prolific@latest`.
- Add `owner` and `description` fields to `project create` command.
- Fix study list filtering to allow project and status filters to be used together.

## 0.0.56

- Add Apache 2 License.
- Add `aitaskbuilder` command to the root of the application.
- Bump the project to Go 1.25.
- Remove the naivety distribution rate, which has been deprecated by Prolific.
- General maintenance and dependency updates.

## 0.0.55

- Provide more context for the filter view command
- Support project field in study creation templates
- Dependency updates
- Provide examples for all API doc examples
- Rebrand from `benmatselby/prolificli` back to `prolific-oss/cli`

## 0.0.54

- Bump the project to Go 1.24.
- Some formatting fixes for the Docker image (Updated linter).
- Define the Copilot instructions for the project.

## 0.0.53

- Dependency management

## 0.0.52

- Provide the ability to transition a study `study transition 65ecbe2cba93fe76699213f5 -a START`.
  - Actions include: `START`, `PAUSE`, `STOP`, `PUBLISH`.

## 0.0.51

- Consistently use Go 1.23 in the project.

## 0.0.50

- Allow the user to create a study with the `filters` attribute.
- Remove the description from the interactive study list view.

## 0.0.49

- Allow the user to define the default `workspace` value in the config file `$HOME/.config/benmatselby/prolific.yaml`.
  - This will be the default value for all commands that take a `--workspace` argument.

## 0.0.48

- Display the screeners/filters applied to study. `prolific study view [study-id]`
  - This will now render a "Filters" section.
  - Documented [here](https://docs.prolific.com/docs/api-docs/public/#tag/Studies/operation/GetStudy)

## 0.0.47

- Show selected ranges for the filter set view command `filter-sets view 65ecbe2cba93fe76699213f5`.
- Better error handling if you encounter a permission error.
- Project maintenance, removed deprecated go linters.

## 0.0.46

- Provide the ability to list your campaigns (Bring your own participants) `campaign list -w [workspace_id]`.
- Bump go version in the Docker image to 1.22.3.

## 0.0.45

- Provide a filter set view command `filter-sets view 65ecbe2cba93fe76699213f5`.
  - This includes the ability to open the filter set in the web application using your system browser.
  - Just add the `-W` flag.

## 0.0.44

- Provide the ability to open some resources in a browser with the `-W flag`.
  - `project view [id] -W`
  - `study view [id] -W`
  - This will open the resource in the Prolific web application with your system browser.

## 0.0.43

- Capture the fact a project may not exist when trying to view it.
- Render Workspace on the `project view [id]` command.
- Render the application link to the project view.
- Standardise the way we render the application link.

## 0.0.42

- Update the go runtime to 1.22.

## 0.0.41

- Provide a project detail view `project view [id]`.

## 0.0.38

- Update dependencies.
- Minor tidyings.

## 0.0.37

- Update dependencies.

## 0.0.36

- Move to `.com` in the configuration and docs.

## 0.0.35

- Remove the balance data from `whoami`.

## 0.0.34

- Update the go runtime to 1.21

## 0.0.33

- Validation on zero results for list views.

## 0.0.32

- Update the README with the summary output of the application.
- Document the licenses used for our dependencies.
- Update the go.mod file to Go 1.20.

## 0.0.31

- Add paging to the `participant list` command.

## 0.0.30

- Provide `filter-sets list` command that allows you to pull back a list of your Filter Sets.

## 0.0.29

- Make sure all commands have both a long description and also usage examples.

## 0.0.28

- Provide an easy access link to the message centre in the `message list` command.
- Show the submissions configuration data when rendering a study.
- Provide the ability to create studies with eligibility requirements.
- Provide more details in the `requirements` view so you can construct the `eligibility_requirements` payload in create study.

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
