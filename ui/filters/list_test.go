package filters_test

import (
	"testing"

	"github.com/prolific-oss/cli/model"
	uifilters "github.com/prolific-oss/cli/ui/filters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewListItemsFlattensFilters(t *testing.T) {
	items := uifilters.NewListItems([]model.Filter{
		{
			FilterID:          "age",
			FilterTitle:       "Age",
			FilterDescription: "Filter by age",
			Question:          "How old are you?",
			Type:              "range",
			DataType:          "integer",
			Min:               18,
			Max:               100,
		},
		{
			FilterID:    "handedness",
			FilterTitle: "Handedness",
			Type:        "select",
			DataType:    "string",
			Choices:     map[string]string{"1": "Right-handed", "2": "Left-handed"},
		},
	})

	require.Len(t, items, 2)

	assert.Equal(t, "age", items[0].FilterID)
	assert.Equal(t, "Age", items[0].Title)
	assert.Equal(t, "Filter by age", items[0].Description)
	assert.Equal(t, "How old are you?", items[0].Question)
	assert.Equal(t, "range", items[0].Type)
	assert.Equal(t, "integer", items[0].DataType)
	assert.Equal(t, "18", items[0].Min)
	assert.Equal(t, "100", items[0].Max)
	assert.Equal(t, 0, items[0].ChoicesTotal)

	assert.Equal(t, 2, items[1].ChoicesTotal)
}

// A range filter with no bounds, and a select filter with no range, must not
// leak Go's nil rendering into a table or CSV cell.
func TestNewListItemsRendersMissingBoundsAsEmpty(t *testing.T) {
	items := uifilters.NewListItems([]model.Filter{{FilterID: "handedness"}})

	require.Len(t, items, 1)
	assert.Empty(t, items[0].Min)
	assert.Empty(t, items[0].Max)
}

func TestNewListItemsHandlesNoFilters(t *testing.T) {
	assert.Empty(t, uifilters.NewListItems(nil))
}

// A table shows what identifies a filter; a CSV adds the data type and how
// many choices it has.
func TestListFieldsDifferByFormat(t *testing.T) {
	assert.Equal(t, "FilterID,Title,Type", uifilters.ListFields.Table)
	assert.Equal(t, "FilterID,Title,Type,DataType,ChoicesTotal", uifilters.ListFields.CSV)
}
