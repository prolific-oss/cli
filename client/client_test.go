package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/prolific-oss/cli/model"
	"github.com/prolific-oss/cli/version"
)

// testWorkspaceID is the workspace ID the request-shape tests send.
const testWorkspaceID = "ws-id"

func TestFormatBatchErrorBody(t *testing.T) {
	tests := []struct {
		name     string
		body     []byte
		expected string
	}{
		{
			name:     "non-JSON body returned as-is",
			body:     []byte("internal server error"),
			expected: "internal server error",
		},
		{
			name:     "non-INVALID_BATCH_ITEMS error returned as-is",
			body:     []byte(`{"type":"SOME_OTHER_ERROR","message":"something went wrong"}`),
			expected: `{"type":"SOME_OTHER_ERROR","message":"something went wrong"}`,
		},
		{
			name:     "INVALID_BATCH_ITEMS with no issues",
			body:     []byte(`{"type":"INVALID_BATCH_ITEMS","issues":[]}`),
			expected: "batch_items validation failed:",
		},
		{
			name:     "INVALID_BATCH_ITEMS with single issue, no field",
			body:     []byte(`{"type":"INVALID_BATCH_ITEMS","issues":[{"page":0,"row":0,"column":0,"item":0,"type":"free_text","message":"description is required"}]}`),
			expected: "batch_items validation failed:\n  Page 1, Row 1, Column 1, Item 1 (free_text): description is required",
		},
		{
			name:     "INVALID_BATCH_ITEMS with field reference",
			body:     []byte(`{"type":"INVALID_BATCH_ITEMS","issues":[{"page":0,"row":1,"column":0,"item":2,"type":"dataset_field","field":"missing_col","message":"Field does not exist in the dataset schema"}]}`),
			expected: "batch_items validation failed:\n  Page 1, Row 2, Column 1, Item 3 (dataset_field) \"missing_col\": Field does not exist in the dataset schema",
		},
		{
			name:     "INVALID_BATCH_ITEMS with multiple issues",
			body:     []byte(`{"type":"INVALID_BATCH_ITEMS","issues":[{"page":0,"row":0,"column":0,"item":0,"type":"free_text","message":"description is required"},{"page":1,"row":2,"column":1,"item":3,"type":"multiple_choice","message":"answer_limit exceeds number of options"}]}`),
			expected: "batch_items validation failed:\n  Page 1, Row 1, Column 1, Item 1 (free_text): description is required\n  Page 2, Row 3, Column 2, Item 4 (multiple_choice): answer_limit exceeds number of options",
		},
		{
			// A non-content-block item (e.g. an instruction) on a display_position "intro"/"outro"
			// page — see DISPLAY_POSITION_CONTENT_ONLY_MESSAGE in data-collection-tool. No `field`
			// is set for this issue type, so it should format the same as any other fieldless issue.
			name:     "INVALID_BATCH_ITEMS with display_position content-only violation",
			body:     []byte(`{"type":"INVALID_BATCH_ITEMS","issues":[{"page":0,"row":0,"column":0,"item":0,"type":"free_text","message":"Pages with display_position \"intro\" or \"outro\" may only contain image or rich_text content blocks"}]}`),
			expected: `batch_items validation failed:` + "\n" + `  Page 1, Row 1, Column 1, Item 1 (free_text): Pages with display_position "intro" or "outro" may only contain image or rich_text content blocks`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatBatchErrorBody(tt.body)
			if got != tt.expected {
				t.Fatalf("expected:\n%q\ngot:\n%q", tt.expected, got)
			}
		})
	}
}

