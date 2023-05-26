package client

// MessagePayload represents the JSON payload for sending a message
type SendMessagePayload struct {
	RecipientID string `json:"recipient_id"`
	StudyID     string `json:"study_id"`
	Body        string `json:"body"`
}
