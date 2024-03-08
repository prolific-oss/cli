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
	StudyCode     string        `json:"study_code"`
	StarAwarded   bool          `json:"star_awarded"`
	BonusPayments []interface{} `json:"bonus_payments"`
	IP            string        `json:"ip"`
}

// Workspace represents the workspace model
type Workspace struct {
	ID                      string  `json:"id"`
	Title                   string  `json:"title"`
	Description             string  `json:"description"`
	Users                   []User  `json:"users"`
	NaivetyDistributionRate float64 `json:"naivety_distribution_rate"`
}

// Project represents the project model
type Project struct {
	ID                      string  `json:"id"`
	Title                   string  `json:"title"`
	Description             string  `json:"description"`
	Owner                   string  `json:"owner"`
	Users                   []User  `json:"users"`
	NaivetyDistributionRate float64 `json:"naivety_distribution_rate"`
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

// Message represents the message model
type Message struct {
	DatetimeCreated time.Time              `json:"datetime_created"`
	Body            string                 `json:"body"`
	SenderID        string                 `json:"sender_id"`
	StudyID         string                 `json:"study_id,omitempty"`
	Data            map[string]interface{} `json:"data,omitempty"`
}

// UnreadMessage represents the unread message model
type UnreadMessage struct {
	DatetimeCreated time.Time `json:"datetime_created"`
	Body            string    `json:"body"`
	Sender          string    `json:"sender"`
}
