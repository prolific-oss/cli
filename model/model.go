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

// CreateStudy is responsible for capturing what fields we need to send
// to Prolific to create a study.
type CreateStudy struct {
	Name             string `json:"name"`
	InternalName     string `json:"internal_name"`
	Description      string `json:"description"`
	ExternalStudyURL string `json:"external_study_url"`
	// Enum "question", "url_parameters" (Recommended), "not_required"
	ProlificIDOption string `json:"prolific_id_option"`
	CompletionCode   string `json:"completion_code"`
	// Enum: "url", "code"
	CompletionOption     string `json:"completion_option"`
	TotalAvailablePlaces int    `json:"total_available_places"`
	// Minutes
	EstimatedCompletionTime int     `json:"estimated_completion_time"`
	MaximumAllowedTime      int     `json:"maximum_allowed_time"`
	Reward                  float64 `json:"reward"`
	// Enum: "desktop", "tablet", "mobile"
	DeviceCompatibility []string `json:"device_compatibility"`
	// Enum: "audio", "camera", "download", "microphone"
	PeripheralRequirements []string `json:"peripheral_requirements"`
}

// FilterValue will help the bubbletea views run
func (s Study) FilterValue() string { return s.Name }

// Title will set the main string for the view.
func (s Study) Title() string { return s.Name }

// Description will set the secondary string the view.
func (s Study) Description() string {
	return fmt.Sprintf("%s - %s - %d places available - %s", s.Status, s.StudyType, s.TotalAvailablePlaces, s.Desc)
}

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

// Requirement represents an eligibility requirement in the system.
type Requirement struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	// Attributes []struct {
	// 	Label string `json:"label"`
	// 	Name  string `json:"name"`
	// 	Value string `json:"value"`
	// 	Index int    `json:"index"`
	// } `json:"attributes"`
	Query struct {
		ID                  string `json:"id"`
		Question            string `json:"question"`
		Description         string `json:"description"`
		Title               string `json:"title"`
		HelpText            string `json:"help_text"`
		ParticipantHelpText string `json:"participant_help_text"`
		ResearcherHelpText  string `json:"researcher_help_text"`
		IsNew               bool   `json:"is_new"`
	} `json:"query,omitempty"`
	Cls             string      `json:"_cls"`
	Category        string      `json:"category"`
	Subcategory     interface{} `json:"subcategory"`
	Order           int         `json:"order"`
	Recommended     bool        `json:"recommended"`
	DetailsDisplay  string      `json:"details_display"`
	RequirementType string      `json:"requirement_type"`
}

// FilterValue will help the bubbletea views run
func (r Requirement) FilterValue() string {
	title := r.Query.Question
	if title == "" {
		title = r.Query.Title
	}
	return title
}

// Title will set the main string for the view.
func (r Requirement) Title() string {
	title := r.Query.Question
	if title == "" {
		title = r.Query.Title
	}
	return title
}

// Description will set the secondary string the view.
func (r Requirement) Description() string {
	desc := fmt.Sprintf("Category: %s", r.Category)

	if r.Query.Description != "" {
		desc += fmt.Sprintf(". %s", r.Query.Description)
	}
	return desc
}

// Hook represents a subscription to an event
type Hook struct {
	ID          string `json:"id"`
	EventType   string `json:"event_type"`
	TargetURL   string `json:"target_url"`
	IsEnabled   bool   `json:"is_enabled"`
	WorkspaceID string `json:"workspace_id"`
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
	Value string `json:"value"`
}
