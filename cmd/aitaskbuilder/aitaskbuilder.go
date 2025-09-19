package aitaskbuilder

import (
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// NewAITaskBuilderCommand creates a new `aitaskbuilder` command
func NewAITaskBuilderCommand(client client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "aitaskbuilder",
		Short: "AI Task Builder tools and utilities",
		Long:  "Manage AI task building workflows, datasets, and batch operations for the Prolific platform",
	}

	cmd.AddCommand(
		NewGetBatchCommand(client, w),
		NewGetBatchStatusCommand(client, w),
		NewGetBatchesCommand(client, w),
		NewGetResponsesCommand(client, w),
		NewGetDatasetStatusCommand(client, w),
	)

	return cmd
}
