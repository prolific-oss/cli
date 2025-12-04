package aitaskbuilder

import (
	"errors"
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type BatchGetBatchesOptions struct {
	Args        []string
	WorkspaceID string
}

func renderAITaskBuilderBatches(c client.API, opts BatchGetBatchesOptions, w io.Writer) error {
	if opts.WorkspaceID == "" {
		return errors.New(ErrWorkspaceIDRequired)
	}

	response, err := c.GetAITaskBuilderBatches(opts.WorkspaceID)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "AI Task Builder Batches by Workspace:\n")
	fmt.Fprintf(w, "Workspace ID: %s\n", opts.WorkspaceID)
	fmt.Fprintf(w, "Batches: %d\n", len(response.Results))
	for i, batch := range response.Results {
		fmt.Fprintf(w, "  Batch %d: %s | Name: %s | Status: %s\n", i+1, batch.ID, batch.Name, batch.Status)
	}

	if len(response.Results) == 0 {
		fmt.Fprintf(w, "No batches found for workspace %s\n", opts.WorkspaceID)
	}

	return nil
}

func NewGetBatchesListCommand(client client.API, w io.Writer) *cobra.Command {
	var opts BatchGetBatchesOptions

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List batches in a workspace",
		Long: `List the batches in a workspace.

This command lists all batches in a workspace by providing the workspace ID.`,
		Example: `
Get AI Task Builder batches:
$ prolific aitaskbuilder batch list -w <workspace_id>
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := renderAITaskBuilderBatches(client, opts, w)
			if err != nil {
				return err
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.WorkspaceID, "workspace-id", "w", viper.GetString("workspace"), "Workspace ID (required) - The ID of the workspace to retrieve batches from.")

	_ = cmd.MarkFlagRequired("workspace-id")

	return cmd
}
