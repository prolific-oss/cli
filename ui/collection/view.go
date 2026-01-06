package collection

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/model"
	"github.com/prolific-oss/cli/ui"
)

// ListView is responsible for presenting a list view to the user.
type ListView struct {
	List        list.Model
	Collections map[string]model.Collection
	Collection  *model.Collection
	Client      client.API
}

// Init will initialise the view.
func (lv ListView) Init() tea.Cmd {
	return nil
}

// Update will update the view.
func (lv ListView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return lv, tea.Quit
		}

		if msg.String() == "enter" {
			i, ok := lv.List.SelectedItem().(model.Collection)
			if ok {
				lv.Collection = &i
			}
			return lv, tea.Quit
		}

	case tea.WindowSizeMsg:
		h, v := lipgloss.NewStyle().GetFrameSize()
		lv.List.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	lv.List, cmd = lv.List.Update(msg)
	return lv, cmd
}

// View will render the view.
func (lv ListView) View() string {
	if lv.Collection != nil {
		return RenderCollection(*lv.Collection)
	}
	return lv.List.View()
}

// RenderCollection will produce a detailed view of the selected collection.
func RenderCollection(collection model.Collection) string {
	content := fmt.Sprintln(ui.RenderHeading(collection.Name))
	content += fmt.Sprintf("ID:         %s\n", collection.ID)
	content += fmt.Sprintf("Created by: %s\n", collection.CreatedBy)
	content += fmt.Sprintf("Created at: %s\n", collection.CreatedAt.Format("2006-01-02 15:04:05"))
	content += fmt.Sprintf("Item count: %d\n", collection.ItemCount)

	return content
}
