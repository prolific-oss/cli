package submission

import (
	"errors"
	"fmt"
	"io"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/ui/submission"
	"github.com/spf13/cobra"
)

// ListOptions is the options for the listing studies command.
type ListOptions struct {
	Args           []string
	NonInteractive bool
	Fields         string
	Study          string
	Csv            bool
	Limit          int
	Offset         int
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

You can render the results as a CSV format
$ prolific submission list -s 63c123af913a974f87e8e7fc -l 5 -c

You can specify the fields you want to render in either the standard or CSV view
$ prolific submission list -s 63c123af913a974f87e8e7fc -f ID,Status,TimeTaken

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

			renderer := submission.ListRenderer{}

			if opts.Csv {
				renderer.SetStrategy(&submission.CsvRenderer{})
			} else {
				renderer.SetStrategy(&submission.NonInteractiveRenderer{})
			}

			err := renderer.Render(c, submission.ListUsedOptions{
				StudyID:        opts.Study,
				Csv:            opts.Csv,
				NonInteractive: opts.NonInteractive,
				Fields:         opts.Fields,
				Limit:          opts.Limit,
				Offset:         opts.Offset,
			}, w)

			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.BoolVarP(&opts.NonInteractive, "non-interactive", "n", true, "Render the list details straight to the terminal.")
	flags.BoolVarP(&opts.Csv, "csv", "c", false, "Render the list details in a CSV format.")
	flags.StringVarP(&opts.Study, "study", "s", "", "The study we want submissions for.")
	flags.StringVarP(&opts.Fields, "fields", "f", "", "Comma separated list of fields you want to display in non-interactive/csv mode.")
	flags.IntVarP(&opts.Limit, "limit", "l", client.DefaultRecordLimit, "Limit the number of events returned.")
	flags.IntVarP(&opts.Offset, "offset", "o", client.DefaultRecordOffset, "The number of events to offset.")

	return cmd
}
