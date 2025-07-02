package requirement

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
	List         list.Model
	Requirements map[string]model.Requirement
	Requirement  *model.Requirement
	Client       client.API
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
			i, ok := lv.List.SelectedItem().(model.Requirement)
			if ok {
				lv.Requirement = &i
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
	if lv.Requirement != nil {
		return RenderRequirement(*lv.Requirement)
	}
	return lv.List.View()
}

// RenderRequirement will provide a more indepth view of the requirement.
func RenderRequirement(req model.Requirement) string {
	content := fmt.Sprintln(ui.RenderHeading(req.Title()))

	content += fmt.Sprintf("ID:                 %s\n", req.ID)
	content += fmt.Sprintf("CLS (_cls):         %s\n", req.Cls)
	content += fmt.Sprintf("Category:           %s\n", req.Category)
	if req.Subcategory != nil {
		content += fmt.Sprintf("Subcategory:        %s\n", req.Subcategory)
	}

	content += ui.RenderSectionMarker()

	content += fmt.Sprintln(ui.RenderHeading("Query"))
	content += fmt.Sprintf("ID:                 %s\n", req.Query.ID)
	content += fmt.Sprintf("Question:           %s\n", req.Query.Question)
	content += fmt.Sprintf("Title:              %s\n", req.Query.Title)
	content += fmt.Sprintf("Description:        %s\n", req.Query.Description)

	content += ui.RenderSectionMarker()

	content += fmt.Sprintln(ui.RenderHeading("Attributes"))
	for _, attribute := range req.Attributes {
		content += fmt.Sprintf("Name:               %v\n", attribute.Name)
		content += fmt.Sprintf("Label:              %v\n", attribute.Label)
		content += fmt.Sprintf("Index:              %v\n", attribute.Index)
		content += fmt.Sprintf("Value:              %v\n", attribute.Value)
		content += "\n"
	}

	return fmt.Sprintln(content)
}
