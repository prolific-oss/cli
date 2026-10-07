package client

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/prolific-oss/cli/model"
	"golang.org/x/exp/slices"
)

// SearchStudies searches accessible studies by name, internal name, or study ID,
// paginating by page number rather than limit and offset.
//
// openapi.yaml declares only status and state for this operation, so search,
// page and workspace_id are undocumented: they are not covered by contract_test
// (see its SPECGAP note) and the page-number behaviour below is an observation
// about the live API, not something the spec states. Confirm against the API
// before relying on it, and prefer FetchNumberedPages over calling this
// directly so callers never depend on the pagination model.
func (c *Client) SearchStudies(query, workspaceID string, page int) (*ListStudiesResponse, error) {
	params := url.Values{"search": {query}, "page": {strconv.Itoa(page)}}
	if workspaceID != "" {
		params.Set("workspace_id", workspaceID)
	}
	var response ListStudiesResponse
	if _, err := c.ExecuteBuilder().GetInto("/api/v1/studies/?"+params.Encode(), &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// CreateStudy is responsible for hitting the Prolific API to create a study.
func (c *Client) CreateStudy(study model.CreateStudy) (*model.Study, error) {
	var response model.Study

	url := "/api/v1/studies/"
	_, err := c.ExecuteBuilder().
		PostRequest(url).
		Body(study).
		Status(http.StatusCreated).
		Decode(&response).
		Execute()
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// DuplicateStudy will duplicate an existing study.
func (c *Client) DuplicateStudy(ID string) (*model.Study, error) {
	var response model.Study

	url := fmt.Sprintf("/api/v1/studies/%s/clone/", ID)
	_, err := c.ExecuteBuilder().
		PostRequest(url).
		Status(http.StatusOK).
		Decode(&response).
		Execute()
	if err != nil {
		return nil, err
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

	_, err := c.ExecuteBuilder().GetInto(url, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// GetStudy will return a single study
func (c *Client) GetStudy(ID string) (*model.Study, error) {
	var response model.Study

	url := fmt.Sprintf("/api/v1/studies/%s", ID)
	_, err := c.ExecuteBuilder().Get(url, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// GetStudySubmissionCounts returns submission counts grouped by status for a study.
func (c *Client) GetStudySubmissionCounts(ID string) (*model.SubmissionCounts, error) {
	var response model.SubmissionCounts

	url := fmt.Sprintf("/api/v1/studies/%s/submissions/counts/", ID)
	_, err := c.ExecuteBuilder().Get(url, &response)
	if err != nil {
		return nil, err
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
	_, err := c.ExecuteBuilder().
		PostRequest(url).
		Body(transition).
		Decode(&response).
		Execute()
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// UpdateStudy is responsible for updating the Study with a PATCH request.
func (c *Client) UpdateStudy(ID string, study any) (*model.Study, error) {
	var response model.Study

	url := fmt.Sprintf("/api/v1/studies/%s/", ID)
	_, err := c.ExecuteBuilder().
		PatchRequest(url).
		Body(study).
		Status(http.StatusOK).
		Decode(&response).
		Execute()
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// GetStudyCredentialsUsageReportCSV will return the credentials usage report for a study as CSV.
func (c *Client) GetStudyCredentialsUsageReportCSV(ID string) (string, error) {
	endpointURL := fmt.Sprintf("/api/v1/studies/%s/credentials/report/", ID)
	httpResponse, err := c.ExecuteBuilder().GetRequest(endpointURL).Execute()
	if err != nil {
		return "", err
	}

	responseBody, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return "", fmt.Errorf("unable to read response body: %w", err)
	}

	return string(responseBody), nil
}

// TestStudy creates a test run of a study to validate configuration before going live.
func (c *Client) TestStudy(ID string) (*TestStudyResponse, error) {
	var response TestStudyResponse

	url := fmt.Sprintf("/api/v1/studies/%s/test-study/", ID)
	_, err := c.ExecuteBuilder().
		PostRequest(url).
		Status(http.StatusOK).
		Decode(&response).
		Execute()
	if err != nil {
		return nil, err
	}

	return &response, nil
}
