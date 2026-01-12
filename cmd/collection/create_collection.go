package collection

import (
	"errors"
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/shared"
	"github.com/prolific-oss/cli/model"
	"github.com/prolific-oss/cli/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// CreateCollectionOptions is the options for creating a collection command.
type CreateCollectionOptions struct {
	Args         []string
	TemplatePath string
}

// NewCreateCollectionCommand creates a new `collection create` command to allow you to create
// a collection.
func NewCreateCollectionCommand(c client.API, w io.Writer) *cobra.Command {
	var opts CreateCollectionOptions

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new collection",
		Long:  `Create a new AI Task Builder collection from a JSON or YAML file`,
		Example: `
To create a collection via the CLI, define your collection as a JSON/YAML file
$ prolific collection create -t docs/examples/collection.json
$ prolific collection create -t docs/examples/collection.yaml

An example of a JSON collection file:

{
  "workspace_id": "67890abcdef12345678901234",
  "name": "example-collection",
  "collection_items": [
    {
      "order": 0,
      "page_items": [
        {
          "order": 0,
          "type": "free_text",
          "description": "How was your experience completing this task?"
        },
        {
          "order": 1,
          "type": "multiple_choice",
          "description": "Which option do you prefer?",
          "options": [
            {
              "label": "Response 1",
              "value": "response1"
            },
            {
              "label": "Response 2",
              "value": "response2"
            }
          ],
          "answer_limit": -1
        }
      ]
    }
  ]
}

An example of a YAML collection file:

---
workspace_id: 67890abcdef12345678901234
name: example-collection
collection_items:
  - order: 0
    page_items:
      - order: 0
        type: free_text
        description: How was your experience completing this task?
      - order: 1
        type: multiple_choice
        description: Which option do you prefer?
        options:
          - label: Response 1
            value: response1
          - label: Response 2
            value: response2
        answer_limit: -1
---`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			if opts.TemplatePath == "" {
				return fmt.Errorf("error: template path required. Use -t or --template-path flag")
			}

			err := createCollection(c, opts, w)
			if err != nil {
				if shared.IsFeatureNotEnabledError(err) {
					ui.RenderFeatureAccessMessage(FeatureNameAITBCollection, FeatureContactEmailAITBCollection)
					return nil
				}
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.TemplatePath, "template-path", "t", "", "Path to a YAML or JSON file containing the collection you want to create")
	_ = cmd.MarkFlagRequired("template-path")

	return cmd
}

func validatePayload(payload model.CreateAITaskBuilderCollection) error {
	// Validate required fields
	if payload.Name == "" {
		return errors.New(ErrNameRequired)
	}

	if payload.WorkspaceID == "" {
		return errors.New(ErrWorkspaceIDRequired)
	}

	if len(payload.CollectionItems) == 0 {
		return errors.New(ErrCollectionItemsRequired)
	}

	return nil
}

func createCollection(c client.API, opts CreateCollectionOptions, w io.Writer) error {
	v := viper.New()
	v.SetConfigFile(opts.TemplatePath)
	err := v.ReadInConfig()
	if err != nil {
		return fmt.Errorf("unable to read config file: %s", err)
	}

	var payload model.CreateAITaskBuilderCollection
	err = v.Unmarshal(&payload)
	if err != nil {
		return fmt.Errorf("unable to map %s to collection model: %s", opts.TemplatePath, err)
	}

	if err := validatePayload(payload); err != nil {
		return err
	}

	collection, err := c.CreateAITaskBuilderCollection(payload)
	if err != nil {
		return err
	}

	// Output collection details
	fmt.Fprintf(w, "Collection created successfully!\n")
	fmt.Fprintf(w, "ID:              %s\n", collection.ID)
	fmt.Fprintf(w, "Name:            %s\n", collection.Name)
	fmt.Fprintf(w, "Workspace ID:    %s\n", collection.WorkspaceID)
	fmt.Fprintf(w, "Schema Version:  %d\n", collection.SchemaVersion)
	fmt.Fprintf(w, "Created By:      %s\n", collection.CreatedBy)
	fmt.Fprintf(w, "Pages:           %d\n", len(collection.CollectionItems))

	return nil
}
