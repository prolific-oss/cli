package ui

import (
	"fmt"

	"github.com/benmatselby/prolificli/model"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

// ListView is responsible for presenting a list view to the user.
type ListView struct {
	List    list.Model
	Studies map[string]model.Study
	Study   *model.Study
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
			i, ok := lv.List.SelectedItem().(model.Study)
			if ok {
				lv.Study = &i
			}
			return lv, tea.Quit
		}

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		lv.List.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	lv.List, cmd = lv.List.Update(msg)
	return lv, cmd
}

// View will render the view.
func (lv ListView) View() string {
	if lv.Study != nil {
		return docStyle.Render(lv.RenderStudy())
	}
	return docStyle.Render(lv.List.View())
}

// RenderStudy will produce a detailed view of the selected study.
func (lv ListView) RenderStudy() string {
	content := fmt.Sprintln(RenderTitle(lv.Study.Name, lv.Study.Status))
	content += fmt.Sprintf("%s\n\n", lv.Study.Desc)
	content += fmt.Sprintf("Status:                    %s\n", lv.Study.Status)
	content += fmt.Sprintf("Total cost:                %.2f\n", float64(lv.Study.TotalCost)/100)
	content += fmt.Sprintf("Reward:                    %.2f\n", float64(lv.Study.Reward)/100)
	content += fmt.Sprintf("Hourly rate:               %.2f\n", float64(lv.Study.AverageRewardPerHour)/100)
	content += fmt.Sprintf("Estimated completion time: %d\n", lv.Study.EstimatedCompletionTime)
	content += fmt.Sprintf("Maximum allowed time:      %d\n", lv.Study.MaximumAllowedTime)
	content += fmt.Sprintf("Study URL:                 %s\n", lv.Study.ExternalStudyURL)
	content += fmt.Sprintf("Places taken:              %d\n", lv.Study.PlacesTaken)
	content += fmt.Sprintf("Available places:          %d\n", lv.Study.TotalAvailablePlaces)
	content += "\n---\n\n"

	if len(lv.Study.EligibilityRequirements) > 0 {
		content += fmt.Sprintln("Eligibility requirements:")
		for _, er := range lv.Study.EligibilityRequirements {
			content += fmt.Sprintf("  - %s\n", er.Question.Title)
		}
	}

	return content
}
