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
		Short: "Setup a batch",
		Long: `Setup an AI Task Builder batch with a dataset and task configuration

This command configures a batch by linking it to a dataset and specifying how many tasks
each participant group should complete. The batch must already be created and have
instructions added before setup.`,
		Example: `
Setup a batch:
$ prolific aitaskbuilder batch setup -b <batch_id> -d <dataset_id> --tasks-per-group 3
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := setupAITaskBuilderBatch(client, opts, w)
			if err != nil {
				return err
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.BatchID, "batch-id", "b", "", "Batch ID (required) - The ID of the batch to setup.")
	flags.StringVarP(&opts.DatasetID, "dataset-id", "d", "", "Dataset ID (required) - The ID of the dataset to use for setup.")
	flags.IntVar(&opts.TasksPerGroup, "tasks-per-group", 1, "Tasks per group - The number of tasks to assign per group (default 1).")

	_ = cmd.MarkFlagRequired("batch-id")
	_ = cmd.MarkFlagRequired("dataset-id")

	return cmd
}

// setupAITaskBuilderBatch will setup an AI Task Builder batch
func setupAITaskBuilderBatch(c client.API, opts BatchSetupOptions, w io.Writer) error {
	if opts.BatchID == "" {
		return errors.New(ErrBatchIDRequired)
	}
	if opts.DatasetID == "" {
		return errors.New(ErrDatasetIDRequired)
	}
	if opts.TasksPerGroup < 1 {
		return errors.New(ErrTasksPerGroupMinimum)
	}

	_, err := c.SetupAITaskBuilderBatch(opts.BatchID, opts.DatasetID, opts.TasksPerGroup)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "Successfully setup batch %s\n", opts.BatchID)
	fmt.Fprintf(w, "Dataset ID: %s\n", opts.DatasetID)
	fmt.Fprintf(w, "Tasks per Group: %d\n", opts.TasksPerGroup)

	return nil
}
