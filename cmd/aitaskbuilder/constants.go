package aitaskbuilder

const (
	// Error messages
	ErrAtLeastOneInstructionRequired = "at least one instruction must be provided"
	ErrBatchIDRequired               = "batch ID is required"
	ErrBatchNotFound                 = "batch not found"
	ErrBothInstructionInputsProvided = "cannot specify both instructions file (-f) and JSON string (-j)"
	ErrBothBatchItemsInputsProvided  = "cannot specify both batch-items file (-f) and batch-items JSON (-j)"
	ErrBatchItemsMutuallyExclusive   = "cannot combine --clear-batch-items with batch-items file (-f) or batch-items JSON (-j)"
	ErrBatchItemsMustBeArray         = "batch_items must be a JSON array"
	ErrBatchItemsMustBeNonEmpty      = "batch_items must contain at least one page"
	ErrDatasetIDRequired             = "dataset ID is required"
	ErrDatasetNotFound               = "dataset not found"
	ErrInstructionInputRequired      = "either instructions file (-f) or JSON string (-j) must be provided"
	ErrNameRequired                  = "name is required"
	ErrSchemaInvalidJSON             = "schema contains invalid JSON"
	ErrSchemaStrictSetInBoth         = "cannot set strict in both --schema and --strict"
	ErrStrictRequiresSchema          = "--strict requires --schema"
	ErrSchemaMustBeObject            = "schema must be a JSON object"
	ErrSchemaFieldsRequired          = "schema must define at least one field"
	ErrSchemaMultipleTaskGroupID     = "schema may define at most one task_group_id field"
	ErrTaskIntroductionRequired      = "task introduction is required"
	ErrTaskNameRequired              = "task name is required"
	ErrTaskStepsRequired             = "task steps is required"
	ErrTasksPerGroupMinimum          = "tasks per group must be at least 1"
	ErrWorkspaceIDRequired           = "workspace ID is required"
	ErrAtLeastOneUpdateFieldRequired = "at least one of --name, --dataset-id, task detail flags, batch-items flags, or --auto-sync must be provided"
)
