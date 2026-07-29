package client

import (
	"fmt"
	"net/http"
	"net/url"
)

// executeSensitive runs a request without debug logging. Feedback responses
// must never be written to logs, even when PROLIFIC_DEBUG is enabled.
func (c *Client) executeSensitive(method, requestURL string, body any, response any) (*http.Response, error) {
	sensitiveClient := *c
	sensitiveClient.Debug = false
	return sensitiveClient.Execute(method, requestURL, body, response)
}

// GetStudyFeedback returns participant feedback responses for a study. A
// limit or offset of 0 is omitted so the API returns every record.
func (c *Client) GetStudyFeedback(studyID string, hasWrittenFeedback bool, limit, offset int) (*ListStudyFeedbackResponse, error) {
	var response ListStudyFeedbackResponse

	params := url.Values{}
	params.Set("has_written_feedback", fmt.Sprintf("%v", hasWrittenFeedback))
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%v", limit))
	}
	if offset > 0 {
		params.Set("offset", fmt.Sprintf("%v", offset))
	}

	requestURL := fmt.Sprintf("/api/v1/studies/%s/feedback/responses/?%s", studyID, params.Encode())
	_, err := c.executeSensitive(http.MethodGet, requestURL, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", requestURL, err)
	}

	return &response, nil
}

// GetStudyRatings returns the aggregated study-level feedback ratings.
func (c *Client) GetStudyRatings(studyID string) (*StudyRatingsResponse, error) {
	var response StudyRatingsResponse

	requestURL := fmt.Sprintf("/api/v1/studies/%s/feedback/ratings/", studyID)
	_, err := c.executeSensitive(http.MethodGet, requestURL, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", requestURL, err)
	}

	return &response, nil
}
