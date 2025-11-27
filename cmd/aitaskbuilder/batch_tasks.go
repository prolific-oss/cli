package aitaskbuilder

import (
	"errors"
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

type BatchTasksOptions struct {
	Args    []string
	BatchID string
}

func NewBatchTasksCommand(client client.API, w io.Writer) *cobra.Command {
	var opts BatchTasksOptions

	cmd := &cobra.Command{
		Use:   "tasks",
		Short: "Get AI Task Builder batch tasks",
		Long: `Get the tasks for a specific AI Task Builder batch

This command allows you to retrieve all tasks for a specific AI Task Builder batch by providing
the batch ID.`,
		Example: `
Get AI Task Builder batch tasks:
$ prolific aitaskbuilder batch tasks -b <batch_id>
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := renderAITaskBuilderTasks(client, opts, w)
			if err != nil {
				return err
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.BatchID, "batch-id", "b", "", "Batch ID (required) - The ID of the batch to retrieve tasks from.")

	_ = cmd.MarkFlagRequired("batch-id")

	return cmd
}

func renderAITaskBuilderTasks(c client.API, opts BatchTasksOptions, w io.Writer) error {
	if opts.BatchID == "" {
		return errors.New(ErrBatchIDRequired)
	}

	taskIDs, err := c.GetAITaskBuilderTasks(opts.BatchID)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "AI Task Builder Batch Tasks:\n")
	fmt.Fprintf(w, "Batch ID: %s\n", opts.BatchID)
	fmt.Fprintf(w, "Total Tasks: %d\n", len(*taskIDs))
	fmt.Fprintf(w, "\n")

	if len(*taskIDs) == 0 {
		fmt.Fprintf(w, "No tasks found for batch %s\n", opts.BatchID)
		return nil
	}

	fmt.Fprintf(w, "Task IDs:\n")
	for i, taskID := range *taskIDs {
		fmt.Fprintf(w, "  %d. %s\n", i+1, taskID)
	}

	return nil
}
