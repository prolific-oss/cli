package filters

import (
	"testing"

	"github.com/acarl005/stripansi"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/prolific-oss/cli/model"
	"github.com/prolific-oss/cli/ui"
	"github.com/stretchr/testify/assert"
)

// withColour forces a colour profile for the duration of a test so that
// highlight styling is emitted even though tests have no TTY.
func withColour(t *testing.T) {
	previous := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(previous) })
}

func TestHighlightText(t *testing.T) {
	hl := ui.RenderHighlightedText

	tests := []struct {
		name  string
		text  string
		spans []HighlightSpan
		want  string
	}{
		{name: "no spans", text: "Software developers", want: "Software developers"},
		{name: "empty text", text: "", spans: []HighlightSpan{{0, 3}}, want: ""},
		{name: "single span", text: "Software developers", spans: []HighlightSpan{{9, 19}}, want: "Software " + hl("developers")},
		{name: "span at start", text: "Age", spans: []HighlightSpan{{0, 3}}, want: hl("Age")},
		{name: "multiple spans", text: "a b c", spans: []HighlightSpan{{0, 1}, {4, 5}}, want: hl("a") + " b " + hl("c")},
		{name: "unsorted overlapping spans merged", text: "abcdef", spans: []HighlightSpan{{2, 4}, {0, 3}}, want: hl("abcd") + "ef"},
		{name: "adjacent spans merged", text: "abcdef", spans: []HighlightSpan{{0, 2}, {2, 4}}, want: hl("abcd") + "ef"},
		{name: "out of range clamped", text: "abc", spans: []HighlightSpan{{-2, 10}}, want: hl("abc")},
		{name: "inverted span ignored", text: "abc", spans: []HighlightSpan{{2, 1}}, want: "abc"},
		// Offsets are code points, not bytes: "é" is two bytes but one code point.
		{name: "code point offsets", text: "Café owner", spans: []HighlightSpan{{5, 10}}, want: "Café " + hl("owner")},
		{name: "multi-byte inside span", text: "Zoë Jones", spans: []HighlightSpan{{0, 3}}, want: hl("Zoë") + " Jones"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, HighlightText(tt.text, tt.spans))
		})
	}
}

func TestRenderSearchResultMinimal(t *testing.T) {
	r := model.FilterSearchResult{
		FilterID: "handedness",
		Title:    "Handedness",
		Type:     "select",
		DataType: "ChoiceID",
	}

	out := stripansi.Strip(RenderSearchResult(3, r))

	assert.Equal(t, "3. Handedness\n   select · ChoiceID\n   Filter ID    handedness\n", out)
}

func TestRenderSearchResultCategoryAndSubcategory(t *testing.T) {
	category, subcategory := "Employment", "Occupation"
	r := model.FilterSearchResult{
		FilterID:    "job-title",
		Title:       "Job title",
		Type:        "select",
		DataType:    "ChoiceID",
		Category:    &category,
		Subcategory: &subcategory,
	}

	out := stripansi.Strip(RenderSearchResult(1, r))
	assert.Contains(t, out, "   select · ChoiceID · Employment / Occupation\n")
}

func TestRenderSearchResultRangeWithMissingBound(t *testing.T) {
	r := model.FilterSearchResult{
		FilterID: "age",
		Title:    "Age",
		Type:     "range",
		DataType: "integer",
		Min:      18,
	}

	out := stripansi.Strip(RenderSearchResult(1, r))
	assert.Contains(t, out, "   Range        18 to -\n")
}

func TestRenderSearchResultMatchedChoicesNotTruncated(t *testing.T) {
	r := model.FilterSearchResult{
		FilterID: "job-title",
		Title:    "Job title",
		Type:     "select",
		DataType: "ChoiceID",
		MatchedChoices: &model.FilterMatchingChoices{
			Matched:   1,
			Truncated: false,
			Results: []model.FilterChoiceSearchResult{
				{ID: "7", Label: "Nurse"},
			},
		},
	}

	out := stripansi.Strip(RenderSearchResult(1, r))
	assert.Contains(t, out, "                7  Nurse\n")
	assert.NotContains(t, out, "more matching")
	assert.NotContains(t, out, "nested")
	// No num_choices means no Choices line, even with a matched preview.
	assert.NotContains(t, out, "Choices")
}

func TestRenderSearchResultAppliesHighlights(t *testing.T) {
	withColour(t)

	question := "What is your job title?"
	r := model.FilterSearchResult{
		FilterID: "job-title",
		Title:    "Job title",
		Question: &question,
		Type:     "select",
		DataType: "ChoiceID",
		Match: model.FilterSearchMatch{
			Fields: []string{"question", "choices"},
			Highlights: []model.FilterSearchHighlight{
				{Field: "question", QueryTerm: "job", MatchedText: "job", Start: 13, End: 16},
			},
		},
		MatchedChoices: &model.FilterMatchingChoices{
			Matched: 1,
			Results: []model.FilterChoiceSearchResult{
				{ID: "100", Label: "Software developers", Match: model.FilterChoiceMatch{Highlights: []model.FilterSearchHighlight{
					{Field: "label", QueryTerm: "developers", MatchedText: "developers", Start: 9, End: 19},
				}}},
			},
		},
	}

	out := RenderSearchResult(1, r)

	assert.Contains(t, out, "Question     What is your "+ui.RenderHighlightedText("job")+" title?")
	assert.Contains(t, out, "  Software "+ui.RenderHighlightedText("developers"))
	// Title has no highlight for this match, so it must be rendered as a plain heading.
	assert.Contains(t, out, ui.RenderHeading("Job title")+"\n")
	assert.NotContains(t, out, ui.RenderHighlightedText("Job"))
}

func TestRenderSearchResultHighlightedTitleKeepsHeadingStyle(t *testing.T) {
	withColour(t)

	r := model.FilterSearchResult{
		FilterID: "software-development-experience",
		Title:    "Software development experience",
		Type:     "range",
		DataType: "integer",
		Match: model.FilterSearchMatch{
			Fields: []string{"title"},
			Highlights: []model.FilterSearchHighlight{
				{Field: "title", QueryTerm: "developers", MatchedText: "development", Start: 9, End: 20},
			},
		},
	}

	out := RenderSearchResult(1, r)

	// Each plain segment is bolded separately so the text after the highlight
	// keeps the heading style.
	want := ui.RenderHeading("Software ") + ui.RenderHighlightedText("development") + ui.RenderHeading(" experience") + "\n"
	assert.Contains(t, out, want)
}

func TestRenderSearchHeader(t *testing.T) {
	tests := []struct {
		name         string
		shown, total int
		want         string
	}{
		{name: "all shown", shown: 2, total: 2, want: "Filters matching \"dev\"\nShowing 2 of 2 results\n\n"},
		{name: "single result", shown: 1, total: 1, want: "Filters matching \"dev\"\nShowing 1 of 1 result\n\n"},
		{name: "more available", shown: 25, total: 340, want: "Filters matching \"dev\"\nShowing 25 of 340 results. Use --limit or --all to see more\n\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, stripansi.Strip(RenderSearchHeader("dev", tt.shown, tt.total)))
		})
	}
}
