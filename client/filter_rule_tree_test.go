package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Verify the HTTP contract until the endpoint is published in the OpenAPI spec.
func TestGetFilterRuleTree(t *testing.T) {
	const tree = `{"rule_tree":{"and":{"max_children":null,"max_by_type":{"or":0},"children":{"leaf":true},"future_constraint":true}}}`
	for _, tc := range []struct {
		name, workspace, body string
		status                int
		wantError             bool
	}{
		{"default rules", "", tree, http.StatusOK, false},
		{"workspace rules", "workspace & other", tree, http.StatusOK, false},
		{"access denied", "", `{"detail":"access denied"}`, http.StatusForbidden, true},
		{"missing tree", "", `{}`, http.StatusOK, true},
		{"empty tree", "", `{"rule_tree":{}}`, http.StatusOK, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/api/v1/filters/rule-tree/", r.URL.Path)
				assert.Equal(t, "Token test-token", r.Header.Get("Authorization"))
				assert.Equal(t, tc.workspace, r.URL.Query().Get("workspace_id"))
				assert.Equal(t, tc.workspace != "", r.URL.Query().Has("workspace_id"))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				_, err := w.Write([]byte(tc.body))
				assert.NoError(t, err)
			}))
			defer server.Close()
			c := Client{Client: server.Client(), BaseURL: server.URL, Token: "test-token"}
			result, err := c.GetFilterRuleTree(tc.workspace)
			if tc.wantError {
				require.Error(t, err)
				require.Nil(t, result)
				return
			}
			require.NoError(t, err)
			encoded, err := json.Marshal(result)
			require.NoError(t, err)
			require.JSONEq(t, tree, string(encoded))
		})
	}
}
