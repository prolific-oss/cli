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
	HTMLDir      string
}

// NewUpdateCommand returns the `collection update` command.
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

Read html_file paths from another folder:
$ prolific collection update collec12345 -t collection.yaml --html-dir ./course-pages

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
The path is relative to the template file, or to --html-dir when that flag is set.
It must stay inside that folder.
You can set html_content instead of html_file. Do not set both.
Set id to the stored block id. The id is a UUID.
The API then updates that block.
Without id, the API creates a new block and deletes the stored block.

- id: 11111111-1111-4111-8111-111111111111
  order: 0
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
	cmd.Flags().StringVar(&opts.HTMLDir, "html-dir", "", "Folder that html_file paths are relative to. A relative value is read from the current folder. Defaults to the template folder.")
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

	if err := loadUpdateCustomHTML(&updatePayload, opts.TemplatePath, opts.HTMLDir); err != nil {
		return updatePayload, err
	}

	return updatePayload, nil
}

func loadUpdateCustomHTML(payload *model.UpdateCollection, templatePath, htmlDir string) error {
	folder, err := shared.OpenHTMLFolder(filepath.Dir(templatePath), htmlDir)
	if err != nil {
		return err
	}
	defer folder.Close()

	for pageIndex := range payload.CollectionItems {
		page := &payload.CollectionItems[pageIndex]
		for itemIndex := range page.PageItems {
			item := &page.PageItems[itemIndex]
			if item.Type != model.ContentBlockTypeCustomHTML {
				continue
			}

			html, err := readCustomHTML(folder, pageIndex, itemIndex, item.HTMLFile, item.HTMLContent)
			if err != nil {
				return err
			}

			item.HTMLContent = html
			item.HTMLFile = ""
		}
	}

	return nil
}
