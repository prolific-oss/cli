package aitaskbuilder

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/shared"
	"github.com/prolific-oss/cli/ui"
	"github.com/spf13/cobra"
)

// BatchExportListOptions holds the options for the batch export list command.
type BatchExportListOptions struct {
	Args   []string
	Output shared.OutputOptions
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

You can output as a table
$ prolific aitaskbuilder batch export list 5f8e3c2a-1d4b-4e6f-9a7c-2b0d8f3e1c5a --table

You can output as CSV
$ prolific aitaskbuilder batch export list 5f8e3c2a-1d4b-4e6f-9a7c-2b0d8f3e1c5a --csv

You can output as JSON
$ prolific aitaskbuilder batch export list 5f8e3c2a-1d4b-4e6f-9a7c-2b0d8f3e1c5a --json
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			if len(opts.Args) < 1 || opts.Args[0] == "" {
				return errors.New("please provide a batch ID")
			}

			return listBatchExports(c, opts, w)
		},
	}

	shared.AddOutputFlags(cmd, &opts.Output)

	return cmd
}

func listBatchExports(c client.API, opts BatchExportListOptions, w io.Writer) error {
	batchID := opts.Args[0]

	jobs, err := c.ListBatchExportJobs(batchID)
	if err != nil {
		return fmt.Errorf("error listing exports: %s", err.Error())
	}

	switch shared.ResolveFormat(opts.Output) {
	case "json":
		r := ui.JSONRenderer[client.ExportJobListItem]{}
		if err := r.Render(jobs, w); err != nil {
			return fmt.Errorf("error: %s", err)
		}
	case "csv":
		r := ui.CsvRenderer[BatchExportListItem]{}
		if err := r.Render(NewBatchExportListItems(jobs), BatchExportListFields, w); err != nil {
			return fmt.Errorf("error: %s", err)
		}
	default:
		if len(jobs) == 0 {
			fmt.Fprintf(w, "No export jobs found for batch %s\n", batchID)
			return nil
		}
		r := ui.TableRenderer[BatchExportListItem]{}
		if err := r.Render(NewBatchExportListItems(jobs), BatchExportListFields, w); err != nil {
			return fmt.Errorf("error: %s", err)
		}
	}

	return nil
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
