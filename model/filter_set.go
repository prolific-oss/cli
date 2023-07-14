package model

// FilterSet holds information about the filter
type FilterSet struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	WorkspaceID string `json:"workspace_id"`
}
