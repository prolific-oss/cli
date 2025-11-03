package aitaskbuilder

import (
	"errors"
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

type BatchGetOptions struct {
	Args    []string
	BatchID string
}

func NewGetBatchCommand(client client.API, w io.Writer) *cobra.Command {
	var opts BatchGetOptions

	cmd := &cobra.Command{
		Use:   "view",
		Short: "Get batch details",
		Long: `View details for a specific batch.

Retrieve details of a batch by id.`,
		Example: `
Get an AI Task Builder batch:
$ prolific aitaskbuilder batch view -b <batch_id>
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := renderAITaskBuilderBatch(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.BatchID, "batch-id", "b", "", "Batch ID (required) - The ID of the batch to retrieve.")

	_ = cmd.MarkFlagRequired("batch-id")

	return cmd
}

// renderAITaskBuilderBatch will show details of a specific AI Task Builder batch
func renderAITaskBuilderBatch(c client.API, opts BatchGetOptions, w io.Writer) error {
	if opts.BatchID == "" {
		return errors.New(ErrBatchIDRequired)
	}

	response, err := c.GetAITaskBuilderBatch(opts.BatchID)
	if err != nil {
		return err
	}

	batch := response.AITaskBuilderBatch

	fmt.Fprintf(w, "AI Task Builder Batch Details:\n")
	fmt.Fprintf(w, "ID: %s\n", batch.ID)
	fmt.Fprintf(w, "Name: %s\n", batch.Name)
	fmt.Fprintf(w, "Status: %s\n", batch.Status)
	fmt.Fprintf(w, "Total Task Count: %d\n", batch.TotalTaskCount)
	fmt.Fprintf(w, "Total Instruction Count: %d\n", batch.TotalInstructionCount)
	fmt.Fprintf(w, "Workspace ID: %s\n", batch.WorkspaceID)
	fmt.Fprintf(w, "Created By: %s\n", batch.CreatedBy)
	fmt.Fprintf(w, "Created At: %s\n", batch.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(w, "Schema Version: %d\n", batch.SchemaVersion)

	if len(batch.Datasets) > 0 {
		fmt.Fprintf(w, "Datasets: %d\n", len(batch.Datasets))
		for i, dataset := range batch.Datasets {
			fmt.Fprintf(w, "  Dataset %d: %s (%d datapoints)\n", i+1, dataset.ID, dataset.TotalDatapointCount)
		}
	}

	if batch.TaskDetails.TaskName != "" {
		fmt.Fprintf(w, "\nTask Details:\n")
		fmt.Fprintf(w, "  Name: %s\n", batch.TaskDetails.TaskName)
		fmt.Fprintf(w, "  Introduction: %s\n", batch.TaskDetails.TaskIntroduction)
		fmt.Fprintf(w, "  Steps: %s\n", batch.TaskDetails.TaskSteps)
	}

	return nil
}