func TestExecuteTruncatesOversizedErrorDetail(t *testing.T) {
	oversizedDetail := strings.Repeat("x", maxErrorDetailLen*3)

	tests := []struct {
		name          string
		statusCode    int
		body          string
		wantTruncated bool
	}{
		{
			name:          "JSONAPIError with oversized detail is truncated and hints at debug mode",
			statusCode:    http.StatusBadRequest,
			body:          fmt.Sprintf(`{"error":{"status":400,"title":"Bad Request","detail":%q}}`, oversizedDetail),
			wantTruncated: true,
		},
		{
			name:          "JSONAPIError with small detail is unchanged",
			statusCode:    http.StatusBadRequest,
			body:          `{"error":{"status":400,"title":"Bad Request","detail":"study ID is invalid"}}`,
			wantTruncated: false,
		},
		{
			name:          "SimpleAPIError with oversized detail is truncated and hints at debug mode",
			statusCode:    http.StatusBadRequest,
			body:          fmt.Sprintf(`{"message":"Bad Request","detail":%q}`, oversizedDetail),
			wantTruncated: true,
		},
		{
			name:          "SimpleAPIError with small detail is unchanged",
			statusCode:    http.StatusBadRequest,
			body:          `{"message":"Bad Request","detail":"study not found"}`,
			wantTruncated: false,
		},
		{
			name:          "unrecognized error body is truncated and hints at debug mode",
			statusCode:    http.StatusInternalServerError,
			body:          strings.Repeat("y", maxErrorDetailLen*3),
			wantTruncated: true,
		},
		{
			name:          "small unrecognized error body is unchanged",
			statusCode:    http.StatusInternalServerError,
			body:          "internal server error",
			wantTruncated: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			c := Client{Client: server.Client(), BaseURL: server.URL, Token: "test-token"}

			_, err := c.Execute(http.MethodGet, "/", nil, nil)
			if err == nil {
				t.Fatalf("expected an error, got nil")
			}

			if tt.wantTruncated {
				if !strings.Contains(err.Error(), "PROLIFIC_DEBUG=1") {
					t.Fatalf("expected truncated error to mention PROLIFIC_DEBUG=1, got: %q", err.Error())
				}
				if len(err.Error()) >= len(tt.body) {
					t.Fatalf("expected error message to be shorter than the oversized body (%d bytes), got %d bytes", len(tt.body), len(err.Error()))
				}
			} else {
				if strings.Contains(err.Error(), "PROLIFIC_DEBUG=1") {
					t.Fatalf("expected small error not to be truncated, got: %q", err.Error())
				}
			}
		})
	}
}

func TestExecutePreservesHTTPStatusOnAPIErrors(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{
			name: "nested API error",
			body: `{"error":{"status":403,"title":"Forbidden","detail":"feature unavailable"}}`,
		},
		{
			name: "simple API error",
			body: `{"message":"Forbidden","detail":"feature unavailable"}`,
		},
		{
			name: "unrecognized API error",
			body: "forbidden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			c := Client{Client: server.Client(), BaseURL: server.URL, Token: "test-token"}
			_, err := c.Execute(http.MethodGet, "/", nil, nil)

			if !IsHTTPStatusError(err, http.StatusForbidden) {
				t.Fatalf("expected error to retain status 403, got %T: %v", err, err)
			}
			if IsHTTPStatusError(err, http.StatusNotFound) {
				t.Fatalf("did not expect error to match status 404: %v", err)
			}
		})
	}
}

// TestTruncateErrorDetailMultiByteRuneCount guards against comparing byte length to a rune-based cap.
func TestTruncateErrorDetailMultiByteRuneCount(t *testing.T) {
	// 300 runes of a 3-byte CJK character: 900 bytes, but only 300 runes.
	s := strings.Repeat("世", 300)

	got := truncateErrorDetail(s)

	if got != s {
		t.Fatalf("expected a string under the rune cap to be returned unchanged, got: %q", got)
	}
}

// TestUnrecognizedAPIErrorBodyRemainsIntactAfterTruncation guards the INVALID_BATCH_ITEMS path's json.Unmarshal of Body.
func TestUnrecognizedAPIErrorBodyRemainsIntactAfterTruncation(t *testing.T) {
	issue := struct {
		Page    int    `json:"page"`
		Row     int    `json:"row"`
		Column  int    `json:"column"`
		Item    int    `json:"item"`
		Type    string `json:"type"`
		Message string `json:"message"`
	}{Message: strings.Repeat("z", maxErrorDetailLen*3)}

	payload := struct {
		Type   string        `json:"type"`
		Issues []interface{} `json:"issues"`
	}{Type: "INVALID_BATCH_ITEMS", Issues: []interface{}{issue}}

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal fixture: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(body)
	}))
	defer server.Close()

	c := Client{Client: server.Client(), BaseURL: server.URL, Token: "test-token"}

	_, execErr := c.Execute(http.MethodGet, "/", nil, nil)
	if execErr == nil {
		t.Fatalf("expected an error, got nil")
	}

	var apiErr *UnrecognizedAPIError
	if !errors.As(execErr, &apiErr) {
		t.Fatalf("expected *UnrecognizedAPIError, got %T: %v", execErr, execErr)
	}

	if !bytes.Equal(apiErr.Body, body) {
		t.Fatalf("UnrecognizedAPIError.Body was mutated/truncated: got %d bytes, want %d bytes", len(apiErr.Body), len(body))
	}

	got := formatBatchErrorBody(apiErr.Body)
	if !strings.Contains(got, issue.Message) {
		t.Fatalf("formatBatchErrorBody lost the full untruncated message; got %d chars, want to contain %d-char message", len(got), len(issue.Message))
	}

	if !strings.Contains(apiErr.Error(), "PROLIFIC_DEBUG=1") {
		t.Fatalf("expected Error() to be truncated with a debug hint, got: %q", apiErr.Error())
	}
}

