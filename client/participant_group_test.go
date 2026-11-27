package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// TestGetParticipantGroupsSendsWorkspaceIDQueryParam guards the request shape
// directly, since contract_test skips this operation (HARNESSGAP: kin-openapi
// can't decode its oneOf-of-objects query param).
func TestGetParticipantGroupsSendsWorkspaceIDQueryParam(t *testing.T) {
	var gotMethod, gotPath string
	var gotQuery url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotQuery = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(ListParticipantGroupsResponse{}); err != nil {
			t.Logf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := Client{
		Client:  server.Client(),
		BaseURL: server.URL,
		Token:   "test-token",
	}

	_, err := c.GetParticipantGroups("ws/+&name=value", 10, 0)
	if err != nil {
		t.Fatalf("GetParticipantGroups returned error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("method = %q, want %q", gotMethod, http.MethodGet)
	}
	if want := "/api/v1/participant-groups/"; gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
	if got := gotQuery.Get("workspace_id"); got != "ws/+&name=value" {
		t.Errorf("workspace_id = %q, want %q", got, "ws/+&name=value")
	}
	if got := gotQuery.Get("limit"); got != "10" {
		t.Errorf("limit = %q, want %q", got, "10")
	}
	if got := gotQuery.Get("offset"); got != "0" {
		t.Errorf("offset = %q, want %q", got, "0")
	}
	if gotQuery.Has("project_id") {
		t.Errorf("project_id should not be sent, got %q", gotQuery.Get("project_id"))
	}
}

// TestParticipantGroupMembersRequestShape guards the verb, path and body of the
// add/remove membership calls, which are otherwise only mocked in the command tests.
func TestParticipantGroupMembersRequestShape(t *testing.T) {
	tests := map[string]struct {
		method string
		call   func(c *Client) (*ViewParticipantGroupResponse, error)
	}{
		"add": {
			method: http.MethodPost,
			call: func(c *Client) (*ViewParticipantGroupResponse, error) {
				return c.AddParticipantGroupMembers("group-1", []string{"p1", "p2"})
			},
		},
		"remove": {
			method: http.MethodDelete,
			call: func(c *Client) (*ViewParticipantGroupResponse, error) {
				return c.RemoveParticipantGroupMembers("group-1", []string{"p1", "p2"})
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var gotMethod, gotPath string
			var gotBody map[string][]string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotMethod = r.Method
				gotPath = r.URL.Path
				if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
					t.Errorf("failed to decode request body: %v", err)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"results":[{"participant_id":"p3"}]}`))
			}))
			defer server.Close()

			c := Client{Client: server.Client(), BaseURL: server.URL, Token: "test-token"}

			response, err := tc.call(&c)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if gotMethod != tc.method {
				t.Errorf("method = %q, want %q", gotMethod, tc.method)
			}
			if want := "/api/v1/participant-groups/group-1/participants/"; gotPath != want {
				t.Errorf("path = %q, want %q", gotPath, want)
			}
			if ids := gotBody["participant_ids"]; len(ids) != 2 || ids[0] != "p1" || ids[1] != "p2" {
				t.Errorf("participant_ids = %v, want [p1 p2]", ids)
			}
			if len(response.Results) != 1 || response.Results[0].ParticipantID != "p3" {
				t.Errorf("unexpected response: %+v", response)
			}
		})
	}
}

func TestParticipantGroupMembersReturnsErrorOnNon200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"detail":"bad request"}}`))
	}))
	defer server.Close()

	c := Client{Client: server.Client(), BaseURL: server.URL, Token: "test-token"}

	if _, err := c.AddParticipantGroupMembers("group-1", []string{"p1"}); err == nil {
		t.Error("AddParticipantGroupMembers: expected error, got nil")
	}
	if _, err := c.RemoveParticipantGroupMembers("group-1", []string{"p1"}); err == nil {
		t.Error("RemoveParticipantGroupMembers: expected error, got nil")
	}
}
