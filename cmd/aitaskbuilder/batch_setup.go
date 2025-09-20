package aitaskbuilder

import (
	"errors"
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

type BatchSetupOptions struct {
	Args          []string
	BatchID       string
	DatasetID     string
	TasksPerGroup int
}

func NewBatchSetupCommand(client client.API, w io.Writer) *cobra.Command {
	var opts BatchSetupOptions

	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Setup an AI Task Builder batch",
		Long: `Setup an AI Task Builder batch with a dataset and task configuration

This command allows you to setup a specific AI Task Builder batch by providing the batch ID,
dataset ID, and number of tasks per group.`,
		Example: `
Setup an AI Task Builder batch:
$ prolific aitaskbuilder batch setup -b 01954894-65b3-779e-aaf6-348698e23634 -d 8c4c51f1-f6f3-43bc-b65d-7415e8ef22c0 --tasks-per-group 3
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := setupAITaskBuilderBatch(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.BatchID, "batch-id", "b", "", "Batch ID (required) - The ID of the batch to setup.")
	flags.StringVarP(&opts.DatasetID, "dataset-id", "d", "", "Dataset ID (required) - The ID of the dataset to use for setup.")
	flags.IntVar(&opts.TasksPerGroup, "tasks-per-group", 1, "Tasks per group (required) - The number of tasks to assign per group.")

	_ = cmd.MarkFlagRequired("batch-id")
	_ = cmd.MarkFlagRequired("dataset-id")
	_ = cmd.MarkFlagRequired("tasks-per-group")

	return cmd
}

// setupAITaskBuilderBatch will setup an AI Task Builder batch
func setupAITaskBuilderBatch(c client.API, opts BatchSetupOptions, w io.Writer) error {
	if opts.BatchID == "" {
		return errors.New("batch ID is required")
	}
	if opts.DatasetID == "" {
		return errors.New("dataset ID is required")
	}
	if opts.TasksPerGroup < 1 {
		return errors.New("tasks per group must be at least 1")
	}

	_, err := c.SetupAITaskBuilderBatch(opts.BatchID, opts.DatasetID, opts.TasksPerGroup)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "AI Task Builder Batch Setup Complete:\n")
	fmt.Fprintf(w, "Batch ID: %s\n", opts.BatchID)
	fmt.Fprintf(w, "Dataset ID: %s\n", opts.DatasetID)
	fmt.Fprintf(w, "Tasks per Group: %d\n", opts.TasksPerGroup)

	return nil
}
