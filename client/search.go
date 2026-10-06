package client

import (
	"net/url"
	"strconv"
)

// SearchStudies searches accessible studies by name, internal name, or study ID.
// Studies use page-number pagination; the endpoint does not support limit/offset.
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

// SearchParticipantGroups searches group names within a workspace on the server.
func (c *Client) SearchParticipantGroups(query, workspaceID string, limit, offset int) (*ListParticipantGroupsResponse, error) {
	params := url.Values{"search": {query}, "workspace_id": {workspaceID}, "limit": {strconv.Itoa(limit)}, "offset": {strconv.Itoa(offset)}}
	var response ListParticipantGroupsResponse
	if _, err := c.ExecuteBuilder().GetInto("/api/v1/participant-groups/?"+params.Encode(), &response); err != nil {
		return nil, err
	}
	return &response, nil
}
