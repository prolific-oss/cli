package feedback

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/prolific-oss/cli/model"
	"github.com/prolific-oss/cli/ui"
)

type feedbackListItem struct {
	feedback model.StudyFeedback
}

func (i feedbackListItem) FilterValue() string {
	return strings.Join([]string{
		i.feedback.ParticipantID,
		formatOptionalString(i.feedback.Category),
		formatOptionalString(i.feedback.Text),
	}, " ")
}

func (i feedbackListItem) Title() string {
	return i.feedback.ParticipantID
}

func (i feedbackListItem) Description() string {
	return fmt.Sprintf(
		"%s · clarity %s · ease %s · fairness %s",
		formatOptionalString(i.feedback.Category),
		formatRating(i.feedback.Ratings.Clarity),
		formatRating(i.feedback.Ratings.Ease),
		formatRating(i.feedback.Ratings.Fairness),
	)
}

// InteractiveRenderer displays feedback in a searchable Bubbletea list.
type InteractiveRenderer struct{}

// Render launches the TUI for the supplied feedback records.
func (r *InteractiveRenderer) Render(records []model.StudyFeedback, _ io.Writer) error {
	items := make([]list.Item, 0, len(records))
	for _, record := range records {
		items = append(items, feedbackListItem{feedback: record})
	}

	lv := ListView{
		List: list.New(items, list.NewDefaultDelegate(), 0, 0),
	}
	lv.List.Title = "Study feedback"

	program := tea.NewProgram(lv)
	if _, err := program.Run(); err != nil {
		return fmt.Errorf("cannot render study feedback: %s", err)
	}

	return nil
}

// ListView presents a searchable list of participant feedback.
type ListView struct {
	List     list.Model
	Feedback *model.StudyFeedback
}

// Init initialises the view.
func (lv ListView) Init() tea.Cmd {
	return nil
}

// Update handles navigation and selection.
func (lv ListView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return lv, tea.Quit
		}

		if msg.String() == "enter" {
			item, ok := lv.List.SelectedItem().(feedbackListItem)
			if ok {
				feedback := item.feedback
				lv.Feedback = &feedback
			}
			return lv, tea.Quit
		}
	case tea.WindowSizeMsg:
		horizontal, vertical := lipgloss.NewStyle().GetFrameSize()
		lv.List.SetSize(msg.Width-horizontal, msg.Height-vertical)
	}

	var cmd tea.Cmd
	lv.List, cmd = lv.List.Update(msg)
	return lv, cmd
}

// View renders the list or the selected feedback record.
func (lv ListView) View() string {
	if lv.Feedback == nil {
		return lv.List.View()
	}

	var content strings.Builder
	content.WriteString(fmt.Sprintln(ui.RenderHeading("Participant feedback")))
	content.WriteString(fmt.Sprintf("Participant: %s\n", lv.Feedback.ParticipantID))
	content.WriteString(fmt.Sprintf("Category:    %s\n", formatOptionalString(lv.Feedback.Category)))
	content.WriteString(fmt.Sprintf("Clarity:     %s\n", formatRating(lv.Feedback.Ratings.Clarity)))
	content.WriteString(fmt.Sprintf("Ease:        %s\n", formatRating(lv.Feedback.Ratings.Ease)))
	content.WriteString(fmt.Sprintf("Fairness:    %s\n", formatRating(lv.Feedback.Ratings.Fairness)))
	content.WriteString(fmt.Sprintf("Text:        %s\n", formatOptionalString(lv.Feedback.Text)))
	return content.String()
}
