package model

import "time"

// ParticipantGroup holds information about the group
type ParticipantGroup struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ProjectID string `json:"project_id"`
}

// ParticipantGroupMembership holds information about a member in a group
type ParticipantGroupMembership struct {
	ParticipantID   string    `json:"participant_id"`
	DatetimeCreated time.Time `json:"datetime_created"`
}
