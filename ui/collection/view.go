package collection

import (
	"fmt"
	"io"

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
		return renderCollectionString(*lv.Collection)
	}
	return lv.List.View()
}

// RenderCollection will write a detailed view of the selected collection to the provided writer.
func RenderCollection(collection *model.Collection, w io.Writer) error {
	content := fmt.Sprintln(ui.RenderHeading(collection.Name))
	content += fmt.Sprintf("ID:         %s\n", collection.ID)
	content += fmt.Sprintf("Created by: %s\n", collection.CreatedBy)
	content += fmt.Sprintf("Created at: %s\n", collection.CreatedAt.Format("2006-01-02 15:04:05"))

	if collection.TaskDetails != nil {
		content += "\nTask Details:\n"
		content += fmt.Sprintf("  Task Name:         %s\n", collection.TaskDetails.TaskName)
		content += fmt.Sprintf("  Task Introduction: %s\n", collection.TaskDetails.TaskIntroduction)
		content += fmt.Sprintf("  Task Steps:        %s\n", collection.TaskDetails.TaskSteps)
	}

	_, err := fmt.Fprint(w, content)
	return err
}

// renderCollectionString will produce a detailed view of the selected collection as a string.
func renderCollectionString(collection model.Collection) string {
	content := fmt.Sprintln(ui.RenderHeading(collection.Name))
	content += fmt.Sprintf("ID:         %s\n", collection.ID)
	content += fmt.Sprintf("Created by: %s\n", collection.CreatedBy)
	content += fmt.Sprintf("Created at: %s\n", collection.CreatedAt.Format("2006-01-02 15:04:05"))

	if collection.TaskDetails != nil {
		content += "\nTask Details:\n"
		content += fmt.Sprintf("  Task Name:         %s\n", collection.TaskDetails.TaskName)
		content += fmt.Sprintf("  Task Introduction: %s\n", collection.TaskDetails.TaskIntroduction)
		content += fmt.Sprintf("  Task Steps:        %s\n", collection.TaskDetails.TaskSteps)
	}

	return content
}
