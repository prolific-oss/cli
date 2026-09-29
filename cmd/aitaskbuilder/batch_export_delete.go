package aitaskbuilder

import (
	"errors"
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// BatchExportDeleteOptions holds the options for the batch export delete command.
type BatchExportDeleteOptions struct {
	Args []string
}

// NewBatchExportDeleteCommand creates a new `aitaskbuilder batch export delete`
// command to permanently delete an export job for a batch.
func NewBatchExportDeleteCommand(c client.API, w io.Writer) *cobra.Command {
	var opts BatchExportDeleteOptions

	cmd := &cobra.Command{
		Use:   "delete <batch-id> <export-id>",
		Args:  cobra.ExactArgs(2),
		Short: "Delete a batch's export job",
		Long: `Delete a batch's export job

This command permanently deletes an export job and, if it completed, its
ZIP archive. An export that is still generating cannot be deleted.`,
		Example: `
Delete an export job for a batch:

$ prolific aitaskbuilder batch export delete 5f8e3c2a-1d4b-4e6f-9a7c-2b0d8f3e1c5a export-job-uuid-456
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			if len(opts.Args) < 2 || opts.Args[0] == "" || opts.Args[1] == "" {
				return errors.New("please provide a batch ID and an export ID")
			}

			return deleteBatchExport(c, opts, w)
		},
	}

	return cmd
}

func deleteBatchExport(c client.API, opts BatchExportDeleteOptions, w io.Writer) error {
	batchID := opts.Args[0]
	exportID := opts.Args[1]

	if err := c.DeleteBatchExport(batchID, exportID); err != nil {
		return fmt.Errorf("error deleting export: %s", err.Error())
	}

	fmt.Fprintf(w, "Export %s for batch %s deleted.\n", exportID, batchID)

	return nil
}
