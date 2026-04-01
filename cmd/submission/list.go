package submission

import (
	"errors"
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/shared"
	"github.com/prolific-oss/cli/model"
	"github.com/prolific-oss/cli/ui"
	"github.com/spf13/cobra"
)

// defaultListFields is the default fields shown when the user has not specified --fields.
const defaultListFields = "ParticipantID,StartedAt,TimeTaken,StudyCode,Status"

// ListOptions is the options for the listing submissions command.
type ListOptions struct {
	Args   []string
	Fields string
	Output shared.OutputOptions
	Study  string
	Limit  int
	Offset int
}

// NewListCommand creates a new `submission list` command to give you details about
// your submissions for a study.
func NewListCommand(c client.API, w io.Writer) *cobra.Command {
	var opts ListOptions

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Provide details about your submissions, requires Study ID",
		Long: `List submissions for a given study

A published study will have submissions taken by the Prolific Participants. This
commands allows you to list those submissions.`,
		Example: `
You can list all the submissions for a given study
$ prolific submission list -s 63c123af913a974f87e8e7fc

You can use the paging options to limit the submissions returned, for example 5
$ prolific submission list -s 63c123af913a974f87e8e7fc -l 5

You can also offset the results, for example skipping 5
$ prolific submission list -s 63c123af913a974f87e8e7fc -l 5 -o 5

You can output as a table
$ prolific submission list -s 63c123af913a974f87e8e7fc --table
$ prolific submission list -s 63c123af913a974f87e8e7fc -t

You can output as CSV
$ prolific submission list -s 63c123af913a974f87e8e7fc --csv
$ prolific submission list -s 63c123af913a974f87e8e7fc -c

You can output as JSON
$ prolific submission list -s 63c123af913a974f87e8e7fc --json
$ prolific submission list -s 63c123af913a974f87e8e7fc -j

You can specify the fields you want to render in table or CSV output
$ prolific submission list -s 63c123af913a974f87e8e7fc -f ID,Status,TimeTaken -t
$ prolific submission list -s 63c123af913a974f87e8e7fc -f ID,Status,TimeTaken -c

The fields you can use are
- ID
- ParticipantID
- StartedAt
- CompletedAt
- IsComplete
- TimeTaken
- Reward
- Status
- StudyCode
- StarAwarded
- BonusPayments
- IP`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			if opts.Study == "" {
				return errors.New("please provide a study ID")
			}

			submissions, err := c.GetSubmissions(opts.Study, opts.Limit, opts.Offset)
			if err != nil {
				return err
			}

			format := shared.ResolveFormat(opts.Output)
			fields := opts.Fields
			if fields == "" {
				fields = defaultListFields
			}

			switch format {
			case "json":
				r := ui.JSONRenderer[model.Submission]{}
				if err := r.Render(submissions.Results, w); err != nil {
					return fmt.Errorf("error: %s", err)
				}
			case "csv":
				r := ui.CsvRenderer[model.Submission]{}
				if err := r.Render(submissions.Results, fields, w); err != nil {
					return fmt.Errorf("error: %s", err)
				}
				fmt.Fprintf(w, "\n%s\n", ui.RenderRecordCounter(len(submissions.Results), submissions.Meta.Count))
			case "table":
				r := ui.TableRenderer[model.Submission]{}
				if err := r.Render(submissions.Results, fields, w); err != nil {
					return fmt.Errorf("error: %s", err)
				}
				fmt.Fprintf(w, "\n%s\n", ui.RenderRecordCounter(len(submissions.Results), submissions.Meta.Count))
			default:
				r := &InteractiveRenderer{}
				if err := r.Render(*submissions, w); err != nil {
					return fmt.Errorf("error: %s", err)
				}
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Study, "study", "s", "", "The study we want submissions for.")
	flags.StringVarP(&opts.Fields, "fields", "f", "", "Comma separated list of fields you want to display in table or CSV output.")
	flags.IntVarP(&opts.Limit, "limit", "l", client.DefaultRecordLimit, "Limit the number of submissions returned.")
	flags.IntVarP(&opts.Offset, "offset", "o", client.DefaultRecordOffset, "The number of submissions to offset.")
	shared.AddOutputFlags(cmd, &opts.Output)

	return cmd
}

// InteractiveRenderer runs the Bubbletea UI framework to provide a rich
// UI experience for the user.
type InteractiveRenderer struct{}

// Render builds the item list and launches the interactive TUI.
func (r *InteractiveRenderer) Render(submissions client.ListSubmissionsResponse, w io.Writer) error {
	var items []list.Item
	submissionMap := make(map[string]model.Submission)

	for _, s := range submissions.Results {
		items = append(items, s)
		submissionMap[s.ID] = s
	}

	lv := ListView{
		List:        list.New(items, list.NewDefaultDelegate(), 0, 0),
		Submissions: submissionMap,
	}
	lv.List.Title = "Submissions"

	p := tea.NewProgram(lv)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("cannot render submissions: %s", err)
	}

	return nil
}

// ListView presents an interactive list of submissions using the Bubbletea TUI.
type ListView struct {
	List        list.Model
	Submissions map[string]model.Submission
	Submission  *model.Submission
}

// Init initialises the view.
func (lv ListView) Init() tea.Cmd {
	return nil
}

// Update handles TUI messages and key events.
func (lv ListView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return lv, tea.Quit
		}

		if msg.String() == "enter" {
			i, ok := lv.List.SelectedItem().(model.Submission)
			if ok {
				lv.Submission = &i
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

// View renders the current state of the TUI.
func (lv ListView) View() string {
	if lv.Submission != nil {
		return RenderSubmission(*lv.Submission)
	}
	return lv.List.View()
}
