// Package model defines the data structures representing Prolific domain
// entities such as studies, submissions, workspaces, projects, and users.
// These types are used for JSON serialization when communicating with the
// Prolific API.
package model

import (
	"time"
)

// DefaultCurrency is set to GBP if we cannot figure out what currency to
// render based on other factors.
const DefaultCurrency string = "GBP"

// Submission represents a submission to a study from a participant.
type Submission struct {
	ID            string    `json:"id"`
	ParticipantID string    `json:"participant_id"`
	StartedAt     time.Time `json:"started_at"`
	CompletedAt   time.Time `json:"completed_at"`
	IsComplete    bool      `json:"is_complete"`
	TimeTaken     int       `json:"time_taken"`
	Reward        int       `json:"reward"`
	Status        string    `json:"status"`
	Strata        struct {
		DateOfBirth         string `json:"date of birth"`
		EthnicitySimplified string `json:"ethnicity (simplified)"`
		Sex                 string `json:"sex"`
	} `json:"strata"`
	StudyCode     string `json:"study_code"`
	StarAwarded   bool   `json:"star_awarded"`
	BonusPayments []any  `json:"bonus_payments"`
	IP            string `json:"ip"`
}

// Workspace represents the workspace model
type Workspace struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Users       []User `json:"users"`
}

// Project represents the project model
type Project struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Workspace   string `json:"workspace"`
	Owner       string `json:"owner"`
	Users       []User `json:"users"`
}

// User represents a user in the system
type User struct {
	ID    string   `json:"id"`
	Name  string   `json:"name"`
	Email string   `json:"email"`
	Roles []string `json:"roles"`
}

// Secret represents the secrets passed back from Prolific.
type Secret struct {
	ID          string `json:"id"`
	Value       string `json:"value"`
	WorkspaceID string `json:"workspace_id"`
}

// Message represents a message on the Prolific platform.
// The regular messages endpoint returns "sender_id" while the unread
// endpoint returns "sender", so both fields are present.
type Message struct {
	ID              string       `json:"id"`
	SenderID        string       `json:"sender_id,omitempty"`
	Sender          string       `json:"sender,omitempty"`
	Body            string       `json:"body"`
	DatetimeCreated time.Time    `json:"datetime_created"`
	Type            string       `json:"type,omitempty"`
	ChannelID       string       `json:"channel_id"`
	Data            *MessageData `json:"data,omitempty"`
}

// GetSenderID returns the sender identifier regardless of which API
// endpoint populated the message. The regular messages endpoint uses
// "sender_id" while the unread endpoint uses "sender".
func (m Message) GetSenderID() string {
	if m.SenderID != "" {
		return m.SenderID
	}
	return m.Sender
}

// MessageData contains metadata associated with a message.
type MessageData struct {
	StudyID  string `json:"study_id,omitempty"`
	Category string `json:"category,omitempty"`
}
