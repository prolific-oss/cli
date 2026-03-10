package aitaskbuilder

const (
	// Error messages
	ErrAtLeastOneInstructionRequired = "at least one instruction must be provided"
	ErrBatchIDRequired               = "batch ID is required"
	ErrBatchNotFound                 = "batch not found"
	ErrBothInstructionInputsProvided = "cannot specify both instructions file (-f) and JSON string (-j)"
	ErrDatasetIDRequired             = "dataset ID is required"
	ErrDatasetNotFound               = "dataset not found"
	ErrInstructionInputRequired      = "either instructions file (-f) or JSON string (-j) must be provided"
	ErrNameRequired                  = "name is required"
	ErrTaskIntroductionRequired      = "task introduction is required"
	ErrTaskNameRequired              = "task name is required"
	ErrTaskStepsRequired             = "task steps is required"
	ErrTasksPerGroupMinimum          = "tasks per group must be at least 1"
	ErrWorkspaceIDRequired           = "workspace ID is required"
)
