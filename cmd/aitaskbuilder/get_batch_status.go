package aitaskbuilder

import (
	"errors"
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

type BatchGetStatusOptions struct {
	Args    []string
	BatchID string
}

func NewGetBatchStatusCommand(client client.API, w io.Writer) *cobra.Command {
	var opts BatchGetStatusOptions

	cmd := &cobra.Command{
		Use:   "getbatchstatus",
		Short: "Get an AI Task Builder batch status",
		Long: `Get the status of a specific AI Task Builder batch

This command allows you to retrieve the status of a specific AI Task Builder batch by providing
the batch ID.`,
		Example: `
Get an AI Task Builder batch status:
$ prolific aitaskbuilder getbatchstatus -b <batch_id>
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := renderAITaskBuilderBatchStatus(client, opts, w)
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

func renderAITaskBuilderBatchStatus(c client.API, opts BatchGetStatusOptions, w io.Writer) error {
	if opts.BatchID == "" {
		return errors.New(ErrBatchIDRequired)
	}

	response, err := c.GetAITaskBuilderBatchStatus(opts.BatchID)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "AI Task Builder Batch Status:\n")
	fmt.Fprintf(w, "Batch ID: %s\n", opts.BatchID)
	fmt.Fprintf(w, "Status: %s\n", response.Status)

	return nil
}
