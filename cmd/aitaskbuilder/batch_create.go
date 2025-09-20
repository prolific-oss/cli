package aitaskbuilder

import (
	"errors"
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

type BatchCreateOptions struct {
	Args             []string
	Name             string
	WorkspaceID      string
	DatasetID        string
	TaskName         string
	TaskIntroduction string
	TaskSteps        string
}

func NewBatchCreateCommand(client client.API, w io.Writer) *cobra.Command {
	var opts BatchCreateOptions

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an AI Task Builder batch",
		Long: `Create a new AI Task Builder batch

This command allows you to create a new AI Task Builder batch by providing the batch name,
workspace ID, dataset ID, and task details.`,
		Example: `
Create an AI Task Builder batch:
$ prolific aitaskbuilder batch create -n "My Data Collection Batch" -w 6278acb09062db3b35bcbeb0 -d 1234acb09999db4b99bcded1 --task-name "Sample Task" --task-introduction "This is a sample task for testing" --task-steps "1. Review the data\n2. Provide your response"
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := createAITaskBuilderBatch(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Name, "name", "n", "", "Batch name (required) - The name of the batch to create.")
	flags.StringVarP(&opts.WorkspaceID, "workspace-id", "w", "", "Workspace ID (required) - The ID of the workspace where the batch will be created.")
	flags.StringVarP(&opts.DatasetID, "dataset-id", "d", "", "Dataset ID (required) - The ID of the dataset to use for the batch.")
	flags.StringVar(&opts.TaskName, "task-name", "", "Task name (required) - The name of the task.")
	flags.StringVar(&opts.TaskIntroduction, "task-introduction", "", "Task introduction (required) - The introduction text for the task.")
	flags.StringVar(&opts.TaskSteps, "task-steps", "", "Task steps (required) - The steps for completing the task.")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("workspace-id")
	_ = cmd.MarkFlagRequired("dataset-id")
	_ = cmd.MarkFlagRequired("task-name")
	_ = cmd.MarkFlagRequired("task-introduction")
	_ = cmd.MarkFlagRequired("task-steps")

	return cmd
}

// createAITaskBuilderBatch will create a new AI Task Builder batch
func createAITaskBuilderBatch(c client.API, opts BatchCreateOptions, w io.Writer) error {
	if opts.Name == "" {
		return errors.New("batch name is required")
	}
	if opts.WorkspaceID == "" {
		return errors.New("workspace ID is required")
	}
	if opts.DatasetID == "" {
		return errors.New("dataset ID is required")
	}
	if opts.TaskName == "" {
		return errors.New("task name is required")
	}
	if opts.TaskIntroduction == "" {
		return errors.New("task introduction is required")
	}
	if opts.TaskSteps == "" {
		return errors.New("task steps is required")
	}

	response, err := c.CreateAITaskBuilderBatch(opts.Name, opts.WorkspaceID, opts.DatasetID, opts.TaskName, opts.TaskIntroduction, opts.TaskSteps)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "AI Task Builder Batch Created Successfully:\n")
	fmt.Fprintf(w, "ID: %s\n", response.ID)
	fmt.Fprintf(w, "Name: %s\n", response.Name)
	fmt.Fprintf(w, "Status: %s\n", response.Status)
	fmt.Fprintf(w, "Total Task Count: %d\n", response.TotalTaskCount)
	fmt.Fprintf(w, "Total Instruction Count: %d\n", response.TotalInstructionCount)
	if response.TotalTaskGroups != nil {
		fmt.Fprintf(w, "Total Task Groups: %d\n", *response.TotalTaskGroups)
	}
	fmt.Fprintf(w, "Workspace ID: %s\n", response.WorkspaceID)
	fmt.Fprintf(w, "Created By: %s\n", response.CreatedBy)
	fmt.Fprintf(w, "Created At: %s\n", response.CreatedAt)

	if len(response.Datasets) > 0 {
		fmt.Fprintf(w, "Datasets: %d\n", len(response.Datasets))
		for i, dataset := range response.Datasets {
			fmt.Fprintf(w, "  Dataset %d: %s (%d datapoints)\n", i+1, dataset.ID, dataset.TotalDatapointCount)
		}
	}

	fmt.Fprintf(w, "\nTask Details:\n")
	fmt.Fprintf(w, "  Name: %s\n", response.TaskDetails.TaskName)
	fmt.Fprintf(w, "  Introduction: %s\n", response.TaskDetails.TaskIntroduction)
	fmt.Fprintf(w, "  Steps: %s\n", response.TaskDetails.TaskSteps)

	return nil
}
