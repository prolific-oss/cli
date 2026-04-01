package model

import "time"

// ParticipantGroup holds information about the group
type ParticipantGroup struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	ProjectID        string `json:"project_id"`
	WorkspaceID      string `json:"workspace_id"`
	Description      string `json:"description"`
	ParticipantCount int    `json:"participant_count"`
}

// CreateParticipantGroup is the payload to create a new participant group
type CreateParticipantGroup struct {
	Name           string   `json:"name"`
	WorkspaceID    string   `json:"workspace_id,omitempty"`
	Description    string   `json:"description,omitempty"`
	ParticipantIDs []string `json:"participant_ids,omitempty"`
}

// ParticipantGroupMembership holds information about a member in a group
type ParticipantGroupMembership struct {
	ParticipantID   string    `json:"participant_id"`
	DatetimeCreated time.Time `json:"datetime_created"`
}
