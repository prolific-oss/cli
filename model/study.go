package model

import (
	"fmt"
	"time"
)

const (
	// StatusUnpublished is a valid study status
	StatusUnpublished = "unpublished"
	// StatusActive is a valid study status
	StatusActive = "active"
	// StatusScheduled is a valid study status
	StatusScheduled = "scheduled"
	// StatusAwaitingReview is a valid study status
	StatusAwaitingReview = "awaiting review"
	// StatusCompleted is a valid study status
	StatusCompleted = "completed"
	// StatusAll is a mock status that allows us to list all studies.
	StatusAll = "all"
)

// StudyStatuses represents the allows statuses for the system
var StudyStatuses = []string{
	StatusUnpublished,
	StatusActive,
	StatusScheduled,
	StatusAwaitingReview,
	StatusCompleted,
}

// StudyListStatus represents what status we can filter on for the list
var StudyListStatus = []string{
	StatusUnpublished,
	StatusActive,
	StatusCompleted,
	StatusAll,
}

const (
	// TransitionStudyPublish will allow us to publish a study
	TransitionStudyPublish = "PUBLISH"
	// TransitionStudyPause will allow us to pause a study
	TransitionStudyPause = "PAUSE"
	// TransitionStudyStart will allow us to start a study
	TransitionStudyStart = "START"
	// TransitionStudyStop will allow us to stop a study
	TransitionStudyStop = "STOP"
)

// TransitionList is the list of transitions we can use on a Study.
var TransitionList = []string{
	TransitionStudyPublish,
	TransitionStudyStart,
	TransitionStudyPause,
	TransitionStudyStop,
}

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
		DisplayDetails string `json:"details_display"`
	} `json:"eligibility_requirements"`
	Filters                 []Filter    `json:"filters"`
	Desc                    string      `json:"description"`
	EstimatedCompletionTime int         `json:"estimated_completion_time"`
	MaximumAllowedTime      int         `json:"maximum_allowed_time"`
	CompletionURL           string      `json:"completion_url"`
	ExternalStudyURL        string      `json:"external_study_url"`
	PublishedAt             interface{} `json:"published_at"`
	StartedPublishingAt     interface{} `json:"started_publishing_at"`
	AwardPoints             int         `json:"award_points"`
	PresentmentCurrencyCode string      `json:"presentment_currency_code"`
	CurrencyCode            string      `json:"currency_code"`
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
	Status                 string            `json:"status"`
	AverageRewardPerHour   float64           `json:"average_reward_per_hour"`
	DeviceCompatibility    []string          `json:"device_compatibility"`
	PeripheralRequirements []interface{}     `json:"peripheral_requirements"`
	PlacesTaken            int               `json:"places_taken"`
	EstimatedRewardPerHour float64           `json:"estimated_reward_per_hour"`
	Ref                    interface{}       `json:"_ref"`
	StudyType              string            `json:"study_type"`
	TotalCost              float64           `json:"total_cost"`
	PublishAt              interface{}       `json:"publish_at"`
	IsPilot                bool              `json:"is_pilot"`
	IsUnderpaying          interface{}       `json:"is_underpaying"`
	SubmissionsConfig      SubmissionsConfig `json:"submissions_config"`
}

// CreateStudy is responsible for capturing what fields we need to send
// to Prolific to create a study. The `mapstructure` is so we can take a viper
// configuration file.
type CreateStudy struct {
	Name             string `json:"name" mapstructure:"name"`
	InternalName     string `json:"internal_name" mapstructure:"internal_name"`
	Description      string `json:"description" mapstructure:"description"`
	ExternalStudyURL string `json:"external_study_url" mapstructure:"external_study_url"`
	// Enum "question", "url_parameters" (Recommended), "not_required"
	ProlificIDOption string `json:"prolific_id_option" mapstructure:"prolific_id_option"`
	CompletionCode   string `json:"completion_code" mapstructure:"completion_code"`
	// Enum: "url", "code"
	CompletionOption     string `json:"completion_option" mapstructure:"completion_option"`
	TotalAvailablePlaces int    `json:"total_available_places" mapstructure:"total_available_places"`
	// Minutes
	EstimatedCompletionTime int     `json:"estimated_completion_time" mapstructure:"estimated_completion_time"`
	MaximumAllowedTime      int     `json:"maximum_allowed_time" mapstructure:"maximum_allowed_time"`
	Reward                  float64 `json:"reward" mapstructure:"reward"`
	// Enum: "desktop", "tablet", "mobile"
	DeviceCompatibility []string `json:"device_compatibility" mapstructure:"device_compatibility"`
	// Enum: "audio", "camera", "download", "microphone"
	PeripheralRequirements []string `json:"peripheral_requirements" mapstructure:"peripheral_requirements"`
	SubmissionsConfig      struct {
		MaxSubmissionsPerParticipant int `json:"max_submissions_per_participant" mapstructure:"max_submissions_per_participant"`
		MaxConcurrentSubmissions     int `json:"max_concurrent_submissions" mapstructure:"max_concurrent_submissions"`
	} `json:"submissions_config" mapstructure:"submissions_config"`
	EligibilityRequirements []struct {
		Attributes []struct {
			ID    string `json:"id" mapstructure:"id"`
			Index any    `json:"index,omitempty" mapstructure:"index,omitempty"`
			Value any    `json:"value" mapstructure:"value"`
		} `json:"attributes" mapstructure:"attributes"`
		Query struct {
			ID string `json:"id" mapstructure:"id"`
		} `json:"query" mapstructure:"query"`
		Cls string `json:"_cls" mapstructure:"_cls"`
	} `json:"eligibility_requirements" mapstructure:"eligibility_requirements"`
}

// UpdateStudy represents the model we will send back to Prolific to update
// the study.
type UpdateStudy struct {
	TotalAvailablePlaces int `json:"total_available_places,omitempty"`
}

// SubmissionsConfig represents configuration around submission gathering
type SubmissionsConfig struct {
	MaxSubmissionsPerParticipant int `json:"max_submissions_per_participant"`
	MaxConcurrentSubmissions     int `json:"max_concurrent_submissions"`
}

// FilterValue will help the bubbletea views run
func (s Study) FilterValue() string { return s.Name }

// Title will set the main string for the view.
func (s Study) Title() string { return s.Name }

// Description will set the secondary string the view.
func (s Study) Description() string {
	return fmt.Sprintf("%s - %s - %d places available - %s", s.Status, s.StudyType, s.TotalAvailablePlaces, s.Desc)
}

// GetCurrencyCode handles the logic about which internal fields to use to decide
// which currency to display. Defaults to GBP.
func (s Study) GetCurrencyCode() string {
	if s.PresentmentCurrencyCode != "" {
		return s.PresentmentCurrencyCode
	}

	if s.CurrencyCode != "" {
		return s.CurrencyCode
	}

	return DefaultCurrency
}
