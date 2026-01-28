package collection

const (
	// Error messages
	ErrCollectionItemsRequired  = "at least one collection item must be provided"
	ErrPageItemsRequired        = "each page must have at least one item in page_items"
	ErrWorkspaceIDRequired      = "workspace ID is required"
	ErrNameRequired             = "name is required"
	ErrWorkspaceNotFound        = "workspace not found"
	ErrTaskDetailsRequired      = "task_details is required"
	ErrTaskNameRequired         = "task_details.task_name is required"
	ErrTaskIntroductionRequired = "task_details.task_introduction is required"
	ErrTaskStepsRequired        = "task_details.task_steps is required"

	// Feature access constants for AI Task Builder Collections (see DCP-2152)
	FeatureNameAITBCollection       = "AI Task Builder Collections"
	FeatureContactURLAITBCollection = "https://researcher-help.prolific.com/en/"
)
