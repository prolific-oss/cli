package ui_test

import (
	"testing"

	"github.com/prolific-oss/cli/ui"
	"github.com/stretchr/testify/assert"
)

func TestFieldHighlightsRender(t *testing.T) {
	hl := ui.RenderHighlightedText

	tests := []struct {
		name  string
		text  string
		spans []ui.HighlightSpan
		want  string
	}{
		{name: "no spans", text: "Software developers", want: "Software developers"},
		{name: "empty text", text: "", spans: []ui.HighlightSpan{{Start: 0, End: 3}}, want: ""},
		{name: "single span", text: "Software developers", spans: []ui.HighlightSpan{{Start: 9, End: 19}}, want: "Software " + hl("developers")},
		{name: "span at start", text: "Age", spans: []ui.HighlightSpan{{Start: 0, End: 3}}, want: hl("Age")},
		{name: "multiple spans", text: "a b c", spans: []ui.HighlightSpan{{Start: 0, End: 1}, {Start: 4, End: 5}}, want: hl("a") + " b " + hl("c")},
		{name: "unsorted overlapping spans merged", text: "abcdef", spans: []ui.HighlightSpan{{Start: 2, End: 4}, {Start: 0, End: 3}}, want: hl("abcd") + "ef"},
		{name: "adjacent spans merged", text: "abcdef", spans: []ui.HighlightSpan{{Start: 0, End: 2}, {Start: 2, End: 4}}, want: hl("abcd") + "ef"},
		{name: "out of range clamped", text: "abc", spans: []ui.HighlightSpan{{Start: -2, End: 10}}, want: hl("abc")},
		{name: "inverted span ignored", text: "abc", spans: []ui.HighlightSpan{{Start: 2, End: 1}}, want: "abc"},
		// Offsets are code points, not bytes: "é" is two bytes but one code point.
		{name: "code point offsets", text: "Café owner", spans: []ui.HighlightSpan{{Start: 5, End: 10}}, want: "Café " + hl("owner")},
		{name: "multi-byte inside span", text: "Zoë Jones", spans: []ui.HighlightSpan{{Start: 0, End: 3}}, want: hl("Zoë") + " Jones"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := ui.FieldHighlights{"title": tt.spans}
			assert.Equal(t, tt.want, h.Render("title", tt.text))
		})
	}
}

func TestFieldHighlightsRenderIgnoresOtherFields(t *testing.T) {
	h := ui.FieldHighlights{"question": {{Start: 0, End: 3}}}
	assert.Equal(t, "Age", h.Render("title", "Age"))
}

func TestFieldHighlightsRenderStyled(t *testing.T) {
	bold := func(s string) string { return "<b>" + s + "</b>" }
	hl := ui.RenderHighlightedText

	tests := []struct {
		name  string
		text  string
		spans []ui.HighlightSpan
		want  string
	}{
		{name: "no spans styles whole text", text: "Job title", want: "<b>Job title</b>"},
		{name: "plain segments styled separately", text: "Software development experience", spans: []ui.HighlightSpan{{Start: 9, End: 20}}, want: "<b>Software </b>" + hl("development") + "<b> experience</b>"},
		{name: "span covering whole text has no plain segments", text: "Age", spans: []ui.HighlightSpan{{Start: 0, End: 3}}, want: hl("Age")},
		{name: "empty text untouched", text: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := ui.FieldHighlights{"title": tt.spans}
			assert.Equal(t, tt.want, h.RenderStyled("title", tt.text, bold))
		})
	}
}

func TestFieldHighlightsNilMap(t *testing.T) {
	var h ui.FieldHighlights
	assert.Equal(t, "Age", h.Render("title", "Age"))
}
