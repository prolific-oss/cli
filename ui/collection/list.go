package collection

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/model"
)

// InteractiveRenderer runs the bubbles UI framework to provide a rich
// UI experience for the user.
type InteractiveRenderer struct{}

// Render will render the list in an interactive manner.
func (r *InteractiveRenderer) Render(client client.API, collections client.ListCollectionsResponse, w io.Writer) error {
	var items []list.Item
	collectionMap := make(map[string]model.Collection)

	for _, collection := range collections.Results {
		items = append(items, collection)
		collectionMap[collection.ID] = collection
	}

	lv := ListView{
		List:        list.New(items, list.NewDefaultDelegate(), 0, 0),
		Collections: collectionMap,
		Client:      client,
	}
	lv.List.Title = "Collections"

	p := tea.NewProgram(lv)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("cannot render collections: %s", err)
	}

	return nil
}
