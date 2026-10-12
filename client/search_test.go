package client

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResourceSearchRequests(t *testing.T) {
	for _, tc := range []struct {
		name, path string
		query      url.Values
		call       func(*Client) error
	}{
		{"studies", "/api/v1/studies/", url.Values{"search": {"memory & attention"}, "workspace_id": {"ws/+"}, "page": {"2"}, "page_size": {"200"}}, func(c *Client) error { _, err := c.SearchStudies("memory & attention", "ws/+", 2); return err }},
		{"studies without workspace", "/api/v1/studies/", url.Values{"search": {"memory"}, "page": {"1"}, "page_size": {"200"}}, func(c *Client) error { _, err := c.SearchStudies("memory", "", 1); return err }},
		{"participant groups", "/api/v1/participant-groups/", url.Values{"search": {"memory & attention"}, "workspace_id": {"ws/+"}, "limit": {"10"}, "offset": {"20"}}, func(c *Client) error {
			_, err := c.SearchParticipantGroups("memory & attention", "ws/+", 10, 20)
			return err
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(t, tc.path, r.URL.Path)
				require.Equal(t, tc.query, r.URL.Query())
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write([]byte(`{"results":[],"meta":{"count":0}}`))
				require.NoError(t, err)
			}))
			defer server.Close()
			c := &Client{Client: server.Client(), BaseURL: server.URL, Token: "test-token"}
			require.NoError(t, tc.call(c))
		})
	}
}
