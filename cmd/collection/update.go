package collection

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/shared"
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
task_details:
  task_name: Updated task name
  task_introduction: "<p>Updated task introduction</p>"
  task_steps: "<ol><li>Updated Step 1</li></ol>"
collection_items:
  - order: 0
    page_items:
      - type: free_text
        description: "What is your feedback?"
        order: 0

Example JSON config file:
{
  "name": "My Updated Collection",
  "task_details": {
    "task_name": "Updated task name",
    "task_introduction": "<p>Updated task introduction</p>",
    "task_steps": "<ol><li>Updated Step 1</li></ol>"
  },
  "collection_items": [{
    "order": 0,
    "page_items": [{
      "type": "free_text",
      "description": "What is your feedback?",
      "order": 0
    }]
  }]
}

A snippet for one custom_html item. This is not a complete file.
The path is relative to the template file.
You can set html_content instead of html_file. Do not set both.

- order: 0
  type: custom_html
  html_file: course.html`,
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

	if err := loadUpdateCustomHTML(&updatePayload, opts.TemplatePath); err != nil {
		return updatePayload, err
	}

	return updatePayload, nil
}

// loadUpdateCustomHTML reads html_file for each custom_html item and checks the size limits.
// PageInstruction.Type is an InstructionType, so it is compared without a string conversion.
func loadUpdateCustomHTML(payload *model.UpdateCollection, templatePath string) error {
	baseDir := filepath.Dir(templatePath)
	var check shared.CustomHTMLSizeCheck

	for pageIndex := range payload.CollectionItems {
		page := &payload.CollectionItems[pageIndex]
		for itemIndex := range page.PageItems {
			item := &page.PageItems[itemIndex]
			if item.Type != model.ContentBlockTypeCustomHTML {
				continue
			}

			location := fmt.Sprintf("page %d, item %d", page.Order, item.Order)
			html, err := shared.ResolveCustomHTML(baseDir, item.HTMLFile, item.HTMLContent)
			if err != nil {
				return fmt.Errorf("%s: %w", location, err)
			}

			item.HTMLContent = html
			item.HTMLFile = ""
			if err := check.Add(location, item.HTMLContent); err != nil {
				return err
			}
		}
	}

	return nil
}
