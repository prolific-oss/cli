package aitaskbuilder

import (
	"errors"
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// CreateDatasetOptions holds the options for creating a dataset
type CreateDatasetOptions struct {
	Args        []string
	Name        string
	WorkspaceID string
}

// NewCreateDatasetCommand creates a new command for creating datasets
func NewCreateDatasetCommand(client client.API, w io.Writer) *cobra.Command {
	var opts CreateDatasetOptions

	cmd := &cobra.Command{
		Use:   "create-dataset",
		Short: "Create an AI Task Builder dataset",
		Long: `Create a new AI Task Builder dataset

Provide a name and workspace ID to create a new dataset in your workspace.`,
		Example: `
Create a dataset:
$ prolific aitaskbuilder create-dataset -n "test" -w <workspace_id>
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := createAITaskBuilderDataset(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Name, "name", "n", "", "Name of the dataset (required)")
	flags.StringVarP(&opts.WorkspaceID, "workspace-id", "w", "", "Workspace ID (required)")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("workspace-id")

	return cmd
}

// createAITaskBuilderDataset will create a new dataset
func createAITaskBuilderDataset(c client.API, opts CreateDatasetOptions, w io.Writer) error {
	// Validate required fields
	if opts.Name == "" {
		return errors.New(ErrNameRequired)
	}
	if opts.WorkspaceID == "" {
		return errors.New(ErrWorkspaceIDRequired)
	}

	// Build payload from options
	payload := client.CreateAITaskBuilderDatasetPayload{
		Name: opts.Name,
	}

	// Call API to create
	response, err := c.CreateAITaskBuilderDataset(opts.WorkspaceID, payload)
	if err != nil {
		return err
	}

	// Output success with ID
	fmt.Fprintf(w, "Created dataset: %s\n", response.Dataset.ID)
	fmt.Fprintf(w, "Total datapoint count: %d\n", response.Dataset.TotalDatapointCount)

	return nil
}
