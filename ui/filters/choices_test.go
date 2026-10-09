package filters

import (
	"testing"

	"github.com/acarl005/stripansi"
	"github.com/prolific-oss/cli/model"
	"github.com/prolific-oss/cli/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func choice(id, label string, parentID *string, children, descendants int) model.FilterChoiceSearchResult {
	return model.FilterChoiceSearchResult{
		ID:             id,
		Label:          label,
		ParentID:       parentID,
		NumChildren:    children,
		NumDescendants: descendants,
	}
}

func TestNewChoiceListItemsFlattensChoices(t *testing.T) {
	items := NewChoiceListItems([]model.FilterChoiceSearchResult{
		choice("0", "Management Occupations", nil, 4, 476),
		choice("1016", "Registered Nurses", ptr("0"), 12, 12),
	})

	require.Len(t, items, 2)

	assert.Equal(t, ChoiceListItem{
		ID: "0", Label: "Management Occupations", ParentID: "", NumChildren: 4, NumDescendants: 476,
	}, items[0])
	assert.Equal(t, ChoiceListItem{
		ID: "1016", Label: "Registered Nurses", ParentID: "0", NumChildren: 12, NumDescendants: 12,
	}, items[1])
}

func TestNewChoiceListItemsEmpty(t *testing.T) {
	assert.Empty(t, NewChoiceListItems(nil))
	assert.NotNil(t, NewChoiceListItems(nil))
}

// A table shows the two columns that identify a choice; a CSV carries the
// hierarchy columns a spreadsheet can work with.
func TestChoiceFieldsDifferByFormat(t *testing.T) {
	assert.Equal(t, "ID,Label", ChoiceFields.Table)
	assert.Equal(t, "ID,Label,ParentID,NumChildren,NumDescendants", ChoiceFields.CSV)
}

func TestRenderChoicesHeader(t *testing.T) {
	tests := map[string]struct {
		query     string
		total     int
		truncated bool
		want      string
	}{
		"listing": {
			total: 4123,
			want:  "Choices in \"job-title\"\n4123 choices\n\n",
		},
		"searching": {
			query: "nurse",
			total: 128,
			want:  "Choices in \"job-title\" matching \"nurse\"\n128 matching choices\n\n",
		},
		"one match": {
			query: "nurse",
			total: 1,
			want:  "Choices in \"job-title\" matching \"nurse\"\n1 matching choice\n\n",
		},
		"truncated": {
			total:     4123,
			truncated: true,
			want:      "Choices in \"job-title\"\n4123 choices. Use --limit or --all to see more\n\n",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := stripansi.Strip(RenderChoicesHeader("job-title", tt.query, tt.total, tt.truncated))
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRenderNoChoices(t *testing.T) {
	assert.Equal(t, "No choices found in \"job-title\"\n", RenderNoChoices("job-title", ""))
	assert.Equal(t, "No choices found in \"job-title\" matching \"zzz\"\n", RenderNoChoices("job-title", "zzz"))
}

func TestRenderChoiceRowShowsIDLabelParentAndCounts(t *testing.T) {
	row := stripansi.Strip(RenderChoiceRow(choice("18873", "Obstetrics Nurse", ptr("1016"), 0, 0)))
	assert.Equal(t, "18873        Obstetrics Nurse  (parent 1016)\n", row)

	// num_descendants includes the direct children, so the nested count is
	// what is left below them: 476 descendants under 4 children is 472 deeper.
	root := stripansi.Strip(RenderChoiceRow(choice("0", "Management Occupations", nil, 4, 476)))
	assert.Equal(t, "0            Management Occupations  (4 children, +472 nested)\n", root)

	// A parent of leaves has as many descendants as children, so repeating the
	// count as "+12 nested" would say nothing.
	leafParent := stripansi.Strip(RenderChoiceRow(choice("1016", "Registered Nurses", ptr("0"), 12, 12)))
	assert.Equal(t, "1016         Registered Nurses  (parent 0, 12 children)\n", leafParent)

	// A leaf has nothing to say about its place beyond its parent.
	leaf := stripansi.Strip(RenderChoiceRow(choice("18873", "Obstetrics Nurse", ptr("1016"), 0, 0)))
	assert.Equal(t, "18873        Obstetrics Nurse  (parent 1016)\n", leaf)

	// A standalone choice carries no parenthetical at all.
	alone := stripansi.Strip(RenderChoiceRow(choice("0", "Nurses", nil, 0, 0)))
	assert.Equal(t, "0            Nurses\n", alone)
}

// The whole point of the search view is seeing why a choice matched, so it
// reuses the catalogue search's highlight rendering rather than its own.
func TestRenderChoiceRowHighlightsMatches(t *testing.T) {
	withColour(t)

	matched := model.FilterChoiceSearchResult{
		ID:    "18873",
		Label: "Obstetrics Nurse",
		Matches: []model.FilterSearchHighlight{
			{Field: "label", QueryTerm: "nurse", MatchedText: "Nurse", Start: 11, End: 16},
		},
	}

	assert.Contains(t, RenderChoiceRow(matched), "Obstetrics "+ui.RenderHighlightedText("Nurse"))
}

func TestRenderChoicesTableHeader(t *testing.T) {
	assert.Equal(t, "Choice ID    Label\n", stripansi.Strip(RenderChoicesTableHeader()))
}
