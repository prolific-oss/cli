package agentenv

import (
	"testing"
)

func TestDetected(t *testing.T) {
	// Every env var Detected looks at. Each subtest zeroes all of them first
	// so it's isolated from whatever environment its running in (e.g. running
	// inside an AI agent, or a local dev machine) has set.
	knownVars := []string{"CLAUDE_CODE", "ANTIGRAVITY_AGENT", "AI_AGENT", "LLM_AGENT"}

	tests := []struct {
		name string
		env  map[string]string
		want string
	}{
		{
			name: "no agent detected",
			env:  map[string]string{},
			want: "",
		},
		{
			name: "claude code, branded slug regardless of value",
			env:  map[string]string{"CLAUDE_CODE": "1"},
			want: "claude-code",
		},
		{
			name: "antigravity, branded slug regardless of value",
			env:  map[string]string{"ANTIGRAVITY_AGENT": "ignored-value"},
			want: "antigravity",
		},
		{
			name: "generic AI_AGENT forwards its value",
			env:  map[string]string{"AI_AGENT": "cursor"},
			want: "cursor",
		},
		{
			// tricky scenario, decision for branded
			name: "branded wins over generic (precedence)",
			env:  map[string]string{"CLAUDE_CODE": "1", "AI_AGENT": "cursor"},
			want: "claude-code",
		},
		{
			name: "empty string treated as unset",
			env:  map[string]string{"CLAUDE_CODE": ""},
			want: "",
		},
		{
			name: "invalid generic value skipped, falls through to next source",
			env:  map[string]string{"AI_AGENT": "bad\nvalue", "LLM_AGENT": "gemma4"},
			want: "gemma4",
		},
		{
			name: "valid non-ASCII generic value forwarded unchanged",
			env:  map[string]string{"AI_AGENT": "claudé-日本"},
			want: "claudé-日本",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, k := range knownVars {
				t.Setenv(k, "") // isolate from ambient env vars
			}
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			if got := Detected(); got != tt.want {
				t.Fatalf("Detected() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidHeaderValue(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{"plain ascii", "claude-code", true},
		{"empty string", "", true},
		{"tab is allowed", "a\tb", true},
		{"utf-8 multibyte", "日本語", true},
		{"high byte 0xFF", "\xff", true},
		{"carriage return", "a\rb", false},
		{"line feed", "a\nb", false},
		{"null byte", "a\x00b", false},
		{"del", "a\x7fb", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validHeaderValue(tt.value); got != tt.want {
				t.Fatalf("validHeaderValue(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}
