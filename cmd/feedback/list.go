package feedback

import (
	"errors"
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/shared"
	"github.com/prolific-oss/cli/model"
	"github.com/prolific-oss/cli/ui"
	"github.com/spf13/cobra"
)

// ListOptions is the options for the listing study feedback command.
type ListOptions struct {
	Args               []string
	Study              string
	Output             shared.OutputOptions
	HasWrittenFeedback bool
	Limit              int
	Offset             int
}

// NewListCommand creates a new `feedback list` command to give you details
// about participant feedback for a study.
func NewListCommand(c client.API, w io.Writer) *cobra.Command {
	var opts ListOptions

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List participant feedback for a study, requires Study ID",
		Long: `List participant feedback for a study

Retrieves permissioned, record-level participant feedback for a study,
including written feedback and clarity, difficulty and fairness ratings, through
the same permissioned server-side path used by the Feedback dashboard.

By default, up to 200 feedback records are returned. Pass --limit 0 to fetch
every record, or use --limit and --offset to page through results.`,
		Example: `
You can list all feedback for a given study
$ prolific feedback list -s 63c123af913a974f87e8e7fc

You can filter to only feedback that includes written text
$ prolific feedback list -s 63c123af913a974f87e8e7fc --has-written-feedback

You can page through results instead, for example 10 at a time
$ prolific feedback list -s 63c123af913a974f87e8e7fc -l 10 -o 10

You can render the results as a non-interactive table
$ prolific feedback list -s 63c123af913a974f87e8e7fc --table

You can render the results as JSON for machine-readable output
$ prolific feedback list -s 63c123af913a974f87e8e7fc --json

You can render the results as a CSV format
$ prolific feedback list -s 63c123af913a974f87e8e7fc -c`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			if opts.Study == "" {
				return errors.New("please provide a study ID")
			}

			if opts.Limit < 0 {
				return errors.New("limit must be greater than or equal to 0")
			}

			response, err := c.GetStudyFeedback(
				opts.Study,
				opts.HasWrittenFeedback,
				opts.Limit,
				opts.Offset,
			)
			if err != nil {
				return handleAPIError(err)
			}

			records := response.Results
			if records == nil {
				records = []model.StudyFeedback{}
			}
			total := len(records)
			if response.JSONAPIMeta != nil {
				total = response.Meta.Count
			}

			switch shared.ResolveFormat(opts.Output) {
			case "json":
				renderer := ui.JSONRenderer[model.StudyFeedback]{}
				if err := renderer.Render(records, w); err != nil {
					return fmt.Errorf("error: %s", err)
				}
			case "csv":
				renderer := ui.CsvRenderer[ListItem]{}
				if err := renderer.Render(NewListItems(records), ListFields, w); err != nil {
					return fmt.Errorf("error: %s", err)
				}
			case "table":
				renderer := ui.TableRenderer[ListItem]{}
				if err := renderer.Render(NewListItems(records), ListFields, w); err != nil {
					return fmt.Errorf("error: %s", err)
				}
				fmt.Fprintf(w, "\n%s\n", ui.RenderRecordCounter(len(records), total))
			default:
				renderer := InteractiveRenderer{}
				if err := renderer.Render(records, w); err != nil {
					return fmt.Errorf("error: %s", err)
				}
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Study, "study", "s", "", "The study you want feedback for (required).")
	flags.BoolVar(&opts.HasWrittenFeedback, "has-written-feedback", false, "Only return feedback that includes written text.")
	flags.IntVarP(&opts.Limit, "limit", "l", client.DefaultRecordLimit, "Limit the number of feedback records returned per page. Use 0 to fetch every record.")
	flags.IntVarP(&opts.Offset, "offset", "o", client.DefaultRecordOffset, "The number of feedback records to offset.")
	shared.AddOutputFlags(cmd, &opts.Output)

	return cmd
}
