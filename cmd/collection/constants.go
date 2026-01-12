package collection

const (
	// Error messages
	ErrCollectionItemsRequired = "at least one collection item must be provided"
	ErrPageItemsRequired       = "each page must have at least one item in page_items"
	ErrWorkspaceIDRequired     = "workspace ID is required"
	ErrNameRequired            = "name is required"
	ErrWorkspaceNotFound       = "workspace not found"

	// Feature access constants for AI Task Builder Collections (see DCP-2152)
	FeatureNameAITBCollection       = "AI Task Builder Collections"
	FeatureContactURLAITBCollection = "https://researcher-help.prolific.com/en/"
)
