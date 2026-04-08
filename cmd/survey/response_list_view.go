package survey

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/prolific-oss/cli/model"
	"github.com/prolific-oss/cli/ui"
)

// ResponseListView is responsible for presenting a response list view to the user.
type ResponseListView struct {
	List     list.Model
	Response *model.SurveyResponse
}

// Init will initialise the view.
func (lv ResponseListView) Init() tea.Cmd {
	return nil
}

// Update will update the view.
func (lv ResponseListView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return lv, tea.Quit
		}

		if msg.String() == "enter" {
			i, ok := lv.List.SelectedItem().(model.SurveyResponseListItem)
			if ok {
				r := i.SurveyResponse
				lv.Response = &r
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
func (lv ResponseListView) View() string {
	if lv.Response != nil {
		return renderResponseString(*lv.Response)
	}
	return lv.List.View()
}

// renderResponseString produces a detailed view of a survey response as a string.
func renderResponseString(r model.SurveyResponse) string {
	var content strings.Builder
	content.WriteString(fmt.Sprintln(ui.RenderHeading("Survey Response")))

	content.WriteString(fmt.Sprintf("ID:            %v\n", r.ID))
	content.WriteString(fmt.Sprintf("Participant:   %v\n", r.ParticipantID))
	content.WriteString(fmt.Sprintf("Submission:    %v\n", r.SubmissionID))
	content.WriteString(fmt.Sprintf("Date Created:  %v\n", r.DateCreated.Format("2006-01-02 15:04:05")))
	content.WriteString(fmt.Sprintf("Date Modified: %v\n", r.DateModified.Format("2006-01-02 15:04:05")))

	if len(r.Sections) > 0 {
		content.WriteString(ui.RenderSectionMarker())
		for i, section := range r.Sections {
			content.WriteString(fmt.Sprintf("Section: %s\n", section.SectionID))
			renderResponseQuestions(&content, section.Questions)
			if i < len(r.Sections)-1 {
				content.WriteString("\n")
			}
		}
	}

	if len(r.Questions) > 0 {
		content.WriteString(ui.RenderSectionMarker())
		renderResponseQuestions(&content, r.Questions)
	}

	if len(r.Sections) == 0 && len(r.Questions) == 0 {
		content.WriteString("\nNo answers recorded\n")
	}

	return content.String()
}

func renderResponseQuestions(content *strings.Builder, questions []model.SurveyQuestionResponse) {
	for _, q := range questions {
		content.WriteString(fmt.Sprintf("  Question: %s\n", q.QuestionTitle))
		for _, a := range q.Answers {
			content.WriteString(fmt.Sprintf("    - %s\n", a.Value))
		}
	}
}
