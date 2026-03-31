package study

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/model"
	submissionui "github.com/prolific-oss/cli/ui/submission"
	"github.com/spf13/cobra"
)

// SubmissionCountsOptions is the options for the submission-counts command.
type SubmissionCountsOptions struct {
	JSON           bool
	NonInteractive bool
}

// NewSubmissionCountsCommand creates a new `study submission-counts` command to
// retrieve submission counts grouped by status for a study.
func NewSubmissionCountsCommand(client client.API, w io.Writer) *cobra.Command {
	var opts SubmissionCountsOptions

	cmd := &cobra.Command{
		Use:   "submission-counts",
		Short: "Get submission counts by status for a study",
		Long:  `Retrieve a summary of submission counts grouped by status for a given study`,
		Example: `
To get submission counts for a study (interactive):
$ prolific study submission-counts 64395e9c2332b8a59a65d51e

To get submission counts as a table:
$ prolific study submission-counts 64395e9c2332b8a59a65d51e -n

To get submission counts as JSON:
$ prolific study submission-counts 64395e9c2332b8a59a65d51e --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			studyID := args[0]

			counts, err := client.GetStudySubmissionCounts(studyID)
			if err != nil {
				return err
			}

			if opts.JSON {
				data, err := json.MarshalIndent(counts, "", "  ")
				if err != nil {
					return err
				}
				fmt.Fprintln(w, string(data))
				return nil
			}

			if opts.NonInteractive {
				fmt.Fprint(w, renderSubmissionCounts(counts))
				return nil
			}

			countItems := counts.ToItems()
			if len(countItems) == 0 {
				fmt.Fprintln(w, "No submissions found for this study.")
				return nil
			}

			var items []list.Item
			for _, item := range countItems {
				items = append(items, item)
			}

			cv := submissionui.NewCountsView(items, studyID, client)
			p := tea.NewProgram(cv)
			if _, err := p.Run(); err != nil {
				return fmt.Errorf("cannot render submission counts: %s", err)
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.BoolVar(&opts.JSON, "json", false, "Output as JSON")
	flags.BoolVarP(&opts.NonInteractive, "non-interactive", "n", false, "Render as a table")

	return cmd
}

func renderSubmissionCounts(counts *model.SubmissionCounts) string {
	var buf bytes.Buffer
	tw := tabwriter.NewWriter(&buf, 0, 0, 3, ' ', 0)

	fmt.Fprintln(tw, "STATUS\tCOUNT")
	fmt.Fprintf(tw, "Active\t%d\n", counts.Active)
	fmt.Fprintf(tw, "Approved\t%d\n", counts.Approved)
	fmt.Fprintf(tw, "Awaiting Review\t%d\n", counts.AwaitingReview)
	fmt.Fprintf(tw, "Rejected\t%d\n", counts.Rejected)
	fmt.Fprintf(tw, "Reserved\t%d\n", counts.Reserved)
	fmt.Fprintf(tw, "Returned\t%d\n", counts.Returned)
	fmt.Fprintf(tw, "Timed Out\t%d\n", counts.TimedOut)
	fmt.Fprintf(tw, "Partially Approved\t%d\n", counts.PartiallyApproved)
	fmt.Fprintf(tw, "Screened Out\t%d\n", counts.ScreenedOut)
	fmt.Fprintf(tw, "Total\t%d\n", counts.Total)
	tw.Flush()

	return buf.String()
}
