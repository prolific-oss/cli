package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/benmatselby/prolificli/model"
	"github.com/spf13/viper"
	"golang.org/x/exp/slices"
)

// API represents what is allowed to be called on the Prolific client.
type API interface {
	GetMe() (*MeResponse, error)

	CreateStudy(model.CreateStudy) (*model.Study, error)
	DuplicateStudy(ID string) (*model.Study, error)
	GetEligibilityRequirements() (*ListRequirementsResponse, error)
	GetStudies(status, projectID string) (*ListStudiesResponse, error)
	GetStudy(ID string) (*model.Study, error)
	GetSubmissions(ID string) (*ListSubmissionsResponse, error)
	TransitionStudy(ID, action string) (*TransitionStudyResponse, error)

	GetHooks(enabled bool) (*ListHooksResponse, error)
	GetHookEventTypes() (*ListHookEventTypesResponse, error)
	GetHookSecrets(workspaceID string) (*ListSecretsResponse, error)
	GetEvents(subscriptionID string) (*ListHookEventsResponse, error)

	GetWorkspaces() (*ListWorkspacesResponse, error)
	CreateWorkspace(workspace model.Workspace) (*CreateWorkspacesResponse, error)

	GetProjects(workspaceID string) (*ListProjectsResponse, error)
	CreateProject(workspaceID string, project model.Project) (*CreateProjectResponse, error)

	GetParticipantGroups(projectID string) (*ListParticipantGroupsResponse, error)
}

// Client is responsible for interacting with the Prolific API.
type Client struct {
	Client  *http.Client
	BaseURL string
	Token   string
	Debug   bool
}

// New will return a new Prolific client.
func New() Client {
	viper.SetDefault("PROLIFIC_URL", "https://api.prolific.co")

	client := Client{
		Client:  http.DefaultClient,
		Token:   viper.GetString("PROLIFIC_TOKEN"),
		BaseURL: strings.TrimRight(viper.GetString("PROLIFIC_URL"), "/"),
		Debug:   viper.GetBool("PROLIFIC_DEBUG"),
	}

	return client
}

// Execute runs an HTTP request.
func (c *Client) Execute(method, url string, body interface{}, response interface{}) (*http.Response, error) {
	if c.Token == "" {
		return nil, errors.New("PROLIFIC_TOKEN not set")
	}

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		err := enc.Encode(body)
		if err != nil {
			return nil, err
		}
	}

	request, err := http.NewRequestWithContext(context.Background(), method, c.BaseURL+url, buf)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "benmatselby/prolificli")
	request.Header.Set("Authorization", fmt.Sprintf("Token %s", c.Token))

	httpResponse, err := c.Client.Do(request)
	if err != nil {
		return nil, err
	}
	defer httpResponse.Body.Close()

	responseBody, _ := io.ReadAll(httpResponse.Body)
	httpResponse.Body = io.NopCloser(bytes.NewBuffer(responseBody))

	if c.Debug {
		fmt.Println(string(responseBody))
	}

	if err := json.NewDecoder(io.NopCloser(bytes.NewBuffer(responseBody))).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding JSON response from %s failed: %v", request.URL, err)
	}

	return httpResponse, nil
}

// CreateStudy is responsible for hitting the Prolific API to create a study.
func (c *Client) CreateStudy(study model.CreateStudy) (*model.Study, error) {
	var response model.Study

	url := "/api/v1/studies/"
	httpResponse, err := c.Execute(http.MethodPost, url, study, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	if httpResponse.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(httpResponse.Body)
		return nil, fmt.Errorf("unable to create study: %v", string(body))
	}

	return &response, nil
}

// GetMe will return your user account details.
func (c *Client) GetMe() (*MeResponse, error) {
	var response MeResponse

	url := "/api/v1/users/me"
	_, err := c.Execute(http.MethodGet, url, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	return &response, nil
}

// DuplicateStudy will duplicate an existing study.
func (c *Client) DuplicateStudy(ID string) (*model.Study, error) {
	var response model.Study

	url := fmt.Sprintf("/api/v1/studies/%s/clone/", ID)
	httpResponse, err := c.Execute(http.MethodPost, url, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	if httpResponse.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResponse.Body)
		return nil, fmt.Errorf("unable to duplicate study: %v", string(body))
	}

	return &response, nil
}

