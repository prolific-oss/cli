package aitaskbuilder

import (
	"errors"
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// DatasetCreateOptions are the options for creating an AI Task Builder dataset.
type DatasetCreateOptions struct {
	Args        []string
	Name        string
	WorkspaceID string
}

// NewDatasetCreateCommand creates a new command for creating an AI Task Builder dataset.
func NewDatasetCreateCommand(client client.API, w io.Writer) *cobra.Command {
	var opts DatasetCreateOptions

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an AI Task Builder dataset",
		Long: `Create an AI Task Builder dataset in your workspace

Create a new dataset that can be used to store data for AI Task Builder batches.
Once created, you can upload data to the dataset and use it in batch operations.`,
		Example: `
Create an AI Task Builder dataset:
$ prolific aitaskbuilder dataset create -n "Test Dataset" -w 6278acb09062db3b35bcbeb0
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := createDataset(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Name, "name", "n", "", "The name of the dataset (required)")
	flags.StringVarP(&opts.WorkspaceID, "workspace-id", "w", "", "The ID of the workspace to create the dataset in (required)")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("workspace-id")

	return cmd
}

// createDataset will create an AI Task Builder dataset
func createDataset(client client.API, opts DatasetCreateOptions, w io.Writer) error {
	if opts.Name == "" {
		return errors.New("name is required")
	}

	if opts.WorkspaceID == "" {
		return errors.New("workspace ID is required")
	}

	response, err := client.CreateAITaskBuilderDataset(opts.Name, opts.WorkspaceID)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "AI Task Builder Dataset Created:\n")
	fmt.Fprintf(w, "ID: %s\n", response.ID)
	fmt.Fprintf(w, "Name: %s\n", response.Name)
	fmt.Fprintf(w, "Workspace ID: %s\n", response.WorkspaceID)
	fmt.Fprintf(w, "Status: %s\n", response.Status)
	fmt.Fprintf(w, "Created At: %s\n", response.CreatedAt)

	return nil
}
