package submission

import (
	"fmt"
	"io"

	"github.com/prolific-oss/prolificli/client"
	"github.com/prolific-oss/prolificli/ui/submission"
	"github.com/spf13/cobra"
)

// ListOptions is the options for the listing studies command.
type ListOptions struct {
	Args           []string
	NonInteractive bool
	Fields         string
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
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			renderer := submission.ListRenderer{}

			if opts.Csv {
				renderer.SetStrategy(&submission.CsvRenderer{})
			} else {
				renderer.SetStrategy(&submission.NonInteractiveRenderer{})
			}

			err := renderer.Render(c, submission.ListUsedOptions{
				StudyID:        args[0],
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
	flags.StringVarP(&opts.Fields, "fields", "f", "", "Comma separated list of fields you want to display in non-interactive/csv mode.")
	flags.IntVarP(&opts.Limit, "limit", "l", client.DefaultRecordLimit, "Limit the number of events returned")
	flags.IntVarP(&opts.Offset, "offset", "o", client.DefaultRecordOffset, "The number of events to offset")

	return cmd
}
