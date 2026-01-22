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
	Filters                 []Filter `json:"filters"`
	Desc                    string   `json:"description"`
	EstimatedCompletionTime int      `json:"estimated_completion_time"`
	MaximumAllowedTime      int      `json:"maximum_allowed_time"`
	CompletionURL           string   `json:"completion_url"`
	ExternalStudyURL        string   `json:"external_study_url"`
	PublishedAt             any      `json:"published_at"`
	StartedPublishingAt     any      `json:"started_publishing_at"`
	AwardPoints             int      `json:"award_points"`
	PresentmentCurrencyCode string   `json:"presentment_currency_code"`
	CurrencyCode            string   `json:"currency_code"`
	Researcher              struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Email       string `json:"email"`
		Country     string `json:"country"`
		Institution struct {
			Name any    `json:"name"`
			Logo any    `json:"logo"`
			Link string `json:"link"`
		} `json:"institution"`
	} `json:"researcher"`
	Status                 string            `json:"status"`
	AverageRewardPerHour   float64           `json:"average_reward_per_hour"`
	DeviceCompatibility    []string          `json:"device_compatibility"`
	PeripheralRequirements []any             `json:"peripheral_requirements"`
	PlacesTaken            int               `json:"places_taken"`
	EstimatedRewardPerHour float64           `json:"estimated_reward_per_hour"`
	Ref                    any               `json:"_ref"`
	StudyType              string            `json:"study_type"`
	TotalCost              float64           `json:"total_cost"`
	PublishAt              any               `json:"publish_at"`
	IsPilot                bool              `json:"is_pilot"`
	IsUnderpaying          any               `json:"is_underpaying"`
	SubmissionsConfig      SubmissionsConfig `json:"submissions_config"`
	CredentialPoolID       string            `json:"credential_pool_id"`
}

// CompletionCode represents a completion code configuration for a study.
type CompletionCode struct {
	Code     string                 `json:"code" mapstructure:"code"`
	CodeType string                 `json:"code_type" mapstructure:"code_type"`
	Actions  []CompletionCodeAction `json:"actions" mapstructure:"actions"`
}

// CompletionCodeAction represents an action to take when a completion code is used.
type CompletionCodeAction struct {
	Action           string `json:"action" mapstructure:"action"`
	ParticipantGroup string `json:"participant_group,omitempty" mapstructure:"participant_group,omitempty"`
}

// AccessDetail represents a taskflow study URL allocation.
type AccessDetail struct {
	ExternalURL     string `json:"external_url" mapstructure:"external_url"`
	TotalAllocation int    `json:"total_allocation" mapstructure:"total_allocation"`
}

