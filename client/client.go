// Package client provides an HTTP client for interacting with the Prolific API.
// It handles authentication, request execution, and response parsing for all
// supported Prolific resources including studies, workspaces, projects, hooks,
// messages, and AI Task Builder operations.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/prolific-oss/cli/config"
	"github.com/prolific-oss/cli/model"
	"github.com/spf13/viper"
	"golang.org/x/exp/slices"
)

// DefaultRecordOffset defines how many records we should offset to start with.
const DefaultRecordOffset = 0

// DefaultRecordLimit defines how many records to return by default.
const DefaultRecordLimit = 200

// API represents what is allowed to be called on the Prolific client.
type API interface {
	GetMe() (*MeResponse, error)

	CreateStudy(model.CreateStudy) (*model.Study, error)
	DuplicateStudy(ID string) (*model.Study, error)
	GetEligibilityRequirements() (*ListRequirementsResponse, error)
	GetStudies(status, projectID string) (*ListStudiesResponse, error)
	GetStudy(ID string) (*model.Study, error)
	GetSubmissions(ID string, limit, offset int) (*ListSubmissionsResponse, error)
	TransitionStudy(ID, action string) (*TransitionStudyResponse, error)
	UpdateStudy(ID string, study model.UpdateStudy) (*model.Study, error)
	GetStudyCredentialsUsageReportCSV(ID string) (string, error)
	CreateCredentialPool(credentials string, workspaceID string) (*CredentialPoolResponse, error)
	UpdateCredentialPool(credentialPoolID string, credentials string, workspaceID string) (*CredentialPoolResponse, error)
	ListCredentialPools(workspaceID string) (*ListCredentialPoolsResponse, error)

	GetCampaigns(workspaceID string, limit, offset int) (*ListCampaignsResponse, error)

	GetCollections(workspaceID string, limit, offset int) (*ListCollectionsResponse, error)
	GetCollection(ID string) (*model.Collection, error)
	UpdateCollection(ID string, collection model.UpdateCollection) (*model.Collection, error)

	GetHooks(workspaceID string, enabled bool, limit, offset int) (*ListHooksResponse, error)
	GetHookEventTypes() (*ListHookEventTypesResponse, error)
	GetHookSecrets(workspaceID string) (*ListSecretsResponse, error)
	GetEvents(subscriptionID string, limit, offset int) (*ListHookEventsResponse, error)

	GetWorkspaces(limit, offset int) (*ListWorkspacesResponse, error)
	CreateWorkspace(workspace model.Workspace) (*CreateWorkspacesResponse, error)

	GetProjects(workspaceID string, limit, offset int) (*ListProjectsResponse, error)
	CreateProject(workspaceID string, project model.Project) (*CreateProjectResponse, error)
	GetProject(ID string) (*model.Project, error)

	GetParticipantGroups(workspaceID string, limit, offset int) (*ListParticipantGroupsResponse, error)
	GetParticipantGroup(groupID string) (*ViewParticipantGroupResponse, error)

	GetFilters() (*ListFiltersResponse, error)

	GetFilterSets(workspaceID string, limit, offset int) (*ListFilterSetsResponse, error)
	GetFilterSet(ID string) (*model.FilterSet, error)

	GetRewardRecommendations(currency string, estimatedTime int, filterIDs []string) (*RewardRecommendationsResponse, error)

	GetMessages(userID *string, createdAfter *string) (*ListMessagesResponse, error)
	SendMessage(body, recipientID, studyID string) error
	GetUnreadMessages() (*ListUnreadMessagesResponse, error)
	BulkSendMessage(ids []string, body, studyID string) error
	SendGroupMessage(participantGroupID, body string, studyID *string) error

	CreateBonusPayments(payload CreateBonusPaymentsPayload) (*CreateBonusPaymentsResponse, error)
	PayBonusPayments(id string) error

	CreateAITaskBuilderBatch(params CreateBatchParams) (*CreateAITaskBuilderBatchResponse, error)
	CreateAITaskBuilderInstructions(batchID string, instructions CreateAITaskBuilderInstructionsPayload) (*CreateAITaskBuilderInstructionsResponse, error)
	SetupAITaskBuilderBatch(batchID, datasetID string, tasksPerGroup int) (*SetupAITaskBuilderBatchResponse, error)
	CreateAITaskBuilderDataset(workspaceID string, payload CreateAITaskBuilderDatasetPayload) (*CreateAITaskBuilderDatasetResponse, error)
	CreateAITaskBuilderCollection(payload model.CreateAITaskBuilderCollection) (*CreateAITaskBuilderCollectionResponse, error)
	GetAITaskBuilderBatch(batchID string) (*GetAITaskBuilderBatchResponse, error)
	GetAITaskBuilderBatchStatus(batchID string) (*GetAITaskBuilderBatchStatusResponse, error)
	GetAITaskBuilderBatches(workspaceID string) (*GetAITaskBuilderBatchesResponse, error)
	GetAITaskBuilderResponses(batchID string) (*GetAITaskBuilderResponsesResponse, error)
	GetAITaskBuilderTasks(batchID string) (*GetAITaskBuilderTasksResponse, error)
	GetAITaskBuilderDatasetStatus(datasetID string) (*GetAITaskBuilderDatasetStatusResponse, error)
	GetAITaskBuilderDatasetUploadURL(datasetID, fileName string) (*GetAITaskBuilderDatasetUploadURLResponse, error)
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
	viper.SetDefault("PROLIFIC_URL", config.GetAPIURL())

	client := Client{
		Client:  http.DefaultClient,
		Token:   viper.GetString("PROLIFIC_TOKEN"),
		BaseURL: strings.TrimRight(viper.GetString("PROLIFIC_URL"), "/"),
		Debug:   viper.GetBool("PROLIFIC_DEBUG"),
	}

	return client
}