// GetStudies will return you a list of Study objects.
func (c *Client) GetStudies(status, projectID string) (*ListStudiesResponse, error) {
	var response ListStudiesResponse
	var url string

	if projectID != "" {
		url = fmt.Sprintf("/api/v1/projects/%s/studies/", projectID)
	} else {
		if !slices.Contains(model.StudyListStatus, status) {
			return nil, fmt.Errorf("%s is not a valid status: %s", status, strings.Join(model.StudyListStatus, ", "))
		}

		statusFragment := ""
		if status == model.StatusUnpublished {
			statusFragment = "published=0"
		} else {
			statusFragment = fmt.Sprintf("%s=1", status)
		}

		url = fmt.Sprintf("/api/v1/studies/?%s", statusFragment)
	}

	_, err := c.Execute(http.MethodGet, url, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	return &response, nil
}

// GetStudy will return a single study
func (c *Client) GetStudy(ID string) (*model.Study, error) {
	var response model.Study

	url := fmt.Sprintf("/api/v1/studies/%s", ID)
	_, err := c.Execute(http.MethodGet, url, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	return &response, nil
}

// GetSubmissions will return submission data for a given study.
func (c *Client) GetSubmissions(ID string) (*ListSubmissionsResponse, error) {
	var response ListSubmissionsResponse

	url := fmt.Sprintf("/api/v1/studies/%s/submissions/?offset=0&limit=200", ID)
	_, err := c.Execute(http.MethodGet, url, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	return &response, nil
}

// GetEligibilityRequirements will return requirement data.
func (c *Client) GetEligibilityRequirements() (*ListRequirementsResponse, error) {
	var response ListRequirementsResponse

	url := "/api/v1/eligibility-requirements/"
	_, err := c.Execute(http.MethodGet, url, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	return &response, nil
}

// TransitionStudy will move the study status to a desired state.
func (c *Client) TransitionStudy(ID, action string) (*TransitionStudyResponse, error) {
	var response TransitionStudyResponse

	transtion := struct {
		Action string `json:"action"`
	}{
		Action: action,
	}

	url := fmt.Sprintf("/api/v1/studies/%s/transition/", ID)
	_, err := c.Execute(http.MethodPost, url, transtion, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to transition study to %s: %v", action, err)
	}

	return &response, nil
}

// GetHooks will return the subscriptions to event types for current user.
func (c *Client) GetHooks(enabled bool) (*ListHooksResponse, error) {
	var response ListHooksResponse

	url := fmt.Sprintf("/api/v1/hooks/subscriptions?is_enabled=%v", enabled)
	_, err := c.Execute(http.MethodGet, url, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	return &response, nil
}

// GetHookEventTypes will return all of the event types you can subscribe a
// hook for.
func (c *Client) GetHookEventTypes() (*ListHookEventTypesResponse, error) {
	var response ListHookEventTypesResponse

	url := "/api/v1/hooks/event-types/"
	_, err := c.Execute(http.MethodGet, url, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	return &response, nil
}

// GetHookSecrets will return the secrets for a Workspace
func (c *Client) GetHookSecrets(workspaceID string) (*ListSecretsResponse, error) {
	var response ListSecretsResponse

	url := fmt.Sprintf("/api/v1/hooks/secrets/?workspace_id=%s", workspaceID)
	_, err := c.Execute(http.MethodGet, url, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	return &response, nil
}

// GetEvents will return events created for a subscription
func (c *Client) GetEvents(subscriptionID string) (*ListHookEventsResponse, error) {
	var response ListHookEventsResponse

	url := fmt.Sprintf("/api/v1/hooks/subscriptions/%s/events/", subscriptionID)
	_, err := c.Execute(http.MethodGet, url, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	return &response, nil
}

// GetWorkspaces will return you the workspaces you can see
func (c *Client) GetWorkspaces() (*ListWorkspacesResponse, error) {
	var response ListWorkspacesResponse

	url := "/api/v1/workspaces/"
	_, err := c.Execute(http.MethodGet, url, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	return &response, nil
}

// CreateWorkspace will create you a workspace
func (c *Client) CreateWorkspace(workspace model.Workspace) (*CreateWorkspacesResponse, error) {
	var response CreateWorkspacesResponse

	url := "/api/v1/workspaces/"
	_, err := c.Execute(http.MethodPost, url, workspace, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	return &response, nil
}

// GetProjects will return the projects for the given workspace ID
func (c *Client) GetProjects(workspaceID string) (*ListProjectsResponse, error) {
	var response ListProjectsResponse

	url := fmt.Sprintf("/api/v1/workspaces/%s/projects/", workspaceID)
	_, err := c.Execute(http.MethodGet, url, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	return &response, nil
}

// CreateProject will create you a project
func (c *Client) CreateProject(workspaceID string, project model.Project) (*CreateProjectResponse, error) {
	var response CreateProjectResponse

	url := fmt.Sprintf("/api/v1/workspaces/%s/projects/", workspaceID)
	_, err := c.Execute(http.MethodPost, url, project, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	return &response, nil
}

// GetParticipantGroups will return all the participant groups you have access to for a given ProjectID
func (c *Client) GetParticipantGroups(projectID string) (*ListParticipantGroupsResponse, error) {
	var response ListParticipantGroupsResponse

	url := fmt.Sprintf("/api/v1/participant-groups/?project_id=%s", projectID)
	_, err := c.Execute(http.MethodGet, url, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	return &response, nil
}
