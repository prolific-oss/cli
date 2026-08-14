package client

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prolific-oss/cli/version"
)

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
