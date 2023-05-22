package requirement

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/prolific-oss/prolificli/client"
	"github.com/prolific-oss/prolificli/model"
	"github.com/prolific-oss/prolificli/ui"
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
		return lipgloss.NewStyle().Render(RenderRequirement(*lv.Requirement))
	}
	return lipgloss.NewStyle().Render(lv.List.View())
}

// RenderRequirement will provide a more indepth view of the requirement.
func RenderRequirement(req model.Requirement) string {
	content := fmt.Sprintln(ui.RenderHeading(req.Title()))
	content += fmt.Sprintf("ID:                 %s\n", req.Query.ID)
	content += fmt.Sprintf("Question:           %s\n", req.Query.Question)
	content += fmt.Sprintf("Title:              %s\n", req.Query.Title)
	content += fmt.Sprintf("Description:        %s\n", req.Query.Description)
	content += fmt.Sprintf("Category:           %s\n", req.Category)
	content += fmt.Sprintf("Subcategory:        %s\n", req.Subcategory)
	content += fmt.Sprintf("Type:               %s\n", req.RequirementType)

	return fmt.Sprintln(content)
}
