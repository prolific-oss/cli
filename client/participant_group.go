package client

import (
	"net/url"
	"strconv"
)

// SearchParticipantGroups searches group names within a workspace on the server.
func (c *Client) SearchParticipantGroups(query, workspaceID string, limit, offset int) (*ListParticipantGroupsResponse, error) {
	params := url.Values{"search": {query}, "workspace_id": {workspaceID}, "limit": {strconv.Itoa(limit)}, "offset": {strconv.Itoa(offset)}}
	var response ListParticipantGroupsResponse
	if _, err := c.ExecuteBuilder().GetInto("/api/v1/participant-groups/?"+params.Encode(), &response); err != nil {
		return nil, err
	}
	return &response, nil
}
