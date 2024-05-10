package study

import (
	"fmt"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/config"
	"github.com/benmatselby/prolificli/model"
	"github.com/benmatselby/prolificli/ui"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

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
		h, v := lipgloss.NewStyle().GetFrameSize()
		lv.List.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	lv.List, cmd = lv.List.Update(msg)
	return lv, cmd
}

// View will render the view.
func (lv ListView) View() string {
	if lv.Study != nil {
		return RenderStudy(*lv.Study)
	}
	return lv.List.View()
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

	content += ui.RenderSectionMarker()

	content += fmt.Sprintln(ui.RenderHeading("Submissions configuration"))

	content += fmt.Sprintf("Max submissions per participant: %v\n", study.SubmissionsConfig.MaxSubmissionsPerParticipant)
	content += fmt.Sprintf("Max concurrent submissions:      %v\n", study.SubmissionsConfig.MaxConcurrentSubmissions)

	content += ui.RenderSectionMarker()

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

	content += ui.RenderApplicationLink("study", GetStudyPath(study.ID))

	return content
}

// GetStudyPath returns the URL path to a study, agnostic of domain
func GetStudyPath(ID string) string {
	return fmt.Sprintf("researcher/studies/%s", ID)
}

// GetStudyURL returns the full URL to a study using configuration
func GetStudyURL(ID string) string {
	return fmt.Sprintf("%s/%s", config.GetApplicationURL(), GetStudyPath(ID))
}
