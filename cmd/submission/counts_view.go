package submission

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/model"
)

// CountsView is a two-level bubbletea model for submission counts drill-down.
// Level 1 shows submission counts by status; Level 2 shows submissions filtered
// to the selected status.
type CountsView struct {
	countsList     list.Model
	submissionList *list.Model
	selectedSub    *model.Submission
	drillStatus    string
	studyID        string
	client         client.API
	submissions    []model.Submission
	totalCount     int
	fetched        bool
	loading        bool
	err            error
}

// NewCountsView creates a new CountsView from submission count items.
func NewCountsView(items []list.Item, studyID string, c client.API) CountsView {
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Submission counts"
	return CountsView{
		countsList: l,
		studyID:    studyID,
		client:     c,
	}
}

type submissionsFetchedMsg struct {
	submissions []model.Submission
	totalCount  int
	err         error
}

func fetchSubmissions(c client.API, studyID string) tea.Cmd {
	return func() tea.Msg {
		resp, err := c.GetSubmissions(studyID, client.DefaultRecordLimit, client.DefaultRecordOffset)
		if err != nil {
			return submissionsFetchedMsg{err: err}
		}
		totalCount := len(resp.Results)
		if resp.JSONAPIMeta != nil {
			totalCount = resp.Meta.Count
		}
		return submissionsFetchedMsg{
			submissions: resp.Results,
			totalCount:  totalCount,
		}
	}
}

// Init implements tea.Model.
func (cv CountsView) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (cv CountsView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case submissionsFetchedMsg:
		cv.loading = false
		if msg.err != nil {
			cv.err = msg.err
			return cv, tea.Quit
		}
		cv.submissions = msg.submissions
		cv.totalCount = msg.totalCount
		cv.fetched = true
		cv.buildSubmissionList()
		return cv, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return cv, tea.Quit

		case "esc":
			if cv.submissionList != nil {
				cv.submissionList = nil
				cv.drillStatus = ""
				return cv, nil
			}
			return cv, tea.Quit

		case "enter":
			if cv.submissionList != nil {
				i, ok := cv.submissionList.SelectedItem().(model.Submission)
				if ok {
					cv.selectedSub = &i
				}
				return cv, tea.Quit
			}

			i, ok := cv.countsList.SelectedItem().(model.SubmissionCountItem)
			if ok {
				cv.drillStatus = i.StatusKey
				if cv.fetched {
					cv.buildSubmissionList()
					return cv, nil
				}
				cv.loading = true
				return cv, fetchSubmissions(cv.client, cv.studyID)
			}
		}

	case tea.WindowSizeMsg:
		h, v := lipgloss.NewStyle().GetFrameSize()
		w, ht := msg.Width-h, msg.Height-v
		cv.countsList.SetSize(w, ht)
		if cv.submissionList != nil {
			cv.submissionList.SetSize(w, ht)
		}
	}

	var cmd tea.Cmd
	if cv.submissionList != nil {
		*cv.submissionList, cmd = cv.submissionList.Update(msg)
	} else {
		cv.countsList, cmd = cv.countsList.Update(msg)
	}
	return cv, cmd
}

// View implements tea.Model.
func (cv CountsView) View() string {
	if cv.selectedSub != nil {
		return RenderSubmission(*cv.selectedSub)
	}
	if cv.err != nil {
		return fmt.Sprintf("Error: %s\n", cv.err)
	}
	if cv.loading {
		return "Fetching submissions...\n"
	}
	if cv.submissionList != nil {
		return cv.submissionList.View()
	}
	return cv.countsList.View()
}

func (cv *CountsView) buildSubmissionList() {
	filtered := FilterSubmissionsByStatus(cv.submissions, cv.drillStatus)
	var items []list.Item
	for _, sub := range filtered {
		items = append(items, sub)
	}
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)

	title := fmt.Sprintf("Submissions - %s", cv.drillStatus)
	if cv.totalCount > len(cv.submissions) {
		title += fmt.Sprintf(" (showing from first %d of %d total submissions)", len(cv.submissions), cv.totalCount)
	}
	l.Title = title

	cv.submissionList = &l
}

// FilterSubmissionsByStatus filters submissions to those matching the given status key.
func FilterSubmissionsByStatus(submissions []model.Submission, status string) []model.Submission {
	status = strings.ToUpper(status)
	var filtered []model.Submission
	for _, sub := range submissions {
		if strings.ToUpper(sub.Status) == status {
			filtered = append(filtered, sub)
		}
	}
	return filtered
}

// RenderSubmission produces a detailed view of a submission.
func RenderSubmission(sub model.Submission) string {
	content := fmt.Sprintf("Participant: %s\n", sub.ParticipantID)
	content += fmt.Sprintf("ID:          %s\n", sub.ID)
	content += fmt.Sprintf("Status:      %s\n", sub.Status)
	content += fmt.Sprintf("Study Code:  %s\n", sub.StudyCode)
	content += fmt.Sprintf("Started At:  %s\n", sub.StartedAt.Format("2006-01-02 15:04:05"))
	if !sub.CompletedAt.IsZero() {
		content += fmt.Sprintf("Completed:   %s\n", sub.CompletedAt.Format("2006-01-02 15:04:05"))
	}
	content += fmt.Sprintf("Time Taken:  %ds\n", sub.TimeTaken)
	content += fmt.Sprintf("Reward:      %d\n", sub.Reward)
	return content
}
