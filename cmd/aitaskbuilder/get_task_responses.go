package aitaskbuilder

import (
	"errors"
	"fmt"
	"io"
	"sort"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/model"
	"github.com/spf13/cobra"
)

type BatchGetResponsesOptions struct {
	Args    []string
	BatchID string
}

func NewGetResponsesCommand(client client.API, w io.Writer) *cobra.Command {
	var opts BatchGetResponsesOptions

	cmd := &cobra.Command{
		Use:   "responses",
		Short: "List batch task responses",
		Long: `Get the responses for a specific AI Task Builder batch

This command allows you to retrieve all responses for a specific AI Task Builder batch by providing
the batch ID.`,
		Example: `
Get AI Task Builder batch responses:
$ prolific aitaskbuilder batch responses -b <batch_id>
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
	fmt.Fprintf(w, "Total Responses: %d\n", response.Meta.Count)
	fmt.Fprintf(w, "\n")

	if len(response.Results) == 0 {
		fmt.Fprintf(w, "No responses found for batch %s\n", opts.BatchID)
		return nil
	}

	for i, resp := range response.Results {
		fmt.Fprintf(w, "Response %d:\n", i+1)
		fmt.Fprintf(w, "  ID: %s\n", resp.ID)
		fmt.Fprintf(w, "  Batch ID: %s\n", resp.BatchID)
		fmt.Fprintf(w, "  Participant ID: %s\n", resp.ParticipantID)
		fmt.Fprintf(w, "  Task ID: %s\n", resp.TaskID)
		fmt.Fprintf(w, "  Correlation ID: %s\n", resp.CorrelationID)
		fmt.Fprintf(w, "  Submission ID: %s\n", resp.SubmissionID)
		fmt.Fprintf(w, "  Schema Version: %d\n", resp.SchemaVersion)
		fmt.Fprintf(w, "  Created At: %s\n", resp.CreatedAt.Format("2006-01-02 15:04:05"))

		if len(resp.Metadata) > 0 {
			fmt.Fprintf(w, "  Metadata:\n")
			// Sort keys to ensure deterministic output
			keys := make([]string, 0, len(resp.Metadata))
			for key := range resp.Metadata {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			for _, key := range keys {
				fmt.Fprintf(w, "    %s: %s\n", key, resp.Metadata[key])
			}
		}

		fmt.Fprintf(w, "  Response:\n")
		fmt.Fprintf(w, "    Instruction ID: %s\n", resp.Response.InstructionID)
		fmt.Fprintf(w, "    Type: %s\n", string(resp.Response.Type))

		// Handle different response types
		switch resp.Response.Type {
		case model.AITaskBuilderResponseTypeFreeText:
			if resp.Response.Text != nil {
				fmt.Fprintf(w, "    Text: %s\n", *resp.Response.Text)
			} else {
				fmt.Fprintf(w, "    Text: \n")
			}
		case model.AITaskBuilderResponseTypeMultipleChoice:
			if len(resp.Response.Answer) > 0 {
				fmt.Fprintf(w, "    Selected Options:\n")
				for _, option := range resp.Response.Answer {
					fmt.Fprintf(w, "      - %s\n", option.Value)
				}
			} else {
				fmt.Fprintf(w, "    Selected Options: \n")
			}
		case model.AITaskBuilderResponseTypeMultipleChoiceWithFreeText:
			if len(resp.Response.Answer) > 0 {
				fmt.Fprintf(w, "    Selected Options:\n")
				for _, option := range resp.Response.Answer {
					fmt.Fprintf(w, "      - %s\n", option.Value)
				}
			} else {
				fmt.Fprintf(w, "    Selected Options: \n")
			}
			if resp.Response.Text != nil {
				fmt.Fprintf(w, "    Additional Text: %s\n", *resp.Response.Text)
			} else {
				fmt.Fprintf(w, "    Additional Text: \n")
			}
		case model.AITaskBuilderResponseTypeMultipleChoiceWithUnit:
			if len(resp.Response.Answer) > 0 {
				fmt.Fprintf(w, "    Selected Options:\n")
				for _, option := range resp.Response.Answer {
					fmt.Fprintf(w, "      - %s\n", option.Value)
				}
			} else {
				fmt.Fprintf(w, "    Selected Options: \n")
			}
			if resp.Response.Unit != nil {
				fmt.Fprintf(w, "    Unit: %s\n", *resp.Response.Unit)
			} else {
				fmt.Fprintf(w, "    Unit: \n")
			}
		case model.AITaskBuilderResponseTypeFileUpload:
			if resp.Response.FileReference != nil {
				fmt.Fprintf(w, "    File Reference: %s\n", *resp.Response.FileReference)
			} else {
				fmt.Fprintf(w, "    File Reference: \n")
			}
		}

		if i < len(response.Results)-1 {
			fmt.Fprintf(w, "\n")
		}
	}

	return nil
}
