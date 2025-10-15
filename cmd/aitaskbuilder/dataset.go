package aitaskbuilder

import (
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// NewDatasetCommand creates a new `dataset` command under aitaskbuilder
func NewDatasetCommand(client client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dataset",
		Short: "AI Task Builder dataset operations",
		Long:  "Manage AI Task Builder datasets - check status and manage dataset operations",
	}

	cmd.AddCommand(
		NewDatasetCreateCommand(client, w),
		NewDatasetStatusCommand(client, w),
		NewDatasetUploadCommand(client, w),
	)

	return cmd
}
