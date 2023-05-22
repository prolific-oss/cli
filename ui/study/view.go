package study

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/prolific-oss/prolificli/client"
	"github.com/prolific-oss/prolificli/model"
	"github.com/prolific-oss/prolificli/ui"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

// ListView is responsible for presenting a list view to the user.
type ListView struct {
	List    list.Model
	Studies map[string]model.Study
	Study   *model.Study
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
		return docStyle.Render(RenderStudy(*lv.Study))
	}
	return docStyle.Render(lv.List.View())
}

// RenderStudy will produce a detailed view of the selected study.
func RenderStudy(study model.Study) string {
	content := fmt.Sprintln(ui.RenderHeading(study.Name))
	content += fmt.Sprintf("%s\n\n", study.Desc)
	content += fmt.Sprintf("ID:                        %s\n", study.ID)
	content += fmt.Sprintf("Status:                    %s\n", study.Status)
	content += fmt.Sprintf("Type:                      %s\n", study.StudyType)
	content += fmt.Sprintf("Total cost:                %s\n", ui.RenderMoney((study.TotalCost/100), study.GetCurrencyCode()))
	content += fmt.Sprintf("Reward:                    %s\n", ui.RenderMoney((study.Reward/100), study.GetCurrencyCode()))
	content += fmt.Sprintf("Hourly rate:               %s\n", ui.RenderMoney((study.AverageRewardPerHour/100), study.GetCurrencyCode()))
	content += fmt.Sprintf("Estimated completion time: %d\n", study.EstimatedCompletionTime)
	content += fmt.Sprintf("Maximum allowed time:      %d\n", study.MaximumAllowedTime)
	content += fmt.Sprintf("Study URL:                 %s\n", study.ExternalStudyURL)
	content += fmt.Sprintf("Places taken:              %d\n", study.PlacesTaken)
	content += fmt.Sprintf("Available places:          %d\n", study.TotalAvailablePlaces)
	content += fmt.Sprintf("\n%s\n\n", ui.RenderSectionMarker())

	content += fmt.Sprintln(ui.RenderHeading("Eligibility requirements"))

	erCount := 0
	erContent := ""
	for _, er := range study.EligibilityRequirements {
		if er.Question.Title != "" {
			erContent += fmt.Sprintf("- %s\n", er.Question.Title)
			erCount++
		}
	}

	if erCount == 0 {
		content += fmt.Sprintln("No eligibility requirements are defined for this study.")
	} else {
		content += erContent
	}

	content += fmt.Sprintf("\n%s\n\n", ui.RenderSectionMarker())

	content += fmt.Sprintf("View study in the application: https://app.prolific.co/researcher/studies/%s", study.ID)

	return content
}