// TestUpdateAITaskBuilderBatch characterizes the current error-branching behaviour
// before any ExecuteBuilder migration: it does typed error inspection
// (errors.As for *UnrecognizedAPIError) directly on the raw error from
// Client.Execute, which only works because that error isn't wrapped before
// the check. Any migration must preserve this or the INVALID_BATCH_ITEMS
// formatting silently stops firing.
func TestUpdateAITaskBuilderBatch(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantErr    string
		wantOK     bool
	}{
		{
			name:       "success",
			statusCode: http.StatusOK,
			body:       `{"batch_id":"batch-1"}`,
			wantOK:     true,
		},
		{
			name:       "INVALID_BATCH_ITEMS on a 4xx response is formatted via errors.As(*UnrecognizedAPIError)",
			statusCode: http.StatusBadRequest,
			body:       `{"type":"INVALID_BATCH_ITEMS","issues":[{"page":0,"row":0,"column":0,"item":0,"type":"free_text","message":"description is required"}]}`,
			wantErr:    "unable to update batch: batch_items validation failed:\n  Page 1, Row 1, Column 1, Item 1 (free_text): description is required",
		},
		{
			name:       "INVALID_BATCH_ITEMS on a non-error success status is formatted via the StatusCode check",
			statusCode: http.StatusAccepted,
			body:       `{"type":"INVALID_BATCH_ITEMS","issues":[{"page":0,"row":0,"column":0,"item":0,"type":"free_text","message":"description is required"}]}`,
			wantErr:    "unable to update batch: batch_items validation failed:\n  Page 1, Row 1, Column 1, Item 1 (free_text): description is required",
		},
		{
			name:       "a recognized error shape on a 4xx response takes the generic error path, not formatBatchErrorBody",
			statusCode: http.StatusBadRequest,
			body:       `{"detail":"batch not found"}`,
			wantErr:    "unable to fulfil request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			c := Client{
				Client:  server.Client(),
				BaseURL: server.URL,
				Token:   "test-token",
			}

			resp, err := c.UpdateAITaskBuilderBatch(UpdateBatchParams{BatchID: "batch-1", Name: "updated"})

			if tt.wantOK {
				if err != nil {
					t.Fatalf("expected no error, got: %v", err)
				}
				if resp == nil {
					t.Fatalf("expected a response, got nil")
				}
				return
			}

			if err == nil {
				t.Fatalf("expected an error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error to contain %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

// TestCreateAITaskBuilderBatch characterizes the same errors.As(*UnrecognizedAPIError)
// pattern as TestUpdateAITaskBuilderBatch above.
func TestCreateAITaskBuilderBatch(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantErr    string
		wantOK     bool
	}{
		{
			name:       "success",
			statusCode: http.StatusCreated,
			body:       `{"batch_id":"batch-1"}`,
			wantOK:     true,
		},
		{
			name:       "INVALID_BATCH_ITEMS on a 4xx response is formatted via errors.As(*UnrecognizedAPIError)",
			statusCode: http.StatusBadRequest,
			body:       `{"type":"INVALID_BATCH_ITEMS","issues":[{"page":0,"row":0,"column":0,"item":0,"type":"free_text","message":"description is required"}]}`,
			wantErr:    "unable to create batch: batch_items validation failed:\n  Page 1, Row 1, Column 1, Item 1 (free_text): description is required",
		},
		{
			name:       "INVALID_BATCH_ITEMS on a non-error success status is formatted via the StatusCode check",
			statusCode: http.StatusOK,
			body:       `{"type":"INVALID_BATCH_ITEMS","issues":[{"page":0,"row":0,"column":0,"item":0,"type":"free_text","message":"description is required"}]}`,
			wantErr:    "unable to create batch: batch_items validation failed:\n  Page 1, Row 1, Column 1, Item 1 (free_text): description is required",
		},
		{
			name:       "a recognized error shape on a 4xx response takes the generic error path, not formatBatchErrorBody",
			statusCode: http.StatusBadRequest,
			body:       `{"detail":"workspace not found"}`,
			wantErr:    "unable to fulfil request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			c := Client{
				Client:  server.Client(),
				BaseURL: server.URL,
				Token:   "test-token",
			}

			resp, err := c.CreateAITaskBuilderBatch(CreateBatchParams{
				Name: "batch", WorkspaceID: "ws-1", DatasetID: "ds-1",
				TaskName: "task", TaskIntroduction: "intro", TaskSteps: "steps",
			})

			if tt.wantOK {
				if err != nil {
					t.Fatalf("expected no error, got: %v", err)
				}
				if resp == nil {
					t.Fatalf("expected a response, got nil")
				}
				return
			}

			if err == nil {
				t.Fatalf("expected an error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error to contain %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

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

func TestComposeUserAgent(t *testing.T) {
	knownVars := []string{"CLAUDECODE", "ANTIGRAVITY_AGENT", "AI_AGENT", "LLM_AGENT"}

	tests := []struct {
		name     string
		skill    string
		agentEnv map[string]string
		want     string
	}{
		{
			name:  "no skill, no agent",
			skill: "",
			want:  "prolific-oss/cli/" + version.Get(),
		},
		{
			name:  "skill only",
			skill: "cli-command-create",
			want:  "prolific-oss/cli/" + version.Get() + " skill/cli-command-create",
		},
		{
			name:     "agent and skill together",
			skill:    "cli-command-create",
			agentEnv: map[string]string{"CLAUDECODE": "1"},
			want:     "prolific-oss/cli/" + version.Get() + " agent/claude-code skill/cli-command-create",
		},
		{
			name:  "invalid skill (control characters) is dropped",
			skill: "bad\nvalue",
			want:  "prolific-oss/cli/" + version.Get(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, k := range knownVars {
				t.Setenv(k, "") // isolate from ambient agent env vars
			}
			for k, v := range tt.agentEnv {
				t.Setenv(k, v)
			}

			if got := ComposeUserAgent(tt.skill); got != tt.want {
				t.Fatalf("ComposeUserAgent(%q) = %q, want %q", tt.skill, got, tt.want)
			}
		})
	}
}

func TestExecuteSetsSkillInUserAgent(t *testing.T) {
	// Isolate from ambient agent env vars (this shell has AI_AGENT set).
	for _, k := range []string{"CLAUDECODE", "ANTIGRAVITY_AGENT", "AI_AGENT", "LLM_AGENT"} {
		t.Setenv(k, "")
	}

	var gotUserAgent string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserAgent = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := Client{
		Client:  server.Client(),
		BaseURL: server.URL,
		Token:   "test-token",
		Skill:   "cli-command-create",
	}

	if _, err := c.Execute(http.MethodGet, "/studies", nil, nil); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if want := "prolific-oss/cli/" + version.Get() + " skill/cli-command-create"; gotUserAgent != want {
		t.Fatalf("User-Agent = %q, want %q", gotUserAgent, want)
	}
}

func TestExecuteSetsAgentInUserAgent(t *testing.T) {
	// Isolate from ambient agent env vars (this shell has AI_AGENT set).
	for _, k := range []string{"CLAUDECODE", "ANTIGRAVITY_AGENT", "AI_AGENT", "LLM_AGENT"} {
		t.Setenv(k, "")
	}
	t.Setenv("CLAUDECODE", "1")

	var gotUserAgent string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserAgent = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := Client{
		Client:  server.Client(),
		BaseURL: server.URL,
		Token:   "test-token",
	}

	if _, err := c.Execute(http.MethodGet, "/studies", nil, nil); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if want := "prolific-oss/cli/" + version.Get() + " agent/claude-code"; gotUserAgent != want {
		t.Fatalf("User-Agent = %q, want %q", gotUserAgent, want)
	}
}

func TestGetFiltersScopesTheCatalogueToAWorkspace(t *testing.T) {
	var gotPath string
	var gotQuery url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(ListFiltersResponse{})
	}))
	defer server.Close()

	c := Client{Client: server.Client(), BaseURL: server.URL, Token: "test-token"}

	if _, err := c.GetFilters(testWorkspaceID); err != nil {
		t.Fatalf("GetFilters returned error: %v", err)
	}
	if want := "/api/v1/filters/"; gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
	if got := gotQuery.Get("workspace_id"); got != testWorkspaceID {
		t.Errorf("workspace_id = %q, want %q", got, testWorkspaceID)
	}
}

func TestGetFiltersOmitsWorkspaceIDWhenEmpty(t *testing.T) {
	var gotQuery url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(ListFiltersResponse{})
	}))
	defer server.Close()

	c := Client{Client: server.Client(), BaseURL: server.URL, Token: "test-token"}

	if _, err := c.GetFilters(""); err != nil {
		t.Fatalf("GetFilters returned error: %v", err)
	}
	if gotQuery.Has("workspace_id") {
		t.Errorf("workspace_id should not be sent when empty, got %q", gotQuery.Get("workspace_id"))
	}
}

