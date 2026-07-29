package feedback

import (
	"errors"
	"fmt"
	"io"
	"regexp"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/ui/feedback"
	"github.com/spf13/cobra"
)

// studyIDPattern matches the study ID format accepted by the feedback API.
var studyIDPattern = regexp.MustCompile(`^[a-f\d]{24}$`)

// ListOptions is the options for the listing study feedback command.
type ListOptions struct {
	Args               []string
	Study              string
	Json               bool
	Csv                bool
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
including written feedback and clarity, ease and fairness ratings, through
the same permissioned server-side path used by the Feedback dashboard.

By default, every feedback record for the study is returned in one response.
Pass --limit (and optionally --offset) to page through results instead.`,
		Example: `
You can list all feedback for a given study, along with the study's aggregated
ratings
$ prolific feedback list -s 63c123af913a974f87e8e7fc

You can filter to only feedback that includes written text
$ prolific feedback list -s 63c123af913a974f87e8e7fc --has-written-feedback

You can page through results instead, for example 10 at a time
$ prolific feedback list -s 63c123af913a974f87e8e7fc -l 10 -o 10

You can render the results as JSON for machine-readable output
$ prolific feedback list -s 63c123af913a974f87e8e7fc --json

You can render the results as a CSV format
$ prolific feedback list -s 63c123af913a974f87e8e7fc -c`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			if opts.Study == "" {
				return errors.New("please provide a study ID")
			}

			if !studyIDPattern.MatchString(opts.Study) {
				return fmt.Errorf("invalid study ID %q: must be a 24-character hex string", opts.Study)
			}

			if opts.Limit < 0 {
				return errors.New("limit must be greater than or equal to 0")
			}

			renderer := feedback.ListRenderer{}

			if opts.Json {
				renderer.SetStrategy(&feedback.JSONRenderer{})
			} else if opts.Csv {
				renderer.SetStrategy(&feedback.CsvRenderer{})
			} else {
				renderer.SetStrategy(&feedback.NonInteractiveRenderer{})
			}

			err := renderer.Render(c, feedback.ListUsedOptions{
				StudyID:            opts.Study,
				HasWrittenFeedback: opts.HasWrittenFeedback,
				Limit:              opts.Limit,
				Offset:             opts.Offset,
			}, w)

			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Study, "study", "s", "", "The study you want feedback for (required).")
	flags.BoolVar(&opts.HasWrittenFeedback, "has-written-feedback", false, "Only return feedback that includes written text.")
	flags.BoolVar(&opts.Json, "json", false, "Render the results in JSON format for machine-readable output.")
	flags.BoolVarP(&opts.Csv, "csv", "c", false, "Render the results in a CSV format.")
	flags.IntVarP(&opts.Limit, "limit", "l", 0, "Limit the number of feedback records returned per page. Omit to fetch every record for the study.")
	flags.IntVarP(&opts.Offset, "offset", "o", client.DefaultRecordOffset, "The number of feedback records to offset.")

	return cmd
}
