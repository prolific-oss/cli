package collection

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

// ExportListOptions is the options for the collection export list command.
type ExportListOptions struct {
	Args   []string
	Output shared.OutputOptions
}

// NewExportListCommand creates a new `collection export list` command to
// list every export job requested for a collection.
func NewExportListCommand(c client.API, w io.Writer) *cobra.Command {
	var opts ExportListOptions

	cmd := &cobra.Command{
		Use:   "list <collection-id>",
		Args:  cobra.MinimumNArgs(1),
		Short: "List a collection's export jobs",
		Long: `List a collection's export jobs

This command lists every export job that has been requested for a
collection, most recent first, including its status and the filter (if any)
it was requested with. A collection can have at most 10 export jobs at
once.`,
		Example: `
List all export jobs for a collection:

$ prolific collection export list 5f8e3c2a-1d4b-4e6f-9a7c-2b0d8f3e1c5a

You can output as a table
$ prolific collection export list 5f8e3c2a-1d4b-4e6f-9a7c-2b0d8f3e1c5a --table

You can output as CSV
$ prolific collection export list 5f8e3c2a-1d4b-4e6f-9a7c-2b0d8f3e1c5a --csv

You can output as JSON
$ prolific collection export list 5f8e3c2a-1d4b-4e6f-9a7c-2b0d8f3e1c5a --json
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			if len(opts.Args) < 1 || opts.Args[0] == "" {
				return errors.New("please provide a collection ID")
			}

			return listCollectionExports(c, opts, w)
		},
	}

	shared.AddOutputFlags(cmd, &opts.Output)

	return cmd
}

func listCollectionExports(c client.API, opts ExportListOptions, w io.Writer) error {
	collectionID := opts.Args[0]

	jobs, err := c.ListCollectionExportJobs(collectionID)
	if err != nil {
		if shared.IsFeatureNotEnabledError(err) {
			ui.RenderFeatureAccessMessage(FeatureNameAITBCollection, FeatureContactURLAITBCollection)
			return nil
		}
		return fmt.Errorf("error listing exports: %s", err.Error())
	}

	switch shared.ResolveFormat(opts.Output) {
	case "json":
		r := ui.JSONRenderer[client.ExportJobListItem]{}
		if err := r.Render(jobs, w); err != nil {
			return fmt.Errorf("error: %s", err)
		}
	case "csv":
		r := ui.CsvRenderer[ExportListItem]{}
		if err := r.Render(NewExportListItems(jobs), ExportListFields, w); err != nil {
			return fmt.Errorf("error: %s", err)
		}
	default:
		if len(jobs) == 0 {
			fmt.Fprintf(w, "No export jobs found for collection %s\n", collectionID)
			return nil
		}
		r := ui.TableRenderer[ExportListItem]{}
		if err := r.Render(NewExportListItems(jobs), ExportListFields, w); err != nil {
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
