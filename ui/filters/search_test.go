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

func ptr[T any](v T) *T { return &v }

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
	r := model.FilterSearchResult{
		FilterID:    "job-title",
		Title:       "Job title",
		Type:        "select",
		DataType:    "ChoiceID",
		Category:    ptr("Employment"),
		Subcategory: ptr("Occupation"),
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

func TestRenderSearchResultMatchedChoices(t *testing.T) {
	r := model.FilterSearchResult{
		FilterID:   "job-title",
		Title:      "Job title",
		Type:       "select",
		DataType:   "ChoiceID",
		NumChoices: ptr(5),
		MatchedChoices: &model.FilterMatchingChoices{
			Matched:   4,
			Truncated: true,
			Results: []model.FilterChoiceSearchResult{
				{ID: "100", Label: "Software developers", NumDescendants: 3},
			},
		},
	}

	out := stripansi.Strip(RenderSearchResult(1, r))
	assert.Contains(t, out, "   Choices      5 (4 matching)\n")
	assert.Contains(t, out, "                100  Software developers  (+3 nested)\n")
	assert.Contains(t, out, "                …and 3 more matching choices\n")
}

func TestRenderSearchResultMatchedChoicesNotTruncated(t *testing.T) {
	r := model.FilterSearchResult{
		FilterID: "job-title",
		Title:    "Job title",
		Type:     "select",
		DataType: "ChoiceID",
		MatchedChoices: &model.FilterMatchingChoices{
			Matched: 1,
			Results: []model.FilterChoiceSearchResult{{ID: "7", Label: "Nurse"}},
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

	r := model.FilterSearchResult{
		FilterID: "job-title",
		Title:    "Job title",
		Question: ptr("What is your job title?"),
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

	want := ui.RenderHeading("Software ") + ui.RenderHighlightedText("development") + ui.RenderHeading(" experience") + "\n"
	assert.Contains(t, out, want)
}

func TestRenderSearchHeader(t *testing.T) {
	tests := []struct {
		name      string
		total     int
		truncated bool
		want      string
	}{
		{name: "all requested", total: 2, want: "Filters matching \"dev\"\n2 matching filters\n\n"},
		{name: "single result", total: 1, want: "Filters matching \"dev\"\n1 matching filter\n\n"},
		{name: "truncated", total: 340, truncated: true, want: "Filters matching \"dev\"\n340 matching filters. Use --limit or --all to see more\n\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, stripansi.Strip(RenderSearchHeader("dev", tt.total, tt.truncated)))
		})
	}
}

func TestRenderSearchFooter(t *testing.T) {
	assert.Equal(t, "\nShowing 25 records of 340\n", stripansi.Strip(RenderSearchFooter(25, 340)))
	assert.Equal(t, "\nShowing 1 record of 1\n", stripansi.Strip(RenderSearchFooter(1, 1)))
}

func TestRenderSearchRule(t *testing.T) {
	out := stripansi.Strip(RenderSearchRule())
	assert.Equal(t, 60, len([]rune(out))-1)
	assert.True(t, out[len(out)-1] == '\n')
}

func TestNewSearchListItems(t *testing.T) {
	results := []model.FilterSearchResult{
		{
			FilterID:    "job-title",
			Title:       "Job title",
			Type:        "select",
			DataType:    "ChoiceID",
			Category:    ptr("Employment"),
			Subcategory: ptr("Occupation"),
			Match:       model.FilterSearchMatch{Fields: []string{"title", "choices"}},
		},
		{FilterID: "age", Title: "Age", Type: "range", DataType: "integer"},
	}

	items := NewSearchListItems(results, 5)

	assert.Equal(t, []SearchListItem{
		{Rank: 5, FilterID: "job-title", Title: "Job title", Type: "select", DataType: "ChoiceID", Category: "Employment / Occupation", MatchedOn: "title, choices"},
		{Rank: 6, FilterID: "age", Title: "Age", Type: "range", DataType: "integer"},
	}, items)
}

func TestNewSearchListItemsEmpty(t *testing.T) {
	assert.Empty(t, NewSearchListItems(nil, 1))
	assert.NotNil(t, NewSearchListItems(nil, 1))
}
