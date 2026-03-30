package model

import (
	"fmt"
	"time"
)

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

// FilterValue enables filtering in the Bubbletea list view.
func (s Submission) FilterValue() string { return s.ParticipantID }

// Title is the primary display string in the Bubbletea list view.
func (s Submission) Title() string { return s.ParticipantID }

// Description is the secondary display string in the Bubbletea list view.
func (s Submission) Description() string {
	return fmt.Sprintf("%s - %ds", s.Status, s.TimeTaken)
}
