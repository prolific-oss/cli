# AI Task Builder Batch Management Commands

## Summary

This PR adds comprehensive batch management functionality to the AI Task Builder CLI, enabling users to create and configure AI data collection batches through the command line. It includes three new core commands for batch creation, instruction configuration, and setup, plus a command to retrieve task IDs. Additionally, it fixes several API integration issues with dataset and batch commands to ensure proper communication with the AI Task Builder API.

## Changes

### New Commands

- **`aitaskbuilder batch create`** - Create new AI Task Builder batches with task details (name, introduction, steps)
- **`aitaskbuilder batch instructions`** - Add evaluation instructions to batches (supports free text, multiple choice, and mixed types)
- **`aitaskbuilder batch setup`** - Configure batches with dataset and task grouping parameters
- **`aitaskbuilder batch tasks`** - Retrieve all task IDs for a given batch

### Bug Fixes

- **Dataset Create Command**:
  - Fixed API endpoint (changed from `/api/v1/data-collection/workspaces/{id}/datasets/` to `/api/v1/data-collection/datasets`)
  - Added `workspace_id` field to request payload
  - Updated response handling to match API structure (fields at top level instead of nested)
  - Enhanced output to display complete dataset details

- **Batch Instructions Command**:
  - Fixed response parsing to handle array of instruction objects (not wrapped in object)
  - Enhanced output to show created instruction IDs and metadata

- **Batch Setup Command**:
  - Fixed handling of 202 Accepted responses with empty body (was causing EOF errors)
  - Improved error message parsing to support both nested and flat error formats
  - Added clear, actionable error messages from API

### API Client Updates

- Added `GetAITaskBuilderTasks()` method to retrieve batch task IDs
- Added `SimpleAPIError` struct for flat error format support
- Enhanced error handling to try multiple error response formats
- Added request/response types for all new batch operations

### Documentation

- Added AI Task Builder study template example (`docs/examples/standard-sample-aitaskbuilder.json`)
- Updated `CHANGELOG.md` with all new features and fixes
- Comprehensive test coverage for all new commands

### Tests

- 100% coverage for all new batch commands
- Tests cover success cases, error handling, validation, and edge cases
- Mock API responses based on actual API behavior
