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

// defaultListFields is the default fields shown in table mode.
const defaultListFields = "ID,Title,DateCreated"

// ListOptions is the options for the listing surveys command.
type ListOptions struct {
	Args   []string
	Output shared.OutputOptions
	Limit  int
	Offset int
}

// NewListCommand creates a new command to list surveys
func NewListCommand(commandName string, c client.API, w io.Writer) *cobra.Command {
	var opts ListOptions

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "Provide a list of your surveys",
		Long: `List your surveys

Surveys are associated with your researcher account.
`,
		Example: `
List your surveys interactively
$ prolific survey list

List your surveys as a table
$ prolific survey list -n

List your surveys as JSON
$ prolific survey list --json
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			me, err := c.GetMe()
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			surveys, err := c.GetSurveys(me.ID, opts.Limit, opts.Offset)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			format := shared.ResolveFormat(opts.Output)
			switch format {
			case "json":
				r := ui.JSONRenderer[model.Survey]{}
				if err := r.Render(surveys.Results, w); err != nil {
					return fmt.Errorf("error: %s", err)
				}
			case "table":
				r := ui.TableRenderer[model.Survey]{}
				if err := r.Render(surveys.Results, defaultListFields, w); err != nil {
					return fmt.Errorf("error: %s", err)
				}
			default:
				r := &InteractiveRenderer{}
				if err := r.Render(c, *surveys, w); err != nil {
					return fmt.Errorf("error: %s", err)
				}
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.IntVarP(&opts.Limit, "limit", "l", client.DefaultRecordLimit, "Limit the number of surveys returned")
	flags.IntVarP(&opts.Offset, "offset", "o", client.DefaultRecordOffset, "The number of surveys to offset")
	shared.AddOutputFlags(cmd, &opts.Output)

	return cmd
}

// InteractiveRenderer runs the Bubbletea UI framework for an interactive list.
type InteractiveRenderer struct{}

// Render will render the survey list in an interactive manner.
func (r *InteractiveRenderer) Render(c client.API, surveys client.ListSurveysResponse, w io.Writer) error {
	var items []list.Item
	surveyMap := make(map[string]model.Survey)

	for _, s := range surveys.Results {
		items = append(items, model.SurveyListItem{Survey: s})
		surveyMap[s.ID] = s
	}

	lv := SurveyListView{
		List:    list.New(items, list.NewDefaultDelegate(), 0, 0),
		Surveys: surveyMap,
		Client:  c,
	}
	lv.List.Title = "Surveys"

	p := tea.NewProgram(lv)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("cannot render surveys: %s", err)
	}

	return nil
}
