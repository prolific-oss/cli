package survey

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

// SurveyListView is responsible for presenting a list view to the user.
type SurveyListView struct {
	List    list.Model
	Surveys map[string]model.Survey
	Survey  *model.Survey
	Client  client.API
}

// Init will initialise the view.
func (lv SurveyListView) Init() tea.Cmd {
	return nil
}

// Update will update the view.
func (lv SurveyListView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return lv, tea.Quit
		}

		if msg.String() == "enter" {
			i, ok := lv.List.SelectedItem().(model.SurveyListItem)
			if ok {
				s := i.Survey
				lv.Survey = &s
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
func (lv SurveyListView) View() string {
	if lv.Survey != nil {
		return renderSurveyString(*lv.Survey)
	}
	return lv.List.View()
}

// renderSurveyString produces a detailed view of the selected survey as a string.
func renderSurveyString(s model.Survey) string {
	var content strings.Builder
	content.WriteString(fmt.Sprintln(ui.RenderHeading(s.Title)))

	content.WriteString(fmt.Sprintf("ID:            %v\n", s.ID))
	content.WriteString(fmt.Sprintf("Researcher:    %v\n", s.ResearcherID))
	content.WriteString(fmt.Sprintf("Date Created:  %v\n", s.DateCreated.Format("2006-01-02 15:04:05")))
	content.WriteString(fmt.Sprintf("Date Modified: %v\n", s.DateModified.Format("2006-01-02 15:04:05")))

	if len(s.Sections) > 0 {
		content.WriteString(ui.RenderSectionMarker())
		for i, section := range s.Sections {
			content.WriteString(fmt.Sprintf("Section: %s\n", section.Title))
			for _, q := range section.Questions {
				content.WriteString(fmt.Sprintf("  Question: %s (%s)\n", q.Title, q.Type))
				for _, a := range q.Answers {
					content.WriteString(fmt.Sprintf("    - %s\n", a.Value))
				}
			}
			if i < len(s.Sections)-1 {
				content.WriteString("\n")
			}
		}
	}

	if len(s.Questions) > 0 {
		content.WriteString(ui.RenderSectionMarker())
		for _, q := range s.Questions {
			content.WriteString(fmt.Sprintf("  Question: %s (%s)\n", q.Title, q.Type))
			for _, a := range q.Answers {
				content.WriteString(fmt.Sprintf("    - %s\n", a.Value))
			}
		}
	}

	if len(s.Sections) == 0 && len(s.Questions) == 0 {
		content.WriteString("\nNo questions defined\n")
	}

	return content.String()
}
