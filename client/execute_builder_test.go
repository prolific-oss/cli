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

// GetInto accepts any status Client.Execute already treated as a success,
// unlike Get which asserts 200 — this covers callers that never checked the
// status code beyond that.
func TestExecuteBuilderGetIntoAcceptsAnySuccessStatus(t *testing.T) {
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
	httpResponse, err := c.ExecuteBuilder().GetInto("/some-path", &resp)

	require.NoError(t, err)
	require.Equal(t, http.StatusAccepted, httpResponse.StatusCode)
	require.Equal(t, "123", resp.ID)
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

// Without a Status call, the builder accepts whatever Client.Execute treated
// as a success (i.e. anything below 400) rather than erroring on every
// request — this covers callers that never asserted a specific status code.
func TestExecuteBuilderExecuteWithoutStatusAcceptsAnySuccess(t *testing.T) {
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
	httpResponse, err := c.ExecuteBuilder().
		GetRequest("/some-path").
		Decode(&resp).
		Execute()

	require.NoError(t, err)
	require.Equal(t, http.StatusAccepted, httpResponse.StatusCode)
	require.Equal(t, "123", resp.ID)
}

func TestExecuteBuilderPostRequestSendsBodyAndMethod(t *testing.T) {
	var gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		var got testResponse
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(got); err != nil {
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
		PostRequest("/some-path").
		Body(testResponse{ID: "123", Name: "test"}).
		Status(http.StatusCreated).
		Decode(&resp).
		Execute()

	require.NoError(t, err)
	require.Equal(t, http.MethodPost, gotMethod)
	require.Equal(t, http.StatusCreated, httpResponse.StatusCode)
	require.Equal(t, "123", resp.ID)
}

func TestExecuteBuilderPatchRequestUsesPatchMethod(t *testing.T) {
	var gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := Client{
		Client:  http.DefaultClient,
		BaseURL: srv.URL,
		Token:   "fake-token",
	}

	_, err := c.ExecuteBuilder().
		PatchRequest("/some-path").
		Body(testResponse{ID: "123"}).
		Status(http.StatusOK).
		Execute()

	require.NoError(t, err)
	require.Equal(t, http.MethodPatch, gotMethod)
}

func TestExecuteBuilderDeleteRequestUsesDeleteMethod(t *testing.T) {
	var gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := Client{
		Client:  http.DefaultClient,
		BaseURL: srv.URL,
		Token:   "fake-token",
	}

	_, err := c.ExecuteBuilder().
		DeleteRequest("/some-path").
		Status(http.StatusNoContent).
		Execute()

	require.NoError(t, err)
	require.Equal(t, http.MethodDelete, gotMethod)
}