// CreateStudy is responsible for capturing what fields we need to send
// to Prolific to create a study. The `mapstructure` is so we can take a viper
// configuration file.
type CreateStudy struct {
	Name             string `json:"name" mapstructure:"name"`
	InternalName     string `json:"internal_name" mapstructure:"internal_name"`
	Description      string `json:"description" mapstructure:"description"`
	ExternalStudyURL string `json:"external_study_url,omitempty" mapstructure:"external_study_url"`
	// Enum "question", "url_parameters" (Recommended), "not_required"
	ProlificIDOption string `json:"prolific_id_option" mapstructure:"prolific_id_option"`

	// New: Array of completion code configurations (replaces completion_code and completion_option)
	CompletionCodes []CompletionCode `json:"completion_codes,omitempty" mapstructure:"completion_codes"`

	// DEPRECATED: Use CompletionCodes instead. Kept for backward compatibility.
	CompletionCode string `json:"completion_code,omitempty" mapstructure:"completion_code"`
	// DEPRECATED: Use CompletionCodes instead. Kept for backward compatibility.
	// Enum: "url", "code"
	CompletionOption string `json:"completion_option,omitempty" mapstructure:"completion_option"`

	TotalAvailablePlaces int `json:"total_available_places" mapstructure:"total_available_places"`
	// Minutes
	EstimatedCompletionTime int     `json:"estimated_completion_time" mapstructure:"estimated_completion_time"`
	MaximumAllowedTime      int     `json:"maximum_allowed_time,omitempty" mapstructure:"maximum_allowed_time"`
	Reward                  float64 `json:"reward" mapstructure:"reward"`
	// Enum: "desktop", "tablet", "mobile"
	DeviceCompatibility []string `json:"device_compatibility" mapstructure:"device_compatibility"`
	// Enum: "audio", "camera", "download", "microphone"
	PeripheralRequirements []string `json:"peripheral_requirements,omitempty" mapstructure:"peripheral_requirements"`
	// Study labels for categorization (e.g., "ai_annotation")
	StudyLabels []string `json:"study_labels,omitempty" mapstructure:"study_labels"`

	// New: Array of access details for taskflow studies with multiple URLs (replaces access_details_collection_id)
	AccessDetails []AccessDetail `json:"access_details,omitempty" mapstructure:"access_details"`

	// DEPRECATED: Use AccessDetails instead. Kept for backward compatibility.
	// Access details collection ID: ID of the collection to attach to the study (for Taskflow studies)
	AccessDetailsCollectionID string `json:"access_details_collection_id,omitempty" mapstructure:"access_details_collection_id"`

	// Data collection method: "AI_TASK_BUILDER"
	DataCollectionMethod string `json:"data_collection_method,omitempty" mapstructure:"data_collection_method"`
	// Data collection ID: Project/collection/batch ID for data collection
	DataCollectionID string `json:"data_collection_id,omitempty" mapstructure:"data_collection_id"`
	// Data collection metadata: Configuration parameters (optional dict)
	DataCollectionMetadata map[string]any `json:"data_collection_metadata,omitempty" mapstructure:"data_collection_metadata"`

	// New: Predefined filter set configuration
	FilterSetID      string `json:"filter_set_id,omitempty" mapstructure:"filter_set_id"`
	FilterSetVersion int    `json:"filter_set_version,omitempty" mapstructure:"filter_set_version"`

	// New: Custom screening flag
	IsCustomScreening bool `json:"is_custom_screening,omitempty" mapstructure:"is_custom_screening"`

	// New: Content warnings
	ContentWarnings       []string `json:"content_warnings,omitempty" mapstructure:"content_warnings"`
	ContentWarningDetails string   `json:"content_warning_details,omitempty" mapstructure:"content_warning_details"`

	// New: Custom metadata
	Metadata map[string]any `json:"metadata,omitempty" mapstructure:"metadata"`

	// New: JWT security flag for external study URLs
	IsExternalStudyURLSecure bool `json:"is_external_study_url_secure,omitempty" mapstructure:"is_external_study_url_secure"`

	SubmissionsConfig struct {
		MaxSubmissionsPerParticipant int      `json:"max_submissions_per_participant,omitempty" mapstructure:"max_submissions_per_participant"`
		MaxConcurrentSubmissions     int      `json:"max_concurrent_submissions,omitempty" mapstructure:"max_concurrent_submissions"`
		AutoRejectionCategories      []string `json:"auto_rejection_categories,omitempty" mapstructure:"auto_rejection_categories"`
	} `json:"submissions_config,omitempty" mapstructure:"submissions_config"`

	// DEPRECATED: Use Filters or FilterSetID instead. Kept for backward compatibility.
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
	} `json:"eligibility_requirements,omitempty" mapstructure:"eligibility_requirements"`
	Filters          []Filter `json:"filters,omitempty" mapstructure:"filters"`
	Project          string   `json:"project,omitempty" mapstructure:"project"`
	CredentialPoolID string   `json:"credential_pool_id,omitempty" mapstructure:"credential_pool_id"`
}

// UpdateStudy represents the model we will send back to Prolific to update
// the study.
type UpdateStudy struct {
	TotalAvailablePlaces int    `json:"total_available_places,omitempty"`
	CredentialPoolID     string `json:"credential_pool_id,omitempty"`
}

// SubmissionsConfig represents configuration around submission gathering
type SubmissionsConfig struct {
	MaxSubmissionsPerParticipant int      `json:"max_submissions_per_participant"`
	MaxConcurrentSubmissions     int      `json:"max_concurrent_submissions"`
	AutoRejectionCategories      []string `json:"auto_rejection_categories,omitempty"`
}

// FilterValue will help the bubbletea views run
func (s Study) FilterValue() string { return s.Name }

// Title will set the main string for the view.
func (s Study) Title() string { return s.Name }

// Description will set the secondary string the view.
func (s Study) Description() string {
	return fmt.Sprintf("%s - %s - %d places available", s.Status, s.StudyType, s.TotalAvailablePlaces)
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
