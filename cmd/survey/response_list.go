package survey

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/shared"
	"github.com/prolific-oss/cli/model"
	"github.com/prolific-oss/cli/ui"
	"github.com/spf13/cobra"
)

const defaultResponseListFields = "ID,ParticipantID,SubmissionID,DateCreated"

// ResponseListOptions is the options for the listing survey responses command.
type ResponseListOptions struct {
	Args   []string
	Output shared.OutputOptions
	Limit  int
	Offset int
}

// NewResponseListCommand creates a new command to list survey responses
func NewResponseListCommand(commandName string, c client.API, w io.Writer) *cobra.Command {
	var opts ResponseListOptions

	cmd := &cobra.Command{
		Use:   commandName + " <survey_id>",
		Args:  cobra.MinimumNArgs(1),
		Short: "List responses for a survey",
		Long: `List responses for a survey

Shows all participant responses submitted for the given survey.
`,
		Example: `
List survey responses interactively
$ prolific survey response list 6261321e223a605c7a4f7678

List survey responses as a table
$ prolific survey response list 6261321e223a605c7a4f7678 -n

List survey responses as JSON
$ prolific survey response list 6261321e223a605c7a4f7678 --json
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			surveyID := args[0]

			responses, err := c.GetSurveyResponses(surveyID, opts.Limit, opts.Offset)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			format := shared.ResolveFormat(opts.Output)
			switch format {
			case "json":
				r := ui.JSONRenderer[model.SurveyResponse]{}
				if err := r.Render(responses.Results, w); err != nil {
					return fmt.Errorf("error: %s", err)
				}
			case "csv":
				r := ui.CsvRenderer[model.SurveyResponse]{}
				if err := r.Render(responses.Results, defaultResponseListFields, w); err != nil {
					return fmt.Errorf("error: %s", err)
				}
			case "table":
				r := ui.TableRenderer[model.SurveyResponse]{}
				if err := r.Render(responses.Results, defaultResponseListFields, w); err != nil {
					return fmt.Errorf("error: %s", err)
				}
			default:
				r := &ResponseInteractiveRenderer{}
				if err := r.Render(*responses, w); err != nil {
					return fmt.Errorf("error: %s", err)
				}
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.IntVarP(&opts.Limit, "limit", "l", client.DefaultRecordLimit, "Limit the number of responses returned")
	flags.IntVarP(&opts.Offset, "offset", "o", client.DefaultRecordOffset, "The number of responses to offset")
	shared.AddOutputFlags(cmd, &opts.Output)

	return cmd
}

// ResponseInteractiveRenderer runs the Bubbletea UI framework for an interactive response list.
type ResponseInteractiveRenderer struct{}

// Render will render the survey response list in an interactive manner.
func (r *ResponseInteractiveRenderer) Render(responses client.ListSurveyResponsesResponse, w io.Writer) error {
	var items []list.Item

	for _, resp := range responses.Results {
		items = append(items, model.SurveyResponseListItem{SurveyResponse: resp})
	}

	lv := ResponseListView{
		List: list.New(items, list.NewDefaultDelegate(), 0, 0),
	}
	lv.List.Title = "Survey Responses"

	p := tea.NewProgram(lv)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("cannot render survey responses: %s", err)
	}

	return nil
}