// TestSearchFiltersSendsExpectedRequest guards the request shape directly,
// since the operation is not yet in the published spec and so is not covered
// by contract_test.
func TestSearchFiltersSendsExpectedRequest(t *testing.T) {
	var gotMethod, gotPath string
	var gotQuery url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotQuery = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(SearchFiltersResponse{}); err != nil {
			t.Logf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := Client{
		Client:  server.Client(),
		BaseURL: server.URL,
		Token:   "test-token",
	}

	_, err := c.SearchFilters("software developer", testWorkspaceID, 25, 50)
	if err != nil {
		t.Fatalf("SearchFilters returned error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("method = %q, want %q", gotMethod, http.MethodGet)
	}
	if want := "/api/v1/filters/search/"; gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
	if got := gotQuery.Get("q"); got != "software developer" {
		t.Errorf("q = %q, want %q", got, "software developer")
	}
	if got := gotQuery.Get("workspace_id"); got != testWorkspaceID {
		t.Errorf("workspace_id = %q, want %q", got, testWorkspaceID)
	}
	if got := gotQuery.Get("limit"); got != "25" {
		t.Errorf("limit = %q, want %q", got, "25")
	}
	if got := gotQuery.Get("offset"); got != "50" {
		t.Errorf("offset = %q, want %q", got, "50")
	}
}

func TestSearchFiltersOmitsWorkspaceIDWhenEmpty(t *testing.T) {
	var gotQuery url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(SearchFiltersResponse{})
	}))
	defer server.Close()

	c := Client{Client: server.Client(), BaseURL: server.URL, Token: "test-token"}

	if _, err := c.SearchFilters("dev", "", 25, 0); err != nil {
		t.Fatalf("SearchFilters returned error: %v", err)
	}
	if gotQuery.Has("workspace_id") {
		t.Errorf("workspace_id should not be sent when empty, got %q", gotQuery.Get("workspace_id"))
	}
}

