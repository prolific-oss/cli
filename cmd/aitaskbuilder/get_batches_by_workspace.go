package aitaskbuilder

import (
	"errors"
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

type BatchGetBatchesByWorkspaceOptions struct {
	Args        []string
	WorkspaceID string
}

func renderAITaskBuilderBatchesByWorkspace(c client.API, opts BatchGetBatchesByWorkspaceOptions, w io.Writer) error {
	if opts.WorkspaceID == "" {
		return errors.New("workspace ID is required")
	}

	response, err := c.GetAITaskBuilderBatchesByWorkspace(opts.WorkspaceID)
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

func NewGetBatchesByWorkspaceCommand(client client.API, w io.Writer) *cobra.Command {
	var opts BatchGetBatchesByWorkspaceOptions

	cmd := &cobra.Command{
		Use:   "getbatchesbyworkspace",
		Short: "Get AI Task Builder batches by workspace",
		Long: `Get the batches for a given workspace.

This command allows you to retrieve the batches for a given workspace by providing
the workspace ID.`,
		Example: `
Get AI Task Builder batches by workspace:
$ prolific aitaskbuilder getbatchesbyworkspace -w <workspace_id>
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := renderAITaskBuilderBatchesByWorkspace(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.WorkspaceID, "workspace-id", "w", "", "Workspace ID (required) - The ID of the workspace to retrieve batches from.")

	_ = cmd.MarkFlagRequired("workspace-id")

	return cmd
}
