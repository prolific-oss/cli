package client

import (
	"github.com/benmatselby/prolificli/model"
)

// JSONAPILinks is the standard pagination data structure.
type JSONAPILinks struct {
	Links struct {
		Self struct {
			Href  string `json:"href"`
			Title string `json:"title"`
		} `json:"self"`
		Next struct {
			Href  interface{} `json:"href"`
			Title string      `json:"title"`
		} `json:"next"`
		Previous struct {
			Href  interface{} `json:"href"`
			Title string      `json:"title"`
		} `json:"previous"`
		Last struct {
			Href  string `json:"href"`
			Title string `json:"title"`
		} `json:"last"`
	} `json:"_links"`
}

// JSONAPIMeta is the standard meta data structure.
type JSONAPIMeta struct {
	Meta struct {
		Count int `json:"count"`
	} `json:"meta"`
}

// MeResponse is a struct that represents your account.
type MeResponse struct {
	ID               string `json:"id"`
	Email            string `json:"email"`
	FirstName        string `json:"first_name"`
	LastName         string `json:"last_name"`
	Name             string `json:"name"`
	Username         string `json:"username"`
	UserType         string `json:"user_type"`
	CurrencyCode     string `json:"currency_code"`
	Balance          int    `json:"balance"`
	AvailableBalance int    `json:"available_balance"`
}

// ListStudiesResponse is the response for the /studies API response.
type ListStudiesResponse struct {
	Results []model.Study `json:"results"`
	*JSONAPILinks
	*JSONAPIMeta
}

// ListSubmissionsResponse is the response for the submissions request.
type ListSubmissionsResponse struct {
	Results []model.Submission `json:"results"`
	*JSONAPILinks
	*JSONAPIMeta
}

// ListRequirementsResponse is the response for the requirements request.
type ListRequirementsResponse struct {
	Results []model.Requirement `json:"results"`
	*JSONAPILinks
	*JSONAPIMeta
}

// TransitionStudyResponse is the response for transitioning a study to another status.
type TransitionStudyResponse struct {
	ID                      string        `json:"id"`
	Name                    string        `json:"name"`
	InternalName            string        `json:"internal_name"`
	Description             string        `json:"description"`
	ExternalStudyURL        string        `json:"external_study_url"`
	ProlificIDOption        string        `json:"prolific_id_option"`
	CompletionCode          string        `json:"completion_code"`
	CompletionOption        string        `json:"completion_option"`
	TotalAvailablePlaces    int           `json:"total_available_places"`
	EstimatedCompletionTime int           `json:"estimated_completion_time"`
	MaximumAllowedTime      int           `json:"maximum_allowed_time"`
	Reward                  int           `json:"reward"`
	DeviceCompatibility     []string      `json:"device_compatibility"`
	PeripheralRequirements  []interface{} `json:"peripheral_requirements"`
	EligibilityRequirements []interface{} `json:"eligibility_requirements"`
	Status                  string        `json:"status"`
}

// ListHooksResponse is the response for the hook subscriptions.
type ListHooksResponse struct {
	Results []model.Hook `json:"results"`
}

// ListHookEventTypesResponse is the response for the event types hook API.
type ListHookEventTypesResponse struct {
	Results []model.HookEventType `json:"results"`
}

// ListHookEventsResponse is the response for the hook events API.
type ListHookEventsResponse struct {
	Results []model.HookEvent `json:"results"`
}

// ListWorkspacesResponse is the response for the list workspaces endpoint.
type ListWorkspacesResponse struct {
	Results []model.Workspace `json:"results"`
}

// CreateWorkspacesResponse is the response for creating a workspace.
type CreateWorkspacesResponse struct {
	model.Workspace
}

// ListProjectsResponse is the response for the list projects endpoint.
type ListProjectsResponse struct {
	Results []model.Project `json:"results"`
}

// CreateProjectResponse is the response for creating a project.
type CreateProjectResponse struct {
	model.Project
}

// ListSecretsResponse is the list secrets response.
type ListSecretsResponse struct {
	Results []model.Secret `json:"results"`
}

// ListParticipantGroupsResponse is the list of participant groups response.
type ListParticipantGroupsResponse struct {
	Results []model.ParticipantGroup `json:"results"`
}

// ViewParticipantGroupResponse is the list of members in a group.
type ViewParticipantGroupResponse struct {
	Results []model.ParticipantGroupMembership `json:"results"`
}
