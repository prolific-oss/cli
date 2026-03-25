package requirement

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/model"
	"github.com/prolific-oss/cli/ui"
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
	var content strings.Builder
	content.WriteString(fmt.Sprintln(ui.RenderHeading(req.Title())))

	content.WriteString(fmt.Sprintf("ID:                 %s\n", req.ID))
	content.WriteString(fmt.Sprintf("CLS (_cls):         %s\n", req.Cls))
	content.WriteString(fmt.Sprintf("Category:           %s\n", req.Category))
	if req.Subcategory != nil {
		content.WriteString(fmt.Sprintf("Subcategory:        %s\n", req.Subcategory))
	}

	content.WriteString(ui.RenderSectionMarker())

	content.WriteString(fmt.Sprintln(ui.RenderHeading("Query")))
	content.WriteString(fmt.Sprintf("ID:                 %s\n", req.Query.ID))
	content.WriteString(fmt.Sprintf("Question:           %s\n", req.Query.Question))
	content.WriteString(fmt.Sprintf("Title:              %s\n", req.Query.Title))
	content.WriteString(fmt.Sprintf("Description:        %s\n", req.Query.Description))

	content.WriteString(ui.RenderSectionMarker())

	content.WriteString(fmt.Sprintln(ui.RenderHeading("Attributes")))
	for _, attribute := range req.Attributes {
		content.WriteString(fmt.Sprintf("Name:               %v\n", attribute.Name))
		content.WriteString(fmt.Sprintf("Label:              %v\n", attribute.Label))
		content.WriteString(fmt.Sprintf("Index:              %v\n", attribute.Index))
		content.WriteString(fmt.Sprintf("Value:              %v\n", attribute.Value))
		content.WriteString("\n")
	}

	return fmt.Sprintln(content.String())
}
