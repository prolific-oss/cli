package aitaskbuilder

import (
	"errors"
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

type BatchGetResponsesOptions struct {
	Args    []string
	BatchID string
}

func NewGetResponsesCommand(client client.API, w io.Writer) *cobra.Command {
	var opts BatchGetResponsesOptions

	cmd := &cobra.Command{
		Use:   "getresponses",
		Short: "Get AI Task Builder batch responses",
		Long: `Get the responses for a specific AI Task Builder batch

This command allows you to retrieve all responses for a specific AI Task Builder batch by providing
the batch ID.`,
		Example: `
Get AI Task Builder batch responses:
$ prolific aitaskbuilder getresponses -b <batch_id>
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := renderAITaskBuilderResponses(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.BatchID, "batch-id", "b", "", "Batch ID (required) - The ID of the batch to retrieve responses from.")

	_ = cmd.MarkFlagRequired("batch-id")

	return cmd
}

func renderAITaskBuilderResponses(c client.API, opts BatchGetResponsesOptions, w io.Writer) error {
	if opts.BatchID == "" {
		return errors.New(ErrBatchIDRequired)
	}

	response, err := c.GetAITaskBuilderResponses(opts.BatchID)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "AI Task Builder Batch Responses:\n")
	fmt.Fprintf(w, "Batch ID: %s\n", opts.BatchID)
	fmt.Fprintf(w, "Total Responses: %d\n", len(response.Responses))
	fmt.Fprintf(w, "\n")

	if len(response.Responses) == 0 {
		fmt.Fprintf(w, "No responses found for batch %s\n", opts.BatchID)
		return nil
	}

	for i, resp := range response.Responses {
		fmt.Fprintf(w, "Response %d:\n", i+1)
		fmt.Fprintf(w, "  ID: %s\n", resp.ID)
		fmt.Fprintf(w, "  Participant ID: %s\n", resp.ParticipantID)
		fmt.Fprintf(w, "  Task ID: %s\n", resp.TaskID)
		fmt.Fprintf(w, "  Created At: %s\n", resp.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Fprintf(w, "  Response:\n")
		fmt.Fprintf(w, "    Instruction ID: %s\n", resp.Response.InstructionID)
		fmt.Fprintf(w, "    Type: %s\n", resp.Response.Type)
		fmt.Fprintf(w, "    Answer: %s\n", resp.Response.Answer)

		if i < len(response.Responses)-1 {
			fmt.Fprintf(w, "\n")
		}
	}

	return nil
}
