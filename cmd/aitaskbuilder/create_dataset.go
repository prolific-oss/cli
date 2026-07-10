package aitaskbuilder

import (
	"errors"
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// CreateDatasetOptions holds the options for creating a dataset
type CreateDatasetOptions struct {
	Args        []string
	Name        string
	WorkspaceID string
	Schema      string // raw --schema value (inline JSON or path)
	Strict      bool   // --strict flag value
	StrictSet   bool   // whether --strict was explicitly passed
}

// NewCreateDatasetCommand creates a new command for creating datasets
func NewCreateDatasetCommand(client client.API, w io.Writer) *cobra.Command {
	var opts CreateDatasetOptions

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Dataset",
		Long: `Create a new AI Task Builder dataset

A dataset contains the data that will be used for annotation tasks. You must provide:
- A name for the dataset
- The workspace ID where the dataset will be created

The workspace ID determines which workspace owns and has access to this dataset.

Optionally attach a typed schema with --schema (inline JSON or a path to a JSON
file) to define the dataset's fields up front. When passing inline JSON in a
shell, quote the entire value (for example with single quotes) so the shell
passes it through unchanged. The value is the full schema object, for example:

  {
    "strict": true,
    "fields": {
      "question": { "type": "text", "label": "Question" },
      "image":    { "type": "image_url" },
      "source":   { "type": "metadata" },
      "group":    { "type": "task_group_id" }
    }
  }

Field types are text, image_url, metadata, and task_group_id (at most one). By
default schemas are created with "strict": false. Use --strict to enable strict
mode when the schema JSON does not already set "strict" (passing --strict
alongside a schema that sets "strict" is an error). See
docs/examples/dataset-schema.json for a full example.

Schemas are only accepted for workspaces where the typed-dataset feature is
enabled; supplying --schema for a workspace without it will be rejected by the
API. Omit --schema to leave the schema unset.`,
		Example: `
Create a dataset:
$ prolific aitaskbuilder dataset create -n "test" -w <workspace_id>

Create a dataset with a schema from a file:
$ prolific aitaskbuilder dataset create -n "test" -w <workspace_id> --schema docs/examples/dataset-schema.json

Create a dataset with an inline schema in strict mode:
$ prolific aitaskbuilder dataset create -n "test" -w <workspace_id> --strict --schema '{"fields":{"q":{"type":"text"}}}'
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			opts.StrictSet = cmd.Flags().Changed("strict")

			err := createAITaskBuilderDataset(client, opts, w)
			if err != nil {
				return err
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Name, "name", "n", "", "Name of the dataset (required)")
	flags.StringVarP(&opts.WorkspaceID, "workspace-id", "w", "", "Workspace ID (required)")
	flags.StringVar(&opts.Schema, "schema", "", "Optional dataset schema as quoted inline JSON or a path to a JSON file")
	flags.BoolVar(&opts.Strict, "strict", false, "Enable strict mode: reject records missing any schema field (requires --schema)")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("workspace-id")

	return cmd
}

// createAITaskBuilderDataset will create a new dataset
func createAITaskBuilderDataset(c client.API, opts CreateDatasetOptions, w io.Writer) error {
	// Validate required fields
	if opts.Name == "" {
		return errors.New(ErrNameRequired)
	}
	if opts.WorkspaceID == "" {
		return errors.New(ErrWorkspaceIDRequired)
	}

	// Resolve the optional schema (nil when --schema is omitted, so it is left
	// out of the payload). The dataset version is decided server-side.
	schema, err := resolveDatasetSchema(opts.Schema, opts.Strict, opts.StrictSet)
	if err != nil {
		return err
	}

	// Build payload from options
	payload := client.CreateAITaskBuilderDatasetPayload{
		Name:   opts.Name,
		Schema: schema,
	}

	// Call API to create dataset in the specified workspace
	response, err := c.CreateAITaskBuilderDataset(opts.WorkspaceID, payload)
	if err != nil {
		return err
	}

	// Output full dataset details
	fmt.Fprintf(w, "ID: %s\n", response.ID)
	fmt.Fprintf(w, "Name: %s\n", response.Name)
	fmt.Fprintf(w, "Created At: %s\n", response.CreatedAt)
	fmt.Fprintf(w, "Created By: %s\n", response.CreatedBy)
	fmt.Fprintf(w, "Status: %s\n", response.Status)
	fmt.Fprintf(w, "Total Datapoint Count: %d\n", response.TotalDatapointCount)
	fmt.Fprintf(w, "Workspace ID: %s\n", response.WorkspaceID)
	fmt.Fprintf(w, "Schema Version: %d\n", response.SchemaVersion)
	if payload.Schema != nil {
		if payload.Schema.Strict != nil {
			fmt.Fprintf(w, "Strict: %t\n", *payload.Schema.Strict)
		}
		fmt.Fprintf(w, "Schema Fields: %d\n", len(payload.Schema.Fields))
	}

	return nil
}