// TestGetFilterBreakdownSendsRequestAndDecodesResponse guards the request and
// response shape directly, since this endpoint isn't published in the
// OpenAPI spec yet and so contract_test cannot validate it.
func TestGetFilterBreakdownSendsRequestAndDecodesResponse(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody FilterBreakdownPayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(FilterBreakdownResponse{
			Breakdown: map[string]int{"0": 4, "1": 3, FilterBreakdownNAKey: 5},
		}); err != nil {
			t.Logf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := Client{
		Client:  server.Client(),
		BaseURL: server.URL,
		Token:   "test-token",
	}

	payload := FilterBreakdownPayload{
		Filters: []model.Filter{
			{FilterID: "age", SelectedRange: &model.FilterRange{Lower: float64(18), Upper: float64(65)}},
		},
		BreakdownFilter: model.Filter{FilterID: "handedness", SelectedValues: []string{"0", "1"}},
		WorkspaceID:     testWorkspaceID,
	}

	response, err := c.GetFilterBreakdown(payload)
	if err != nil {
		t.Fatalf("GetFilterBreakdown returned error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want %q", gotMethod, http.MethodPost)
	}
	if want := "/api/v1/eligibility-count/filter-breakdown/"; gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}

	if gotBody.WorkspaceID != payload.WorkspaceID {
		t.Errorf("request workspace_id = %q, want %q", gotBody.WorkspaceID, payload.WorkspaceID)
	}
	if len(gotBody.Filters) != 1 || gotBody.Filters[0].FilterID != "age" {
		t.Errorf("request filters = %+v, want a single 'age' filter", gotBody.Filters)
	}
	if gotBody.Filters[0].SelectedRange == nil || gotBody.Filters[0].SelectedRange.Lower != float64(18) || gotBody.Filters[0].SelectedRange.Upper != float64(65) {
		t.Errorf("request filters[0].selected_range = %+v, want {18 65}", gotBody.Filters[0].SelectedRange)
	}
	if gotBody.BreakdownFilter.FilterID != "handedness" {
		t.Errorf("request breakdown_filter.filter_id = %q, want %q", gotBody.BreakdownFilter.FilterID, "handedness")
	}
	if want := []string{"0", "1"}; len(gotBody.BreakdownFilter.SelectedValues) != len(want) ||
		gotBody.BreakdownFilter.SelectedValues[0] != want[0] || gotBody.BreakdownFilter.SelectedValues[1] != want[1] {
		t.Errorf("request breakdown_filter.selected_values = %v, want %v", gotBody.BreakdownFilter.SelectedValues, want)
	}

	wantBreakdown := map[string]int{"0": 4, "1": 3, FilterBreakdownNAKey: 5}
	if len(response.Breakdown) != len(wantBreakdown) {
		t.Fatalf("Breakdown = %+v, want %+v", response.Breakdown, wantBreakdown)
	}
	for key, want := range wantBreakdown {
		if got := response.Breakdown[key]; got != want {
			t.Errorf("Breakdown[%q] = %d, want %d", key, got, want)
		}
	}
}

