package aitaskbuilder

import (
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// NewDatasetsCommand creates a new `dataset` command
func NewDatasetsCommand(client client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dataset",
		Short: "Manage your datasets",
		Long:  "Create and manage AI Task Builder datasets in your workspace",
	}

	cmd.AddCommand(
		NewGetDatasetStatusCommand(client, w),
		NewCreateDatasetCommand(client, w),
	)

	return cmd
}
