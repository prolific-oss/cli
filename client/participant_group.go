package client

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/prolific-oss/cli/model"
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

// GetParticipantGroups will return all the participant groups you have access to for a given WorkspaceID
func (c *Client) GetParticipantGroups(workspaceID string, limit, offset int) (*ListParticipantGroupsResponse, error) {
	var response ListParticipantGroupsResponse

	params := url.Values{"workspace_id": {workspaceID}, "limit": {strconv.Itoa(limit)}, "offset": {strconv.Itoa(offset)}}
	_, err := c.ExecuteBuilder().GetInto("/api/v1/participant-groups/?"+params.Encode(), &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// GetParticipantGroup will return the membership in the group
func (c *Client) GetParticipantGroup(groupID string) (*ViewParticipantGroupResponse, error) {
	var response ViewParticipantGroupResponse

	url := fmt.Sprintf("/api/v1/participant-groups/%s/participants/", groupID)
	_, err := c.ExecuteBuilder().GetInto(url, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// CreateParticipantGroup will create a new participant group
func (c *Client) CreateParticipantGroup(group model.CreateParticipantGroup) (*CreateParticipantGroupResponse, error) {
	var response CreateParticipantGroupResponse

	url := "/api/v1/participant-groups/"
	_, err := c.ExecuteBuilder().
		PostRequest(url).
		Body(group).
		Decode(&response).
		Execute()
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *Client) RemoveParticipantGroupMembers(groupID string, participantIDs []string) (*ViewParticipantGroupResponse, error) {
	payload := RemoveParticipantGroupMembersPayload{
		ParticipantIDs: participantIDs,
	}
	var response ViewParticipantGroupResponse

	url := fmt.Sprintf("/api/v1/participant-groups/%s/participants/", groupID)
	_, err := c.ExecuteBuilder().
		DeleteRequest(url).
		Body(payload).
		Status(http.StatusOK).
		Decode(&response).
		Execute()
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *Client) AddParticipantGroupMembers(groupID string, participantIDs []string) (*ViewParticipantGroupResponse, error) {
	payload := AddParticipantGroupMembersPayload{
		ParticipantIDs: participantIDs,
	}
	var response ViewParticipantGroupResponse

	url := fmt.Sprintf("/api/v1/participant-groups/%s/participants/", groupID)
	_, err := c.ExecuteBuilder().
		PostRequest(url).
		Body(payload).
		Status(http.StatusOK).
		Decode(&response).
		Execute()
	if err != nil {
		return nil, err
	}

	return &response, nil
}
