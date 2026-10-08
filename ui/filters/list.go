package filters

import (
	"fmt"

	"github.com/prolific-oss/cli/cmd/shared"
	"github.com/prolific-oss/cli/model"
)

// ListFields is the default column set for each format.
var ListFields = shared.FieldSet{
	CSV:   "FilterID,Title,Type,DataType,ChoicesTotal",
	Table: "FilterID,Title,Type",
}

// ListItem is a flattened catalogue filter for table and CSV output. Bounds
// are pre-formatted because the API types them loosely, and Choices is a count
// rather than the choices themselves, which a row cannot usefully hold.
type ListItem struct {
	FilterID     string
	Title        string
	Description  string
	Question     string
	Type         string
	DataType     string
	Min          string
	Max          string
	ChoicesTotal int
}

// NewListItems flattens catalogue filters for table and CSV output, in the
// order the API returned them.
func NewListItems(filters []model.Filter) []ListItem {
	items := make([]ListItem, 0, len(filters))
	for _, f := range filters {
		items = append(items, ListItem{
			FilterID:     f.FilterID,
			Title:        f.Title(),
			Description:  f.Description(),
			Question:     f.Question,
			Type:         f.Type,
			DataType:     f.DataType,
			Min:          formatBound(f.Min),
			Max:          formatBound(f.Max),
			ChoicesTotal: len(f.Choices),
		})
	}
	return items
}

func formatBound(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}