// Execute runs an HTTP request.
func (c *Client) Execute(method, url string, body any, response any) (*http.Response, error) {
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
	request.Header.Set("User-Agent", "prolific-oss/cli")
	request.Header.Set("Authorization", fmt.Sprintf("Token %s", c.Token))

	if c.Debug {
		fmt.Println(request)
	}

	httpResponse, err := c.Client.Do(request) //nolint:gosec // G704: URL constructed from user-configured API base
	if err != nil {
		return nil, err
	}
	defer httpResponse.Body.Close()

	responseBody, _ := io.ReadAll(httpResponse.Body)
	httpResponse.Body = io.NopCloser(bytes.NewBuffer(responseBody))

	if c.Debug {
		fmt.Println(string(responseBody))
	}

	if httpResponse.StatusCode >= 400 {
		// Try the nested error format first
		var apiError JSONAPIError
		if err := json.NewDecoder(io.NopCloser(bytes.NewBuffer(responseBody))).Decode(&apiError); err == nil && apiError.Error.Detail != nil {
			return nil, fmt.Errorf("request failed: %v", apiError.Error.Detail)
		}

		// Try the simple error format
		var simpleError SimpleAPIError
		if err := json.NewDecoder(io.NopCloser(bytes.NewBuffer(responseBody))).Decode(&simpleError); err == nil && simpleError.Detail != "" {
			return nil, fmt.Errorf("request failed: %s - %s", simpleError.Message, simpleError.Detail)
		}

		// If both fail, return generic error with status code
		return nil, fmt.Errorf("request failed with status %d: %s", httpResponse.StatusCode, string(responseBody))
	}

	if response != nil {
		if err := json.NewDecoder(io.NopCloser(bytes.NewBuffer(responseBody))).Decode(response); err != nil {
			return nil, fmt.Errorf("decoding JSON response from %s failed: %v", request.URL, err)
		}
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

	// Validate status if it's not "all" or empty
	if status != "" && status != model.StatusAll {
		if !slices.Contains(model.StudyListStatus, status) {
			return nil, fmt.Errorf("%s is not a valid status: %s", status, strings.Join(model.StudyListStatus, ", "))
		}
	}

	// Build status fragment if status filtering is needed
	statusFragment := ""
	if status != "" && status != model.StatusAll {
		if status == model.StatusUnpublished {
			statusFragment = "published=0"
		} else {
			statusFragment = fmt.Sprintf("%s=1", status)
		}
	}

	// Build URL based on whether projectID is provided
	if projectID != "" {
		if statusFragment != "" {
			url = fmt.Sprintf("/api/v1/projects/%s/studies/?%s", projectID, statusFragment)
		} else {
			url = fmt.Sprintf("/api/v1/projects/%s/studies/", projectID)
		}
	} else {
		if statusFragment != "" {
			url = fmt.Sprintf("/api/v1/studies/?%s", statusFragment)
		} else {
			url = "/api/v1/studies/"
		}
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
	httpResponse, err := c.Execute(http.MethodGet, url, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	if httpResponse.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResponse.Body)
		return nil, fmt.Errorf("unable to get study: %v", string(body))
	}

	return &response, nil
}

// GetSubmissions will return submission data for a given study.
func (c *Client) GetSubmissions(ID string, limit, offset int) (*ListSubmissionsResponse, error) {
	var response ListSubmissionsResponse

	url := fmt.Sprintf("/api/v1/studies/%s/submissions/?limit=%v&offset=%v", ID, limit, offset)
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

	transition := struct {
		Action string `json:"action"`
	}{
		Action: action,
	}

	url := fmt.Sprintf("/api/v1/studies/%s/transition/", ID)
	_, err := c.Execute(http.MethodPost, url, transition, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to transition study to %s: %v", action, err)
	}

	return &response, nil
}

// GetCampaigns will return you a list of Campaign objects.
func (c *Client) GetCampaigns(workspaceID string, limit, offset int) (*ListCampaignsResponse, error) {
	var response ListCampaignsResponse

	url := fmt.Sprintf("/api/v1/campaigns/?workspace_id=%s&limit=%v&offset=%v", workspaceID, limit, offset)
	_, err := c.Execute(http.MethodGet, url, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	return &response, nil
}

// GetCollections will return a list of Collection objects for a workspace.
func (c *Client) GetCollections(workspaceID string, limit, offset int) (*ListCollectionsResponse, error) {
	var response ListCollectionsResponse

	url := fmt.Sprintf("/api/v1/data-collection/collections?workspace_id=%s&limit=%v&offset=%v", workspaceID, limit, offset)
	_, err := c.Execute(http.MethodGet, url, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	return &response, nil
}

// GetCollection will return a single Collection by ID.
func (c *Client) GetCollection(ID string) (*model.Collection, error) {
	var response model.Collection

	url := fmt.Sprintf("/api/v1/data-collection/collections/%s", ID)
	httpResponse, err := c.Execute(http.MethodGet, url, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	if httpResponse.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResponse.Body)
		return nil, fmt.Errorf("unable to get collection: %v", string(body))
	}

	return &response, nil
}

// UpdateStudy is responsible for updating the Study with a PATCH request.
func (c *Client) UpdateStudy(ID string, study model.UpdateStudy) (*model.Study, error) {
	var response model.Study

	url := fmt.Sprintf("/api/v1/studies/%s/", ID)
	httpResponse, err := c.Execute(http.MethodPatch, url, study, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to update study: %v", err)
	}

	if httpResponse.StatusCode != http.StatusOK {
		return nil, errors.New(`unable to update study`)
	}

	return &response, nil
}

// GetStudyCredentialsUsageReportCSV will return the credentials usage report for a study as CSV.
func (c *Client) GetStudyCredentialsUsageReportCSV(ID string) (string, error) {
	endpointURL := fmt.Sprintf("/api/v1/studies/%s/credentials/report/", ID)
	httpResponse, err := c.Execute(http.MethodGet, endpointURL, nil, nil)
	if err != nil {
		return "", err
	}

	responseBody, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return "", fmt.Errorf("unable to read response body: %w", err)
	}

	return string(responseBody), nil
}

// GetHooks will return the subscriptions to event types for current user.
func (c *Client) GetHooks(workspaceID string, enabled bool, limit, offset int) (*ListHooksResponse, error) {
	var response ListHooksResponse

	url := fmt.Sprintf(
		"/api/v1/hooks/subscriptions?workspace_id=%s&is_enabled=%v&limit=%v&offset=%v",
		workspaceID,
		enabled,
		limit,
		offset,
	)
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
func (c *Client) GetEvents(subscriptionID string, limit, offset int) (*ListHookEventsResponse, error) {
	var response ListHookEventsResponse

	url := fmt.Sprintf("/api/v1/hooks/subscriptions/%s/events/?limit=%v&offset=%v", subscriptionID, limit, offset)
	_, err := c.Execute(http.MethodGet, url, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	return &response, nil
}

// GetWorkspaces will return you the workspaces you can see
func (c *Client) GetWorkspaces(limit, offset int) (*ListWorkspacesResponse, error) {
	var response ListWorkspacesResponse

	url := fmt.Sprintf("/api/v1/workspaces/?limit=%v&offset=%v", limit, offset)
	httpResponse, err := c.Execute(http.MethodGet, url, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	if httpResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code was %v, so therefore unable to get workspaces", httpResponse.StatusCode)
	}

	return &response, nil
}

// CreateWorkspace will create you a workspace
func (c *Client) CreateWorkspace(workspace model.Workspace) (*CreateWorkspacesResponse, error) {
	var response CreateWorkspacesResponse

	url := "/api/v1/workspaces/"
	httpResponse, err := c.Execute(http.MethodPost, url, workspace, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	if httpResponse.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(httpResponse.Body)
		return nil, fmt.Errorf("unable to create workspace: %v", string(body))
	}

	return &response, nil
}

// GetProjects will return the projects for the given workspace ID
func (c *Client) GetProjects(workspaceID string, limit, offset int) (*ListProjectsResponse, error) {
	var response ListProjectsResponse

	url := fmt.Sprintf("/api/v1/workspaces/%s/projects/?limit=%v&offset=%v", workspaceID, limit, offset)
	_, err := c.Execute(http.MethodGet, url, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	return &response, nil
}

// GetProject will return the project for the given project ID
func (c *Client) GetProject(ID string) (*model.Project, error) {
	var response model.Project

	url := fmt.Sprintf("/api/v1/projects/%s/", ID)
	httpResponse, err := c.Execute(http.MethodGet, url, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	if httpResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code was %v, so therefore unable to get project: %v", httpResponse.StatusCode, ID)
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

// GetParticipantGroups will return all the participant groups you have access to for a given WorkspaceID
func (c *Client) GetParticipantGroups(workspaceID string, limit, offset int) (*ListParticipantGroupsResponse, error) {
	var response ListParticipantGroupsResponse

	url := fmt.Sprintf("/api/v1/participant-groups/?workspace_id=%s&limit=%v&offset=%v", workspaceID, limit, offset)
	_, err := c.Execute(http.MethodGet, url, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	return &response, nil
}

// GetParticipantGroup will return the membership in the group
func (c *Client) GetParticipantGroup(groupID string) (*ViewParticipantGroupResponse, error) {
	var response ViewParticipantGroupResponse

	url := fmt.Sprintf("/api/v1/participant-groups/%s/participants/", groupID)
	_, err := c.Execute(http.MethodGet, url, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	return &response, nil
}

func (c *Client) GetFilters() (*ListFiltersResponse, error) {
	var response ListFiltersResponse

	url := "/api/v1/filters/"
	_, err := c.Execute(http.MethodGet, url, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	return &response, nil
}

// GetFilterSets will return the filter sets in a workspace
func (c *Client) GetFilterSets(workspaceID string, limit, offset int) (*ListFilterSetsResponse, error) {
	var response ListFilterSetsResponse

	url := fmt.Sprintf("/api/v1/filter-sets/?workspace_id=%s&limit=%v&offset=%v", workspaceID, limit, offset)
	_, err := c.Execute(http.MethodGet, url, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	return &response, nil
}

// GetFilterSet will return the filter set for the given filter set ID
func (c *Client) GetFilterSet(ID string) (*model.FilterSet, error) {
	var response model.FilterSet

	url := fmt.Sprintf("/api/v1/filter-sets/%s/", ID)
	httpResponse, err := c.Execute(http.MethodGet, url, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	if httpResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code was %v, so therefore unable to get filter set: %v", httpResponse.StatusCode, ID)
	}

	return &response, nil
}

// GetRewardRecommendations will return recommended reward rates for participants
func (c *Client) GetRewardRecommendations(currency string, estimatedTime int, filterIDs []string) (*RewardRecommendationsResponse, error) {
	var response RewardRecommendationsResponse

	params := url.Values{}
	params.Add("currency_code", currency)
	params.Add("estimated_time", fmt.Sprintf("%d", estimatedTime))

	if len(filterIDs) > 0 {
		params.Add("filter_ids", strings.Join(filterIDs, ","))
	}

	url := fmt.Sprintf("/api/v1/studies/reward-recommendations/?%s", params.Encode())
	_, err := c.Execute(http.MethodGet, url, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	return &response, nil
}

// UpdateCollection will update a collection with the given ID
func (c *Client) UpdateCollection(ID string, collection model.UpdateCollection) (*model.Collection, error) {
	var response model.Collection

	url := fmt.Sprintf("/api/v1/data-collection/collections/%s/", ID)
	_, err := c.Execute(http.MethodPut, url, collection, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	return &response, nil
}

// GetMessages will return the messages for the authenticated user
func (c *Client) GetMessages(userID *string, createdAfter *string) (*ListMessagesResponse, error) {
	var response ListMessagesResponse

	if userID == nil && createdAfter == nil {
		return nil, fmt.Errorf("either userID or createdAfter must be provided")
	}

	baseURL := "/api/v1/messages/"
	params := url.Values{}

	if userID != nil {
		params.Add("user_id", *userID)
	}

	if createdAfter != nil {
		params.Add("created_after", *createdAfter)
	}

	url := baseURL + "?" + params.Encode()

	_, err := c.Execute(http.MethodGet, url, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	return &response, nil
}

// SendMessage will send a message
func (c *Client) SendMessage(body string, recipientID string, studyID string) error {
	payload := SendMessagePayload{
		RecipientID: recipientID,
		StudyID:     studyID,
		Body:        body,
	}

	url := "/api/v1/messages/"
	_, err := c.Execute(http.MethodPost, url, payload, nil)
	if err != nil {
		return fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	return nil
}

// GetUnreadMessages will return the unread messages for the authenticated user
func (c *Client) GetUnreadMessages() (*ListUnreadMessagesResponse, error) {
	var response ListUnreadMessagesResponse

	url := "/api/v1/messages/unread/"

	_, err := c.Execute(http.MethodGet, url, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	return &response, nil
}

// BulkSendMessage will send a message to multiple participants
func (c *Client) BulkSendMessage(ids []string, body, studyID string) error {
	payload := BulkSendMessagePayload{
		IDs:     ids,
		Body:    body,
		StudyID: studyID,
	}

	url := "/api/v1/messages/bulk/"
	_, err := c.Execute(http.MethodPost, url, payload, nil)
	if err != nil {
		return fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	return nil
}

// SendGroupMessage will send a message to a participant group
func (c *Client) SendGroupMessage(participantGroupID, body string, studyID *string) error {
	payload := SendGroupMessagePayload{
		ParticipantGroupID: participantGroupID,
		Body:               body,
	}

	if studyID != nil {
		payload.StudyID = *studyID
	}

	url := "/api/v1/messages/participant-group/"
	_, err := c.Execute(http.MethodPost, url, payload, nil)
	if err != nil {
		return fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	return nil
}

// CreateBonusPayments creates bonus payment records for participants in a study.
func (c *Client) CreateBonusPayments(payload CreateBonusPaymentsPayload) (*CreateBonusPaymentsResponse, error) {
	var response CreateBonusPaymentsResponse

	url := "/api/v1/submissions/bonus-payments/"
	httpResponse, err := c.Execute(http.MethodPost, url, payload, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	if httpResponse.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(httpResponse.Body)
		return nil, fmt.Errorf("unable to create bonus payments: %v", string(body))
	}

	return &response, nil
}

// PayBonusPayments triggers asynchronous payment of previously created bonus records.
func (c *Client) PayBonusPayments(id string) error {
	url := fmt.Sprintf("/api/v1/bulk-bonus-payments/%s/pay/", id)
	httpResponse, err := c.Execute(http.MethodPost, url, nil, nil)
	if err != nil {
		return fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	if httpResponse.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(httpResponse.Body)
		return fmt.Errorf("unable to pay bonus payments: %v", string(body))
	}

	return nil
}

// GetAITaskBuilderBatch will return details of an AI Task Builder batch.
func (c *Client) GetAITaskBuilderBatch(batchID string) (*GetAITaskBuilderBatchResponse, error) {
	var response GetAITaskBuilderBatchResponse

	url := fmt.Sprintf("/api/v1/data-collection/batches/%s", batchID)
	_, err := c.Execute(http.MethodGet, url, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	return &response, nil
}

// GetAITaskBuilderBatchStatus will return the status of an AI Task Builder batch.
func (c *Client) GetAITaskBuilderBatchStatus(batchID string) (*GetAITaskBuilderBatchStatusResponse, error) {
	var response GetAITaskBuilderBatchStatusResponse

	url := fmt.Sprintf("/api/v1/data-collection/batches/%s/status", batchID)
	_, err := c.Execute(http.MethodGet, url, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}
	return &response, nil
}

// GetAITaskBuilderBatches will return the batches for a given workspace.
func (c *Client) GetAITaskBuilderBatches(workspaceID string) (*GetAITaskBuilderBatchesResponse, error) {
	var response GetAITaskBuilderBatchesResponse

	url := fmt.Sprintf("/api/v1/data-collection/batches/?workspace_id=%s", workspaceID)
	_, err := c.Execute(http.MethodGet, url, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}
	return &response, nil
}

// GetAITaskBuilderResponses will return the responses for an AI Task Builder batch.
func (c *Client) GetAITaskBuilderResponses(batchID string) (*GetAITaskBuilderResponsesResponse, error) {
	var response GetAITaskBuilderResponsesResponse

	url := fmt.Sprintf("/api/v1/data-collection/batches/%s/responses", batchID)
	_, err := c.Execute(http.MethodGet, url, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}
	return &response, nil
}

// GetAITaskBuilderTasks will return the tasks for an AI Task Builder batch.
func (c *Client) GetAITaskBuilderTasks(batchID string) (*GetAITaskBuilderTasksResponse, error) {
	var response GetAITaskBuilderTasksResponse

	url := fmt.Sprintf("/api/v1/data-collection/batches/%s/tasks", batchID)
	_, err := c.Execute(http.MethodGet, url, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}
	return &response, nil
}

// GetAITaskBuilderDatasetStatus will return the status of an AI Task Builder dataset.
func (c *Client) GetAITaskBuilderDatasetStatus(datasetID string) (*GetAITaskBuilderDatasetStatusResponse, error) {
	var response GetAITaskBuilderDatasetStatusResponse

	url := fmt.Sprintf("/api/v1/data-collection/datasets/%s/status", datasetID)
	_, err := c.Execute(http.MethodGet, url, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}
	return &response, nil
}

// GetAITaskBuilderDatasetUploadURL will get an upload URL for an AI Task Builder dataset.
func (c *Client) GetAITaskBuilderDatasetUploadURL(datasetID, fileName string) (*GetAITaskBuilderDatasetUploadURLResponse, error) {
	var response GetAITaskBuilderDatasetUploadURLResponse

	url := fmt.Sprintf("/api/v1/data-collection/datasets/%s/upload-url/%s.csv/", datasetID, fileName)
	_, err := c.Execute(http.MethodGet, url, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}
	return &response, nil
}

// CreateAITaskBuilderBatch will create an AI Task Builder batch.
func (c *Client) CreateAITaskBuilderBatch(params CreateBatchParams) (*CreateAITaskBuilderBatchResponse, error) {
	var response CreateAITaskBuilderBatchResponse

	payload := CreateAITaskBuilderBatchPayload{
		Name:        params.Name,
		WorkspaceID: params.WorkspaceID,
		DatasetID:   params.DatasetID,
		TaskDetails: TaskDetails{
			TaskName:         params.TaskName,
			TaskIntroduction: params.TaskIntroduction,
			TaskSteps:        params.TaskSteps,
		},
	}

	url := "/api/v1/data-collection/batches"
	httpResponse, err := c.Execute(http.MethodPost, url, payload, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	if httpResponse.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(httpResponse.Body)
		return nil, fmt.Errorf("unable to create batch: %v", string(body))
	}

	return &response, nil
}

// CreateAITaskBuilderInstructions will create instructions for an AI Task Builder batch.
func (c *Client) CreateAITaskBuilderInstructions(batchID string, instructions CreateAITaskBuilderInstructionsPayload) (*CreateAITaskBuilderInstructionsResponse, error) {
	var response CreateAITaskBuilderInstructionsResponse

	url := fmt.Sprintf("/api/v1/data-collection/batches/%s/instructions", batchID)
	httpResponse, err := c.Execute(http.MethodPost, url, instructions, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	if httpResponse.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(httpResponse.Body)
		return nil, fmt.Errorf("unable to create instructions: %v", string(body))
	}

	return &response, nil
}

// SetupAITaskBuilderBatch will setup an AI Task Builder batch.
func (c *Client) SetupAITaskBuilderBatch(batchID, datasetID string, tasksPerGroup int) (*SetupAITaskBuilderBatchResponse, error) {
	var response SetupAITaskBuilderBatchResponse

	payload := SetupAITaskBuilderBatchPayload{
		DatasetID:     datasetID,
		TasksPerGroup: tasksPerGroup,
	}

	url := fmt.Sprintf("/api/v1/data-collection/batches/%s/setup", batchID)
	httpResponse, err := c.Execute(http.MethodPost, url, payload, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	// Check for 202 Accepted status
	if httpResponse.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("unexpected status code: %d", httpResponse.StatusCode)
	}

	return &response, nil
}

// CreateAITaskBuilderDataset will create a new AI Task Builder dataset.
// The workspaceID parameter specifies which workspace the dataset belongs to.
func (c *Client) CreateAITaskBuilderDataset(workspaceID string, payload CreateAITaskBuilderDatasetPayload) (*CreateAITaskBuilderDatasetResponse, error) {
	var response CreateAITaskBuilderDatasetResponse

	// Ensure workspace_id in payload matches the parameter
	payload.WorkspaceID = workspaceID

	url := "/api/v1/data-collection/datasets"
	httpResponse, err := c.Execute(http.MethodPost, url, payload, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	if httpResponse.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(httpResponse.Body)
		return nil, fmt.Errorf("unable to create dataset: %v", string(body))
	}

	return &response, nil
}

// CreateAITaskBuilderCollection creates a new AI Task Builder collection.
func (c *Client) CreateAITaskBuilderCollection(payload model.CreateAITaskBuilderCollection) (*CreateAITaskBuilderCollectionResponse, error) {
	var response CreateAITaskBuilderCollectionResponse

	url := "/api/v1/data-collection/collections"
	httpResponse, err := c.Execute(http.MethodPost, url, payload, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	if httpResponse.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(httpResponse.Body)
		return nil, fmt.Errorf("unable to create collection: %v", string(body))
	}

	return &response, nil
}

// CreateCredentialPool creates a new credential pool with the provided credentials.
// credentials should be a comma-separated string with newlines between entries.
func (c *Client) CreateCredentialPool(credentials string, workspaceID string) (*CredentialPoolResponse, error) {
	var response CredentialPoolResponse

	payload := CredentialPoolPayload{
		Credentials: credentials,
		WorkspaceID: workspaceID,
	}

	endpointURL := "/api/v1/credentials/"
	httpResponse, err := c.Execute(http.MethodPost, endpointURL, payload, &response)
	if err != nil {
		return nil, err
	}

	if httpResponse.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code: expected 201, got %d", httpResponse.StatusCode)
	}

	return &response, nil
}

// UpdateCredentialPool updates an existing credential pool with new credentials.
// credentials should be a comma-separated string with newlines between entries.
func (c *Client) UpdateCredentialPool(credentialPoolID string, credentials string, workspaceID string) (*CredentialPoolResponse, error) {
	var response CredentialPoolResponse

	payload := UpdateCredentialPoolPayload{
		Credentials: credentials,
	}

	endpointURL := fmt.Sprintf("/api/v1/credentials/%s/", credentialPoolID)
	httpResponse, err := c.Execute(http.MethodPatch, endpointURL, payload, &response)
	if err != nil {
		return nil, err
	}

	if httpResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: expected 200, got %d", httpResponse.StatusCode)
	}

	return &response, nil
}

// ListCredentialPools retrieves a list of credential pools for a specific workspace.
func (c *Client) ListCredentialPools(workspaceID string) (*ListCredentialPoolsResponse, error) {
	var response ListCredentialPoolsResponse

	endpointURL := fmt.Sprintf("/api/v1/credentials/?workspace_id=%s", workspaceID)
	httpResponse, err := c.Execute(http.MethodGet, endpointURL, nil, &response)
	if err != nil {
		return nil, err
	}

	if httpResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: expected 200, got %d", httpResponse.StatusCode)
	}

	return &response, nil
}
