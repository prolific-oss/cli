package filter

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/model"
	"github.com/prolific-oss/cli/ui"
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
	content += fmt.Sprintf("Type:              %s\n", filter.Type)
	content += fmt.Sprintf("Data Type:         %s\n", filter.DataType)
	content += fmt.Sprintf("Min:               %v\n", filter.Min)
	content += fmt.Sprintf("Max:               %v\n", filter.Max)

	if len(filter.Choices) > 0 {
		content += "Choices:\n"

		// Ensure ordering
		keys := make([]string, 0, len(filter.Choices))
		for k := range filter.Choices {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			content += fmt.Sprintf("  %s: %s\n", k, filter.Choices[k])
		}
	}

	return fmt.Sprintln(content)
}
