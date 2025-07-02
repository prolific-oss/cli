package filter

import (
	"fmt"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/model"
	"github.com/benmatselby/prolificli/ui"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ListView is responsible for presenting a list view to the user.
type ListView struct {
	List    list.Model
	Filters map[string]model.Filter
	Filter  *model.Filter
	Client  client.API
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
			i, ok := lv.List.SelectedItem().(model.Filter)
			if ok {
				lv.Filter = &i
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
	if lv.Filter != nil {
		return RenderFilter(*lv.Filter)
	}
	return lv.List.View()
}

// RenderFilter will provide a more indepth view of the filter.
func RenderFilter(filter model.Filter) string {
	content := fmt.Sprintln(ui.RenderHeading(filter.Title()))

	content += fmt.Sprintf("ID:                %s\n", filter.FilterID)
	content += fmt.Sprintf("Filter ID:         %s\n", filter.FilterID)
	content += fmt.Sprintf("Title:             %s\n", filter.Title())
	content += fmt.Sprintf("Question:          %s\n", filter.Question)
	content += fmt.Sprintf("Description:       %s\n", filter.Description())

	return fmt.Sprintln(content)
}
