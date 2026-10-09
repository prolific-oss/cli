package client

import (
	"net/url"
	"strconv"
)

// StudyPageSize is the number of studies requested per page. It is sent
// explicitly rather than relying on the endpoint's own default of 20, so the
// window a page covers is a number the CLI chose and can report truthfully,
// and it matches the record count every other listing returns.
const StudyPageSize = DefaultRecordLimit

// SearchStudies searches accessible studies by name, internal name, or study ID.
// Studies paginate by page number rather than limit/offset, so the caller asks
// for a page and the page size fixes how many records that page holds.
func (c *Client) SearchStudies(query, workspaceID string, page int) (*ListStudiesResponse, error) {
	params := url.Values{
		"search":    {query},
		"page":      {strconv.Itoa(page)},
		"page_size": {strconv.Itoa(StudyPageSize)},
	}
	if workspaceID != "" {
		params.Set("workspace_id", workspaceID)
	}
	var response ListStudiesResponse
	if _, err := c.ExecuteBuilder().GetInto("/api/v1/studies/?"+params.Encode(), &response); err != nil {
		return nil, err
	}
	return &response, nil
}
