package client

import (
	"fmt"
	"net/url"
)

// sensitiveExecuteBuilder builds requests without debug logging. Feedback
// responses must never be written to logs, even when PROLIFIC_DEBUG is enabled.
func (c *Client) sensitiveExecuteBuilder() *ExecuteBuilder {
	sensitiveClient := *c
	sensitiveClient.Debug = false
	return sensitiveClient.ExecuteBuilder()
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
	if _, err := c.sensitiveExecuteBuilder().GetInto(requestURL, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// GetStudyRatings returns the aggregated study-level feedback ratings.
func (c *Client) GetStudyRatings(studyID string) (*StudyRatingsResponse, error) {
	var response StudyRatingsResponse

	requestURL := fmt.Sprintf("/api/v1/studies/%s/feedback/ratings/", studyID)
	if _, err := c.sensitiveExecuteBuilder().GetInto(requestURL, &response); err != nil {
		return nil, err
	}

	return &response, nil
}
