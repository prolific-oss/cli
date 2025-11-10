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
		Use:   "create",
		Short: "Create a Dataset",
		Long: `Create a new AI Task Builder dataset

A dataset contains the data that will be used for annotation tasks. You must provide:
- A name for the dataset
- The workspace ID where the dataset will be created

The workspace ID determines which workspace owns and has access to this dataset.`,
		Example: `
Create a dataset:
$ prolific aitaskbuilder dataset create -n "test" -w <workspace_id>
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

	// Call API to create dataset in the specified workspace
	response, err := c.CreateAITaskBuilderDataset(opts.WorkspaceID, payload)
	if err != nil {
		return err
	}

	// Output full dataset details
	fmt.Fprintf(w, "ID: %s\n", response.ID)
	fmt.Fprintf(w, "Name: %s\n", response.Name)
	fmt.Fprintf(w, "Created At: %s\n", response.CreatedAt)
	fmt.Fprintf(w, "Created By: %s\n", response.CreatedBy)
	fmt.Fprintf(w, "Status: %s\n", response.Status)
	fmt.Fprintf(w, "Total Datapoint Count: %d\n", response.TotalDatapointCount)
	fmt.Fprintf(w, "Workspace ID: %s\n", response.WorkspaceID)

	return nil
}
