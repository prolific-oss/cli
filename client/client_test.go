package client

import (
	"net/http"
	"net/http/httptest"
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
func TestComposeUserAgent(t *testing.T) {
	knownVars := []string{"CLAUDE_CODE", "ANTIGRAVITY_AGENT", "AI_AGENT", "LLM_AGENT"}

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
			agentEnv: map[string]string{"CLAUDE_CODE": "1"},
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
	for _, k := range []string{"CLAUDE_CODE", "GEMINI_CLI", "AI_AGENT", "LLM_AGENT"} {
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
	for _, k := range []string{"CLAUDE_CODE", "ANTIGRAVITY_AGENT", "AI_AGENT", "LLM_AGENT"} {
		t.Setenv(k, "")
	}
	t.Setenv("CLAUDE_CODE", "1")

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
