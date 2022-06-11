package client

import (
	"fmt"
	"time"
)

const (
	StatusUnpublished    = "UNPUBLISHED"
	StatusActive         = "ACTIVE"
	StatusScheduled      = "SCHEDULED"
	StatusAwaitingReview = "AWAITING REVIEW"
	StatusCompleted      = "COMPLETED"
)

// Study represents a Prolific Study
type Study struct {
	ID                      string    `json:"id"`
	Name                    string    `json:"name"`
	InternalName            string    `json:"internal_name"`
	DateCreated             time.Time `json:"date_created"`
	TotalAvailablePlaces    int       `json:"total_available_places"`
	Reward                  float64   `json:"reward"`
	CanAutoReview           bool      `json:"can_auto_review"`
	EligibilityRequirements []struct {
		ID       string `json:"id"`
		Question struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		} `json:"question"`
	} `json:"eligibility_requirements"`
	Desc                    string      `json:"description"`
	EstimatedCompletionTime int         `json:"estimated_completion_time"`
	MaximumAllowedTime      int         `json:"maximum_allowed_time"`
	CompletionURL           string      `json:"completion_url"`
	ExternalStudyURL        string      `json:"external_study_url"`
	PublishedAt             interface{} `json:"published_at"`
	StartedPublishingAt     interface{} `json:"started_publishing_at"`
	AwardPoints             int         `json:"award_points"`
	PresentmentCurrencyCode string      `json:"presentment_currency_code"`
	Researcher              struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Email       string `json:"email"`
		Country     string `json:"country"`
		Institution struct {
			Name interface{} `json:"name"`
			Logo interface{} `json:"logo"`
			Link string      `json:"link"`
		} `json:"institution"`
	} `json:"researcher"`
	Status                 string        `json:"status"`
	AverageRewardPerHour   float64       `json:"average_reward_per_hour"`
	DeviceCompatibility    []string      `json:"device_compatibility"`
	PeripheralRequirements []interface{} `json:"peripheral_requirements"`
	PlacesTaken            int           `json:"places_taken"`
	EstimatedRewardPerHour float64       `json:"estimated_reward_per_hour"`
	Ref                    interface{}   `json:"_ref"`
	StudyType              string        `json:"study_type"`
	TotalCost              float64       `json:"total_cost"`
	PublishAt              interface{}   `json:"publish_at"`
	IsPilot                bool          `json:"is_pilot"`
	IsUnderpaying          interface{}   `json:"is_underpaying"`
}

func (s Study) FilterValue() string { return s.Name }
func (s Study) Title() string       { return s.Name }
func (s Study) Description() string {

	return fmt.Sprintf("%s - %d places available - %s", s.Status, s.TotalAvailablePlaces, s.Desc)

}