// The filter choices endpoints are live but not published in the OpenAPI
// spec, so contract_test cannot validate them; these guard the request shape
// directly.
func TestGetFilterChoicesSendsExpectedRequest(t *testing.T) {
	var gotMethod, gotPath string
	var gotQuery url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotQuery = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(ListFilterChoicesResponse{})
	}))
	defer server.Close()

	c := Client{Client: server.Client(), BaseURL: server.URL, Token: "test-token"}

	if _, err := c.GetFilterChoices("job-title", testWorkspaceID, 100, 50); err != nil {
		t.Fatalf("GetFilterChoices returned error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("method = %q, want %q", gotMethod, http.MethodGet)
	}
	if want := "/api/v1/filters/job-title/choices/"; gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
	if got := gotQuery.Get("limit"); got != "100" {
		t.Errorf("limit = %q, want %q", got, "100")
	}
	if got := gotQuery.Get("offset"); got != "50" {
		t.Errorf("offset = %q, want %q", got, "50")
	}
	if got := gotQuery.Get("workspace_id"); got != testWorkspaceID {
		t.Errorf("workspace_id = %q, want %q", got, testWorkspaceID)
	}
}

// A filter ID is user-supplied and goes into the path, so it has to be
// escaped rather than concatenated.
func TestGetFilterChoicesEscapesTheFilterID(t *testing.T) {
	// The raw request URI is what shows the escaping; r.URL.Path has already
	// been decoded by the time the handler sees it.
	var gotURI string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURI = r.RequestURI
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(ListFilterChoicesResponse{})
	}))
	defer server.Close()

	c := Client{Client: server.Client(), BaseURL: server.URL, Token: "test-token"}

	if _, err := c.GetFilterChoices("job title/../x", "", 10, 0); err != nil {
		t.Fatalf("GetFilterChoices returned error: %v", err)
	}
	if !strings.HasPrefix(gotURI, "/api/v1/filters/job%20title%2F..%2Fx/choices/") {
		t.Errorf("request URI = %q, want the filter ID percent-escaped in the path", gotURI)
	}
}

func TestGetFilterChoicesOmitsWorkspaceIDWhenEmpty(t *testing.T) {
	var gotQuery url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(ListFilterChoicesResponse{})
	}))
	defer server.Close()

	c := Client{Client: server.Client(), BaseURL: server.URL, Token: "test-token"}

	if _, err := c.GetFilterChoices("job-title", "", 10, 0); err != nil {
		t.Fatalf("GetFilterChoices returned error: %v", err)
	}
	if gotQuery.Has("workspace_id") {
		t.Errorf("workspace_id should not be sent when empty, got %q", gotQuery.Get("workspace_id"))
	}
}

