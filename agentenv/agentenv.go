// Package agentenv detects well-known environment variables set by AI
// coding agents (Claude Code, Antigravity, etc.) and resolves the driving
// agent's name so the CLI can advertise it in its User-Agent header, letting
// the Prolific API attribute requests to the agent that drove them.
package agentenv

import "os"

type agentSource struct {
	env  string // env var variable to check
	name string // the agent / model slug, empty is treated as 'forward value'
}

var sources = []agentSource{
	{"CLAUDE_CODE", "claude-code"},
	{"ANTIGRAVITY_AGENT", "antigravity"},
	{"AI_AGENT", ""},
	{"LLM_AGENT", ""},
}

// Detected returns the name of the AI agent driving the CLI, or "" if none is
// detected. Branded tools map to a fixed slug; generic vars forward their
// value verbatim, provided it contains no control characters or whitespace.
// Unset, empty, or malformed values yield "".
func Detected() string {
	for _, s := range sources {
		val := os.Getenv(s.env)
		if val == "" {
			continue
		}
		name := s.name
		if name == "" { // generic var: forward its value
			name = val
		}
		if !ValidHeaderValue(name) {
			continue
		}
		return name // first usable match wins
	}
	return ""
}

// ValidHeaderValue reports whether s is safe to embed as a single
// space-separated User-Agent token: no control characters, and no
// whitespace (which would split the token across multiple segments).
func ValidHeaderValue(s string) bool {
	for i := 0; i < len(s); i++ {
		b := s[i]
		if b < 0x20 || b == 0x7f || b == ' ' {
			return false
		}
	}
	return true
}
