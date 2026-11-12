package client

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prolific-oss/cli/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetStudiesRequests(t *testing.T) {
	const studiesPath = "/api/v1/studies/"

	methods := []struct {
		name string
		call func(*Client, string, string) error
	}{
		{"typed", func(c *Client, status, projectID string) error {
			_, err := c.GetStudies(status, projectID)
			return err
		}},
		{"raw", func(c *Client, status, projectID string) error {
			_, err := c.GetStudiesRaw(status, projectID)
			return err
		}},
	}
	statuses := []struct {
		name, status, query string
	}{
		{"empty", "", ""},
		{"all", model.StatusAll, ""},
		{"unpublished", model.StatusUnpublished, "published=0"},
		{"active", model.StatusActive, "active=1"},
		{"completed", model.StatusCompleted, "completed=1"},
		{"invalid", "invalid", ""},
	}
	for _, method := range methods {
		for _, status := range statuses {
			for _, projectID := range []string{"", "proj-id"} {
				t.Run(method.name+"/"+status.name+"/"+projectID, func(t *testing.T) {
					requests := 0
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						requests++
						path := studiesPath
						if projectID != "" {
							path = "/api/v1/projects/proj-id/studies/"
						}
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, path, r.URL.Path)
						assert.Equal(t, status.query, r.URL.RawQuery)
						assert.Equal(t, "Token test-token", r.Header.Get("Authorization"))
						_, err := w.Write([]byte(`{"results":[],"meta":{"count":0}}`))
						assert.NoError(t, err)
					}))
					defer server.Close()
					c := &Client{Client: server.Client(), BaseURL: server.URL, Token: "test-token"}
					err := method.call(c, status.status, projectID)
					if status.name == "invalid" {
						require.ErrorContains(t, err, "invalid is not a valid status")
						require.Zero(t, requests)
						return
					}
					require.NoError(t, err)
					require.Equal(t, 1, requests)
				})
			}
		}
	}
}

func TestGetStudiesRawPreservesResponse(t *testing.T) {
	for _, tc := range []struct {
		name, body string
	}{
		{"unknown fields", `{"results":[{"id":"study-id","future_field":{"large_integer":9007199254740993,"precise_decimal":0.1234567890123456789,"nullable":null,"enabled":false,"empty_list":[],"empty_object":{}}}],"meta":{"count":1,"future_meta":true},"_links":{"next":{"href":null}},"future_root":"value"}`},
		{"model-incompatible fields", `{"results":[{"id":"study-id","date_created":"a future date format","reward":"a future reward format"}]}`},
		{"empty results", `{"results":[],"meta":{"count":0}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, err := w.Write([]byte(tc.body))
				assert.NoError(t, err)
			}))
			defer server.Close()
			c := Client{Client: server.Client(), BaseURL: server.URL, Token: "test-token"}
			response, err := c.GetStudiesRaw(model.StatusAll, "")
			require.NoError(t, err)
			// Exact bytes catch float rounding that a map[string]any comparison could miss.
			require.Equal(t, tc.body, string(response))
		})
	}
}

func TestGetStudiesRawErrors(t *testing.T) {
	for _, tc := range []struct {
		name, body, wantError string
		status                int
	}{
		{"nested API error", `{"error":{"detail":"invalid request"}}`, "request failed: invalid request", http.StatusBadRequest},
		{"simple API error", `{"message":"Forbidden","detail":"access denied"}`, "request failed: Forbidden - access denied", http.StatusForbidden},
		{"unrecognized API error", "server error", "server error", http.StatusInternalServerError},
		{"invalid JSON", `{"results":`, "decoding JSON response", http.StatusOK},
		{"empty response", "", "decoding JSON response", http.StatusOK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.status)
				_, err := w.Write([]byte(tc.body))
				assert.NoError(t, err)
			}))
			defer server.Close()
			c := Client{Client: server.Client(), BaseURL: server.URL, Token: "test-token"}
			response, err := c.GetStudiesRaw(model.StatusAll, "")
			require.ErrorContains(t, err, tc.wantError)
			require.Nil(t, response)
		})
	}
}

func TestGetStudiesRawRequiresToken(t *testing.T) {
	c := Client{}
	response, err := c.GetStudiesRaw(model.StatusAll, "")
	require.ErrorContains(t, err, "PROLIFIC_TOKEN not set")
	require.Nil(t, response)
}
