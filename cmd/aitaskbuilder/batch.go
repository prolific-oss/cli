package aitaskbuilder

import (
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// NewBatchCommand creates a new `batch` command under aitaskbuilder
func NewBatchCommand(client client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "batch",
		Short: "AI Task Builder batch operations",
		Long:  "Manage AI Task Builder batches - get batch details, check status, and list batches",
	}

	cmd.AddCommand(
		NewBatchCreateCommand(client, w),
		NewBatchGetCommand(client, w),
		NewBatchSetupCommand(client, w),
		NewBatchStatusCommand(client, w),
		NewBatchListCommand(client, w),
	)

	return cmd
}
