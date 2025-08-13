package filter_test

import (
	"strings"
	"testing"

	"github.com/acarl005/stripansi"
	"github.com/prolific-oss/cli/model"
	"github.com/prolific-oss/cli/ui/filter"
)

func TestRenderFilter(t *testing.T) {
	record := model.Filter{
		ID:                "id",
		FilterID:          "filter-id",
		FilterTitle:       "filter title",
		FilterDescription: "filter description",
		Question:          "filter question",
		Type:              "filter type",
		DataType:          "filter data type",
		Min:               1,
		Max:               10,
		Choices:           map[string]string{"choice1": "Choice 1", "choice2": "Choice 2"},
		SelectedValues:    []string{"choice1"},
		SelectedRange:     model.FilterRange{Lower: 1, Upper: 10},
	}

	actual := filter.RenderFilter(record)

	expected := `filtertitle
ID:filter-id
FilterID:filter-id
Title:filtertitle
Question:filterquestion
Description:filterdescription
Type:filtertype
DataType:filterdatatype
Min:1
Max:10
Choices:
choice1:Choice1
choice2:Choice2

`

	actual = stripansi.Strip(actual)
	actual = strings.ReplaceAll(actual, " ", "")
	expected = strings.ReplaceAll(expected, " ", "")

	if expected != actual {
		t.Fatalf("expected '%s', got '%s'", expected, actual)
	}
}
