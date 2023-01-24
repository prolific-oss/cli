package model

// ParticipantGroup holds information about the group
type ParticipantGroup struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ProjectID string `json:"project_id"`
}
