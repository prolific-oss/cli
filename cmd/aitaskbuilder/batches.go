package aitaskbuilder

import (
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// NewBatchesCommand creates a new `batch` command
func NewBatchesCommand(client client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "batch",
		Short: "Manage your batches",
		Long:  "Create and manage AI Task Builder batches in your workspace",
	}

	cmd.AddCommand(
		NewGetBatchCommand(client, w),
		NewGetBatchStatusCommand(client, w),
		NewGetBatchesListCommand(client, w),
		NewGetResponsesCommand(client, w),
	)

	return cmd
}
