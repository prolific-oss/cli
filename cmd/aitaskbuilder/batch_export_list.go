package aitaskbuilder

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// BatchExportListOptions holds the options for the batch export list command.
type BatchExportListOptions struct {
	Args []string
}

// NewBatchExportListCommand creates a new `aitaskbuilder batch export list`
// command to list every export job requested for a batch.
func NewBatchExportListCommand(c client.API, w io.Writer) *cobra.Command {
	var opts BatchExportListOptions

	cmd := &cobra.Command{
		Use:   "list <batch-id>",
		Args:  cobra.MinimumNArgs(1),
		Short: "List a batch's export jobs",
		Long: `List a batch's export jobs

This command lists every export job that has been requested for a batch,
most recent first, including its status and the filter (if any) it was
requested with. A batch can have at most 10 export jobs at once.`,
		Example: `
List all export jobs for a batch:

$ prolific aitaskbuilder batch export list 5f8e3c2a-1d4b-4e6f-9a7c-2b0d8f3e1c5a
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			if len(opts.Args) < 1 || opts.Args[0] == "" {
				return errors.New("please provide a batch ID")
			}

			return listBatchExports(c, opts, w)
		},
	}

	return cmd
}

func listBatchExports(c client.API, opts BatchExportListOptions, w io.Writer) error {
	batchID := opts.Args[0]

	jobs, err := c.ListBatchExportJobs(batchID)
	if err != nil {
		return fmt.Errorf("error listing exports: %s", err.Error())
	}

	if len(jobs) == 0 {
		fmt.Fprintf(w, "No export jobs found for batch %s\n", batchID)
		return nil
	}

	tw := tabwriter.NewWriter(w, 0, 1, 1, ' ', 0)
	fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", "Export ID", "Filter", "Status", "Created At")
	for _, job := range jobs {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", job.ExportID, formatExportJobFilter(job.Filter), job.Status, job.CreatedAt)
	}

	return tw.Flush()
}

// formatExportJobFilter renders an export job's filter as a short,
// human-readable summary, e.g. "study_id=X, from=Y, to=Z". A nil filter
// (unfiltered/full export) renders as "-".
func formatExportJobFilter(filter *client.ExportJobFilter) string {
	if filter == nil {
		return "-"
	}

	var parts []string
	if filter.StudyID != "" {
		parts = append(parts, fmt.Sprintf("study_id=%s", filter.StudyID))
	}
	if filter.From != "" {
		parts = append(parts, fmt.Sprintf("from=%s", filter.From))
	}
	if filter.To != "" {
		parts = append(parts, fmt.Sprintf("to=%s", filter.To))
	}

	if len(parts) == 0 {
		return "none"
	}

	return strings.Join(parts, ", ")
}
