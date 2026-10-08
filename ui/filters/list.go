package filters

import (
	"fmt"

	"github.com/prolific-oss/cli/model"
)

// ListFields is the default column set for table and CSV output.
const ListFields = "FilterID,Title,Type,DataType,Choices"

// ListItem is a flattened catalogue filter for table and CSV output. Bounds
// are pre-formatted because the API types them loosely, and Choices is a count
// rather than the choices themselves, which a row cannot usefully hold.
type ListItem struct {
	FilterID    string
	Title       string
	Description string
	Question    string
	Type        string
	DataType    string
	Min         string
	Max         string
	Choices     int
}

// NewListItems flattens catalogue filters for table and CSV output, in the
// order the API returned them.
func NewListItems(filters []model.Filter) []ListItem {
	items := make([]ListItem, 0, len(filters))
	for _, f := range filters {
		items = append(items, ListItem{
			FilterID:    f.FilterID,
			Title:       f.Title(),
			Description: f.Description(),
			Question:    f.Question,
			Type:        f.Type,
			DataType:    f.DataType,
			Min:         formatBound(f.Min),
			Max:         formatBound(f.Max),
			Choices:     len(f.Choices),
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
