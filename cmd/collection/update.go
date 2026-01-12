package collection

import (
	"errors"
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/model"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type UpdateOptions struct {
	TemplatePath string
}

// NewUpdateCommand creates a new `collection update` command to update a collection
func NewUpdateCommand(client client.API, w io.Writer) *cobra.Command {
	var opts UpdateOptions

	cmd := &cobra.Command{
		Use:   "update <collection-id>",
		Short: "Update a collection",
		Long:  "Update a collection with new values from a YAML or JSON config file",
		Example: `
Update a collection using a YAML config file:
$ prolific collection update collec12345 -t collection.yaml

Update a collection using a JSON config file:
$ prolific collection update collec12345 -t collection.json

Example YAML config file:
---
name: My Updated Collection
collection_items:
  - order: 0
    page_items:
      - type: free_text
        description: "What is your feedback?"
        order: 0

Example JSON config file:
{
  "name": "My Updated Collection",
  "collection_items": [{
  	"type": "free_text",
    "description": "What is your feedback?",
    "order": 0
  }]
}`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			collectionID := args[0]

			updatePayload, err := validateTemplate(opts)
			if err != nil {
				return err
			}

			collection, err := client.UpdateCollection(collectionID, updatePayload)
			if err != nil {
				return err
			}

			fmt.Fprintf(w, "Collection updated successfully\n")
			fmt.Fprintf(w, "ID: %s\n", collection.ID)
			fmt.Fprintf(w, "Name: %s\n", collection.Name)

			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.TemplatePath, "template-path", "t", "", "Path to a YAML or JSON file containing your collection updates")
	_ = cmd.MarkFlagRequired("template-path")

	return cmd
}

func validateTemplate(opts UpdateOptions) (model.UpdateCollection, error) {
	var updatePayload model.UpdateCollection

	v := viper.New()
	v.SetConfigFile(opts.TemplatePath)
	err := v.ReadInConfig()

	if err != nil {
		return updatePayload, fmt.Errorf("unable to read config file: %s", err)
	}

	err = v.UnmarshalExact(&updatePayload)
	if err != nil {
		return updatePayload, fmt.Errorf("unable to unmarshal config file: %s", err)
	}

	if updatePayload.Name == "" {
		return updatePayload, errors.New(ErrNameRequired)
	}

	// Validate that collection_items array has at least 1 page
	if len(updatePayload.CollectionItems) == 0 {
		return updatePayload, errors.New(ErrCollectionItemsRequired)
	}

	// Validate that all pages have at least 1 item in page_items
	for i, page := range updatePayload.CollectionItems {
		if len(page.PageItems) == 0 {
			return updatePayload, fmt.Errorf("page at index %d: %s", i, ErrPageItemsRequired)
		}
	}

	return updatePayload, nil
}
