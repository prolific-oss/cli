package model

// FilterSet holds information about the filter
type FilterSet struct {
	ID                       string   `json:"id" mapstructure:"id"`
	Name                     string   `json:"name" mapstructure:"name"`
	OrganisationID           string   `json:"organisation_id" mapstructure:"organisation_id"`
	WorkspaceID              string   `json:"workspace_id" mapstructure:"workspace_id"`
	Version                  int      `json:"version" mapstructure:"version"`
	IsDeleted                bool     `json:"is_deleted" mapstructure:"is_deleted"`
	IsLocked                 bool     `json:"is_locked" mapstructure:"is_locked"`
	EligibleParticipantCount int      `json:"eligible_participant_count" mapstructure:"eligible_participant_count"`
	Filters                  []Filter `json:"filters" mapstructure:"filters"`
}

// CreateFilterSet holds the fields needed to create a new filter set.
type CreateFilterSet struct {
	Name           string   `json:"name,omitempty" mapstructure:"name"`
	WorkspaceID    string   `json:"workspace_id,omitempty" mapstructure:"workspace_id"`
	OrganisationID string   `json:"organisation_id,omitempty" mapstructure:"organisation_id"`
	Filters        []Filter `json:"filters,omitempty" mapstructure:"filters"`
}
