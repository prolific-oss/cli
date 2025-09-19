//nolint:dupl // Similar patterns are expected for CLI commands
package aitaskbuilder

import (
	"errors"
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

type DatasetGetStatusOptions struct {
	Args      []string
	DatasetID string
}

func NewGetDatasetStatusCommand(client client.API, w io.Writer) *cobra.Command {
	var opts DatasetGetStatusOptions

	cmd := &cobra.Command{
		Use:   "getdatasetstatus",
		Short: "Get an AI Task Builder dataset status",
		Long: `Get the status of a specific AI Task Builder dataset

This command allows you to retrieve the status of a specific AI Task Builder dataset by providing
the dataset ID.

The status of a dataset can transition to one of the following:

• UNINITIALISED - This means that the dataset has been created, but no data has been uploaded to it yet.
• PROCESSING - This means that the dataset is being processed into datapoints to use in the task configuration process.
• READY - This means that the dataset is completely processed into datapoints and ready to be used within a batch.
• ERROR - This means that something has gone wrong during processing and the data may not be usable.`,
		Example: `
Get an AI Task Builder dataset status:
$ prolific aitaskbuilder getdatasetstatus -d <dataset_id>
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := renderAITaskBuilderDatasetStatus(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.DatasetID, "dataset-id", "d", "", "Dataset ID (required) - The ID of the dataset to retrieve.")

	_ = cmd.MarkFlagRequired("dataset-id")

	return cmd
}

func renderAITaskBuilderDatasetStatus(c client.API, opts DatasetGetStatusOptions, w io.Writer) error {
	if opts.DatasetID == "" {
		return errors.New(ErrDatasetIDRequired)
	}

	response, err := c.GetAITaskBuilderDatasetStatus(opts.DatasetID)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "AI Task Builder Dataset Status:\n")
	fmt.Fprintf(w, "Dataset ID: %s\n", opts.DatasetID)
	fmt.Fprintf(w, "Status: %s\n", response.Status)

	return nil
}
