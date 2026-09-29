package collection

import (
	"errors"
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/shared"
	"github.com/prolific-oss/cli/ui"
	"github.com/spf13/cobra"
)

// ExportDeleteOptions is the options for the collection export delete command.
type ExportDeleteOptions struct {
	Args []string
}

// NewExportDeleteCommand creates a new `collection export delete` command to
// permanently delete an export job for a collection.
func NewExportDeleteCommand(c client.API, w io.Writer) *cobra.Command {
	var opts ExportDeleteOptions

	cmd := &cobra.Command{
		Use:   "delete <collection-id> <export-id>",
		Args:  cobra.ExactArgs(2),
		Short: "Delete a collection's export job",
		Long: `Delete a collection's export job

This command permanently deletes an export job and, if it completed, its
ZIP archive. An export that is still generating cannot be deleted.`,
		Example: `
Delete an export job for a collection:

$ prolific collection export delete 5f8e3c2a-1d4b-4e6f-9a7c-2b0d8f3e1c5a export-job-uuid-456
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			if len(opts.Args) < 2 || opts.Args[0] == "" || opts.Args[1] == "" {
				return errors.New("please provide a collection ID and an export ID")
			}

			return deleteCollectionExport(c, opts, w)
		},
	}

	return cmd
}

func deleteCollectionExport(c client.API, opts ExportDeleteOptions, w io.Writer) error {
	collectionID := opts.Args[0]
	exportID := opts.Args[1]

	if err := c.DeleteCollectionExport(collectionID, exportID); err != nil {
		if shared.IsFeatureNotEnabledError(err) {
			ui.RenderFeatureAccessMessage(FeatureNameAITBCollection, FeatureContactURLAITBCollection)
			return nil
		}
		return fmt.Errorf("error deleting export: %s", err.Error())
	}

	fmt.Fprintf(w, "Export %s for collection %s deleted.\n", exportID, collectionID)

	return nil
}