func TestGetFilterChoicesDecodesResultsAndCount(t *testing.T) {
	body := `{
	  "results": [
	    {"id": "0", "label": "Management Occupations", "parent_id": null, "num_children": 4, "num_descendants": 476},
	    {"id": "1016", "label": "Registered Nurses", "parent_id": "0", "num_children": 12, "num_descendants": 12}
	  ],
	  "_links": {"self": {"href": "x", "title": "Current"}},
	  "meta": {"count": 4123}
	}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()

	c := Client{Client: server.Client(), BaseURL: server.URL, Token: "test-token"}

	response, err := c.GetFilterChoices("job-title", "", 100, 0)
	if err != nil {
		t.Fatalf("GetFilterChoices returned error: %v", err)
	}
	if len(response.Results) != 2 {
		t.Fatalf("got %d results, want 2", len(response.Results))
	}
	if response.Results[0].Label != "Management Occupations" {
		t.Errorf("label = %q", response.Results[0].Label)
	}
	if response.Results[0].ParentID != nil {
		t.Errorf("parent_id = %v, want nil for a root choice", *response.Results[0].ParentID)
	}
	if response.Results[1].ParentID == nil || *response.Results[1].ParentID != "0" {
		t.Errorf("parent_id = %v, want 0", response.Results[1].ParentID)
	}
	if response.Results[0].NumDescendants != 476 {
		t.Errorf("num_descendants = %d, want 476", response.Results[0].NumDescendants)
	}
	if response.Meta.Count != 4123 {
		t.Errorf("meta.count = %d, want 4123", response.Meta.Count)
	}
}

// The API 404s both for an unknown filter and for one absent from the
// requested workspace; the error has to reach the caller unchanged.
func TestGetFilterChoicesSurfacesNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":{"status":404,"title":"The resource requested was not found.","error_code":140401}}`))
	}))
	defer server.Close()

	c := Client{Client: server.Client(), BaseURL: server.URL, Token: "test-token"}

	_, err := c.GetFilterChoices("nope", "", 10, 0)
	if err == nil {
		t.Fatal("expected an error for a 404 response")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, want it to carry the API's message", err)
	}
}

func TestSearchFilterChoicesSendsExpectedRequest(t *testing.T) {
	var gotPath string
	var gotQuery url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(SearchFilterChoicesResponse{})
	}))
	defer server.Close()

	c := Client{Client: server.Client(), BaseURL: server.URL, Token: "test-token"}

	if _, err := c.SearchFilterChoices("job-title", "nurse", testWorkspaceID, 100, 0); err != nil {
		t.Fatalf("SearchFilterChoices returned error: %v", err)
	}

	if want := "/api/v1/filters/job-title/choices/search/"; gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
	if got := gotQuery.Get("q"); got != "nurse" {
		t.Errorf("q = %q, want %q", got, "nurse")
	}
	if got := gotQuery.Get("limit"); got != "100" {
		t.Errorf("limit = %q, want %q", got, "100")
	}
	if got := gotQuery.Get("workspace_id"); got != testWorkspaceID {
		t.Errorf("workspace_id = %q, want %q", got, testWorkspaceID)
	}
}

func TestSearchFilterChoicesDecodesHighlights(t *testing.T) {
	body := `{
	  "results": [
	    {"id": "18873", "label": "Obstetrics Nurse (OB Nurse)", "parent_id": "1016", "num_children": 0, "num_descendants": 0,
	     "matches": [{"field": "label", "query_term": "nurse", "matched_text": "Nurse", "start": 11, "end": 16}]}
	  ],
	  "meta": {"count": 128}
	}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()

	c := Client{Client: server.Client(), BaseURL: server.URL, Token: "test-token"}

	response, err := c.SearchFilterChoices("job-title", "nurse", "", 100, 0)
	if err != nil {
		t.Fatalf("SearchFilterChoices returned error: %v", err)
	}
	if len(response.Results) != 1 {
		t.Fatalf("got %d results, want 1", len(response.Results))
	}
	match := response.Results[0].Matches
	if len(match) != 1 || match[0].Field != "label" || match[0].Start != 11 || match[0].End != 16 {
		t.Errorf("matches = %+v, want one label highlight spanning 11-16", match)
	}
	if response.Meta.Count != 128 {
		t.Errorf("meta.count = %d, want 128", response.Meta.Count)
	}
}
