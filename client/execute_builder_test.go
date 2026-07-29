package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// testResponse is a simple struct for decoding test responses
type testResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func TestExecuteBuilderGet(t *testing.T) {
	// Create a mock server that returns JSON
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(testResponse{ID: "123", Name: "test"}); err != nil {
			t.Logf("failed to encode response: %v", err)
		}
	}))
	defer srv.Close()

	c := Client{
		Client:  http.DefaultClient,
		BaseURL: srv.URL,
		Token:   "fake-token", // satisfies the token check in Execute()
	}

	var resp testResponse
	_, err := c.ExecuteBuilder().Get("/some-path", &resp)
	require.NoError(t, err)
	require.Equal(t, "123", resp.ID)
	require.Equal(t, "test", resp.Name)
}

func TestExecuteBuilderGetReturnsErrorOnHTTPFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"}); err != nil {
			t.Logf("failed to encode response: %v", err)
		}
	}))
	defer srv.Close()

	c := Client{
		Client:  http.DefaultClient,
		BaseURL: srv.URL,
		Token:   "fake-token", // satisfies the token check in Execute()
	}

	var resp testResponse
	_, err := c.ExecuteBuilder().Get("/some-path", &resp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unable to fulfil request")
}

// A 202 is a success as far as Client.Execute is concerned (it only errors on
// 4xx and above), so this exercises the builder's own status check rather than
// the error path inside Execute.
func TestExecuteBuilderGetReturnsErrorOnUnexpectedSuccessStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		if err := json.NewEncoder(w).Encode(testResponse{ID: "123", Name: "test"}); err != nil {
			t.Logf("failed to encode response: %v", err)
		}
	}))
	defer srv.Close()

	c := Client{
		Client:  http.DefaultClient,
		BaseURL: srv.URL,
		Token:   "fake-token",
	}

	var resp testResponse
	_, err := c.ExecuteBuilder().Get("/some-path", &resp)

	require.Error(t, err)
	require.Contains(t, err.Error(), "unexpected status code 202")
}

func TestExecuteBuilderExecuteAcceptsAConfiguredStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(testResponse{ID: "123", Name: "test"}); err != nil {
			t.Logf("failed to encode response: %v", err)
		}
	}))
	defer srv.Close()

	c := Client{
		Client:  http.DefaultClient,
		BaseURL: srv.URL,
		Token:   "fake-token",
	}

	var resp testResponse
	httpResponse, err := c.ExecuteBuilder().
		GetRequest("/some-path").
		Status(http.StatusCreated).
		Decode(&resp).
		Execute()

	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, httpResponse.StatusCode)
	require.Equal(t, "123", resp.ID)
}
